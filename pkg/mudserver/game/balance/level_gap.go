package balance

import "math"

// LevelGapPerLevel is the change applied for each level of attacker advantage
// (attackerLevel - defenderLevel), before clamping.
type LevelGapPerLevel struct {
	HitChance   float64 `yaml:"hit_chance"`   // Added to hit probability. +0.05 ≈ +1 on a d20.
	CritChance  float64 `yaml:"crit_chance"`  // Added to the base crit chance (natural 20 = 5%).
	DamageDealt float64 `yaml:"damage_dealt"` // Multiplier term on outgoing damage.
	DamageTaken float64 `yaml:"damage_taken"` // Multiplier term on damage the defender receives.
}

// LevelGapConfig is the `level_gap` section of config/combat_balance.yaml.
// World-specific names do not belong here; tune the numbers, not the content.
type LevelGapConfig struct {
	MaxLevels           int              `yaml:"max_levels"`
	PerLevel            LevelGapPerLevel `yaml:"per_level"`
	MinDamageMultiplier float64          `yaml:"min_damage_multiplier"`
	MaxDamageMultiplier float64          `yaml:"max_damage_multiplier"`
}

// LevelGapMods is the resolved modifier for one attacker/defender pair.
// Gap is attackerLevel - defenderLevel after clamping. Positive means the
// attacker is higher level than the defender.
type LevelGapMods struct {
	Gap              int
	RawGap           int
	HitChanceDelta   float64
	HitBonus         int // d20 to-hit bonus (HitChanceDelta / 0.05, rounded)
	CritChanceDelta  float64
	DamageDealtMult  float64
	DamageTakenMult  float64
	DamageMultiplier float64 // Dealt × taken, clamped
}

func defaultLevelGap() LevelGapConfig {
	return LevelGapConfig{
		MaxLevels: 6,
		PerLevel: LevelGapPerLevel{
			HitChance:   0.05,
			CritChance:  0.015,
			DamageDealt: 0.06,
			DamageTaken: 0.04,
		},
		MinDamageMultiplier: 0.40,
		MaxDamageMultiplier: 1.80,
	}
}

func effectiveLevelGap(cfg *CombatBalanceConfig) LevelGapConfig {
	def := defaultLevelGap()
	if cfg == nil || cfg.LevelGap.MaxLevels <= 0 {
		return def
	}
	gap := cfg.LevelGap
	if gap.MinDamageMultiplier <= 0 {
		gap.MinDamageMultiplier = def.MinDamageMultiplier
	}
	if gap.MaxDamageMultiplier <= 0 {
		gap.MaxDamageMultiplier = def.MaxDamageMultiplier
	}
	if gap.MaxDamageMultiplier < gap.MinDamageMultiplier {
		gap.MaxDamageMultiplier = gap.MinDamageMultiplier
	}
	return gap
}

// LevelGapModifiers returns hit, crit, and damage modifiers for an attack.
// A zero gap (equal level, or both levels unset) is neutral.
func LevelGapModifiers(attackerLevel, defenderLevel int32) LevelGapMods {
	return levelGapModifiers(GetConfig(), attackerLevel, defenderLevel)
}

func levelGapModifiers(cfg *CombatBalanceConfig, attackerLevel, defenderLevel int32) LevelGapMods {
	gapCfg := effectiveLevelGap(cfg)
	raw := int(int64(attackerLevel) - int64(defenderLevel))
	gap := raw
	max := gapCfg.MaxLevels
	if gap > max {
		gap = max
	}
	if gap < -max {
		gap = -max
	}

	per := gapCfg.PerLevel
	hitDelta := float64(gap) * per.HitChance
	hitBonus := int(math.Round(hitDelta / 0.05))

	dealt := 1 + float64(gap)*per.DamageDealt
	taken := 1 + float64(gap)*per.DamageTaken
	if dealt < 0 {
		dealt = 0
	}
	if taken < 0 {
		taken = 0
	}
	mult := dealt * taken
	if mult < gapCfg.MinDamageMultiplier {
		mult = gapCfg.MinDamageMultiplier
	}
	if mult > gapCfg.MaxDamageMultiplier {
		mult = gapCfg.MaxDamageMultiplier
	}

	return LevelGapMods{
		Gap:              gap,
		RawGap:           raw,
		HitChanceDelta:   hitDelta,
		HitBonus:         hitBonus,
		CritChanceDelta:  float64(gap) * per.CritChance,
		DamageDealtMult:  dealt,
		DamageTakenMult:  taken,
		DamageMultiplier: mult,
	}
}

// ScaleDamage applies the level-gap damage multiplier (dealt × taken).
// Non-positive damage is unchanged. A neutral gap returns damage as-is.
// Positive results are at least 1.
func ScaleDamage(attackerLevel, defenderLevel, damage int32) int32 {
	if damage <= 0 {
		return damage
	}
	mult := LevelGapModifiers(attackerLevel, defenderLevel).DamageMultiplier
	if mult == 1 {
		return damage
	}
	scaled := int32(math.Round(float64(damage) * mult))
	if scaled < 1 {
		scaled = 1
	}
	return scaled
}
