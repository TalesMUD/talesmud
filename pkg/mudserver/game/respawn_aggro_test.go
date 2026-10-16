package game

import (
	"strings"
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/ruleset"
	"github.com/talesmud/talesmud/pkg/service"
)

func quietAggro(t *testing.T, g *Game, respawn, post, spawn time.Duration) *time.Time {
	t.Helper()
	now := useAggroClock(t, g)
	ruleset.SetAggroOnSight(true, 500*time.Millisecond, 5, 0)
	ruleset.SetAggroQuiet(respawn, post, spawn)
	return now
}

func TestUpdateSpawnerRespectsRespawnFloor(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	if ruleset.RespawnMin() != 30*time.Second {
		t.Fatalf("floor = %s", ruleset.RespawnMin())
	}
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "R-floor", nil)
	if _, err := facade.NPCsService().Import(&npc.NPC{
		Entity:           &entities.Entity{ID: "TPL-FLOOR"},
		Name:             "Grave Warden",
		IsTemplate:       true,
		Level:            1,
		MaxHitPoints:     10,
		CurrentHitPoints: 10,
		EnemyTrait:       &npc.EnemyTrait{AggroOnSight: true, Difficulty: "hard", AttackPower: 1},
	}); err != nil {
		t.Fatal(err)
	}
	spawner := &npc.NPCSpawner{
		Entity:        &entities.Entity{ID: "SPW-FLOOR"},
		TemplateID:    "TPL-FLOOR",
		RoomID:        "R-floor",
		MaxInstances:  1,
		InitialCount:  1,
		SpawnInterval: 10 * time.Minute,
	}
	g.updateSpawner(spawner)
	if g.NPCManager.CountAliveForSpawner("SPW-FLOOR") != 1 {
		t.Fatal("server start should fill InitialCount with no death yet")
	}
	id := livingSpawnerID(t, g, "R-floor")
	if !g.NPCManager.DamageInstance(id, 100) {
		t.Fatal("kill failed")
	}
	g.updateSpawner(spawner)
	if g.NPCManager.CountAliveForSpawner("SPW-FLOOR") != 0 {
		t.Fatal("respawn floor should hold a spawner with no authored delay")
	}
	g.NPCManager.GetSpawnerState("SPW-FLOOR").LastDeathTime = time.Now().Add(-31 * time.Second)
	g.updateSpawner(spawner)
	if g.NPCManager.CountAliveForSpawner("SPW-FLOOR") != 1 {
		t.Fatal("spawner should replace the corpse once the floor has elapsed")
	}
	if got := livingSpawnerID(t, g, "R-floor"); got == id {
		t.Fatal("replacement should be a new instance")
	}
}

func TestUpdateSpawnerRespawnUsesRespawnGrace(t *testing.T) {
	g, facade := newHookTestGame(t)
	now := quietAggro(t, g, 8*time.Second, 0, 0)
	storeTestRoom(t, facade, "R-wire", nil)
	if _, err := facade.NPCsService().Import(&npc.NPC{
		Entity:           &entities.Entity{ID: "TPL-WIRE"},
		Name:             "Grave Warden",
		IsTemplate:       true,
		Level:            1,
		MaxHitPoints:     10,
		CurrentHitPoints: 10,
		EnemyTrait:       &npc.EnemyTrait{AggroOnSight: true, Difficulty: "hard", AttackPower: 1},
	}); err != nil {
		t.Fatal(err)
	}
	spawner := &npc.NPCSpawner{
		Entity:        &entities.Entity{ID: "SPW-WIRE"},
		TemplateID:    "TPL-WIRE",
		RoomID:        "R-wire",
		MaxInstances:  1,
		InitialCount:  1,
		SpawnInterval: 10 * time.Minute,
	}
	g.updateSpawner(spawner)
	_, hero := sharePlayer(t, facade, "user-wire", "ref-wire", "char-wire", "Hero", "R-wire")
	online(g, facade, hero)

	id := livingSpawnerID(t, g, "R-wire")
	if !g.NPCManager.DamageInstance(id, 100) {
		t.Fatal("kill failed")
	}
	g.updateSpawner(spawner)
	if g.NPCManager.CountAliveForSpawner("SPW-WIRE") != 0 {
		t.Fatal("floor should hold the replacement")
	}
	g.NPCManager.GetSpawnerState("SPW-WIRE").LastDeathTime = time.Now().Add(-31 * time.Second)
	g.updateSpawner(spawner)
	if g.NPCManager.CountAliveForSpawner("SPW-WIRE") != 1 {
		t.Fatal("expected a replacement")
	}
	advanceAggro(now, g, 500*time.Millisecond)
	if g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("a death replacement used the short sight grace")
	}
	advanceAggro(now, g, 7500*time.Millisecond)
	if !g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("respawn grace did not engage the player already in the room")
	}
	if got := g.NPCManager.GetInstance(livingSpawnerID(t, g, "R-wire")).SpawnerID; got != "SPW-WIRE" {
		t.Fatalf("SpawnerID = %q", got)
	}
}

func TestRespawnGraceHoldsOccupiedRoom(t *testing.T) {
	g, facade := newHookTestGame(t)
	now := quietAggro(t, g, 8*time.Second, 0, 0)
	hero, enemy := placeAggro(t, g, facade, "R-hold", "wolf-hold", "Wolf", 1, nil)
	online(g, facade, hero)
	g.NoteNPCRespawned(enemy.ID, "R-hold")
	advanceAggro(now, g, 500*time.Millisecond)
	if g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("respawn aggro fired on the short grace")
	}
	advanceAggro(now, g, 7500*time.Millisecond)
	if !g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("respawn grace did not fire")
	}
}

func TestWalkInAfterRespawnKeepsShortGrace(t *testing.T) {
	g, facade := newHookTestGame(t)
	now := quietAggro(t, g, 8*time.Second, 0, 0)
	storeTestRoom(t, facade, "R-away", nil)
	hero, enemy := placeAggro(t, g, facade, "R-return", "wolf-return", "Wolf", 1, nil)
	online(g, facade, hero)
	if err := facade.CharactersService().Modify(hero.ID, func(ch *characters.Character) error {
		ch.CurrentRoomID = "R-away"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	g.NoteNPCRespawned(enemy.ID, "R-return")
	advanceAggro(now, g, 8*time.Second)
	if g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("respawn aggroed a player who was in another room")
	}
	if err := facade.CharactersService().Modify(hero.ID, func(ch *characters.Character) error {
		ch.CurrentRoomID = "R-return"
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	g.NotePlayerEntered(hero.ID, "R-return")
	advanceAggro(now, g, 500*time.Millisecond)
	if !g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("walking into the room should use the short grace")
	}
}

func TestWinThenMove(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	g, facade := newHookTestGame(t)
	now := quietAggro(t, g, 8*time.Second, 10*time.Second, 0)
	hero, enemy := placeAggro(t, g, facade, "R-win", "wolf-win", "Wolf", 1, &npc.EnemyTrait{AttackPower: 0, Difficulty: "easy"})
	online(g, facade, hero)
	setRoomExits(t, facade, "R-win", rooms.Exits{{Name: "north", Target: "R-after-win"}})
	storeTestRoom(t, facade, "R-after-win", nil)
	g.NotePlayerEntered(hero.ID, "R-win")
	advanceAggro(now, g, 500*time.Millisecond)
	endAggroVictory(t, g, hero.ID, enemy.ID)

	born := aggroEnemy("wolf-born", "Warden", 1, nil)
	g.NPCManager.RegisterExistingNPC(born, "R-win")
	g.NoteNPCRespawned(born.ID, "R-win")
	advanceAggro(now, g, 8*time.Second)
	if g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("fresh respawn aggroed during the post-combat window")
	}

	next := aggroEnemy("wolf-after-win", "Timber Wolf", 1, nil)
	g.NPCManager.RegisterExistingNPC(next, "R-after-win")
	if !takeNorth(t, g, facade, hero.ID, "R-win") {
		t.Fatal("victory left the player unable to move")
	}
	moved, err := facade.CharactersService().FindByID(hero.ID)
	if err != nil {
		t.Fatal(err)
	}
	if moved.CurrentRoomID != "R-after-win" {
		t.Fatalf("room = %s", moved.CurrentRoomID)
	}
	if moved.InCombat {
		t.Fatal("stale InCombat survived the exit")
	}
	advanceAggro(now, g, 500*time.Millisecond)
	if !g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("the next room should aggro on the short grace")
	}
}

func TestFleeThenMove(t *testing.T) {
	g, facade := newHookTestGame(t)
	now := quietAggro(t, g, 8*time.Second, 10*time.Second, 0)
	hero, enemy := placeAggro(t, g, facade, "R-flee-move", "wolf-flee-move", "Wolf", 1, &npc.EnemyTrait{AttackPower: 0, Difficulty: "easy"})
	online(g, facade, hero)
	setRoomExits(t, facade, "R-flee-move", rooms.Exits{{Name: "north", Target: "R-after-flee"}})
	storeTestRoom(t, facade, "R-after-flee", nil)
	g.NotePlayerEntered(hero.ID, "R-flee-move")
	advanceAggro(now, g, 500*time.Millisecond)
	if !fleeFight(t, g, hero.ID) {
		t.Fatal("flee did not end the fight")
	}

	born := aggroEnemy("wolf-flee-born", "Warden", 1, nil)
	g.NPCManager.RegisterExistingNPC(born, "R-flee-move")
	g.NoteNPCRespawned(born.ID, "R-flee-move")
	advanceAggro(now, g, 8*time.Second)
	if g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("fresh respawn aggroed after flee")
	}
	_ = enemy

	next := aggroEnemy("wolf-after-flee", "Timber Wolf", 1, nil)
	g.NPCManager.RegisterExistingNPC(next, "R-after-flee")
	if !takeNorth(t, g, facade, hero.ID, "R-flee-move") {
		t.Fatal("flee left the player unable to move")
	}
	moved, err := facade.CharactersService().FindByID(hero.ID)
	if err != nil {
		t.Fatal(err)
	}
	if moved.CurrentRoomID != "R-after-flee" {
		t.Fatalf("room = %s", moved.CurrentRoomID)
	}
	advanceAggro(now, g, 500*time.Millisecond)
	if !g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("the next room should aggro on the short grace")
	}
}

func TestSpawnCooldownBlocksSamePlayer(t *testing.T) {
	g, facade := newHookTestGame(t)
	now := quietAggro(t, g, 8*time.Second, 0, 60*time.Second)
	hero, enemy := placeAggro(t, g, facade, "R-same", "wolf-same", "Wolf", 1, &npc.EnemyTrait{AttackPower: 0, Difficulty: "easy"})
	enemy.SpawnerID = "SPW-SAME"
	online(g, facade, hero)
	g.NotePlayerEntered(hero.ID, "R-same")
	advanceAggro(now, g, 500*time.Millisecond)
	if !g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("setup fight did not start")
	}
	if !fleeFight(t, g, hero.ID) {
		t.Fatal("flee did not end the fight")
	}
	g.NPCManager.RemoveInstance(enemy.ID)

	next := aggroEnemy("wolf-same-2", "Wolf", 1, nil)
	next.SpawnerID = "SPW-SAME"
	g.NPCManager.RegisterExistingNPC(next, "R-same")
	g.NoteNPCRespawned(next.ID, "R-same")
	advanceAggro(now, g, 8*time.Second)
	if g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("same player was auto-aggroed inside the spawn cooldown")
	}
	advanceAggro(now, g, 50*time.Second)
	if g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("spawn cooldown elapsed early")
	}

	g.CombatController.ClearCombatGrace(hero.ID)
	fresh, err := facade.CharactersService().FindByID(hero.ID)
	if err != nil {
		t.Fatal(err)
	}
	user, err := facade.UsersService().FindByID(fresh.BelongsUserID)
	if err != nil {
		t.Fatal(err)
	}
	msg := &messages.Message{FromUser: user, Character: fresh, Data: "attack Wolf"}
	if !(&commands.AttackCommand{}).Execute(g, msg) {
		t.Fatal("attack was not handled")
	}
	if !g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("manual attack should ignore the spawn cooldown")
	}
}

func TestSpawnCooldownLetsAnotherPlayer(t *testing.T) {
	g, facade := newHookTestGame(t)
	now := quietAggro(t, g, 8*time.Second, 0, 60*time.Second)
	hero, enemy := placeAggro(t, g, facade, "R-other", "wolf-other", "Wolf", 1, &npc.EnemyTrait{AttackPower: 0, Difficulty: "easy"})
	enemy.SpawnerID = "SPW-OTHER"
	_, other := sharePlayer(t, facade, "user-other", "ref-other", "char-other", "Ally", "R-other")
	online(g, facade, hero)
	online(g, facade, other)
	g.NotePlayerEntered(hero.ID, "R-other")
	advanceAggro(now, g, 500*time.Millisecond)
	if !g.CombatController.IsPlayerInCombat(hero.ID) || g.CombatController.IsPlayerInCombat(other.ID) {
		t.Fatal("setup should aggro only the entering player")
	}
	if !fleeFight(t, g, hero.ID) {
		t.Fatal("flee did not end the fight")
	}
	g.NPCManager.RemoveInstance(enemy.ID)

	next := aggroEnemy("wolf-other-2", "Wolf", 1, nil)
	next.SpawnerID = "SPW-OTHER"
	g.NPCManager.RegisterExistingNPC(next, "R-other")
	g.NoteNPCRespawned(next.ID, "R-other")
	advanceAggro(now, g, 8*time.Second)
	if g.CombatController.IsPlayerInCombat(hero.ID) {
		t.Fatal("the player who was already aggroed was picked again")
	}
	if !g.CombatController.IsPlayerInCombat(other.ID) {
		t.Fatal("a different player in the room should still be aggroed")
	}
}

func livingSpawnerID(t *testing.T, g *Game, roomID string) string {
	t.Helper()
	for _, inst := range g.NPCManager.GetInstancesInRoom(roomID) {
		if inst != nil && inst.Entity != nil && !inst.IsDead {
			return inst.Entity.ID
		}
	}
	t.Fatalf("no living npc in %s", roomID)
	return ""
}

func setRoomExits(t *testing.T, facade service.Facade, roomID string, exits rooms.Exits) {
	t.Helper()
	room, err := facade.RoomsService().FindByID(roomID)
	if err != nil || room == nil {
		t.Fatalf("room %s: %v", roomID, err)
	}
	room.Exits = &exits
	if err := facade.RoomsService().Update(roomID, room); err != nil {
		t.Fatal(err)
	}
}

func endAggroVictory(t *testing.T, g *Game, playerID, enemyID string) {
	t.Helper()
	inst := g.CombatController.GetCombatInstance(playerID)
	if inst == nil {
		t.Fatal("no fight to win")
	}
	ref := inst.GetEnemyByID(enemyID)
	if ref == nil {
		t.Fatal("enemy missing from the fight")
	}
	ref.IsAlive = false
	ref.CurrentHP = 0
	g.CombatController.engine.UpdateCombatant(inst, ref)
	g.CombatController.engine.EndCombat(inst, combat.CombatStateVictory)
	g.CombatController.cleanupCombatInstance(inst, combat.CombatStateVictory)
	if g.CombatController.IsPlayerInCombat(playerID) {
		t.Fatal("victory left the player in combat")
	}
}

func fleeFight(t *testing.T, g *Game, playerID string) bool {
	t.Helper()
	fled := false
	for i := 0; i < 40 && g.CombatController.IsPlayerInCombat(playerID); i++ {
		ok, _, _, _ := g.CombatController.ProcessPlayerFlee(playerID)
		if ok {
			fled = true
		}
	}
	return fled && !g.CombatController.IsPlayerInCombat(playerID)
}

func takeNorth(t *testing.T, g *Game, facade service.Facade, playerID, roomID string) bool {
	t.Helper()
	fresh, err := facade.CharactersService().FindByID(playerID)
	if err != nil || fresh == nil {
		t.Fatal(err)
	}
	fresh.InCombat = true
	if err := facade.CharactersService().Update(playerID, fresh); err != nil {
		t.Fatal(err)
	}
	user, err := facade.UsersService().FindByID(fresh.BelongsUserID)
	if err != nil || user == nil {
		t.Fatal(err)
	}
	room, err := facade.RoomsService().FindByID(roomID)
	if err != nil || room == nil {
		t.Fatal(err)
	}
	_ = drainGameMessages(g.SendMessage())
	msg := &messages.Message{FromUser: user, Character: fresh, Data: "north"}
	handled := commands.TakeExit("north")(room, g, msg)
	for _, out := range drainGameMessages(g.SendMessage()) {
		rsp, ok := out.(messages.MessageResponse)
		if ok && strings.Contains(rsp.Message, "can't leave while in combat") {
			t.Fatal("movement was blocked after the fight")
		}
	}
	return handled
}
