package portraits

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
)

func TestPortraitURLUsesTemplateForInstances(t *testing.T) {
	got := URL("Meadow Wolf-abc123", "NPC0010")
	if got != "/api/portraits/NPC0010.png" {
		t.Fatalf("got %q", got)
	}
}

func TestPortraitURLFallsBackToID(t *testing.T) {
	got := URL("NPC0001", "")
	if got != "/api/portraits/NPC0001.png" {
		t.Fatalf("got %q", got)
	}
}

func TestPortraitURLStripsInstanceSuffix(t *testing.T) {
	got := FileName("R0210~a1b2c3d4", "")
	if got != "R0210.png" {
		t.Fatalf("got %q", got)
	}
}

func TestForNPC(t *testing.T) {
	n := &npc.NPC{Entity: &entities.Entity{ID: "wolf-1"}, TemplateID: "NPC0010"}
	if ForNPC(n) != "/api/portraits/NPC0010.png" {
		t.Fatalf("got %q", ForNPC(n))
	}
}

func TestForPlayer(t *testing.T) {
	tests := []struct{ race, class, want string }{
		{"human", "warrior", "/api/portraits/player-human-warrior.png"},
		{"elve", "wizard", "/api/portraits/player-elf-mage.png"},
		{"dwarf", "hunter", "/api/portraits/player-dwarf-ranger.png"},
		{"orc", "warrior", ""},
	}
	for _, tt := range tests {
		ch := &characters.Character{Race: characters.Race{ID: tt.race}, Class: characters.Class{ID: tt.class}}
		if got := ForPlayer(ch); got != tt.want {
			t.Errorf("%s/%s: got %q, want %q", tt.race, tt.class, got, tt.want)
		}
	}
}
