package commands

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// EquipCommand handles equipping items from inventory
type EquipCommand struct {
}

// Key returns the command key matcher
func (command *EquipCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute handles the equip/wear command
func (command *EquipCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	// Parse: "equip <item>"
	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		game.SendMessage() <- message.Reply("Equip what? Usage: equip <item>")
		return true
	}

	itemName := strings.Join(parts[1:], " ")

	// Find item in inventory
	item := message.Character.Inventory.FindItemByName(itemName)
	if item == nil {
		item = message.Character.Inventory.FindItemByTargetName(itemName)
	}
	if item == nil {
		game.SendMessage() <- message.Reply("You don't have '" + itemName + "' in your inventory.")
		return true
	}

	// Check if item can be equipped
	if item.Slot == "" || item.Slot == items.ItemSlotInventory || item.Slot == items.ItemSlotContainer || item.Slot == items.ItemSlotPurse {
		game.SendMessage() <- message.Reply(item.Name + " cannot be equipped.")
		return true
	}

	// Ensure EquippedItems map exists
	if message.Character.EquippedItems == nil {
		message.Character.EquippedItems = make(map[items.ItemSlot]*items.Item)
	}

	// Handle two-handed weapons
	targetSlots := []items.ItemSlot{item.Slot}
	if item.SubType == items.ItemSubTypeTwoHandSword {
		targetSlots = []items.ItemSlot{items.ItemSlotMainHand, items.ItemSlotOffHand}
	}

	// Unequip existing items in target slots
	var unequippedItems []*items.Item
	for _, slot := range targetSlots {
		if existing := message.Character.EquippedItems[slot]; existing != nil {
			// Add existing item back to inventory
			err := message.Character.Inventory.AddItem(existing)
			if err != nil {
				game.SendMessage() <- message.Reply("Your inventory is full. Cannot swap equipment.")
				return true
			}
			unequippedItems = append(unequippedItems, existing)
			delete(message.Character.EquippedItems, slot)
		}
	}

	// Remove item from inventory
	_, err := message.Character.Inventory.RemoveItem(item.ID)
	if err != nil {
		// Rollback unequipped items
		for _, unequipped := range unequippedItems {
			message.Character.Inventory.RemoveItem(unequipped.ID)
			message.Character.EquippedItems[unequipped.Slot] = unequipped
		}
		game.SendMessage() <- message.Reply("Error equipping item.")
		return true
	}

	// Equip item
	for _, slot := range targetSlots {
		message.Character.EquippedItems[slot] = item
	}

	// Persist character
	err = game.GetFacade().CharactersService().Update(message.Character.ID, message.Character)
	if err != nil {
		log.WithError(err).Error("Failed to update character")
	}

	// Build response message
	var msg strings.Builder
	msg.WriteString("You equip ")
	msg.WriteString(item.Name)
	msg.WriteString(".")

	if len(unequippedItems) > 0 {
		msg.WriteString(" (Unequipped: ")
		for i, unequipped := range unequippedItems {
			if i > 0 {
				msg.WriteString(", ")
			}
			msg.WriteString(unequipped.Name)
		}
		msg.WriteString(")")
	}

	game.SendMessage() <- message.Reply(msg.String())
	if inv := messages.NewInventoryUpdateMessage(message); inv != nil {
		game.SendMessage() <- inv
	}
	return true
}
