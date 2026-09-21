package daily

import (
	"path/filepath"
	"testing"
	"time"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
)

func TestConsumeAndMidnightRollover(t *testing.T) {
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "daily.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close() })

	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}
	store := New(client.DB(), loc)
	day := time.Date(2026, 9, 21, 23, 30, 0, 0, loc)
	store.SetNow(func() time.Time { return day })

	got, err := store.Consume("char-1", ForestFightsKey, 15, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got.Remaining != 14 || got.Day != "2026-09-21" {
		t.Fatalf("after consume: %+v", got)
	}
	if _, err := store.Consume("char-1", ForestFightsKey, 15, 14); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Consume("char-1", ForestFightsKey, 15, 1); err != ErrExhausted {
		t.Fatalf("expected exhausted, got %v", err)
	}

	inn, err := store.Get("char-1", "inn_draught", 1)
	if err != nil {
		t.Fatal(err)
	}
	if inn.Remaining != 1 {
		t.Fatalf("other key was coupled to fights: %+v", inn)
	}

	day = time.Date(2026, 9, 22, 0, 5, 0, 0, loc)
	refilled, err := store.Get("char-1", ForestFightsKey, 15)
	if err != nil {
		t.Fatal(err)
	}
	if refilled.Remaining != 15 || refilled.Day != "2026-09-22" {
		t.Fatalf("rollover: %+v", refilled)
	}
}
