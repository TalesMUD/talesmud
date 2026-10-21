package combat_test

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/skills"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
	"github.com/talesmud/talesmud/pkg/mudserver/game/combat/simutil"
)

func TestEarlyCurveProbe(t *testing.T) {
	if os.Getenv("EARLY_PROBE") == "" {
		t.Skip("set EARLY_PROBE=1 to print the early-game table")
	}
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	prev := skills.AllSkills()
	skills.LoadFromDB(skills.SeedSkills())
	t.Cleanup(func() { skills.LoadFromDB(prev) })

	iters := 40
	if v := os.Getenv("EARLY_ITERS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			t.Fatalf("EARLY_ITERS %q", v)
		}
		iters = n
	}
	only := os.Getenv("EARLY_ONLY")
	var rows []simutil.EarlyRow
	if only == "" {
		rows = simutil.RunEarlyTable(iters)
	} else {
		for _, preset := range characters.SystemCharacterTemplatePresets() {
			if preset == nil {
				continue
			}
			for _, level := range simutil.EarlyLevels() {
				for _, enemy := range simutil.EarlyPackEnemies() {
					if enemy.ID != only && enemy.Name != only {
						continue
					}
					rows = append(rows, simutil.RunEarlyCell(preset, level, enemy, iters))
				}
			}
		}
	}
	t.Logf("\n%s", simutil.FormatEarlyMarkdown(rows))

	if os.Getenv("EARLY_SKIP_BOUNTY") == "1" {
		return
	}
	wantBounty := map[string]bool{}
	if v := os.Getenv("EARLY_BOUNTY"); v != "" {
		for _, id := range strings.Split(v, ",") {
			wantBounty[strings.TrimSpace(id)] = true
		}
	}
	runBounty := func(id string) bool {
		return len(wantBounty) == 0 || wantBounty[id]
	}

	presets := characters.SystemCharacterTemplatePresets()
	thug := simutil.EarlyEnemy{ID: "ENM0007", Name: "Sample Thug", Level: 2, HP: 18, Attack: 4, Defense: 1, Difficulty: "normal"}
	spider := simutil.EarlyEnemy{ID: "ENM0014", Name: "Sample Spider", Level: 9, HP: 30, Attack: 8, Defense: 2, Difficulty: "normal"}
	toad := simutil.EarlyEnemy{ID: "ENM0015", Name: "Sample Toad", Level: 10, HP: 40, Attack: 6, Defense: 4, Difficulty: "normal"}
	brute := simutil.EarlyEnemy{ID: "ENM0022", Name: "Sample Raider", Level: 16, HP: 78, Attack: 14, Defense: 6, Difficulty: "hard"}
	for _, preset := range presets {
		if runBounty("0900") {
			one := simutil.RunBountyFight(preset, 1, thug, 3, true, "", iters)
			t.Logf("QST0900 L1 x3 rest %s win %.0f%% medRounds %.1f hp %.0f%% enemyHP %d",
				preset.Name, one.WinRate*100, one.MedRounds, one.MedHPLeft, one.EnemyHP)
			two := simutil.RunBountyFight(preset, 2, thug, 3, true, "", iters)
			t.Logf("QST0900 L2 x3 rest %s win %.0f%% medRounds %.1f hp %.0f%%",
				preset.Name, two.WinRate*100, two.MedRounds, two.MedHPLeft)
		}
		if runBounty("0907") {
			sp := simutil.RunBountyFight(preset, 5, spider, 1, true, simutil.GearAppropriate, iters)
			t.Logf("QST0907 L5 x1 %s win %.0f%% medRounds %.1f hp %.0f%% enemy %d/%d/%d",
				preset.Name, sp.WinRate*100, sp.MedRounds, sp.MedHPLeft, sp.EnemyHP, sp.EnemyATK, sp.EnemyDEF)
		}
		if runBounty("0908") {
			td := simutil.RunBountyFight(preset, 10, toad, 3, true, simutil.GearAppropriate, iters)
			t.Logf("QST0908 L10 x3 %s win %.0f%% medRounds %.1f hp %.0f%% enemy %d/%d/%d",
				preset.Name, td.WinRate*100, td.MedRounds, td.MedHPLeft, td.EnemyHP, td.EnemyATK, td.EnemyDEF)
		}
		if runBounty("0909") {
			br := simutil.RunBountyFight(preset, 15, brute, 2, true, simutil.GearAppropriate, iters)
			t.Logf("QST0909 L15 x2 %s win %.0f%% medRounds %.1f hp %.0f%% enemy %d/%d/%d",
				preset.Name, br.WinRate*100, br.MedRounds, br.MedHPLeft, br.EnemyHP, br.EnemyATK, br.EnemyDEF)
		}
	}
}

// TestEarlyCurveBands locks the starter curve. The seed is fixed inside
// RunEarlyCell, and the windows are wider than one unlucky fight so a
// one-count wobble does not fail the suite. Targets, measured at 80 iterations
// of seed 20261018: same-level Bandit is a short win for every class; a +2
// Bandit is about 65–82% (the rogue sits on a crit threshold at 82%, and the next
// integer of damage drops that cell to 50%).
func TestEarlyCurveBands(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	prev := skills.AllSkills()
	skills.LoadFromDB(skills.SeedSkills())
	t.Cleanup(func() { skills.LoadFromDB(prev) })

	const iters = 80
	var bandit simutil.EarlyEnemy
	for _, enemy := range simutil.EarlyPackEnemies() {
		if enemy.ID == "ENM0004" {
			bandit = enemy
		}
	}
	if bandit.ID == "" {
		t.Fatal("missing bandit")
	}
	thug := simutil.EarlyEnemy{ID: "ENM0007", Name: "Sample Thug", Level: 2, HP: 18, Attack: 4, Defense: 1, Difficulty: "normal"}
	spider := simutil.EarlyEnemy{ID: "ENM0014", Name: "Sample Spider", Level: 9, HP: 30, Attack: 8, Defense: 2, Difficulty: "normal"}

	var lo, hi float64 = 1, 0
	for _, preset := range characters.SystemCharacterTemplatePresets() {
		if preset == nil {
			continue
		}
		same := simutil.RunEarlyCell(preset, 3, bandit, iters)
		if same.EnemyHP != 24 || same.EnemyATK != 9 {
			t.Fatalf("%s bandit live %d/%d, want 24/9", preset.Name, same.EnemyHP, same.EnemyATK)
		}
		if same.WinRate < 0.90 {
			t.Fatalf("%s L3 bandit win %.2f", preset.Name, same.WinRate)
		}
		if same.MedRounds < 3 || same.MedRounds > 7 {
			t.Fatalf("%s L3 bandit rounds %.1f", preset.Name, same.MedRounds)
		}
		up := simutil.RunEarlyCell(preset, 1, bandit, iters)
		if up.WinRate < 0.55 || up.WinRate > 0.90 {
			t.Fatalf("%s L1 bandit +2 win %.2f", preset.Name, up.WinRate)
		}
		if up.WinRate < lo {
			lo = up.WinRate
		}
		if up.WinRate > hi {
			hi = up.WinRate
		}
		posting := simutil.RunBountyFight(preset, 1, thug, 3, true, "", iters)
		if posting.WinRate < 0.90 || posting.EnemyHP != 20 || posting.EnemyATK != 7 {
			t.Fatalf("QST0900 %s win %.2f enemy %d/%d", preset.Name, posting.WinRate, posting.EnemyHP, posting.EnemyATK)
		}
		hollow := simutil.RunBountyFight(preset, 5, spider, 1, true, simutil.GearAppropriate, iters)
		if hollow.WinRate < 0.55 || hollow.EnemyHP != 81 {
			t.Fatalf("QST0907 %s win %.2f enemyHP %d", preset.Name, hollow.WinRate, hollow.EnemyHP)
		}
	}
	if hi-lo > 0.25 {
		t.Fatalf("+2 bandit spread %.2f (%.2f–%.2f)", hi-lo, lo, hi)
	}
}
