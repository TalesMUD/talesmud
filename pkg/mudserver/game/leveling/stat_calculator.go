package leveling

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/characters"
)

const (
	// HP scaling constants
	BaseHPPerLevel = 5  // All classes gain at least 5 HP per level
	STAMultiplier  = 2  // Stamina modifier has 2x effect on HP

	// Attribute scaling milestones
	PrimaryAttrEvery   = 5  // Primary attribute increases every 5 levels
	SecondaryAttrEvery = 10 // Secondary attribute increases every 10 levels
)

// CalculateHPGain returns the total HP gained from leveling up
// Formula: HP_Gain_Per_Level = BASE_HP + (STA_Modifier * STA_MULTIPLIER)
func CalculateHPGain(char *characters.Character, levelsGained int) int32 {
	// Get Stamina attribute
	sta := getAttributeValue(char, "STA")

	// Calculate Stamina modifier: (STA - 10) / 2
	staModifier := (sta - 10) / 2

	// HP gained per level
	hpPerLevel := int32(BaseHPPerLevel + (staModifier * STAMultiplier))

	// Total HP gain
	totalHPGain := hpPerLevel * int32(levelsGained)

	// Ensure at least some HP is gained (minimum 1 per level even with very low STA)
	if totalHPGain < int32(levelsGained) {
		totalHPGain = int32(levelsGained)
	}

	return totalHPGain
}

// GetPrimaryAttribute returns the primary attribute short name for a class
func GetPrimaryAttribute(class characters.Class) string {
	switch strings.ToLower(class.ID) {
	case "warrior":
		return "STR"
	case "hunter", "ranger":
		return "DEX"
	case "rogue":
		return "DEX"
	case "wizard", "mage":
		return "INT"
	case "cleric":
		return "WIS"
	case "druid":
		return "INT"
	default:
		return "STR" // Default to STR if unknown class
	}
}

// GetSecondaryAttribute returns the secondary attribute short name for a class
func GetSecondaryAttribute(class characters.Class) string {
	switch strings.ToLower(class.ID) {
	case "warrior":
		return "STA"
	case "hunter", "ranger":
		return "STA"
	case "rogue":
		return "STR"
	case "wizard", "mage":
		return "WIS"
	case "cleric":
		return "INT"
	case "druid":
		return "WIS"
	default:
		return "STA" // Default to STA if unknown class
	}
}

// CalculateAttributeGains returns a map of attribute increases from oldLevel to newLevel
// Primary attribute: +1 every 5 levels (5, 10, 15, 20, 25, 30, 35, 40, 45, 50)
// Secondary attribute: +1 every 10 levels (10, 20, 30, 40, 50)
func CalculateAttributeGains(char *characters.Character, oldLevel, newLevel int32) map[string]int32 {
	gains := make(map[string]int32)

	primaryAttr := GetPrimaryAttribute(char.Class)
	secondaryAttr := GetSecondaryAttribute(char.Class)

	// Count primary attribute gains (every 5 levels)
	primaryGains := int32(0)
	for level := oldLevel + 1; level <= newLevel; level++ {
		if level%PrimaryAttrEvery == 0 {
			primaryGains++
		}
	}
	if primaryGains > 0 {
		gains[primaryAttr] = primaryGains
	}

	// Count secondary attribute gains (every 10 levels)
	secondaryGains := int32(0)
	for level := oldLevel + 1; level <= newLevel; level++ {
		if level%SecondaryAttrEvery == 0 {
			secondaryGains++
		}
	}
	if secondaryGains > 0 {
		gains[secondaryAttr] = secondaryGains
	}

	return gains
}

// getAttributeValue retrieves the value of an attribute by short name
func getAttributeValue(char *characters.Character, shortName string) int32 {
	for _, attr := range char.Attributes {
		if attr.Short == shortName {
			return attr.Value
		}
	}
	return 10 // Default value if attribute not found
}

// applyAttributeGain increases an attribute value by the specified amount
func applyAttributeGain(char *characters.Character, shortName string, gainAmount int32) {
	for i := range char.Attributes {
		if char.Attributes[i].Short == shortName {
			char.Attributes[i].Value += gainAmount
			return
		}
	}
}

// FormatAttributeGains creates a human-readable string of attribute gains
func FormatAttributeGains(gains map[string]int32, char *characters.Character) string {
	if len(gains) == 0 {
		return ""
	}

	result := ""
	for attrShort, gainAmount := range gains {
		// Get new value after gain
		newValue := getAttributeValue(char, attrShort)
		result += fmt.Sprintf("  + %d %s (now %d)\n", gainAmount, attrShort, newValue)
	}

	return result
}
