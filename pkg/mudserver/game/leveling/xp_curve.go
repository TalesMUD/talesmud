package leveling

import "math"

// XPTable is precomputed XP requirements for levels 1-50
// Index represents the level, value is the cumulative XP needed to reach that level
var XPTable [51]int32

// init precomputes the XP table on package initialization
func init() {
	XPTable[0] = 0
	XPTable[1] = 0 // Level 1 is starting level, no XP needed

	for level := int32(2); level <= 50; level++ {
		XPTable[level] = CalculateXPRequired(level)
	}
}

// CalculateXPRequired computes the cumulative XP needed to reach a specific level
// Formula: 85 + (17 * level^1.10) per level increment
// This creates a gentle exponential curve where early levels are fast and later levels progressively harder
// Targets: L2≈150, L5≈661, L10≈1681, L20≈5478, L30≈11935, L40≈20905, L50≈35455
func CalculateXPRequired(level int32) int32 {
	if level <= 1 {
		return 0
	}

	// Cumulative XP calculation
	var total float64
	for l := int32(2); l <= level; l++ {
		// XP needed to go from level (l-1) to level l
		xpForLevel := 85.0 + (17.0 * math.Pow(float64(l), 1.10))
		total += xpForLevel
	}

	return int32(total)
}

// GetXPRequired returns the total XP needed to reach a specific level
// Uses precomputed table for efficiency
func GetXPRequired(level int32) int32 {
	if level < 0 {
		return 0
	}
	if level > 50 {
		level = 50
	}
	return XPTable[level]
}

// GetXPForNextLevel returns how much more XP is needed to reach the next level
// Returns 0 if already at max level (50)
func GetXPForNextLevel(currentLevel int32, currentXP int32) int32 {
	if currentLevel >= 50 {
		return 0 // Already at max level
	}

	nextLevelXP := GetXPRequired(currentLevel + 1)
	remaining := nextLevelXP - currentXP

	if remaining < 0 {
		return 0
	}

	return remaining
}

// CalculateEnemyXPReward returns the XP granted by an enemy based on its level
// Formula: 10 * enemy_level
// This creates natural progression where higher-level enemies grant more XP
func CalculateEnemyXPReward(enemyLevel int32) int64 {
	if enemyLevel < 1 {
		enemyLevel = 1
	}
	return int64(enemyLevel) * 10
}

// GetXPProgress returns the percentage progress to next level (0-100)
// Returns 100 if at max level
func GetXPProgress(currentLevel int32, currentXP int32) float64 {
	if currentLevel >= 50 {
		return 100.0
	}

	currentLevelXP := GetXPRequired(currentLevel)
	nextLevelXP := GetXPRequired(currentLevel + 1)

	if nextLevelXP <= currentLevelXP {
		return 100.0
	}

	progress := float64(currentXP-currentLevelXP) / float64(nextLevelXP-currentLevelXP) * 100.0

	if progress < 0 {
		return 0
	}
	if progress > 100 {
		return 100
	}

	return progress
}
