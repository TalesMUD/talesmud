package service

import (
	"math/rand"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/items"
)

// LootPreview is a seeded frequency count for one loot table.
// It is a dry run: nothing is created or stored.
type LootPreview struct {
	N           int               `json:"n"`
	Seed        int64             `json:"seed"`
	PlayerLevel int32             `json:"playerLevel"`
	Boss        bool              `json:"boss"`
	Drops       []LootPreviewDrop `json:"drops"`
}

// LootPreviewDrop is how often one item template appeared.
// Hits counts successful entries and can exceed N when several entries share an item.
// Rolls counts the rolls that dropped it at least once. Frequency is Rolls/N.
type LootPreviewDrop struct {
	ItemTemplateID string  `json:"itemTemplateId"`
	Name           string  `json:"name,omitempty"`
	Hits           int     `json:"hits"`
	Rolls          int     `json:"rolls"`
	Frequency      float64 `json:"frequency"`
	Quantity       int64   `json:"quantity"`
}

// TemplateInfo names an item template and says whether the template itself is unique.
// A nil lookup leaves the name empty and treats only a rarity of "unique" as unique.
type TemplateInfo func(templateID string) (name string, unique bool)

// PreviewLoot rolls table n times with math/rand seeded by seed.
// The chance and quantity order matches RollLootFromTableInContext.
// Boss-only entries are skipped unless boss is set. Ownership is not checked.
// The function does not create items.
func PreviewLoot(table *items.LootTable, n int, seed int64, playerLevel int32, boss bool, lookup TemplateInfo) LootPreview {
	out := LootPreview{
		N:           n,
		Seed:        seed,
		PlayerLevel: playerLevel,
		Boss:        boss,
		Drops:       []LootPreviewDrop{},
	}
	if table == nil || n < 1 {
		return out
	}

	type info struct {
		name   string
		unique bool
		index  int
	}
	known := map[string]*info{}
	order := make([]string, 0)
	for _, entry := range table.Entries {
		id := entry.ItemTemplateID
		if _, ok := known[id]; ok {
			continue
		}
		row := &info{index: len(order)}
		if lookup != nil && id != "" {
			row.name, row.unique = lookup(id)
		}
		known[id] = row
		order = append(order, id)
	}

	hits := map[string]int{}
	rolls := map[string]int{}
	qty := map[string]int64{}
	rng := rand.New(rand.NewSource(seed))

	for i := 0; i < n; i++ {
		seen := map[string]bool{}
		for _, entry := range table.Entries {
			if entry.MinPlayerLevel > 0 && playerLevel < entry.MinPlayerLevel {
				continue
			}
			if entry.BossOnly && !boss {
				continue
			}
			meta := known[entry.ItemTemplateID]
			shouldDrop := entry.Guaranteed
			if !shouldDrop {
				chance := entry.DropChance + table.DropBonus
				if chance > 1 {
					chance = 1
				}
				shouldDrop = rng.Float64() < chance
			}
			if !shouldDrop {
				continue
			}
			unique := strings.EqualFold(strings.TrimSpace(entry.Rarity), "unique") || (meta != nil && meta.unique)
			quantity := entry.MinQuantity
			if !unique && entry.MaxQuantity > entry.MinQuantity {
				quantity = entry.MinQuantity + int32(rng.Intn(int(entry.MaxQuantity-entry.MinQuantity+1)))
			}
			if quantity < 1 || unique {
				quantity = 1
			}
			id := entry.ItemTemplateID
			hits[id]++
			qty[id] += int64(quantity)
			if !seen[id] {
				seen[id] = true
				rolls[id]++
			}
		}
	}

	out.Drops = make([]LootPreviewDrop, 0, len(order))
	for _, id := range order {
		meta := known[id]
		drop := LootPreviewDrop{
			ItemTemplateID: id,
			Name:           meta.name,
			Hits:           hits[id],
			Rolls:          rolls[id],
			Quantity:       qty[id],
		}
		if n > 0 {
			drop.Frequency = float64(drop.Rolls) / float64(n)
		}
		out.Drops = append(out.Drops, drop)
	}
	return out
}
