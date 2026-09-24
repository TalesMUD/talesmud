package commands_test

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func TestAttackWarnsOnceOnSkullEnemy(t *testing.T) {
	g, facade := newSocialTestGame(t)
	user, char := storeSocialPlayer(t, facade, "user-threat", "ref-threat", "char-threat", "Wisp", "room-threat")
	char.Level = 1
	if err := facade.CharactersService().Update(char.ID, char); err != nil {
		t.Fatal(err)
	}

	exits := rooms.Exits{}
	chars := rooms.Characters{char.ID}
	room := &rooms.Room{
		Entity:     &entities.Entity{ID: "room-threat"},
		Name:       "Threat Room",
		Exits:      &exits,
		Characters: &chars,
	}
	if _, err := facade.RoomsService().Import(room); err != nil {
		t.Fatal(err)
	}

	enemy := &npc.NPC{
		Entity:           &entities.Entity{ID: "skull-beast"},
		Name:             "Skull Beast",
		Level:            6,
		CurrentHitPoints: 40,
		MaxHitPoints:     40,
		EnemyTrait:       &npc.EnemyTrait{AttackPower: 8, Defense: 2, Difficulty: "boss"},
	}
	enemy.CurrentRoomID = "room-threat"
	g.NPCManager.RegisterExistingNPC(enemy, "room-threat")
	g.ConnectUserSession(user)
	g.SetUserSessionCharacter(user, char)

	cmd := &commands.AttackCommand{}
	if !cmd.Execute(g, &messages.Message{FromUser: user, Character: char, Data: "attack Skull Beast"}) {
		t.Fatal("first attack did not handle")
	}
	if g.GetCombatEngine().IsPlayerInCombat(char.ID) {
		t.Fatal("first attack on a skull enemy should not start combat")
	}
	warned := false
	for _, out := range drainSocialMessages(g.SendMessage()) {
		if rsp, ok := out.(messages.MessageResponse); ok && strings.Contains(rsp.Message, "Skull Beast is much stronger than you") {
			warned = true
		}
		if _, ok := out.(*messages.CombatStartMessage); ok {
			t.Fatal("warning should not open combat")
		}
	}
	if !warned {
		t.Fatal("expected much-stronger warning")
	}

	if !cmd.Execute(g, &messages.Message{FromUser: user, Character: char, Data: "attack Skull Beast"}) {
		t.Fatal("second attack did not handle")
	}
	if !g.GetCombatEngine().IsPlayerInCombat(char.ID) {
		t.Fatal("second attack should engage")
	}
	var sawThreat string
	for _, out := range drainSocialMessages(g.SendMessage()) {
		if start, ok := out.(*messages.CombatStartMessage); ok {
			if len(start.Enemies) == 0 {
				t.Fatal("combat start missing enemies")
			}
			sawThreat = start.Enemies[0].Threat
		}
	}
	if sawThreat != "skull" {
		t.Fatalf("combat payload threat = %q, want skull", sawThreat)
	}
}

func TestAttackBangSkipsThreatWarning(t *testing.T) {
	g, facade := newSocialTestGame(t)
	user, char := storeSocialPlayer(t, facade, "user-bang", "ref-bang", "char-bang", "Thorn", "room-bang")
	char.Level = 1
	if err := facade.CharactersService().Update(char.ID, char); err != nil {
		t.Fatal(err)
	}

	exits := rooms.Exits{}
	chars := rooms.Characters{char.ID}
	room := &rooms.Room{
		Entity:     &entities.Entity{ID: "room-bang"},
		Name:       "Bang Room",
		Exits:      &exits,
		Characters: &chars,
	}
	if _, err := facade.RoomsService().Import(room); err != nil {
		t.Fatal(err)
	}

	enemy := &npc.NPC{
		Entity:           &entities.Entity{ID: "orange-brute"},
		Name:             "Orange Brute",
		Level:            3, // +2 → orange
		CurrentHitPoints: 30,
		MaxHitPoints:     30,
		EnemyTrait:       &npc.EnemyTrait{AttackPower: 4, Defense: 1, Difficulty: "normal"},
	}
	enemy.CurrentRoomID = "room-bang"
	g.NPCManager.RegisterExistingNPC(enemy, "room-bang")
	g.ConnectUserSession(user)
	g.SetUserSessionCharacter(user, char)

	if !(&commands.AttackCommand{}).Execute(g, &messages.Message{FromUser: user, Character: char, Data: "attack! Orange Brute"}) {
		t.Fatal("attack! did not handle")
	}
	if !g.GetCombatEngine().IsPlayerInCombat(char.ID) {
		t.Fatal("attack! should engage an orange enemy immediately")
	}
	for _, out := range drainSocialMessages(g.SendMessage()) {
		if rsp, ok := out.(messages.MessageResponse); ok && strings.Contains(rsp.Message, "much stronger") {
			t.Fatal("attack! should not warn")
		}
	}
}

func TestAttackEvenLevelDoesNotWarn(t *testing.T) {
	g, facade := newSocialTestGame(t)
	user, char := storeSocialPlayer(t, facade, "user-even", "ref-even", "char-even", "Ember", "room-even")
	char.Level = 2
	if err := facade.CharactersService().Update(char.ID, char); err != nil {
		t.Fatal(err)
	}

	exits := rooms.Exits{}
	chars := rooms.Characters{char.ID}
	room := &rooms.Room{
		Entity:     &entities.Entity{ID: "room-even"},
		Name:       "Even Room",
		Exits:      &exits,
		Characters: &chars,
	}
	if _, err := facade.RoomsService().Import(room); err != nil {
		t.Fatal(err)
	}

	enemy := &npc.NPC{
		Entity:           &entities.Entity{ID: "yellow-wolf"},
		Name:             "Meadow Wolf",
		Level:            3, // +1 → yellow
		CurrentHitPoints: 20,
		MaxHitPoints:     20,
		EnemyTrait:       &npc.EnemyTrait{AttackPower: 3, Defense: 0, Difficulty: "normal"},
	}
	enemy.CurrentRoomID = "room-even"
	g.NPCManager.RegisterExistingNPC(enemy, "room-even")
	g.ConnectUserSession(user)
	g.SetUserSessionCharacter(user, char)

	if !(&commands.AttackCommand{}).Execute(g, &messages.Message{FromUser: user, Character: char, Data: "attack Meadow Wolf"}) {
		t.Fatal("attack did not handle")
	}
	if !g.GetCombatEngine().IsPlayerInCombat(char.ID) {
		t.Fatal("yellow enemy should not require a confirm")
	}
}
