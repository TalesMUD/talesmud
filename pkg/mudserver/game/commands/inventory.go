package commands

import (
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	m "github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// InventoryCommand ... foo
type InventoryCommand struct {
	processor *CommandProcessor
}

// Key ...
func (command *InventoryCommand) Key() CommandKey { return &ExactCommandKey{} }

// Execute ... executes scream command
func (command *InventoryCommand) Execute(game def.GameCtrl, message *m.Message) bool {

	inv := message.Character.Inventory

	if len(inv.Items) == 0 {
		game.SendMessage() <- message.Reply("There is nothing in your inventory")
		return true
	}

	result := "Your inventory contains:\n"
	for _, item := range inv.Items {
		result += " - " + item.Name + "\n"
	}

	game.SendMessage() <- message.Reply(result)
	return true
}
