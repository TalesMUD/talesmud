package leveling

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/characters"
)

func TestValidateAttributeSpend_Valid(t *testing.T) {
	char := createTestCharacter(characters.ClassWarrior, 1)
	char.UnspentAttributePoints = 10

	errMsg := ValidateAttributeSpend(char, "STR", 3)
	if errMsg != "" {
		t.Errorf("Expected valid spend, got: %s", errMsg)
	}
}

func TestValidateAttributeSpend_InsufficientPoints(t *testing.T) {
	char := createTestCharacter(characters.ClassWarrior, 1)
	char.UnspentAttributePoints = 2

	errMsg := ValidateAttributeSpend(char, "STR", 5)
	if errMsg == "" {
		t.Error("Expected error for insufficient points, got none")
	}
}

func TestValidateAttributeSpend_InvalidAttribute(t *testing.T) {
	char := createTestCharacter(characters.ClassWarrior, 1)
	char.UnspentAttributePoints = 10

	errMsg := ValidateAttributeSpend(char, "FOO", 1)
	if errMsg == "" {
		t.Error("Expected error for invalid attribute, got none")
	}
}

func TestValidateAttributeSpend_NegativeAmount(t *testing.T) {
	char := createTestCharacter(characters.ClassWarrior, 1)
	char.UnspentAttributePoints = 10

	errMsg := ValidateAttributeSpend(char, "STR", -1)
	if errMsg == "" {
		t.Error("Expected error for negative amount, got none")
	}
}

func TestValidateAttributeSpend_ExceedsCap(t *testing.T) {
	char := createTestCharacter(characters.ClassWizard, 1)
	char.UnspentAttributePoints = 50

	// Wizard STR cap is 5
	errMsg := ValidateAttributeSpend(char, "STR", 6)
	if errMsg == "" {
		t.Error("Expected cap violation error, got none")
	}

	// Should allow exactly 5
	errMsg = ValidateAttributeSpend(char, "STR", 5)
	if errMsg != "" {
		t.Errorf("Expected valid spend of 5, got: %s", errMsg)
	}
}

func TestValidateAttributeSpend_AlreadyAtCap(t *testing.T) {
	char := createTestCharacter(characters.ClassWizard, 1)
	char.UnspentAttributePoints = 50
	char.SpentAttributePoints = map[string]int32{"STR": 5} // Already at wizard STR cap

	errMsg := ValidateAttributeSpend(char, "STR", 1)
	if errMsg == "" {
		t.Error("Expected cap reached error, got none")
	}
}

func TestApplyAttributeSpend_Basic(t *testing.T) {
	char := createTestCharacter(characters.ClassWarrior, 1)
	char.UnspentAttributePoints = 10
	initialSTR := getAttributeValue(char, "STR")

	ApplyAttributeSpend(char, "STR", 3)

	if char.UnspentAttributePoints != 7 {
		t.Errorf("Expected 7 unspent, got %d", char.UnspentAttributePoints)
	}
	if char.SpentAttributePoints["STR"] != 3 {
		t.Errorf("Expected 3 spent on STR, got %d", char.SpentAttributePoints["STR"])
	}
	if getAttributeValue(char, "STR") != initialSTR+3 {
		t.Errorf("Expected STR=%d, got %d", initialSTR+3, getAttributeValue(char, "STR"))
	}
}

func TestApplyAttributeSpend_NilSpentMap(t *testing.T) {
	char := createTestCharacter(characters.ClassWarrior, 1)
	char.UnspentAttributePoints = 5
	char.SpentAttributePoints = nil // Simulate pre-existing character

	ApplyAttributeSpend(char, "DEX", 2)

	if char.SpentAttributePoints == nil {
		t.Fatal("SpentAttributePoints should have been initialized")
	}
	if char.SpentAttributePoints["DEX"] != 2 {
		t.Errorf("Expected 2 spent on DEX, got %d", char.SpentAttributePoints["DEX"])
	}
}

func TestApplyAttributeSpend_Cumulative(t *testing.T) {
	char := createTestCharacter(characters.ClassWarrior, 1)
	char.UnspentAttributePoints = 20

	ApplyAttributeSpend(char, "STR", 5)
	ApplyAttributeSpend(char, "STR", 3)

	if char.SpentAttributePoints["STR"] != 8 {
		t.Errorf("Expected 8 total spent on STR, got %d", char.SpentAttributePoints["STR"])
	}
	if char.UnspentAttributePoints != 12 {
		t.Errorf("Expected 12 unspent, got %d", char.UnspentAttributePoints)
	}
}

func TestLevelUpAwardsPoints(t *testing.T) {
	char := createTestCharacter(characters.ClassWarrior, 1)
	char.XP = GetXPRequired(3) // Level to 3

	levelsGained, _ := CheckLevelUp(char)
	result := ApplyLevelUp(char, levelsGained)

	expectedPoints := int32(levelsGained) * int32(PointsPerLevel)
	if char.UnspentAttributePoints != expectedPoints {
		t.Errorf("Expected %d unspent points, got %d", expectedPoints, char.UnspentAttributePoints)
	}
	if result.AttributePointsGained != expectedPoints {
		t.Errorf("Expected %d in result, got %d", expectedPoints, result.AttributePointsGained)
	}
}

func TestLevelUpAwardsPoints_MultiLevel(t *testing.T) {
	char := createTestCharacter(characters.ClassWarrior, 1)
	char.XP = GetXPRequired(10) // Level to 10

	levelsGained, _ := CheckLevelUp(char)
	result := ApplyLevelUp(char, levelsGained)

	// 9 levels * 2 points = 18
	expectedPoints := int32(9) * int32(PointsPerLevel)
	if char.UnspentAttributePoints != expectedPoints {
		t.Errorf("Expected %d unspent points, got %d", expectedPoints, char.UnspentAttributePoints)
	}
	if result.AttributePointsGained != expectedPoints {
		t.Errorf("Expected %d in result, got %d", expectedPoints, result.AttributePointsGained)
	}
}

func TestLevelUpAwardsPoints_Accumulates(t *testing.T) {
	char := createTestCharacter(characters.ClassWarrior, 1)
	char.UnspentAttributePoints = 5 // Already had some points

	char.XP = GetXPRequired(3) // Level to 3
	levelsGained, _ := CheckLevelUp(char)
	ApplyLevelUp(char, levelsGained)

	// Should accumulate: 5 + (2 levels * 2 points) = 9
	expectedPoints := int32(5) + int32(levelsGained)*int32(PointsPerLevel)
	if char.UnspentAttributePoints != expectedPoints {
		t.Errorf("Expected %d unspent points, got %d", expectedPoints, char.UnspentAttributePoints)
	}
}

func TestClassAttributeCaps_AllClassesHaveAllAttributes(t *testing.T) {
	classes := []string{"warrior", "rogue", "wizard", "cleric", "ranger", "hunter", "druid"}
	attrs := []string{"STR", "DEX", "INT", "WIS", "STA"}

	for _, classID := range classes {
		caps := ClassAttributeCaps(classID)
		for _, attr := range attrs {
			if _, ok := caps[attr]; !ok {
				t.Errorf("Class %s missing cap for %s", classID, attr)
			}
			if caps[attr] <= 0 {
				t.Errorf("Class %s has non-positive cap for %s: %d", classID, attr, caps[attr])
			}
		}
	}
}

func TestClassAttributeCaps_UnknownClass(t *testing.T) {
	caps := ClassAttributeCaps("unknown_class")

	// Should return default caps
	if caps["STR"] != 20 {
		t.Errorf("Expected default STR cap of 20, got %d", caps["STR"])
	}
}

func TestClassAttributeCaps_WarriorFavorsSTR(t *testing.T) {
	caps := ClassAttributeCaps("warrior")

	if caps["STR"] <= caps["INT"] {
		t.Errorf("Warrior STR cap (%d) should be higher than INT cap (%d)", caps["STR"], caps["INT"])
	}
}

func TestClassAttributeCaps_WizardFavorsINT(t *testing.T) {
	caps := ClassAttributeCaps("wizard")

	if caps["INT"] <= caps["STR"] {
		t.Errorf("Wizard INT cap (%d) should be higher than STR cap (%d)", caps["INT"], caps["STR"])
	}
}

func TestPointsPerLevelConstant(t *testing.T) {
	if PointsPerLevel != 2 {
		t.Errorf("Expected PointsPerLevel=2, got %d", PointsPerLevel)
	}
}

func TestLevelUpMessageIncludesPoints(t *testing.T) {
	char := createTestCharacter(characters.ClassWarrior, 1)
	char.XP = GetXPRequired(2)

	levelsGained, _ := CheckLevelUp(char)
	result := ApplyLevelUp(char, levelsGained)

	if result.Message == "" {
		t.Fatal("Expected non-empty message")
	}

	// Message should mention attribute points
	if !containsSubstring(result.Message, "Attribute Point") {
		t.Error("Level-up message should mention attribute points")
	}
	if !containsSubstring(result.Message, "spend") {
		t.Error("Level-up message should mention the spend command")
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
