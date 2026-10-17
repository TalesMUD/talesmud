package quests

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	// RepeatResetDaily waits until the ruleset hour on the next UTC day boundary.
	RepeatResetDaily = "daily"
	// RepeatResetCooldown waits RepeatCooldownSeconds after completion.
	RepeatResetCooldown = "cooldown"
)

var (
	nowMu sync.Mutex
	nowFn = time.Now
)

// Now is the clock quest reset checks use. Tests replace it with SetNow.
func Now() time.Time {
	nowMu.Lock()
	fn := nowFn
	nowMu.Unlock()
	if fn == nil {
		return time.Now()
	}
	return fn()
}

// SetNow replaces the quest clock. A nil function restores time.Now.
func SetNow(fn func() time.Time) {
	nowMu.Lock()
	defer nowMu.Unlock()
	if fn == nil {
		nowFn = time.Now
		return
	}
	nowFn = fn
}

// RepeatKind is "daily", "cooldown", or empty.
// Empty means a repeatable quest can be accepted again immediately.
// A reset is ignored unless the quest is repeatable.
func (q *Quest) RepeatKind() string {
	if q == nil || !q.Repeatable {
		return ""
	}
	switch strings.ToLower(strings.TrimSpace(q.RepeatReset)) {
	case RepeatResetDaily:
		return RepeatResetDaily
	case RepeatResetCooldown:
		if q.RepeatCooldownSeconds > 0 {
			return RepeatResetCooldown
		}
		return ""
	case "":
		if q.RepeatCooldownSeconds > 0 {
			return RepeatResetCooldown
		}
		return ""
	default:
		return ""
	}
}

// Repeats is true when completion opens a timed window before the next accept.
func (q *Quest) Repeats() bool {
	return q.RepeatKind() != ""
}

// RepeatTag is the short label a client can show. Empty when there is no window.
func (q *Quest) RepeatTag() string {
	switch q.RepeatKind() {
	case RepeatResetDaily:
		return "Daily"
	case RepeatResetCooldown:
		return "Repeats"
	default:
		return ""
	}
}

// NextAvailableAt is the UTC instant the quest can be accepted again.
// The bool is false when the quest has no reset window.
func (q *Quest) NextAvailableAt(completedAt time.Time, resetHour int) (time.Time, bool) {
	if !q.Repeats() || completedAt.IsZero() {
		return time.Time{}, false
	}
	switch q.RepeatKind() {
	case RepeatResetDaily:
		return nextDailyReset(completedAt, resetHour), true
	case RepeatResetCooldown:
		return completedAt.UTC().Add(time.Duration(q.RepeatCooldownSeconds) * time.Second), true
	default:
		return time.Time{}, false
	}
}

// AvailableAgain reports whether a completed quest is still inside its window.
func (q *Quest) AvailableAgain(completedAt, now time.Time, resetHour int) (until time.Time, waiting bool) {
	next, ok := q.NextAvailableAt(completedAt, resetHour)
	if !ok {
		return time.Time{}, false
	}
	if !now.Before(next) {
		return next, false
	}
	return next, true
}

// RepeatPayload is the client view of a reset. NextAvailableAt is set only
// while a completed quest is still waiting, as absolute UTC.
type RepeatPayload struct {
	Repeats         bool
	RepeatReset     string
	NextAvailableAt string
}

// RepeatPayload builds the flag and the next-available instant for one character.
func (q *Quest) RepeatPayload(status string, completedAt, now time.Time, resetHour int) RepeatPayload {
	kind := q.RepeatKind()
	if kind == "" {
		return RepeatPayload{}
	}
	out := RepeatPayload{Repeats: true, RepeatReset: kind}
	if strings.EqualFold(strings.TrimSpace(status), string(QuestStatusCompleted)) {
		if until, waiting := q.AvailableAgain(completedAt, now, resetHour); waiting {
			out.NextAvailableAt = until.UTC().Format(time.RFC3339)
		}
	}
	return out
}

// FormatAvailableAgain renders "Available again in 5h 12m".
// A partial minute rounds up so a quest that is still blocked never says 0m.
func FormatAvailableAgain(until, now time.Time) string {
	if !now.Before(until) {
		return ""
	}
	remain := until.Sub(now)
	minutes := int((remain + time.Minute - 1) / time.Minute)
	if minutes < 1 {
		minutes = 1
	}
	days := minutes / (24 * 60)
	hours := (minutes % (24 * 60)) / 60
	mins := minutes % 60
	switch {
	case days > 0 && hours > 0:
		return fmt.Sprintf("Available again in %dd %dh", days, hours)
	case days > 0:
		return fmt.Sprintf("Available again in %dd", days)
	case hours > 0 && mins > 0:
		return fmt.Sprintf("Available again in %dh %dm", hours, mins)
	case hours > 0:
		return fmt.Sprintf("Available again in %dh", hours)
	default:
		return fmt.Sprintf("Available again in %dm", mins)
	}
}

func nextDailyReset(completedAt time.Time, hour int) time.Time {
	if hour < 0 || hour > 23 {
		hour = 4
	}
	t := completedAt.UTC()
	boundary := time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, time.UTC)
	if !t.Before(boundary) {
		boundary = boundary.Add(24 * time.Hour)
	}
	return boundary
}
