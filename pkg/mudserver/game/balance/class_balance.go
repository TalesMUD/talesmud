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
		// Ward. One slow swing until Grit stacks. Guard and Slam are applied in combat.
		// 0.95 keeps the opener under Fenwatch (starter sword is 6 vs 7) and off the old 5.
		// hitch is the stored id from before this kit and uses the same row.
		"ward": {DamageDealt: 0.95, DamageTaken: 1.05, BehindDealt: BehindDealtCap, Swings: 1},
		// Rigger. One light swing. Bolt and Rig are applied in combat, not here.
		"rigger": {DamageDealt: 0.85, DamageTaken: 1.00, BehindDealt: BehindDealtCap, Swings: 1},
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
	case "hitch", "ward":
		return "ward"
	case "rigger":
		return "rigger"
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
		// A config written before Ward still has the hitch row.
		if key == "ward" {
			if row, ok := cfg.ClassBalance["hitch"]; ok {
				return row, true
			}
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
	case "ward":
		return 1.05
	case "rigger":
		return 1
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

// SignatureCharges is Brace and Slip uses granted at combat start.
// Ranger and hunter fold into Alley's Slip. Ward's Guard is a hotbar skill, not a charge.
func SignatureCharges(id string) (brace, slip, pin int) {
	switch classKey(id) {
	case "warrior":
		return 1, 0, 0
	case "rogue":
		return 0, 1, 0
	default:
		return 0, 0, 0
	}
}

// GritCap is the Ward soak stack. It lasts the current fight and clears when combat ends.
const GritCap = 5

// WardOpeningGrit is granted on entering combat so the first Slam and the first
// bite are not the empty 0 stack. The cap, and the peak, are still GritCap.
const WardOpeningGrit = 1

// WardStarterSwing is added to a level-1 basic swing only.
// Sword 5 + STR 12 at 0.95 rounds to 6, and 6 still takes four hits to kill a 20 HP rat.
// +1 makes that swing 7, tying Fenwatch's starter hit without putting the coefficient over 1.
const WardStarterSwing int32 = 1

// IsWard reports Ward, including characters still stored as hitch.
func IsWard(id string) bool {
	return classKey(id) == "ward"
}

func clampGrit(grit int) int {
	if grit < 0 {
		return 0
	}
	if grit > GritCap {
		return GritCap
	}
	return grit
}

// WardSlamAbsolute is the swing multiplier versus a 1.00 baseline.
// 0 Grit is 1.00×. 2 Grit is 1.40×. 5 Grit is 2.00×.
func WardSlamAbsolute(grit int) float64 {
	return 1.0 + 0.20*float64(clampGrit(grit))
}

// WardSlamSwingMult undoes the class swing so Slam lands on WardSlamAbsolute.
func WardSlamSwingMult(grit int) float64 {
	dealt := 0.95
	if row, ok := lookupClass("ward"); ok && row.DamageDealt > 0 {
		dealt = row.DamageDealt
	}
	return WardSlamAbsolute(grit) / dealt
}

// WardRetaliateDamage is 10% × Grit of the hit, using Grit from before this hit.
// 0 Grit throws nothing back. 5 Grit throws half the hit back.
func WardRetaliateDamage(hit int32, grit int) int32 {
	grit = clampGrit(grit)
	if hit <= 0 || grit <= 0 {
		return 0
	}
	back := int32(math.Round(float64(hit) * 0.10 * float64(grit)))
	if back < 0 {
		return 0
	}
	return back
}

// WardGritAfter adds one stack, or two when the hit was a self-guard soak. Cap is GritCap.
func WardGritAfter(grit int, selfSoak bool) int {
	gain := 1
	if selfSoak {
		gain = 2
	}
	return clampGrit(grit + gain)
}

// RiggerCharges is Bolt and Rig uses granted at combat start. Once each.
func RiggerCharges(id string) (bolt, rig int) {
	if classKey(id) == "rigger" {
		return 1, 1
	}
	return 0, 0
}

// IsRigger reports the rigger class. It is not folded into another row.
func IsRigger(id string) bool {
	return classKey(id) == "rigger"
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
