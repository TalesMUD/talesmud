package balance

import "testing"

func TestScaleClassDamageIdentity(t *testing.T) {
	cfg := &CombatBalanceConfig{ClassBalance: defaultClassBalance()}
	configMu.Lock()
	prev := config
	config = cfg
	configMu.Unlock()
	t.Cleanup(func() {
		configMu.Lock()
		config = prev
		configMu.Unlock()
	})

	if got := ScaleClassDamage("", "", 10, 10, 10); got != 10 {
		t.Fatalf("empty class damage = %d", got)
	}
	if got := ScaleClassDamage("cleric", "", 10, 10, 10); got != 10 {
		t.Fatalf("unset class should be identity, got %d", got)
	}
	if got := ScaleClassDamage("warrior", "", 10, 10, 10); got != 10 {
		t.Fatalf("even warrior should be identity, got %d", got)
	}
	if got := ScaleClassDamage("warrior", "", 10, 13, 10); got != 12 {
		t.Fatalf("warrior behind 10 * 1.20 = %d, want 12", got)
	}
	if got := ScaleClassDamage("rogue", "", 10, 10, 10); got != 14 {
		t.Fatalf("rogue even 10 * 1.35 = %d, want 14", got)
	}
	if got := ScaleClassDamage("rogue", "", 10, 13, 10); got != 32 {
		t.Fatalf("rogue behind 10 * 1.35 * 2.35 = %d, want 32", got)
	}
	if got := ScaleClassDamage("ranger", "", 10, 10, 10); got != 13 {
		t.Fatalf("ranger 10 * 1.26 = %d, want 13", got)
	}
	if got := ScaleClassDamage("", "wizard", 12, 10, 20); got != 9 {
		t.Fatalf("mage taken 20 * 0.46 = %d, want 9", got)
	}
	if got := ScaleClassDamage("mage", "", 10, 10, 10); got != 27 {
		t.Fatalf("mage dealt 10 * 2.65 = %d, want 27", got)
	}
}
