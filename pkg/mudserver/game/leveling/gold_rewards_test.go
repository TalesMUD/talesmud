package leveling

import "testing"

func TestCalculateEnemyGoldRewardNormal(t *testing.T) {
	tests := []struct {
		level   int32
		wantMin int32
		wantMax int32
	}{
		{1, 1, 2},    // 1+0=1, 1+0+0=1 → max=2 (min+1)
		{2, 1, 2},    // 1+0=1, 1+1+0=2
		{3, 1, 2},    // 1+0=1, 1+1+0=2
		{5, 2, 4},    // 1+1=2, 1+2+1=4
		{10, 3, 8},   // 1+2=3, 1+5+2=8
		{15, 4, 11},  // 1+3=4, 1+7+3=11
		{20, 5, 15},  // 1+4=5, 1+10+4=15
		{30, 7, 22},  // 1+6=7, 1+15+6=22
		{50, 11, 36}, // 1+10=11, 1+25+10=36
	}

	for _, tt := range tests {
		minGold, maxGold := CalculateEnemyGoldReward(tt.level, "normal")
		if minGold != tt.wantMin || maxGold != tt.wantMax {
			t.Errorf("Level %d normal: expected %d-%d gold, got %d-%d",
				tt.level, tt.wantMin, tt.wantMax, minGold, maxGold)
		}
	}
}

func TestCalculateEnemyGoldRewardDifficulty(t *testing.T) {
	tests := []struct {
		level      int32
		difficulty string
		wantMin    int32
		wantMax    int32
	}{
		// Trivial (0.5x): L10 base 3-8 → 2-4
		{10, "trivial", 2, 4},
		// Easy (0.75x): L10 base 3-8 → 2-6
		{10, "easy", 2, 6},
		// Hard (1.5x): L10 base 3-8 → 5-12
		{10, "hard", 5, 12},
		// Boss (3x): L10 base 3-8 → 9-24
		{10, "boss", 9, 24},
		// Boss L1: base 1-2 → 3-6
		{1, "boss", 3, 6},
		// Hard L5: base 2-4 → 3-6
		{5, "hard", 3, 6},
	}

	for _, tt := range tests {
		minGold, maxGold := CalculateEnemyGoldReward(tt.level, tt.difficulty)
		if minGold != tt.wantMin || maxGold != tt.wantMax {
			t.Errorf("Level %d %s: expected %d-%d gold, got %d-%d",
				tt.level, tt.difficulty, tt.wantMin, tt.wantMax, minGold, maxGold)
		}
	}
}

func TestCalculateEnemyGoldRewardMinimums(t *testing.T) {
	// Level 0 or negative should clamp to level 1
	minGold, maxGold := CalculateEnemyGoldReward(0, "normal")
	if minGold < 1 {
		t.Errorf("Min gold should be at least 1, got %d", minGold)
	}
	if maxGold <= minGold {
		t.Errorf("Max gold (%d) should be greater than min gold (%d)", maxGold, minGold)
	}

	minGold, maxGold = CalculateEnemyGoldReward(-5, "normal")
	if minGold < 1 {
		t.Errorf("Negative level: min gold should be at least 1, got %d", minGold)
	}

	// Unknown difficulty defaults to normal
	minNorm, maxNorm := CalculateEnemyGoldReward(10, "normal")
	minUnk, maxUnk := CalculateEnemyGoldReward(10, "unknown")
	if minNorm != minUnk || maxNorm != maxUnk {
		t.Errorf("Unknown difficulty should default to normal: normal=%d-%d, unknown=%d-%d",
			minNorm, maxNorm, minUnk, maxUnk)
	}
}

func TestCalculateEnemyGoldRewardMaxAlwaysGreaterThanMin(t *testing.T) {
	difficulties := []string{"trivial", "easy", "normal", "hard", "boss"}
	for _, diff := range difficulties {
		for level := int32(1); level <= 50; level++ {
			minGold, maxGold := CalculateEnemyGoldReward(level, diff)
			if maxGold <= minGold {
				t.Errorf("Level %d %s: max (%d) should be > min (%d)",
					level, diff, maxGold, minGold)
			}
			if minGold < 1 {
				t.Errorf("Level %d %s: min (%d) should be >= 1",
					level, diff, minGold)
			}
		}
	}
}

func TestRollEnemyGold(t *testing.T) {
	// With a deterministic "random" that always returns 0, should get min
	gold := RollEnemyGold(10, "normal", func(n int) int { return 0 })
	minGold, _ := CalculateEnemyGoldReward(10, "normal")
	if gold != int64(minGold) {
		t.Errorf("RollEnemyGold with min roll: expected %d, got %d", minGold, gold)
	}

	// With a "random" that returns max, should get max
	_, maxGold := CalculateEnemyGoldReward(10, "normal")
	gold = RollEnemyGold(10, "normal", func(n int) int { return n - 1 })
	if gold != int64(maxGold) {
		t.Errorf("RollEnemyGold with max roll: expected %d, got %d", maxGold, gold)
	}
}

func TestGetDifficultyGoldMultiplier(t *testing.T) {
	tests := []struct {
		difficulty string
		want       float64
	}{
		{"trivial", 0.5},
		{"easy", 0.75},
		{"normal", 1.0},
		{"hard", 1.5},
		{"boss", 3.0},
		{"", 1.0},
		{"unknown", 1.0},
	}

	for _, tt := range tests {
		got := GetDifficultyGoldMultiplier(tt.difficulty)
		if got != tt.want {
			t.Errorf("Difficulty %q: expected %.2f, got %.2f", tt.difficulty, tt.want, got)
		}
	}
}
