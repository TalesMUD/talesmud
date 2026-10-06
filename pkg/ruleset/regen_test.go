package ruleset_test

import (
	"strings"
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
	err := ruleset.LoadBytes([]byte("regen:\n  out_of_combat:\n    hp:\n      enabled: false\n"))
	if err != nil {
		t.Fatal(err)
	}
	hp, mana := ruleset.OutOfCombatRegen()
	if hp.Enabled || hp.Percent != 2 || hp.Flat != 0 || hp.IntervalSeconds != 10 {
		t.Fatalf("hp %+v", hp)
	}
	if !mana.Enabled || mana.Percent != 5 || mana.Flat != 0 || mana.IntervalSeconds != 10 {
		t.Fatalf("mana %+v", mana)
	}
}

func TestRegenExplicitValues(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	err := ruleset.LoadBytes([]byte(`
regen:
  out_of_combat:
    hp:
      enabled: true
      percent: 0.5
      flat: 3
      interval_seconds: 60
    mana:
      enabled: false
      percent: 1
      flat: 2
      interval_seconds: 30
`))
	if err != nil {
		t.Fatal(err)
	}
	hp, mana := ruleset.OutOfCombatRegen()
	if !hp.Enabled || hp.Percent != 0.5 || hp.Flat != 3 || hp.IntervalSeconds != 60 {
		t.Fatalf("hp %+v", hp)
	}
	if mana.Enabled || mana.Percent != 1 || mana.Flat != 2 || mana.IntervalSeconds != 30 {
		t.Fatalf("mana %+v", mana)
	}
}

func TestRegenRejectsInvalid(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	cases := []string{
		"regen:\n  out_of_combat:\n    hp:\n      percent: -0.1\n",
		"regen:\n  out_of_combat:\n    mana:\n      flat: -1\n",
		"regen:\n  out_of_combat:\n    hp:\n      interval_seconds: 0\n",
		"regen:\n  out_of_combat:\n    mana:\n      interval_seconds: -5\n",
	}
	for _, body := range cases {
		err := ruleset.LoadBytes([]byte(body))
		if err == nil || !strings.Contains(err.Error(), "regen") {
			t.Fatalf("body %q err %v", body, err)
		}
		assertDefaultRegen(t)
	}
}

func assertDefaultRegen(t *testing.T) {
	t.Helper()
	hp, mana := ruleset.OutOfCombatRegen()
	if !hp.Enabled || hp.Percent != 2 || hp.Flat != 0 || hp.IntervalSeconds != 10 {
		t.Fatalf("hp %+v", hp)
	}
	if !mana.Enabled || mana.Percent != 5 || mana.Flat != 0 || mana.IntervalSeconds != 10 {
		t.Fatalf("mana %+v", mana)
	}
}
