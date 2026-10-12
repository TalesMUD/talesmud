package service

import (
	"math/rand"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/items"
)

func previewTable() *items.LootTable {
	return &items.LootTable{
		GoldMultiplier: 1,
		DropBonus:      0,
		Entries: []items.LootEntry{
			{ItemTemplateID: "LOW", DropChance: 0.5, MinQuantity: 1, MaxQuantity: 1, MinPlayerLevel: 5},
			{ItemTemplateID: "SURE", DropChance: 0, MinQuantity: 2, MaxQuantity: 4, Guaranteed: true},
			{ItemTemplateID: "BOSS", DropChance: 1, MinQuantity: 1, MaxQuantity: 3, BossOnly: true, Rarity: "unique"},
			{ItemTemplateID: "MAYBE", DropChance: 0.25, MinQuantity: 1, MaxQuantity: 1},
		},
	}
}

func TestPreviewLootIsSeededAndDoesNotNeedItems(t *testing.T) {
	table := previewTable()
	first := PreviewLoot(table, 40, 99, 1, false, nil)
	second := PreviewLoot(table, 40, 99, 1, false, nil)
	if len(first.Drops) != 4 || first.Drops[0].ItemTemplateID != "LOW" {
		t.Fatalf("drops = %+v", first.Drops)
	}
	if first.Drops[0].Hits != 0 || first.Drops[0].Rolls != 0 {
		t.Fatalf("level gate consumed a roll: %+v", first.Drops[0])
	}
	sure := first.Drops[1]
	if sure.Rolls != 40 || sure.Hits != 40 || sure.Frequency != 1 {
		t.Fatalf("guaranteed = %+v", sure)
	}
	if sure.Quantity < 80 || sure.Quantity > 160 {
		t.Fatalf("guaranteed quantity %d", sure.Quantity)
	}
	if first.Drops[2].Hits != 0 {
		t.Fatalf("boss-only dropped without boss: %+v", first.Drops[2])
	}
	if second.Drops[1].Quantity != sure.Quantity || second.Drops[3].Hits != first.Drops[3].Hits {
		t.Fatalf("seed diverged\n%+v\n%+v", first.Drops, second.Drops)
	}

	boss := PreviewLoot(table, 5, 99, 1, true, func(id string) (string, bool) {
		if id == "BOSS" {
			return "Crown", false
		}
		return id, false
	})
	crown := boss.Drops[2]
	if crown.Name != "Crown" || crown.Rolls != 5 || crown.Quantity != 5 {
		t.Fatalf("unique boss drop = %+v", crown)
	}

	// Chance is rolled before quantity. A guaranteed unique consumes neither a
	// failed-chance draw nor an Intn, so the next entry sees a fresh Float64.
	rng := rand.New(rand.NewSource(7))
	wantHits := 0
	for i := 0; i < 20; i++ {
		if rng.Float64() < 0.5 {
			wantHits++
		}
	}
	rolled := PreviewLoot(&items.LootTable{
		Entries: []items.LootEntry{
			{ItemTemplateID: "A", DropChance: 1, MinQuantity: 1, MaxQuantity: 1, Guaranteed: true, Rarity: "unique"},
			{ItemTemplateID: "B", DropChance: 0.5, MinQuantity: 1, MaxQuantity: 1},
		},
	}, 20, 7, 1, false, nil)
	if rolled.Drops[0].Quantity != 20 || rolled.Drops[1].Hits != wantHits {
		t.Fatalf("rng order = %+v want hits %d", rolled.Drops, wantHits)
	}
}
