package game

import (
	"math/rand"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/service"
)

// LootDropResult represents the items and gold dropped from an NPC death
type LootDropResult struct {
	Items []*items.Item
	Gold  int64
}

// DropLootFromNPC handles loot drops when an NPC dies
// It processes guaranteed loot, loot tables, and gold drops
// Parameters:
// - facade: the service facade for accessing services
// - npc: the NPC that was killed
// - room: the room where loot should be placed
// - killerLevel: the level of the player who killed the NPC (for level-restricted drops)
// Returns: the loot result and any error
func DropLootFromNPC(facade service.Facade, deadNPC *npc.NPC, room *rooms.Room, killerLevel int32) (*LootDropResult, error) {
	result := &LootDropResult{
		Items: make([]*items.Item, 0),
		Gold:  0,
	}

	// Check if NPC is an enemy with loot
	if !deadNPC.IsEnemy() || deadNPC.EnemyTrait == nil {
		return result, nil
	}

	enemy := deadNPC.EnemyTrait

	// Roll gold drop
	if enemy.GoldDrop.Max > 0 {
		goldMin := enemy.GoldDrop.Min
		goldMax := enemy.GoldDrop.Max
		if goldMax > goldMin {
			result.Gold = int64(goldMin) + int64(rand.Intn(int(goldMax-goldMin+1)))
		} else {
			result.Gold = int64(goldMin)
		}
	}

	// Process guaranteed loot
	for _, templateID := range enemy.GuaranteedLoot {
		item, err := facade.ItemsService().CreateInstanceFromTemplate(templateID)
		if err != nil {
			log.WithError(err).WithField("templateID", templateID).Warn("Failed to create guaranteed loot item")
			continue
		}
		result.Items = append(result.Items, item)
	}

	// Process loot table
	if enemy.LootTableID != "" {
		lootResult, err := facade.LootTablesService().RollLoot(enemy.LootTableID, killerLevel, 0)
		if err != nil {
			log.WithError(err).WithField("lootTableID", enemy.LootTableID).Warn("Failed to roll loot table")
		} else if lootResult != nil {
			// Apply max drops limit
			if enemy.MaxDrops > 0 && int32(len(lootResult.Items)) > enemy.MaxDrops {
				// Randomly shuffle and take first MaxDrops items
				shuffleItems(lootResult.Items)
				lootResult.Items = lootResult.Items[:enemy.MaxDrops]
			}

			result.Items = append(result.Items, lootResult.Items...)
			result.Gold += lootResult.Gold
		}
	}

	// Place items in room
	for _, item := range result.Items {
		// Store the item
		storedItem, err := facade.ItemsService().Store(item)
		if err != nil {
			log.WithError(err).WithField("item", item.Name).Warn("Failed to store dropped item")
			continue
		}

		// Add to room
		if err := room.AddItem(storedItem.ID); err != nil {
			log.WithError(err).WithField("item", item.Name).Warn("Failed to add item to room")
		}
	}

	// Update room if items were placed
	if len(result.Items) > 0 {
		if err := facade.RoomsService().Update(room.ID, room); err != nil {
			log.WithError(err).WithField("roomID", room.ID).Warn("Failed to update room with loot")
		}
	}

	log.WithFields(log.Fields{
		"npc":        deadNPC.GetDisplayName(),
		"items":      len(result.Items),
		"gold":       result.Gold,
		"room":       room.ID,
		"lootTable":  enemy.LootTableID,
		"guaranteed": len(enemy.GuaranteedLoot),
	}).Debug("Loot dropped from NPC")

	return result, nil
}

// shuffleItems randomly shuffles a slice of items in place
func shuffleItems(items []*items.Item) {
	for i := len(items) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		items[i], items[j] = items[j], items[i]
	}
}

// FormatLootMessage creates a player-facing message about loot drops
func FormatLootMessage(result *LootDropResult, npcName string) string {
	if len(result.Items) == 0 && result.Gold == 0 {
		return ""
	}

	var msg string
	if len(result.Items) > 0 || result.Gold > 0 {
		msg = npcName + " drops:"
	}

	for _, item := range result.Items {
		msg += "\n - " + item.Name
		if item.Stackable && item.Quantity > 1 {
			msg += " (x" + itoa(int(item.Quantity)) + ")"
		}
	}

	if result.Gold > 0 {
		msg += "\n - " + itoa64(result.Gold) + " gold"
	}

	return msg
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
