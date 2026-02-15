package combat_test

import (
	"fmt"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/mudserver/game/combat/simutil"
)

// TestWarriorBalanceAnalysis runs detailed "what-if" scenarios for Warrior tuning
func TestWarriorBalanceAnalysis(t *testing.T) {
	warrior := *simutil.ClassConfigByName("Warrior")

	fmt.Println("\n========================================")
	fmt.Println("  WARRIOR BALANCE ANALYSIS")
	fmt.Println("========================================")

	// Current Hollow Knight stats
	fmt.Println("\n--- CURRENT: Hollow Knight (HP=150, ATK=13, DEF=6) ---")
	for _, lvl := range []int32{5, 6, 7, 8, 10} {
		r := simutil.RunMatchup(simutil.MatchupConfig{
			PlayerConfigs: []simutil.ClassConfig{warrior},
			PlayerLevel:   lvl,
			EnemyConfigs:  []simutil.EnemyConfig{{"Hollow Knight", 6, 150, 13, 6, "boss"}},
			Iterations:    500,
		})
		fmt.Printf("  Warrior L%d vs HK(150/13/6): win=%5.1f%% avgRnds=%4.1f  P[ATK=%d DEF=%d HP=%d]\n",
			lvl, r.WinRate*100, r.AvgRounds, r.PlayerAttackPower, r.PlayerDefense, r.PlayerMaxHP)
	}

	// What-if: Reduce HK HP
	fmt.Println("\n--- WHAT-IF: Hollow Knight with reduced HP ---")
	for _, hp := range []int32{120, 100, 90, 80, 70} {
		r := simutil.RunMatchup(simutil.MatchupConfig{
			PlayerConfigs: []simutil.ClassConfig{warrior},
			PlayerLevel:   6,
			EnemyConfigs:  []simutil.EnemyConfig{{"Hollow Knight", 6, hp, 13, 6, "boss"}},
			Iterations:    500,
		})
		fmt.Printf("  Warrior L6 vs HK(HP=%3d, ATK=13, DEF=6): win=%5.1f%%  avgRnds=%4.1f\n",
			hp, r.WinRate*100, r.AvgRounds)
	}

	// What-if: Reduce HK DEF
	fmt.Println("\n--- WHAT-IF: Hollow Knight with reduced DEF ---")
	for _, def := range []int32{6, 5, 4, 3, 2} {
		r := simutil.RunMatchup(simutil.MatchupConfig{
			PlayerConfigs: []simutil.ClassConfig{warrior},
			PlayerLevel:   6,
			EnemyConfigs:  []simutil.EnemyConfig{{"Hollow Knight", 6, 150, 13, def, "boss"}},
			Iterations:    500,
		})
		fmt.Printf("  Warrior L6 vs HK(HP=150, ATK=13, DEF=%d): win=%5.1f%%  avgRnds=%4.1f\n",
			def, r.WinRate*100, r.AvgRounds)
	}

	// What-if: Reduce HK ATK
	fmt.Println("\n--- WHAT-IF: Hollow Knight with reduced ATK ---")
	for _, atk := range []int32{13, 11, 10, 9, 8} {
		r := simutil.RunMatchup(simutil.MatchupConfig{
			PlayerConfigs: []simutil.ClassConfig{warrior},
			PlayerLevel:   6,
			EnemyConfigs:  []simutil.EnemyConfig{{"Hollow Knight", 6, 150, atk, 6, "boss"}},
			Iterations:    500,
		})
		fmt.Printf("  Warrior L6 vs HK(HP=150, ATK=%2d, DEF=6): win=%5.1f%%  avgRnds=%4.1f\n",
			atk, r.WinRate*100, r.AvgRounds)
	}

	// Combined adjustments targeting 20-30% win rate at L6
	fmt.Println("\n--- COMBINED ADJUSTMENTS (targeting ~20-30% at L6) ---")
	combos := []struct{ hp, atk, def int32 }{
		{100, 10, 4}, {100, 11, 4}, {100, 10, 3},
		{90, 10, 4}, {90, 11, 3}, {80, 11, 4},
		{80, 10, 3}, {110, 10, 4}, {110, 11, 3}, {120, 10, 3},
	}
	for _, c := range combos {
		r := simutil.RunMatchup(simutil.MatchupConfig{
			PlayerConfigs: []simutil.ClassConfig{warrior},
			PlayerLevel:   6,
			EnemyConfigs:  []simutil.EnemyConfig{{"Hollow Knight", 6, c.hp, c.atk, c.def, "boss"}},
			Iterations:    500,
		})
		fmt.Printf("  Warrior L6 vs HK(HP=%3d, ATK=%2d, DEF=%d): win=%5.1f%%  avgRnds=%4.1f\n",
			c.hp, c.atk, c.def, r.WinRate*100, r.AvgRounds)
	}

	// Best candidates across multiple player levels
	fmt.Println("\n--- BEST CANDIDATES ACROSS LEVELS ---")
	bestCombos := []struct {
		hp, atk, def int32
		label        string
	}{
		{100, 10, 4, "Balanced"},
		{90, 11, 3, "Glass cannon"},
		{110, 10, 3, "Tanky but soft"},
	}
	for _, c := range bestCombos {
		fmt.Printf("\n  [%s] HK(HP=%d, ATK=%d, DEF=%d):\n", c.label, c.hp, c.atk, c.def)
		for _, lvl := range []int32{4, 5, 6, 7, 8, 10} {
			r := simutil.RunMatchup(simutil.MatchupConfig{
				PlayerConfigs: []simutil.ClassConfig{warrior},
				PlayerLevel:   lvl,
				EnemyConfigs:  []simutil.EnemyConfig{{"Hollow Knight", 6, c.hp, c.atk, c.def, "boss"}},
				Iterations:    500,
			})
			fmt.Printf("    L%2d: win=%5.1f%%  avgRnds=%4.1f\n", lvl, r.WinRate*100, r.AvgRounds)
		}
	}

	// L1 Warrior vs L2 enemies - is armor too strong?
	fmt.Println("\n--- L1 WARRIOR VS L2 ENEMIES (current DEF=8) ---")
	l2Enemies := []string{"Sewer Rat", "Tunnel Mole", "Meadow Wolf", "Alley Thug", "Wild Boar"}
	for _, name := range l2Enemies {
		enemy := simutil.EnemyConfigByName(name)
		r := simutil.RunMatchup(simutil.MatchupConfig{
			PlayerConfigs: []simutil.ClassConfig{warrior},
			PlayerLevel:   1,
			EnemyConfigs:  []simutil.EnemyConfig{*enemy},
			Iterations:    500,
		})
		fmt.Printf("  vs %s: win=%5.1f%%  avgRnds=%4.1f\n",
			name, r.WinRate*100, r.AvgRounds)
	}

	// What if Leather Armor = 4 instead of 8?
	fmt.Println("\n--- WHAT-IF: Leather Armor DEF=4 (instead of 8) ---")
	for _, name := range l2Enemies {
		enemy := simutil.EnemyConfigByName(name)
		r := runCustomMatchup(warrior, 1, *enemy, 5, 4, 500)
		fmt.Printf("  Warrior L1(DEF=4) vs %s: win=%5.1f%%  avgRnds=%4.1f\n",
			name, r.WinRate*100, r.AvgRounds)
	}

	fmt.Println("\n--- WHAT-IF: Leather Armor DEF=5 ---")
	for _, name := range l2Enemies {
		enemy := simutil.EnemyConfigByName(name)
		r := runCustomMatchup(warrior, 1, *enemy, 5, 5, 500)
		fmt.Printf("  Warrior L1(DEF=5) vs %s: win=%5.1f%%  avgRnds=%4.1f\n",
			name, r.WinRate*100, r.AvgRounds)
	}

	fmt.Println("\n--- WHAT-IF: Leather Armor DEF=3 ---")
	for _, name := range l2Enemies {
		enemy := simutil.EnemyConfigByName(name)
		r := runCustomMatchup(warrior, 1, *enemy, 5, 3, 500)
		fmt.Printf("  Warrior L1(DEF=3) vs %s: win=%5.1f%%  avgRnds=%4.1f\n",
			name, r.WinRate*100, r.AvgRounds)
	}

	// Check armor changes against full enemy range
	fmt.Println("\n--- ARMOR COMPARISON: DEF=8 vs DEF=4 across all enemies ---")
	for _, enemy := range simutil.AllEnemyConfigs() {
		r8 := simutil.RunMatchup(simutil.MatchupConfig{
			PlayerConfigs: []simutil.ClassConfig{warrior},
			PlayerLevel:   1,
			EnemyConfigs:  []simutil.EnemyConfig{enemy},
			Iterations:    500,
		})
		r4 := runCustomMatchup(warrior, 1, enemy, 5, 4, 500)
		fmt.Printf("  vs %-16s L%d:  DEF=8 win=%5.1f%%  |  DEF=4 win=%5.1f%%\n",
			enemy.Name, enemy.Level, r8.WinRate*100, r4.WinRate*100)
	}
}

// TestHollowKnightSweetSpot focuses on finding the right HK stats for ~20-30% at L6, ~5% at L5
func TestHollowKnightSweetSpot(t *testing.T) {
	warrior := *simutil.ClassConfigByName("Warrior")

	fmt.Println("\n========================================")
	fmt.Println("  HOLLOW KNIGHT SWEET SPOT SEARCH")
	fmt.Println("========================================")

	// Key insight from analysis: DEF is the problem. DEF=6 means warrior does only 4 dmg/hit.
	// We need DEF=3 or 4 so warrior does 6-7 dmg/hit, but keep HP high for boss feel.
	// ATK=13 kills too fast at low armor, ATK=10-11 gives player breathing room.
	fmt.Println("\n--- HIGH HP + LOW DEF + MODERATE ATK (boss should feel tanky, not unhittable) ---")
	combos := []struct{ hp, atk, def int32 }{
		// Keep DEF at 3-4 so warrior does 5-6 damage, keep HP high for long fights
		{130, 12, 3}, {130, 12, 4}, {130, 11, 3}, {130, 11, 4},
		{120, 12, 3}, {120, 12, 4}, {120, 11, 3}, {120, 11, 4},
		{140, 12, 3}, {140, 12, 4}, {140, 11, 3}, {140, 11, 4},
		{150, 12, 3}, {150, 12, 4}, {150, 11, 3},
	}

	fmt.Printf("\n  %-30s  %6s  %6s  %6s  %6s  %6s\n", "HK Stats", "L4", "L5", "L6", "L7", "L8")
	fmt.Println("  " + "----------------------------------------------------------------------")
	for _, c := range combos {
		label := fmt.Sprintf("HP=%3d ATK=%2d DEF=%d", c.hp, c.atk, c.def)
		fmt.Printf("  %-30s", label)
		for _, lvl := range []int32{4, 5, 6, 7, 8} {
			r := simutil.RunMatchup(simutil.MatchupConfig{
				PlayerConfigs: []simutil.ClassConfig{warrior},
				PlayerLevel:   lvl,
				EnemyConfigs:  []simutil.EnemyConfig{{"Hollow Knight", 6, c.hp, c.atk, c.def, "boss"}},
				Iterations:    500,
			})
			fmt.Printf("  %5.1f%%", r.WinRate*100)
		}
		fmt.Println()
	}

	// Also check the Thornback Bear and Burrow Brute
	fmt.Println("\n--- THORNBACK BEAR CURRENT (HP=80, ATK=11, DEF=4) ---")
	fmt.Printf("  %-30s  %6s  %6s  %6s  %6s  %6s\n", "Bear Stats", "L3", "L4", "L5", "L6", "L7")
	fmt.Println("  " + "----------------------------------------------------------------------")
	bearCombos := []struct{ hp, atk, def int32 }{
		{80, 11, 4},  // current
		{60, 9, 3},   // nerfed
		{70, 10, 3},  // slight nerf
		{50, 8, 2},   // big nerf
	}
	for _, c := range bearCombos {
		label := fmt.Sprintf("HP=%3d ATK=%2d DEF=%d", c.hp, c.atk, c.def)
		fmt.Printf("  %-30s", label)
		for _, lvl := range []int32{3, 4, 5, 6, 7} {
			r := simutil.RunMatchup(simutil.MatchupConfig{
				PlayerConfigs: []simutil.ClassConfig{warrior},
				PlayerLevel:   lvl,
				EnemyConfigs:  []simutil.EnemyConfig{{"Thornback Bear", 5, c.hp, c.atk, c.def, "hard"}},
				Iterations:    500,
			})
			fmt.Printf("  %5.1f%%", r.WinRate*100)
		}
		fmt.Println()
	}
}

// runCustomMatchup runs a matchup with custom weapon/armor stats
func runCustomMatchup(cls simutil.ClassConfig, level int32, enemy simutil.EnemyConfig, weaponDmg, armorDef int, iterations int) *simutil.SimResult {
	result := &simutil.SimResult{
		PlayerName:  cls.Name,
		PlayerLevel: level,
		EnemyName:   enemy.Name,
		EnemyLevel:  enemy.Level,
		Iterations:  iterations,
	}

	var totalRounds float64
	for i := 0; i < iterations; i++ {
		char := simutil.CreateCharacterWithGear(cls, level, weaponDmg, armorDef)
		npcEnemy := simutil.CreateEnemy(enemy)

		single := simutil.RunSimulation(
			[]*characters.Character{char},
			[]*npc.NPC{npcEnemy},
		)
		totalRounds += float64(single.Rounds)

		switch single.State {
		case "victory":
			result.PlayerWins++
		case "defeat":
			result.EnemyWins++
		default:
			result.Timeouts++
		}
	}

	result.WinRate = float64(result.PlayerWins) / float64(iterations)
	result.AvgRounds = totalRounds / float64(iterations)
	return result
}
