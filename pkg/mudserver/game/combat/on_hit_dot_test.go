package combat

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	entcombat "github.com/talesmud/talesmud/pkg/entities/combat"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
)

func TestWeaponOnHitDotAppliesAndRefreshes(t *testing.T) {
	e := NewEngine(NewManager(), DefaultConfig())
	blade := &items.Item{
		Entity: &entities.Entity{ID: "ITM0265"},
		Name:   "Unmarked Vigil Blade",
		Slot:   items.ItemSlotMainHand,
		Attributes: map[string]interface{}{
			"damage": 15,
			"on_hit_dot": map[string]interface{}{
				"id":       "vigil_burn",
				"name":     "Vigil Burn",
				"damage":   2,
				"duration": 3,
				"message":  "Faint blue runes pulse — Vigil Burn takes hold.",
			},
		},
	}
	hero := &characters.Character{
		Entity:           entities.NewEntity(),
		Name:             "Gimli",
		Level:            6,
		MaxHitPoints:     200,
		CurrentHitPoints: 200,
		EquippedItems: map[items.ItemSlot]*items.Item{
			items.ItemSlotMainHand: blade,
		},
	}
	enemy := &npc.NPC{
		Entity:           entities.NewEntity(),
		Name:             "Dummy",
		Level:            1,
		MaxHitPoints:     500,
		CurrentHitPoints: 500,
		EnemyTrait:       &npc.EnemyTrait{AttackPower: 1, Defense: 0, Difficulty: "normal"},
	}
	inst := e.InitiateCombat("onhit", []*characters.Character{hero}, []*npc.NPC{enemy})
	attacker := inst.GetCombatantByID(hero.Entity.ID)
	if attacker == nil || !attacker.OnHitDot.Active {
		t.Fatalf("expected OnHitDot snapshot, got %+v", attacker)
	}
	if attacker.OnHitDot.Damage != 2 || attacker.OnHitDot.Duration != 3 {
		t.Fatalf("unexpected OnHitDot: %+v", attacker.OnHitDot)
	}

	// Force a hit by attacking until we land one (AC is low).
	var landed bool
	for i := 0; i < 40; i++ {
		res := e.ProcessAttack(inst, hero.Entity.ID, enemy.Entity.ID)
		if res.Hit {
			landed = true
			if res.Message == "" || !contains(res.Message, "Vigil Burn") {
				t.Fatalf("expected Vigil Burn message in %q", res.Message)
			}
			break
		}
	}
	if !landed {
		t.Fatal("failed to land a hit for on-hit DoT test")
	}
	target := inst.GetCombatantByID(enemy.Entity.ID)
	if target == nil || len(target.StatusEffects) == 0 {
		t.Fatalf("expected DoT on target, got %+v", target)
	}
	se := target.StatusEffects[0]
	if se.Type != "dot" || se.Value != 2 || se.Duration != 3 || se.Name != "Vigil Burn" {
		t.Fatalf("unexpected status effect: %+v", se)
	}
	// Reapply refreshes duration (no stack).
	target.StatusEffects[0].Duration = 1
	e.UpdateCombatant(inst, target)
	for i := 0; i < 40; i++ {
		res := e.ProcessAttack(inst, hero.Entity.ID, enemy.Entity.ID)
		if res.Hit {
			break
		}
	}
	target = inst.GetCombatantByID(enemy.Entity.ID)
	if len(target.StatusEffects) != 1 {
		t.Fatalf("expected single refreshed DoT, got %d: %+v", len(target.StatusEffects), target.StatusEffects)
	}
	if target.StatusEffects[0].Duration != 3 {
		t.Fatalf("expected refreshed duration 3, got %d", target.StatusEffects[0].Duration)
	}
	_ = entcombat.CombatActionAttack
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(func() bool {
			for i := 0; i+len(sub) <= len(s); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		})())
}
