package service

import (
	"math/rand"

	"github.com/talesmud/talesmud/pkg/entities/items"
	r "github.com/talesmud/talesmud/pkg/repository"
)

// LootDropResult represents the result of rolling a loot table
type LootDropResult struct {
	Items []*items.Item
	Gold  int64
}

// LootTablesService delivers logical functions on top of the loot tables repository
type LootTablesService interface {
	r.LootTablesRepository

	// RollLoot rolls against a loot table and returns dropped items and gold
	// Parameters:
	// - tableID: the ID of the loot table to roll against
	// - playerLevel: the player's level (for level-restricted drops)
	// - baseGold: the base gold amount (will be multiplied by table's GoldMultiplier)
	// Returns: the loot result containing items and gold
	RollLoot(tableID string, playerLevel int32, baseGold int64) (*LootDropResult, error)

	// RollLootFromTable rolls against a provided loot table (for testing or custom drops)
	RollLootFromTable(table *items.LootTable, playerLevel int32, baseGold int64) (*LootDropResult, error)
}

type lootTablesService struct {
	r.LootTablesRepository
	itemsService ItemsService
}

// NewLootTablesService creates a new loot tables service
func NewLootTablesService(lootTablesRepo r.LootTablesRepository, itemsService ItemsService) LootTablesService {
	return &lootTablesService{
		LootTablesRepository: lootTablesRepo,
		itemsService:         itemsService,
	}
}

// RollLoot implements LootTablesService.RollLoot
func (srv *lootTablesService) RollLoot(tableID string, playerLevel int32, baseGold int64) (*LootDropResult, error) {
	table, err := srv.FindByID(tableID)
	if err != nil {
		return nil, err
	}
	if table == nil {
		return &LootDropResult{Items: []*items.Item{}, Gold: 0}, nil
	}

	return srv.RollLootFromTable(table, playerLevel, baseGold)
}

// RollLootFromTable implements LootTablesService.RollLootFromTable
func (srv *lootTablesService) RollLootFromTable(table *items.LootTable, playerLevel int32, baseGold int64) (*LootDropResult, error) {
	result := &LootDropResult{
		Items: make([]*items.Item, 0),
		Gold:  int64(float64(baseGold) * table.GoldMultiplier),
	}

	for _, entry := range table.Entries {
		// Check player level requirement
		if entry.MinPlayerLevel > 0 && playerLevel < entry.MinPlayerLevel {
			continue
		}

		// Check if item should drop
		shouldDrop := entry.Guaranteed
		if !shouldDrop {
			// Apply drop bonus from table
			effectiveChance := entry.DropChance + table.DropBonus
			if effectiveChance > 1.0 {
				effectiveChance = 1.0
			}
			shouldDrop = rand.Float64() < effectiveChance
		}

		if !shouldDrop {
			continue
		}

		// Determine quantity
		quantity := entry.MinQuantity
		if entry.MaxQuantity > entry.MinQuantity {
			quantity = entry.MinQuantity + int32(rand.Intn(int(entry.MaxQuantity-entry.MinQuantity+1)))
		}
		if quantity < 1 {
			quantity = 1
		}

		// Create item instance from template
		item, err := srv.itemsService.CreateInstanceFromTemplate(entry.ItemTemplateID)
		if err != nil {
			// Log error but continue with other drops
			continue
		}

		// Set quantity for stackable items
		if item.Stackable {
			item.Quantity = quantity
		} else {
			// For non-stackable items, create multiple instances
			for i := int32(0); i < quantity; i++ {
				itemInstance, err := srv.itemsService.CreateInstanceFromTemplate(entry.ItemTemplateID)
				if err != nil {
					continue
				}
				result.Items = append(result.Items, itemInstance)
			}
			continue
		}

		result.Items = append(result.Items, item)
	}

	return result, nil
}
