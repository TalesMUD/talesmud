package leveling

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/characters"
)

const (
	// PointsPerLevel is the number of distributable attribute points earned per level-up
	PointsPerLevel = 2
)

// AttributeCaps defines the maximum distributable points a class can invest in each attribute.
// These cap SPENT distributable points only — automatic milestone gains are separate and uncapped.
type AttributeCaps map[string]int32

// ClassAttributeCaps returns the distributable point caps for a given class ID.
func ClassAttributeCaps(classID string) AttributeCaps {
	caps, ok := classCapTable[strings.ToLower(classID)]
	if !ok {
		return defaultCaps
	}
	return caps
}

var validAttributes = map[string]bool{
	"STR": true, "DEX": true, "INT": true, "WIS": true, "STA": true,
}

var defaultCaps = AttributeCaps{
	"STR": 20, "DEX": 20, "INT": 20, "WIS": 20, "STA": 20,
}

var classCapTable = map[string]AttributeCaps{
	"warrior": {"STR": 30, "DEX": 15, "INT": 5, "WIS": 5, "STA": 25},
	"rogue":   {"STR": 20, "DEX": 30, "INT": 5, "WIS": 5, "STA": 15},
	"wizard":  {"STR": 5, "DEX": 10, "INT": 30, "WIS": 25, "STA": 10},
	"cleric":  {"STR": 10, "DEX": 10, "INT": 20, "WIS": 30, "STA": 15},
	"ranger":  {"STR": 15, "DEX": 30, "INT": 5, "WIS": 10, "STA": 20},
	"hunter":  {"STR": 15, "DEX": 30, "INT": 5, "WIS": 10, "STA": 20},
	"druid":   {"STR": 5, "DEX": 15, "INT": 25, "WIS": 30, "STA": 10},
}

// ValidateAttributeSpend checks if spending `amount` points on `attrShort` is allowed.
// Returns an error message if invalid, or empty string if valid.
func ValidateAttributeSpend(char *characters.Character, attrShort string, amount int32) string {
	if amount <= 0 {
		return "Amount must be positive."
	}
	if amount > char.UnspentAttributePoints {
		return fmt.Sprintf("Not enough unspent points. You have %d.", char.UnspentAttributePoints)
	}

	if !validAttributes[attrShort] {
		return fmt.Sprintf("Unknown attribute '%s'. Valid: STR, DEX, INT, WIS, STA", attrShort)
	}

	caps := ClassAttributeCaps(char.Class.ID)
	currentSpent := int32(0)
	if char.SpentAttributePoints != nil {
		currentSpent = char.SpentAttributePoints[attrShort]
	}
	cap := caps[attrShort]
	if currentSpent+amount > cap {
		remaining := cap - currentSpent
		if remaining <= 0 {
			return fmt.Sprintf("You've reached the %s cap for your class (%d/%d).", attrShort, currentSpent, cap)
		}
		return fmt.Sprintf("Can only spend %d more on %s (cap: %d, spent: %d).", remaining, attrShort, cap, currentSpent)
	}

	return ""
}

// ApplyAttributeSpend applies distributable point spending to a character.
// Caller is responsible for persisting the character afterward.
func ApplyAttributeSpend(char *characters.Character, attrShort string, amount int32) {
	attrShort = strings.ToUpper(attrShort)

	if char.SpentAttributePoints == nil {
		char.SpentAttributePoints = make(map[string]int32)
	}

	char.UnspentAttributePoints -= amount
	char.SpentAttributePoints[attrShort] += amount

	for i := range char.Attributes {
		if strings.EqualFold(char.Attributes[i].Short, attrShort) {
			char.Attributes[i].Value += amount
			char.Attributes[i].Short = attrShort // normalize to uppercase
			return
		}
	}
}
