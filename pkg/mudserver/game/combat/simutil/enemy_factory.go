package simutil

import (
	"github.com/talesmud/talesmud/pkg/entities"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
)

// EnemyConfig holds the definition for an enemy type
type EnemyConfig struct {
	Name        string
	Level       int32
	HP          int32
	AttackPower int32
	Defense     int32
	Difficulty  string
}

// AllEnemyConfigs returns all known enemy definitions from the game data
func AllEnemyConfigs() []EnemyConfig {
	return []EnemyConfig{
		{"Catacomb Rat", 1, 8, 1, 0, "trivial"},
		{"Sewer Rat", 2, 12, 3, 0, "easy"},
		{"Tunnel Mole", 2, 15, 3, 2, "easy"},
		{"Meadow Wolf", 2, 18, 4, 1, "normal"},
		{"Alley Thug", 2, 18, 4, 1, "normal"},
		{"Wild Boar", 2, 25, 5, 2, "normal"},
		{"Bandit", 3, 22, 5, 2, "normal"},
		{"Night Whisper", 4, 30, 7, 0, "normal"},
		{"Burrow Brute", 4, 55, 8, 3, "boss"},
		{"Thornback Bear", 5, 80, 11, 4, "hard"},
		{"Hollow Knight", 6, 140, 12, 4, "boss"},
	}
}

// EnemyConfigByName returns the enemy config for the given name, or nil
func EnemyConfigByName(name string) *EnemyConfig {
	for _, e := range AllEnemyConfigs() {
		if e.Name == name {
			return &e
		}
	}
	return nil
}

// EnemyConfigsByDifficulty returns enemy configs matching the given difficulty
func EnemyConfigsByDifficulty(difficulty string) []EnemyConfig {
	var result []EnemyConfig
	for _, e := range AllEnemyConfigs() {
		if e.Difficulty == difficulty {
			result = append(result, e)
		}
	}
	return result
}

// EnemyConfigsByLevel returns enemy configs at or below the given level
func EnemyConfigsByLevel(maxLevel int32) []EnemyConfig {
	var result []EnemyConfig
	for _, e := range AllEnemyConfigs() {
		if e.Level <= maxLevel {
			result = append(result, e)
		}
	}
	return result
}

// CreateEnemy creates an NPC enemy from an EnemyConfig
// Applies difficulty-based multipliers to base stats
func CreateEnemy(config EnemyConfig) *npc.NPC {
	// Apply difficulty multipliers to base stats
	finalHP, finalAttack, finalDefense := balance.ApplyMultipliers(
		config.HP,
		config.AttackPower,
		config.Defense,
		config.Difficulty,
	)

	return &npc.NPC{
		Entity:           entities.NewEntity(),
		Name:             config.Name,
		Level:            config.Level,
		MaxHitPoints:     finalHP,
		CurrentHitPoints: finalHP,
		EnemyTrait: &npc.EnemyTrait{
			AttackPower: finalAttack,
			Defense:     finalDefense,
			Difficulty:  config.Difficulty,
		},
	}
}

// CreateScaledEnemy creates an enemy with stats scaled to a specific level.
// Useful for testing arbitrary level matchups beyond the predefined enemies.
func CreateScaledEnemy(name string, level int32, difficulty string) *npc.NPC {
	// Scale stats based on level and difficulty
	var hpMult, atkMult, defMult float64
	switch difficulty {
	case "trivial":
		hpMult, atkMult, defMult = 6.0, 0.8, 0.0
	case "easy":
		hpMult, atkMult, defMult = 8.0, 1.2, 0.5
	case "normal":
		hpMult, atkMult, defMult = 10.0, 2.0, 0.8
	case "hard":
		hpMult, atkMult, defMult = 16.0, 2.2, 1.0
	case "boss":
		hpMult, atkMult, defMult = 25.0, 2.5, 1.2
	default:
		hpMult, atkMult, defMult = 10.0, 2.0, 0.8
	}

	hp := int32(float64(level) * hpMult)
	atk := int32(float64(level) * atkMult)
	def := int32(float64(level) * defMult)

	if hp < 1 {
		hp = 1
	}
	if atk < 1 {
		atk = 1
	}

	return &npc.NPC{
		Entity:           entities.NewEntity(),
		Name:             name,
		Level:            level,
		MaxHitPoints:     hp,
		CurrentHitPoints: hp,
		EnemyTrait: &npc.EnemyTrait{
			AttackPower: atk,
			Defense:     def,
			Difficulty:  difficulty,
		},
	}
}
