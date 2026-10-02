package combat_test

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
	"github.com/talesmud/talesmud/pkg/mudserver/game/combat/simutil"
)

// C6 duration bands (player turns ≈ combat rounds in 1v1).
// Reference: Warrior with auto-spent primary attributes + skills AI.
const durationIterations = 300

func TestCombatDurationTargets(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Logf("ReloadConfig: %v (using defaults if file missing)", err)
	}

	warrior := *simutil.ClassConfigByName("Warrior")

	cases := []struct {
		name      string
		level     int32
		enemy     string
		band      string
		minRounds float64
		maxRounds float64
		minWin    float64 // minimum win rate for a fair at-level fight
	}{
		{"trash_rat", 1, "Sample Rat", "trash", 3, 6, 0.90},
		{"trash_sewer", 2, "Sample Pest", "trash", 3, 6, 0.90},
		{"trash_mole", 2, "Sample Burrower", "trash", 3, 7, 0.85}, // mole DEF can nudge past 6
		{"elite_wolf", 2, "Sample Wolf", "elite", 8, 15, 0.85},
		{"elite_bandit", 3, "Sample Bandit", "elite", 8, 15, 0.75},
		{"elite_bear", 5, "Sample Bear", "elite", 8, 16, 0.55},
		{"boss_brute", 4, "Sample Brute", "boss", 15, 30, 0.45},
		{"boss_warden", 6, "Sample Warden", "boss", 15, 30, 0.35},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			en := simutil.EnemyConfigByName(tc.enemy)
			if en == nil {
				t.Fatalf("enemy %q not found", tc.enemy)
			}
			result := simutil.RunMatchup(simutil.MatchupConfig{
				PlayerConfigs: []simutil.ClassConfig{warrior},
				PlayerLevel:   tc.level,
				EnemyConfigs:  []simutil.EnemyConfig{*en},
				Iterations:    durationIterations,
			})

			t.Logf("%s: win=%.1f%% avgRnds=%.1f (band %s %.0f–%.0f) P[ATK=%d DEF=%d HP=%d] E[ATK=%d DEF=%d HP=%d]",
				result.Label, result.WinRate*100, result.AvgRounds, tc.band, tc.minRounds, tc.maxRounds,
				result.PlayerAttackPower, result.PlayerDefense, result.PlayerMaxHP,
				result.EnemyAttackPower, result.EnemyDefense, result.EnemyMaxHP)

			if result.AvgRounds < tc.minRounds || result.AvgRounds > tc.maxRounds {
				t.Errorf("avg rounds %.1f outside %s band [%.0f, %.0f]",
					result.AvgRounds, tc.band, tc.minRounds, tc.maxRounds)
			}
			if result.WinRate < tc.minWin {
				t.Errorf("win rate %.1f%% below floor %.0f%% for at-level %s",
					result.WinRate*100, tc.minWin*100, tc.band)
			}
		})
	}
}

// TestDurationBeforeAfterSnapshot logs a compact matrix for docs (informational).
func TestDurationBeforeAfterSnapshot(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Logf("ReloadConfig: %v", err)
	}
	warrior := *simutil.ClassConfigByName("Warrior")
	enemies := []struct {
		level int32
		name  string
	}{
		{1, "Sample Rat"},
		{2, "Sample Pest"},
		{2, "Sample Wolf"},
		{3, "Sample Bandit"},
		{5, "Sample Bear"},
		{4, "Sample Brute"},
		{6, "Sample Warden"},
	}
	for _, e := range enemies {
		en := simutil.EnemyConfigByName(e.name)
		r := simutil.RunMatchup(simutil.MatchupConfig{
			PlayerConfigs: []simutil.ClassConfig{warrior},
			PlayerLevel:   e.level,
			EnemyConfigs:  []simutil.EnemyConfig{*en},
			Iterations:    200,
		})
		t.Logf("AFTER  Warrior L%d vs %s: win=%.1f%% avgRnds=%.1f E[HP=%d ATK=%d DEF=%d]",
			e.level, e.name, r.WinRate*100, r.AvgRounds, r.EnemyMaxHP, r.EnemyAttackPower, r.EnemyDefense)
	}
}
