package leveling

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/characters"
)

// TestXPCurveFormula verifies the XP formula produces values close to expected targets
func TestXPCurveFormula(t *testing.T) {
	tests := []struct {
		level    int32
		target   int32
		maxError float64 // Maximum acceptable percentage error
	}{
		{2, 150, 20.0},      // Target ~150 XP
		{5, 661, 20.0},      // Target ~661 XP
		{10, 1681, 20.0},    // Target ~1681 XP
		{20, 5478, 20.0},    // Target ~5478 XP
		{30, 11935, 20.0},   // Target ~11935 XP
		{40, 20905, 20.0},   // Target ~20905 XP
		{50, 35455, 20.0},   // Target ~35455 XP
	}

	for _, tt := range tests {
		xp := GetXPRequired(tt.level)
		diff := float64(xp-tt.target) / float64(tt.target) * 100.0
		if diff < 0 {
			diff = -diff
		}
		if diff > tt.maxError {
			t.Errorf("Level %d: expected ~%d XP (±%.0f%%), got %d (%.1f%% error)",
				tt.level, tt.target, tt.maxError, xp, diff)
		}
	}
}

// TestXPTableGeneration verifies the precomputed XP table matches formula
func TestXPTableGeneration(t *testing.T) {
	for level := int32(1); level <= 50; level++ {
		computed := CalculateXPRequired(level)
		stored := XPTable[level]

		if computed != stored {
			t.Errorf("Level %d: XPTable mismatch. Computed=%d, Stored=%d",
				level, computed, stored)
		}
	}
}

// TestLevelUpSingleLevel tests leveling from 1 to 2
func TestLevelUpSingleLevel(t *testing.T) {
	char := createTestCharacter(characters.ClassWarrior, 1)

	// Give exact XP for level 2
	char.XP = GetXPRequired(2)

	levelsGained, newLevel := CheckLevelUp(char)

	if levelsGained != 1 {
		t.Errorf("Expected 1 level gained, got %d", levelsGained)
	}
	if newLevel != 2 {
		t.Errorf("Expected new level 2, got %d", newLevel)
	}
}

// TestLevelUpMultipleLevels tests gaining multiple levels at once
func TestLevelUpMultipleLevels(t *testing.T) {
	char := createTestCharacter(characters.ClassWarrior, 1)

	// Give XP for level 5
	char.XP = GetXPRequired(5)

	levelsGained, newLevel := CheckLevelUp(char)

	if levelsGained != 4 {
		t.Errorf("Expected 4 levels gained, got %d", levelsGained)
	}
	if newLevel != 5 {
		t.Errorf("Expected new level 5, got %d", newLevel)
	}
}

// TestLevelUpAtMaxCap tests that level 50 is the hard cap
func TestLevelUpAtMaxCap(t *testing.T) {
	char := createTestCharacter(characters.ClassWarrior, 50)

	// Give massive XP
	char.XP = 999999

	levelsGained, newLevel := CheckLevelUp(char)

	if levelsGained != 0 {
		t.Errorf("Expected 0 levels gained at max cap, got %d", levelsGained)
	}
	if newLevel != 50 {
		t.Errorf("Expected level to stay at 50, got %d", newLevel)
	}
}

// TestLevelUpJustBelowThreshold tests that XP just below threshold doesn't level up
func TestLevelUpJustBelowThreshold(t *testing.T) {
	char := createTestCharacter(characters.ClassWarrior, 1)

	// Give XP just below level 2 threshold
	char.XP = GetXPRequired(2) - 1

	levelsGained, _ := CheckLevelUp(char)

	if levelsGained != 0 {
		t.Errorf("Expected 0 levels gained with insufficient XP, got %d", levelsGained)
	}
}

// TestHPScalingWarrior tests HP scaling for high-STA class
func TestHPScalingWarrior(t *testing.T) {
	char := createTestCharacter(characters.ClassWarrior, 1)
	initialHP := char.MaxHitPoints

	// Set Stamina to 20 (modifier +5)
	setAttributeValue(char, "STA", 20)

	// Level up once
	hpGained := CalculateHPGain(char, 1)

	// Expected: 5 + (5 * 2) = 15 HP per level
	expectedHPGain := int32(15)

	if hpGained != expectedHPGain {
		t.Errorf("Warrior HP gain: expected %d, got %d", expectedHPGain, hpGained)
	}

	// Apply level-up
	result := ApplyLevelUp(char, 1)

	if char.MaxHitPoints != initialHP+hpGained {
		t.Errorf("MaxHitPoints not updated correctly. Expected %d, got %d",
			initialHP+hpGained, char.MaxHitPoints)
	}

	if char.CurrentHitPoints != char.MaxHitPoints {
		t.Errorf("CurrentHitPoints not restored. Expected %d, got %d",
			char.MaxHitPoints, char.CurrentHitPoints)
	}

	if result == nil {
		t.Error("ApplyLevelUp returned nil result")
	}
}

// TestHPScalingWizard tests HP scaling for low-STA class
func TestHPScalingWizard(t *testing.T) {
	char := createTestCharacter(characters.ClassWizard, 1)

	// Set Stamina to 8 (modifier -1)
	setAttributeValue(char, "STA", 8)

	// Level up once
	hpGained := CalculateHPGain(char, 1)

	// Expected: 5 + (-1 * 2) = 3 HP per level
	expectedHPGain := int32(3)

	if hpGained != expectedHPGain {
		t.Errorf("Wizard HP gain: expected %d, got %d", expectedHPGain, hpGained)
	}
}

// TestHPScalingVeryLowSTA tests that minimum HP is gained even with very low STA
func TestHPScalingVeryLowSTA(t *testing.T) {
	char := createTestCharacter(characters.ClassWizard, 1)

	// Set extremely low Stamina (modifier -5)
	setAttributeValue(char, "STA", 0)

	// Level up once
	hpGained := CalculateHPGain(char, 1)

	// Should still gain at least 1 HP per level (minimum enforced)
	if hpGained < 1 {
		t.Errorf("HP gain should be at least 1, got %d", hpGained)
	}
}

// TestAttributeScalingPrimary tests +1 primary attribute every 5 levels
func TestAttributeScalingPrimary(t *testing.T) {
	char := createTestCharacter(characters.ClassWarrior, 1)
	initialSTR := getAttributeValue(char, "STR")

	// Level from 1 to 10 (should gain +2 STR at levels 5 and 10)
	gains := CalculateAttributeGains(char, 1, 10)

	expectedSTRGain := int32(2)
	if gains["STR"] != expectedSTRGain {
		t.Errorf("Expected +%d STR from levels 1-10, got %d",
			expectedSTRGain, gains["STR"])
	}

	// Apply gains
	for attr, gainAmount := range gains {
		applyAttributeGain(char, attr, gainAmount)
	}

	newSTR := getAttributeValue(char, "STR")
	if newSTR != initialSTR+expectedSTRGain {
		t.Errorf("STR not updated correctly. Expected %d, got %d",
			initialSTR+expectedSTRGain, newSTR)
	}
}

// TestAttributeScalingSecondary tests +1 secondary attribute every 10 levels
func TestAttributeScalingSecondary(t *testing.T) {
	char := createTestCharacter(characters.ClassWarrior, 1)

	// Level from 1 to 20 (should gain +2 STA at levels 10 and 20)
	gains := CalculateAttributeGains(char, 1, 20)

	expectedSTAGain := int32(2)
	if gains["STA"] != expectedSTAGain {
		t.Errorf("Expected +%d STA from levels 1-20, got %d",
			expectedSTAGain, gains["STA"])
	}
}

// TestAttributeScalingPartialLevels tests attribute gains for partial level ranges
func TestAttributeScalingPartialLevels(t *testing.T) {
	char := createTestCharacter(characters.ClassWarrior, 8)

	// Level from 8 to 11 (should gain +1 STR at level 10, +1 STA at level 10)
	gains := CalculateAttributeGains(char, 8, 11)

	if gains["STR"] != 1 {
		t.Errorf("Expected +1 STR from levels 8-11, got %d", gains["STR"])
	}
	if gains["STA"] != 1 {
		t.Errorf("Expected +1 STA from levels 8-11, got %d", gains["STA"])
	}
}

// TestAttributeScalingNoGains tests levels where no attributes are gained
func TestAttributeScalingNoGains(t *testing.T) {
	char := createTestCharacter(characters.ClassWarrior, 1)

	// Level from 1 to 3 (no milestone levels)
	gains := CalculateAttributeGains(char, 1, 3)

	if len(gains) != 0 {
		t.Errorf("Expected no attribute gains from levels 1-3, got %v", gains)
	}
}

// TestApplyLevelUpFullIntegration tests the full level-up flow
func TestApplyLevelUpFullIntegration(t *testing.T) {
	char := createTestCharacter(characters.ClassWarrior, 4)
	char.XP = GetXPRequired(6) // Enough for level 6

	initialHP := char.MaxHitPoints
	initialSTR := getAttributeValue(char, "STR")

	// Check and apply level-up
	levelsGained, _ := CheckLevelUp(char)
	result := ApplyLevelUp(char, levelsGained)

	// Verify result
	if result == nil {
		t.Fatal("ApplyLevelUp returned nil result")
	}

	if result.OldLevel != 4 {
		t.Errorf("OldLevel: expected 4, got %d", result.OldLevel)
	}
	if result.NewLevel != 6 {
		t.Errorf("NewLevel: expected 6, got %d", result.NewLevel)
	}
	if result.LevelsGained != 2 {
		t.Errorf("LevelsGained: expected 2, got %d", result.LevelsGained)
	}

	// Verify level was updated
	if char.Level != 6 {
		t.Errorf("Character level: expected 6, got %d", char.Level)
	}

	// Verify HP increased
	if char.MaxHitPoints <= initialHP {
		t.Errorf("MaxHitPoints should have increased from %d, got %d",
			initialHP, char.MaxHitPoints)
	}

	// Verify HP was fully restored
	if char.CurrentHitPoints != char.MaxHitPoints {
		t.Errorf("HP not fully restored: current=%d, max=%d",
			char.CurrentHitPoints, char.MaxHitPoints)
	}

	// Verify attribute gain at level 5
	newSTR := getAttributeValue(char, "STR")
	if newSTR != initialSTR+1 {
		t.Errorf("STR should have increased by 1 (level 5 milestone), from %d to %d, got %d",
			initialSTR, initialSTR+1, newSTR)
	}

	// Verify message was generated
	if result.Message == "" {
		t.Error("Level-up message should not be empty")
	}
}

// TestEnemyXPReward tests XP calculation for enemies
func TestEnemyXPReward(t *testing.T) {
	tests := []struct {
		enemyLevel int32
		expectedXP int64
	}{
		{1, 10},
		{5, 50},
		{10, 100},
		{20, 200},
		{30, 300},
		{50, 500},
	}

	for _, tt := range tests {
		xp := CalculateEnemyXPReward(tt.enemyLevel)
		if xp != tt.expectedXP {
			t.Errorf("Enemy level %d: expected %d XP, got %d",
				tt.enemyLevel, tt.expectedXP, xp)
		}
	}
}

// TestGetXPForNextLevel tests the XP remaining calculation
func TestGetXPForNextLevel(t *testing.T) {
	// Level 1 with 50 XP (needs 150 for level 2)
	remaining := GetXPForNextLevel(1, 50)
	expected := GetXPRequired(2) - 50

	if remaining != expected {
		t.Errorf("Expected %d XP remaining, got %d", expected, remaining)
	}

	// At max level
	remaining = GetXPForNextLevel(50, 999999)
	if remaining != 0 {
		t.Errorf("At max level, expected 0 XP remaining, got %d", remaining)
	}
}

// TestGetXPProgress tests the progress percentage calculation
func TestGetXPProgress(t *testing.T) {
	// 50% progress to level 2
	level2XP := GetXPRequired(2)
	halfwayXP := level2XP / 2

	progress := GetXPProgress(1, halfwayXP)

	if progress < 45.0 || progress > 55.0 {
		t.Errorf("Expected ~50%% progress, got %.2f%%", progress)
	}

	// 100% at max level
	progress = GetXPProgress(50, 999999)
	if progress != 100.0 {
		t.Errorf("At max level, expected 100%% progress, got %.2f%%", progress)
	}
}

// TestClassAttributeMapping tests that each class has correct primary/secondary attributes
func TestClassAttributeMapping(t *testing.T) {
	tests := []struct {
		class     characters.Class
		primary   string
		secondary string
	}{
		{characters.ClassWarrior, "STR", "STA"},
		{characters.ClassHunter, "DEX", "STA"},
		{characters.ClassRogue, "DEX", "STR"},
		{characters.ClassWizard, "INT", "WIS"},
	}

	for _, tt := range tests {
		primary := GetPrimaryAttribute(tt.class)
		secondary := GetSecondaryAttribute(tt.class)

		if primary != tt.primary {
			t.Errorf("%s: expected primary %s, got %s",
				tt.class, tt.primary, primary)
		}
		if secondary != tt.secondary {
			t.Errorf("%s: expected secondary %s, got %s",
				tt.class, tt.secondary, secondary)
		}
	}
}

// Helper functions for testing

func createTestCharacter(class characters.Class, level int32) *characters.Character {
	char := &characters.Character{
		Level:            level,
		XP:               0,
		MaxHitPoints:     25,
		CurrentHitPoints: 25,
		Class:            class,
		Attributes: []characters.Attribute{
			{Short: "STR", Value: 10},
			{Short: "DEX", Value: 10},
			{Short: "INT", Value: 10},
			{Short: "WIS", Value: 10},
			{Short: "STA", Value: 10},
		},
	}
	return char
}

func setAttributeValue(char *characters.Character, shortName string, value int32) {
	for i := range char.Attributes {
		if char.Attributes[i].Short == shortName {
			char.Attributes[i].Value = value
			return
		}
	}
}
