package ruleset

import "fmt"

// RegenPolicy is passive regeneration for one pool, outside combat and rest.
// Percent is a percent of that pool's maximum per tick, not a fraction.
type RegenPolicy struct {
	Enabled         bool
	Percent         float64
	Flat            int32
	IntervalSeconds int
}

type regenPoolFile struct {
	Enabled         *bool    `yaml:"enabled"`
	Percent         *float64 `yaml:"percent"`
	Flat            *int32   `yaml:"flat"`
	IntervalSeconds *int     `yaml:"interval_seconds"`
}

func defaultOutOfCombatRegen() (hp, mana RegenPolicy) {
	hp = RegenPolicy{Enabled: true, Percent: 2, Flat: 0, IntervalSeconds: 10}
	mana = RegenPolicy{Enabled: true, Percent: 5, Flat: 0, IntervalSeconds: 10}
	return hp, mana
}

// OutOfCombatRegen is the passive HP and mana policy.
// Resting and in-combat rates are not part of it.
func OutOfCombatRegen() (hp, mana RegenPolicy) {
	mu.RLock()
	defer mu.RUnlock()
	return current.oocHP, current.oocMana
}

func applyRegen(next *state, file fileShape) error {
	hp, err := mergeRegenPool("hp", file.Regen.OutOfCombat.HP, next.oocHP)
	if err != nil {
		return err
	}
	mana, err := mergeRegenPool("mana", file.Regen.OutOfCombat.Mana, next.oocMana)
	if err != nil {
		return err
	}
	next.oocHP = hp
	next.oocMana = mana
	return nil
}

func mergeRegenPool(name string, in regenPoolFile, base RegenPolicy) (RegenPolicy, error) {
	out := base
	if in.Enabled != nil {
		out.Enabled = *in.Enabled
	}
	if in.Percent != nil {
		if *in.Percent < 0 {
			return RegenPolicy{}, fmt.Errorf("regen out_of_combat.%s.percent must be >= 0", name)
		}
		out.Percent = *in.Percent
	}
	if in.Flat != nil {
		if *in.Flat < 0 {
			return RegenPolicy{}, fmt.Errorf("regen out_of_combat.%s.flat must be >= 0", name)
		}
		out.Flat = *in.Flat
	}
	if in.IntervalSeconds != nil {
		if *in.IntervalSeconds < 1 {
			return RegenPolicy{}, fmt.Errorf("regen out_of_combat.%s.interval_seconds must be >= 1", name)
		}
		out.IntervalSeconds = *in.IntervalSeconds
	}
	return out, nil
}
