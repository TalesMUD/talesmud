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
// Uses a piecewise formula: early levels are very flat, transitioning to steeper scaling at higher levels
// Targets: L2≈45, L5≈203, L10≈1120, L20≈4947, L30≈11794, L40≈21271, L50≈33454
func CalculateXPRequired(level int32) int32 {
	if level <= 1 {
		return 0
	}

	// Cumulative XP calculation with piecewise scaling
	var total float64
	for l := int32(2); l <= level; l++ {
		var xpForLevel float64
		if l <= 5 {
			// Early game: much flatter curve for fast initial leveling
			xpForLevel = 30.0 + (10.0 * math.Pow(float64(l), 0.60))
		} else if l <= 15 {
			// Mid-early: transition zone with moderate scaling
			xpForLevel = 50.0 + (15.0 * math.Pow(float64(l), 1.05))
		} else {
			// Mid-to-late: original steeper scaling
			xpForLevel = 85.0 + (17.0 * math.Pow(float64(l), 1.10))
		}
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
// Formula: 15 * enemy_level + 5
// This creates natural progression where higher-level enemies grant more XP
// Note: Only used as fallback when an enemy's EnemyTrait.XPReward is 0
func CalculateEnemyXPReward(enemyLevel int32) int64 {
	if enemyLevel < 1 {
		enemyLevel = 1
	}
	return int64(enemyLevel)*15 + 5
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
