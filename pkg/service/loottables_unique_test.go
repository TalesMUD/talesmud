package service

import (
	"math/rand"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/items"
	r "github.com/talesmud/talesmud/pkg/repository"
)

type uniqueItemStub struct {
	templates map[string]*items.Item
}

func (s *uniqueItemStub) CreateInstanceFromTemplate(templateID string) (*items.Item, error) {
	tpl := s.templates[templateID]
	if tpl == nil {
		return nil, errMissingTemplate(templateID)
	}
	cp := *tpl
	cp.Entity = &entities.Entity{ID: templateID + "-inst"}
	cp.IsTemplate = false
	cp.TemplateID = templateID
	cp.Unique = tpl.Unique
	return &cp, nil
}

func (s *uniqueItemStub) FindByID(id string) (*items.Item, error) {
	tpl := s.templates[id]
	if tpl == nil {
		return nil, errMissingTemplate(id)
	}
	return tpl, nil
}

func (s *uniqueItemStub) Drop() error { return nil }
func (s *uniqueItemStub) FindByName(string) ([]*items.Item, error) {
	return nil, nil
}
func (s *uniqueItemStub) FindAll(r.ItemsQuery) ([]*items.Item, error) { return nil, nil }
func (s *uniqueItemStub) Update(string, *items.Item) error            { return nil }
func (s *uniqueItemStub) Delete(string) error                         { return nil }
func (s *uniqueItemStub) Store(item *items.Item) (*items.Item, error) { return item, nil }
func (s *uniqueItemStub) Import(item *items.Item) (*items.Item, error) {
	return item, nil
}
func (s *uniqueItemStub) FindAllTemplates(r.ItemsQuery) ([]*items.Item, error) {
	return nil, nil
}
func (s *uniqueItemStub) FindAllInstances(r.ItemsQuery) ([]*items.Item, error) {
	return nil, nil
}
func (s *uniqueItemStub) FindTemplateByName(string) ([]*items.Item, error) { return nil, nil }
func (s *uniqueItemStub) FindByTemplateID(string) ([]*items.Item, error)   { return nil, nil }
func (s *uniqueItemStub) ItemSlots() items.ItemSlots                       { return nil }
func (s *uniqueItemStub) ItemQualities() items.ItemQualities               { return nil }
func (s *uniqueItemStub) ItemTypes() items.ItemTypes                       { return nil }
func (s *uniqueItemStub) ItemSubTypes() items.ItemSubTypes                 { return nil }

type missingTemplate string

func (e missingTemplate) Error() string { return "missing " + string(e) }

func errMissingTemplate(id string) error { return missingTemplate(id) }

func uniqueTable(id string, chance float64, rarity string, bossOnly bool) *items.LootTable {
	return &items.LootTable{
		GoldMultiplier: 1,
		Entries: []items.LootEntry{{
			ItemTemplateID: id,
			DropChance:     chance,
			MinQuantity:    1,
			MaxQuantity:    1,
			Rarity:         rarity,
			BossOnly:       bossOnly,
		}},
	}
}

func (s *uniqueItemStub) service() *lootTablesService {
	return &lootTablesService{itemsService: s}
}

func rollCount(t *testing.T, srv *lootTablesService, table *items.LootTable, n int, seed int64, ctx LootRollContext) int {
	t.Helper()
	rng := rand.New(rand.NewSource(seed))
	ctx.Float64 = rng.Float64
	ctx.Intn = rng.Intn
	hits := 0
	for i := 0; i < n; i++ {
		res, err := srv.RollLootFromTableInContext(table, 1, 0, ctx)
		if err != nil {
			t.Fatal(err)
		}
		hits += len(res.Items)
	}
	return hits
}

func TestUniqueDropRatesWithinThreePoints(t *testing.T) {
	// Seed 1, 200 independent rolls, one Float64 each. Counts: 24, 12, 16.
	// Bands are the authored chances ±3 percentage points.
	const seed = 1
	const n = 200
	srv := (&uniqueItemStub{templates: map[string]*items.Item{
		"REL": {Entity: &entities.Entity{ID: "REL"}, Name: "Relic", Unique: true},
	}}).service()
	cases := []struct {
		chance float64
		want   int
	}{
		{0.10, 24},
		{0.06, 12},
		{0.08, 16},
	}
	for _, tc := range cases {
		got := rollCount(t, srv, uniqueTable("REL", tc.chance, "unique", true), n, seed, LootRollContext{Boss: true})
		rate := float64(got) / float64(n)
		if rate < tc.chance-0.03 || rate > tc.chance+0.03 {
			t.Errorf("chance %.2f rate %.3f (%d/%d) outside ±3pp", tc.chance, rate, got, n)
		}
		if got != tc.want {
			t.Errorf("chance %.2f hits %d, seed %d previously landed %d", tc.chance, got, seed, tc.want)
		}
	}
}

func TestUniqueRollSkipsWhenOwnedAndWhenNotBoss(t *testing.T) {
	srv := (&uniqueItemStub{templates: map[string]*items.Item{
		"REL":   {Entity: &entities.Entity{ID: "REL"}, Name: "Relic", Unique: true},
		"PLAIN": {Entity: &entities.Entity{ID: "PLAIN"}, Name: "Scrap"},
	}}).service()

	owned := rollCount(t, srv, uniqueTable("REL", 1, "unique", true), 200, 1, LootRollContext{
		Boss: true,
		Owns: func(string) bool { return true },
	})
	if owned != 0 {
		t.Fatalf("owned unique dropped %d times", owned)
	}

	open := rollCount(t, srv, uniqueTable("REL", 1, "unique", true), 1, 1, LootRollContext{Boss: true})
	if open != 1 {
		t.Fatalf("unowned unique hits %d", open)
	}

	held := false
	drops := 0
	rng := rand.New(rand.NewSource(1))
	table := uniqueTable("REL", 1, "unique", true)
	for i := 0; i < 200; i++ {
		res, err := srv.RollLootFromTableInContext(table, 1, 0, LootRollContext{
			Boss:    true,
			Owns:    func(string) bool { return held },
			Float64: rng.Float64,
			Intn:    rng.Intn,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Items) > 0 {
			drops++
			held = true
			if !res.Items[0].Unique {
				t.Fatal("dropped item lost unique flag")
			}
		}
	}
	if drops != 1 {
		t.Fatalf("accumulating owner got %d copies", drops)
	}

	skipped := rollCount(t, srv, uniqueTable("REL", 1, "unique", true), 20, 1, LootRollContext{Boss: false})
	if skipped != 0 {
		t.Fatalf("boss-only unique dropped off a non-boss: %d", skipped)
	}

	plain := uniqueTable("PLAIN", 1, "", false)
	plainHits := rollCount(t, srv, plain, 5, 1, LootRollContext{
		Owns: func(string) bool { return true },
	})
	if plainHits != 5 {
		t.Fatalf("ordinary drop blocked by ownership: %d", plainHits)
	}

	// Template unique flag gates the roll even when the entry has no rarity tag.
	flagged := uniqueTable("REL", 1, "", false)
	blocked := rollCount(t, srv, flagged, 10, 1, LootRollContext{
		Owns: func(string) bool { return true },
	})
	if blocked != 0 {
		t.Fatalf("template unique still dropped: %d", blocked)
	}
}

func TestBossOnlyDoesNotConsumeTheNextRoll(t *testing.T) {
	srv := (&uniqueItemStub{templates: map[string]*items.Item{
		"REL":   {Entity: &entities.Entity{ID: "REL"}, Name: "Relic", Unique: true},
		"PLAIN": {Entity: &entities.Entity{ID: "PLAIN"}, Name: "Scrap"},
	}}).service()
	table := &items.LootTable{
		GoldMultiplier: 1,
		Entries: []items.LootEntry{
			{ItemTemplateID: "REL", DropChance: 1, MinQuantity: 1, MaxQuantity: 1, Rarity: "unique", BossOnly: true},
			{ItemTemplateID: "PLAIN", DropChance: 1, MinQuantity: 1, MaxQuantity: 1},
		},
	}
	res, err := srv.RollLootFromTableInContext(table, 1, 0, LootRollContext{
		Boss:    false,
		Float64: func() float64 { return 0 },
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Items) != 1 || res.Items[0].TemplateID != "PLAIN" {
		t.Fatalf("items %#v", res.Items)
	}
}
