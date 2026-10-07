package combat

import (
	"testing"

	entities "github.com/talesmud/talesmud/pkg/entities"
	combatent "github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
)

func TestEnemySwingCountAndHolds(t *testing.T) {
	if EnemySwingCount(0) != 1 || EnemySwingCount(-2) != 1 || EnemySwingCount(1) != 1 {
		t.Fatalf("unset, 0, and 1.0 are one swing: 0=%d 1=%d", EnemySwingCount(0), EnemySwingCount(1))
	}
	if EnemySwingCount(2) != 2 || EnemySwingCount(2.4) != 2 || EnemySwingCount(2.6) != 3 || EnemySwingCount(9) != 3 {
		t.Fatalf("round and clamp swings: 2=%d 2.4=%d 2.6=%d 9=%d", EnemySwingCount(2), EnemySwingCount(2.4), EnemySwingCount(2.6), EnemySwingCount(9))
	}
	if EnemySwingCount(0.5) != 1 || EnemySwingCount(0.25) != 1 || EnemySwingCount(0.1) != 1 {
		t.Fatalf("slow enemies still swing once per action: 0.5=%d 0.25=%d 0.1=%d", EnemySwingCount(0.5), EnemySwingCount(0.25), EnemySwingCount(0.1))
	}
	if EnemyHoldsAttack(0, 0) || EnemyHoldsAttack(0, 3) || EnemyHoldsAttack(1, 4) || EnemyHoldsAttack(2, 1) {
		t.Fatal("unset, 1.0, and faster never hold")
	}
	if EnemyHoldsAttack(0.5, 0) || !EnemyHoldsAttack(0.5, 1) || EnemyHoldsAttack(0.5, 2) {
		t.Fatal("0.5 swings, holds, swings")
	}
	if EnemyHoldsAttack(0.25, 0) || !EnemyHoldsAttack(0.25, 1) || !EnemyHoldsAttack(0.25, 3) || EnemyHoldsAttack(0.25, 4) {
		t.Fatal("0.25 swings every fourth attack action")
	}
	if EnemyHoldsAttack(0.8, 0) || !EnemyHoldsAttack(0.8, 1) || EnemyHoldsAttack(0.8, 2) {
		t.Fatal("0.8 clamps to a period of 2")
	}
}

func TestNPCAttackSpeedSwingsOncePerAction(t *testing.T) {
	mgr := NewManager()
	e := NewEngine(mgr, nil)
	inst := mgr.CreateInstance("room")
	inst.State = combatent.CombatStateActive
	inst.Players = []combatent.CombatantRef{{
		ID: "hero", Type: combatent.CombatantTypePlayer, Name: "Hero",
		IsAlive: true, MaxHP: 500, CurrentHP: 500, Defense: 0,
		StatusEffects: []combatent.StatusEffect{{Stat: "dodge", Percent: 1}},
	}}
	inst.Enemies = []combatent.CombatantRef{{
		ID: "brute", Type: combatent.CombatantTypeNPC, Name: "Brute",
		IsAlive: true, MaxHP: 80, CurrentHP: 80, AttackPower: 5, AttackSpeed: 2, Difficulty: "easy",
	}}

	e.ProcessAttack(inst, "brute", "hero")
	if got := inst.GetEnemyByID("brute").AttackActions; got != 1 {
		t.Fatalf("two swings are one attack action, actions=%d", got)
	}
	if dodges := countResult(inst, "brute", "dodged"); dodges != 2 {
		t.Fatalf("speed 2 wants two swings, dodged=%d", dodges)
	}

	inst.Enemies[0].AttackSpeed = 0
	inst.Enemies[0].AttackActions = 0
	inst.Log = nil
	e.ProcessAttack(inst, "brute", "hero")
	if got := inst.GetEnemyByID("brute").AttackActions; got != 1 {
		t.Fatalf("speed 0 still counts one action, actions=%d", got)
	}
	if dodges := countResult(inst, "brute", "dodged"); dodges != 1 {
		t.Fatalf("speed 0 wants one swing, dodged=%d", dodges)
	}
}

func TestCreateCombatantFromNPCCopiesHooksAndSpeed(t *testing.T) {
	e := NewEngine(NewManager(), nil)
	n := &npc.NPC{
		Entity: &entities.Entity{ID: "chan"},
		Name:   "Chanter",
		EnemyTrait: &npc.EnemyTrait{
			AttackSpeed: 2, Difficulty: "easy", AttackPower: 3,
			OnAggroScript: "SCR-A", OnDeathScript: "SCR-D", OnFleeScript: "SCR-F",
		},
	}
	ref := e.CreateCombatantFromNPC(n)
	if ref.AttackSpeed != 2 || ref.OnAggroScript != "SCR-A" || ref.OnDeathScript != "SCR-D" || ref.OnFleeScript != "SCR-F" {
		t.Fatalf("snapshot = %+v", ref)
	}
	plain := e.CreateCombatantFromNPC(&npc.NPC{Entity: &entities.Entity{ID: "rat"}, Name: "Rat"})
	if plain.AttackSpeed != 0 || plain.OnAggroScript != "" {
		t.Fatalf("unset speed must stay 0, got %+v", plain)
	}
}

func countResult(inst *combatent.CombatInstance, actor, result string) int {
	n := 0
	for _, entry := range inst.Log {
		if entry.ActorID == actor && entry.Result == result {
			n++
		}
	}
	return n
}
