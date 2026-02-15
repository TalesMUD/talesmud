package commands

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/mudserver/game/util"
)

// PickupCommand handles picking up items from the room
type PickupCommand struct {
}

// Key returns the command key matcher
func (command *PickupCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute handles the pickup/get/take command
func (command *PickupCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	if message.Character.CurrentRoomID == "" {
		game.SendMessage() <- message.Reply("You are not in a room.")
		return true
	}

	// Parse item name from command: "pickup sword" or "get sword" or "take sword"
	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		game.SendMessage() <- message.Reply("Pick up what? Usage: pickup <item>")
		return true
	}

	itemName := strings.Join(parts[1:], " ")

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

	// Find matching item in room
	itemIDs := room.GetItemIDs()
	if len(itemIDs) == 0 {
		game.SendMessage() <- message.Reply("There are no items here.")
		return true
	}

	// Search for item by name
	var foundItem *struct {
		id   string
		name string
	}
	itemNameLower := strings.ToLower(itemName)

	for _, itemID := range itemIDs {
		item, err := game.GetFacade().ItemsService().FindByID(itemID)
		if err != nil || item == nil {
			continue
		}

		// Check for exact match first
		if strings.ToLower(item.Name) == itemNameLower {
			foundItem = &struct {
				id   string
				name string
			}{id: item.ID, name: item.Name}
			break
		}

		// Check for target name match (name-suffix)
		if strings.ToLower(item.GetTargetName()) == itemNameLower {
			foundItem = &struct {
				id   string
				name string
			}{id: item.ID, name: item.Name}
			break
		}

		// Check for prefix match
		if foundItem == nil && strings.HasPrefix(strings.ToLower(item.Name), itemNameLower) {
			foundItem = &struct {
				id   string
				name string
			}{id: item.ID, name: item.Name}
		}

		// Check for contains match
		if foundItem == nil && strings.Contains(strings.ToLower(item.Name), itemNameLower) {
			foundItem = &struct {
				id   string
				name string
			}{id: item.ID, name: item.Name}
		}
	}

	if foundItem == nil {
		game.SendMessage() <- message.Reply("You don't see a '" + itemName + "' here.")
		return true
	}

	// Fetch the item
	item, err := game.GetFacade().ItemsService().FindByID(foundItem.id)
	if err != nil || item == nil {
		game.SendMessage() <- message.Reply("Error picking up item.")
		return true
	}

	// Check NoPickup flag
	if item.NoPickup {
		game.SendMessage() <- message.Reply("You can't pick up " + item.Name + ".")
		return true
	}

	// Handle CopyOnPickup items — create a personal copy instead of removing from room
	if item.CopyOnPickup {
		templateID := item.TemplateID
		if templateID == "" {
			templateID = item.ID // Item IS the template
		}

		// Check if player already collected this
		if message.Character.HasCollectedCopyItem(templateID) {
			game.SendMessage() <- message.Reply("You have already collected " + item.Name + ".")
			return true
		}

		// Check inventory capacity
		if message.Character.Inventory.IsFull() {
			game.SendMessage() <- message.Reply("Your inventory is full.")
			return true
		}

		// Create instance from template
		instance, err := game.GetFacade().ItemsService().CreateInstanceFromTemplate(templateID)
		if err != nil {
			log.WithError(err).Error("Error creating copy-on-pickup instance")
			game.SendMessage() <- message.Reply("Error picking up item.")
			return true
		}

		// Add instance to inventory
		err = message.Character.Inventory.AddItem(instance)
		if err != nil {
			game.SendMessage() <- message.Reply("Failed to pick up item: " + err.Error())
			return true
		}

		// Mark as collected (so it's hidden and can't be picked up again)
		message.Character.MarkCollectedCopyItem(templateID)

		// Persist character (inventory + flags updated; room is NOT modified)
		err = game.GetFacade().CharactersService().Update(message.Character.ID, message.Character)
		if err != nil {
			log.WithError(err).Error("Error updating character")
		}

		// Track quest progress (instance has TemplateID set, so quest tracker matches)
		NotifyQuestItemPickup(game, message.Character.ID, message.FromUser.ID, instance)

		// Send pickup message
		game.SendMessage() <- message.Reply("You pick up " + item.Name + ".")
		if inv := messages.NewInventoryUpdateMessage(message); inv != nil {
			game.SendMessage() <- inv
		}

		// Refresh room display so the collected item disappears for this player
		roomView := util.RoomWithCharacterReveals(room, message.Character)
		enterRoom := messages.NewEnterRoomMessage(roomView, message.FromUser, game, message.Character)
		enterRoom.AudienceID = message.FromUser.ID
		game.SendMessage() <- enterRoom

		return true
	}

	// Check inventory capacity
	if message.Character.Inventory.IsFull() {
		game.SendMessage() <- message.Reply("Your inventory is full.")
		return true
	}

	// Add item to character inventory
	err = message.Character.Inventory.AddItem(item)
	if err != nil {
		game.SendMessage() <- message.Reply("Failed to pick up item: " + err.Error())
		return true
	}

	// Remove item from room
	err = room.RemoveItem(item.ID)
	if err != nil {
		log.WithError(err).Error("Error removing item from room")
		// Rollback inventory change
		message.Character.Inventory.RemoveItem(item.ID)
		game.SendMessage() <- message.Reply("Error picking up item.")
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

	// Track quest progress for item pickup
	NotifyQuestItemPickup(game, message.Character.ID, message.FromUser.ID, item)

	// Send pickup message
	quantityStr := ""
	if item.Stackable && item.Quantity > 1 {
		quantityStr = " (x" + itoa(int(item.Quantity)) + ")"
	}
	game.SendMessage() <- message.Reply("You pick up " + item.Name + quantityStr + ".")
	if inv := messages.NewInventoryUpdateMessage(message); inv != nil {
		game.SendMessage() <- inv
	}

	return true
}
