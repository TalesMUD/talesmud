package balance

import (
	"math"
	"strings"
)

// ClassBalance scales one class's outgoing and incoming damage.
// 1 (or 0, treated as unset) leaves that side unchanged.
// BehindDealt is an extra multiplier on damage_dealt when this class is
// the lower level, so a fight uphill can be tuned apart from an even fight.
// Keys are class ids. "wizard" is read as "mage".
type ClassBalance struct {
	DamageDealt float64 `yaml:"damage_dealt"`
	DamageTaken float64 `yaml:"damage_taken"`
	BehindDealt float64 `yaml:"behind_dealt"`
}

func defaultClassBalance() map[string]ClassBalance {
	return map[string]ClassBalance{
		// Even fights stay at 1 so at-level trash duration is unchanged.
		// BehindDealt is only the uphill fight. The thicker scaled boss body
		// is what pulls an even boss down into the 50–65% band.
		"warrior": {DamageDealt: 1, DamageTaken: 1, BehindDealt: 1.20},
		// Even fights stay near the old dagger. BehindDealt is the +3 boss.
		"rogue":  {DamageDealt: 1.35, DamageTaken: 1, BehindDealt: 2.35},
		"ranger": {DamageDealt: 1.26, DamageTaken: 1},
		// Cloth stays fragile. Spell hits have to land the kill before the robe does.
		"mage": {DamageDealt: 2.65, DamageTaken: 0.46},
	}
}

func classKey(id string) string {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "wizard" {
		return "mage"
	}
	return id
}

func lookupClass(id string) (ClassBalance, bool) {
	key := classKey(id)
	if key == "" {
		return ClassBalance{}, false
	}
	cfg := GetConfig()
	if cfg != nil && cfg.ClassBalance != nil {
		if row, ok := cfg.ClassBalance[key]; ok {
			return row, true
		}
	}
	row, ok := defaultClassBalance()[key]
	return row, ok
}

func multOrOne(v float64) float64 {
	if v <= 0 {
		return 1
	}
	return v
}

// ScaleClassDamage applies the attacker's damage_dealt and the defender's
// damage_taken. When the attacker is lower level, damage_dealt is also
// multiplied by behind_dealt. An empty class id, or a multiplier of 1,
// leaves that side alone. The result stays at least 1 when damage is positive.
func ScaleClassDamage(attackerClass, defenderClass string, attackerLevel, defenderLevel, damage int32) int32 {
	if damage <= 0 {
		return damage
	}
	dealt, taken := 1.0, 1.0
	if row, ok := lookupClass(attackerClass); ok {
		dealt = multOrOne(row.DamageDealt)
		if attackerLevel < defenderLevel {
			dealt *= multOrOne(row.BehindDealt)
		}
	}
	if row, ok := lookupClass(defenderClass); ok {
		taken = multOrOne(row.DamageTaken)
	}
	if dealt == 1 && taken == 1 {
		return damage
	}
	out := int32(math.Round(float64(damage) * dealt * taken))
	if out < 1 {
		return 1
	}
	return out
}
