package combat_test

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/skills"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
	"github.com/talesmud/talesmud/pkg/mudserver/game/combat/simutil"
)

func TestGapMatrixTargets(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	prev := skills.AllSkills()
	skills.LoadFromDB(skills.SeedSkills())
	t.Cleanup(func() { skills.LoadFromDB(prev) })

	rows := simutil.RunGapMatrix(simutil.GapMatrixConfig{Iterations: 24})
	t.Log("\n" + simutil.FormatGapMarkdown(rows))

	// Bands are wide enough for a 24-iteration sample. Mage bosses sit lower
	// (cloth HP); the durable classes carry the at-level and +3 good-gear targets.
	check := func(class, gear, tier string, gap int, min, max float64) {
		t.Helper()
		for _, r := range rows {
			if r.Class == class && r.Gear == gear && r.Tier == tier && r.Gap == gap {
				if r.WinRate < min || r.WinRate > max {
					t.Errorf("%s %s %s gap %+d win %.0f%% outside %.0f–%.0f%%",
						class, gear, tier, gap, r.WinRate*100, min*100, max*100)
				}
				return
			}
		}
		t.Errorf("missing row %s %s %s gap %+d", class, gear, tier, gap)
	}

	check("Warrior", simutil.GearAppropriate, "trash", 0, 0.90, 1)
	check("Warrior", simutil.GearAppropriate, "elite", 0, 0.70, 1)
	check("Warrior", simutil.GearAppropriate, "boss", 0, 0.40, 0.95)
	check("Warrior", simutil.GearAppropriate, "elite", 5, 0, 0.20)
	check("Warrior", simutil.GearAppropriate, "boss", 5, 0, 0.15)
	check("Warrior", simutil.GearGood, "boss", 3, 0.35, 0.85)

	check("Rogue", simutil.GearAppropriate, "trash", 0, 0.75, 1)
	check("Rogue", simutil.GearAppropriate, "boss", 5, 0, 0.20)
	check("Rogue", simutil.GearGood, "boss", 3, 0.05, 0.70)

	check("Ranger", simutil.GearAppropriate, "trash", 0, 0.85, 1)
	check("Ranger", simutil.GearAppropriate, "elite", 0, 0.55, 1)
	check("Ranger", simutil.GearAppropriate, "boss", 5, 0, 0.15)
	check("Ranger", simutil.GearGood, "boss", 3, 0.30, 0.85)

	check("Mage", simutil.GearAppropriate, "trash", 0, 0.45, 1)
	check("Mage", simutil.GearAppropriate, "boss", 5, 0, 0.20)
}
