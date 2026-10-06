package game

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/ruleset"
)

func defaultRegenPolicies() (ruleset.RegenPolicy, ruleset.RegenPolicy) {
	return ruleset.RegenPolicy{Enabled: true, Percent: 2, Flat: 0, IntervalSeconds: 10},
		ruleset.RegenPolicy{Enabled: true, Percent: 5, Flat: 0, IntervalSeconds: 10}
}

func legacyPool(max int32, rate float64) int32 {
	amount := int32(float64(max) * rate)
	if amount < 1 {
		amount = 1
	}
	return amount
}

func living(maxHP, maxMana int32) *characters.Character {
	return &characters.Character{
		MaxHitPoints:     maxHP,
		CurrentHitPoints: 1,
		MaxMana:          maxMana,
		CurrentMana:      0,
	}
}

func TestPassiveRegenMatchesLegacyOnTens(t *testing.T) {
	hpPolicy, manaPolicy := defaultRegenPolicies()
	for max := int32(1); max <= 5000; max++ {
		char := living(max, max)
		gotHP, gotMana := regenAmounts(char, 10, hpPolicy, manaPolicy)
		wantHP := legacyPool(max, 0.02)
		wantMana := legacyPool(max, 0.05)
		if gotHP != wantHP || gotMana != wantMana {
			t.Fatalf("max %d hp %d want %d mana %d want %d", max, gotHP, wantHP, gotMana, wantMana)
		}
		for _, tick := range []uint64{9, 11, 19} {
			offHP, offMana := regenAmounts(char, tick, hpPolicy, manaPolicy)
			if offHP != 0 || offMana != 0 {
				t.Fatalf("max %d tick %d hp %d mana %d", max, tick, offHP, offMana)
			}
		}
		onHP, onMana := regenAmounts(char, 20, hpPolicy, manaPolicy)
		if onHP != wantHP || onMana != wantMana {
			t.Fatalf("max %d tick 20 hp %d mana %d", max, onHP, onMana)
		}
	}
}

func TestPassiveRegenDisabledHP(t *testing.T) {
	hpPolicy, manaPolicy := defaultRegenPolicies()
	hpPolicy.Enabled = false
	char := living(250, 250)
	gotHP, gotMana := regenAmounts(char, 10, hpPolicy, manaPolicy)
	if gotHP != 0 || gotMana != legacyPool(250, 0.05) {
		t.Fatalf("hp %d mana %d", gotHP, gotMana)
	}
	gotHP, gotMana = regenAmounts(char, 20, hpPolicy, manaPolicy)
	if gotHP != 0 || gotMana != legacyPool(250, 0.05) {
		t.Fatalf("tick 20 hp %d mana %d", gotHP, gotMana)
	}
}

func TestPassiveRegenCustomPercentFlatInterval(t *testing.T) {
	hpPolicy := ruleset.RegenPolicy{Enabled: true, Percent: 0.5, Flat: 3, IntervalSeconds: 60}
	manaPolicy := ruleset.RegenPolicy{Enabled: true, Percent: 1, Flat: 0, IntervalSeconds: 10}
	char := living(2000, 2000)
	wantHP := int32(float64(2000)*(0.5/100)) + 3
	wantMana := legacyPool(2000, 0.01)

	gotHP, gotMana := regenAmounts(char, 60, hpPolicy, manaPolicy)
	if gotHP != wantHP || gotMana != wantMana {
		t.Fatalf("tick 60 hp %d want %d mana %d want %d", gotHP, wantHP, gotMana, wantMana)
	}
	gotHP, gotMana = regenAmounts(char, 10, hpPolicy, manaPolicy)
	if gotHP != 0 || gotMana != wantMana {
		t.Fatalf("tick 10 hp %d mana %d", gotHP, gotMana)
	}
	gotHP, gotMana = regenAmounts(char, 30, hpPolicy, manaPolicy)
	if gotHP != 0 || gotMana != wantMana {
		t.Fatalf("tick 30 hp %d mana %d", gotHP, gotMana)
	}
	gotHP, gotMana = regenAmounts(char, 59, hpPolicy, manaPolicy)
	if gotHP != 0 || gotMana != 0 {
		t.Fatalf("tick 59 hp %d mana %d", gotHP, gotMana)
	}

	small := living(1, 1)
	gotHP, _ = regenAmounts(small, 60, hpPolicy, manaPolicy)
	// 0.5% of 1 truncates to 0, then flat 3. The sum is already above the minimum.
	if gotHP != 3 {
		t.Fatalf("small hp %d", gotHP)
	}
	noFlat := hpPolicy
	noFlat.Flat = 0
	gotHP, _ = regenAmounts(small, 60, noFlat, manaPolicy)
	if gotHP != 1 {
		t.Fatalf("truncated percent hp %d", gotHP)
	}

	flatOnly := ruleset.RegenPolicy{Enabled: true, Percent: 0, Flat: 4, IntervalSeconds: 10}
	zero := ruleset.RegenPolicy{Enabled: true, Percent: 0, Flat: 0, IntervalSeconds: 10}
	gotHP, _ = regenAmounts(char, 10, flatOnly, zero)
	if gotHP != 4 {
		t.Fatalf("flat hp %d", gotHP)
	}
	gotHP, gotMana = regenAmounts(char, 10, zero, zero)
	if gotHP != 0 || gotMana != 0 {
		t.Fatalf("zero rates hp %d mana %d", gotHP, gotMana)
	}
}

func TestCombatIgnoresPassiveRegenConfig(t *testing.T) {
	hpPolicy := ruleset.RegenPolicy{Enabled: false, Percent: 50, Flat: 9, IntervalSeconds: 60}
	manaPolicy := ruleset.RegenPolicy{Enabled: false, Percent: 50, Flat: 9, IntervalSeconds: 60}
	char := living(2000, 2000)
	char.InCombat = true
	char.Flags = map[string]interface{}{"resting": true}

	gotHP, gotMana := regenAmounts(char, 10, hpPolicy, manaPolicy)
	if gotHP != legacyPool(2000, 0.005) || gotMana != legacyPool(2000, 0.01) {
		t.Fatalf("tick 10 hp %d mana %d", gotHP, gotMana)
	}
	gotHP, gotMana = regenAmounts(char, 60, hpPolicy, manaPolicy)
	if gotHP != legacyPool(2000, 0.005) || gotMana != legacyPool(2000, 0.01) {
		t.Fatalf("tick 60 hp %d mana %d", gotHP, gotMana)
	}
	gotHP, gotMana = regenAmounts(char, 7, hpPolicy, manaPolicy)
	if gotHP != 0 || gotMana != 0 {
		t.Fatalf("tick 7 hp %d mana %d", gotHP, gotMana)
	}
	char.CurrentMana = char.MaxMana
	if _, gotMana = regenAmounts(char, 10, hpPolicy, manaPolicy); gotMana != 0 {
		t.Fatalf("full mana %d", gotMana)
	}
}

func TestRestingIgnoresPassiveRegenConfig(t *testing.T) {
	hpPolicy := ruleset.RegenPolicy{Enabled: false, Percent: 50, Flat: 9, IntervalSeconds: 60}
	manaPolicy := ruleset.RegenPolicy{Enabled: true, Percent: 0.5, Flat: 3, IntervalSeconds: 60}
	char := living(2000, 2000)
	char.Flags = map[string]interface{}{"resting": true}

	gotHP, gotMana := regenAmounts(char, 10, hpPolicy, manaPolicy)
	if gotHP != legacyPool(2000, 0.10) || gotMana != legacyPool(2000, 0.15) {
		t.Fatalf("tick 10 hp %d mana %d", gotHP, gotMana)
	}
	gotHP, gotMana = regenAmounts(char, 60, hpPolicy, manaPolicy)
	if gotHP != legacyPool(2000, 0.10) || gotMana != legacyPool(2000, 0.15) {
		t.Fatalf("tick 60 hp %d mana %d", gotHP, gotMana)
	}
	gotHP, gotMana = regenAmounts(char, 7, hpPolicy, manaPolicy)
	if gotHP != 0 || gotMana != 0 {
		t.Fatalf("tick 7 hp %d mana %d", gotHP, gotMana)
	}
}
