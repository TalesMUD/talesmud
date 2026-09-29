package balance

import "testing"

func TestBossTelegraphAndEnrage(t *testing.T) {
	configMu.Lock()
	prev := config
	config = &CombatBalanceConfig{BossMechanics: defaultBossMechanics()}
	configMu.Unlock()
	t.Cleanup(func() {
		configMu.Lock()
		config = prev
		configMu.Unlock()
	})

	if BossTelegraphTurns("boss", false) != 1 {
		t.Fatalf("boss telegraph turns = %d", BossTelegraphTurns("boss", false))
	}
	if BossTelegraphTurns("hard", false) != 1 {
		t.Fatalf("elite telegraph turns = %d", BossTelegraphTurns("hard", false))
	}
	if BossTelegraphTurns("easy", false) != 0 {
		t.Fatal("trash should not telegraph")
	}
	if BossTelegraphTurns("boss", true) != 0 {
		t.Fatal("enrage should skip the wind-up")
	}
	if !ShouldEnrage("boss", 16, 100, 100, false) {
		t.Fatal("round 16 should enrage a boss")
	}
	if ShouldEnrage("boss", 4, 100, 100, false) {
		t.Fatal("early full-hp boss should not enrage")
	}
	if !ShouldEnrage("boss", 2, 20, 100, false) {
		t.Fatal("30% hp or below should enrage")
	}
	if ShouldEnrage("hard", 20, 10, 100, false) {
		t.Fatal("elites telegraph but do not enrage")
	}
	if got := ScaleEnrageDamage(10); got != 12 {
		t.Fatalf("enrage 10 * 1.20 = %d, want 12", got)
	}
}

func TestBossPhaseConfigOverrides(t *testing.T) {
	original := config
	defer func() { config = original }()
	config = &CombatBalanceConfig{BossMechanics: defaultBossMechanics()}
	config.BossMechanics.Phases = []BossPhaseConfig{
		{Label: "Opening", BelowHP: 1},
		{Label: "Final", BelowHP: .5, TelegraphLabel: "Final Blow", DamageDealt: 1.5, EnrageDamage: 2},
	}
	if err := ValidateBossPhases(config.BossMechanics); err != nil {
		t.Fatal(err)
	}
	if len(BossPhases("hard")) != 0 || len(BossPhases("boss")) != 2 {
		t.Fatal("phases must default to boss-only")
	}
	if BossPhaseTelegraphLabel("boss", 2) != "Final Blow" || BossPhaseTelegraphLabel("boss", 1) != BossTelegraphLabel() {
		t.Fatal("telegraph inheritance failed")
	}
	if got := ScaleBossPhaseDamage(10, "boss", 2, false); got != 15 {
		t.Fatalf("phase damage: %d", got)
	}
	if got := ScaleBossPhaseDamage(10, "boss", 2, true); got != 30 {
		t.Fatalf("enrage override applied twice: %d", got)
	}
	if got := ScaleBossPhaseDamage(10, "boss", 1, true); got != 12 {
		t.Fatalf("global enrage inheritance: %d", got)
	}
	config.BossMechanics.PhaseTiers = []string{"elite"}
	if len(BossPhases("hard")) != 2 || len(BossPhases("boss")) != 0 {
		t.Fatal("explicit tier opt-in failed")
	}
	config.BossMechanics.Phases = nil
	if len(BossPhases("hard")) != 0 {
		t.Fatal("omitted phases must disable mechanic")
	}
}

func TestBossPhaseRejectInvalidConfig(t *testing.T) {
	for _, phases := range [][]BossPhaseConfig{
		{{Label: "Missing opening", BelowHP: .66}},
		{{Label: "Opening", BelowHP: 1}, {Label: "Duplicate", BelowHP: 1}},
		{{Label: "Opening", BelowHP: 1}, {Label: "", BelowHP: .5}},
		{{Label: "Opening", BelowHP: 1}, {Label: "Invalid", BelowHP: -.2}},
		{{Label: "Opening", BelowHP: 1}, {Label: "Invalid", BelowHP: .5, DamageDealt: -1}},
	} {
		if ValidateBossPhases(BossMechanicsConfig{Phases: phases}) == nil {
			t.Fatalf("accepted invalid phases: %+v", phases)
		}
	}
}
