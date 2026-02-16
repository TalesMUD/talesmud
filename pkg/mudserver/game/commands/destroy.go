package commands

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/mudserver/game/util"
)

// DestroyCommand handles destroying items from inventory
type DestroyCommand struct {
}

// Key returns the command key matcher
func (command *DestroyCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute handles the destroy/discard command
func (command *DestroyCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	// Parse command: "destroy sword"
	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		game.SendMessage() <- message.Reply("Destroy what? Usage: destroy <item>")
		return true
	}

	itemName := strings.Join(parts[1:], " ")

	// Find item in inventory
	item := message.Character.Inventory.FindItemByName(itemName)
	if item == nil {
		item = message.Character.Inventory.FindItemByTargetName(itemName)
	}

	if item == nil {
		game.SendMessage() <- message.Reply("You don't have a '" + itemName + "' in your inventory.")
		return true
	}

	// Remove from inventory
	_, err := message.Character.Inventory.RemoveItem(item.ID)
	if err != nil {
		log.WithError(err).Error("Error removing item from inventory")
		game.SendMessage() <- message.Reply("You can't seem to destroy " + item.Name + ".")
		return true
	}

	// For bound items (CopyOnPickup instances), clear the collected flag
	// so the player can pick up the template again
	if item.IsBound() && item.TemplateID != "" {
		if message.Character.Flags != nil {
			delete(message.Character.Flags, "collected_item:"+item.TemplateID)
		}
	}

	// Persist character
	err = game.GetFacade().CharactersService().Update(message.Character.ID, message.Character)
	if err != nil {
		log.WithError(err).Error("Error updating character")
	}

	// Delete the item from the database
	game.GetFacade().ItemsService().Delete(item.ID)

	// Choose message based on binding
	if item.IsBound() {
		game.SendMessage() <- message.Reply("You destroy " + item.Name + ". It fades away.")
	} else {
		game.SendMessage() <- message.Reply("You destroy " + item.Name + ". It crumbles to dust.")
	}

	if inv := messages.NewInventoryUpdateMessage(message); inv != nil {
		game.SendMessage() <- inv
	}

	// For bound items, refresh the room display so the CopyOnPickup template reappears
	if item.IsBound() && message.Character.CurrentRoomID != "" {
		room, rerr := game.GetFacade().RoomsService().FindByID(message.Character.CurrentRoomID)
		if rerr == nil && room != nil {
			roomView := util.RoomWithCharacterReveals(room, message.Character)
			enterRoom := messages.NewEnterRoomMessage(roomView, message.FromUser, game, message.Character)
			enterRoom.AudienceID = message.FromUser.ID
			game.SendMessage() <- enterRoom
		}
	}

	return true
}
