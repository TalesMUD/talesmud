package balance

import "strings"

// BossMechanicsConfig is the telegraph window and enrage spike for boss fights.
// Tiers match EnemyTrait.Difficulty (boss, hard). A turn of telegraph spends
// the enemy action on a warning; the hit lands on a later action.
type BossMechanicsConfig struct {
	TelegraphTurns       int      `yaml:"telegraph_turns"`
	TelegraphMS          int      `yaml:"telegraph_ms"`
	TelegraphLabel       string   `yaml:"telegraph_label"`
	TelegraphTiers       []string `yaml:"telegraph_tiers"`
	EnrageAfterRounds    int      `yaml:"enrage_after_rounds"`
	EnrageBelowHP        float64  `yaml:"enrage_below_hp"`
	EnrageDamage         float64  `yaml:"enrage_damage"`
	EnrageTiers          []string `yaml:"enrage_tiers"`
	EnrageSkipsTelegraph bool     `yaml:"enrage_skips_telegraph"`
}

func defaultBossMechanics() BossMechanicsConfig {
	return BossMechanicsConfig{
		TelegraphTurns:       1,
		TelegraphMS:          1400,
		TelegraphLabel:       "Crushing Blow",
		TelegraphTiers:       []string{"boss", "hard"},
		EnrageAfterRounds:    16,
		EnrageBelowHP:        0.30,
		EnrageDamage:         1.20,
		EnrageTiers:          []string{"boss"},
		EnrageSkipsTelegraph: true,
	}
}

func bossMechanics() BossMechanicsConfig {
	cfg := GetConfig()
	if cfg == nil {
		return defaultBossMechanics()
	}
	m := cfg.BossMechanics
	if m.TelegraphLabel == "" && m.TelegraphTurns == 0 && len(m.TelegraphTiers) == 0 && m.EnrageDamage == 0 {
		return defaultBossMechanics()
	}
	if m.TelegraphLabel == "" {
		m.TelegraphLabel = defaultBossMechanics().TelegraphLabel
	}
	if m.TelegraphMS <= 0 {
		m.TelegraphMS = defaultBossMechanics().TelegraphMS
	}
	if len(m.TelegraphTiers) == 0 {
		m.TelegraphTiers = defaultBossMechanics().TelegraphTiers
	}
	if len(m.EnrageTiers) == 0 {
		m.EnrageTiers = defaultBossMechanics().EnrageTiers
	}
	if m.EnrageDamage <= 0 {
		m.EnrageDamage = defaultBossMechanics().EnrageDamage
	}
	return m
}

func tierListed(difficulty string, tiers []string) bool {
	d := strings.ToLower(strings.TrimSpace(difficulty))
	if d == "elite" {
		d = "hard"
	}
	for _, t := range tiers {
		tt := strings.ToLower(strings.TrimSpace(t))
		if tt == "elite" {
			tt = "hard"
		}
		if tt == d {
			return true
		}
	}
	return false
}

// BossTelegraphTurns is the wind-up length for this difficulty. 0 means strike now.
func BossTelegraphTurns(difficulty string, enraged bool) int {
	m := bossMechanics()
	if enraged && m.EnrageSkipsTelegraph {
		return 0
	}
	if !tierListed(difficulty, m.TelegraphTiers) || m.TelegraphTurns <= 0 {
		return 0
	}
	return m.TelegraphTurns
}

// BossTelegraphLabel is the ability name shown during the wind-up.
func BossTelegraphLabel() string {
	return bossMechanics().TelegraphLabel
}

// BossTelegraphMS is the on-screen pulse length for the warning.
func BossTelegraphMS() int {
	return bossMechanics().TelegraphMS
}

// ShouldEnrage reports whether this difficulty and state cross the enrage line.
func ShouldEnrage(difficulty string, round int, hp, maxHP int32, already bool) bool {
	if already || maxHP <= 0 {
		return false
	}
	m := bossMechanics()
	if !tierListed(difficulty, m.EnrageTiers) {
		return false
	}
	if m.EnrageAfterRounds > 0 && round >= m.EnrageAfterRounds {
		return true
	}
	if m.EnrageBelowHP > 0 && float64(hp)/float64(maxHP) <= m.EnrageBelowHP {
		return true
	}
	return false
}

// ScaleEnrageDamage multiplies a hit when the attacker is enraged.
// A multiplier of 1 leaves the number unchanged.
func ScaleEnrageDamage(damage int32) int32 {
	if damage <= 0 {
		return damage
	}
	mult := bossMechanics().EnrageDamage
	if mult == 1 || mult <= 0 {
		return damage
	}
	out := int32(float64(damage)*mult + 0.5)
	if out < 1 {
		return 1
	}
	return out
}
