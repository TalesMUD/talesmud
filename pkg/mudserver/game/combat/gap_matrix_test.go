package combat_test

import (
	"math/rand"
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

	// The combat engine rolls the global math/rand source. A 24-fight draw
	// of mage at-level bosses has crossed the top of the band. Each row
	// restarts the same seed and takes 200 fights, so the rate does not
	// depend on which row ran before it.
	const iterations = 200

	check := func(class, gear, tier string, gap int, min, max float64) {
		t.Helper()
		rand.Seed(20260927)
		cls := simutil.ClassConfigByName(class)
		if cls == nil {
			t.Fatalf("missing class %s", class)
		}
		sim := simutil.RunScaledMatchup(*cls, gap, tier, gear, iterations)
		t.Logf("%s %s %s gap %+d win %.1f%% over %d", class, gear, tier, gap, sim.WinRate*100, iterations)
		if sim.WinRate < min || sim.WinRate > max {
			t.Errorf("%s %s %s gap %+d win %.0f%% outside %.0f–%.0f%%",
				class, gear, tier, gap, sim.WinRate*100, min*100, max*100)
		}
	}

	// Bands follow the signed replacement (behind cap 1.15, no rogue 2.35, no mage 0.46).
	// Seed 20260927, 200 fights. Slack is around that draw, not the old tuning.
	check("Warrior", simutil.GearAppropriate, "trash", 0, 0.90, 1)
	check("Warrior", simutil.GearAppropriate, "elite", 0, 0.70, 1)
	check("Warrior", simutil.GearAppropriate, "boss", 0, 0.60, 0.90)
	check("Warrior", simutil.GearAppropriate, "elite", 5, 0, 0.20)
	check("Warrior", simutil.GearAppropriate, "boss", 5, 0, 0.15)
	check("Warrior", simutil.GearGood, "boss", 3, 0.35, 0.80)

	check("Rogue", simutil.GearAppropriate, "trash", 0, 0.75, 1)
	check("Rogue", simutil.GearAppropriate, "boss", 0, 0.05, 0.30)
	check("Rogue", simutil.GearAppropriate, "boss", 5, 0, 0.20)
	check("Rogue", simutil.GearGood, "boss", 3, 0.05, 0.40) // Slip is a button, not a free miss

	check("Ranger", simutil.GearAppropriate, "trash", 0, 0.85, 1)
	check("Ranger", simutil.GearAppropriate, "elite", 0, 0.55, 1)
	check("Ranger", simutil.GearAppropriate, "boss", 0, 0.05, 0.35)
	check("Ranger", simutil.GearAppropriate, "boss", 5, 0, 0.15)
	check("Ranger", simutil.GearGood, "boss", 3, 0.08, 0.40)

	check("Mage", simutil.GearAppropriate, "trash", 0, 0.30, 0.60)
	check("Mage", simutil.GearAppropriate, "elite", 0, 0.20, 0.70) // Inscribe is a button, not a free burn
	check("Mage", simutil.GearAppropriate, "boss", 0, 0, 0.15)
	check("Mage", simutil.GearAppropriate, "boss", 5, 0, 0.20)
	check("Mage", simutil.GearGood, "boss", 3, 0, 0.20)
}
