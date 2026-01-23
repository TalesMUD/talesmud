package items

import (
	"github.com/talesmud/talesmud/pkg/entities"
)

// LootTable represents a collection of potential item drops
type LootTable struct {
	*entities.Entity `bson:",inline"`

	Name        string `bson:"name" json:"name"`
	Description string `bson:"description,omitempty" json:"description,omitempty"`

	// Entries defines the potential drops
	Entries []LootEntry `bson:"entries" json:"entries"`

	// GoldMultiplier multiplies the base gold drop (1.0 = normal)
	GoldMultiplier float64 `bson:"goldMultiplier" json:"goldMultiplier"`

	// DropBonus adds a flat bonus to all drop chances (0.0 = normal)
	DropBonus float64 `bson:"dropBonus" json:"dropBonus"`
}

// LootEntry represents a single potential drop in a loot table
type LootEntry struct {
	// ItemTemplateID is the ID of the item template to drop
	ItemTemplateID string `bson:"itemTemplateId" json:"itemTemplateId"`

	// DropChance is the probability of dropping (0.0 to 1.0, where 1.0 = 100%)
	DropChance float64 `bson:"dropChance" json:"dropChance"`

	// MinQuantity is the minimum number of items to drop (for stackable items)
	MinQuantity int32 `bson:"minQuantity" json:"minQuantity"`

	// MaxQuantity is the maximum number of items to drop (for stackable items)
	MaxQuantity int32 `bson:"maxQuantity" json:"maxQuantity"`

	// Guaranteed means this item always drops regardless of DropChance
	Guaranteed bool `bson:"guaranteed" json:"guaranteed"`

	// MinPlayerLevel is the minimum player level required for this drop
	MinPlayerLevel int32 `bson:"minPlayerLevel,omitempty" json:"minPlayerLevel,omitempty"`

	// RequiredTags are tags the player must have for this drop (e.g., quest progress)
	RequiredTags []string `bson:"requiredTags,omitempty" json:"requiredTags,omitempty"`
}

// LootTables type alias for slice of LootTable pointers
type LootTables []*LootTable

// NewLootTable creates a new loot table with default values
func NewLootTable() *LootTable {
	return &LootTable{
		Entity:         entities.NewEntity(),
		Entries:        make([]LootEntry, 0),
		GoldMultiplier: 1.0,
		DropBonus:      0.0,
	}
}

// AddEntry adds a new loot entry to the table
func (lt *LootTable) AddEntry(entry LootEntry) {
	if entry.MinQuantity == 0 {
		entry.MinQuantity = 1
	}
	if entry.MaxQuantity == 0 {
		entry.MaxQuantity = 1
	}
	lt.Entries = append(lt.Entries, entry)
}

// GetGuaranteedEntries returns all entries marked as guaranteed
func (lt *LootTable) GetGuaranteedEntries() []LootEntry {
	var result []LootEntry
	for _, entry := range lt.Entries {
		if entry.Guaranteed {
			result = append(result, entry)
		}
	}
	return result
}

// GetRandomEntries returns entries that are not guaranteed (require rolling)
func (lt *LootTable) GetRandomEntries() []LootEntry {
	var result []LootEntry
	for _, entry := range lt.Entries {
		if !entry.Guaranteed {
			result = append(result, entry)
		}
	}
	return result
}
