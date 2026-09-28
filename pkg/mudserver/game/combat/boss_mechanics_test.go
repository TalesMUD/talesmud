package combat

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	entcombat "github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
)

func testFighter(name string, hp int32) *characters.Character {
	return &characters.Character{
		Entity:           entities.NewEntity(),
		Name:             name,
		Level:            5,
		MaxHitPoints:     hp,
		CurrentHitPoints: hp,
	}
}

func testEnemy(name, difficulty string, hp, atk int32) *npc.NPC {
	return &npc.NPC{
		Entity:           entities.NewEntity(),
		Name:             name,
		Level:            5,
		MaxHitPoints:     hp,
		CurrentHitPoints: hp,
		EnemyTrait: &npc.EnemyTrait{
			AttackPower: atk,
			Defense:     0,
			Difficulty:  difficulty,
		},
	}
}

func stepNextEnemy(t *testing.T, e *Engine, inst *entcombat.CombatInstance) NPCAttackStep {
	t.Helper()
	for i := 0; i < 12; i++ {
		cur := inst.GetCurrentTurnCombatant()
		if cur != nil && cur.Type == entcombat.CombatantTypeNPC {
			living := inst.GetLivingPlayers()
			if len(living) == 0 {
				t.Fatal("no living player")
			}
			step := e.StepNPCAttack(inst, cur.ID, living[0].ID)
			e.NextTurn(inst)
			return step
		}
		e.NextTurn(inst)
	}
	t.Fatal("no enemy turn")
	return NPCAttackStep{}
}

func TestBossTelegraphDelaysDamage(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	e := NewEngine(NewManager(), DefaultConfig())
	player := testFighter("Hero", 80)
	boss := testEnemy("Stone Warden", "boss", 400, 8)
	inst := e.InitiateCombat("room", []*characters.Character{player}, []*npc.NPC{boss})
	start := inst.Players[0].CurrentHP

	first := stepNextEnemy(t, e, inst)
	if !first.Telegraph {
		t.Fatalf("first boss action should telegraph, message %q", first.Message)
	}
	if inst.Players[0].CurrentHP != start {
		t.Fatalf("telegraph dealt damage, hp %d -> %d", start, inst.Players[0].CurrentHP)
	}
	windup := inst.GetCombatantByID(boss.Entity.ID)
	if windup == nil || windup.TelegraphAbility == "" {
		t.Fatal("wind-up ability was not stored")
	}

	second := stepNextEnemy(t, e, inst)
	if second.Telegraph {
		t.Fatal("the stored wind-up should resolve on the next enemy action")
	}
	after := inst.GetCombatantByID(boss.Entity.ID)
	if after == nil || after.TelegraphAbility != "" || after.TelegraphTurns != 0 {
		t.Fatal("telegraph should clear when the hit resolves")
	}
	if inst.Players[0].CurrentHP == start && !second.Attack.Miss {
		t.Fatalf("resolved hit missed the hp bar and was not a miss: %+v", second.Attack)
	}
}

func TestTrashDoesNotTelegraph(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	e := NewEngine(NewManager(), DefaultConfig())
	inst := e.InitiateCombat("room", []*characters.Character{testFighter("Hero", 40)}, []*npc.NPC{testEnemy("Rat", "easy", 20, 3)})
	step := stepNextEnemy(t, e, inst)
	if step.Telegraph {
		t.Fatal("trash should strike without a wind-up")
	}
}

func TestBossEnrageBoostsDamage(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	e := NewEngine(NewManager(), DefaultConfig())
	inst := e.InitiateCombat("room", []*characters.Character{testFighter("Hero", 40)}, []*npc.NPC{testEnemy("Stone Warden", "boss", 100, 10)})
	boss := &inst.Enemies[0]
	boss.CurrentHP = 20
	e.refreshEnrage(inst, boss)
	if !inst.Enemies[0].Enraged {
		t.Fatal("boss under 30% HP should enrage")
	}

	plain := &entcombat.CombatantRef{AttackPower: 10, Level: 5, Type: entcombat.CombatantTypeNPC, Difficulty: "boss"}
	target := &entcombat.CombatantRef{Defense: 0, Level: 5}
	base := e.CalculateDamage(plain, target, false)
	plain.Enraged = true
	raged := e.CalculateDamage(plain, target, false)
	if raged <= base {
		t.Fatalf("enrage damage %d should exceed %d", raged, base)
	}

	inst.Round = 1
	elite := testEnemy("Bear", "hard", 80, 6)
	eliteInst := e.InitiateCombat("room-2", []*characters.Character{testFighter("Hero", 40)}, []*npc.NPC{elite})
	eliteInst.Enemies[0].CurrentHP = 5
	eliteInst.Round = 30
	e.refreshEnrage(eliteInst, &eliteInst.Enemies[0])
	if eliteInst.Enemies[0].Enraged {
		t.Fatal("elites telegraph but do not enrage")
	}
}
