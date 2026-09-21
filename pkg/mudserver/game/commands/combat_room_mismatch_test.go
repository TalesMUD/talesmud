package commands_test

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func TestTakeExitBlockedWhileInCombat(t *testing.T) {
	g, facade := newSocialTestGame(t)
	user, character := storeSocialPlayer(t, facade, "user-c1", "ref-c1", "char-c1", "Fighter", "room-fight")

	oldExits := rooms.Exits{{Name: "north", Target: "room-away"}}
	oldChars := rooms.Characters{character.ID}
	fightRoom := &rooms.Room{
		Entity:     &entities.Entity{ID: "room-fight"},
		Name:       "Fight Room",
		Exits:      &oldExits,
		Characters: &oldChars,
	}
	awayExits := rooms.Exits{}
	awayChars := rooms.Characters{}
	awayRoom := &rooms.Room{
		Entity:     &entities.Entity{ID: "room-away"},
		Name:       "Away Room",
		Exits:      &awayExits,
		Characters: &awayChars,
	}
	if _, err := facade.RoomsService().Import(fightRoom); err != nil {
		t.Fatal(err)
	}
	if _, err := facade.RoomsService().Import(awayRoom); err != nil {
		t.Fatal(err)
	}

	enemy := &npc.NPC{
		Entity:           &entities.Entity{ID: "enemy-1"},
		Name:             "Catacomb Rat",
		CurrentHitPoints: 20,
		MaxHitPoints:     20,
		EnemyTrait:       &npc.EnemyTrait{AttackPower: 1, Difficulty: "trivial"},
	}
	enemy.CurrentRoomID = "room-fight"
	g.NPCManager.RegisterExistingNPC(enemy, "room-fight")

	g.ConnectUserSession(user)
	g.SetUserSessionCharacter(user, character)

	instance := g.GetCombatEngine().InitiateCombat("room-fight", []*characters.Character{character}, []*npc.NPC{enemy})
	if instance == nil {
		t.Fatal("failed to initiate combat")
	}
	character.InCombat = true
	character.CombatInstanceID = instance.ID
	_ = facade.CharactersService().Update(character.ID, character)

	_ = drainSocialMessages(g.SendMessage())
	msg := &messages.Message{FromUser: user, Character: character, Data: "north"}
	if !commands.TakeExit("north")(fightRoom, g, msg) {
		t.Fatal("take exit should handle (block) movement")
	}

	fresh, err := facade.CharactersService().FindByID(character.ID)
	if err != nil {
		t.Fatal(err)
	}
	if fresh.CurrentRoomID != "room-fight" {
		t.Fatalf("expected to stay in fight room, got %q", fresh.CurrentRoomID)
	}

	var sawBlock bool
	for _, out := range drainSocialMessages(g.SendMessage()) {
		if rsp, ok := out.(messages.MessageResponse); ok && strings.Contains(rsp.Message, "can't leave while in combat") {
			sawBlock = true
		}
	}
	if !sawBlock {
		t.Fatal("expected combat leave block message")
	}
}

func TestAttackAbandonsOrphanedCombatFromOtherRoom(t *testing.T) {
	g, facade := newSocialTestGame(t)
	user, character := storeSocialPlayer(t, facade, "user-c2", "ref-c2", "char-c2", "Wanderer", "room-bunker")

	bunkerExits := rooms.Exits{}
	bunkerChars := rooms.Characters{character.ID}
	bunker := &rooms.Room{
		Entity:     &entities.Entity{ID: "room-bunker"},
		Name:       "Burrow Bunker",
		Exits:      &bunkerExits,
		Characters: &bunkerChars,
	}
	if _, err := facade.RoomsService().Import(bunker); err != nil {
		t.Fatal(err)
	}

	rat := &npc.NPC{
		Entity:           &entities.Entity{ID: "rat-1"},
		Name:             "Catacomb Rat",
		CurrentHitPoints: 20,
		MaxHitPoints:     20,
		EnemyTrait:       &npc.EnemyTrait{AttackPower: 1, Difficulty: "trivial"},
	}
	rat.CurrentRoomID = "room-nest"
	g.NPCManager.RegisterExistingNPC(rat, "room-nest")

	brute := &npc.NPC{
		Entity:           &entities.Entity{ID: "brute-1"},
		Name:             "Burrow Brute",
		CurrentHitPoints: 55,
		MaxHitPoints:     55,
		EnemyTrait:       &npc.EnemyTrait{AttackPower: 8, Difficulty: "boss"},
	}
	brute.CurrentRoomID = "room-bunker"
	g.NPCManager.RegisterExistingNPC(brute, "room-bunker")

	g.ConnectUserSession(user)
	g.SetUserSessionCharacter(user, character)

	// Orphaned combat: started in nest, player is now in bunker
	old := g.GetCombatEngine().InitiateCombat("room-nest", []*characters.Character{character}, []*npc.NPC{rat})
	if old == nil {
		t.Fatal("failed to initiate orphaned combat")
	}
	character.InCombat = true
	character.CombatInstanceID = old.ID
	character.CurrentRoomID = "room-bunker"
	_ = facade.CharactersService().Update(character.ID, character)

	_ = drainSocialMessages(g.SendMessage())
	msg := &messages.Message{FromUser: user, Character: character, Data: "attack Burrow Brute"}
	if !(&commands.AttackCommand{}).Execute(g, msg) {
		t.Fatal("attack did not handle")
	}

	if g.GetCombatEngine().IsPlayerInCombat(character.ID) {
		inst := g.GetCombatEngine().GetCombatInstance(character.ID)
		if inst == nil {
			t.Fatal("player marked in combat but no instance")
		}
		if inst.OriginRoomID != "room-bunker" {
			t.Fatalf("expected new combat in bunker, got origin %q", inst.OriginRoomID)
		}
		if len(inst.Enemies) != 1 || inst.Enemies[0].Name != "Burrow Brute" {
			t.Fatalf("expected Burrow Brute in new combat, got %#v", inst.Enemies)
		}
	} else {
		t.Fatal("expected player to be in new combat with Burrow Brute")
	}

	var sawBreakOff, sawInvalid bool
	for _, out := range drainSocialMessages(g.SendMessage()) {
		if rsp, ok := out.(messages.MessageResponse); ok {
			if strings.Contains(rsp.Message, "previous fight is too far away") {
				sawBreakOff = true
			}
			if strings.Contains(rsp.Message, "Invalid target") {
				sawInvalid = true
			}
		}
	}
	if !sawBreakOff {
		t.Fatal("expected orphaned-combat break-off message")
	}
	if sawInvalid {
		t.Fatal("must not report Invalid target Burrow Brute when Brute is in room")
	}
}
