package commands

import (
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// DropCommand handles dropping items to the room
type DropCommand struct {
}

// Key returns the command key matcher
func (command *DropCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute handles the drop command
func (command *DropCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	if message.Character.CurrentRoomID == "" {
		game.SendMessage() <- message.Reply("You are not in a room.")
		return true
	}

	// Parse command: "drop sword" or "drop sword 5" (for stackable items)
	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		game.SendMessage() <- message.Reply("Drop what? Usage: drop <item> [quantity]")
		return true
	}

	// Check if the last part is a quantity
	var quantity int32 = 0
	var itemName string

	lastPart := parts[len(parts)-1]
	if q, err := strconv.Atoi(lastPart); err == nil && len(parts) > 2 {
		quantity = int32(q)
		itemName = strings.Join(parts[1:len(parts)-1], " ")
	} else {
		itemName = strings.Join(parts[1:], " ")
	}

	// Find item in inventory
	item := message.Character.Inventory.FindItemByName(itemName)
	if item == nil {
		// Try by target name
		item = message.Character.Inventory.FindItemByTargetName(itemName)
	}

	if item == nil {
		game.SendMessage() <- message.Reply("You don't have a '" + itemName + "' in your inventory.")
		return true
	}

	// Get current room
	room, err := game.GetFacade().RoomsService().FindByID(message.Character.CurrentRoomID)
	if err != nil {
		log.WithError(err).Error("Error finding room")
		game.SendMessage() <- message.Reply("Error finding room.")
		return true
	}

	if room == nil {
		game.SendMessage() <- message.Reply("You are not in a valid room.")
		return true
	}

	// Handle quantity for stackable items
	if item.Stackable && quantity > 0 && quantity < item.Quantity {
		// Split the stack: create a new item with the dropped quantity
		droppedItem := *item // Copy the item
		droppedItem.Quantity = quantity
		item.Quantity -= quantity

		// Store the dropped item as a new entity
		storedItem, err := game.GetFacade().ItemsService().Store(&droppedItem)
		if err != nil {
			log.WithError(err).Error("Error storing dropped item")
			// Rollback quantity change
			item.Quantity += quantity
			game.SendMessage() <- message.Reply("Error dropping item.")
			return true
		}

		// Add dropped item to room
		err = room.AddItem(storedItem.ID)
		if err != nil {
			log.WithError(err).Error("Error adding item to room")
			game.SendMessage() <- message.Reply("Error dropping item.")
			return true
		}

		// Persist room update
		err = game.GetFacade().RoomsService().Update(room.ID, room)
		if err != nil {
			log.WithError(err).Error("Error updating room")
		}

		// Persist character update (original item stack was reduced)
		err = game.GetFacade().CharactersService().Update(message.Character.ID, message.Character)
		if err != nil {
			log.WithError(err).Error("Error updating character")
		}

		game.SendMessage() <- message.Reply("You drop " + item.Name + " (x" + itoa(int(quantity)) + ").")
		if inv := messages.NewInventoryUpdateMessage(message); inv != nil {
			game.SendMessage() <- inv
		}
		return true
	}

	// Remove item from inventory (full stack or non-stackable)
	removedItem, err := message.Character.Inventory.RemoveItem(item.ID)
	if err != nil {
		game.SendMessage() <- message.Reply("Error removing item from inventory.")
		return true
	}

	// Add item to room
	err = room.AddItem(removedItem.ID)
	if err != nil {
		log.WithError(err).Error("Error adding item to room")
		// Rollback inventory change
		message.Character.Inventory.AddItem(removedItem)
		game.SendMessage() <- message.Reply("Error dropping item.")
		return true
	}

	// Persist room update
	err = game.GetFacade().RoomsService().Update(room.ID, room)
	if err != nil {
		log.WithError(err).Error("Error updating room")
	}

	// Persist character update
	err = game.GetFacade().CharactersService().Update(message.Character.ID, message.Character)
	if err != nil {
		log.WithError(err).Error("Error updating character")
	}

	// Send drop message
	quantityStr := ""
	if removedItem.Stackable && removedItem.Quantity > 1 {
		quantityStr = " (x" + itoa(int(removedItem.Quantity)) + ")"
	}
	game.SendMessage() <- message.Reply("You drop " + removedItem.Name + quantityStr + ".")
	if inv := messages.NewInventoryUpdateMessage(message); inv != nil {
		game.SendMessage() <- inv
	}

	return true
}
