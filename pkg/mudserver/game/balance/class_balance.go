package balance

import (
	"math"
	"strings"

	"github.com/talesmud/talesmud/pkg/classkit"
)

// BehindDealtCap is the uphill damage cap. Live rows sit on it.
// Older rogue 2.35 and warrior 1.20 must not come back, even from YAML.
const BehindDealtCap = 1.15

// ClassBalance scales one class's outgoing and incoming damage.
// 1 (or 0, treated as unset) leaves that side unchanged.
// BehindDealt is an extra multiplier on damage_dealt when this class is
// the lower level. It is capped at BehindDealtCap.
// Swings is how many basic attacks one action takes. 0 means 1.
// Keys are class ids. The catalog wins. The config file is the fallback
// when a class is not in the catalog. "wizard" falls back as "mage".
// "ranger" and "hunter" fall back as "rogue". "hitch" falls back as "ward".
type ClassBalance struct {
	DamageDealt float64 `yaml:"damage_dealt"`
	DamageTaken float64 `yaml:"damage_taken"`
	BehindDealt float64 `yaml:"behind_dealt"`
	Swings      int     `yaml:"swings"`
}

func fromKit(row classkit.Row) ClassBalance {
	return ClassBalance{
		DamageDealt: row.DamageDealt,
		DamageTaken: row.DamageTaken,
		BehindDealt: row.BehindDealt,
		Swings:      row.Swings,
	}
}

func defaultClassBalance() map[string]ClassBalance {
	raw := classkit.BalanceMap()
	out := make(map[string]ClassBalance, len(raw))
	for id, row := range raw {
		out[id] = fromKit(row)
	}
	return out
}

func classKey(id string) string {
	id = strings.ToLower(strings.TrimSpace(id))
	switch id {
	case "wizard", "mage":
		return "mage"
	case "rogue", "ranger", "hunter":
		return "rogue"
	case "hitch":
		return "ward"
	default:
		return id
	}
}

func lookupClass(id string) (ClassBalance, bool) {
	if row, ok := classkit.Balance(id); ok {
		return fromKit(row), true
	}
	key := classKey(id)
	if key == "" {
		return ClassBalance{}, false
	}
	cfg := GetConfig()
	if cfg != nil && cfg.ClassBalance != nil {
		if row, ok := cfg.ClassBalance[key]; ok {
			return row, true
		}
		// A config written before the plate class still has the hitch row.
		if key == "ward" {
			if row, ok := cfg.ClassBalance["hitch"]; ok {
				return row, true
			}
		}
	}
	return ClassBalance{}, false
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
	return classkit.HPMultiplier(id)
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
// Ranger and hunter share the rogue slip. Guard is a hotbar skill, not a charge.
func SignatureCharges(id string) (brace, slip, pin int) {
	brace, slip = classkit.Charges(id)
	return brace, slip, 0
}

// GritCap is the Ward soak stack. It lasts the current fight and clears when combat ends.
const GritCap = 5

// WardOpeningGrit is granted on entering combat so the first Slam and the first
// bite are not the empty 0 stack. The cap, and the peak, are still GritCap.
const WardOpeningGrit = 1

// WardStarterSwing is added to a level-1 basic swing only.
// Sword 5 + STR 12 at 0.95 rounds to 6, and 6 still takes four hits to kill a 20 HP rat.
// +1 makes that swing 7, tying the plate opener without putting the coefficient over 1.
const WardStarterSwing int32 = 1

// IsWard reports a class that uses the soak primitive, including a stored hitch id.
func IsWard(id string) bool {
	return classkit.HasGrit(id)
}

func gritBlock() *classkit.Grit {
	if g := classkit.GritOf("ward"); g != nil {
		return g
	}
	return classkit.GritOf("hitch")
}

func clampGrit(grit int) int {
	cap := GritCap
	if g := gritBlock(); g != nil {
		cap = g.Cap
	}
	if grit < 0 {
		return 0
	}
	if grit > cap {
		return cap
	}
	return grit
}

// WardSlamAbsolute is the swing multiplier versus a 1.00 baseline.
// 0 Grit is 1.00×. 2 Grit is 1.40×. 5 Grit is 2.00×.
func WardSlamAbsolute(grit int) float64 {
	per := 0.20
	var slamCap float64
	hasCap := false
	if g := gritBlock(); g != nil {
		per = g.SlamPerStack
		slamCap = g.SlamCap
		hasCap = true
	}
	abs := 1.0 + per*float64(clampGrit(grit))
	if hasCap && abs > slamCap {
		abs = slamCap
	}
	return abs
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
	per := 0.10
	if g := gritBlock(); g != nil {
		per = g.RetaliatePerStack
	}
	back := int32(math.Round(float64(hit) * per * float64(grit)))
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
		if g := gritBlock(); g != nil {
			gain = g.SelfGuardGain
		}
	}
	return clampGrit(grit + gain)
}

// OpeningGrit is granted when a soak class enters combat.
func OpeningGrit(id string) int {
	if g := classkit.GritOf(id); g != nil {
		return g.Opening
	}
	return 0
}

// StarterSwing is added to a level-1 basic swing for a soak class.
func StarterSwing(id string) int32 {
	if g := classkit.GritOf(id); g != nil {
		return g.StarterSwing
	}
	return 0
}

// BoltRigCharges is bolt and rig uses granted at combat start.
func BoltRigCharges(id string) (bolt, rig int) {
	return classkit.BoltRig(id)
}

// ArmsScrap reports a class that starts a fight with bolt or rig uses.
func ArmsScrap(id string) bool {
	bolt, rig := classkit.BoltRig(id)
	return bolt > 0 || rig > 0
}

// IsRuneHand reports the class that inscribes on a basic hit.
func IsRuneHand(id string) bool {
	return classkit.Inscribes(id)
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
