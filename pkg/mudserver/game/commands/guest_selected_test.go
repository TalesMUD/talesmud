package commands_test

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func TestSelectCharacterReportsGuest(t *testing.T) {
	for _, guest := range []bool{false, true} {
		g, _, user, fighter, _ := setupSwitchFight(t)
		user.IsGuest = guest
		(&commands.SelectCharacterCommand{}).Execute(g, &messages.Message{FromUser: user, Character: fighter, Data: "sc " + fighter.Name})
		var selected *messages.CharacterSelected
		for _, out := range drainSocialMessages(g.SendMessage()) {
			if msg, ok := out.(*messages.CharacterSelected); ok {
				selected = msg
			}
		}
		if selected == nil || selected.Guest != guest || selected.Message != "You are now playing as [WarriorKing]" {
			t.Fatalf("guest=%v selection=%+v", guest, selected)
		}
	}
}
