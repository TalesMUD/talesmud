package game

import (
	"sort"
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
)

// TestSummonedAddKillOrdersDoNotStallTurn drives the real combat tick after
// adds join through SummonCombatAllies. The first victim is last in the order
// and another enemy has already acted, which is the shape that used to leave
// CurrentTurnIdx past the order.
func TestSummonedAddKillOrdersDoNotStallTurn(t *testing.T) {
	orders := [][]string{
		{"add1", "add2", "boss"},
		{"add2", "add1", "boss"},
		{"boss", "add1", "add2"},
		{"boss", "add2", "add1"},
		{"add1", "boss", "add2"},
	}
	for _, order := range orders {
		order := order
		t.Run(joinKillOrder(order), func(t *testing.T) {
			runSummonKillOrder(t, order)
		})
	}
}

func joinKillOrder(order []string) string {
	out := order[0]
	for _, name := range order[1:] {
		out += "-" + name
	}
	return out
}

func runSummonKillOrder(t *testing.T, order []string) {
	t.Helper()
	room := "R-hang-" + joinKillOrder(order)
	g, inst, hero := startLowFight(t, room, "", "", "knight", 80, &npc.EnemyTrait{
		AttackPower: 1, Defense: 0, Difficulty: "boss",
	})
	importEnemyTemplate(t, g, "TPL-RAT", "", true, 0)
	if got := g.CombatController.SummonCombatAllies("knight", "TPL-RAT", 2); got != 2 {
		t.Fatalf("summoned %d adds, want 2", got)
	}

	var addIDs []string
	for _, enemy := range inst.Enemies {
		if enemy.Summoned {
			addIDs = append(addIDs, enemy.ID)
		}
	}
	sort.Strings(addIDs)
	if len(addIDs) != 2 {
		t.Fatalf("summoned ids = %v", addIDs)
	}
	ids := map[string]string{
		"boss": "knight",
		"add1": addIDs[0],
		"add2": addIDs[1],
	}
	killIDs := make([]string, len(order))
	for i, name := range order {
		killIDs[i] = ids[name]
	}

	player := inst.GetPlayerByID(hero.ID)
	player.AttackPower = 40
	player.STRMod = 30
	player.MaxHP = 100000
	player.CurrentHP = 100000
	player.Defense = 500
	g.CombatController.engine.UpdateCombatant(inst, player)
	for i := range inst.Enemies {
		inst.Enemies[i].AttackPower = 1
		inst.Enemies[i].Defense = 0
		if inst.Enemies[i].MaxHP < 5000 {
			inst.Enemies[i].MaxHP = 5000
		}
		inst.Enemies[i].CurrentHP = inst.Enemies[i].MaxHP
		g.CombatController.engine.UpdateCombatant(inst, &inst.Enemies[i])
	}

	before := make([]string, 0, len(killIDs)-1)
	for _, id := range killIDs[1:] {
		before = append(before, id)
	}
	pinStallOrder(t, g, inst, hero.ID, before, killIDs[0])

	const maxTicks = 180
	const ticksAfterLastDeath = 4
	lastDeathTick := -1
	for tick := 0; tick < maxTicks; tick++ {
		if inst.State == combat.CombatStateVictory {
			if lastDeathTick >= 0 && tick-lastDeathTick > ticksAfterLastDeath {
				t.Fatalf("victory took %d ticks after the last death", tick-lastDeathTick)
			}
			assertBossDeadAddsGone(t, g, addIDs)
			return
		}
		if len(inst.GetLivingEnemies()) == 0 && lastDeathTick < 0 {
			lastDeathTick = tick
		}
		if lastDeathTick >= 0 && tick-lastDeathTick > ticksAfterLastDeath && inst.State != combat.CombatStateVictory {
			t.Fatalf("state %s %d ticks after the last enemy died (round %d idx %d)",
				inst.State, tick-lastDeathTick, inst.Round, inst.CurrentTurnIdx)
		}

		victim := nextLivingKill(inst, killIDs)
		if victim != "" {
			ref := inst.GetEnemyByID(victim)
			if ref != nil && ref.IsAlive {
				ref.CurrentHP = 1
				g.CombatController.engine.UpdateCombatant(inst, ref)
				g.CombatController.SetAutoAttackTarget(hero.ID, victim)
			}
			for _, enemy := range inst.Enemies {
				if !enemy.IsAlive || enemy.ID == victim {
					continue
				}
				live := inst.GetEnemyByID(enemy.ID)
				if live == nil || !live.IsAlive {
					continue
				}
				if live.CurrentHP < live.MaxHP {
					live.CurrentHP = live.MaxHP
					g.CombatController.engine.UpdateCombatant(inst, live)
				}
			}
		}

		beforeLog := len(inst.Log)
		beforeIdx := inst.CurrentTurnIdx
		beforeRound := inst.Round
		beforePhase := inst.Phase
		beforeState := inst.State
		beforeID := ""
		if cur := inst.GetCurrentTurnCombatant(); cur != nil {
			beforeID = cur.ID
		}

		inst.NextActionAt = time.Time{}
		inst.DecisionDeadline = time.Now().Add(-time.Second)
		g.CombatController.processAllTurns(inst)
		_ = drainGameMessages(g.SendMessage())

		if inst.State == combat.CombatStateActive {
			if inst.GetCurrentTurnCombatant() == nil {
				t.Fatalf("tick %d stalled with a nil current combatant (round %d idx %d order %d)",
					tick, inst.Round, inst.CurrentTurnIdx, len(inst.TurnOrder))
			}
			if inst.CurrentTurnIdx < 0 || inst.CurrentTurnIdx >= len(inst.TurnOrder) {
				t.Fatalf("tick %d index %d past order %d", tick, inst.CurrentTurnIdx, len(inst.TurnOrder))
			}
		}
		if len(inst.GetLivingEnemies()) == 0 && lastDeathTick < 0 {
			lastDeathTick = tick
		}

		afterID := ""
		if cur := inst.GetCurrentTurnCombatant(); cur != nil {
			afterID = cur.ID
		}
		progressed := inst.State != beforeState || inst.Phase != beforePhase || inst.Round != beforeRound ||
			inst.CurrentTurnIdx != beforeIdx || len(inst.Log) != beforeLog || afterID != beforeID
		if inst.State == combat.CombatStateActive && !progressed {
			t.Fatalf("tick %d resolved nothing (phase %s round %d idx %d)", tick, inst.Phase, inst.Round, inst.CurrentTurnIdx)
		}
	}
	t.Fatalf("fight did not end in %d ticks: state=%s living=%d round=%d",
		maxTicks, inst.State, len(inst.GetLivingEnemies()), inst.Round)
}

func nextLivingKill(inst *combat.CombatInstance, ids []string) string {
	for _, id := range ids {
		ref := inst.GetCombatantByID(id)
		if ref != nil && ref.IsAlive {
			return id
		}
	}
	return ""
}

func pinStallOrder(t *testing.T, g *Game, inst *combat.CombatInstance, playerID string, before []string, victimID string) {
	t.Helper()
	init := 100
	for _, id := range before {
		ref := inst.GetCombatantByID(id)
		if ref == nil {
			t.Fatalf("missing combatant %s", id)
		}
		ref.Initiative = init
		init -= 5
	}
	player := inst.GetPlayerByID(playerID)
	if player == nil {
		t.Fatal("missing player")
	}
	player.Initiative = 40
	victim := inst.GetCombatantByID(victimID)
	if victim == nil {
		t.Fatalf("missing victim %s", victimID)
	}
	victim.Initiative = 1
	g.CombatController.engine.BuildTurnOrder(inst)
	if len(inst.TurnOrder) < 3 {
		t.Fatalf("turn order too short: %d", len(inst.TurnOrder))
	}
	if inst.TurnOrder[0].Type != combat.CombatantTypeNPC {
		t.Fatal("expected a living enemy to act before the player")
	}
	last := inst.TurnOrder[len(inst.TurnOrder)-1]
	if last.ID != victimID {
		t.Fatalf("victim %s is not last; last is %s", victimID, last.ID)
	}
}

func assertBossDeadAddsGone(t *testing.T, g *Game, addIDs []string) {
	t.Helper()
	boss := g.NPCManager.GetInstance("knight")
	if boss == nil || !boss.IsDead {
		t.Fatalf("boss IsDead = %#v", boss)
	}
	for _, id := range addIDs {
		if g.NPCManager.GetInstance(id) != nil {
			t.Fatalf("summoned add %s still registered after victory", id)
		}
	}
}
