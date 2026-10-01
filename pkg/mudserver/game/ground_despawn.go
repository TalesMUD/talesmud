package game

import (
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

// Ground-drop TTL defaults (tunable constants).
// Only applies to loot/player instances on the ground — never room blueprints.
const (
	// Junk / trash / common crafting & trophy leftovers from combat.
	groundDespawnJunkTTL = 3 * time.Minute
	// Default for ordinary non-junk instances (consumables, currency piles, etc.).
	groundDespawnDefaultTTL = 5 * time.Minute
	// Normal-quality weapons and armor left on the ground.
	groundDespawnGearTTL = 15 * time.Minute
	// Magic-quality drops.
	groundDespawnMagicTTL = 30 * time.Minute
	// Quest items, rare+, and tagged-important drops — long enough to come back.
	groundDespawnImportantTTL = 60 * time.Minute
)

// groundDespawnTTL returns how long a ground item may linger before auto-despawn.
// Zero means "never auto-despawn" (fixtures / missing timestamps).
func groundDespawnTTL(item *items.Item) time.Duration {
	if item == nil {
		return 0
	}
	// Permanent room fixtures and catalog blueprints stay forever.
	if item.IsTemplate || item.CopyOnPickup || item.NoPickup || item.IsRoomBlueprint() {
		return 0
	}
	// Only despawn true loot/player instances (template + suffix).
	if !item.IsInstance() {
		return 0
	}
	if item.Created.IsZero() {
		return 0
	}

	if isGroundImportant(item) {
		return groundDespawnImportantTTL
	}
	if item.Quality == items.ItemQualityMagic {
		return groundDespawnMagicTTL
	}
	if isGroundGear(item) {
		return groundDespawnGearTTL
	}
	if isGroundJunk(item) {
		return groundDespawnJunkTTL
	}
	return groundDespawnDefaultTTL
}

func isGroundImportant(item *items.Item) bool {
	if item.Type == items.ItemTypeQuest {
		return true
	}
	switch item.Quality {
	case items.ItemQualityRare, items.ItemQualityLegendary, items.ItemQualityMythic:
		return true
	}
	if hasItemTag(item, "quest") || hasItemTag(item, "important") || hasItemTag(item, "unique") {
		return true
	}
	sub := strings.ToLower(string(item.SubType))
	return sub == "quest" || sub == "quest_item" || sub == "artifact_fragment"
}

func isGroundGear(item *items.Item) bool {
	return item.Type == items.ItemTypeWeapon || item.Type == items.ItemTypeArmor
}

func isGroundJunk(item *items.Item) bool {
	typeName := strings.ToLower(string(item.Type))
	sub := strings.ToLower(string(item.SubType))
	if typeName == "junk" || sub == "junk" || sub == "trash" {
		return true
	}
	if hasItemTag(item, "junk") || hasItemTag(item, "trash") || hasItemTag(item, "trophy") {
		return true
	}
	// Common crafting / collectible fodder at normal quality.
	if item.Quality == "" || item.Quality == items.ItemQualityNormal {
		if item.Type == items.ItemTypeCollectible || item.Type == items.ItemTypeCraftingMaterial {
			return true
		}
	}
	return false
}

func hasItemTag(item *items.Item, want string) bool {
	want = strings.ToLower(want)
	for _, t := range item.Tags {
		if strings.ToLower(t) == want {
			return true
		}
	}
	return false
}

func groundItemExpired(item *items.Item, now time.Time) bool {
	ttl := groundDespawnTTL(item)
	if ttl <= 0 {
		return false
	}
	return !item.Created.After(now.Add(-ttl))
}

// despawnExpiredGroundItems removes timed-out loot instances from rooms,
// deletes their item rows, and refreshes occupant clients when anything changed.
func (g *Game) despawnExpiredGroundItems(now time.Time, allRooms []*rooms.Room) {
	for _, room := range allRooms {
		if room == nil || room.Items == nil || len(*room.Items) == 0 {
			continue
		}
		removed := g.despawnExpiredInRoom(room, now)
		if removed == 0 {
			continue
		}
		if err := g.Facade.RoomsService().Update(room.ID, room); err != nil {
			log.WithError(err).WithField("roomID", room.ID).Warn("ground despawn: failed to persist room")
			continue
		}
		log.WithFields(log.Fields{
			"roomID":  room.ID,
			"removed": removed,
		}).Info("ground despawn: cleared expired drops")
		g.sendRoomUpdateToOccupants(room.ID)
	}
}

func (g *Game) despawnExpiredInRoom(room *rooms.Room, now time.Time) int {
	ids := room.GetItemIDs()
	removed := 0
	for _, itemID := range ids {
		item, err := g.Facade.ItemsService().FindByID(itemID)
		if err != nil || item == nil {
			continue
		}
		if !groundItemExpired(item, now) {
			continue
		}
		if err := room.RemoveItem(itemID); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"roomID": room.ID,
				"itemID": itemID,
			}).Warn("ground despawn: remove from room failed")
			continue
		}
		if err := g.Facade.ItemsService().Delete(itemID); err != nil {
			log.WithError(err).WithField("itemID", itemID).Warn("ground despawn: delete item failed")
		}
		removed++
		log.WithFields(log.Fields{
			"roomID":   room.ID,
			"itemID":   itemID,
			"name":     item.Name,
			"template": item.TemplateID,
			"quality":  item.Quality,
			"type":     item.Type,
			"age":      now.Sub(item.Created).Round(time.Second).String(),
		}).Debug("ground despawn: expired drop removed")
	}
	return removed
}
