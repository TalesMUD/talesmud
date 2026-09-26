package ruleset

import (
	"strings"
	"time"

	"github.com/talesmud/talesmud/pkg/entities/characters"
)

// ApplyNewDay fills HP and mana when the profile asks for a dawn heal and the
// calendar day in the profile timezone has changed. It does not touch resource
// balances. A false result means the character was not modified.
func ApplyNewDay(char *characters.Character, now time.Time) bool {
	if char == nil {
		return false
	}
	mu.RLock()
	fullHeal := current.fullHeal
	tz := current.timezone
	mu.RUnlock()
	if !fullHeal {
		return false
	}
	day := now.In(location(tz)).Format("2006-01-02")
	if char.LastResetDay == day {
		return false
	}
	if char.MaxHitPoints > 0 {
		char.CurrentHitPoints = char.MaxHitPoints
	}
	if char.MaxMana > 0 {
		char.CurrentMana = char.MaxMana
	}
	char.AwaitingReset = false
	char.LastResetDay = day
	return true
}

func location(name string) *time.Location {
	name = strings.TrimSpace(name)
	if name == "" || strings.EqualFold(name, "UTC") {
		return time.UTC
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}
