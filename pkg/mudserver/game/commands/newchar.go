package commands

import (
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// NewCharacterCommand ...
type NewCharacterCommand struct {
}

// Execute ... executes the command
func (command *NewCharacterCommand) Execute(game def.GameCtrl, message *messages.Message) bool {

	// just send a message to the client to start character creation
	game.SendMessage() <- messages.NewCreateCharacterMessage(message.FromUser.ID)
	return true
}
