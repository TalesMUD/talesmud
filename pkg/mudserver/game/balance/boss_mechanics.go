package balance

import (
	"fmt"
	"math"
	"strings"
)

// BossMechanicsConfig is the telegraph window and enrage spike for boss fights.
// Tiers match EnemyTrait.Difficulty (boss, hard). A turn of telegraph spends
// the enemy action on a warning; the hit lands on a later action.
// BossPhaseConfig describes a descending HP band. Optional overrides compose
// with the global telegraph/enrage rules; zero damage means the inherited value.
type BossPhaseConfig struct {
	Label          string  `yaml:"label"`
	BelowHP        float64 `yaml:"below_hp"`
	TelegraphLabel string  `yaml:"telegraph_label"`
	DamageDealt    float64 `yaml:"damage_dealt"`
	EnrageDamage   float64 `yaml:"enrage_damage"`
}

type BossMechanicsConfig struct {
	PhaseTiers           []string          `yaml:"phase_tiers"`
	Phases               []BossPhaseConfig `yaml:"phases"`
	TelegraphTurns       int               `yaml:"telegraph_turns"`
	TelegraphMS          int               `yaml:"telegraph_ms"`
	TelegraphLabel       string            `yaml:"telegraph_label"`
	TelegraphTiers       []string          `yaml:"telegraph_tiers"`
	EnrageAfterRounds    int               `yaml:"enrage_after_rounds"`
	EnrageBelowHP        float64           `yaml:"enrage_below_hp"`
	EnrageDamage         float64           `yaml:"enrage_damage"`
	EnrageTiers          []string          `yaml:"enrage_tiers"`
	EnrageSkipsTelegraph bool              `yaml:"enrage_skips_telegraph"`
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
	if m.TelegraphLabel == "" && m.TelegraphTurns == 0 && len(m.TelegraphTiers) == 0 && m.EnrageDamage == 0 && len(m.Phases) == 0 {
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

// ValidateBossPhases rejects ambiguous ordering and invalid balance values.
func ValidateBossPhases(m BossMechanicsConfig) error {
	previous := 2.0
	for i, p := range m.Phases {
		if strings.TrimSpace(p.Label) == "" || math.IsNaN(p.BelowHP) || math.IsInf(p.BelowHP, 0) || p.BelowHP <= 0 || p.BelowHP > 1 || p.BelowHP >= previous || (i == 0 && p.BelowHP != 1) {
			return fmt.Errorf("boss phase %d requires a label and descending below_hp (opening must be 1)", i+1)
		}
		for _, mult := range []float64{p.DamageDealt, p.EnrageDamage} {
			if math.IsNaN(mult) || math.IsInf(mult, 0) || mult < 0 {
				return fmt.Errorf("boss phase %d has invalid damage multiplier", i+1)
			}
		}
		previous = p.BelowHP
	}
	return nil
}

// BossPhases defaults to bosses only. An omitted phase list disables phases.
func BossPhases(difficulty string) []BossPhaseConfig {
	m := bossMechanics()
	tiers := m.PhaseTiers
	if len(tiers) == 0 {
		tiers = []string{"boss"}
	}
	if !tierListed(difficulty, tiers) {
		return nil
	}
	return m.Phases
}

// BossPhase uses a one-based phase number; zero means no phase mechanics.
func BossPhase(difficulty string, phase int) BossPhaseConfig {
	phases := BossPhases(difficulty)
	if phase <= 0 || phase > len(phases) {
		return BossPhaseConfig{}
	}
	return phases[phase-1]
}

func BossPhaseTelegraphLabel(difficulty string, phase int) string {
	if label := BossPhase(difficulty, phase).TelegraphLabel; label != "" {
		return label
	}
	return BossTelegraphLabel()
}

// ScaleBossPhaseDamage applies a phase multiplier and an optional enrage override.
// The enrage multiplier replaces the global value, rather than multiplying it twice.
func ScaleBossPhaseDamage(damage int32, difficulty string, phase int, enraged bool) int32 {
	if damage <= 0 {
		return damage
	}
	p := BossPhase(difficulty, phase)
	mult := p.DamageDealt
	if mult <= 0 {
		mult = 1
	}
	if enraged {
		rage := p.EnrageDamage
		if rage <= 0 {
			rage = bossMechanics().EnrageDamage
		}
		mult *= rage
	}
	return int32(math.Max(1, math.Round(float64(damage)*mult)))
}
