package balance

import "math"

// RewardScaleConfig is the `reward_scale` section of config/combat_balance.yaml.
// Multipliers apply to an enemy's base XP and gold by threat tier
// (enemyLevel - referenceLevel). referenceLevel is chosen by the victory
// code: the highest level among characters who receive the split.
type RewardScaleConfig struct {
	Grey           float64 `yaml:"grey"`
	Green          float64 `yaml:"green"`
	Yellow         float64 `yaml:"yellow"`
	Orange         float64 `yaml:"orange"`
	Red            float64 `yaml:"red"`
	Skull          float64 `yaml:"skull"`
	FirstKillBonus float64 `yaml:"first_kill_bonus"` // fraction of that character's boss share
}

func defaultRewardScale() RewardScaleConfig {
	return RewardScaleConfig{
		Grey:           0.15,
		Green:          0.60,
		Yellow:         1.0,
		Orange:         1.25,
		Red:            1.50,
		Skull:          2.0,
		FirstKillBonus: 0.50,
	}
}

func effectiveRewardScale(cfg *CombatBalanceConfig) RewardScaleConfig {
	if cfg == nil || cfg.RewardScale == (RewardScaleConfig{}) {
		return defaultRewardScale()
	}
	return cfg.RewardScale
}

// RewardMultiplier returns the XP/gold multiplier for a threat tier.
func RewardMultiplier(tier string) float64 {
	return rewardMultiplier(GetConfig(), tier)
}

func rewardMultiplier(cfg *CombatBalanceConfig, tier string) float64 {
	s := effectiveRewardScale(cfg)
	switch tier {
	case ThreatGrey:
		return s.Grey
	case ThreatGreen:
		return s.Green
	case ThreatYellow:
		return s.Yellow
	case ThreatOrange:
		return s.Orange
	case ThreatRed:
		return s.Red
	case ThreatSkull:
		return s.Skull
	default:
		if s.Yellow > 0 {
			return s.Yellow
		}
		return 1
	}
}

// FirstKillBonusRate is the fraction added to a character's own share of a
// boss they have not killed before.
func FirstKillBonusRate() float64 {
	return effectiveRewardScale(GetConfig()).FirstKillBonus
}

// ScaleReward multiplies an XP or gold amount and rounds.
// A positive amount with a positive multiplier is at least 1.
// A multiplier of 1 returns amount unchanged.
func ScaleReward(amount int64, mult float64) int64 {
	if amount <= 0 || mult <= 0 {
		return 0
	}
	if mult == 1 {
		return amount
	}
	scaled := int64(math.Round(float64(amount) * mult))
	if scaled < 1 {
		return 1
	}
	return scaled
}

// BonusReward is the first-kill extra: round(share * rate). Zero rate or
// zero share yields 0 (no forced minimum).
func BonusReward(share int64, rate float64) int64 {
	if share <= 0 || rate <= 0 {
		return 0
	}
	return int64(math.Round(float64(share) * rate))
}
