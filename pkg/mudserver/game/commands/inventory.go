package commands

import (
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	m "github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// InventoryCommand ... foo
type InventoryCommand struct {
	processor *CommandProcessor
}

// Key ...
func (command *InventoryCommand) Key() CommandKey { return &ExactCommandKey{} }

// Execute ... executes inventory command
func (command *InventoryCommand) Execute(game def.GameCtrl, message *m.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	inv := message.Character.Inventory
	var sb strings.Builder

	// Header with slot count
	slotCount := inv.Count()
	maxSlots := inv.Size
	if maxSlots == 0 {
		maxSlots = 20 // Default
	}
	sb.WriteString("=== Inventory (")
	sb.WriteString(itoa(slotCount))
	sb.WriteString("/")
	sb.WriteString(itoa(int(maxSlots)))
	sb.WriteString(" slots) ===\n")

	// Show gold
	sb.WriteString("Gold: ")
	sb.WriteString(itoa64(message.Character.Gold))
	sb.WriteString("\n")

	if len(inv.Items) == 0 {
		sb.WriteString("\nYour inventory is empty.")
		game.SendMessage() <- message.Reply(sb.String())
		return true
	}

	// Group items by type
	weapons := make([]*items.Item, 0)
	armor := make([]*items.Item, 0)
	consumables := make([]*items.Item, 0)
	quest := make([]*items.Item, 0)
	other := make([]*items.Item, 0)

	for _, item := range inv.Items {
		switch item.Type {
		case items.ItemTypeWeapon:
			weapons = append(weapons, item)
		case items.ItemTypeArmor:
			armor = append(armor, item)
		case items.ItemTypeConsumable:
			consumables = append(consumables, item)
		case items.ItemTypeQuest:
			quest = append(quest, item)
		default:
			other = append(other, item)
		}
	}

	// Check equipped items
	equipped := make(map[string]bool)
	for _, item := range message.Character.EquippedItems {
		if item != nil {
			equipped[item.ID] = true
		}
	}

	// Format each category
	if len(weapons) > 0 {
		sb.WriteString("\n[Weapons]\n")
		formatItemList(&sb, weapons, equipped)
	}

	if len(armor) > 0 {
		sb.WriteString("\n[Armor]\n")
		formatItemList(&sb, armor, equipped)
	}

	if len(consumables) > 0 {
		sb.WriteString("\n[Consumables]\n")
		formatItemList(&sb, consumables, equipped)
	}

	if len(quest) > 0 {
		sb.WriteString("\n[Quest Items]\n")
		formatItemList(&sb, quest, equipped)
	}

	if len(other) > 0 {
		sb.WriteString("\n[Other]\n")
		formatItemList(&sb, other, equipped)
	}

	game.SendMessage() <- message.Reply(sb.String())
	if inv := m.NewInventoryUpdateMessage(message); inv != nil {
		game.SendMessage() <- inv
	}
	return true
}

// formatItemList formats a list of items for display
func formatItemList(sb *strings.Builder, itemList []*items.Item, equipped map[string]bool) {
	for _, item := range itemList {
		sb.WriteString(" - ")
		sb.WriteString(item.Name)

		// Show quantity for stackable items
		if item.Stackable && item.Quantity > 1 {
			sb.WriteString(" (x")
			sb.WriteString(itoa(int(item.Quantity)))
			sb.WriteString(")")
		}

		// Show equipped indicator
		if equipped[item.ID] {
			sb.WriteString(" [EQUIPPED]")
		}

		// Show quality indicator for non-normal quality
		if item.Quality != "" && item.Quality != items.ItemQualityNormal {
			sb.WriteString(" [")
			sb.WriteString(strings.ToUpper(string(item.Quality)))
			sb.WriteString("]")
		}

		sb.WriteString("\n")
	}
}

// itoa64 converts int64 to string
func itoa64(n int64) string {
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if negative {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
