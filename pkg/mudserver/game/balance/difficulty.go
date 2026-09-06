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
			"The Hollow Knight": {HP: 1.0, Attack: 0.85, Defense: 0.9},
			"Hollow Knight":     {HP: 1.0, Attack: 0.85, Defense: 0.9},
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
func ApplyEnemyMultipliers(baseHP, baseAttack, baseDef int32, difficulty, enemyName string) (hp, attack, def int32) {
	cfg := GetConfig()

	mult, ok := lookupMultipliers(cfg, difficulty, enemyName)
	if !ok {
		log.WithField("difficulty", difficulty).Warn("Unknown difficulty tier, no multipliers applied")
		return baseHP, baseAttack, baseDef
	}

	hp = int32(float64(baseHP) * mult.HP)
	attack = int32(float64(baseAttack) * mult.Attack)
	def = int32(float64(baseDef) * mult.Defense)

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

func lookupMultipliers(cfg *CombatBalanceConfig, difficulty, enemyName string) (DifficultyMultipliers, bool) {
	if cfg != nil && enemyName != "" && cfg.NamedOverrides != nil {
		if m, ok := cfg.NamedOverrides[enemyName]; ok {
			return m, true
		}
		// Case-insensitive fallback
		lower := strings.ToLower(enemyName)
		for name, m := range cfg.NamedOverrides {
			if strings.ToLower(name) == lower {
				return m, true
			}
		}
	}

	if cfg == nil || cfg.DifficultyMultipliers == nil {
		return DifficultyMultipliers{}, false
	}
	m, ok := cfg.DifficultyMultipliers[difficulty]
	return m, ok
}
