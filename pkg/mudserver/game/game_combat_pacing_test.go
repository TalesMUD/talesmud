package game

import (
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// seedPacedCombat builds a 1v1 instance registered with the combat manager.
func seedPacedCombat(t *testing.T, g *Game, charID, enemyID string, playerFirst bool) *combat.CombatInstance {
	t.Helper()
	mgr := g.CombatController.manager
	inst := mgr.CreateInstance("R-pace")
	player := combat.CombatantRef{
		ID: charID, Type: combat.CombatantTypePlayer, Name: "Hero",
		IsAlive: true, MaxHP: 50, CurrentHP: 50, AttackPower: 8, Defense: 2,
		AutoAttackTargetID: enemyID,
	}
	enemy := combat.CombatantRef{
		ID: enemyID, Type: combat.CombatantTypeNPC, Name: "Rat",
		IsAlive: true, MaxHP: 30, CurrentHP: 30, AttackPower: 3, Defense: 1,
	}
	if playerFirst {
		player.Initiative = 20
		enemy.Initiative = 5
	} else {
		player.Initiative = 5
		enemy.Initiative = 20
	}
	inst.Players = []combat.CombatantRef{player}
	inst.Enemies = []combat.CombatantRef{enemy}
	inst.State = combat.CombatStateActive
	inst.Round = 1
	inst.TurnStartTime = time.Now()
	g.CombatController.engine.BuildTurnOrder(inst)
	mgr.RegisterPlayer(charID, inst.ID)
	mgr.RegisterNPC(enemyID, inst.ID)
	return inst
}

func TestPacingTwoTurnsNeedTimeAdvance(t *testing.T) {
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "R-pace", nil)

	char, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "char-pace"},
		Name:             "Hero",
		BelongsUser:      *traits.BelongsToUser("user-pace"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "R-pace"},
		MaxHitPoints:     50,
		CurrentHitPoints: 50,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Short beats for the test; still non-zero so a second Update without advancing NextActionAt is gated.
	cfg := g.CombatController.engine.Config
	cfg.DecisionWindowSeconds = 10
	cfg.TurnBeatMs = 5000
	cfg.ReactionMs = 5000

	enemyID := "npc-pace-rat"
	g.NPCManager.RegisterExistingNPC(&npc.NPC{
		Entity:           &entities.Entity{ID: enemyID},
		Name:             "Rat",
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "R-pace"},
		MaxHitPoints:     30,
		CurrentHitPoints: 30,
		Level:            1,
		EnemyTrait:       &npc.EnemyTrait{AttackPower: 3, Defense: 1},
	}, "R-pace")

	inst := seedPacedCombat(t, g, char.ID, enemyID, true)

	// First Update: opens player decision window, does not resolve.
	g.CombatController.processAllTurns(inst)
	if inst.Phase != combat.CombatPhaseWaitingPlayer {
		t.Fatalf("expected waitingPlayer after first Update, got %q", inst.Phase)
	}
	turnIdxBefore := inst.CurrentTurnIdx
	enemyHPBefore := inst.GetEnemyByID(enemyID).CurrentHP

	// Expire decision window so auto-attack can fire, keep NextActionAt clear.
	inst.DecisionDeadline = time.Now().Add(-time.Second)
	_ = drainGameMessages(g.SendMessage())
	g.CombatController.processAllTurns(inst)

	if inst.Phase == combat.CombatPhaseWaitingPlayer {
		t.Fatal("expected decision window to resolve after deadline")
	}
	if inst.NextActionAt.IsZero() || !time.Now().Before(inst.NextActionAt) {
		t.Fatalf("expected NextActionAt in the future after resolve, got %v", inst.NextActionAt)
	}
	// Turn must have advanced (NextTurn) after the single resolve.
	if inst.CurrentTurnIdx == turnIdxBefore && inst.Round == 1 && len(inst.TurnOrder) > 1 {
		// If only one living combatant somehow, skip; otherwise fail.
		t.Fatalf("expected turn to advance after resolve (idx still %d)", turnIdxBefore)
	}
	phaseAfterFirst := inst.Phase
	nextAt := inst.NextActionAt
	turnIdxAfterFirst := inst.CurrentTurnIdx
	roundAfterFirst := inst.Round
	enemyHPAfterFirst := inst.GetEnemyByID(enemyID).CurrentHP

	var sawAction bool
	for _, out := range drainGameMessages(g.SendMessage()) {
		if msg, ok := out.(*messages.CombatActionMessage); ok && msg.Action == "attack" {
			sawAction = true
			break
		}
	}
	if !sawAction {
		t.Fatal("expected structured combatAction from first resolve")
	}

	// Second Update immediately: pacing gate must block another resolve.
	g.CombatController.processAllTurns(inst)
	if inst.CurrentTurnIdx != turnIdxAfterFirst || inst.Round != roundAfterFirst {
		t.Fatal("two turns resolved without time advancing (turn index/round changed)")
	}
	if inst.Phase != phaseAfterFirst {
		t.Fatalf("phase changed under pacing gate: %q -> %q", phaseAfterFirst, inst.Phase)
	}
	if !inst.NextActionAt.Equal(nextAt) {
		t.Fatal("NextActionAt should be unchanged while gated")
	}
	// Enemy HP must not change again while gated (player already acted; NPC not yet).
	if inst.GetEnemyByID(enemyID).CurrentHP != enemyHPAfterFirst {
		t.Fatal("enemy HP changed while pacing gate should block NPC turn")
	}
	_ = enemyHPBefore // silence if unused when attack always misses

	// Advance past beat budget → next turn may proceed.
	inst.NextActionAt = time.Now().Add(-time.Millisecond)
	g.CombatController.processAllTurns(inst)
	if inst.NextActionAt.Equal(nextAt) && inst.CurrentTurnIdx == turnIdxAfterFirst && inst.Phase == phaseAfterFirst {
		t.Fatal("expected progress after NextActionAt elapsed")
	}
}

func TestPlayerAutoAttacksAfterDecisionWindow(t *testing.T) {
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "R-pace", nil)

	char, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "char-auto"},
		Name:             "Hero",
		BelongsUser:      *traits.BelongsToUser("user-auto"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "R-pace"},
		MaxHitPoints:     50,
		CurrentHitPoints: 50,
	})
	if err != nil {
		t.Fatal(err)
	}

	cfg := g.CombatController.engine.Config
	cfg.DecisionWindowSeconds = 10
	cfg.TurnBeatMs = 1
	cfg.ReactionMs = 1

	enemyID := "npc-auto-rat"
	g.NPCManager.RegisterExistingNPC(&npc.NPC{
		Entity:           &entities.Entity{ID: enemyID},
		Name:             "Rat",
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "R-pace"},
		MaxHitPoints:     30,
		CurrentHitPoints: 30,
		Level:            1,
		EnemyTrait:       &npc.EnemyTrait{AttackPower: 3, Defense: 1},
	}, "R-pace")

	inst := seedPacedCombat(t, g, char.ID, enemyID, true)

	// Open decision window
	g.CombatController.processAllTurns(inst)
	if inst.Phase != combat.CombatPhaseWaitingPlayer {
		t.Fatalf("expected waitingPlayer, got %q", inst.Phase)
	}

	player := inst.GetPlayerByID(char.ID)
	if player == nil || player.QueuedAction != "" {
		t.Fatal("expected no queued action")
	}

	// Still inside window → no resolve
	enemyHPBefore := inst.GetEnemyByID(enemyID).CurrentHP
	g.CombatController.processAllTurns(inst)
	if inst.GetEnemyByID(enemyID).CurrentHP != enemyHPBefore && inst.Phase == combat.CombatPhaseWaitingPlayer {
		// HP unchanged while waiting is the success path; if phase moved something else happened
	}
	if inst.Phase != combat.CombatPhaseWaitingPlayer {
		t.Fatalf("should still be waiting inside decision window, got %q", inst.Phase)
	}

	// Expire window with no queue → auto-attack
	inst.DecisionDeadline = time.Now().Add(-time.Millisecond)
	_ = drainGameMessages(g.SendMessage())
	g.CombatController.processAllTurns(inst)

	var sawAction *messages.CombatActionMessage
	for _, out := range drainGameMessages(g.SendMessage()) {
		if msg, ok := out.(*messages.CombatActionMessage); ok {
			sawAction = msg
			break
		}
	}
	if sawAction == nil {
		t.Fatal("expected structured combatAction after auto-attack")
	}
	if sawAction.Action != "attack" {
		t.Fatalf("expected action=attack, got %q", sawAction.Action)
	}
	if sawAction.ActorID != char.ID {
		t.Fatalf("expected actorId=%s, got %s", char.ID, sawAction.ActorID)
	}
}

func TestCombatActionPayloadHasDamageAndResult(t *testing.T) {
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "R-pace", nil)

	char, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "char-dmg"},
		Name:             "Hero",
		BelongsUser:      *traits.BelongsToUser("user-dmg"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "R-pace"},
		MaxHitPoints:     50,
		CurrentHitPoints: 50,
	})
	if err != nil {
		t.Fatal(err)
	}

	cfg := g.CombatController.engine.Config
	cfg.DecisionWindowSeconds = 1
	cfg.TurnBeatMs = 1
	cfg.ReactionMs = 1

	enemyID := "npc-dmg-rat"
	g.NPCManager.RegisterExistingNPC(&npc.NPC{
		Entity:           &entities.Entity{ID: enemyID},
		Name:             "Rat",
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "R-pace"},
		MaxHitPoints:     30,
		CurrentHitPoints: 30,
		Level:            1,
		EnemyTrait:       &npc.EnemyTrait{AttackPower: 3, Defense: 0},
	}, "R-pace")

	inst := seedPacedCombat(t, g, char.ID, enemyID, true)
	// Force queue so we resolve without waiting full window after announce tick
	g.CombatController.processAllTurns(inst) // announce
	inst.GetPlayerByID(char.ID).QueuedAction = combat.CombatActionAttack
	inst.GetPlayerByID(char.ID).QueuedTargetID = enemyID
	inst.DecisionDeadline = time.Now().Add(-time.Millisecond)
	inst.NextActionAt = time.Time{}

	_ = drainGameMessages(g.SendMessage())
	g.CombatController.processAllTurns(inst)

	var saw *messages.CombatActionMessage
	for _, out := range drainGameMessages(g.SendMessage()) {
		if msg, ok := out.(*messages.CombatActionMessage); ok && msg.Action == "attack" {
			saw = msg
			break
		}
	}
	if saw == nil {
		t.Fatal("expected combatAction with action=attack")
	}
	if saw.Result == "" {
		t.Fatal("expected non-empty result (hit/miss/crit)")
	}
	if saw.FxID == "" {
		t.Fatal("expected fxId hint")
	}
	if saw.TargetID != enemyID {
		t.Fatalf("expected targetId=%s, got %s", enemyID, saw.TargetID)
	}
	if saw.MaxHP != 30 {
		t.Fatalf("expected maxHp=30, got %d", saw.MaxHP)
	}
	// Damage may be 0 on miss; result field still present. On hit, damage > 0.
	if saw.Result == "hit" || saw.Result == "crit" {
		if saw.Damage <= 0 {
			t.Fatalf("expected damage on %s, got %d", saw.Result, saw.Damage)
		}
	}
	if saw.Message == "" {
		t.Fatal("terminal prose Message must remain populated")
	}
	if len(saw.Combatants) == 0 {
		t.Fatal("expected combatant HP snapshots on combatAction")
	}
}

func TestCombatTurnEmittedWithDeadline(t *testing.T) {
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "R-pace", nil)

	char, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "char-turn"},
		Name:             "Hero",
		BelongsUser:      *traits.BelongsToUser("user-turn"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "R-pace"},
		MaxHitPoints:     50,
		CurrentHitPoints: 50,
	})
	if err != nil {
		t.Fatal(err)
	}

	enemyID := "npc-turn-rat"
	g.NPCManager.RegisterExistingNPC(&npc.NPC{
		Entity:           &entities.Entity{ID: enemyID},
		Name:             "Rat",
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "R-pace"},
		MaxHitPoints:     30,
		CurrentHitPoints: 30,
		Level:            1,
		EnemyTrait:       &npc.EnemyTrait{AttackPower: 3, Defense: 1},
	}, "R-pace")

	inst := seedPacedCombat(t, g, char.ID, enemyID, true)
	_ = drainGameMessages(g.SendMessage())
	g.CombatController.processAllTurns(inst)

	var saw *messages.CombatTurnMessage
	for _, out := range drainGameMessages(g.SendMessage()) {
		if msg, ok := out.(*messages.CombatTurnMessage); ok {
			saw = msg
			break
		}
	}
	if saw == nil {
		t.Fatal("expected combatTurn when player turn starts")
	}
	if saw.ActorID != char.ID {
		t.Fatalf("actorId=%s want %s", saw.ActorID, char.ID)
	}
	if saw.DeadlineMs == 0 {
		t.Fatal("expected deadlineMs for player decision window")
	}
	if saw.Round != 1 {
		t.Fatalf("round=%d want 1", saw.Round)
	}
}
