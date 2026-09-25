package balance

// Threat tier names sent to clients. Compared as enemyLevel - playerLevel.
const (
	ThreatGrey   = "grey"
	ThreatGreen  = "green"
	ThreatYellow = "yellow"
	ThreatOrange = "orange"
	ThreatRed    = "red"
	ThreatSkull  = "skull"
)

// ThreatConfig is the `threat` section of config/combat_balance.yaml.
// Cutoffs are on (enemyLevel - playerLevel) and must ascend:
// grey, then green, yellow, orange, red. Anything above red is skull.
type ThreatConfig struct {
	GreyAtOrBelow   int `yaml:"grey_at_or_below"`
	GreenAtOrBelow  int `yaml:"green_at_or_below"`
	YellowAtOrBelow int `yaml:"yellow_at_or_below"`
	OrangeAtOrBelow int `yaml:"orange_at_or_below"`
	RedAtOrBelow    int `yaml:"red_at_or_below"`
}

func defaultThreat() ThreatConfig {
	return ThreatConfig{
		GreyAtOrBelow:   -3,
		GreenAtOrBelow:  -1,
		YellowAtOrBelow: 1,
		OrangeAtOrBelow: 2,
		RedAtOrBelow:    4,
	}
}

func effectiveThreat(cfg *CombatBalanceConfig) ThreatConfig {
	if cfg == nil || cfg.Threat == (ThreatConfig{}) {
		return defaultThreat()
	}
	return cfg.Threat
}

// ThreatTier returns the con color for an enemy relative to a player.
// Gap 0 (and +1) is yellow. -3 or lower is grey. +5 or higher is skull.
func ThreatTier(playerLevel, enemyLevel int32) string {
	return threatTier(GetConfig(), playerLevel, enemyLevel)
}

func threatTier(cfg *CombatBalanceConfig, playerLevel, enemyLevel int32) string {
	// Characters start at level 1. An unset 0 would paint every level-2
	// enemy orange, so floor the viewer at 1.
	if playerLevel < 1 {
		playerLevel = 1
	}
	gap := int(int64(enemyLevel) - int64(playerLevel))
	t := effectiveThreat(cfg)
	switch {
	case gap <= t.GreyAtOrBelow:
		return ThreatGrey
	case gap <= t.GreenAtOrBelow:
		return ThreatGreen
	case gap <= t.YellowAtOrBelow:
		return ThreatYellow
	case gap <= t.OrangeAtOrBelow:
		return ThreatOrange
	case gap <= t.RedAtOrBelow:
		return ThreatRed
	default:
		return ThreatSkull
	}
}

// ThreatNeedsWarning is true for orange, red, and skull.
func ThreatNeedsWarning(tier string) bool {
	switch tier {
	case ThreatOrange, ThreatRed, ThreatSkull:
		return true
	default:
		return false
	}
}
