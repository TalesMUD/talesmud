package commands

import (
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

type RestCommand struct{}

// Key returns the command key matcher
func (command *RestCommand) Key() CommandKey { return &ExactCommandKey{} }

func (command *RestCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	// Validate character exists
	if message.Character == nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You must select a character first.")
		return true
	}

	char := message.Character

	// Check if in combat
	if char.InCombat {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You cannot rest while in combat!")
		return true
	}

	// Check if already at full HP
	if char.CurrentHitPoints >= char.MaxHitPoints {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You are already at full health.")
		return true
	}

	// Check if already resting
	if isAlreadyResting(char) {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You are already resting.")
		return true
	}

	// Set resting flag
	if char.Flags == nil {
		char.Flags = make(map[string]interface{})
	}
	char.Flags["resting"] = true

	// Save character
	if err := game.GetFacade().CharactersService().Update(char.ID, char); err != nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Failed to begin resting.")
		return true
	}

	// Send confirmation to player
	game.SendMessage() <- messages.Reply(message.FromUser.ID, "You sit down and begin to rest. You will recover health faster while resting.")

	// Notify room (excluding origin)
	roomMessage := char.Name + " sits down to rest."
	game.SendMessage() <- messages.MessageResponse{
		Audience:   messages.MessageAudienceRoomWithoutOrigin,
		AudienceID: char.CurrentRoomID,
		OriginID:   message.FromUser.ID,
		Message:    roomMessage,
	}

	return true
}

// isAlreadyResting checks if a character is currently resting.
func isAlreadyResting(char *characters.Character) bool {
	if char.Flags == nil {
		return false
	}

	resting, ok := char.Flags["resting"]
	if !ok {
		return false
	}

	restingBool, ok := resting.(bool)
	return ok && restingBool
}
