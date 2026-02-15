package balance

import (
	"os"
	"path/filepath"
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
}

var (
	config     *CombatBalanceConfig
	configOnce sync.Once
	configPath = "config/combat_balance.yaml"
)

// LoadConfig loads the combat balance configuration from disk
func LoadConfig() (*CombatBalanceConfig, error) {
	var err error
	configOnce.Do(func() {
		// Try to find the config file
		paths := []string{
			configPath,
			filepath.Join(".", configPath),
			filepath.Join("..", configPath),
			filepath.Join("../..", configPath),
		}

		var data []byte
		var found bool
		for _, path := range paths {
			data, err = os.ReadFile(path)
			if err == nil {
				found = true
				log.WithField("path", path).Debug("Loaded combat balance config")
				break
			}
		}

		if !found {
			log.Warn("Combat balance config not found, using default multipliers")
			config = getDefaultConfig()
			return
		}

		config = &CombatBalanceConfig{}
		if err = yaml.Unmarshal(data, config); err != nil {
			log.WithError(err).Error("Failed to parse combat balance config, using defaults")
			config = getDefaultConfig()
		}
	})

	return config, err
}

// getDefaultConfig returns sensible default multipliers if config file is missing
func getDefaultConfig() *CombatBalanceConfig {
	return &CombatBalanceConfig{
		DifficultyMultipliers: map[string]DifficultyMultipliers{
			"trivial": {HP: 3.75, Attack: 4.0, Defense: 1.0},
			"easy":    {HP: 2.9, Attack: 1.67, Defense: 1.0},
			"normal":  {HP: 2.67, Attack: 1.5, Defense: 2.0},
			"hard":    {HP: 1.25, Attack: 1.18, Defense: 1.25},
			"boss":    {HP: 2.0, Attack: 1.5, Defense: 1.67},
		},
	}
}

// GetConfig returns the loaded combat balance config (loads on first call)
func GetConfig() *CombatBalanceConfig {
	if config == nil {
		LoadConfig()
	}
	return config
}

// ApplyMultipliers applies difficulty-based multipliers to base enemy stats
func ApplyMultipliers(baseHP, baseAttack, baseDef int32, difficulty string) (hp, attack, def int32) {
	cfg := GetConfig()

	mult, ok := cfg.DifficultyMultipliers[difficulty]
	if !ok {
		// Unknown difficulty, return unmodified
		log.WithField("difficulty", difficulty).Warn("Unknown difficulty tier, no multipliers applied")
		return baseHP, baseAttack, baseDef
	}

	// Apply multipliers and round to nearest int32
	hp = int32(float64(baseHP) * mult.HP)
	attack = int32(float64(baseAttack) * mult.Attack)
	def = int32(float64(baseDef) * mult.Defense)

	// Ensure minimum values
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
