package combat

import (
	"testing"

	entcombat "github.com/talesmud/talesmud/pkg/entities/combat"
)

func livingRef(id, name string, kind entcombat.CombatantType, init int) entcombat.CombatantRef {
	return entcombat.CombatantRef{
		ID: id, Name: name, Type: kind, Initiative: init,
		IsAlive: true, MaxHP: 20, CurrentHP: 20,
	}
}

func assertIndexInOrder(t *testing.T, inst *entcombat.CombatInstance) {
	t.Helper()
	if len(inst.TurnOrder) == 0 {
		if inst.CurrentTurnIdx != 0 {
			t.Fatalf("empty order index = %d", inst.CurrentTurnIdx)
		}
		return
	}
	if inst.CurrentTurnIdx < 0 || inst.CurrentTurnIdx >= len(inst.TurnOrder) {
		t.Fatalf("index %d past order len %d", inst.CurrentTurnIdx, len(inst.TurnOrder))
	}
}

func TestNextTurnDeadTailOpensNextRound(t *testing.T) {
	e := NewEngine(NewManager(), DefaultConfig())
	inst := entcombat.NewCombatInstance("R")
	inst.State = entcombat.CombatStateActive
	inst.Round = 4
	add := livingRef("add", "Add", entcombat.CombatantTypeNPC, 30)
	hero := livingRef("hero", "Hero", entcombat.CombatantTypePlayer, 10)
	dead := livingRef("dead", "Dead", entcombat.CombatantTypeNPC, 5)
	dead.IsAlive = false
	dead.CurrentHP = 0
	inst.Players = []entcombat.CombatantRef{hero}
	inst.Enemies = []entcombat.CombatantRef{add, dead}
	inst.TurnOrder = []entcombat.CombatantRef{add, hero, dead}
	inst.CurrentTurnIdx = 1

	next := e.NextTurn(inst)
	assertIndexInOrder(t, inst)
	if next == nil || !next.IsAlive {
		t.Fatal("dead tail did not advance to a living combatant")
	}
	if inst.Round != 5 {
		t.Fatalf("round = %d, want 5", inst.Round)
	}
	if next.ID != "add" {
		t.Fatalf("next = %s, want add", next.ID)
	}
}

func TestNextTurnPastEndSelfHeals(t *testing.T) {
	e := NewEngine(NewManager(), DefaultConfig())
	inst := entcombat.NewCombatInstance("R")
	inst.State = entcombat.CombatStateActive
	inst.Round = 5
	add := livingRef("add", "Add", entcombat.CombatantTypeNPC, 30)
	hero := livingRef("hero", "Hero", entcombat.CombatantTypePlayer, 10)
	inst.Players = []entcombat.CombatantRef{hero}
	inst.Enemies = []entcombat.CombatantRef{add}
	inst.TurnOrder = []entcombat.CombatantRef{add, hero}
	inst.CurrentTurnIdx = len(inst.TurnOrder)

	next := e.NextTurn(inst)
	assertIndexInOrder(t, inst)
	if next == nil || next.ID != "add" {
		t.Fatalf("next = %#v, want add", next)
	}
	if inst.Round != 6 {
		t.Fatalf("round = %d, want 6", inst.Round)
	}
}

func TestNextTurnAdvancesToNextLiving(t *testing.T) {
	e := NewEngine(NewManager(), DefaultConfig())
	inst := entcombat.NewCombatInstance("R")
	inst.State = entcombat.CombatStateActive
	inst.Round = 2
	add := livingRef("add", "Add", entcombat.CombatantTypeNPC, 30)
	hero := livingRef("hero", "Hero", entcombat.CombatantTypePlayer, 10)
	inst.Players = []entcombat.CombatantRef{hero}
	inst.Enemies = []entcombat.CombatantRef{add}
	inst.TurnOrder = []entcombat.CombatantRef{add, hero}
	inst.CurrentTurnIdx = 0

	next := e.NextTurn(inst)
	assertIndexInOrder(t, inst)
	if next == nil || next.ID != "hero" {
		t.Fatalf("next = %#v, want hero", next)
	}
	if inst.Round != 2 {
		t.Fatalf("round = %d, want 2", inst.Round)
	}
}

func TestNextTurnDeadCurrentContinuesRound(t *testing.T) {
	e := NewEngine(NewManager(), DefaultConfig())
	inst := entcombat.NewCombatInstance("R")
	inst.State = entcombat.CombatStateActive
	inst.Round = 3
	dead := livingRef("dead", "Dead", entcombat.CombatantTypeNPC, 30)
	dead.IsAlive = false
	dead.CurrentHP = 0
	hero := livingRef("hero", "Hero", entcombat.CombatantTypePlayer, 10)
	inst.Players = []entcombat.CombatantRef{hero}
	inst.Enemies = []entcombat.CombatantRef{dead}
	inst.TurnOrder = []entcombat.CombatantRef{dead, hero}
	inst.CurrentTurnIdx = 0

	next := e.NextTurn(inst)
	assertIndexInOrder(t, inst)
	if next == nil || next.ID != "hero" {
		t.Fatalf("next = %#v, want hero", next)
	}
	if inst.Round != 3 {
		t.Fatalf("round = %d, want 3", inst.Round)
	}
}
