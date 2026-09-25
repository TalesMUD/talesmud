package simutil

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	combatentity "github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
)

// Gap anchor: a mid-level character. Enemy level is AnchorLevel + gap.
// Gap is enemyLevel - playerLevel, matching threat colors (-3 trivial, +5 skull).
const GapAnchorLevel int32 = 10

// GearAppropriate is starter gear grown slightly with level.
// GearGood is that kit scaled up so it can close about two or three levels.
const (
	GearAppropriate = "appropriate"
	GearGood        = "good"
)

// GapRow is one cell of the balance table.
type GapRow struct {
	Class       string
	Gear        string
	Tier        string
	Gap         int
	PlayerLevel int32
	EnemyLevel  int32
	Iterations  int
	WinRate     float64
	AvgRounds   float64
	HPLeft      float64 // average remaining player HP% on wins; 0 if no wins
	PlayerATK   int32
	PlayerDEF   int32
	PlayerHP    int32
	EnemyATK    int32
	EnemyDEF    int32
	EnemyHP     int32
}

// GapMatrixConfig selects which slice of the table to run.
type GapMatrixConfig struct {
	Iterations int
	Classes    []string // empty = warrior, rogue, ranger, mage
	Gears      []string // empty = appropriate and good
	Tiers      []string // empty = trash, elite, boss
	Gaps       []int    // empty = -3..+5
}

// DefaultGapClasses are the classes the balance table covers.
func DefaultGapClasses() []string {
	return []string{"Warrior", "Rogue", "Ranger", "Mage"}
}

func tierDifficulty(tier string) string {
	switch tier {
	case "trash":
		return "easy"
	case "elite":
		return "hard"
	case "boss":
		return "boss"
	default:
		return "normal"
	}
}

// GearScores returns weapon damage and armor for a class at a level.
// Appropriate gear is the class starter plus a small per-level bump.
// Good gear is about 1.6× that, enough to close roughly two or three levels.
func GearScores(cfg ClassConfig, level int32, gear string) (weapon, armor int) {
	base := CreateCharacter(cfg, 1)
	weapon = int(base.GetWeaponDamage()) + int(level)/2
	armor = int(base.GetArmorDefense()) + int(level)/3
	if weapon < 1 {
		weapon = 1
	}
	if gear == GearGood {
		// About 2.2× level-appropriate weapon and armor: enough to offset
		// roughly three levels of the configured level-gap modifiers.
		weapon = int(float64(weapon)*2.2 + 0.5)
		armor = int(float64(armor)*2.2 + 0.5)
	}
	if weapon < 1 {
		weapon = 1
	}
	return weapon, armor
}

// RunScaledMatchup fights one class at GapAnchorLevel against a level-scaled
// trash, elite, or boss at anchor+gap.
func RunScaledMatchup(class ClassConfig, gap int, tier, gear string, iterations int) *SimResult {
	if iterations <= 0 {
		iterations = 40
	}
	playerLevel := GapAnchorLevel
	enemyLevel := playerLevel + int32(gap)
	if enemyLevel < 1 {
		enemyLevel = 1
	}
	diff := tierDifficulty(tier)
	weapon, armor := GearScores(class, playerLevel, gear)

	result := &SimResult{
		PlayerName:  class.Name,
		PlayerLevel: playerLevel,
		EnemyName:   tier,
		EnemyLevel:  enemyLevel,
		Iterations:  iterations,
		Label:       fmt.Sprintf("%s L%d %s vs %s %+d (L%d)", class.Name, playerLevel, gear, tier, gap, enemyLevel),
	}

	sample := CreateCharacterWithGear(class, playerLevel, weapon, armor)
	result.PlayerAttackPower = sample.GetWeaponDamage() + int32(sample.GetPrimaryAttackMod())
	if result.PlayerAttackPower < 1 {
		result.PlayerAttackPower = 1
	}
	result.PlayerDefense = sample.GetArmorDefense()
	result.PlayerMaxHP = sample.MaxHitPoints
	result.PlayerSTRMod = sample.GetPrimaryAttackMod()

	sampleEnemy := CreateScaledEnemy(tier, enemyLevel, diff)
	result.EnemyAttackPower = sampleEnemy.EnemyTrait.AttackPower
	result.EnemyDefense = sampleEnemy.EnemyTrait.Defense
	result.EnemyMaxHP = sampleEnemy.MaxHitPoints

	var totalRounds float64
	var totalPlayerHP float64
	var wins int
	for i := 0; i < iterations; i++ {
		char := CreateCharacterWithGear(class, playerLevel, weapon, armor)
		enemy := CreateScaledEnemy(tier, enemyLevel, diff)
		single := RunSimulation([]*characters.Character{char}, []*npc.NPC{enemy})
		totalRounds += float64(single.Rounds)
		switch single.State {
		case combatentity.CombatStateVictory:
			result.PlayerWins++
			wins++
			if single.PlayerHPMax > 0 {
				totalPlayerHP += float64(single.PlayerHPEnd) / float64(single.PlayerHPMax) * 100
			}
		case combatentity.CombatStateDefeat:
			result.EnemyWins++
		default:
			result.Timeouts++
		}
	}
	result.WinRate = float64(result.PlayerWins) / float64(iterations)
	result.AvgRounds = totalRounds / float64(iterations)
	if wins > 0 {
		result.AvgPlayerHPPercent = totalPlayerHP / float64(wins)
	}
	return result
}

// RunGapMatrix runs the class × gear × tier × gap table.
func RunGapMatrix(cfg GapMatrixConfig) []GapRow {
	classes := cfg.Classes
	if len(classes) == 0 {
		classes = DefaultGapClasses()
	}
	gears := cfg.Gears
	if len(gears) == 0 {
		gears = []string{GearAppropriate, GearGood}
	}
	tiers := cfg.Tiers
	if len(tiers) == 0 {
		tiers = []string{"trash", "elite", "boss"}
	}
	gaps := cfg.Gaps
	if len(gaps) == 0 {
		gaps = []int{-3, -2, -1, 0, 1, 2, 3, 4, 5}
	}
	iters := cfg.Iterations
	if iters <= 0 {
		iters = 40
	}

	var rows []GapRow
	for _, name := range classes {
		cls := ClassConfigByName(name)
		if cls == nil {
			continue
		}
		for _, gear := range gears {
			for _, tier := range tiers {
				for _, gap := range gaps {
					sim := RunScaledMatchup(*cls, gap, tier, gear, iters)
					rows = append(rows, GapRow{
						Class:       name,
						Gear:        gear,
						Tier:        tier,
						Gap:         gap,
						PlayerLevel: sim.PlayerLevel,
						EnemyLevel:  sim.EnemyLevel,
						Iterations:  iters,
						WinRate:     sim.WinRate,
						AvgRounds:   sim.AvgRounds,
						HPLeft:      sim.AvgPlayerHPPercent,
						PlayerATK:   sim.PlayerAttackPower,
						PlayerDEF:   sim.PlayerDefense,
						PlayerHP:    sim.PlayerMaxHP,
						EnemyATK:    sim.EnemyAttackPower,
						EnemyDEF:    sim.EnemyDefense,
						EnemyHP:     sim.EnemyMaxHP,
					})
				}
			}
		}
	}
	return rows
}

// FormatGapMarkdown renders the matrix as one table per class and gear.
func FormatGapMarkdown(rows []GapRow) string {
	var b strings.Builder
	b.WriteString("# Combat balance — level gap table\n\n")
	b.WriteString("Player level is 10. Gap is enemy level minus player level. ")
	b.WriteString("Trash / elite / boss use the level-scaled bodies in `CreateScaledEnemy` ")
	b.WriteString("(easy / hard / boss). Appropriate gear is class starter plus a small per-level bump. ")
	b.WriteString("Good gear is about 2.2× that. Hit, crit, and the level gap come from `config/combat_balance.yaml` `level_gap`. ")
	b.WriteString("`class_balance` then scales each class's damage dealt and taken.\n\n")
	b.WriteString("Targets: level-appropriate trash at gap 0 wins about 95%+, elites 75–90%, bosses 50–65%. ")
	b.WriteString("Bosses at +3 with good gear about 50%. +5 with appropriate gear stays under 15%.\n\n")

	type key struct{ class, gear string }
	var order []key
	seen := map[key]bool{}
	for _, row := range rows {
		k := key{row.Class, row.Gear}
		if !seen[k] {
			seen[k] = true
			order = append(order, k)
		}
	}
	gaps := []int{-3, -2, -1, 0, 1, 2, 3, 4, 5}
	for _, k := range order {
		fmt.Fprintf(&b, "## %s — %s gear\n\n", k.class, k.gear)
		b.WriteString("| Tier | Stat |")
		for _, g := range gaps {
			fmt.Fprintf(&b, " %+d |", g)
		}
		b.WriteString("\n| --- | --- |")
		for range gaps {
			b.WriteString(" --- |")
		}
		b.WriteString("\n")
		for _, tier := range []string{"trash", "elite", "boss"} {
			writeGapLine(&b, rows, k.class, k.gear, tier, "win%", gaps, func(r GapRow) string {
				return fmt.Sprintf(" %.0f%%", r.WinRate*100)
			})
			writeGapLine(&b, rows, k.class, k.gear, tier, "rounds", gaps, func(r GapRow) string {
				return fmt.Sprintf(" %.1f", r.AvgRounds)
			})
			writeGapLine(&b, rows, k.class, k.gear, tier, "HP left", gaps, func(r GapRow) string {
				if r.WinRate == 0 {
					return " —"
				}
				return fmt.Sprintf(" %.0f%%", r.HPLeft)
			})
		}
		b.WriteString("\n")
	}
	return b.String()
}

func writeGapLine(b *strings.Builder, rows []GapRow, class, gear, tier, stat string, gaps []int, cell func(GapRow) string) {
	fmt.Fprintf(b, "| %s | %s |", tier, stat)
	for _, g := range gaps {
		found := false
		for _, r := range rows {
			if r.Class == class && r.Gear == gear && r.Tier == tier && r.Gap == g {
				b.WriteString(cell(r))
				b.WriteString(" |")
				found = true
				break
			}
		}
		if !found {
			b.WriteString("  |")
		}
	}
	b.WriteString("\n")
}
