package commands

import (
	"strings"

	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// SayCommand allows players to speak to everyone in the current room
type SayCommand struct{}

// Key returns the command key matcher
func (command *SayCommand) Key() CommandKey { return &StartsWithCommandKey{} }

func (command *SayCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	// Validate character exists
	if message.Character == nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You must select a character first.")
		return true
	}

	// Parse message text
	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Say what? Usage: say <message>")
		return true
	}

	// Join all parts after "say" command
	text := strings.Join(parts[1:], " ")
	if text == "" {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Say what? Usage: say <message>")
		return true
	}

	// Format message
	roomMessage := message.Character.Name + " says: " + text

	// Send to room
	out := messages.NewRoomBasedMessage("", roomMessage)
	out.AudienceID = message.Character.CurrentRoomID
	game.SendMessage() <- out

	return true
}
