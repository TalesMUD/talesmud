package game

import (
	"math/rand"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/service"
)

// LootDropResult represents the items and gold dropped from an NPC death
type LootDropResult struct {
	Items []*items.Item
	Gold  int64
	// RareUnique lists drops from a rarity:unique loot entry.
	// Template-unique trophies stay in Items and are not announced.
	RareUnique []*items.Item
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
	return DropLootFromNPCFor(facade, deadNPC, room, killerLevel, nil)
}

// DropLootFromNPCFor drops loot for one dead NPC.
// recipientIDs are the characters the drop is for. A unique entry is skipped
// only when every one of them already holds that template.
func DropLootFromNPCFor(facade service.Facade, deadNPC *npc.NPC, room *rooms.Room, killerLevel int32, recipientIDs []string) (*LootDropResult, error) {
	result := &LootDropResult{
		Items: make([]*items.Item, 0),
		Gold:  0,
	}

	// Check if NPC is an enemy with loot
	if !deadNPC.IsEnemy() || deadNPC.EnemyTrait == nil {
		return result, nil
	}

	enemy := deadNPC.EnemyTrait

	// Roll gold drop - use configured GoldDrop range or calculate from level/difficulty
	if enemy.GoldDrop.Max > 0 {
		goldMin := enemy.GoldDrop.Min
		goldMax := enemy.GoldDrop.Max
		if goldMax > goldMin {
			result.Gold = int64(goldMin) + int64(rand.Intn(int(goldMax-goldMin+1)))
		} else {
			result.Gold = int64(goldMin)
		}
	} else {
		result.Gold = leveling.RollEnemyGold(deadNPC.Level, enemy.Difficulty, rand.Intn)
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
		boss := isBossDifficulty(enemy.Difficulty)
		lootResult, err := facade.LootTablesService().RollLootInContext(enemy.LootTableID, killerLevel, 0, service.LootRollContext{
			Boss: boss,
			Owns: uniqueAlreadyHeld(facade, recipientIDs),
		})
		if err != nil {
			log.WithError(err).WithField("lootTableID", enemy.LootTableID).Warn("Failed to roll loot table")
		} else if lootResult != nil {
			kept := lootResult.Items
			// Apply max drops limit
			if enemy.MaxDrops > 0 && int32(len(kept)) > enemy.MaxDrops {
				// Randomly shuffle and take first MaxDrops items
				shuffleItems(kept)
				kept = kept[:enemy.MaxDrops]
			}

			result.Items = append(result.Items, kept...)
			result.Gold += lootResult.Gold
			result.RareUnique = append(result.RareUnique, rareStillDropped(lootResult.RareUnique, kept)...)
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

// lootReveals is the victory-panel list for items already dropped in the room.
// Unique is set only for items in announced, which are rarity:unique drops.
func lootReveals(dropped []*items.Item, announced []*items.Item) []messages.LootReveal {
	if len(dropped) == 0 {
		return nil
	}
	out := make([]messages.LootReveal, 0, len(dropped))
	for _, item := range dropped {
		if item == nil || strings.TrimSpace(item.Name) == "" {
			continue
		}
		qty := item.Quantity
		if qty < 1 {
			qty = 1
		}
		quality := string(item.Quality)
		if quality == "" {
			quality = string(items.ItemQualityNormal)
		}
		out = append(out, messages.LootReveal{
			Name:     item.Name,
			Quality:  quality,
			Quantity: qty,
			Unique:   announcedDrop(item, announced),
		})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func announcedDrop(item *items.Item, announced []*items.Item) bool {
	for _, other := range announced {
		if item != nil && item == other {
			return true
		}
	}
	return false
}

// rareStillDropped keeps rarity-unique items that survived the max-drops cut.
func rareStillDropped(rare, kept []*items.Item) []*items.Item {
	if len(rare) == 0 || len(kept) == 0 {
		return nil
	}
	out := make([]*items.Item, 0, len(rare))
	for _, item := range rare {
		if announcedDrop(item, kept) {
			out = append(out, item)
		}
	}
	return out
}

// uniqueDropMessage is the room-wide gold chip for a rarity:unique drop.
func uniqueDropMessage(roomID, npcName, itemName string) messages.MessageResponse {
	msg := messages.NewRoomBasedMessage("SYSTEM", "UNIQUE: "+itemName)
	msg.Audience = messages.MessageAudienceRoom
	msg.AudienceID = roomID
	msg.Style = "combatEvent"
	msg.Hook = "unique"
	msg.Source = npcName
	return msg
}

// uniqueAlreadyHeld is true only when every recipient already holds the template.
// An empty list, or a character that cannot be loaded, does not block the drop.
func uniqueAlreadyHeld(facade service.Facade, recipientIDs []string) func(string) bool {
	if facade == nil || len(recipientIDs) == 0 {
		return nil
	}
	return func(templateID string) bool {
		saw := false
		for _, id := range recipientIDs {
			if id == "" {
				continue
			}
			ch, err := facade.CharactersService().FindByID(id)
			if err != nil || ch == nil {
				return false
			}
			saw = true
			if ch.CountOfTemplate(templateID) < 1 {
				return false
			}
		}
		return saw
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
