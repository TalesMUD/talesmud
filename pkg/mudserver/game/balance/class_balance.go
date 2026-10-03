package balance

import (
	"math"
	"strings"
)

// BehindDealtCap is the uphill damage cap. Live rows sit on it.
// Older rogue 2.35 and warrior 1.20 must not come back, even from YAML.
const BehindDealtCap = 1.15

// ClassBalance scales one class's outgoing and incoming damage.
// 1 (or 0, treated as unset) leaves that side unchanged.
// BehindDealt is an extra multiplier on damage_dealt when this class is
// the lower level. It is capped at BehindDealtCap.
// Swings is how many basic attacks one action takes. 0 means 1.
// Keys are class ids. "wizard" is read as "mage".
// "ranger" and "hunter" are read as "rogue" (Alley weapons, not classes).
type ClassBalance struct {
	DamageDealt float64 `yaml:"damage_dealt"`
	DamageTaken float64 `yaml:"damage_taken"`
	BehindDealt float64 `yaml:"behind_dealt"`
	Swings      int     `yaml:"swings"`
}

func defaultClassBalance() map[string]ClassBalance {
	return map[string]ClassBalance{
		// Fenwatch. One swing at full. Incoming is lighter.
		"warrior": {DamageDealt: 1.00, DamageTaken: 0.90, BehindDealt: BehindDealtCap, Swings: 1},
		// Alley. Two lighter swings. Incoming hurts more. Ranger and hunter use this row.
		"rogue": {DamageDealt: 0.55, DamageTaken: 1.15, BehindDealt: BehindDealtCap, Swings: 2},
		// Rune Hand. One heavy swing. Cloth takes more. Inscribe is applied in combat, not here.
		"mage": {DamageDealt: 1.40, DamageTaken: 1.25, BehindDealt: BehindDealtCap, Swings: 1},
		// Hitch. One slightly light swing. Pin is applied in combat, not here.
		"hitch": {DamageDealt: 0.90, DamageTaken: 1.05, BehindDealt: BehindDealtCap, Swings: 1},
	}
}

func classKey(id string) string {
	id = strings.ToLower(strings.TrimSpace(id))
	switch id {
	case "wizard", "mage", "runehand", "rune_hand", "rune hand":
		return "mage"
	case "rogue", "alley", "ranger", "hunter":
		return "rogue"
	case "warrior", "fenwatch":
		return "warrior"
	case "hitch":
		return "hitch"
	default:
		return id
	}
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

func capBehind(v float64) float64 {
	v = multOrOne(v)
	if v > BehindDealtCap {
		return BehindDealtCap
	}
	return v
}

// ClassSwings is the basic-attack count for one action. Unknown classes swing once.
func ClassSwings(id string) int {
	row, ok := lookupClass(id)
	if ok && row.Swings > 0 {
		return row.Swings
	}
	if d, ok := defaultClassBalance()[classKey(id)]; ok && d.Swings > 0 {
		return d.Swings
	}
	return 1
}

// ClassHPMultiplier is the create-time max HP scale. 1 leaves the base alone.
// Applied once when a template is built. Combat must not apply it again.
func ClassHPMultiplier(id string) float64 {
	switch classKey(id) {
	case "warrior":
		return 1.20
	case "rogue":
		return 0.85
	case "mage":
		return 0.75
	case "hitch":
		return 1.05
	default:
		return 1
	}
}

// ScaleClassHP applies ClassHPMultiplier once. base <= 0 is unchanged.
func ScaleClassHP(id string, base int32) int32 {
	if base <= 0 {
		return base
	}
	out := int32(math.Round(float64(base) * ClassHPMultiplier(id)))
	if out < 1 {
		return 1
	}
	return out
}

// SignatureCharges is Brace, Slip, Pin uses granted at combat start.
// Ranger and hunter fold into Alley's Slip.
func SignatureCharges(id string) (brace, slip, pin int) {
	switch classKey(id) {
	case "warrior":
		return 1, 0, 0
	case "rogue":
		return 0, 1, 0
	case "hitch":
		return 0, 0, 1
	default:
		return 0, 0, 0
	}
}

// IsRuneHand reports the wizard/mage class that inscribes on a basic hit.
func IsRuneHand(id string) bool {
	return classKey(id) == "mage"
}

// ScaleClassDamage applies the attacker's damage_dealt and the defender's
// damage_taken. When the attacker is lower level, damage_dealt is also
// multiplied by behind_dealt, capped at BehindDealtCap. An empty class id,
// or a multiplier of 1, leaves that side alone. The result stays at least 1
// when damage is positive.
func ScaleClassDamage(attackerClass, defenderClass string, attackerLevel, defenderLevel, damage int32) int32 {
	if damage <= 0 {
		return damage
	}
	dealt, taken := 1.0, 1.0
	if row, ok := lookupClass(attackerClass); ok {
		dealt = multOrOne(row.DamageDealt)
		if attackerLevel < defenderLevel {
			dealt *= capBehind(row.BehindDealt)
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
