package quests

import (
	"testing"
	"time"
)

func TestDailyResetBoundary(t *testing.T) {
	q := &Quest{Repeatable: true, RepeatReset: RepeatResetDaily}
	hour := 4
	done := time.Date(2026, 10, 17, 3, 0, 0, 0, time.UTC)
	next, ok := q.NextAvailableAt(done, hour)
	if !ok || !next.Equal(time.Date(2026, 10, 17, 4, 0, 0, 0, time.UTC)) {
		t.Fatalf("before the hour: %v ok=%v", next, ok)
	}
	done = time.Date(2026, 10, 17, 4, 0, 0, 0, time.UTC)
	next, ok = q.NextAvailableAt(done, hour)
	if !ok || !next.Equal(time.Date(2026, 10, 18, 4, 0, 0, 0, time.UTC)) {
		t.Fatalf("on the hour: %v", next)
	}
	done = time.Date(2026, 10, 17, 5, 12, 0, 0, time.UTC)
	next, ok = q.NextAvailableAt(done, hour)
	if !ok || !next.Equal(time.Date(2026, 10, 18, 4, 0, 0, 0, time.UTC)) {
		t.Fatalf("after the hour: %v", next)
	}
}

func TestFormatAvailableAgain(t *testing.T) {
	now := time.Date(2026, 10, 17, 10, 0, 0, 0, time.UTC)
	until := now.Add(5*time.Hour + 12*time.Minute)
	got := FormatAvailableAgain(until, now)
	if got != "Available again in 5h 12m" {
		t.Fatalf("got %q", got)
	}
	got = FormatAvailableAgain(until.Add(time.Second), now)
	if got != "Available again in 5h 13m" {
		t.Fatalf("partial minute got %q", got)
	}
}

func TestCooldownAndPlainRepeatable(t *testing.T) {
	now := time.Date(2026, 10, 17, 10, 0, 0, 0, time.UTC)
	cool := &Quest{Repeatable: true, RepeatReset: RepeatResetCooldown, RepeatCooldownSeconds: 3600}
	until, waiting := cool.AvailableAgain(now, now.Add(59*time.Minute), 4)
	if !waiting || !until.Equal(now.Add(time.Hour)) {
		t.Fatalf("cooldown still open: until=%v waiting=%v", until, waiting)
	}
	_, waiting = cool.AvailableAgain(now, now.Add(time.Hour), 4)
	if waiting {
		t.Fatal("cooldown should be open at the instant it ends")
	}
	plain := &Quest{Repeatable: true}
	if plain.Repeats() || plain.RepeatTag() != "" {
		t.Fatal("plain repeatable grew a window")
	}
	if _, waiting := plain.AvailableAgain(now, now, 4); waiting {
		t.Fatal("plain repeatable is waiting")
	}
	blocked := &Quest{Repeatable: false, RepeatReset: RepeatResetDaily}
	if blocked.Repeats() {
		t.Fatal("non-repeatable reset counted")
	}
}

func TestRepeatPayloadWhileWaiting(t *testing.T) {
	q := &Quest{Repeatable: true, RepeatReset: RepeatResetDaily}
	done := time.Date(2026, 10, 17, 10, 0, 0, 0, time.UTC)
	now := done
	p := q.RepeatPayload(string(QuestStatusCompleted), done, now, 4)
	if !p.Repeats || p.RepeatReset != RepeatResetDaily || p.NextAvailableAt != "2026-10-18T04:00:00Z" {
		t.Fatalf("payload %+v", p)
	}
	active := q.RepeatPayload(string(QuestStatusActive), done, now, 4)
	if !active.Repeats || active.NextAvailableAt != "" {
		t.Fatalf("active payload %+v", active)
	}
}
