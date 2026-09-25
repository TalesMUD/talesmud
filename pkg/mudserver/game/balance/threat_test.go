package balance

import "testing"

func TestThreatTierDefaults(t *testing.T) {
	cfg := &CombatBalanceConfig{Threat: defaultThreat()}
	cases := []struct {
		player, enemy int32
		want          string
	}{
		{5, 1, ThreatGrey},   // -4
		{5, 2, ThreatGrey},   // -3
		{5, 3, ThreatGreen},  // -2
		{5, 4, ThreatGreen},  // -1
		{5, 5, ThreatYellow}, // 0
		{5, 6, ThreatYellow}, // +1
		{5, 7, ThreatOrange}, // +2
		{5, 8, ThreatRed},    // +3
		{5, 9, ThreatRed},    // +4
		{5, 10, ThreatSkull}, // +5
		{1, 20, ThreatSkull},
		{0, 2, ThreatYellow}, // unset viewer floors at level 1, so +1 is yellow
		{0, 6, ThreatSkull},
	}
	for _, tc := range cases {
		got := threatTier(cfg, tc.player, tc.enemy)
		if got != tc.want {
			t.Errorf("player %d enemy %d: got %s want %s", tc.player, tc.enemy, got, tc.want)
		}
	}
}

func TestThreatNeedsWarning(t *testing.T) {
	if ThreatNeedsWarning(ThreatYellow) || ThreatNeedsWarning(ThreatGreen) || ThreatNeedsWarning(ThreatGrey) {
		t.Fatal("even and lower tiers should not warn")
	}
	for _, tier := range []string{ThreatOrange, ThreatRed, ThreatSkull} {
		if !ThreatNeedsWarning(tier) {
			t.Fatalf("%s should warn", tier)
		}
	}
}

func TestThreatMissingSectionUsesDefaults(t *testing.T) {
	if threatTier(&CombatBalanceConfig{}, 1, 6) != ThreatSkull {
		t.Fatal("missing threat section should still skull a +5 gap")
	}
}

func TestThreatConfigFileWired(t *testing.T) {
	if _, err := ReloadConfig(); err != nil {
		t.Fatal(err)
	}
	cfg := GetConfig()
	if cfg.Threat.RedAtOrBelow != 4 || cfg.Threat.GreyAtOrBelow != -3 {
		t.Fatalf("yaml threat = %+v", cfg.Threat)
	}
	if ThreatTier(4, 9) != ThreatSkull {
		t.Fatalf("ThreatTier(4, 9) = %s", ThreatTier(4, 9))
	}
	if ThreatTier(4, 4) != ThreatYellow {
		t.Fatalf("equal level = %s", ThreatTier(4, 4))
	}
}
