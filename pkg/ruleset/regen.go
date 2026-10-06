package ruleset

import "fmt"

// RegenPolicy is one passive regeneration pool.
// Percent is a percent of that pool's maximum per tick, not a fraction.
type RegenPolicy struct {
	Enabled         bool
	Percent         float64
	Flat            int32
	IntervalSeconds int
}

// Active reports whether this pool can grant points.
// A disabled pool, or a pool whose percent and flat are both zero, is inactive.
func (p RegenPolicy) Active() bool {
	return p.Enabled && (p.Percent > 0 || p.Flat > 0)
}

// RegenPair is the HP and mana policy for one character state.
type RegenPair struct {
	HP   RegenPolicy
	Mana RegenPolicy
}

// RegenProfile is out-of-combat, resting, and in-combat regeneration.
type RegenProfile struct {
	OutOfCombat RegenPair
	Resting     RegenPair
	InCombat    RegenPair
}

type regenPoolFile struct {
	Enabled         *bool    `yaml:"enabled"`
	Percent         *float64 `yaml:"percent"`
	Flat            *int32   `yaml:"flat"`
	IntervalSeconds *int     `yaml:"interval_seconds"`
}

type regenBlockFile struct {
	HP   regenPoolFile `yaml:"hp"`
	Mana regenPoolFile `yaml:"mana"`
}

type regenFile struct {
	OutOfCombat regenBlockFile `yaml:"out_of_combat"`
	Resting     regenBlockFile `yaml:"resting"`
	InCombat    regenBlockFile `yaml:"in_combat"`
}

func defaultRegen() RegenProfile {
	return RegenProfile{
		OutOfCombat: RegenPair{
			HP:   RegenPolicy{Enabled: true, Percent: 2, Flat: 0, IntervalSeconds: 10},
			Mana: RegenPolicy{Enabled: true, Percent: 5, Flat: 0, IntervalSeconds: 10},
		},
		Resting: RegenPair{
			HP:   RegenPolicy{Enabled: true, Percent: 10, Flat: 0, IntervalSeconds: 10},
			Mana: RegenPolicy{Enabled: true, Percent: 15, Flat: 0, IntervalSeconds: 10},
		},
		InCombat: RegenPair{
			HP:   RegenPolicy{Enabled: true, Percent: 0.5, Flat: 0, IntervalSeconds: 10},
			Mana: RegenPolicy{Enabled: true, Percent: 1, Flat: 0, IntervalSeconds: 10},
		},
	}
}

// Regen is the passive tick profile.
// Per-round combat mana is separate and is not read from this profile.
func Regen() RegenProfile {
	mu.RLock()
	defer mu.RUnlock()
	return current.regen
}

// OutOfCombatRegen is the passive HP and mana policy while not fighting and not resting.
func OutOfCombatRegen() (hp, mana RegenPolicy) {
	mu.RLock()
	defer mu.RUnlock()
	return current.regen.OutOfCombat.HP, current.regen.OutOfCombat.Mana
}

func applyRegen(next *state, file regenFile) error {
	profile, err := mergeRegenProfile(next.regen, file)
	if err != nil {
		return err
	}
	next.regen = profile
	return nil
}

func mergeRegenProfile(base RegenProfile, file regenFile) (RegenProfile, error) {
	var err error
	base.OutOfCombat, err = mergeRegenPair("out_of_combat", file.OutOfCombat, base.OutOfCombat)
	if err != nil {
		return RegenProfile{}, err
	}
	base.Resting, err = mergeRegenPair("resting", file.Resting, base.Resting)
	if err != nil {
		return RegenProfile{}, err
	}
	base.InCombat, err = mergeRegenPair("in_combat", file.InCombat, base.InCombat)
	if err != nil {
		return RegenProfile{}, err
	}
	return base, nil
}

func mergeRegenPair(block string, in regenBlockFile, base RegenPair) (RegenPair, error) {
	hp, err := mergeRegenPool(block, "hp", in.HP, base.HP)
	if err != nil {
		return RegenPair{}, err
	}
	mana, err := mergeRegenPool(block, "mana", in.Mana, base.Mana)
	if err != nil {
		return RegenPair{}, err
	}
	return RegenPair{HP: hp, Mana: mana}, nil
}

func mergeRegenPool(block, name string, in regenPoolFile, base RegenPolicy) (RegenPolicy, error) {
	out := base
	if in.Enabled != nil {
		out.Enabled = *in.Enabled
	}
	if in.Percent != nil {
		if *in.Percent < 0 {
			return RegenPolicy{}, fmt.Errorf("regen %s.%s.percent must be >= 0", block, name)
		}
		out.Percent = *in.Percent
	}
	if in.Flat != nil {
		if *in.Flat < 0 {
			return RegenPolicy{}, fmt.Errorf("regen %s.%s.flat must be >= 0", block, name)
		}
		out.Flat = *in.Flat
	}
	if in.IntervalSeconds != nil {
		if *in.IntervalSeconds < 1 {
			return RegenPolicy{}, fmt.Errorf("regen %s.%s.interval_seconds must be >= 1", block, name)
		}
		out.IntervalSeconds = *in.IntervalSeconds
	}
	return out, nil
}
