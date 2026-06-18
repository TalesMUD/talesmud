package commands

import (
	"strings"

	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// TellCommand sends a private message to another player
type TellCommand struct{}

// Key returns the command key matcher
func (command *TellCommand) Key() CommandKey { return &StartsWithCommandKey{} }

func (command *TellCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	// Validate character exists
	if message.Character == nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You must select a character first.")
		return true
	}

	// Parse command: "tell PlayerName message text here"
	parts := strings.Fields(message.Data)
	if len(parts) < 3 {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Usage: tell <player> <message>")
		return true
	}

	targetName := parts[1]
	text := strings.Join(parts[2:], " ")

	if text == "" {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Tell them what? Usage: tell <player> <message>")
		return true
	}

	// Find online player
	targetPlayer, found := game.FindOnlinePlayerByName(targetName)
	if !found {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Player '"+targetName+"' is not online.")
		return true
	}

	// Check not sending to self
	if targetPlayer.UserID == message.FromUser.ID {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You can't send a tell to yourself.")
		return true
	}

	// Send message to target
	targetMessage := message.Character.Name + " tells you: " + text
	game.SendMessage() <- messages.Reply(targetPlayer.UserID, targetMessage)

	// Send confirmation to sender
	confirmMessage := "You tell " + targetPlayer.CharacterName + ": " + text
	game.SendMessage() <- messages.Reply(message.FromUser.ID, confirmMessage)

	return true
}
