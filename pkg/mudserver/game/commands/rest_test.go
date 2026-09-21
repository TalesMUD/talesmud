package commands_test

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func TestRestSetsFlagAndPushesUpdate(t *testing.T) {
	g, _ := newSocialTestGame(t)
	user, char := storeSocialPlayer(t, g.GetFacade(), "user-rest", "ref-rest", "char-rest", "Aryn", "room-1")
	char.CurrentHitPoints = 5
	char.MaxHitPoints = 20
	if err := g.GetFacade().CharactersService().Update(char.ID, char); err != nil {
		t.Fatalf("prep hp: %v", err)
	}
	g.ConnectUserSession(user)
	g.SetUserSessionCharacter(user, char)
	_ = drainSocialMessages(g.SendMessage())

	msg := &messages.Message{FromUser: user, Character: char, Data: "rest"}
	if !(&commands.RestCommand{}).Execute(g, msg) {
		t.Fatal("rest did not handle")
	}
	out := drainSocialMessages(g.SendMessage())
	var sawText, sawUpdate bool
	for _, raw := range out {
		switch v := raw.(type) {
		case messages.MessageResponse:
			if strings.Contains(v.Message, "begin to rest") {
				sawText = true
			}
		case *messages.CharacterUpdateMessage:
			if v.Resting {
				sawUpdate = true
			}
		}
	}
	if !sawText {
		t.Fatalf("expected rest confirmation, got %#v", out)
	}
	if !sawUpdate {
		t.Fatalf("expected characterUpdate resting=true, got %#v", out)
	}
	loaded, err := g.GetFacade().CharactersService().FindByID(char.ID)
	if err != nil || loaded == nil || loaded.Flags["resting"] != true {
		t.Fatalf("expected Flags.resting persisted, got %#v err=%v", loaded, err)
	}
}

func TestRestRefusesCombatAndSelf(t *testing.T) {
	g, _ := newSocialTestGame(t)
	user, char := storeSocialPlayer(t, g.GetFacade(), "user-rest", "ref-rest", "char-rest", "Aryn", "room-1")
	char.CurrentHitPoints = 20
	char.MaxHitPoints = 20
	g.ConnectUserSession(user)
	g.SetUserSessionCharacter(user, char)
	_ = drainSocialMessages(g.SendMessage())

	full := &messages.Message{FromUser: user, Character: char, Data: "rest"}
	(&commands.RestCommand{}).Execute(g, full)
	out := drainSocialMessages(g.SendMessage())
	if !strings.Contains(strings.Join(friendReplyTexts(out), "\n"), "full health") {
		t.Fatalf("expected full-health refusal, got %#v", friendReplyTexts(out))
	}

	char.CurrentHitPoints = 5
	char.InCombat = true
	combat := &messages.Message{FromUser: user, Character: char, Data: "rest"}
	(&commands.RestCommand{}).Execute(g, combat)
	out = drainSocialMessages(g.SendMessage())
	if !strings.Contains(strings.Join(friendReplyTexts(out), "\n"), "combat") {
		t.Fatalf("expected combat refusal, got %#v", friendReplyTexts(out))
	}
}
