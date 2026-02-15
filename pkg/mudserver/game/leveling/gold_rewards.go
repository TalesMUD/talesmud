package leveling

import "math"

// Difficulty gold multipliers
const (
	DifficultyGoldTrivial = 0.5
	DifficultyGoldEasy    = 0.75
	DifficultyGoldNormal  = 1.0
	DifficultyGoldHard    = 1.5
	DifficultyGoldBoss    = 3.0
)

// CalculateEnemyGoldReward calculates a gold drop range for an enemy based on level and difficulty.
// Used as fallback when an enemy's EnemyTrait.GoldDrop range is not configured.
//
// Base formula (normal difficulty):
//
//	min = 1 + level/5
//	max = 1 + level/2 + level/5
//
// Example outputs (normal difficulty):
//
//	Level  1: 1-2 gold
//	Level  5: 2-4 gold
//	Level 10: 3-8 gold
//	Level 20: 5-15 gold
//
// Difficulty multipliers scale the range: trivial(0.5x), easy(0.75x), normal(1x), hard(1.5x), boss(3x)
func CalculateEnemyGoldReward(enemyLevel int32, difficulty string) (minGold int32, maxGold int32) {
	if enemyLevel < 1 {
		enemyLevel = 1
	}

	// Base gold range
	baseMin := int32(1) + enemyLevel/5
	baseMax := int32(1) + enemyLevel/2 + enemyLevel/5

	// Ensure at least 1 gold spread
	if baseMax <= baseMin {
		baseMax = baseMin + 1
	}

	// Apply difficulty multiplier
	mult := GetDifficultyGoldMultiplier(difficulty)
	minGold = int32(math.Max(1, math.Round(float64(baseMin)*mult)))
	maxGold = int32(math.Max(float64(minGold+1), math.Round(float64(baseMax)*mult)))

	return minGold, maxGold
}

// RollEnemyGold calculates and rolls a random gold amount for an enemy kill.
// Convenience function that combines CalculateEnemyGoldReward with a random roll.
func RollEnemyGold(enemyLevel int32, difficulty string, randIntn func(int) int) int64 {
	minGold, maxGold := CalculateEnemyGoldReward(enemyLevel, difficulty)
	spread := maxGold - minGold
	if spread <= 0 {
		return int64(minGold)
	}
	return int64(minGold) + int64(randIntn(int(spread+1)))
}

// GetDifficultyGoldMultiplier returns the gold multiplier for a given difficulty level.
func GetDifficultyGoldMultiplier(difficulty string) float64 {
	switch difficulty {
	case "trivial":
		return DifficultyGoldTrivial
	case "easy":
		return DifficultyGoldEasy
	case "normal":
		return DifficultyGoldNormal
	case "hard":
		return DifficultyGoldHard
	case "boss":
		return DifficultyGoldBoss
	default:
		return DifficultyGoldNormal
	}
}
