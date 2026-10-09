package exporter

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/importer"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
)

func TestContentYAMLRoundTripDoesNotDoubleScale(t *testing.T) {
	balance.ReloadConfig()

	original := &importer.YAMLNPC{
		ID:           "ENM0009",
		Name:         "The Hollow Knight",
		MaxHitPoints: 150,
		Level:        6,
		EnemyTrait: &importer.YAMLEnemyTrait{
			CreatureType:      "construct",
			CombatStyle:       "melee",
			Difficulty:        "boss",
			AttackPower:       13,
			Defense:           6,
			XPReward:          135,
			LootTableID:       "LT0204",
			OnLowHealthScript: "SCR0207",
			GoldDrop:          importer.YAMLRange{Min: 6, Max: 15},
			GuaranteedLoot:    []string{"ITM0017"},
			MaxDrops:          2,
			ResetOnDisengage:  boolPtr(false),
		},
	}

	imported := original.ToEntity()
	if imported.EnemyTrait == nil || imported.EnemyTrait.BaseStats == nil {
		t.Fatal("import did not store content base stats")
	}
	base := imported.EnemyTrait.BaseStats
	if base.MaxHitPoints != 150 || base.AttackPower != 13 || base.Defense != 6 {
		t.Fatalf("base = %+v, want 150/13/6", base)
	}
	if imported.MaxHitPoints != 150 || imported.EnemyTrait.AttackPower != 11 || imported.EnemyTrait.Defense != 5 {
		t.Fatalf("effective hp/atk/def = %d/%d/%d, want 150/11/5",
			imported.MaxHitPoints, imported.EnemyTrait.AttackPower, imported.EnemyTrait.Defense)
	}
	if imported.EnemyTrait.GuaranteedLoot[0] != "ITM0017" || imported.EnemyTrait.MaxDrops != 2 {
		t.Fatalf("loot fields dropped: %+v", imported.EnemyTrait)
	}

	maxHP, attack, defense, fromBase := ContentCombatStats(imported)
	if !fromBase || maxHP != 150 || attack != 13 || defense != 6 {
		t.Fatalf("export stats = %d/%d/%d fromBase=%v, want content 150/13/6", maxHP, attack, defense, fromBase)
	}

	exported := ContentNPC(imported)
	again := exported.ToEntity()
	if again.MaxHitPoints != imported.MaxHitPoints ||
		again.EnemyTrait.AttackPower != imported.EnemyTrait.AttackPower ||
		again.EnemyTrait.Defense != imported.EnemyTrait.Defense {
		t.Fatalf("round trip scaled twice: first %d/%d/%d second %d/%d/%d",
			imported.MaxHitPoints, imported.EnemyTrait.AttackPower, imported.EnemyTrait.Defense,
			again.MaxHitPoints, again.EnemyTrait.AttackPower, again.EnemyTrait.Defense)
	}
	if again.EnemyTrait.BaseStats.AttackPower != 13 || again.EnemyTrait.OnLowHealthScript != "SCR0207" {
		t.Fatalf("content fields not preserved: %+v", again.EnemyTrait)
	}
	if again.EnemyTrait.ResetOnDisengage == nil || *again.EnemyTrait.ResetOnDisengage {
		t.Fatal("resetOnDisengage was dropped on export")
	}
}

func boolPtr(v bool) *bool { return &v }

func TestContentYAMLUnknownTierKeepsBase(t *testing.T) {
	balance.ReloadConfig()

	y := &importer.YAMLNPC{
		ID:           "ENM0065",
		Name:         "Ley Leech",
		MaxHitPoints: 40,
		EnemyTrait: &importer.YAMLEnemyTrait{
			Difficulty:  "elite",
			AttackPower: 8,
			Defense:     3,
		},
	}
	n := y.ToEntity()
	if n.MaxHitPoints != 40 || n.EnemyTrait.AttackPower != 8 || n.EnemyTrait.Defense != 3 {
		t.Fatalf("unknown tier scaled stats: %d/%d/%d", n.MaxHitPoints, n.EnemyTrait.AttackPower, n.EnemyTrait.Defense)
	}
	maxHP, attack, defense, fromBase := ContentCombatStats(n)
	if !fromBase || maxHP != 40 || attack != 8 || defense != 3 {
		t.Fatalf("export = %d/%d/%d fromBase=%v", maxHP, attack, defense, fromBase)
	}
}

func TestApplyContentBaseKeepsWoundedHP(t *testing.T) {
	balance.ReloadConfig()

	n := (&importer.YAMLNPC{
		ID:           "ENM1",
		Name:         "Bandit",
		MaxHitPoints: 22,
		EnemyTrait: &importer.YAMLEnemyTrait{
			Difficulty:  "normal",
			AttackPower: 4,
			Defense:     1,
		},
	}).ToEntity()
	n.CurrentHitPoints = 10
	n.EnemyTrait.BaseStats.AttackPower = 5
	importer.ApplyContentBase(n)
	if n.CurrentHitPoints != 10 {
		t.Fatalf("wounded HP changed to %d", n.CurrentHitPoints)
	}
	if n.EnemyTrait.AttackPower == 4 {
		t.Fatalf("attack was not rescaled from the new base, still %d", n.EnemyTrait.AttackPower)
	}
}
