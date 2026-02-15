package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/mudserver/game/combat/simutil"
)

func main() {
	iterations := flag.Int("n", 500, "Number of simulations per matchup")
	classFilter := flag.String("class", "", "Filter to specific class (e.g. Warrior)")
	enemyFilter := flag.String("enemy", "", "Filter to specific enemy (e.g. 'Meadow Wolf')")
	levelFilter := flag.Int("level", 0, "Filter to specific player level (0 = all)")
	statsOnly := flag.Bool("stats", false, "Only show stat diagnostics, no combat sims")
	flag.Parse()

	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║         TalesMUD Combat Simulator                ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")

	// Phase 1: Stat diagnostics
	levels := []int32{1, 5, 10, 15, 20}
	fmt.Print(simutil.FormatStatTable(levels))
	fmt.Print(simutil.FormatEnemyTable())

	if *statsOnly {
		return
	}

	// Phase 2: Build matchup configs
	fmt.Printf("\nRunning %d iterations per matchup...\n", *iterations)

	configs := simutil.BuildStandardMatchups(*iterations)

	// Apply filters
	if *classFilter != "" || *enemyFilter != "" || *levelFilter > 0 {
		filtered := make([]simutil.MatchupConfig, 0)
		for _, c := range configs {
			if *classFilter != "" && !strings.EqualFold(c.PlayerConfigs[0].Name, *classFilter) {
				continue
			}
			if *enemyFilter != "" && !strings.EqualFold(c.EnemyConfigs[0].Name, *enemyFilter) {
				continue
			}
			if *levelFilter > 0 && c.PlayerLevel != int32(*levelFilter) {
				continue
			}
			filtered = append(filtered, c)
		}
		configs = filtered
	}

	fmt.Printf("Running %d matchup configurations...\n", len(configs))

	// Phase 3: Run simulations
	results := simutil.RunAllMatchups(configs)

	// Phase 4: Output
	fmt.Print(simutil.FormatResultsTable(results))
	fmt.Print(simutil.FormatBalanceReport(results))

	// Summary
	var totalWins, totalLosses, totalTimeouts int
	for _, r := range results {
		totalWins += r.PlayerWins
		totalLosses += r.EnemyWins
		totalTimeouts += r.Timeouts
	}
	total := totalWins + totalLosses + totalTimeouts
	if total > 0 {
		fmt.Printf("\n=== SUMMARY ===\n")
		fmt.Printf("  Total simulations: %d\n", total)
		fmt.Printf("  Player wins:       %d (%.1f%%)\n", totalWins, float64(totalWins)/float64(total)*100)
		fmt.Printf("  Enemy wins:        %d (%.1f%%)\n", totalLosses, float64(totalLosses)/float64(total)*100)
		fmt.Printf("  Timeouts:          %d (%.1f%%)\n", totalTimeouts, float64(totalTimeouts)/float64(total)*100)
	}
}
