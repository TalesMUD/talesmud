package commands

import (
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// ExamineCommand handles examining items in detail
type ExamineCommand struct {
}

// Key returns the command key matcher
func (command *ExamineCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute handles the examine/inspect command
func (command *ExamineCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	// Parse item name from command
	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		game.SendMessage() <- message.Reply("Examine what? Usage: examine <item>")
		return true
	}

	itemName := strings.Join(parts[1:], " ")

	// First, check inventory
	item := message.Character.Inventory.FindItemByName(itemName)
	if item == nil {
		item = message.Character.Inventory.FindItemByTargetName(itemName)
	}

	// If not in inventory, check the room
	if item == nil && message.Character.CurrentRoomID != "" {
		room, err := game.GetFacade().RoomsService().FindByID(message.Character.CurrentRoomID)
		if err == nil && room != nil {
			item = findItemInRoom(room, game, itemName)
		}
	}

	// Check equipped items
	if item == nil {
		for _, equipped := range message.Character.EquippedItems {
			if equipped == nil {
				continue
			}
			if strings.EqualFold(equipped.Name, itemName) ||
				strings.EqualFold(equipped.GetTargetName(), itemName) ||
				strings.HasPrefix(strings.ToLower(equipped.Name), strings.ToLower(itemName)) {
				item = equipped
				break
			}
		}
	}

	if item == nil {
		game.SendMessage() <- message.Reply("You don't see a '" + itemName + "' here.")
		return true
	}

	// Generate detailed item description
	result := examineItem(item, message.Character.EquippedItems)
	game.SendMessage() <- message.Reply(result)

	return true
}

// examineItem generates a detailed description of an item
func examineItem(item *items.Item, equippedItems map[items.ItemSlot]*items.Item) string {
	var sb strings.Builder

	// Header with quality color indicator
	sb.WriteString("=== ")
	sb.WriteString(item.Name)
	sb.WriteString(" ===\n")

	// Description
	if item.Description != "" {
		sb.WriteString("\n")
		sb.WriteString(item.Description)
		sb.WriteString("\n")
	}

	// Detail from LookAt trait
	if item.Detail != "" {
		sb.WriteString("\n")
		sb.WriteString(item.Detail)
		sb.WriteString("\n")
	}

	sb.WriteString("\n--- Item Details ---\n")

	// Type and subtype
	if item.Type != "" {
		sb.WriteString("Type: ")
		sb.WriteString(formatItemType(item.Type))
		if item.SubType != "" {
			sb.WriteString(" (")
			sb.WriteString(formatItemSubType(item.SubType))
			sb.WriteString(")")
		}
		sb.WriteString("\n")
	}

	// Quality
	if item.Quality != "" {
		sb.WriteString("Quality: ")
		sb.WriteString(formatQuality(item.Quality))
		sb.WriteString("\n")
	}

	// Level requirement
	if item.Level > 0 {
		sb.WriteString("Required Level: ")
		sb.WriteString(itoa(int(item.Level)))
		sb.WriteString("\n")
	}

	// Equipment slot
	if item.Slot != "" && item.Slot != items.ItemSlotInventory {
		sb.WriteString("Equip Slot: ")
		sb.WriteString(formatSlot(item.Slot))
		sb.WriteString("\n")
	}

	// Check if equipped
	isEquipped := false
	for _, equipped := range equippedItems {
		if equipped != nil && equipped.ID == item.ID {
			isEquipped = true
			break
		}
	}
	if isEquipped {
		sb.WriteString("Status: [EQUIPPED]\n")
	}

	// Stack info
	if item.Stackable {
		sb.WriteString("Stack: ")
		sb.WriteString(itoa(int(item.Quantity)))
		if item.MaxStack > 0 {
			sb.WriteString("/")
			sb.WriteString(itoa(int(item.MaxStack)))
		}
		sb.WriteString("\n")
	}

	// Base price
	if item.BasePrice > 0 {
		sb.WriteString("Value: ")
		sb.WriteString(itoa(int(item.BasePrice)))
		sb.WriteString(" gold\n")
	}

	// Attributes (stats)
	if len(item.Attributes) > 0 {
		sb.WriteString("\n--- Attributes ---\n")
		for key, value := range item.Attributes {
			sb.WriteString(formatAttributeName(key))
			sb.WriteString(": ")
			switch v := value.(type) {
			case float64:
				if v == float64(int(v)) {
					sb.WriteString(itoa(int(v)))
				} else {
					// Format float with one decimal
					sb.WriteString(formatFloat(v))
				}
			case int:
				sb.WriteString(itoa(v))
			case int32:
				sb.WriteString(itoa(int(v)))
			case int64:
				sb.WriteString(itoa64(v))
			case string:
				sb.WriteString(v)
			case bool:
				if v {
					sb.WriteString("Yes")
				} else {
					sb.WriteString("No")
				}
			default:
				sb.WriteString("?")
			}
			sb.WriteString("\n")
		}
	}

	// Properties
	if len(item.Properties) > 0 {
		sb.WriteString("\n--- Properties ---\n")
		for key, value := range item.Properties {
			sb.WriteString(formatAttributeName(key))
			sb.WriteString(": ")
			switch v := value.(type) {
			case float64:
				if v == float64(int(v)) {
					sb.WriteString(itoa(int(v)))
				} else {
					sb.WriteString(formatFloat(v))
				}
			case int:
				sb.WriteString(itoa(v))
			case string:
				sb.WriteString(v)
			case bool:
				if v {
					sb.WriteString("Yes")
				} else {
					sb.WriteString("No")
				}
			default:
				sb.WriteString("?")
			}
			sb.WriteString("\n")
		}
	}

	// Tags
	if len(item.Tags) > 0 {
		sb.WriteString("\nTags: ")
		sb.WriteString(strings.Join(item.Tags, ", "))
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatItemType formats item type for display
func formatItemType(t items.ItemType) string {
	switch t {
	case items.ItemTypeWeapon:
		return "Weapon"
	case items.ItemTypeArmor:
		return "Armor"
	case items.ItemTypeConsumable:
		return "Consumable"
	case items.ItemTypeQuest:
		return "Quest Item"
	case items.ItemTypeCurrency:
		return "Currency"
	case items.ItemTypeCollectible:
		return "Collectible"
	case items.ItemTypeCraftingMaterial:
		return "Crafting Material"
	default:
		return string(t)
	}
}

// formatItemSubType formats item subtype for display
func formatItemSubType(st items.ItemSubType) string {
	switch st {
	case items.ItemSubTypeSword:
		return "Sword"
	case items.ItemSubTypeTwoHandSword:
		return "Two-Handed Sword"
	case items.ItemSubTypeAxe:
		return "Axe"
	case items.ItemSubTypeSpear:
		return "Spear"
	case items.ItemSubTypeShield:
		return "Shield"
	default:
		return string(st)
	}
}

// formatQuality formats item quality for display
func formatQuality(q items.ItemQuality) string {
	switch q {
	case items.ItemQualityNormal:
		return "Normal"
	case items.ItemQualityMagic:
		return "Magic"
	case items.ItemQualityRare:
		return "Rare"
	case items.ItemQualityLegendary:
		return "Legendary"
	case items.ItemQualityMythic:
		return "Mythic"
	default:
		return string(q)
	}
}

// formatSlot formats equipment slot for display
func formatSlot(s items.ItemSlot) string {
	switch s {
	case items.ItemSlotHead:
		return "Head"
	case items.ItemSlotChest:
		return "Chest"
	case items.ItemSlotLegs:
		return "Legs"
	case items.ItemSlotBoots:
		return "Boots"
	case items.ItemSlotHands:
		return "Hands"
	case items.ItemSlotMainHand:
		return "Main Hand"
	case items.ItemSlotOffHand:
		return "Off Hand"
	case items.ItemSlotNeck:
		return "Neck"
	case items.ItemSlotRing1:
		return "Ring 1"
	case items.ItemSlotRing2:
		return "Ring 2"
	default:
		return string(s)
	}
}

// formatAttributeName formats attribute name for display (converts camelCase/snake_case to Title Case)
func formatAttributeName(name string) string {
	// Replace underscores with spaces
	name = strings.ReplaceAll(name, "_", " ")

	// Handle camelCase
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune(' ')
		}
		result.WriteRune(r)
	}

	// Title case
	return strings.Title(strings.ToLower(result.String()))
}

// formatFloat formats a float with one decimal place
func formatFloat(f float64) string {
	intPart := int(f)
	decPart := int((f - float64(intPart)) * 10)
	if decPart < 0 {
		decPart = -decPart
	}
	return itoa(intPart) + "." + itoa(decPart)
}
