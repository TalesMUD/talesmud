package commands

import (
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// HelpCommand ... foo
type HelpCommand struct {
	processor *CommandProcessor
}

// Execute ... executes scream command
func (command *HelpCommand) Execute(game def.GameCtrl, message *messages.Message) bool {

	result := "List of all global commands:\n"

	for key, element := range command.processor.Help {
		result += key + " " + element + "\n"
	}

	game.SendMessage(messages.Reply(message.FromUser.ID.Hex(), result))
	return true
}
