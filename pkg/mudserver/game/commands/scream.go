package commands

import (
	"strings"

	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// ScreamCommand ... foo
type ScreamCommand struct {
}

// Execute ... executes scream command
func (screamCommand *ScreamCommand) Execute(game def.GameCtrl, message *messages.Message) bool {

	parts := strings.Fields(message.Data)
	newMsg := strings.Join(parts[1:], " ")

	newMessage := "-- " + message.Character.Name + " screams " + strings.ToUpper(newMsg) + "!!!!!"
	out := messages.NewRoomBasedMessage("", newMessage)
	out.AudienceID = message.Character.CurrentRoomID

	game.SendMessage(out)

	return true
}
