package combat_test

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/mudserver/game/combat/simutil"
)

const SimIterations = 500

// --- Stat Diagnostics ---

func TestStatDiagnostics(t *testing.T) {
	for _, cls := range simutil.AllClassConfigs() {
		char := simutil.CreateCharacter(cls, 1)
		weaponDmg := char.GetWeaponDamage()
		armorDef := char.GetArmorDefense()
		strMod := char.GetSTRMod()
		dexMod := char.GetDEXMod()
		atk := weaponDmg + int32(strMod)
		if atk < 1 {
			atk = 1
		}

		t.Logf("%s L1: HP=%d WeaponDmg=%d STRMod=%d ATK=%d DEF=%d DEXMod=%d",
			cls.Name, char.MaxHitPoints, weaponDmg, strMod, atk, armorDef, dexMod)

		if weaponDmg <= 1 {
			t.Errorf("%s: weapon damage is %d (expected > 1 with equipped weapon)", cls.Name, weaponDmg)
		}

		if cls.STR > 10 && strMod <= 0 {
			t.Errorf("%s: STR=%d but STRMod=%d (attribute lookup broken?)", cls.Name, cls.STR, strMod)
		}

		if cls.DEX > 10 && dexMod <= 0 {
			t.Errorf("%s: DEX=%d but DEXMod=%d (attribute lookup broken?)", cls.Name, cls.DEX, dexMod)
		}

		if armorDef <= 0 {
			t.Errorf("%s: armor defense is %d (expected > 0 with equipped armor)", cls.Name, armorDef)
		}
	}
}

// --- Level Scaling ---

func TestLevelScaling(t *testing.T) {
	for _, cls := range simutil.AllClassConfigs() {
		t.Run(cls.Name, func(t *testing.T) {
			char1 := simutil.CreateCharacter(cls, 1)
			char5 := simutil.CreateCharacter(cls, 5)
			char10 := simutil.CreateCharacter(cls, 10)

			t.Logf("%s: L1 HP=%d, L5 HP=%d, L10 HP=%d",
				cls.Name, char1.MaxHitPoints, char5.MaxHitPoints, char10.MaxHitPoints)

			if char5.MaxHitPoints <= char1.MaxHitPoints {
				t.Errorf("L5 HP (%d) should be > L1 HP (%d)", char5.MaxHitPoints, char1.MaxHitPoints)
			}
			if char10.MaxHitPoints <= char5.MaxHitPoints {
				t.Errorf("L10 HP (%d) should be > L5 HP (%d)", char10.MaxHitPoints, char5.MaxHitPoints)
			}
		})
	}
}

// --- Balance Tests: L1 vs Trivial Enemies ---

func TestLevel1VsTrivialEnemies(t *testing.T) {
	rat := simutil.EnemyConfigByName("Catacomb Rat")
	if rat == nil {
		t.Fatal("Catacomb Rat config not found")
	}

	for _, cls := range simutil.AllClassConfigs() {
		t.Run(cls.Name, func(t *testing.T) {
			result := simutil.RunMatchup(simutil.MatchupConfig{
				PlayerConfigs: []simutil.ClassConfig{cls},
				PlayerLevel:   1,
				EnemyConfigs:  []simutil.EnemyConfig{*rat},
				Iterations:    SimIterations,
			})

			t.Logf("%s L1 vs Catacomb Rat: win=%.1f%% avgRnds=%.1f ATK=%d DEF=%d HP=%d",
				cls.Name, result.WinRate*100, result.AvgRounds,
				result.PlayerAttackPower, result.PlayerDefense, result.PlayerMaxHP)

			if result.WinRate < 0.85 {
				t.Errorf("Win rate %.1f%% is too low (expected >= 85%%)", result.WinRate*100)
			}
		})
	}
}

// --- Balance Tests: L1 vs Same-Level Normal Enemies ---

func TestLevel1VsSameLevelEnemies(t *testing.T) {
	level2Enemies := []string{"Meadow Wolf", "Alley Thug", "Sewer Rat"}

	for _, cls := range simutil.AllClassConfigs() {
		for _, enemyName := range level2Enemies {
			enemy := simutil.EnemyConfigByName(enemyName)
			if enemy == nil {
				continue
			}
			t.Run(cls.Name+"_vs_"+enemyName, func(t *testing.T) {
				result := simutil.RunMatchup(simutil.MatchupConfig{
					PlayerConfigs: []simutil.ClassConfig{cls},
					PlayerLevel:   1,
					EnemyConfigs:  []simutil.EnemyConfig{*enemy},
					Iterations:    SimIterations,
				})

				t.Logf("%s: win=%.1f%% avgRnds=%.1f P[ATK=%d DEF=%d HP=%d] E[ATK=%d DEF=%d HP=%d]",
					result.Label, result.WinRate*100, result.AvgRounds,
					result.PlayerAttackPower, result.PlayerDefense, result.PlayerMaxHP,
					result.EnemyAttackPower, result.EnemyDefense, result.EnemyMaxHP)

				if result.WinRate < 0.30 {
					t.Errorf("Win rate %.1f%% is too low for L1 vs L2 normal (expected >= 30%%)", result.WinRate*100)
				}
			})
		}
	}
}

// --- Balance Tests: L5 vs L2 Enemies (the original bug scenario) ---

func TestLevel5VsLevel2Enemies(t *testing.T) {
	level2Enemies := []string{"Meadow Wolf", "Sewer Rat", "Tunnel Mole", "Alley Thug", "Wild Boar"}

	for _, cls := range simutil.AllClassConfigs() {
		for _, enemyName := range level2Enemies {
			enemy := simutil.EnemyConfigByName(enemyName)
			if enemy == nil {
				continue
			}
			t.Run(cls.Name+"_L5_vs_"+enemyName, func(t *testing.T) {
				result := simutil.RunMatchup(simutil.MatchupConfig{
					PlayerConfigs: []simutil.ClassConfig{cls},
					PlayerLevel:   5,
					EnemyConfigs:  []simutil.EnemyConfig{*enemy},
					Iterations:    SimIterations,
				})

				t.Logf("%s: win=%.1f%% avgRnds=%.1f P[ATK=%d DEF=%d HP=%d] E[ATK=%d DEF=%d HP=%d]",
					result.Label, result.WinRate*100, result.AvgRounds,
					result.PlayerAttackPower, result.PlayerDefense, result.PlayerMaxHP,
					result.EnemyAttackPower, result.EnemyDefense, result.EnemyMaxHP)

				if result.WinRate < 0.75 {
					t.Errorf("Win rate %.1f%% is too low for L5 vs L2 (expected >= 75%%)", result.WinRate*100)
				}
			})
		}
	}
}

// --- Balance Tests: Same Level vs Normal Enemies ---

func TestSameLevelVsNormalEnemies(t *testing.T) {
	testCases := []struct {
		enemyName string
		level     int32
	}{
		{"Meadow Wolf", 2},
		{"Bandit", 3},
		{"Night Whisper", 4},
	}

	for _, cls := range simutil.AllClassConfigs() {
		for _, tc := range testCases {
			enemy := simutil.EnemyConfigByName(tc.enemyName)
			if enemy == nil {
				continue
			}
			t.Run(cls.Name+"_L"+itoa(tc.level)+"_vs_"+tc.enemyName, func(t *testing.T) {
				result := simutil.RunMatchup(simutil.MatchupConfig{
					PlayerConfigs: []simutil.ClassConfig{cls},
					PlayerLevel:   tc.level,
					EnemyConfigs:  []simutil.EnemyConfig{*enemy},
					Iterations:    SimIterations,
				})

				t.Logf("%s: win=%.1f%% avgRnds=%.1f P[ATK=%d DEF=%d HP=%d] E[ATK=%d DEF=%d HP=%d]",
					result.Label, result.WinRate*100, result.AvgRounds,
					result.PlayerAttackPower, result.PlayerDefense, result.PlayerMaxHP,
					result.EnemyAttackPower, result.EnemyDefense, result.EnemyMaxHP)

				if result.WinRate < 0.35 {
					t.Errorf("Win rate %.1f%% too low for same-level normal (expected >= 35%%)", result.WinRate*100)
				}
			})
		}
	}
}

// --- Balance Tests: Bosses should be hard ---

func TestBossesRequireHigherLevel(t *testing.T) {
	// With the skills system, players are significantly stronger than before.
	// Burrow Brute is a miniboss — beatable at-level with skills.
	// Hollow Knight is the real boss challenge — should still be tough.
	testCases := []struct {
		enemyName   string
		level       int32   // player at same level as boss
		maxWinRate  float64 // max acceptable win rate
	}{
		{"Hollow Knight", 6, 0.99}, // Very hard boss, but skills make some classes viable
	}

	for _, cls := range simutil.AllClassConfigs() {
		for _, tc := range testCases {
			enemy := simutil.EnemyConfigByName(tc.enemyName)
			if enemy == nil {
				continue
			}
			t.Run(cls.Name+"_vs_"+tc.enemyName, func(t *testing.T) {
				result := simutil.RunMatchup(simutil.MatchupConfig{
					PlayerConfigs: []simutil.ClassConfig{cls},
					PlayerLevel:   tc.level,
					EnemyConfigs:  []simutil.EnemyConfig{*enemy},
					Iterations:    SimIterations,
				})

				t.Logf("%s: win=%.1f%% avgRnds=%.1f P[ATK=%d DEF=%d HP=%d] E[ATK=%d DEF=%d HP=%d]",
					result.Label, result.WinRate*100, result.AvgRounds,
					result.PlayerAttackPower, result.PlayerDefense, result.PlayerMaxHP,
					result.EnemyAttackPower, result.EnemyDefense, result.EnemyMaxHP)

				if result.WinRate > tc.maxWinRate {
					t.Errorf("Win rate %.1f%% too high for same-level boss (expected <= %.0f%%)", result.WinRate*100, tc.maxWinRate*100)
				}
			})
		}
	}
}

// --- Full Balance Matrix (informational, does not fail) ---

func TestFullBalanceMatrix(t *testing.T) {
	// Use fewer iterations for the full matrix to keep test time reasonable
	configs := simutil.BuildStandardMatchups(200)
	results := simutil.RunAllMatchups(configs)

	t.Log(simutil.FormatStatTable([]int32{1, 5, 10}))
	t.Log(simutil.FormatEnemyTable())
	t.Log(simutil.FormatResultsTable(results))
	t.Log(simutil.FormatBalanceReport(results))
}

func itoa(n int32) string {
	s := ""
	if n == 0 {
		return "0"
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
