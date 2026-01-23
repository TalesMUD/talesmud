package commands

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// UnequipCommand handles unequipping items
type UnequipCommand struct {
}

// Key returns the command key matcher
func (command *UnequipCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute handles the unequip/remove command
func (command *UnequipCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	// Parse: "unequip <slot|item>"
	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		game.SendMessage() <- message.Reply("Unequip what? Usage: unequip <slot or item name>")
		return true
	}

	target := strings.Join(parts[1:], " ")
	targetLower := strings.ToLower(target)

	if message.Character.EquippedItems == nil || len(message.Character.EquippedItems) == 0 {
		game.SendMessage() <- message.Reply("You don't have anything equipped.")
		return true
	}

	// Try to find by slot name first
	slot := parseSlotName(targetLower)
	var item *items.Item
	var foundSlot items.ItemSlot

	if slot != "" {
		// Found a slot name
		item = message.Character.EquippedItems[slot]
		foundSlot = slot
	} else {
		// Try to find by item name
		for s, equipped := range message.Character.EquippedItems {
			if equipped == nil {
				continue
			}
			if strings.EqualFold(equipped.Name, target) ||
				strings.HasPrefix(strings.ToLower(equipped.Name), targetLower) ||
				strings.EqualFold(equipped.GetTargetName(), target) {
				item = equipped
				foundSlot = s
				break
			}
		}
	}

	if item == nil {
		game.SendMessage() <- message.Reply("Nothing equipped there or no item named '" + target + "'.")
		return true
	}

	// Check inventory space
	if message.Character.Inventory.IsFull() {
		game.SendMessage() <- message.Reply("Your inventory is full.")
		return true
	}

	// Handle two-handed weapons (remove from both slots)
	slotsToRemove := []items.ItemSlot{foundSlot}
	if item.SubType == items.ItemSubTypeTwoHandSword {
		slotsToRemove = []items.ItemSlot{items.ItemSlotMainHand, items.ItemSlotOffHand}
	}

	// Remove from all relevant slots
	for _, s := range slotsToRemove {
		if message.Character.EquippedItems[s] != nil && message.Character.EquippedItems[s].ID == item.ID {
			delete(message.Character.EquippedItems, s)
		}
	}

	// Add to inventory
	err := message.Character.Inventory.AddItem(item)
	if err != nil {
		// Rollback
		message.Character.EquippedItems[foundSlot] = item
		game.SendMessage() <- message.Reply("Error unequipping item.")
		return true
	}

	// Persist character
	err = game.GetFacade().CharactersService().Update(message.Character.ID, message.Character)
	if err != nil {
		log.WithError(err).Error("Failed to update character")
	}

	game.SendMessage() <- message.Reply("You unequip " + item.Name + ".")
	return true
}

// parseSlotName converts slot names/aliases to ItemSlot
func parseSlotName(name string) items.ItemSlot {
	switch name {
	case "head", "helm", "helmet":
		return items.ItemSlotHead
	case "chest", "body", "torso", "armor":
		return items.ItemSlotChest
	case "legs", "pants", "leggings":
		return items.ItemSlotLegs
	case "boots", "feet", "shoes":
		return items.ItemSlotBoots
	case "hands", "gloves", "gauntlets":
		return items.ItemSlotHands
	case "main_hand", "mainhand", "main", "weapon", "right":
		return items.ItemSlotMainHand
	case "off_hand", "offhand", "off", "shield", "left":
		return items.ItemSlotOffHand
	case "neck", "necklace", "amulet":
		return items.ItemSlotNeck
	case "ring1", "ring_1", "ring 1":
		return items.ItemSlotRing1
	case "ring2", "ring_2", "ring 2":
		return items.ItemSlotRing2
	default:
		return ""
	}
}
