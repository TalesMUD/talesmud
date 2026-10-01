package game

import (
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/items"
)

func testInstance(name string, typ items.ItemType, quality items.ItemQuality, tags []string, age time.Duration) *items.Item {
	return &items.Item{
		Entity:         &entities.Entity{ID: "inst-" + name},
		Name:           name,
		Type:           typ,
		SubType:        items.ItemSubType(tagsSub(tags)),
		Quality:        quality,
		Tags:           tags,
		TemplateID:     "TMPL-" + name,
		InstanceSuffix: "abcd1234",
		Created:        time.Now().Add(-age),
	}
}

func tagsSub(tags []string) string {
	for _, t := range tags {
		if t == "junk" || t == "trash" {
			return t
		}
	}
	return ""
}

func TestGroundDespawnTTLCategories(t *testing.T) {
	cases := []struct {
		name string
		item *items.Item
		want time.Duration
	}{
		{
			name: "junk rat tail",
			item: testInstance("Rat Tail", items.ItemTypeCollectible, items.ItemQualityNormal, []string{"junk", "trophy", "rat"}, 0),
			want: groundDespawnJunkTTL,
		},
		{
			name: "quest type instance",
			item: func() *items.Item {
				it := testInstance("Bone", items.ItemTypeQuest, items.ItemQualityNormal, []string{"quest"}, 0)
				it.SubType = "artifact_fragment"
				return it
			}(),
			want: groundDespawnImportantTTL,
		},
		{
			name: "rare weapon",
			item: testInstance("Rare Blade", items.ItemTypeWeapon, items.ItemQualityRare, nil, 0),
			want: groundDespawnImportantTTL,
		},
		{
			name: "magic armor",
			item: testInstance("Glint Mail", items.ItemTypeArmor, items.ItemQualityMagic, nil, 0),
			want: groundDespawnMagicTTL,
		},
		{
			name: "normal weapon",
			item: testInstance("Rusty Sword", items.ItemTypeWeapon, items.ItemQualityNormal, nil, 0),
			want: groundDespawnGearTTL,
		},
		{
			name: "crafting material",
			item: testInstance("Scrap Iron", items.ItemTypeCraftingMaterial, items.ItemQualityNormal, nil, 0),
			want: groundDespawnJunkTTL,
		},
		{
			name: "room blueprint never",
			item: &items.Item{
				Entity:       &entities.Entity{ID: "ITM0023"},
				Name:         "Rune-Carved Bone",
				Type:         items.ItemTypeQuest,
				CopyOnPickup: true,
				Created:      time.Now().Add(-2 * time.Hour),
			},
			want: 0,
		},
		{
			name: "template never",
			item: &items.Item{
				Entity:     &entities.Entity{ID: "ITM0018"},
				Name:       "Rat Tail",
				Type:       items.ItemTypeCollectible,
				IsTemplate: true,
				Tags:       []string{"junk"},
				Created:    time.Now().Add(-2 * time.Hour),
			},
			want: 0,
		},
		{
			name: "missing created never",
			item: &items.Item{
				Entity:         &entities.Entity{ID: "legacy"},
				Name:           "Old Drop",
				Type:           items.ItemTypeCollectible,
				Tags:           []string{"junk"},
				TemplateID:     "ITM0018",
				InstanceSuffix: "deadbeef",
			},
			want: 0,
		},
		{
			name: "default consumable",
			item: testInstance("Potion", items.ItemTypeConsumable, items.ItemQualityNormal, nil, 0),
			want: groundDespawnDefaultTTL,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := groundDespawnTTL(tc.item)
			if got != tc.want {
				t.Fatalf("TTL=%v want %v", got, tc.want)
			}
		})
	}
}

func TestGroundItemExpired(t *testing.T) {
	now := time.Now()
	fresh := testInstance("Rat Tail", items.ItemTypeCollectible, items.ItemQualityNormal, []string{"junk"}, time.Minute)
	if groundItemExpired(fresh, now) {
		t.Fatal("1-minute-old junk should still be present")
	}
	stale := testInstance("Rat Tail", items.ItemTypeCollectible, items.ItemQualityNormal, []string{"junk"}, 4*time.Minute)
	if !groundItemExpired(stale, now) {
		t.Fatal("4-minute-old junk should expire (TTL 3m)")
	}
	quest := testInstance("Medal", items.ItemTypeQuest, items.ItemQualityNormal, []string{"quest"}, 30*time.Minute)
	if groundItemExpired(quest, now) {
		t.Fatal("30-minute-old quest drop should still be present (TTL 60m)")
	}
	oldQuest := testInstance("Medal", items.ItemTypeQuest, items.ItemQualityNormal, []string{"quest"}, 61*time.Minute)
	if !groundItemExpired(oldQuest, now) {
		t.Fatal("61-minute-old quest drop should expire")
	}
}
