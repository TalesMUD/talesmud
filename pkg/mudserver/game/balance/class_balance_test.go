package balance

import "testing"

func TestScaleClassDamageRoster(t *testing.T) {
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
		t.Fatalf("fenwatch even 10 * 1.00 = %d, want 10", got)
	}
	if got := ScaleClassDamage("fenwatch", "", 10, 13, 10); got != 12 {
		t.Fatalf("fenwatch behind 10 * 1.00 * 1.15 = %d, want 12", got)
	}
	if got := ScaleClassDamage("rogue", "", 10, 10, 10); got != 6 {
		t.Fatalf("alley even 10 * 0.55 = %d, want 6", got)
	}
	if got := ScaleClassDamage("alley", "", 10, 13, 10); got != 6 {
		t.Fatalf("alley behind 10 * 0.55 * 1.15 = %d, want 6", got)
	}
	if got := ScaleClassDamage("ranger", "", 10, 10, 10); got != 6 {
		t.Fatalf("ranger folds into alley, got %d want 6", got)
	}
	if got := ScaleClassDamage("hunter", "", 10, 10, 10); got != 6 {
		t.Fatalf("hunter folds into alley, got %d want 6", got)
	}
	if got := ScaleClassDamage("", "wizard", 12, 10, 20); got != 25 {
		t.Fatalf("rune hand taken 20 * 1.25 = %d, want 25", got)
	}
	if got := ScaleClassDamage("mage", "", 10, 10, 10); got != 14 {
		t.Fatalf("rune hand dealt 10 * 1.40 = %d, want 14", got)
	}
	if got := ScaleClassDamage("hitch", "", 10, 10, 10); got != 9 {
		t.Fatalf("hitch dealt 10 * 0.90 = %d, want 9", got)
	}
	if got := ScaleClassDamage("", "hitch", 10, 10, 20); got != 21 {
		t.Fatalf("hitch taken 20 * 1.05 = %d, want 21", got)
	}
	if got := ScaleClassDamage("warrior", "rogue", 10, 10, 10); got != 12 {
		t.Fatalf("fenwatch into alley 10 * 1.00 * 1.15 = %d, want 12", got)
	}

	// A stale overtuned behind value must clamp, not stack.
	cfg.ClassBalance["rogue"] = ClassBalance{DamageDealt: 0.55, DamageTaken: 1.15, BehindDealt: 2.35, Swings: 2}
	if got := ScaleClassDamage("rogue", "", 10, 13, 10); got != 6 {
		t.Fatalf("behind cap 10 * 0.55 * 1.15 = %d, want 6 (not 2.35)", got)
	}
	cfg.ClassBalance["mage"] = ClassBalance{DamageDealt: 1.40, DamageTaken: 0.46, BehindDealt: 1, Swings: 1}
	if got := ScaleClassDamage("", "wizard", 12, 10, 20); got != 9 {
		t.Fatalf("explicit 0.46 taken still applies when under the behind cap, got %d", got)
	}
}

func TestClassRosterShape(t *testing.T) {
	if ClassSwings("rogue") != 2 || ClassSwings("warrior") != 1 || ClassSwings("hitch") != 1 || ClassSwings("wizard") != 1 {
		t.Fatalf("swings rogue=%d warrior=%d hitch=%d wizard=%d", ClassSwings("rogue"), ClassSwings("warrior"), ClassSwings("hitch"), ClassSwings("wizard"))
	}
	if got := ScaleClassHP("warrior", 25); got != 30 {
		t.Fatalf("fenwatch hp %d", got)
	}
	if got := ScaleClassHP("rogue", 25); got != 21 {
		t.Fatalf("alley hp %d", got)
	}
	if got := ScaleClassHP("wizard", 25); got != 19 {
		t.Fatalf("rune hand hp %d", got)
	}
	if got := ScaleClassHP("hitch", 25); got != 26 {
		t.Fatalf("hitch hp %d", got)
	}
	b, s, p := SignatureCharges("warrior")
	if b != 1 || s != 0 || p != 0 {
		t.Fatalf("brace charges %d %d %d", b, s, p)
	}
	b, s, p = SignatureCharges("hunter")
	if b != 0 || s != 1 || p != 0 {
		t.Fatalf("hunter should slip with alley, got %d %d %d", b, s, p)
	}
	b, s, p = SignatureCharges("hitch")
	if b != 0 || s != 0 || p != 1 {
		t.Fatalf("pin charges %d %d %d", b, s, p)
	}
	if !IsRuneHand("wizard") || IsRuneHand("hitch") {
		t.Fatal("rune hand detect")
	}
	for id, row := range defaultClassBalance() {
		if row.BehindDealt > BehindDealtCap {
			t.Fatalf("%s behind %v over cap", id, row.BehindDealt)
		}
		if id == "mage" && row.DamageTaken < 1 {
			t.Fatalf("mage taken %v still the old cloth discount", row.DamageTaken)
		}
		if id == "rogue" && row.BehindDealt > BehindDealtCap {
			t.Fatalf("rogue behind %v", row.BehindDealt)
		}
	}
}
