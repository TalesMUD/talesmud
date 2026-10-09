package service

import (
	"math/rand"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/items"
	r "github.com/talesmud/talesmud/pkg/repository"
)

// LootRollContext gates boss-only and unique entries.
// An empty context matches RollLoot: nothing is boss-only blocked, and ownership is not checked.
type LootRollContext struct {
	// Boss is true when the dead NPC's difficulty is boss.
	Boss bool
	// Owns reports that every recipient already holds the template.
	// Nil does not block the drop.
	Owns func(templateID string) bool
	// Float64 and Intn override math/rand. Nil uses the global source.
	Float64 func() float64
	Intn    func(n int) int
}

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

	// RollLootInContext is RollLoot with boss and ownership gates.
	RollLootInContext(tableID string, playerLevel int32, baseGold int64, ctx LootRollContext) (*LootDropResult, error)

	// RollLootFromTable rolls against a provided loot table (for testing or custom drops)
	RollLootFromTable(table *items.LootTable, playerLevel int32, baseGold int64) (*LootDropResult, error)

	// RollLootFromTableInContext is RollLootFromTable with boss and ownership gates.
	RollLootFromTableInContext(table *items.LootTable, playerLevel int32, baseGold int64, ctx LootRollContext) (*LootDropResult, error)
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
	return srv.RollLootInContext(tableID, playerLevel, baseGold, LootRollContext{})
}

// RollLootInContext implements LootTablesService.RollLootInContext
func (srv *lootTablesService) RollLootInContext(tableID string, playerLevel int32, baseGold int64, ctx LootRollContext) (*LootDropResult, error) {
	table, err := srv.FindByID(tableID)
	if err != nil {
		return nil, err
	}
	if table == nil {
		return &LootDropResult{Items: []*items.Item{}, Gold: 0}, nil
	}

	return srv.RollLootFromTableInContext(table, playerLevel, baseGold, ctx)
}

// RollLootFromTable implements LootTablesService.RollLootFromTable
func (srv *lootTablesService) RollLootFromTable(table *items.LootTable, playerLevel int32, baseGold int64) (*LootDropResult, error) {
	return srv.RollLootFromTableInContext(table, playerLevel, baseGold, LootRollContext{})
}

// RollLootFromTableInContext implements LootTablesService.RollLootFromTableInContext
func (srv *lootTablesService) RollLootFromTableInContext(table *items.LootTable, playerLevel int32, baseGold int64, ctx LootRollContext) (*LootDropResult, error) {
	result := &LootDropResult{
		Items: make([]*items.Item, 0),
		Gold:  int64(float64(baseGold) * table.GoldMultiplier),
	}

	for _, entry := range table.Entries {
		// Check player level requirement
		if entry.MinPlayerLevel > 0 && playerLevel < entry.MinPlayerLevel {
			continue
		}
		if srv.entryBlocked(entry, ctx) {
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
			shouldDrop = ctx.float64() < effectiveChance
		}

		if !shouldDrop {
			continue
		}

		unique := srv.entryIsUnique(entry)
		// Determine quantity. A unique drops a single copy.
		quantity := entry.MinQuantity
		if !unique && entry.MaxQuantity > entry.MinQuantity {
			quantity = entry.MinQuantity + int32(ctx.intn(int(entry.MaxQuantity-entry.MinQuantity+1)))
		}
		if quantity < 1 || unique {
			quantity = 1
		}

		// Create item instance from template
		item, err := srv.itemsService.CreateInstanceFromTemplate(entry.ItemTemplateID)
		if err != nil {
			// Log error but continue with other drops
			continue
		}
		srv.stampUnique(item, entry)

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
				srv.stampUnique(itemInstance, entry)
				result.Items = append(result.Items, itemInstance)
			}
			continue
		}

		result.Items = append(result.Items, item)
	}

	return result, nil
}

func (ctx LootRollContext) float64() float64 {
	if ctx.Float64 != nil {
		return ctx.Float64()
	}
	return rand.Float64()
}

func (ctx LootRollContext) intn(n int) int {
	if n <= 0 {
		return 0
	}
	if ctx.Intn != nil {
		return ctx.Intn(n)
	}
	return rand.Intn(n)
}

func (srv *lootTablesService) entryBlocked(entry items.LootEntry, ctx LootRollContext) bool {
	if entry.BossOnly && !ctx.Boss {
		return true
	}
	if !srv.entryIsUnique(entry) || ctx.Owns == nil {
		return false
	}
	return ctx.Owns(entry.ItemTemplateID)
}

func (srv *lootTablesService) entryIsUnique(entry items.LootEntry) bool {
	if strings.EqualFold(strings.TrimSpace(entry.Rarity), "unique") {
		return true
	}
	if srv == nil || srv.itemsService == nil || entry.ItemTemplateID == "" {
		return false
	}
	tpl, err := srv.itemsService.FindByID(entry.ItemTemplateID)
	if err != nil || tpl == nil {
		return false
	}
	return tpl.Unique
}

func (srv *lootTablesService) stampUnique(item *items.Item, entry items.LootEntry) {
	if item == nil || item.Unique || !strings.EqualFold(strings.TrimSpace(entry.Rarity), "unique") {
		return
	}
	item.Unique = true
	if srv != nil && srv.itemsService != nil && item.ID != "" {
		_ = srv.itemsService.Update(item.ID, item)
	}
}
