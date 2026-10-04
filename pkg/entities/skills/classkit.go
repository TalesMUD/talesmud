package skills

import (
	"sort"

	"github.com/talesmud/talesmud/pkg/entities"
)

// Class-kit ids. These are the only skills the five signed classes learn.
const (
	KitBrace    = "brace"
	KitSlam     = "slam"
	KitStand    = "stand"
	KitSlip     = "slip"
	KitNick     = "nick"
	KitSmoke    = "smoke"
	KitInscribe = "inscribe"
	KitSear     = "sear"
	KitGlyph    = "glyph"
	KitGuard    = "guard"
	KitBolt     = "bolt"
	KitRig      = "rig"
	KitOverload = "overload"
)

const HotbarCap = 4

// IsKitClass reports Fenwatch, Alley, Rune Hand, Ward, and Rigger.
// The stored class id hitch is Ward.
func IsKitClass(classID string) bool {
	switch normalizeClassID(classID) {
	case "warrior", "rogue", "mage", "ward", "rigger":
		return true
	default:
		return false
	}
}

// ClassKit is the v1 kit. Two later unlocks per class. Rigger's third is Overload at 6.
func ClassKit() []*Skill {
	return []*Skill{
		kit("warrior_brace", "Brace", "warrior", 1, KitBrace, true, true, 0, 0,
			"Once a fight. The next hit on you is halved. You still swing."),
		kit("warrior_slam", "Slam", "warrior", 4, KitSlam, false, false, 4, 1.40,
			"A heavy swing, 1.40×. Replaces your swing this round."),
		kit("warrior_stand", "Stand", "warrior", 8, KitStand, false, true, 0, 0,
			"Once a fight, for two rounds hits aimed at others hit you."),

		kit("rogue_slip", "Slip", "rogue", 1, KitSlip, false, true, 0, 0,
			"Once a fight. Drop the fight and take one exit."),
		kit("rogue_nick", "Nick", "rogue", 4, KitNick, false, false, 3, 0.55,
			"Extra 0.55× swing on top of your cadence. Replaces the autoattack this round."),
		kit("rogue_smoke", "Smoke", "rogue", 8, KitSmoke, false, true, 0, 0,
			"Once a fight. The target misses their next swing."),

		kit("mage_inscribe", "Inscribe", "mage", 1, KitInscribe, false, false, 4, 0,
			"Mark them. 4 a round for 3 rounds. Refresh, no stack. Costs no mana."),
		kit("mage_sear", "Sear", "mage", 4, KitSear, false, false, 4, 1.80,
			"A cast at 1.80×. No burn. Replaces your swing this round."),
		kit("mage_glyph", "Glyph", "mage", 8, KitGlyph, false, true, 0, 0,
			"Once a fight. The next hit on you is reduced by 4."),

		wardGuard(),
		kit("ward_slam", "Slam", "ward", 4, KitSlam, false, false, 4, 1.00,
			"Replaces your swing. 1.00×, plus 0.20× for each Grit, up to 2.00×."),

		kit("rigger_bolt", "Bolt", "rigger", 1, KitBolt, false, true, 0, 0,
			"Once a fight. Spend your swing. The next hit still lands, and the attacker takes it back."),
		kit("rigger_rig", "Rig", "rigger", 1, KitRig, false, true, 0, 0.50,
			"Once a fight. Drop a turret. It hits twice at 0.50×, then falls apart."),
		kit("rigger_overload", "Overload", "rigger", 6, KitOverload, false, true, 0, 0.80,
			"Once a fight, while the turret is up. Its remaining hits deal 0.80×."),
	}
}

func wardGuard() *Skill {
	s := kit("ward_guard", "Guard", "ward", 1, KitGuard, true, true, 0, 0,
		"Once a fight. The next hit aimed at an ally hits you. Guarding yourself, that hit stacks two Grit. You still swing.")
	s.Target = TargetAlly
	return s
}

func kit(id, name, classID string, level int32, kind string, keeps, once bool, cd int, mult float64, desc string) *Skill {
	return &Skill{
		Entity:         &entities.Entity{ID: id},
		Name:           name,
		Description:    desc,
		ClassIDs:       []string{classID},
		LevelRequired:  level,
		ResourceType:   ResourceCooldown,
		ManaCost:       0,
		CooldownRounds: cd,
		Target:         TargetEnemy,
		Effect:         EffectDamage,
		Kit:            kind,
		KeepsSwing:     keeps,
		OncePerFight:   once,
		SwingMult:      mult,
	}
}

// FillHotbar appends known kit skills into empty slots, in unlock order.
// It does not remove a skill the player already chose, and it never passes HotbarCap.
func FillHotbar(classID string, level int32, equipped []string) []string {
	if !IsKitClass(classID) {
		return equipped
	}
	if normalizeClassID(classID) == "ward" {
		equipped = migrateWardHotbar(level, equipped)
	}
	out := make([]string, 0, HotbarCap)
	seen := map[string]bool{}
	for _, id := range equipped {
		if id == "" || seen[id] || len(out) >= HotbarCap {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	known := AvailableSkills(classID, level)
	sort.SliceStable(known, func(i, j int) bool {
		if known[i].LevelRequired != known[j].LevelRequired {
			return known[i].LevelRequired < known[j].LevelRequired
		}
		return known[i].Entity.ID < known[j].Entity.ID
	})
	for _, s := range known {
		if s == nil || s.Kit == "" || s.Entity == nil {
			continue
		}
		if seen[s.Entity.ID] || len(out) >= HotbarCap {
			continue
		}
		seen[s.Entity.ID] = true
		out = append(out, s.Entity.ID)
	}
	return out
}

// migrateWardHotbar drops Pin, Hobble, and Reel and keeps Guard and Slam in unlock order.
// It reads ClassKit directly so a character loaded before the skill cache still keeps a bar.
func migrateWardHotbar(level int32, equipped []string) []string {
	out := make([]string, 0, HotbarCap)
	seen := map[string]bool{}
	for _, id := range equipped {
		switch id {
		case "", "hitch_pin", "hitch_hobble", "hitch_reel":
			continue
		}
		if seen[id] || len(out) >= HotbarCap {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	for _, s := range ClassKit() {
		if s == nil || s.Entity == nil || !s.HasClass("ward") || s.LevelRequired > level {
			continue
		}
		if seen[s.Entity.ID] || len(out) >= HotbarCap {
			continue
		}
		seen[s.Entity.ID] = true
		out = append(out, s.Entity.ID)
	}
	return out
}

// AllowsClass reports whether this skill belongs to classID, including aliases.
func (s *Skill) AllowsClass(classID string) bool {
	if s == nil {
		return false
	}
	return s.HasClass(normalizeClassID(classID))
}
