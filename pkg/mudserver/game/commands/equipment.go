package commands

import (
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// EquipmentCommand displays equipped items
type EquipmentCommand struct {
}

// Key returns the command key matcher
func (command *EquipmentCommand) Key() CommandKey { return &ExactCommandKey{} }

// Execute handles the equipment/eq/gear command
func (command *EquipmentCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	var sb strings.Builder
	sb.WriteString("=== Equipment ===\n")

	// Define slot order for display
	slots := []struct {
		slot items.ItemSlot
		name string
	}{
		{items.ItemSlotHead, "Head"},
		{items.ItemSlotNeck, "Neck"},
		{items.ItemSlotChest, "Chest"},
		{items.ItemSlotHands, "Hands"},
		{items.ItemSlotLegs, "Legs"},
		{items.ItemSlotBoots, "Boots"},
		{items.ItemSlotMainHand, "Main Hand"},
		{items.ItemSlotOffHand, "Off Hand"},
		{items.ItemSlotRing1, "Ring 1"},
		{items.ItemSlotRing2, "Ring 2"},
	}

	// Track items we've already shown (for two-handed weapons)
	shownItems := make(map[string]bool)

	for _, slotInfo := range slots {
		sb.WriteString(slotInfo.name)
		sb.WriteString(": ")

		item := message.Character.EquippedItems[slotInfo.slot]
		if item == nil {
			sb.WriteString("[Empty]")
		} else if shownItems[item.ID] {
			// Skip if we've shown this item (two-handed weapon in both hands)
			sb.WriteString("[see " + formatSlotName(items.ItemSlotMainHand) + "]")
		} else {
			sb.WriteString(item.Name)
			shownItems[item.ID] = true

			// Show quality if not normal
			if item.Quality != "" && item.Quality != items.ItemQualityNormal {
				sb.WriteString(" [")
				sb.WriteString(strings.ToUpper(string(item.Quality)))
				sb.WriteString("]")
			}

			// Show key stats
			stats := getItemStats(item)
			if stats != "" {
				sb.WriteString(" (")
				sb.WriteString(stats)
				sb.WriteString(")")
			}
		}
		sb.WriteString("\n")
	}

	game.SendMessage() <- message.Reply(sb.String())
	return true
}

// formatSlotName returns a display-friendly slot name
func formatSlotName(slot items.ItemSlot) string {
	switch slot {
	case items.ItemSlotHead:
		return "Head"
	case items.ItemSlotNeck:
		return "Neck"
	case items.ItemSlotChest:
		return "Chest"
	case items.ItemSlotHands:
		return "Hands"
	case items.ItemSlotLegs:
		return "Legs"
	case items.ItemSlotBoots:
		return "Boots"
	case items.ItemSlotMainHand:
		return "Main Hand"
	case items.ItemSlotOffHand:
		return "Off Hand"
	case items.ItemSlotRing1:
		return "Ring 1"
	case items.ItemSlotRing2:
		return "Ring 2"
	default:
		return string(slot)
	}
}

// getItemStats returns a brief stats summary for an item
func getItemStats(item *items.Item) string {
	if item.Attributes == nil {
		return ""
	}

	var parts []string

	// Common stats to show
	if dmg, ok := item.Attributes["damage"]; ok {
		parts = append(parts, formatStat("Dmg", dmg))
	}
	if def, ok := item.Attributes["defense"]; ok {
		parts = append(parts, formatStat("Def", def))
	}
	if armor, ok := item.Attributes["armor"]; ok {
		parts = append(parts, formatStat("Armor", armor))
	}
	if str, ok := item.Attributes["strength"]; ok {
		parts = append(parts, formatStat("Str", str))
	}
	if agi, ok := item.Attributes["agility"]; ok {
		parts = append(parts, formatStat("Agi", agi))
	}
	if intel, ok := item.Attributes["intelligence"]; ok {
		parts = append(parts, formatStat("Int", intel))
	}

	return strings.Join(parts, ", ")
}

// formatStat formats a stat value for display
func formatStat(name string, value interface{}) string {
	switch v := value.(type) {
	case float64:
		if v >= 0 {
			return name + " +" + itoa(int(v))
		}
		return name + " " + itoa(int(v))
	case int:
		if v >= 0 {
			return name + " +" + itoa(v)
		}
		return name + " " + itoa(v)
	case int32:
		if v >= 0 {
			return name + " +" + itoa(int(v))
		}
		return name + " " + itoa(int(v))
	default:
		return ""
	}
}
