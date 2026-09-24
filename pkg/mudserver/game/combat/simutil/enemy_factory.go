package simutil

import (
	"github.com/talesmud/talesmud/pkg/entities"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
)

// EnemyConfig holds the definition for an enemy type (base stats from content YAML)
type EnemyConfig struct {
	Name        string
	Level       int32
	HP          int32
	AttackPower int32
	Defense     int32
	Difficulty  string
}

// AllEnemyConfigs returns known enemy definitions aligned with talesmud-rpg-1 content.
// Difficulties for Sewer Rat / Tunnel Mole use "easy" (intended trash tier) — content
// currently tags them "normal"; see docs/combat-battle-stage.md C6 content notes.
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
		{"Hollow Knight", 6, 150, 13, 6, "boss"}, // content name: The Hollow Knight
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
// Applies difficulty-based (and named) multipliers to base stats
func CreateEnemy(config EnemyConfig) *npc.NPC {
	finalHP, finalAttack, finalDefense := balance.ApplyEnemyMultipliers(
		config.HP,
		config.AttackPower,
		config.Defense,
		config.Difficulty,
		config.Name,
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
	if level < 1 {
		level = 1
	}
	// Flatter than pure level-multiples so a few levels of growth do not
	// double the body. level_gap in combat_balance.yaml carries most of the
	// gap feel; these numbers set the at-level tier (trash / elite / boss).
	var hp, atk, def int32
	switch difficulty {
	case "trivial":
		hp = 20 + 4*level
		atk = 2 + level/2
		def = 0
	case "easy":
		hp = 48 + 7*level
		atk = 8 + (level*3)/2
		def = level / 4
	case "normal":
		hp = 60 + 10*level
		atk = 8 + (level*5)/4
		def = 1 + level/3
	case "hard":
		hp = 100 + 15*level
		atk = 14 + (level*3)/2
		def = 3 + level/2
	case "boss":
		hp = 160 + 18*level
		atk = 18 + (level*7)/4
		def = 4 + (level*2)/3
	default:
		hp = 50 + 8*level
		atk = 6 + level
		def = 1 + level/3
	}

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
