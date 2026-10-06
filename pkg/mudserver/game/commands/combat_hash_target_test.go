package commands_test

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/mudserver/game/util"
)

func TestInCombatAttackResolvesDisplayHashIndex(t *testing.T) {
	g, facade := newSocialTestGame(t)
	user, char := storeSocialPlayer(t, facade, "hash-user", "hash-ref", "hash-char", "Gimli", "hash-room")
	exits := rooms.Exits{}
	chars := rooms.Characters{char.ID}
	room := &rooms.Room{Entity: &entities.Entity{ID: "hash-room"}, Name: "Sewer Grate", Exits: &exits, Characters: &chars}
	if _, err := facade.RoomsService().Import(room); err != nil {
		t.Fatal(err)
	}
	// IDs sort as hash-rat-1 < hash-rat-2 < hash-rat-3 → UI #1/#2/#3
	for _, id := range []string{"hash-rat-1", "hash-rat-2", "hash-rat-3"} {
		enemy := &npc.NPC{Entity: &entities.Entity{ID: id}, Name: "Sewer Rat", Level: 1,
			CurrentHitPoints: 32, MaxHitPoints: 32,
			EnemyTrait: &npc.EnemyTrait{AttackPower: 1, Difficulty: "trivial", CombatStyle: npc.CombatStyleSwarm}}
		enemy.CurrentRoomID = "hash-room"
		g.NPCManager.RegisterExistingNPC(enemy, "hash-room")
	}

	roomNPCs := g.NPCManager.GetInstancesInRoom("hash-room")
	labels := util.BuildNPCDisplayNames(roomNPCs)
	if labels["hash-rat-1"] != "Sewer Rat#1" || labels["hash-rat-2"] != "Sewer Rat#2" || labels["hash-rat-3"] != "Sewer Rat#3" {
		t.Fatalf("unexpected room labels: %+v", labels)
	}

	g.ConnectUserSession(user)
	g.SetUserSessionCharacter(user, char)
	(&commands.AttackCommand{}).Execute(g, &messages.Message{FromUser: user, Character: char, Data: "attack Sewer Rat#2"})
	drainSocialMessages(g.SendMessage())
	inst := g.GetCombatEngine().GetCombatInstance(char.ID)
	if inst == nil {
		t.Fatal("missing combat instance after initiate with #2")
	}
	if got := inst.GetPlayerByID(char.ID).AutoAttackTargetID; got != "hash-rat-2" {
		t.Fatalf("initiate AutoAttackTargetID = %s want hash-rat-2", got)
	}

	// Switch focus via UI-style label during combat (the bricked path).
	(&commands.AttackCommand{}).Execute(g, &messages.Message{FromUser: user, Character: char, Data: "attack Sewer Rat#3"})
	outs := drainSocialMessages(g.SendMessage())
	for _, out := range outs {
		if rsp, ok := out.(messages.MessageResponse); ok && strings.Contains(rsp.Message, "Invalid target") {
			t.Fatalf("in-combat attack rejected #3: %s", rsp.Message)
		}
	}
	if got := inst.GetPlayerByID(char.ID).AutoAttackTargetID; got != "hash-rat-3" {
		t.Fatalf("in-combat AutoAttackTargetID = %s want hash-rat-3", got)
	}

	cmd := &commands.FocusCommand{}
	cmd.Execute(g, &messages.Message{FromUser: user, Character: char, Data: "focus Sewer Rat#1"})
	if got := inst.GetPlayerByID(char.ID).AutoAttackTargetID; got != "hash-rat-1" {
		t.Fatalf("focus #1 = %s want hash-rat-1", got)
	}
}
