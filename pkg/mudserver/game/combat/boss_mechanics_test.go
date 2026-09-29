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

func TestBossPhaseThresholdsAndNoDoubleFire(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	e := NewEngine(NewManager(), DefaultConfig())
	inst := e.InitiateCombat("phases", []*characters.Character{testFighter("Hero", 800)}, []*npc.NPC{testEnemy("Boss", "boss", 100, 8)})
	boss := &inst.Enemies[0]
	if boss.BossPhase != 1 || boss.BossPhaseLabel != "Opening" || boss.BossPhaseCount != 3 {
		t.Fatalf("opening: %+v", boss)
	}
	count := func() int {
		n := 0
		for _, entry := range inst.Log {
			if entry.Result == "phase-enter" {
				n++
			}
		}
		return n
	}
	for _, check := range []struct {
		hp            int32
		phase, events int
	}{
		{67, 1, 0}, {66, 2, 1}, {66, 2, 1}, {100, 2, 1}, {65, 2, 1}, {34, 2, 1}, {33, 3, 2}, {33, 3, 2}, {100, 3, 2},
	} {
		boss.CurrentHP = check.hp
		e.UpdateCombatant(inst, boss)
		if boss.BossPhase != check.phase || count() != check.events {
			t.Fatalf("HP %d: phase %d events %d", check.hp, boss.BossPhase, count())
		}
		turn := inst.GetCombatantByID(boss.ID)
		if turn.BossPhase != check.phase {
			t.Fatal("turn-order snapshot lost phase")
		}
	}
}

func TestBossPhaseMultiThresholdAndLethalHit(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	e := NewEngine(NewManager(), DefaultConfig())
	inst := e.InitiateCombat("skip", []*characters.Character{testFighter("Hero", 800)}, []*npc.NPC{testEnemy("Boss", "boss", 100, 8)})
	boss := &inst.Enemies[0]
	boss.CurrentHP = 20
	e.UpdateCombatant(inst, boss)
	phases := []string{}
	for _, entry := range inst.Log {
		if entry.Result == "phase-enter" {
			phases = append(phases, entry.Message)
		}
	}
	if len(phases) != 2 || boss.BossPhase != 3 {
		t.Fatalf("multi-threshold: phase %d events %v", boss.BossPhase, phases)
	}
	inst2 := e.InitiateCombat("lethal", []*characters.Character{testFighter("Hero", 800)}, []*npc.NPC{testEnemy("Boss", "boss", 100, 8)})
	dead := &inst2.Enemies[0]
	dead.CurrentHP = 0
	dead.IsAlive = false
	e.UpdateCombatant(inst2, dead)
	e.refreshEnrage(inst2, dead)
	if dead.BossPhase != 1 || dead.Enraged {
		t.Fatal("lethal hit entered a phase or enraged")
	}
}

func TestBossPhaseTrashAndEliteUnchanged(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	for _, tier := range []string{"trivial", "easy", "normal", "hard", "elite"} {
		e := NewEngine(NewManager(), DefaultConfig())
		inst := e.InitiateCombat(tier, []*characters.Character{testFighter("Hero", 800)}, []*npc.NPC{testEnemy("Other", tier, 100, 8)})
		actor := &inst.Enemies[0]
		actor.CurrentHP = 20
		e.UpdateCombatant(inst, actor)
		if actor.BossPhase != 0 || actor.BossPhaseLabel != "" {
			t.Fatalf("%s acquired phases", tier)
		}
		for _, entry := range inst.Log {
			if entry.Result == "phase-enter" {
				t.Fatalf("%s emitted phase-enter", tier)
			}
		}
	}
}

func TestBossPhaseTelegraphAndEnrageCompose(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	e := NewEngine(NewManager(), DefaultConfig())
	inst := e.InitiateCombat("compose", []*characters.Character{testFighter("Hero", 800)}, []*npc.NPC{testEnemy("Boss", "boss", 100, 8)})
	actor := &inst.Enemies[0]
	actor.CurrentHP = 66
	e.UpdateCombatant(inst, actor)
	step := stepNextEnemy(t, e, inst)
	if !step.Telegraph || step.Ability != "Shattering Blow" {
		t.Fatalf("phase telegraph: %+v", step)
	}
	actor = &inst.Enemies[0]
	actor.CurrentHP = 33
	e.UpdateCombatant(inst, actor)
	if actor.Enraged {
		t.Fatal("33% should retain A6 30% threshold")
	}
	step = stepNextEnemy(t, e, inst)
	if step.Telegraph || step.Ability != "Shattering Blow" {
		t.Fatalf("pending blow should retain original label: %+v", step)
	}
	step = stepNextEnemy(t, e, inst)
	if !step.Telegraph || step.Ability != "Desperate Crush" {
		t.Fatalf("final phase telegraph: %+v", step)
	}
	actor = &inst.Enemies[0]
	actor.CurrentHP = 30
	step = stepNextEnemy(t, e, inst)
	if !inst.Enemies[0].Enraged || step.Telegraph || inst.Enemies[0].TelegraphTurns != 0 {
		t.Fatal("A6 enrage must cancel/skip phase wind-ups")
	}
}

func TestBossPhaseDamageOverTime(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	e := NewEngine(NewManager(), DefaultConfig())
	inst := e.InitiateCombat("dot", []*characters.Character{testFighter("Hero", 800)}, []*npc.NPC{testEnemy("Boss", "boss", 100, 8)})
	actor := &inst.Enemies[0]
	actor.CurrentHP = 70
	actor.StatusEffects = []entcombat.StatusEffect{{ID: "burn", Name: "Burn", Type: "dot", Value: 10, Duration: 2}}
	e.ProcessStatusEffects(inst, actor)
	if inst.Enemies[0].BossPhase != 2 || inst.Enemies[0].CurrentHP != 60 {
		t.Fatalf("DoT did not advance phase: %+v", inst.Enemies[0])
	}
}
