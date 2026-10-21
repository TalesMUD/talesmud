package balance

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// DifficultyMultipliers holds stat multipliers for a difficulty tier
type DifficultyMultipliers struct {
	HP      float64 `yaml:"hp"`
	Attack  float64 `yaml:"attack"`
	Defense float64 `yaml:"defense"`
}

// CombatBalanceConfig holds the full combat balance configuration
type CombatBalanceConfig struct {
	DifficultyMultipliers map[string]DifficultyMultipliers `yaml:"difficulty_multipliers"`
	NamedOverrides        map[string]DifficultyMultipliers `yaml:"named_overrides"`
	LevelGap              LevelGapConfig                   `yaml:"level_gap"`
	Threat                ThreatConfig                     `yaml:"threat"`
	RewardScale           RewardScaleConfig                `yaml:"reward_scale"`
	ClassBalance          map[string]ClassBalance          `yaml:"class_balance"`
	BossMechanics         BossMechanicsConfig              `yaml:"boss_mechanics"`
	EarlyNormal           EarlyNormalConfig                `yaml:"early_normal"`
	NewbieCurve           NewbieCurveConfig                `yaml:"newbie_curve"`
}

// EarlyNormalConfig scales the normal-tier multipliers for low-level enemies.
// Other tiers, and normal enemies above MaxLevel, keep the tier table.
// Factors are applied before the integer truncate. 0 means 1.
type EarlyNormalConfig struct {
	MaxLevel int32   `yaml:"max_level"`
	HP       float64 `yaml:"hp"`
	Attack   float64 `yaml:"attack"`
	Defense  float64 `yaml:"defense"`
}

// NewbieClassFactor is an extra damage factor while that class is still in the newbie band.
// UphillTaken, when set, replaces the curve-wide fallback for a +2 attacker.
type NewbieClassFactor struct {
	DamageDealt float64 `yaml:"damage_dealt"`
	DamageTaken float64 `yaml:"damage_taken"`
	UphillTaken float64 `yaml:"uphill_taken"`
}

// NewbieCurveConfig is applied in ScaleClassDamage only when that combatant's
// level is in 1..MaxPlayerLevel. The level-10 gap table does not see it.
// UphillTaken multiplies damage taken when the attacker is at least UphillGap
// levels higher and not above UphillMaxAttackerLevel. Plate and cloth do not
// share one factor: leather stops most early swings, so the hit that lands
// has to matter, while a cloth robe is already being hit.
type NewbieCurveConfig struct {
	MaxPlayerLevel         int32                        `yaml:"max_player_level"`
	UphillGap              int32                        `yaml:"uphill_gap"`
	UphillTaken            float64                      `yaml:"uphill_taken"`
	UphillMaxAttackerLevel int32                        `yaml:"uphill_max_attacker_level"`
	Classes                map[string]NewbieClassFactor `yaml:"classes"`
}

var (
	config     *CombatBalanceConfig
	configOnce sync.Once
	configMu   sync.RWMutex
	configPath = "config/combat_balance.yaml"
)

// LoadConfig loads the combat balance configuration from disk
func LoadConfig() (*CombatBalanceConfig, error) {
	var err error
	configOnce.Do(func() {
		config, err = loadConfigFromDisk()
	})
	return config, err
}

func loadConfigFromDisk() (*CombatBalanceConfig, error) {
	paths := candidateConfigPaths()

	var data []byte
	var foundPath string
	var err error
	for _, path := range paths {
		data, err = os.ReadFile(path)
		if err == nil {
			foundPath = path
			break
		}
	}

	if foundPath == "" {
		log.Warn("Combat balance config not found, using default multipliers")
		return getDefaultConfig(), nil
	}

	log.WithField("path", foundPath).Debug("Loaded combat balance config")
	cfg := &CombatBalanceConfig{}
	if err = yaml.Unmarshal(data, cfg); err != nil {
		log.WithError(err).Error("Failed to parse combat balance config, using defaults")
		return getDefaultConfig(), err
	}
	if err := ValidateBossPhases(cfg.BossMechanics); err != nil {
		return getDefaultConfig(), err
	}
	if cfg.DifficultyMultipliers == nil {
		cfg.DifficultyMultipliers = getDefaultConfig().DifficultyMultipliers
	}
	return cfg, nil
}

// candidateConfigPaths returns likely locations for combat_balance.yaml
// (repo root, cwd, and parents — tests often run with package cwd).
func candidateConfigPaths() []string {
	seen := map[string]bool{}
	add := func(paths *[]string, p string) {
		if p == "" || seen[p] {
			return
		}
		seen[p] = true
		*paths = append(*paths, p)
	}

	var paths []string
	add(&paths, configPath)
	add(&paths, filepath.Join(".", configPath))

	// Walk up from cwd looking for config/combat_balance.yaml
	if cwd, err := os.Getwd(); err == nil {
		dir := cwd
		for i := 0; i < 8; i++ {
			add(&paths, filepath.Join(dir, configPath))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	return paths
}

// ReloadConfig forces a reload from disk (for tests / hot-tune). Resets sync.Once.
func ReloadConfig() (*CombatBalanceConfig, error) {
	configMu.Lock()
	defer configMu.Unlock()
	configOnce = sync.Once{}
	cfg, err := loadConfigFromDisk()
	config = cfg
	// Mark Once as done so GetConfig doesn't reload again until next ReloadConfig
	configOnce.Do(func() {})
	return config, err
}

// getDefaultConfig returns sensible default multipliers if config file is missing
func getDefaultConfig() *CombatBalanceConfig {
	return &CombatBalanceConfig{
		DifficultyMultipliers: map[string]DifficultyMultipliers{
			"trivial": {HP: 2.5, Attack: 3.0, Defense: 1.0},
			"easy":    {HP: 1.9, Attack: 1.4, Defense: 0.5},
			"normal":  {HP: 2.7, Attack: 1.35, Defense: 1.5},
			"hard":    {HP: 0.9, Attack: 1.0, Defense: 1.0},
			"boss":    {HP: 1.85, Attack: 1.15, Defense: 1.5},
		},
		NamedOverrides: map[string]DifficultyMultipliers{
			"Sample Warden": {HP: 1.0, Attack: 0.85, Defense: 0.9},
		},
		LevelGap:      defaultLevelGap(),
		Threat:        defaultThreat(),
		RewardScale:   defaultRewardScale(),
		ClassBalance:  defaultClassBalance(),
		BossMechanics: defaultBossMechanics(),
		EarlyNormal: EarlyNormalConfig{
			MaxLevel: 6,
			HP:       0.42,
			Attack:   1.45,
			Defense:  0.80,
		},
		NewbieCurve: NewbieCurveConfig{
			MaxPlayerLevel:         5,
			UphillGap:              2,
			UphillTaken:            2.2,
			UphillMaxAttackerLevel: 6,
			Classes: map[string]NewbieClassFactor{
				"warrior": {UphillTaken: 2.2},
				"rogue":   {UphillTaken: 1.5},
				"ward":    {UphillTaken: 1.5},
				"mage":    {DamageDealt: 1.15, DamageTaken: 0.32, UphillTaken: 1.0},
				"rigger":  {DamageDealt: 1.30, DamageTaken: 0.95, UphillTaken: 1.0},
			},
		},
	}
}

// GetConfig returns the loaded combat balance config (loads on first call)
func GetConfig() *CombatBalanceConfig {
	configMu.RLock()
	cfg := config
	configMu.RUnlock()
	if cfg == nil {
		LoadConfig()
		configMu.RLock()
		cfg = config
		configMu.RUnlock()
	}
	return cfg
}

// ApplyMultipliers applies difficulty-based multipliers to base enemy stats.
// Prefer ApplyEnemyMultipliers when the enemy name is known (named overrides).
func ApplyMultipliers(baseHP, baseAttack, baseDef int32, difficulty string) (hp, attack, def int32) {
	return ApplyEnemyMultipliers(baseHP, baseAttack, baseDef, difficulty, "")
}

// ApplyEnemyMultipliers applies difficulty multipliers, with optional named overrides
// that replace the tier multipliers when enemyName matches (case-insensitive).
// Level is unknown here, so the early-normal curve is not applied.
func ApplyEnemyMultipliers(baseHP, baseAttack, baseDef int32, difficulty, enemyName string) (hp, attack, def int32) {
	return applyEnemy(baseHP, baseAttack, baseDef, difficulty, enemyName, 0)
}

// ApplyEnemyCombat is ApplyEnemyMultipliers plus the early-normal curve when
// level is in 1..early_normal.max_level and the tier is normal.
func ApplyEnemyCombat(baseHP, baseAttack, baseDef int32, difficulty, enemyName string, level int32) (hp, attack, def int32) {
	return applyEnemy(baseHP, baseAttack, baseDef, difficulty, enemyName, level)
}

func applyEnemy(baseHP, baseAttack, baseDef int32, difficulty, enemyName string, level int32) (hp, attack, def int32) {
	cfg := GetConfig()

	mult, ok := lookupMultipliers(cfg, difficulty, enemyName)
	if !ok {
		log.WithField("difficulty", difficulty).Warn("Unknown difficulty tier, no multipliers applied")
		return baseHP, baseAttack, baseDef
	}

	hpM, atkM, defM := mult.HP, mult.Attack, mult.Defense
	if earlyNormalApplies(cfg, difficulty, level) {
		hpM *= factorOrOne(cfg.EarlyNormal.HP)
		atkM *= factorOrOne(cfg.EarlyNormal.Attack)
		defM *= factorOrOne(cfg.EarlyNormal.Defense)
	}

	hp = int32(float64(baseHP) * hpM)
	attack = int32(float64(baseAttack) * atkM)
	def = int32(float64(baseDef) * defM)

	if hp < 1 {
		hp = 1
	}
	if attack < 1 {
		attack = 1
	}
	if def < 0 {
		def = 0
	}

	return hp, attack, def
}

func earlyNormalApplies(cfg *CombatBalanceConfig, difficulty string, level int32) bool {
	if cfg == nil || level <= 0 || cfg.EarlyNormal.MaxLevel <= 0 || level > cfg.EarlyNormal.MaxLevel {
		return false
	}
	return strings.ToLower(strings.TrimSpace(difficulty)) == "normal"
}

func factorOrOne(v float64) float64 {
	if v <= 0 {
		return 1
	}
	return v
}

func lookupMultipliers(cfg *CombatBalanceConfig, difficulty, enemyName string) (DifficultyMultipliers, bool) {
	m, _, ok := factors(cfg, difficulty, enemyName)
	return m, ok
}

// EnemyFactors reports the multipliers ApplyEnemyMultipliers uses.
// Source is "named" or "tier". ok is false when the name has no override and the tier is not in the table.
func EnemyFactors(difficulty, enemyName string) (DifficultyMultipliers, string, bool) {
	return factors(GetConfig(), difficulty, enemyName)
}

func factors(cfg *CombatBalanceConfig, difficulty, enemyName string) (DifficultyMultipliers, string, bool) {
	if cfg != nil && enemyName != "" && cfg.NamedOverrides != nil {
		if m, ok := cfg.NamedOverrides[enemyName]; ok {
			return m, "named", true
		}
		lower := strings.ToLower(enemyName)
		for name, m := range cfg.NamedOverrides {
			if strings.ToLower(name) == lower {
				return m, "named", true
			}
		}
	}
	if cfg == nil || cfg.DifficultyMultipliers == nil {
		return DifficultyMultipliers{}, "", false
	}
	m, ok := cfg.DifficultyMultipliers[difficulty]
	if !ok {
		return DifficultyMultipliers{}, "", false
	}
	return m, "tier", true
}
