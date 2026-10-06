package balance

import (
	"math"
	"testing"

	"github.com/talesmud/talesmud/pkg/classkit"
)

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
	if got := ScaleClassDamage("hitch", "", 20, 20, 20); got != 19 {
		t.Fatalf("ward dealt 20 * 0.95 = %d, want 19", got)
	}
	if got := ScaleClassDamage("ward", "", 20, 20, 20); got != 19 {
		t.Fatalf("ward id dealt 20 * 0.95 = %d, want 19", got)
	}
	if got := ScaleClassDamage("", "hitch", 10, 10, 20); got != 21 {
		t.Fatalf("ward taken 20 * 1.05 = %d, want 21", got)
	}
	if got := ScaleClassDamage("", "ward", 10, 10, 20); got != 21 {
		t.Fatalf("ward id taken 20 * 1.05 = %d, want 21", got)
	}
	if got := ScaleClassDamage("warrior", "rogue", 10, 10, 10); got != 12 {
		t.Fatalf("fenwatch into alley 10 * 1.00 * 1.15 = %d, want 12", got)
	}
	if got := ScaleClassDamage("rigger", "", 10, 10, 20); got != 17 {
		t.Fatalf("rigger dealt 20 * 0.85 = %d, want 17", got)
	}
	if got := ScaleClassDamage("", "rigger", 10, 10, 20); got != 20 {
		t.Fatalf("rigger taken 20 * 1.00 = %d, want 20", got)
	}
	if got := ScaleClassDamage("rigger", "", 10, 13, 20); got != 20 {
		t.Fatalf("rigger behind 20 * 0.85 * 1.15 = %d, want 20", got)
	}
	if got := ScaleClassDamage("warrior", "", 10, 10, 20); got == ScaleClassDamage("rigger", "", 10, 10, 20) {
		t.Fatalf("rigger damage should not match warrior")
	}

	// A stale overtuned behind value must clamp, not stack.
	// The catalog wins over the config map, so the mutation goes on the catalog.
	origRogue, _ := classkit.Balance("rogue")
	origMage, _ := classkit.Balance("mage")
	t.Cleanup(func() {
		classkit.SetBalance("rogue", origRogue)
		classkit.SetBalance("mage", origMage)
	})
	classkit.SetBalance("rogue", classkit.Row{DamageDealt: 0.55, DamageTaken: 1.15, BehindDealt: 2.35, Swings: 2})
	if got := ScaleClassDamage("rogue", "", 10, 13, 10); got != 6 {
		t.Fatalf("behind cap 10 * 0.55 * 1.15 = %d, want 6 (not 2.35)", got)
	}
	classkit.SetBalance("mage", classkit.Row{DamageDealt: 1.40, DamageTaken: 0.46, BehindDealt: 1, Swings: 1})
	if got := ScaleClassDamage("", "wizard", 12, 10, 20); got != 9 {
		t.Fatalf("explicit 0.46 taken still applies when under the behind cap, got %d", got)
	}
}

func TestClassRosterShape(t *testing.T) {
	if ClassSwings("rogue") != 2 || ClassSwings("warrior") != 1 || ClassSwings("hitch") != 1 || ClassSwings("ward") != 1 || ClassSwings("wizard") != 1 {
		t.Fatalf("swings rogue=%d warrior=%d hitch=%d ward=%d wizard=%d", ClassSwings("rogue"), ClassSwings("warrior"), ClassSwings("hitch"), ClassSwings("ward"), ClassSwings("wizard"))
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
		t.Fatalf("ward hp via hitch %d", got)
	}
	if got := ScaleClassHP("ward", 25); got != 26 {
		t.Fatalf("ward hp %d", got)
	}
	if ClassSwings("rigger") != 1 {
		t.Fatalf("rigger swings %d", ClassSwings("rigger"))
	}
	if got := ScaleClassHP("rigger", 25); got != 25 {
		t.Fatalf("rigger hp 25 -> %d", got)
	}
	if got := ScaleClassHP("rigger", 40); got != 40 {
		t.Fatalf("rigger hp 40 -> %d", got)
	}
	b, s, p := SignatureCharges("rigger")
	if b != 0 || s != 0 || p != 0 {
		t.Fatalf("rigger is not brace/slip/pin, got %d %d %d", b, s, p)
	}
	bolt, rig := BoltRigCharges("rigger")
	if bolt != 1 || rig != 1 {
		t.Fatalf("rigger charges %d %d", bolt, rig)
	}
	if bolt, rig := BoltRigCharges("hitch"); bolt != 0 || rig != 0 {
		t.Fatalf("hitch rigger charges %d %d", bolt, rig)
	}
	if !ArmsScrap("rigger") || ArmsScrap("warrior") {
		t.Fatal("rigger detect")
	}
	b, s, p = SignatureCharges("warrior")
	if b != 1 || s != 0 || p != 0 {
		t.Fatalf("brace charges %d %d %d", b, s, p)
	}
	b, s, p = SignatureCharges("hunter")
	if b != 0 || s != 1 || p != 0 {
		t.Fatalf("hunter should slip with alley, got %d %d %d", b, s, p)
	}
	b, s, p = SignatureCharges("hitch")
	if b != 0 || s != 0 || p != 0 {
		t.Fatalf("ward has no brace/slip/pin charges, got %d %d %d", b, s, p)
	}
	b, s, p = SignatureCharges("ward")
	if b != 0 || s != 0 || p != 0 {
		t.Fatalf("ward id charges %d %d %d", b, s, p)
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
		if id == "rigger" && (row.DamageDealt != 0.85 || row.DamageTaken != 1 || row.Swings != 1 || row.BehindDealt != BehindDealtCap) {
			t.Fatalf("rigger row %+v", row)
		}
		if id == "ward" && (row.DamageDealt != 0.95 || row.DamageTaken != 1.05 || row.Swings != 1 || row.BehindDealt != BehindDealtCap) {
			t.Fatalf("ward row %+v", row)
		}
	}
	if _, ok := defaultClassBalance()["hitch"]; ok {
		t.Fatal("default balance still has a hitch row")
	}
}

func TestWardSoakNumbers(t *testing.T) {
	if WardOpeningGrit != 1 || WardOpeningGrit >= GritCap || WardStarterSwing != 1 {
		t.Fatalf("opening grit %d starter %d", WardOpeningGrit, WardStarterSwing)
	}
	if WardSlamAbsolute(0) != 1 || WardSlamAbsolute(2) != 1.4 || WardSlamAbsolute(5) != 2 || WardSlamAbsolute(9) != 2 {
		t.Fatalf("slam absolute %v %v %v %v", WardSlamAbsolute(0), WardSlamAbsolute(2), WardSlamAbsolute(5), WardSlamAbsolute(9))
	}
	near := func(got, want float64) {
		t.Helper()
		if math.Abs(got-want) > 1e-9 {
			t.Fatalf("got %v want %v", got, want)
		}
	}
	near(WardSlamSwingMult(0), 1.0/0.95)
	near(WardSlamSwingMult(2), 1.4/0.95)
	near(WardSlamSwingMult(5), 2.0/0.95)
	if WardRetaliateDamage(20, 0) != 0 || WardRetaliateDamage(1, 1) != 0 || WardRetaliateDamage(20, 1) != 2 || WardRetaliateDamage(20, 5) != 10 {
		t.Fatalf("retaliate %d %d %d %d", WardRetaliateDamage(20, 0), WardRetaliateDamage(1, 1), WardRetaliateDamage(20, 1), WardRetaliateDamage(20, 5))
	}
	if WardGritAfter(0, false) != 1 || WardGritAfter(0, true) != 2 || WardGritAfter(5, false) != 5 || WardGritAfter(4, true) != 5 {
		t.Fatalf("grit %d %d %d %d", WardGritAfter(0, false), WardGritAfter(0, true), WardGritAfter(5, false), WardGritAfter(4, true))
	}
	if !IsWard("hitch") || !IsWard("ward") || IsWard("warrior") {
		t.Fatal("ward detect")
	}
}
