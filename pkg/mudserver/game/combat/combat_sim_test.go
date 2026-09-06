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

			floor := 0.85
			if cls.Name == "Druid" {
				floor = 0.70
			}
			if result.WinRate < floor {
				t.Errorf("Win rate %.1f%% is too low (expected >= %.0f%%)", result.WinRate*100, floor*100)
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

				floor := 0.30
				if cls.Name == "Mage" || cls.Name == "Druid" {
					floor = 0.05
				}
				if result.WinRate < floor {
					t.Errorf("Win rate %.1f%% is too low for L1 vs L2 normal (expected >= %.0f%%)", result.WinRate*100, floor*100)
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

				// Durable melee classes should dominate down-level trash; casters are informational.
				floor := 0.75
				if cls.Name == "Mage" || cls.Name == "Druid" {
					floor = 0.20
				} else if cls.Name == "Cleric" {
					floor = 0.40
				}
				if result.WinRate < floor {
					t.Errorf("Win rate %.1f%% is too low for L5 vs L2 (expected >= %.0f%%)", result.WinRate*100, floor*100)
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

				floor := 0.35
				if cls.Name == "Mage" || cls.Name == "Druid" {
					// Glass casters lose auto-attack races at starter gear; C6 duration uses Warrior.
					t.Logf("skipping hard floor for glass caster (win=%.1f%%)", result.WinRate*100)
					return
				}
				if result.WinRate < floor {
					t.Errorf("Win rate %.1f%% too low for same-level normal (expected >= %.0f%%)", result.WinRate*100, floor*100)
				}
			})
		}
	}
}

// --- Balance Tests: Bosses should be hard ---

func TestBossesRequireHigherLevel(t *testing.T) {
	// C6: bosses should be long (15–30 player turns) but killable for durable classes
	// at-level. Glass casters may still struggle — duration harness uses Warrior.
	testCases := []struct {
		enemyName  string
		level      int32
		minWinRate float64
		maxWinRate float64
		minRounds  float64
		maxRounds  float64
		classes    []string // empty = all
	}{
		{"Hollow Knight", 6, 0.25, 1.0, 12, 35, []string{"Warrior", "Ranger", "Rogue"}},
		{"Burrow Brute", 4, 0.30, 1.0, 12, 35, []string{"Warrior", "Ranger", "Rogue"}},
	}

	for _, cls := range simutil.AllClassConfigs() {
		for _, tc := range testCases {
			if len(tc.classes) > 0 {
				ok := false
				for _, n := range tc.classes {
					if cls.Name == n {
						ok = true
						break
					}
				}
				if !ok {
					continue
				}
			}
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

				if result.WinRate < tc.minWinRate || result.WinRate > tc.maxWinRate {
					t.Errorf("Win rate %.1f%% outside [%.0f%%, %.0f%%] for at-level boss",
						result.WinRate*100, tc.minWinRate*100, tc.maxWinRate*100)
				}
				if result.AvgRounds < tc.minRounds || result.AvgRounds > tc.maxRounds {
					t.Errorf("Avg rounds %.1f outside boss duration [%.0f, %.0f]",
						result.AvgRounds, tc.minRounds, tc.maxRounds)
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
