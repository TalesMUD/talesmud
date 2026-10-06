//go:debug randseednop=0

package combat_test

import (
	"math"
	"math/rand"
	"os"
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

	// The combat engine rolls the global math/rand source. Go 1.24 makes
	// rand.Seed a no-op unless randseednop=0, which this file sets, so each
	// row restarts at the same seed. A 5000-fight census put the warrior
	// at-level boss at 62.0% (band 60–90). The seeded 200-fight prefix of
	// this seed landed at 56%, outside the band; the 2000-fight prefix
	// landed at 61.2%, inside the band and under one standard error of
	// the census mean. The draw is deterministic, so the band is not widened.
	const iterations = 2000

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
	// Seed 20260927, 2000 fights. Slack is around the census mean, not one short draw.
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

func TestGapMatrixSeedRepeats(t *testing.T) {
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	prev := skills.AllSkills()
	skills.LoadFromDB(skills.SeedSkills())
	t.Cleanup(func() { skills.LoadFromDB(prev) })

	cls := simutil.ClassConfigByName("Warrior")
	if cls == nil {
		t.Fatal("missing warrior")
	}
	rand.Seed(20260927)
	a := simutil.RunScaledMatchup(*cls, 0, "boss", simutil.GearAppropriate, 200)
	rand.Seed(20260927)
	b := simutil.RunScaledMatchup(*cls, 0, "boss", simutil.GearAppropriate, 200)
	if a.PlayerWins != b.PlayerWins || a.AvgRounds != b.AvgRounds || a.EnemyWins != b.EnemyWins {
		t.Fatalf("seed did not repeat: wins %d/%d rounds %.3f/%.3f", a.PlayerWins, b.PlayerWins, a.AvgRounds, b.AvgRounds)
	}
}

// TestGapMatrixCensus measures each band against a large unseeded sample and
// against the seeded draw. It does not run in the default suite.
func TestGapMatrixCensus(t *testing.T) {
	if os.Getenv("GAP_CENSUS") == "" {
		t.Skip("set GAP_CENSUS=1 to measure true rates")
	}
	if _, err := balance.ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	prev := skills.AllSkills()
	skills.LoadFromDB(skills.SeedSkills())
	t.Cleanup(func() { skills.LoadFromDB(prev) })

	type spec struct {
		class, gear, tier string
		gap               int
		min, max          float64
	}
	rows := []spec{
		{"Warrior", simutil.GearAppropriate, "trash", 0, 0.90, 1},
		{"Warrior", simutil.GearAppropriate, "elite", 0, 0.70, 1},
		{"Warrior", simutil.GearAppropriate, "boss", 0, 0.60, 0.90},
		{"Warrior", simutil.GearAppropriate, "elite", 5, 0, 0.20},
		{"Warrior", simutil.GearAppropriate, "boss", 5, 0, 0.15},
		{"Warrior", simutil.GearGood, "boss", 3, 0.35, 0.80},
		{"Rogue", simutil.GearAppropriate, "trash", 0, 0.75, 1},
		{"Rogue", simutil.GearAppropriate, "boss", 0, 0.05, 0.30},
		{"Rogue", simutil.GearAppropriate, "boss", 5, 0, 0.20},
		{"Rogue", simutil.GearGood, "boss", 3, 0.05, 0.40},
		{"Ranger", simutil.GearAppropriate, "trash", 0, 0.85, 1},
		{"Ranger", simutil.GearAppropriate, "elite", 0, 0.55, 1},
		{"Ranger", simutil.GearAppropriate, "boss", 0, 0.05, 0.35},
		{"Ranger", simutil.GearAppropriate, "boss", 5, 0, 0.15},
		{"Ranger", simutil.GearGood, "boss", 3, 0.08, 0.40},
		{"Mage", simutil.GearAppropriate, "trash", 0, 0.30, 0.60},
		{"Mage", simutil.GearAppropriate, "elite", 0, 0.20, 0.70},
		{"Mage", simutil.GearAppropriate, "boss", 0, 0, 0.15},
		{"Mage", simutil.GearAppropriate, "boss", 5, 0, 0.20},
		{"Mage", simutil.GearGood, "boss", 3, 0, 0.20},
	}
	for _, row := range rows {
		cls := simutil.ClassConfigByName(row.class)
		if cls == nil {
			t.Fatalf("missing %s", row.class)
		}
		trueSim := simutil.RunScaledMatchup(*cls, row.gap, row.tier, row.gear, 5000)
		seeded := map[int]float64{}
		for _, n := range []int{200, 1000, 2000} {
			rand.Seed(20260927)
			sim := simutil.RunScaledMatchup(*cls, row.gap, row.tier, row.gear, n)
			seeded[n] = sim.WinRate
		}
		p := trueSim.WinRate
		se := math.Sqrt(p * (1 - p) / 5000)
		outside := p < row.min || p > row.max
		t.Logf("%s %s %s gap %+d true %.2f%% se5000 %.2fpp band %.0f–%.0f outside=%v seed200 %.1f%% seed1000 %.1f%% seed2000 %.1f%%",
			row.class, row.gear, row.tier, row.gap, p*100, se*100, row.min*100, row.max*100, outside,
			seeded[200]*100, seeded[1000]*100, seeded[2000]*100)
	}
}
