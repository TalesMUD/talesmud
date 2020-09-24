package commands

import (
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	m "github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// HelpCommand ... foo
type HelpCommand struct {
	processor *CommandProcessor
}

// Key ...
func (command *HelpCommand) Key() CommandKey { return &ExactCommandKey{} }

// Execute ... executes scream command
func (command *HelpCommand) Execute(game def.GameCtrl, message *m.Message) bool {

	result := "List of all global commands:\n"

	for key, element := range command.processor.Help {
		result += key + " " + element + "\n"
	}

	game.SendMessage() <- m.Reply(message.FromUser.ID, result)
	return true
}
