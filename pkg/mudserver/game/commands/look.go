package commands

import (
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// Look executes the look command, allowing players to observe their surroundings or specific objects
func Look(room *rooms.Room, game def.GameCtrl, message *messages.Message) bool {
	parts := strings.Fields(message.Data)

	// Handle looking at the room (no specific target)
	if len(parts) == 1 {
		return lookAtRoom(room, game, message)
	}

	// Handle looking at a specific target
	target := strings.Join(parts[1:], " ")
	return lookAtTarget(room, game, message, target)
}

// lookAtRoom handles the case when a player looks at the room without specifying a target
func lookAtRoom(room *rooms.Room, game def.GameCtrl, message *messages.Message) bool {
	var sb strings.Builder

	if room.Detail != "" {
		sb.WriteString("You look around...\n")
		sb.WriteString(room.Detail)
	} else {
		sb.WriteString("You look around... nothing else to see here.")
	}

	// Show items in the room
	itemIDs := room.GetItemIDs()
	if len(itemIDs) > 0 {
		// Group items by template for stacking display
		itemCounts := make(map[string]int)
		itemNames := make(map[string]string)
		itemOrder := make([]string, 0)

		for _, itemID := range itemIDs {
			item, err := game.GetFacade().ItemsService().FindByID(itemID)
			if err != nil || item == nil {
				continue
			}

			// Use template ID for grouping if available, otherwise use item ID
			groupKey := item.TemplateID
			if groupKey == "" {
				groupKey = item.ID
			}

			if _, exists := itemCounts[groupKey]; !exists {
				itemOrder = append(itemOrder, groupKey)
				itemNames[groupKey] = item.Name
			}

			// For stackable items, add quantity; otherwise count instances
			if item.Stackable && item.Quantity > 0 {
				itemCounts[groupKey] += int(item.Quantity)
			} else {
				itemCounts[groupKey]++
			}
		}

		if len(itemOrder) > 0 {
			sb.WriteString("\n\nItems on the ground:")
			for _, key := range itemOrder {
				name := itemNames[key]
				count := itemCounts[key]
				sb.WriteString("\n - ")
				sb.WriteString(name)
				if count > 1 {
					sb.WriteString(" (x")
					sb.WriteString(itoa(count))
					sb.WriteString(")")
				}
			}
		}
	}

	game.SendMessage() <- message.Reply(sb.String())
	return true
}

// lookAtTarget handles the case when a player looks at a specific object or NPC
func lookAtTarget(room *rooms.Room, game def.GameCtrl, message *messages.Message, target string) bool {
	// Check for NPCs in the room
	npcManager := game.GetNPCInstanceManager()
	if npcManager != nil {
		npcInstance := npcManager.FindInstanceByNameInRoom(room.ID, target)
		if npcInstance != nil {
			result := lookAtNPC(npcInstance)
			game.SendMessage() <- message.Reply(result)
			return true
		}
	}

	// Check for items in the room
	item := findItemInRoom(room, game, target)
	if item != nil {
		result := lookAtItem(item)
		game.SendMessage() <- message.Reply(result)
		return true
	}

	// Check for items in player's inventory
	if message.Character != nil {
		invItem := message.Character.Inventory.FindItemByName(target)
		if invItem == nil {
			invItem = message.Character.Inventory.FindItemByTargetName(target)
		}
		if invItem != nil {
			result := lookAtItem(invItem)
			game.SendMessage() <- message.Reply(result)
			return true
		}
	}

	// TODO: Check for exit keywords (e.g., "north", "south")

	game.SendMessage() <- message.Reply("You don't see anything special about " + target + ".")
	return true
}

// findItemInRoom finds an item in the room by name
func findItemInRoom(room *rooms.Room, game def.GameCtrl, target string) *items.Item {
	itemIDs := room.GetItemIDs()
	targetLower := strings.ToLower(target)

	// First pass: exact match
	for _, itemID := range itemIDs {
		item, err := game.GetFacade().ItemsService().FindByID(itemID)
		if err != nil || item == nil {
			continue
		}
		if strings.ToLower(item.Name) == targetLower {
			return item
		}
		if strings.ToLower(item.GetTargetName()) == targetLower {
			return item
		}
	}

	// Second pass: prefix match
	for _, itemID := range itemIDs {
		item, err := game.GetFacade().ItemsService().FindByID(itemID)
		if err != nil || item == nil {
			continue
		}
		if strings.HasPrefix(strings.ToLower(item.Name), targetLower) {
			return item
		}
	}

	// Third pass: contains match
	for _, itemID := range itemIDs {
		item, err := game.GetFacade().ItemsService().FindByID(itemID)
		if err != nil || item == nil {
			continue
		}
		if strings.Contains(strings.ToLower(item.Name), targetLower) {
			return item
		}
	}

	return nil
}

// lookAtItem generates a description of an item
func lookAtItem(item *items.Item) string {
	var sb strings.Builder

	// Item name
	sb.WriteString("You look at ")
	sb.WriteString(item.Name)
	sb.WriteString(".\n")

	// Description
	if item.Description != "" {
		sb.WriteString(item.Description)
		sb.WriteString("\n")
	}

	// Item details
	sb.WriteString("\n")

	// Type and subtype
	if item.Type != "" {
		sb.WriteString("Type: ")
		sb.WriteString(string(item.Type))
		if item.SubType != "" {
			sb.WriteString(" (")
			sb.WriteString(string(item.SubType))
			sb.WriteString(")")
		}
		sb.WriteString("\n")
	}

	// Quality
	if item.Quality != "" {
		sb.WriteString("Quality: ")
		sb.WriteString(string(item.Quality))
		sb.WriteString("\n")
	}

	// Level
	if item.Level > 0 {
		sb.WriteString("Level: ")
		sb.WriteString(itoa(int(item.Level)))
		sb.WriteString("\n")
	}

	// Slot
	if item.Slot != "" && item.Slot != items.ItemSlotInventory {
		sb.WriteString("Slot: ")
		sb.WriteString(string(item.Slot))
		sb.WriteString("\n")
	}

	// Stack info
	if item.Stackable {
		sb.WriteString("Quantity: ")
		sb.WriteString(itoa(int(item.Quantity)))
		if item.MaxStack > 0 {
			sb.WriteString("/")
			sb.WriteString(itoa(int(item.MaxStack)))
		}
		sb.WriteString("\n")
	}

	// Attributes (stats)
	if len(item.Attributes) > 0 {
		sb.WriteString("\nAttributes:")
		for key, value := range item.Attributes {
			sb.WriteString("\n  ")
			sb.WriteString(key)
			sb.WriteString(": ")
			switch v := value.(type) {
			case float64:
				sb.WriteString(itoa(int(v)))
			case int:
				sb.WriteString(itoa(v))
			case string:
				sb.WriteString(v)
			default:
				sb.WriteString("?")
			}
		}
	}

	return sb.String()
}

// lookAtNPC generates a description of an NPC
func lookAtNPC(npc *npc.NPC) string {
	var sb strings.Builder

	// NPC name and basic info
	sb.WriteString("You look at ")
	sb.WriteString(npc.GetDisplayName())
	sb.WriteString(".\n")

	// Description
	if npc.Description != "" {
		sb.WriteString(npc.Description)
		sb.WriteString("\n")
	}

	// Race and class info
	if npc.Race.Name != "" || npc.Class.Name != "" {
		sb.WriteString("\n")
		if npc.Race.Name != "" {
			sb.WriteString("Race: ")
			sb.WriteString(npc.Race.Name)
		}
		if npc.Class.Name != "" {
			if npc.Race.Name != "" {
				sb.WriteString(" | ")
			}
			sb.WriteString("Class: ")
			sb.WriteString(npc.Class.Name)
		}
		sb.WriteString("\n")
	}

	// Level and health
	sb.WriteString("Level: ")
	sb.WriteString(itoa(int(npc.Level)))
	sb.WriteString(" | HP: ")
	sb.WriteString(itoa(int(npc.CurrentHitPoints)))
	sb.WriteString("/")
	sb.WriteString(itoa(int(npc.MaxHitPoints)))

	// Enemy indicator
	if npc.IsEnemy() {
		sb.WriteString("\n[HOSTILE]")
	}

	// Merchant indicator
	if npc.IsMerchant() {
		sb.WriteString("\n[MERCHANT]")
	}

	return sb.String()
}

// itoa converts int to string without importing strconv
func itoa(n int) string {
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
