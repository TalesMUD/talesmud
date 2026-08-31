package commands

import (
	"fmt"

	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// RepairCommand restores equipped armor durability at a merchant.
type RepairCommand struct{}

// Key ...
func (command *RepairCommand) Key() CommandKey { return &ExactCommandKey{} }

// Execute ...
func (command *RepairCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}
	merchant := findMerchantInRoom(game, message.Character.CurrentRoomID)
	if merchant == nil {
		game.SendMessage() <- message.Reply("No one here can repair armor. Find a merchant.")
		return true
	}
	n := message.Character.RepairEquippedArmor()
	if n == 0 {
		game.SendMessage() <- message.Reply("Your armor does not need repair.")
		return true
	}
	_ = game.GetFacade().CharactersService().Update(message.Character.ID, message.Character)
	game.SendMessage() <- message.Reply(fmt.Sprintf("%s repairs %d piece(s) of armor.", merchant.Name, n))
	return true
}
