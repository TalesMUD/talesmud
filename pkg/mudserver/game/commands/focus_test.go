package commands_test

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func TestCombatFocusSelectAndRejectInvalid(t *testing.T) {
	g, facade := newSocialTestGame(t)
	user, char := storeSocialPlayer(t, facade, "focus-user", "focus-ref", "focus-char", "Wisp", "focus-room")
	exits := rooms.Exits{}
	chars := rooms.Characters{char.ID}
	room := &rooms.Room{Entity: &entities.Entity{ID: "focus-room"}, Name: "Focus Room", Exits: &exits, Characters: &chars}
	if _, err := facade.RoomsService().Import(room); err != nil {
		t.Fatal(err)
	}
	for _, spec := range []struct{ id, name string }{{"focus-rat", "Rat"}, {"focus-ogre", "Ogre"}} {
		enemy := &npc.NPC{Entity: &entities.Entity{ID: spec.id}, Name: spec.name, Level: 1,
			CurrentHitPoints: 80, MaxHitPoints: 80,
			EnemyTrait: &npc.EnemyTrait{AttackPower: 1, Difficulty: "trivial", CombatStyle: npc.CombatStyleSwarm}}
		enemy.CurrentRoomID = "focus-room"
		g.NPCManager.RegisterExistingNPC(enemy, "focus-room")
	}
	g.ConnectUserSession(user)
	g.SetUserSessionCharacter(user, char)
	(&commands.AttackCommand{}).Execute(g, &messages.Message{FromUser: user, Character: char, Data: "attack Rat"})
	var start *messages.CombatStartMessage
	for _, out := range drainSocialMessages(g.SendMessage()) {
		if msg, ok := out.(*messages.CombatStartMessage); ok {
			start = msg
		}
	}
	if start == nil || start.TargetID != "focus-rat" || len(start.Enemies) != 2 {
		t.Fatalf("combat start target/pack = %+v", start)
	}
	inst := g.GetCombatEngine().GetCombatInstance(char.ID)
	if inst == nil {
		t.Fatal("missing combat instance")
	}
	cmd := &commands.FocusCommand{}
	cmd.Execute(g, &messages.Message{FromUser: user, Character: char, Data: "focus focus-ogre"})
	if got := inst.GetPlayerByID(char.ID).AutoAttackTargetID; got != "focus-ogre" {
		t.Fatalf("focus = %s", got)
	}
	focusEvents := 0
	for _, out := range drainSocialMessages(g.SendMessage()) {
		if msg, ok := out.(*messages.CombatActionMessage); ok && msg.Action == "focus" && msg.TargetID == "focus-ogre" {
			focusEvents++
		}
	}
	if focusEvents != 1 {
		t.Fatalf("focus events = %d", focusEvents)
	}
	cmd.Execute(g, &messages.Message{FromUser: user, Character: char, Data: "focus focus-ogre"})
	if got := len(drainSocialMessages(g.SendMessage())); got != 0 {
		t.Fatalf("repeat focus emitted %d messages", got)
	}
	cmd.Execute(g, &messages.Message{FromUser: user, Character: char, Data: "focus missing"})
	if got := inst.GetPlayerByID(char.ID).AutoAttackTargetID; got != "focus-ogre" {
		t.Fatalf("invalid changed focus to %s", got)
	}
}
