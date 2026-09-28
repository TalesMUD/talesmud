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
