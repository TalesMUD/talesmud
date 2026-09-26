package commands

import "testing"

func TestActionArgs(t *testing.T) {
	args, ok := actionArgs("deposit 20", "deposit")
	if !ok || args != "20" {
		t.Fatalf("deposit args=%q ok=%v", args, ok)
	}
	args, ok = actionArgs("deposit", "deposit")
	if !ok || args != "" {
		t.Fatalf("exact args=%q ok=%v", args, ok)
	}
	if _, ok := actionArgs("examine moons extra", "examine moons"); ok {
		t.Fatal("multi-word action should stay exact")
	}
	if _, ok := actionArgs("a wolf", "a"); ok {
		t.Fatal("single-letter action should stay exact")
	}
}
