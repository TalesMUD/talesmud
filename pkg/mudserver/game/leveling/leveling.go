package leveling

import (
	"fmt"

	"github.com/talesmud/talesmud/pkg/entities/characters"
)

const (
	MaxLevel = 50 // Hard level cap
)

// LevelUpResult contains all changes from a level-up event
type LevelUpResult struct {
	OldLevel              int32
	NewLevel              int32
	LevelsGained          int
	HPGained              int32
	ManaGained            int32
	AttributeGains        map[string]int32 // attribute short name -> gain amount
	AttributePointsGained int32            // distributable points earned
	Message               string           // Formatted level-up message for player
}

// CheckLevelUp determines if a character should level up based on their current XP
// Returns the number of levels gained and the new level
// Supports gaining multiple levels at once (for massive XP grants)
func CheckLevelUp(char *characters.Character) (levelsGained int, newLevel int32) {
	// Already at max level
	if char.Level >= MaxLevel {
		return 0, MaxLevel
	}

	levelsGained = 0
	currentLevel := char.Level
	currentXP := char.XP

	// Check if character has enough XP to level up (possibly multiple times)
	for currentLevel < MaxLevel {
		xpNeeded := GetXPRequired(currentLevel + 1)
		if currentXP >= xpNeeded {
			levelsGained++
			currentLevel++
		} else {
			break
		}
	}

	return levelsGained, currentLevel
}

// ApplyLevelUp applies all stat changes for leveling up
// Updates character level, HP, and attributes
// Returns a LevelUpResult with details for messaging and events
func ApplyLevelUp(char *characters.Character, levelsGained int) *LevelUpResult {
	if levelsGained <= 0 {
		return nil
	}

	oldLevel := char.Level
	newLevel := oldLevel + int32(levelsGained)

	// Enforce level cap
	if newLevel > MaxLevel {
		newLevel = MaxLevel
		levelsGained = int(newLevel - oldLevel)
	}

	// Calculate HP gain
	hpGained := CalculateHPGain(char, levelsGained)

	// Calculate attribute gains
	attributeGains := CalculateAttributeGains(char, oldLevel, newLevel)

	// Apply level increase
	char.Level = newLevel

	// Apply HP increase
	char.MaxHitPoints += hpGained
	char.CurrentHitPoints = char.MaxHitPoints // Fully restore HP on level-up

	// Apply attribute gains
	for attrShortName, gainAmount := range attributeGains {
		applyAttributeGain(char, attrShortName, gainAmount)
	}

	// Award distributable attribute points
	pointsGained := int32(levelsGained) * int32(PointsPerLevel)
	char.UnspentAttributePoints += pointsGained

	// Recalculate and restore mana (after attributes have been updated)
	oldMaxMana := char.MaxMana
	newMaxMana := char.CalculateMaxMana()
	char.MaxMana = newMaxMana
	char.CurrentMana = newMaxMana // Fully restore mana on level-up
	manaGained := newMaxMana - oldMaxMana

	// Build result with formatted message
	result := &LevelUpResult{
		OldLevel:              oldLevel,
		NewLevel:              newLevel,
		LevelsGained:          levelsGained,
		HPGained:              hpGained,
		ManaGained:            manaGained,
		AttributeGains:        attributeGains,
		AttributePointsGained: pointsGained,
		Message:               formatLevelUpMessage(oldLevel, newLevel, levelsGained, hpGained, manaGained, attributeGains, pointsGained, char),
	}

	return result
}

// formatLevelUpMessage creates a formatted level-up notification message
func formatLevelUpMessage(oldLevel, newLevel int32, levelsGained int, hpGained int32, manaGained int32, attributeGains map[string]int32, pointsGained int32, char *characters.Character) string {
	// Header
	message := "╔══════════════════════════════════════════════════╗\n"
	message += "║                    LEVEL UP!                     ║\n"
	message += "╚══════════════════════════════════════════════════╝\n\n"

	// Level announcement
	if levelsGained == 1 {
		message += fmt.Sprintf("You have reached Level %d!\n\n", newLevel)
	} else {
		message += fmt.Sprintf("You have reached Level %d! (+%d levels)\n\n", newLevel, levelsGained)
	}

	// Stat increases
	message += "STAT INCREASES:\n"
	message += fmt.Sprintf("  + %d Max HP (now %d HP)\n", hpGained, char.MaxHitPoints)
	if manaGained > 0 {
		message += fmt.Sprintf("  + %d Max Mana (now %d Mana)\n", manaGained, char.MaxMana)
	}

	// Add attribute gains
	if len(attributeGains) > 0 {
		message += FormatAttributeGains(attributeGains, char)
	}

	// Distributable attribute points
	if pointsGained > 0 {
		message += fmt.Sprintf("  + %d Attribute Point(s) to distribute! (total unspent: %d)\n", pointsGained, char.UnspentAttributePoints)
		message += "  Use 'spend <attribute> [amount]' to allocate.\n"
	}

	// Flavor text
	message += "\n"
	if levelsGained == 1 {
		message += "You feel more powerful!\n"
	} else {
		message += "You feel significantly more powerful!\n"
	}

	message += "\n══════════════════════════════════════════════════"

	return message
}

// GetLevelProgress returns information about current level progress
func GetLevelProgress(char *characters.Character) (currentLevel int32, currentXP int32, xpForNext int32, progressPercent float64) {
	currentLevel = char.Level
	currentXP = char.XP
	xpForNext = GetXPForNextLevel(currentLevel, currentXP)
	progressPercent = GetXPProgress(currentLevel, currentXP)

	return
}
