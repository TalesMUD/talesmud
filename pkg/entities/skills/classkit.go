package skills

import (
	"sort"
	"strings"

	"github.com/talesmud/talesmud/pkg/classkit"
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

// IsKitClass reports a catalog class with a hotbar cap.
// Shared rows such as ranger stay off the kit.
func IsKitClass(classID string) bool {
	return classkit.HotbarCap(classID) > 0
}

// ClassKit is the kit loaded from the class catalog.
func ClassKit() []*Skill {
	specs := classkit.Skills()
	out := make([]*Skill, 0, len(specs))
	for _, spec := range specs {
		s := kit(spec.ID, spec.Name, spec.ClassID, spec.Level, spec.Effect, spec.KeepsSwing, spec.OncePerFight, spec.Cooldown, spec.Multiplier, spec.Tooltip)
		s.Target = skillTarget(spec.Target)
		out = append(out, s)
	}
	return out
}

func skillTarget(target string) TargetType {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "ally":
		return TargetAlly
	case "self":
		return TargetSelf
	case "all_enemies":
		return TargetAllEnemies
	default:
		return TargetEnemy
	}
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
