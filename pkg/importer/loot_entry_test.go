package importer

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYAMLLootEntryChanceOverridesDropChance(t *testing.T) {
	raw := []byte(`
id: LT1
name: Relics
entries:
  - itemTemplateId: REL
    dropChance: 0.5
    chance: 0.08
    minQuantity: 2
    maxQuantity: 4
    rarity: unique
    boss_only: true
  - itemTemplateId: PLAIN
    dropChance: 0.1
  - itemTemplateId: ALT
    dropChance: 0.2
    bossOnly: true
`)
	var table YAMLLootTable
	if err := yaml.Unmarshal(raw, &table); err != nil {
		t.Fatal(err)
	}
	got := table.ToEntity()
	if len(got.Entries) != 3 {
		t.Fatalf("entries %d", len(got.Entries))
	}
	first := got.Entries[0]
	if first.DropChance != 0.08 || first.Rarity != "unique" || !first.BossOnly {
		t.Fatalf("first %+v", first)
	}
	if first.MinQuantity != 2 || first.MaxQuantity != 4 {
		t.Fatalf("qty %+v", first)
	}
	if got.Entries[1].DropChance != 0.1 || got.Entries[1].BossOnly || got.Entries[1].Rarity != "" {
		t.Fatalf("plain %+v", got.Entries[1])
	}
	if !got.Entries[2].BossOnly || got.Entries[2].DropChance != 0.2 {
		t.Fatalf("alt %+v", got.Entries[2])
	}
}
