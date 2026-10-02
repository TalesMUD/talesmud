package commands

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/combat"
)

func TestResolveLivingEnemyByQueryHashIndex(t *testing.T) {
	enemies := []combat.CombatantRef{
		{ID: "rat-1", Name: "Sewer Rat", IsAlive: true, CurrentHP: 32, MaxHP: 32},
		{ID: "rat-2", Name: "Sewer Rat", IsAlive: true, CurrentHP: 32, MaxHP: 32},
		{ID: "rat-3", Name: "Sewer Rat", IsAlive: true, CurrentHP: 32, MaxHP: 32},
		{ID: "brute", Name: "Burrow Brute", IsAlive: true, CurrentHP: 100, MaxHP: 100},
	}

	cases := []struct {
		query, wantID string
	}{
		{"Sewer Rat#1", "rat-1"},
		{"Sewer Rat#2", "rat-2"},
		{"Sewer Rat#3", "rat-3"},
		{"sewer rat#2", "rat-2"},
		{"Sewer Rat #2", "rat-2"}, // space before # tolerated via TrimSpace on base
		{"rat-3", "rat-3"},
		{"Burrow Brute", "brute"},
		{"Sewer Rat", "rat-1"}, // first partial match
		{"Sewer Rat#9", ""},    // out of range
		{"missing", ""},
	}
	for _, tc := range cases {
		gotID, _ := resolveLivingEnemyByQuery(enemies, tc.query)
		if gotID != tc.wantID {
			t.Fatalf("query %q: got %q want %q", tc.query, gotID, tc.wantID)
		}
	}
}

func TestResolveLivingEnemyByQuerySkipsDead(t *testing.T) {
	enemies := []combat.CombatantRef{
		{ID: "rat-1", Name: "Sewer Rat", IsAlive: false},
		{ID: "rat-2", Name: "Sewer Rat", IsAlive: true},
		{ID: "rat-3", Name: "Sewer Rat", IsAlive: true},
	}
	// Living indices: #1=rat-2, #2=rat-3 (dead rat-1 skipped)
	got, _ := resolveLivingEnemyByQuery(enemies, "Sewer Rat#1")
	if got != "rat-2" {
		t.Fatalf("got %q want rat-2", got)
	}
	got, _ = resolveLivingEnemyByQuery(enemies, "Sewer Rat#2")
	if got != "rat-3" {
		t.Fatalf("got %q want rat-3", got)
	}
}

func TestLivingEnemyDisplayLabels(t *testing.T) {
	enemies := []combat.CombatantRef{
		{ID: "a", Name: "Sewer Rat", IsAlive: true},
		{ID: "b", Name: "Sewer Rat", IsAlive: true},
		{ID: "c", Name: "Unique", IsAlive: true},
		{ID: "d", Name: "Sewer Rat", IsAlive: false},
	}
	labels := livingEnemyDisplayLabels(enemies)
	if labels["a"] != "Sewer Rat#1" || labels["b"] != "Sewer Rat#2" {
		t.Fatalf("dup labels = %+v", labels)
	}
	if labels["c"] != "Unique" {
		t.Fatalf("unique label = %q", labels["c"])
	}
	if _, ok := labels["d"]; ok {
		t.Fatalf("dead enemy should not be labeled")
	}
}
