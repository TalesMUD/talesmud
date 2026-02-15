package simutil

import (
	"fmt"
	"strings"
)

// FormatResultsTable prints a formatted ASCII table of simulation results
func FormatResultsTable(results []*SimResult) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("╔══════════╦═══════╦══════════════════╦═══════╦══════════╦═══════════╦════════════╦════════════╗\n")
	sb.WriteString("║ Class    ║ Level ║ Enemy            ║ E.Lvl ║ Win Rate ║ Avg Rnds  ║ P.ATK/DEF  ║ E.ATK/DEF  ║\n")
	sb.WriteString("╠══════════╬═══════╬══════════════════╬═══════╬══════════╬═══════════╬════════════╬════════════╣\n")

	for _, r := range results {
		winRateStr := fmt.Sprintf("%.1f%%", r.WinRate*100)
		marker := " "
		if r.WinRate < 0.30 {
			marker = "!"
		} else if r.WinRate > 0.90 {
			marker = "*"
		}

		sb.WriteString(fmt.Sprintf("║ %-8s ║   %2d  ║ %-16s ║   %2d  ║ %6s %s ║   %5.1f   ║  %3d/%-3d   ║  %3d/%-3d   ║\n",
			truncate(r.PlayerName, 8),
			r.PlayerLevel,
			truncate(r.EnemyName, 16),
			r.EnemyLevel,
			winRateStr, marker,
			r.AvgRounds,
			r.PlayerAttackPower, r.PlayerDefense,
			r.EnemyAttackPower, r.EnemyDefense,
		))
	}

	sb.WriteString("╚══════════╩═══════╩══════════════════╩═══════╩══════════╩═══════════╩════════════╩════════════╝\n")
	sb.WriteString("\n  * = dominant (>90%)  ! = concerning (<30%)\n")

	return sb.String()
}

// FormatStatTable prints a stat diagnostics table for characters at different levels
func FormatStatTable(levels []int32) string {
	var sb strings.Builder
	classes := AllClassConfigs()

	sb.WriteString("\n=== CHARACTER STATS DIAGNOSTICS ===\n\n")
	sb.WriteString(fmt.Sprintf("%-10s %-5s %6s %6s %6s %6s %6s %6s\n",
		"Class", "Level", "HP", "ATK", "DEF", "STR", "DEX", "STA"))
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	for _, cls := range classes {
		for _, lvl := range levels {
			char := CreateCharacter(cls, lvl)
			weaponDmg := char.GetWeaponDamage()
			atkMod := char.GetPrimaryAttackMod()
			atk := weaponDmg + int32(atkMod)
			if atk < 1 {
				atk = 1
			}
			def := char.GetArmorDefense()

			sb.WriteString(fmt.Sprintf("%-10s L%-4d %6d %6d %6d %6d %6d %6d\n",
				cls.Name,
				lvl,
				char.MaxHitPoints,
				atk,
				def,
				char.GetAttribute("STR"),
				char.GetAttribute("DEX"),
				char.GetAttribute("STA"),
			))
		}
	}

	return sb.String()
}

// FormatEnemyTable prints all enemy stats (with difficulty multipliers applied)
func FormatEnemyTable() string {
	var sb strings.Builder

	sb.WriteString("\n=== ENEMY STATS (with difficulty multipliers) ===\n\n")
	sb.WriteString(fmt.Sprintf("%-18s %5s %6s %6s %6s %10s\n",
		"Enemy", "Level", "HP", "ATK", "DEF", "Difficulty"))
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	for _, e := range AllEnemyConfigs() {
		// Create an actual enemy to get the multiplied stats
		enemy := CreateEnemy(e)
		sb.WriteString(fmt.Sprintf("%-18s %5d %6d %6d %6d %10s\n",
			e.Name, e.Level,
			enemy.MaxHitPoints,
			enemy.EnemyTrait.AttackPower,
			enemy.EnemyTrait.Defense,
			e.Difficulty))
	}

	return sb.String()
}

// FormatBalanceReport prints a summary of matchups that are out of expected range
func FormatBalanceReport(results []*SimResult) string {
	var sb strings.Builder

	sb.WriteString("\n=== BALANCE WARNINGS ===\n\n")

	warnings := 0
	for _, r := range results {
		levelDiff := r.PlayerLevel - r.EnemyLevel
		var expectedMin, expectedMax float64

		switch {
		case levelDiff >= 3:
			expectedMin, expectedMax = 0.85, 1.0
		case levelDiff >= 1:
			expectedMin, expectedMax = 0.65, 0.98
		case levelDiff == 0:
			expectedMin, expectedMax = 0.40, 0.80
		case levelDiff >= -2:
			expectedMin, expectedMax = 0.10, 0.50
		default:
			expectedMin, expectedMax = 0.0, 0.30
		}

		// Bosses should be harder
		if r.EnemyName == "Burrow Brute" || r.EnemyName == "Hollow Knight" || r.EnemyName == "Thornback Bear" {
			expectedMin -= 0.15
			expectedMax -= 0.10
		}

		if r.WinRate < expectedMin || r.WinRate > expectedMax {
			direction := "TOO HARD"
			if r.WinRate > expectedMax {
				direction = "TOO EASY"
			}
			sb.WriteString(fmt.Sprintf("  [%s] %s: %.1f%% win rate (expected %.0f-%.0f%%)\n",
				direction, r.Label, r.WinRate*100, expectedMin*100, expectedMax*100))
			sb.WriteString(fmt.Sprintf("         Player: ATK=%d DEF=%d HP=%d | Enemy: ATK=%d DEF=%d HP=%d\n",
				r.PlayerAttackPower, r.PlayerDefense, r.PlayerMaxHP,
				r.EnemyAttackPower, r.EnemyDefense, r.EnemyMaxHP))
			warnings++
		}
	}

	if warnings == 0 {
		sb.WriteString("  No balance issues detected! All matchups within expected ranges.\n")
	} else {
		sb.WriteString(fmt.Sprintf("\n  Total warnings: %d\n", warnings))
	}

	return sb.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "."
}
