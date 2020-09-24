package commands

import (
	"strings"

	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// ShrugCommand ... foo
type ShrugCommand struct {
}

// Key ...
func (command *ShrugCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute ... executes scream command
func (command *ShrugCommand) Execute(game def.GameCtrl, message *messages.Message) bool {

	parts := strings.Fields(message.Data)
	newMsg := strings.Join(parts[1:], " ")
	msg := ""

	if newMsg != "" {
		msg = "'" + newMsg + "'"
	}

	newMessage := "-- " + message.Character.Name + " shrugs " + msg + " ¯\\_(ツ)_/¯"
	out := messages.NewRoomBasedMessage("", newMessage)
	out.AudienceID = message.Character.CurrentRoomID
	game.SendMessage() <- out
	return true
}
