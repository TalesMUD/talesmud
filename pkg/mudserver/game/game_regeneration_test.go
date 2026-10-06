package game

import (
	"math"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/ruleset"
)

func defaultRegenProfile() ruleset.RegenProfile {
	return ruleset.RegenProfile{
		OutOfCombat: ruleset.RegenPair{
			HP:   ruleset.RegenPolicy{Enabled: true, Percent: 2, Flat: 0, IntervalSeconds: 10},
			Mana: ruleset.RegenPolicy{Enabled: true, Percent: 5, Flat: 0, IntervalSeconds: 10},
		},
		Resting: ruleset.RegenPair{
			HP:   ruleset.RegenPolicy{Enabled: true, Percent: 10, Flat: 0, IntervalSeconds: 10},
			Mana: ruleset.RegenPolicy{Enabled: true, Percent: 15, Flat: 0, IntervalSeconds: 10},
		},
		InCombat: ruleset.RegenPair{
			HP:   ruleset.RegenPolicy{Enabled: true, Percent: 0.5, Flat: 0, IntervalSeconds: 10},
			Mana: ruleset.RegenPolicy{Enabled: true, Percent: 1, Flat: 0, IntervalSeconds: 10},
		},
	}
}

func legacyPool(max int32, rate float64) int32 {
	amount := int32(float64(max) * rate)
	if amount < 1 {
		amount = 1
	}
	return amount
}

// living is below full HP so the full-HP guard does not hide the pool formula.
func living(maxHP, maxMana int32) *characters.Character {
	return &characters.Character{
		MaxHitPoints:     maxHP,
		CurrentHitPoints: 0,
		MaxMana:          maxMana,
		CurrentMana:      0,
	}
}

func TestLegacyRatesAreBitIdentical(t *testing.T) {
	pairs := [][2]float64{
		{2.0 / 100, 0.02},
		{5.0 / 100, 0.05},
		{10.0 / 100, 0.10},
		{15.0 / 100, 0.15},
		{0.5 / 100, 0.005},
		{1.0 / 100, 0.01},
	}
	for _, pair := range pairs {
		if math.Float64bits(pair[0]) != math.Float64bits(pair[1]) {
			t.Fatalf("%v (%x) != %v (%x)", pair[0], math.Float64bits(pair[0]), pair[1], math.Float64bits(pair[1]))
		}
	}
}

func TestDefaultRegenMatchesLegacyOnTens(t *testing.T) {
	profile := defaultRegenProfile()
	for max := int32(1); max <= 5000; max++ {
		states := []struct {
			name string
			char *characters.Character
			hp   int32
			mana int32
		}{
			{"ooc", living(max, max), legacyPool(max, 0.02), legacyPool(max, 0.05)},
			{"rest", restingChar(max), legacyPool(max, 0.10), legacyPool(max, 0.15)},
			{"combat", fightingChar(max), legacyPool(max, 0.005), legacyPool(max, 0.01)},
		}
		for _, state := range states {
			for _, tick := range []uint64{10, 20} {
				gotHP, gotMana := regenAmounts(state.char, tick, profile)
				if gotHP != state.hp || gotMana != state.mana {
					t.Fatalf("%s max %d tick %d hp %d want %d mana %d want %d", state.name, max, tick, gotHP, state.hp, gotMana, state.mana)
				}
			}
			for _, tick := range []uint64{9, 11, 19} {
				gotHP, gotMana := regenAmounts(state.char, tick, profile)
				if gotHP != 0 || gotMana != 0 {
					t.Fatalf("%s max %d tick %d hp %d mana %d", state.name, max, tick, gotHP, gotMana)
				}
			}
		}
	}
}

func TestDisabledPoolIsZeroAndNotDue(t *testing.T) {
	disabled := ruleset.RegenPolicy{Enabled: false, Percent: 10, Flat: 4, IntervalSeconds: 1}
	if disabled.Active() || poolDue(1, disabled) || poolDue(10, disabled) {
		t.Fatal("disabled pool is due")
	}
	profile := defaultRegenProfile()
	profile.Resting.HP = disabled
	char := restingChar(100)
	hp, mana := regenAmounts(char, 1, profile)
	if hp != 0 || mana != 0 {
		t.Fatalf("tick 1 hp %d mana %d", hp, mana)
	}
	hp, mana = regenAmounts(char, 10, profile)
	if hp != 0 || mana != legacyPool(100, 0.15) {
		t.Fatalf("tick 10 hp %d mana %d", hp, mana)
	}
}

func TestZeroPercentAndFlatAreNotDue(t *testing.T) {
	zero := ruleset.RegenPolicy{Enabled: true, Percent: 0, Flat: 0, IntervalSeconds: 1}
	if zero.Active() {
		t.Fatal("zero rates are active")
	}
	for tick := uint64(1); tick <= 5; tick++ {
		if poolDue(tick, zero) {
			t.Fatalf("pool tick %d", tick)
		}
	}
	profile := defaultRegenProfile()
	profile.OutOfCombat.HP = zero
	profile.OutOfCombat.Mana = ruleset.RegenPolicy{Enabled: true, Percent: 0, Flat: 4, IntervalSeconds: 10}
	char := living(2000, 2000)
	hp, mana := regenAmounts(char, 10, profile)
	if hp != 0 || mana != 4 {
		t.Fatalf("flat hp %d mana %d", hp, mana)
	}
	hp, mana = regenAmounts(char, 1, profile)
	if hp != 0 || mana != 0 {
		t.Fatalf("off tick hp %d mana %d", hp, mana)
	}
}

func TestCustomRestingAndCombatRegen(t *testing.T) {
	profile := defaultRegenProfile()
	profile.OutOfCombat.HP = ruleset.RegenPolicy{Enabled: false, Percent: 50, Flat: 9, IntervalSeconds: 1}
	profile.OutOfCombat.Mana = profile.OutOfCombat.HP
	profile.Resting.HP = ruleset.RegenPolicy{Enabled: true, Percent: 0.5, Flat: 3, IntervalSeconds: 60}
	profile.Resting.Mana = ruleset.RegenPolicy{Enabled: true, Percent: 1, Flat: 0, IntervalSeconds: 10}
	profile.InCombat.HP = ruleset.RegenPolicy{Enabled: true, Percent: 4, Flat: 2, IntervalSeconds: 15}
	profile.InCombat.Mana = ruleset.RegenPolicy{Enabled: true, Percent: 0, Flat: 7, IntervalSeconds: 15}

	resting := restingChar(2000)
	wantHP := int32(float64(2000)*(0.5/100)) + 3
	wantMana := legacyPool(2000, 0.01)
	hp, mana := regenAmounts(resting, 60, profile)
	if hp != wantHP || mana != wantMana {
		t.Fatalf("rest 60 hp %d want %d mana %d want %d", hp, wantHP, mana, wantMana)
	}
	hp, mana = regenAmounts(resting, 10, profile)
	if hp != 0 || mana != wantMana {
		t.Fatalf("rest 10 hp %d mana %d", hp, mana)
	}
	hp, mana = regenAmounts(resting, 59, profile)
	if hp != 0 || mana != 0 {
		t.Fatalf("rest 59 hp %d mana %d", hp, mana)
	}

	small := restingChar(1)
	hp, _ = regenAmounts(small, 60, profile)
	if hp != 3 {
		t.Fatalf("small rest hp %d", hp)
	}
	noFlat := profile
	noFlat.Resting.HP.Flat = 0
	hp, _ = regenAmounts(small, 60, noFlat)
	if hp != 1 {
		t.Fatalf("min rest hp %d", hp)
	}

	fighting := fightingChar(2000)
	wantHP = int32(float64(2000)*(4.0/100)) + 2
	hp, mana = regenAmounts(fighting, 15, profile)
	if hp != wantHP || mana != 7 {
		t.Fatalf("combat 15 hp %d want %d mana %d", hp, wantHP, mana)
	}
	hp, mana = regenAmounts(fighting, 10, profile)
	if hp != 0 || mana != 0 {
		t.Fatalf("combat 10 hp %d mana %d", hp, mana)
	}
	hp, mana = regenAmounts(fighting, 60, profile)
	if hp != wantHP || mana != 7 {
		t.Fatalf("combat 60 hp %d mana %d", hp, mana)
	}
}

func TestFullHPGuardSkipsHPOnlyTick(t *testing.T) {
	profile := defaultRegenProfile()
	slowMana := func(pair ruleset.RegenPair) ruleset.RegenPair {
		pair.Mana.IntervalSeconds = 60
		return pair
	}
	profile.OutOfCombat = slowMana(profile.OutOfCombat)
	profile.Resting = slowMana(profile.Resting)
	profile.InCombat = slowMana(profile.InCombat)

	cases := []struct {
		name string
		char *characters.Character
		rate float64
	}{
		{"ooc", living(2000, 2000), 0.05},
		{"rest", restingChar(2000), 0.15},
		{"combat", fightingChar(2000), 0.01},
	}
	for _, tc := range cases {
		tc.char.CurrentHitPoints = tc.char.MaxHitPoints
		hp, mana := regenAmounts(tc.char, 10, profile)
		if hp != 0 || mana != 0 {
			t.Fatalf("%s tick 10 hp %d mana %d", tc.name, hp, mana)
		}
		hp, mana = regenAmounts(tc.char, 60, profile)
		if hp != 0 || mana != legacyPool(2000, tc.rate) {
			t.Fatalf("%s tick 60 hp %d mana %d", tc.name, hp, mana)
		}
	}

	partial := living(2000, 2000)
	partial.CurrentHitPoints = 10
	partial.CurrentMana = partial.MaxMana
	hp, mana := regenAmounts(partial, 10, profile)
	if hp != legacyPool(2000, 0.02) || mana != 0 {
		t.Fatalf("partial hp %d mana %d", hp, mana)
	}
}

func TestNothingDueWhenEveryPoolDisabled(t *testing.T) {
	off := func(p ruleset.RegenPolicy) ruleset.RegenPolicy {
		p.Enabled = false
		p.IntervalSeconds = 1
		return p
	}
	profile := defaultRegenProfile()
	profile.OutOfCombat.HP = off(profile.OutOfCombat.HP)
	profile.OutOfCombat.Mana = off(profile.OutOfCombat.Mana)
	profile.Resting.HP = off(profile.Resting.HP)
	profile.Resting.Mana = off(profile.Resting.Mana)
	profile.InCombat.HP = off(profile.InCombat.HP)
	profile.InCombat.Mana = off(profile.InCombat.Mana)
	for tick := uint64(1); tick <= 120; tick++ {
		if anyPoolDue(tick, profile) {
			t.Fatalf("tick %d", tick)
		}
	}
	char := fightingChar(500)
	hp, mana := regenAmounts(char, 1, profile)
	if hp != 0 || mana != 0 {
		t.Fatalf("amount hp %d mana %d", hp, mana)
	}
}

func TestInCombatUsesInCombatPoolsEvenWhenResting(t *testing.T) {
	profile := defaultRegenProfile()
	profile.InCombat.HP.Percent = 3
	profile.InCombat.HP.Flat = 2
	profile.InCombat.Mana.Percent = 4
	profile.InCombat.Mana.Flat = 1
	profile.Resting.HP.Percent = 50
	profile.Resting.Mana.Percent = 50
	char := fightingChar(2000)
	hp, mana := regenAmounts(char, 10, profile)
	wantHP := int32(float64(2000)*(3.0/100)) + 2
	wantMana := int32(float64(2000)*(4.0/100)) + 1
	if hp != wantHP || mana != wantMana {
		t.Fatalf("hp %d want %d mana %d want %d", hp, wantHP, mana, wantMana)
	}
	char.InCombat = false
	hp, mana = regenAmounts(char, 10, profile)
	if hp != legacyPool(2000, 0.50) || mana != legacyPool(2000, 0.50) {
		t.Fatalf("resting fell through hp %d mana %d", hp, mana)
	}
}

func TestFullManaYieldsNoMana(t *testing.T) {
	profile := defaultRegenProfile()
	char := fightingChar(200)
	char.CurrentMana = char.MaxMana
	hp, mana := regenAmounts(char, 10, profile)
	if mana != 0 || hp != legacyPool(200, 0.005) {
		t.Fatalf("hp %d mana %d", hp, mana)
	}
	char.MaxMana = 0
	char.CurrentMana = 0
	_, mana = regenAmounts(char, 10, profile)
	if mana != 0 {
		t.Fatalf("no mana pool %d", mana)
	}
}

func restingChar(max int32) *characters.Character {
	char := living(max, max)
	char.Flags = map[string]interface{}{"resting": true}
	return char
}

func fightingChar(max int32) *characters.Character {
	char := restingChar(max)
	char.InCombat = true
	return char
}
