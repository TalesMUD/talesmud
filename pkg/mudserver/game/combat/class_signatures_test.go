package combat

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	entcombat "github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
)

func newFight(class characters.Class, mana int32) (*Engine, *entcombat.CombatInstance, string, string) {
	e := NewEngine(NewManager(), DefaultConfig())
	hero := &characters.Character{
		Entity:           entities.NewEntity(),
		Name:             "Hero",
		Level:            5,
		Class:            class,
		MaxHitPoints:     400,
		CurrentHitPoints: 400,
		MaxMana:          mana,
		CurrentMana:      mana,
	}
	enemy := &npc.NPC{
		Entity:           entities.NewEntity(),
		Name:             "Dummy",
		Level:            5,
		MaxHitPoints:     5000,
		CurrentHitPoints: 5000,
		EnemyTrait:       &npc.EnemyTrait{AttackPower: 40, Defense: 0, Difficulty: "normal"},
	}
	inst := e.InitiateCombat("sig", []*characters.Character{hero}, []*npc.NPC{enemy})
	return e, inst, hero.Entity.ID, enemy.Entity.ID
}

func TestAlleySwingsTwice(t *testing.T) {
	e, inst, heroID, enemyID := newFight(characters.ClassRogue, 0)
	res := e.ProcessAttack(inst, heroID, enemyID)
	if strings.Count(res.Message, "Hero") < 2 {
		t.Fatalf("expected two alley swings, got %q", res.Message)
	}
}

func TestFenwatchBraceOnce(t *testing.T) {
	e, inst, heroID, enemyID := newFight(characters.ClassWarrior, 0)
	var braced bool
	for i := 0; i < 40; i++ {
		res := e.ProcessAttack(inst, enemyID, heroID)
		if strings.Contains(res.Message, "You brace.") {
			braced = true
			break
		}
	}
	if !braced {
		t.Fatal("expected one brace")
	}
	hero := inst.GetCombatantByID(heroID)
	if hero.BraceLeft != 0 {
		t.Fatalf("brace left %d", hero.BraceLeft)
	}
	res := e.ProcessAttack(inst, enemyID, heroID)
	if strings.Contains(res.Message, "You brace.") {
		t.Fatalf("brace fired twice: %q", res.Message)
	}
}

func TestAlleySlipOnce(t *testing.T) {
	e, inst, heroID, enemyID := newFight(characters.ClassRogue, 0)
	res := e.ProcessAttack(inst, enemyID, heroID)
	if res.Message != "You slip the blow." {
		t.Fatalf("slip message %q", res.Message)
	}
	res = e.ProcessAttack(inst, enemyID, heroID)
	if strings.Contains(res.Message, "You slip the blow.") {
		t.Fatal("slip fired twice")
	}
}

func TestRuneHandInscribeRefreshNoMana(t *testing.T) {
	e, inst, heroID, enemyID := newFight(characters.ClassWizard, 11)
	var landed int
	for i := 0; i < 50 && landed < 2; i++ {
		res := e.ProcessAttack(inst, heroID, enemyID)
		if res.Hit {
			landed++
			if !strings.Contains(res.Message, "You inscribe them.") {
				t.Fatalf("missing inscribe: %q", res.Message)
			}
		}
	}
	if landed < 2 {
		t.Fatal("could not land two rune hand hits")
	}
	hero := inst.GetCombatantByID(heroID)
	if hero.CurrentMana != 11 {
		t.Fatalf("basic spent mana, left %d", hero.CurrentMana)
	}
	target := inst.GetCombatantByID(enemyID)
	var dots int
	for _, se := range target.StatusEffects {
		if se.SkillID == "inscribe" {
			dots++
			if se.Value != 4 || se.Duration != 3 || se.Type != "dot" || se.Name != "Inscribe" {
				t.Fatalf("inscribe effect %+v", se)
			}
		}
	}
	if dots != 1 {
		t.Fatalf("inscribe stacked: %d", dots)
	}
}

func TestHitchPinCancelsNextLeave(t *testing.T) {
	e, inst, heroID, enemyID := newFight(characters.ClassHitch, 0)
	if !e.ArmPin(inst, heroID, enemyID) {
		t.Fatal("ArmPin should arm once")
	}
	if e.ArmPin(inst, heroID, enemyID) {
		t.Fatal("Pin armed twice")
	}
	flee := e.ProcessFlee(inst, enemyID)
	if flee.Success || flee.Message != "You hitch them. They stay." {
		t.Fatalf("pin flee %+v", flee)
	}
	flee = e.ProcessFlee(inst, enemyID)
	if flee.Message == "You hitch them. They stay." {
		t.Fatal("pin consumed but still cancelled the next leave")
	}
}
