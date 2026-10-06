package ruleset_test

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/ruleset"
)

func TestRegenAbsentKeepsDefaults(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	if err := ruleset.LoadBytes([]byte("progression:\n  level_cap: 50\n")); err != nil {
		t.Fatal(err)
	}
	assertDefaultRegen(t)
}

func TestRegenPartialOverride(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	err := ruleset.LoadBytes([]byte(`
regen:
  out_of_combat:
    hp:
      enabled: false
  resting:
    mana:
      percent: 20
  in_combat:
    hp:
      interval_seconds: 30
`))
	if err != nil {
		t.Fatal(err)
	}
	got := ruleset.Regen()
	if got.OutOfCombat.HP.Enabled || got.OutOfCombat.HP.Percent != 2 || got.OutOfCombat.HP.Flat != 0 || got.OutOfCombat.HP.IntervalSeconds != 10 {
		t.Fatalf("ooc hp %+v", got.OutOfCombat.HP)
	}
	if !got.OutOfCombat.Mana.Enabled || got.OutOfCombat.Mana.Percent != 5 || got.OutOfCombat.Mana.Flat != 0 || got.OutOfCombat.Mana.IntervalSeconds != 10 {
		t.Fatalf("ooc mana %+v", got.OutOfCombat.Mana)
	}
	if !got.Resting.HP.Enabled || got.Resting.HP.Percent != 10 || got.Resting.HP.Flat != 0 || got.Resting.HP.IntervalSeconds != 10 {
		t.Fatalf("resting hp %+v", got.Resting.HP)
	}
	if !got.Resting.Mana.Enabled || got.Resting.Mana.Percent != 20 || got.Resting.Mana.Flat != 0 || got.Resting.Mana.IntervalSeconds != 10 {
		t.Fatalf("resting mana %+v", got.Resting.Mana)
	}
	if !got.InCombat.HP.Enabled || got.InCombat.HP.Percent != 0.5 || got.InCombat.HP.Flat != 0 || got.InCombat.HP.IntervalSeconds != 30 {
		t.Fatalf("combat hp %+v", got.InCombat.HP)
	}
	if !got.InCombat.Mana.Enabled || got.InCombat.Mana.Percent != 1 || got.InCombat.Mana.Flat != 0 || got.InCombat.Mana.IntervalSeconds != 10 {
		t.Fatalf("combat mana %+v", got.InCombat.Mana)
	}
}

func TestRegenExplicitRestingAndCombat(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	err := ruleset.LoadBytes([]byte(`
regen:
  resting:
    hp:
      enabled: true
      percent: 12.5
      flat: 4
      interval_seconds: 45
    mana:
      enabled: false
      percent: 8
      flat: 1
      interval_seconds: 20
  in_combat:
    hp:
      enabled: false
      percent: 0.25
      flat: 6
      interval_seconds: 15
    mana:
      enabled: true
      percent: 3
      flat: 2
      interval_seconds: 5
`))
	if err != nil {
		t.Fatal(err)
	}
	got := ruleset.Regen()
	hp, mana := ruleset.OutOfCombatRegen()
	if hp != got.OutOfCombat.HP || mana != got.OutOfCombat.Mana {
		t.Fatalf("getter hp %+v mana %+v profile %+v", hp, mana, got.OutOfCombat)
	}
	if !got.OutOfCombat.HP.Enabled || got.OutOfCombat.HP.Percent != 2 || got.OutOfCombat.Mana.Percent != 5 {
		t.Fatalf("ooc changed %+v", got.OutOfCombat)
	}
	if !got.Resting.HP.Enabled || got.Resting.HP.Percent != 12.5 || got.Resting.HP.Flat != 4 || got.Resting.HP.IntervalSeconds != 45 {
		t.Fatalf("resting hp %+v", got.Resting.HP)
	}
	if got.Resting.Mana.Enabled || got.Resting.Mana.Percent != 8 || got.Resting.Mana.Flat != 1 || got.Resting.Mana.IntervalSeconds != 20 {
		t.Fatalf("resting mana %+v", got.Resting.Mana)
	}
	if got.InCombat.HP.Enabled || got.InCombat.HP.Percent != 0.25 || got.InCombat.HP.Flat != 6 || got.InCombat.HP.IntervalSeconds != 15 {
		t.Fatalf("combat hp %+v", got.InCombat.HP)
	}
	if !got.InCombat.Mana.Enabled || got.InCombat.Mana.Percent != 3 || got.InCombat.Mana.Flat != 2 || got.InCombat.Mana.IntervalSeconds != 5 {
		t.Fatalf("combat mana %+v", got.InCombat.Mana)
	}
}

func TestRegenRejectsInvalid(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	cases := []struct {
		body string
		want string
	}{
		{"regen:\n  out_of_combat:\n    hp:\n      percent: -0.1\n", "regen out_of_combat.hp.percent must be >= 0"},
		{"regen:\n  out_of_combat:\n    mana:\n      flat: -1\n", "regen out_of_combat.mana.flat must be >= 0"},
		{"regen:\n  out_of_combat:\n    hp:\n      interval_seconds: 0\n", "regen out_of_combat.hp.interval_seconds must be >= 1"},
		{"regen:\n  out_of_combat:\n    mana:\n      interval_seconds: -5\n", "regen out_of_combat.mana.interval_seconds must be >= 1"},
		{"regen:\n  resting:\n    hp:\n      percent: -0.1\n", "regen resting.hp.percent must be >= 0"},
		{"regen:\n  resting:\n    mana:\n      flat: -1\n", "regen resting.mana.flat must be >= 0"},
		{"regen:\n  resting:\n    hp:\n      interval_seconds: 0\n", "regen resting.hp.interval_seconds must be >= 1"},
		{"regen:\n  in_combat:\n    hp:\n      percent: -0.5\n", "regen in_combat.hp.percent must be >= 0"},
		{"regen:\n  in_combat:\n    mana:\n      flat: -1\n", "regen in_combat.mana.flat must be >= 0"},
		{"regen:\n  in_combat:\n    mana:\n      interval_seconds: -5\n", "regen in_combat.mana.interval_seconds must be >= 1"},
	}
	for _, tc := range cases {
		err := ruleset.LoadBytes([]byte(tc.body))
		if err == nil || err.Error() != tc.want {
			t.Fatalf("body %q err %v", tc.body, err)
		}
		assertDefaultRegen(t)
	}
}

func TestRegenPolicyActive(t *testing.T) {
	if !(ruleset.RegenPolicy{Enabled: true, Percent: 2}).Active() {
		t.Fatal("percent")
	}
	if !(ruleset.RegenPolicy{Enabled: true, Flat: 1}).Active() {
		t.Fatal("flat")
	}
	if (ruleset.RegenPolicy{Enabled: false, Percent: 2, Flat: 1}).Active() {
		t.Fatal("disabled")
	}
	if (ruleset.RegenPolicy{Enabled: true, Percent: 0, Flat: 0}).Active() {
		t.Fatal("zero rates")
	}
}

func assertDefaultRegen(t *testing.T) {
	t.Helper()
	got := ruleset.Regen()
	want := ruleset.RegenProfile{
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
	if got != want {
		t.Fatalf("regen %+v", got)
	}
	hp, mana := ruleset.OutOfCombatRegen()
	if hp != want.OutOfCombat.HP || mana != want.OutOfCombat.Mana {
		t.Fatalf("out of combat hp %+v mana %+v", hp, mana)
	}
	if !hp.Active() || !mana.Active() || !got.Resting.HP.Active() || !got.InCombat.Mana.Active() {
		t.Fatal("defaults should be active")
	}
}
