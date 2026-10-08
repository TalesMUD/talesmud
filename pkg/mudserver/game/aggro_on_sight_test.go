package game

import (
	"strings"
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/ruleset"
	"github.com/talesmud/talesmud/pkg/service"
)

func useAggroClock(t *testing.T, g *Game) *time.Time {
	t.Helper()
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	g.SetAggroClock(func() time.Time { return now })
	ruleset.SetAggroOnSight(true, 500*time.Millisecond, 5, 15*time.Second)
	t.Cleanup(ruleset.Reset)
	return &now
}

func aggroEnemy(id, name string, level int32, trait *npc.EnemyTrait) *npc.NPC {
	if trait == nil {
		trait = &npc.EnemyTrait{AggroOnSight: true}
	}
	if trait.Difficulty == "" {
		trait.Difficulty = "easy"
	}
	enemy := hookEnemy(id, name, 40, 40, trait)
	enemy.Level = level
	return enemy
}

func placeAggro(t *testing.T, g *Game, facade service.Facade, room, npcID, npcName string, level int32, trait *npc.EnemyTrait) (*characters.Character, *npc.NPC) {
	t.Helper()
	storeTestRoom(t, facade, room, nil)
	_, hero := sharePlayer(t, facade, "user-"+npcID, "ref-"+npcID, "char-"+npcID, "Hero", room)
	enemy := aggroEnemy(npcID, npcName, level, trait)
	enemy.EnemyTrait.AggroOnSight = true
	g.NPCManager.RegisterExistingNPC(enemy, room)
	return hero, enemy
}

func online(g *Game, facade service.Facade, hero *characters.Character) {
	user, err := facade.UsersService().FindByID(hero.BelongsUserID)
	if err != nil || user == nil {
		panic("aggro test user missing")
	}
	connectSharePlayer(g, user, hero)
}

func advanceAggro(now *time.Time, g *Game, d time.Duration) {
	*now = now.Add(d)
	g.FireDueAggro()
}

func spotsYou(msgs []interface{}) (messages.MessageResponse, bool) {
	for _, out := range msgs {
		msg, ok := out.(messages.MessageResponse)
		if ok && strings.Contains(msg.Message, "spots you") {
			return msg, true
		}
	}
	return messages.MessageResponse{}, false
}

func combatStarted(msgs []interface{}) bool {
	for _, out := range msgs {
		if _, ok := out.(*messages.CombatStartMessage); ok {
			return true
		}
	}
	return false
}

func partyLine(msgs []interface{}) string {
	for _, out := range msgs {
		msg, ok := out.(messages.MessageResponse)
		if ok && strings.Contains(msg.Message, "[Party]") {
			return msg.Message
		}
	}
	return ""
}

func TestAggroFiresAfterGrace(t *testing.T) {
	g, facade := newHookTestGame(t)
	now := useAggroClock(t, g)
	hero, _ := placeAggro(t, g, facade, "R-sight", "wolf", "Grey Wolf", 2, nil)
	online(g, facade, hero)

	g.NotePlayerEntered(hero.ID, "R-sight")
	g.FireDueAggro()
	if g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("fired before the grace elapsed")
	}
	advanceAggro(now, g, 500*time.Millisecond)
	msgs := drainGameMessages(g.SendMessage())
	spot, ok := spotsYou(msgs)
	if !ok || spot.Style != "" || spot.Hook != "" || spot.Source != "" || spot.Type != messages.MessageTypeDefault {
		t.Fatalf("spots line = %+v ok=%v", spot, ok)
	}
	if !combatStarted(msgs) || !g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("grace did not start combat")
	}
}

func TestAggroLeavingDuringGraceCancels(t *testing.T) {
	g, facade := newHookTestGame(t)
	now := useAggroClock(t, g)
	storeTestRoom(t, facade, "R-safe", nil)
	hero, _ := placeAggro(t, g, facade, "R-sight", "wolf-leave", "Wolf", 2, nil)
	online(g, facade, hero)
	g.NotePlayerEntered(hero.ID, "R-sight")

	if err := facade.CharactersService().Modify(hero.ID, func(ch *characters.Character) error {
		ch.CurrentRoomID = "R-safe"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	g.NotePlayerEntered(hero.ID, "R-safe")
	advanceAggro(now, g, time.Second)
	if g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("leaving during grace started a fight")
	}

	if err := facade.CharactersService().Modify(hero.ID, func(ch *characters.Character) error {
		ch.CurrentRoomID = "R-sight"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	g.NotePlayerEntered(hero.ID, "R-sight")
	advanceAggro(now, g, 500*time.Millisecond)
	if !g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("re-enter after a grace cancel should still aggro; cooldown must not start")
	}
}

func TestAggroSkipsPlayerAlreadyInCombat(t *testing.T) {
	g, facade := newHookTestGame(t)
	now := useAggroClock(t, g)
	hero, first := placeAggro(t, g, facade, "R-busy", "wolf-a", "Wolf", 2, &npc.EnemyTrait{AggroOnSight: false})
	online(g, facade, hero)
	second := aggroEnemy("wolf-b", "Other Wolf", 2, nil)
	g.NPCManager.RegisterExistingNPC(second, "R-busy")
	fresh, err := facade.CharactersService().FindByID(hero.ID)
	if err != nil {
		t.Fatal(err)
	}
	user, _ := facade.UsersService().FindByID(hero.BelongsUserID)
	if g.CombatController.BeginEngagement("R-busy", fresh, user.ID, first, false) == nil {
		t.Fatal("manual fight did not start")
	}
	g.NotePlayerEntered(hero.ID, "R-busy")
	advanceAggro(now, g, time.Second)
	if g.CombatController.IsNPCInCombat(second.ID) {
		t.Fatal("second aggressive NPC started another fight")
	}
	inst := g.CombatController.GetCombatInstance(hero.ID)
	if inst == nil || len(inst.Enemies) != 1 {
		t.Fatalf("fight enemies = %v", inst)
	}
}

func TestAggroSkipsDeadAndGhost(t *testing.T) {
	g, facade := newHookTestGame(t)
	now := useAggroClock(t, g)
	hero, _ := placeAggro(t, g, facade, "R-dead", "wolf-dead", "Wolf", 2, nil)
	online(g, facade, hero)
	if err := facade.CharactersService().Modify(hero.ID, func(ch *characters.Character) error {
		ch.CurrentHitPoints = 0
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	g.NotePlayerEntered(hero.ID, "R-dead")
	advanceAggro(now, g, time.Second)
	if g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("dead player was aggroed")
	}

	if err := facade.CharactersService().Modify(hero.ID, func(ch *characters.Character) error {
		ch.CurrentHitPoints = 30
		ch.AwaitingReset = true
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	g.NotePlayerEntered(hero.ID, "R-dead")
	advanceAggro(now, g, time.Second)
	if g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("ghost player was aggroed")
	}
}

func TestAggroDisabledRuleset(t *testing.T) {
	g, facade := newHookTestGame(t)
	now := useAggroClock(t, g)
	ruleset.SetAggroOnSight(false, 500*time.Millisecond, 5, 15*time.Second)
	hero, _ := placeAggro(t, g, facade, "R-off", "wolf-off", "Wolf", 2, nil)
	online(g, facade, hero)
	g.NotePlayerEntered(hero.ID, "R-off")
	advanceAggro(now, g, time.Second)
	if g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("disabled ruleset started a fight")
	}
}

func TestAggroLevelGap(t *testing.T) {
	g, facade := newHookTestGame(t)
	now := useAggroClock(t, g)
	hero, _ := placeAggro(t, g, facade, "R-gap", "wolf-gap", "Wolf", 1, nil)
	online(g, facade, hero)
	if err := facade.CharactersService().Modify(hero.ID, func(ch *characters.Character) error {
		ch.Level = 10
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	g.NotePlayerEntered(hero.ID, "R-gap")
	advanceAggro(now, g, time.Second)
	if g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("level gap did not skip")
	}

	ruleset.SetAggroOnSight(true, 500*time.Millisecond, 0, 15*time.Second)
	g.NotePlayerEntered(hero.ID, "R-gap")
	advanceAggro(now, g, 500*time.Millisecond)
	if !g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("gap 0 should aggro any level")
	}
}

func TestAggroCooldownAfterFlee(t *testing.T) {
	g, facade := newHookTestGame(t)
	now := useAggroClock(t, g)
	hero, enemy := placeAggro(t, g, facade, "R-flee", "wolf-flee", "Wolf", 2, &npc.EnemyTrait{AttackPower: 0})
	online(g, facade, hero)
	g.NotePlayerEntered(hero.ID, "R-flee")
	advanceAggro(now, g, 500*time.Millisecond)
	if !g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("setup fight did not start")
	}
	fled := false
	for i := 0; i < 40 && g.CombatController.IsPlayerInCombat(hero.ID); i++ {
		ok, _, _, _ := g.CombatController.ProcessPlayerFlee(hero.ID)
		if ok {
			fled = true
		}
	}
	if !fled || g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("flee did not end the fight")
	}
	_ = enemy
	drainGameMessages(g.SendMessage())

	g.NotePlayerEntered(hero.ID, "R-flee")
	advanceAggro(now, g, 500*time.Millisecond)
	if g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("re-aggro during cooldown")
	}

	*now = now.Add(15 * time.Second)
	g.NotePlayerEntered(hero.ID, "R-flee")
	advanceAggro(now, g, 500*time.Millisecond)
	if !g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("re-aggro after cooldown should start")
	}
}

func TestAggroOnAggroOnceMatchesManual(t *testing.T) {
	g, facade := newHookTestGame(t)
	now := useAggroClock(t, g)
	storeHookScript(t, facade, "SCR-SIGHT", `tales.game.msgToRoom(ctx.roomId, "The chanter hits the drum once.")`)
	trait := &npc.EnemyTrait{OnAggroScript: "SCR-SIGHT", AttackPower: 1}
	hero, _ := placeAggro(t, g, facade, "R-hook-sight", "chanter-sight", "Chanter", 2, trait)
	online(g, facade, hero)
	g.NotePlayerEntered(hero.ID, "R-hook-sight")
	advanceAggro(now, g, 500*time.Millisecond)
	msgs := drainGameMessages(g.SendMessage())
	if got := countRoomText(msgs, "The chanter hits the drum once."); got != 1 {
		t.Fatalf("aggro onAggro = %d", got)
	}
	if _, ok := spotsYou(msgs); !ok {
		t.Fatal("aggro path missing spots line")
	}
	g.CombatController.EndCombatForPlayer(hero.ID)
	drainGameMessages(g.SendMessage())

	_, hero2 := sharePlayer(t, facade, "user-manual", "ref-manual", "char-manual", "Hero", "R-hook-sight")
	manual := aggroEnemy("chanter-manual", "Chanter", 2, &npc.EnemyTrait{OnAggroScript: "SCR-SIGHT", AttackPower: 1, AggroOnSight: false})
	g.NPCManager.RegisterExistingNPC(manual, "R-hook-sight")
	user2, _ := facade.UsersService().FindByID(hero2.BelongsUserID)
	if g.CombatController.BeginEngagement("R-hook-sight", hero2, user2.ID, manual, false) == nil {
		t.Fatal("manual engage did not start")
	}
	manualMsgs := drainGameMessages(g.SendMessage())
	if got := countRoomText(manualMsgs, "The chanter hits the drum once."); got != 1 {
		t.Fatalf("manual onAggro = %d", got)
	}
	if _, ok := spotsYou(manualMsgs); ok {
		t.Fatal("manual engage posted a spots line")
	}
}

func TestAggroPartyNudgeMatchesManual(t *testing.T) {
	g, facade := newHookTestGame(t)
	now := useAggroClock(t, g)
	storeTestRoom(t, facade, "R-party", nil)
	_, hero := sharePlayer(t, facade, "user-lead", "ref-lead", "char-lead", "Hero", "R-party")
	_, ally := sharePlayer(t, facade, "user-ally", "ref-ally", "char-ally", "Ally", "R-party")
	online(g, facade, hero)
	online(g, facade, ally)
	shareParty(t, facade, "Watch", hero.ID, hero.ID, ally.ID)
	enemy := aggroEnemy("wolf-party", "Wolf", 2, nil)
	g.NPCManager.RegisterExistingNPC(enemy, "R-party")
	g.NotePlayerEntered(hero.ID, "R-party")
	advanceAggro(now, g, 500*time.Millisecond)
	aggroMsgs := drainGameMessages(g.SendMessage())
	got := partyLine(aggroMsgs)
	want := "[Party] Hero engaged Wolf nearby! Type 'attack Wolf' to join the fight."
	if got != want {
		t.Fatalf("aggro party line = %q", got)
	}
	if g.CombatController.IsPlayerInCombat(ally.ID) {
		t.Fatal("party member was pulled into the fight")
	}

	g.CombatController.EndCombatForPlayer(hero.ID)
	drainGameMessages(g.SendMessage())
	storeTestRoom(t, facade, "R-party-manual", nil)
	_, hero2 := sharePlayer(t, facade, "user-lead2", "ref-lead2", "char-lead2", "Hero", "R-party-manual")
	_, ally2 := sharePlayer(t, facade, "user-ally2", "ref-ally2", "char-ally2", "Ally", "R-party-manual")
	online(g, facade, hero2)
	online(g, facade, ally2)
	shareParty(t, facade, "Watch2", hero2.ID, hero2.ID, ally2.ID)
	manual := aggroEnemy("wolf-manual", "Wolf", 2, &npc.EnemyTrait{AggroOnSight: false})
	g.NPCManager.RegisterExistingNPC(manual, "R-party-manual")
	user2, _ := facade.UsersService().FindByID(hero2.BelongsUserID)
	if g.CombatController.BeginEngagement("R-party-manual", hero2, user2.ID, manual, false) == nil {
		t.Fatal("manual engage did not start")
	}
	if partyLine(drainGameMessages(g.SendMessage())) != want {
		t.Fatal("manual party line differs from aggro")
	}
}

func TestAggroCooldownCoversRoomBystander(t *testing.T) {
	g, facade := newHookTestGame(t)
	now := useAggroClock(t, g)
	hero, _ := placeAggro(t, g, facade, "R-pair-cd", "wolf-fought", "Wolf", 2, &npc.EnemyTrait{AttackPower: 0})
	bystander := aggroEnemy("wolf-bystander", "Other Wolf", 2, nil)
	g.NPCManager.RegisterExistingNPC(bystander, "R-pair-cd")
	online(g, facade, hero)
	g.NotePlayerEntered(hero.ID, "R-pair-cd")
	advanceAggro(now, g, 500*time.Millisecond)
	if !g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("setup fight did not start")
	}
	g.CombatController.EndCombatForPlayer(hero.ID)
	drainGameMessages(g.SendMessage())

	g.NotePlayerEntered(hero.ID, "R-pair-cd")
	advanceAggro(now, g, 500*time.Millisecond)
	if g.CombatController.IsPlayerInCombat(hero.ID) || g.CombatController.IsNPCInCombat(bystander.ID) {
		t.Fatal("room bystander re-aggroed during cooldown")
	}
}

func TestAggroTwoNPCsOneFight(t *testing.T) {
	g, facade := newHookTestGame(t)
	now := useAggroClock(t, g)
	hero, first := placeAggro(t, g, facade, "R-pair", "wolf-1", "Wolf", 2, nil)
	second := aggroEnemy("wolf-2", "Second Wolf", 2, nil)
	g.NPCManager.RegisterExistingNPC(second, "R-pair")
	online(g, facade, hero)
	g.NotePlayerEntered(hero.ID, "R-pair")
	advanceAggro(now, g, 500*time.Millisecond)
	inst := g.CombatController.GetCombatInstance(hero.ID)
	if inst == nil {
		t.Fatal("no fight")
	}
	in := 0
	if g.CombatController.IsNPCInCombat(first.ID) {
		in++
	}
	if g.CombatController.IsNPCInCombat(second.ID) {
		in++
	}
	if in != 1 || len(inst.Enemies) != 1 {
		t.Fatalf("npcs in combat=%d enemies=%d", in, len(inst.Enemies))
	}
}

func TestAggroNPCSpawnIntoOccupiedRoom(t *testing.T) {
	g, facade := newHookTestGame(t)
	now := useAggroClock(t, g)
	storeTestRoom(t, facade, "R-spawn", nil)
	_, hero := sharePlayer(t, facade, "user-spawn", "ref-spawn", "char-spawn", "Hero", "R-spawn")
	online(g, facade, hero)
	enemy := aggroEnemy("wolf-spawn", "Wolf", 2, nil)
	g.NPCManager.RegisterExistingNPC(enemy, "R-spawn")
	g.NoteNPCAppeared(enemy.ID, "R-spawn")
	g.FireDueAggro()
	if g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("spawn aggro fired before grace")
	}
	advanceAggro(now, g, 500*time.Millisecond)
	if !g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("spawn into an occupied room did not aggro")
	}
}
