package combat

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	entcombat "github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/skills"
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

func equipKit(inst *entcombat.CombatInstance, heroID string, ids ...string) {
	skills.LoadFromDB(skills.SeedSkills())
	hero := inst.GetCombatantByID(heroID)
	hero.EquippedSkills = append([]string(nil), ids...)
	hero.SkillCooldowns = map[string]int{}
	hero.KitSpent = map[string]bool{}
}

func TestFenwatchBraceIsASkill(t *testing.T) {
	e, inst, heroID, enemyID := newFight(characters.ClassWarrior, 0)
	equipKit(inst, heroID, "warrior_brace")
	for i := 0; i < 5; i++ {
		res := e.ProcessAttack(inst, enemyID, heroID)
		if strings.Contains(res.Message, "You brace.") {
			t.Fatalf("brace proc without the button: %q", res.Message)
		}
	}
	got := e.ProcessSkill(inst, heroID, "warrior_brace", "")
	if !got.Success || !got.KeepsSwing {
		t.Fatalf("brace skill %+v", got)
	}
	hero := inst.GetCombatantByID(heroID)
	if hero.BraceLeft != 1 {
		t.Fatalf("brace not armed: %d", hero.BraceLeft)
	}
	again := e.ProcessSkill(inst, heroID, "warrior_brace", "")
	if again.Success {
		t.Fatal("brace used twice")
	}
	var landed bool
	for i := 0; i < 40; i++ {
		res := e.ProcessAttack(inst, enemyID, heroID)
		if res.Hit {
			landed = true
			break
		}
	}
	if !landed {
		t.Fatal("could not land a hit on the brace")
	}
	hero = inst.GetCombatantByID(heroID)
	if hero.BraceLeft != 0 {
		t.Fatalf("brace left %d", hero.BraceLeft)
	}
}

func TestAlleySlipDropsCombat(t *testing.T) {
	e, inst, heroID, enemyID := newFight(characters.ClassRogue, 0)
	equipKit(inst, heroID, "rogue_slip", "rogue_smoke")
	res := e.ProcessAttack(inst, enemyID, heroID)
	if strings.Contains(res.Message, "slip") {
		t.Fatalf("slip proc: %q", res.Message)
	}
	got := e.ProcessSkill(inst, heroID, "rogue_slip", "")
	if !got.Success || !got.SlipMove || got.KeepsSwing {
		t.Fatalf("slip %+v", got)
	}
	hero := inst.GetCombatantByID(heroID)
	if !hero.HasFled {
		t.Fatal("slip did not drop combat")
	}
	if again := e.ProcessSkill(inst, heroID, "rogue_slip", ""); again.Success {
		t.Fatal("slip twice")
	}
	smoke := e.ProcessSkill(inst, heroID, "rogue_smoke", enemyID)
	if !smoke.Success {
		t.Fatalf("smoke %+v", smoke)
	}
	enemy := inst.GetCombatantByID(enemyID)
	if !enemy.SmokeMiss {
		t.Fatal("smoke did not arm a miss")
	}
}

func TestRuneHandInscribeRefreshNoMana(t *testing.T) {
	e, inst, heroID, enemyID := newFight(characters.ClassWizard, 11)
	equipKit(inst, heroID, "mage_inscribe")
	res := e.ProcessAttack(inst, heroID, enemyID)
	if strings.Contains(res.Message, "inscribe") {
		t.Fatalf("inscribe proc: %q", res.Message)
	}
	got := e.ProcessSkill(inst, heroID, "mage_inscribe", enemyID)
	if !got.Success || got.KeepsSwing {
		t.Fatalf("inscribe %+v", got)
	}
	hero := inst.GetCombatantByID(heroID)
	if hero.CurrentMana != 11 {
		t.Fatalf("inscribe spent mana, left %d", hero.CurrentMana)
	}
	if again := e.ProcessSkill(inst, heroID, "mage_inscribe", enemyID); again.Success {
		t.Fatal("inscribe ignored cooldown")
	}
	hero.SkillCooldowns = map[string]int{}
	e.UpdateCombatant(inst, hero)
	if refresh := e.ProcessSkill(inst, heroID, "mage_inscribe", enemyID); !refresh.Success {
		t.Fatalf("refresh %+v", refresh)
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
