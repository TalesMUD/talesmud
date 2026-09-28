package resources

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
)

func openStore(t *testing.T) *Store {
	t.Helper()
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "resources.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	store, err := New(client.DB())
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	return store
}

func TestUnconfiguredKeyDoesNotCreateABalance(t *testing.T) {
	store := openStore(t)
	res, ok, err := store.Get("char-1", "gatherings")
	if err != nil || ok || res.Remaining != 0 {
		t.Fatalf("get unconfigured = %+v ok=%v err=%v", res, ok, err)
	}
	if _, err := store.Consume("char-1", "gatherings", 1); !errors.Is(err, ErrUnknown) {
		t.Fatalf("consume unconfigured err = %v", err)
	}
	var n int
	if err := countRows(t, store, &n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("rows = %d, want 0", n)
	}
}

func TestCalendarRefillUsesTimezone(t *testing.T) {
	store := openStore(t)
	// 2026-09-26 23:30 UTC is still the 26th in UTC and already the 27th in Berlin.
	utcLate := time.Date(2026, 9, 26, 23, 30, 0, 0, time.UTC)
	store.SetNow(func() time.Time { return utcLate })
	store.Configure([]Allowance{{
		Key: "gatherings", Amount: 4, Reset: ResetCalendar, Timezone: "Europe/Berlin",
	}})

	res, ok, err := store.Get("char-1", "gatherings")
	if err != nil || !ok {
		t.Fatalf("get: ok=%v err=%v", ok, err)
	}
	if res.Period != "2026-09-27" || res.Remaining != 4 {
		t.Fatalf("berlin day = %+v", res)
	}
	spent, err := store.Consume("char-1", "gatherings", 3)
	if err != nil || spent.Remaining != 1 {
		t.Fatalf("after consume %+v err=%v", spent, err)
	}

	// Same Berlin date, later UTC instant. Balance must stay spent.
	store.SetNow(func() time.Time {
		return time.Date(2026, 9, 27, 10, 0, 0, 0, time.UTC)
	})
	again, _, err := store.Get("char-1", "gatherings")
	if err != nil || again.Remaining != 1 || again.Period != "2026-09-27" {
		t.Fatalf("same day = %+v err=%v", again, err)
	}

	// Next Berlin morning.
	store.SetNow(func() time.Time {
		return time.Date(2026, 9, 27, 22, 30, 0, 0, time.UTC)
	})
	refilled, _, err := store.Get("char-1", "gatherings")
	if err != nil || refilled.Period != "2026-09-28" || refilled.Remaining != 4 {
		t.Fatalf("next day = %+v err=%v", refilled, err)
	}
}

func TestIntervalRefill(t *testing.T) {
	store := openStore(t)
	start := time.Date(2026, 1, 2, 15, 0, 0, 0, time.UTC)
	now := start
	store.SetNow(func() time.Time { return now })
	store.Configure([]Allowance{{
		Key: "delve", Amount: 3, Reset: ResetInterval, Interval: time.Hour,
	}})

	if _, err := store.Consume("char-1", "delve", 1); err != nil {
		t.Fatal(err)
	}
	now = start.Add(30 * time.Minute)
	mid, _, err := store.Get("char-1", "delve")
	if err != nil || mid.Remaining != 2 {
		t.Fatalf("mid interval %+v err=%v", mid, err)
	}
	now = start.Add(time.Hour)
	next, _, err := store.Get("char-1", "delve")
	if err != nil || next.Remaining != 3 {
		t.Fatalf("next bucket %+v err=%v", next, err)
	}
	if mid.Period == next.Period {
		t.Fatalf("period did not advance: %s", next.Period)
	}
}

func TestExhaustLeavesBalanceAndCharactersAreSeparate(t *testing.T) {
	store := openStore(t)
	store.Configure([]Allowance{{Key: "delve", Amount: 1, Reset: ResetCalendar}})
	if _, err := store.Consume("a", "delve", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Consume("a", "delve", 1); !errors.Is(err, ErrExhausted) {
		t.Fatalf("expected exhausted, got %v", err)
	}
	left, _, err := store.Get("a", "delve")
	if err != nil || left.Remaining != 0 {
		t.Fatalf("balance changed on exhaust: %+v err=%v", left, err)
	}
	other, err := store.Consume("b", "delve", 1)
	if err != nil || other.Remaining != 0 {
		t.Fatalf("other character %+v err=%v", other, err)
	}
	if _, err := store.Consume("a", "delve", 0); err != nil {
		t.Fatalf("zero consume: %v", err)
	}
}

func TestMidPeriodConfigAndModifierDoNotRefill(t *testing.T) {
	store := openStore(t)
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	store.SetNow(func() time.Time { return now })
	store.Configure([]Allowance{{Key: "delve", Amount: 5, Reset: ResetCalendar, Timezone: "UTC"}})
	bonus := 0
	store.AddModifier(func(characterID, key string, allowance int) int {
		return allowance + bonus
	})

	first, _, err := store.Get("c", "delve")
	if err != nil || first.Remaining != 5 {
		t.Fatalf("first %+v err=%v", first, err)
	}
	if _, err := store.Consume("c", "delve", 2); err != nil {
		t.Fatal(err)
	}

	bonus = 4
	store.Configure([]Allowance{{Key: "delve", Amount: 9, Reset: ResetCalendar, Timezone: "UTC"}})
	mid, _, err := store.Get("c", "delve")
	if err != nil {
		t.Fatal(err)
	}
	if mid.Remaining != 3 {
		t.Fatalf("remaining changed mid-period: %+v", mid)
	}
	if mid.Allowance != 13 {
		t.Fatalf("displayed allowance = %d, want 13", mid.Allowance)
	}

	now = now.Add(24 * time.Hour)
	next, _, err := store.Get("c", "delve")
	if err != nil || next.Remaining != 13 {
		t.Fatalf("next period %+v err=%v", next, err)
	}
}

func TestModifierClampAndBadInterval(t *testing.T) {
	store := openStore(t)
	store.Configure([]Allowance{{Key: "delve", Amount: 2, Reset: ResetInterval, Interval: time.Millisecond}})
	if _, _, err := store.Get("c", "delve"); err == nil {
		t.Fatal("expected interval error")
	}
	store.Configure([]Allowance{{Key: "delve", Amount: 2, Reset: ResetCalendar}})
	store.AddModifier(func(string, string, int) int { return -3 })
	res, ok, err := store.Get("c", "delve")
	if err != nil || !ok || res.Remaining != 0 {
		t.Fatalf("clamped %+v ok=%v err=%v", res, ok, err)
	}
}

func TestInvalidTimezoneFallsBackToUTC(t *testing.T) {
	store := openStore(t)
	store.SetNow(func() time.Time {
		return time.Date(2026, 9, 26, 23, 30, 0, 0, time.UTC)
	})
	store.Configure([]Allowance{{
		Key: "delve", Amount: 1, Reset: ResetCalendar, Timezone: "Not/AZone",
	}})
	res, _, err := store.Get("c", "delve")
	if err != nil {
		t.Fatal(err)
	}
	if res.Period != "2026-09-26" {
		t.Fatalf("period = %s, want UTC date", res.Period)
	}
}

func TestSchemaTableExistsBeforeStoreUse(t *testing.T) {
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "schema.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	var name string
	err = client.DB().QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='character_resources'`).Scan(&name)
	if err != nil {
		t.Fatalf("schema table: %v", err)
	}
}

func countRows(t *testing.T, store *Store, n *int) error {
	t.Helper()
	return store.db.QueryRow(`SELECT COUNT(*) FROM character_resources`).Scan(n)
}
