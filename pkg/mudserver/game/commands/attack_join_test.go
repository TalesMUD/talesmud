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

func TestAttackJoinsExistingSameRoomCombat(t *testing.T) {
	g, facade := newSocialTestGame(t)
	userA, charA := storeSocialPlayer(t, facade, "user-a", "ref-a", "char-a", "Aryn", "room-fight")
	userB, charB := storeSocialPlayer(t, facade, "user-b", "ref-b", "char-b", "Bran", "room-fight")

	exits := rooms.Exits{}
	chars := rooms.Characters{charA.ID, charB.ID}
	fightRoom := &rooms.Room{
		Entity:     &entities.Entity{ID: "room-fight"},
		Name:       "Fight Room",
		Exits:      &exits,
		Characters: &chars,
	}
	if _, err := facade.RoomsService().Import(fightRoom); err != nil {
		t.Fatal(err)
	}

	enemy := &npc.NPC{
		Entity:           &entities.Entity{ID: "enemy-join-1"},
		Name:             "Catacomb Rat",
		CurrentHitPoints: 40,
		MaxHitPoints:     40,
		Level:            1,
		EnemyTrait:       &npc.EnemyTrait{AttackPower: 1, Defense: 0, Difficulty: "trivial"},
	}
	enemy.CurrentRoomID = "room-fight"
	g.NPCManager.RegisterExistingNPC(enemy, "room-fight")

	g.ConnectUserSession(userA)
	g.SetUserSessionCharacter(userA, charA)
	g.ConnectUserSession(userB)
	g.SetUserSessionCharacter(userB, charB)

	if !(&commands.AttackCommand{}).Execute(g, &messages.Message{FromUser: userA, Character: charA, Data: "attack Catacomb Rat"}) {
		t.Fatal("A attack did not handle")
	}
	_ = drainSocialMessages(g.SendMessage())

	inst := g.GetCombatEngine().GetCombatInstance(charA.ID)
	if inst == nil {
		t.Fatal("expected combat instance after A engages")
	}
	if len(inst.Players) != 1 {
		t.Fatalf("expected 1 player before join, got %d", len(inst.Players))
	}

	if !(&commands.AttackCommand{}).Execute(g, &messages.Message{FromUser: userB, Character: charB, Data: "attack Catacomb Rat"}) {
		t.Fatal("B attack-join did not handle")
	}

	var sawCombatStart bool
	var sawRefuse bool
	for _, out := range drainSocialMessages(g.SendMessage()) {
		switch msg := out.(type) {
		case messages.MessageResponse:
			if strings.Contains(msg.Message, "already in combat with someone else") {
				sawRefuse = true
			}
		case *messages.CombatStartMessage:
			if msg.AudienceID == userB.ID {
				sawCombatStart = true
			}
		}
	}
	if sawRefuse {
		t.Fatal("same-room join refused with legacy someone-else message")
	}
	if !sawCombatStart {
		t.Fatal("expected combatStart for joining player B")
	}

	inst = g.GetCombatEngine().GetCombatInstance(charB.ID)
	if inst == nil {
		t.Fatal("B should be in combat after join")
	}
	if len(inst.Players) != 2 {
		t.Fatalf("expected 2 players in instance after join, got %d", len(inst.Players))
	}
	if inst.GetPlayerByID(charA.ID) == nil || inst.GetPlayerByID(charB.ID) == nil {
		t.Fatal("both A and B must appear in instance.Players")
	}

	freshB, err := facade.CharactersService().FindByID(charB.ID)
	if err != nil || freshB == nil || !freshB.InCombat || freshB.CombatInstanceID != inst.ID {
		t.Fatalf("expected B InCombat with instance id, got %#v", freshB)
	}
}

func TestAttackJoinRefusedFromOtherRoom(t *testing.T) {
	g, facade := newSocialTestGame(t)
	userA, charA := storeSocialPlayer(t, facade, "user-a2", "ref-a2", "char-a2", "Aryn", "room-fight")
	userB, charB := storeSocialPlayer(t, facade, "user-b2", "ref-b2", "char-b2", "Bran", "room-away")

	exits := rooms.Exits{}
	fightRoom := &rooms.Room{
		Entity:     &entities.Entity{ID: "room-fight"},
		Name:       "Fight Room",
		Exits:      &exits,
		Characters: &rooms.Characters{charA.ID},
	}
	awayRoom := &rooms.Room{
		Entity:     &entities.Entity{ID: "room-away"},
		Name:       "Away Room",
		Exits:      &exits,
		Characters: &rooms.Characters{charB.ID},
	}
	if _, err := facade.RoomsService().Import(fightRoom); err != nil {
		t.Fatal(err)
	}
	if _, err := facade.RoomsService().Import(awayRoom); err != nil {
		t.Fatal(err)
	}

	enemy := &npc.NPC{
		Entity:           &entities.Entity{ID: "enemy-join-2"},
		Name:             "Catacomb Rat",
		CurrentHitPoints: 40,
		MaxHitPoints:     40,
		Level:            1,
		EnemyTrait:       &npc.EnemyTrait{AttackPower: 1, Defense: 0, Difficulty: "trivial"},
	}
	enemy.CurrentRoomID = "room-fight"
	g.NPCManager.RegisterExistingNPC(enemy, "room-fight")

	g.ConnectUserSession(userA)
	g.SetUserSessionCharacter(userA, charA)
	g.ConnectUserSession(userB)
	g.SetUserSessionCharacter(userB, charB)

	inst := g.GetCombatEngine().InitiateCombat("room-fight", []*characters.Character{charA}, []*npc.NPC{enemy})
	if inst == nil {
		t.Fatal("failed to initiate")
	}
	charA.InCombat = true
	charA.CombatInstanceID = inst.ID
	_ = facade.CharactersService().Update(charA.ID, charA)
	enemy.InCombat = true
	enemy.CombatInstanceID = inst.ID

	// Same NPC instance discoverable in B's room, but OriginRoomID is still room-fight.
	g.NPCManager.RegisterExistingNPC(enemy, "room-away")
	charB.CurrentRoomID = "room-away"
	_ = facade.CharactersService().Update(charB.ID, charB)

	_ = drainSocialMessages(g.SendMessage())
	if !(&commands.AttackCommand{}).Execute(g, &messages.Message{FromUser: userB, Character: charB, Data: "attack Catacomb Rat"}) {
		t.Fatal("B attack did not handle")
	}

	var sawRefuse bool
	for _, out := range drainSocialMessages(g.SendMessage()) {
		if rsp, ok := out.(messages.MessageResponse); ok && strings.Contains(rsp.Message, "already in combat with someone else") {
			sawRefuse = true
		}
	}
	if !sawRefuse {
		t.Fatal("expected cross-room join refusal")
	}
	if g.GetCombatEngine().IsPlayerInCombat(charB.ID) {
		t.Fatal("B must not be in combat after refused cross-room join")
	}
	if len(inst.Players) != 1 {
		t.Fatalf("expected still 1 player in instance, got %d", len(inst.Players))
	}
}

func TestVictoryAwardsXPToBothLivingJoiners(t *testing.T) {
	g, facade := newSocialTestGame(t)
	userA, charA := storeSocialPlayer(t, facade, "user-ax", "ref-ax", "char-ax", "Aryn", "room-fight")
	userB, charB := storeSocialPlayer(t, facade, "user-bx", "ref-bx", "char-bx", "Bran", "room-fight")

	exits := rooms.Exits{}
	fightRoom := &rooms.Room{
		Entity:     &entities.Entity{ID: "room-fight"},
		Name:       "Fight Room",
		Exits:      &exits,
		Characters: &rooms.Characters{charA.ID, charB.ID},
	}
	if _, err := facade.RoomsService().Import(fightRoom); err != nil {
		t.Fatal(err)
	}

	enemy := &npc.NPC{
		Entity:           &entities.Entity{ID: "enemy-xp-1"},
		Name:             "Weak Rat",
		CurrentHitPoints: 1,
		MaxHitPoints:     1,
		Level:            2,
		EnemyTrait:       &npc.EnemyTrait{AttackPower: 0, Defense: 0, Difficulty: "trivial", XPReward: 20},
	}
	enemy.CurrentRoomID = "room-fight"
	g.NPCManager.RegisterExistingNPC(enemy, "room-fight")

	g.ConnectUserSession(userA)
	g.SetUserSessionCharacter(userA, charA)
	g.ConnectUserSession(userB)
	g.SetUserSessionCharacter(userB, charB)

	xpBeforeA := charA.XP
	xpBeforeB := charB.XP

	if !(&commands.AttackCommand{}).Execute(g, &messages.Message{FromUser: userA, Character: charA, Data: "attack Weak Rat"}) {
		t.Fatal("A engage failed")
	}
	_ = drainSocialMessages(g.SendMessage())
	if !(&commands.AttackCommand{}).Execute(g, &messages.Message{FromUser: userB, Character: charB, Data: "attack Weak Rat"}) {
		t.Fatal("B join failed")
	}
	_ = drainSocialMessages(g.SendMessage())

	inst := g.GetCombatEngine().GetCombatInstance(charA.ID)
	if inst == nil || len(inst.GetLivingPlayers()) != 2 {
		t.Fatalf("expected 2 living players before victory, got %#v", inst)
	}

	var ended bool
	var endState string
	for i := 0; i < 20; i++ {
		_, combatEnded, state := g.GetCombatEngine().ProcessPlayerAttack(charA.ID, enemy.Entity.ID)
		if combatEnded {
			ended = true
			endState = string(state)
			break
		}
		_, combatEnded, state = g.GetCombatEngine().ProcessPlayerAttack(charB.ID, enemy.Entity.ID)
		if combatEnded {
			ended = true
			endState = string(state)
			break
		}
	}
	if !ended || endState != "victory" {
		t.Fatalf("expected victory end, ended=%v state=%q", ended, endState)
	}

	freshA, _ := facade.CharactersService().FindByID(charA.ID)
	freshB, _ := facade.CharactersService().FindByID(charB.ID)
	if freshA == nil || freshB == nil {
		t.Fatal("missing characters after victory")
	}
	if freshA.XP <= xpBeforeA {
		t.Fatalf("expected A XP to increase (before=%d after=%d)", xpBeforeA, freshA.XP)
	}
	if freshB.XP <= xpBeforeB {
		t.Fatalf("expected B XP to increase (before=%d after=%d)", xpBeforeB, freshB.XP)
	}
}
