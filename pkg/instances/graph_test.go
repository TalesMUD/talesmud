package instances

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

func room(id string, exits ...rooms.Exit) *rooms.Room {
	ex := rooms.Exits(exits)
	return &rooms.Room{Entity: &entities.Entity{ID: id}, Name: id, Exits: &ex}
}

func TestCollectGraphStopsAtHub(t *testing.T) {
	all := map[string]*rooms.Room{
		"R0201": room("R0201", rooms.Exit{Name: "down", Target: "R0210", Type: "instance"}),
		"R0210": room("R0210",
			rooms.Exit{Name: "up", Target: "R0201"},
			rooms.Exit{Name: "east", Target: "R0211"}),
		"R0211": room("R0211", rooms.Exit{Name: "west", Target: "R0210"}),
		"R0202": room("R0202", rooms.Exit{Name: "south", Target: "R0201"}),
	}
	got := CollectGraph(all, "R0201", "R0210")
	if len(got) != 2 {
		t.Fatalf("expected cellar rooms [R0210 R0211], got %v", got)
	}
	set := map[string]bool{}
	for _, id := range got {
		set[id] = true
	}
	if !set["R0210"] || !set["R0211"] {
		t.Fatalf("subgraph %v", got)
	}
	if set["R0201"] || set["R0202"] {
		t.Fatal("hub must stay shared, not cloned")
	}
}

func TestCollectGraphSingleRoomCellar(t *testing.T) {
	all := map[string]*rooms.Room{
		"hub":    room("hub", rooms.Exit{Name: "down", Target: "cellar", Type: "instance"}),
		"cellar": room("cellar", rooms.Exit{Name: "up", Target: "hub"}),
	}
	got := CollectGraph(all, "hub", "cellar")
	if len(got) != 1 || got[0] != "cellar" {
		t.Fatalf("got %v", got)
	}
}

func TestInstanceEntrance(t *testing.T) {
	if !IsInstanceEntrance(rooms.Exit{Type: "instance"}) {
		t.Fatal("type=instance should be an entrance")
	}
	if !IsInstanceEntrance(rooms.Exit{Instance: true, Type: "direction"}) {
		t.Fatal("Instance flag should be an entrance")
	}
	if IsInstanceEntrance(rooms.Exit{Type: "direction"}) {
		t.Fatal("normal direction is not an instance entrance")
	}
}

func TestCloneIDRoundTrip(t *testing.T) {
	id := CloneID("R0210", "aabbccdd")
	if id != "R0210~aabbccdd" {
		t.Fatalf("got %q", id)
	}
	if TemplateIDFromClone(id) != "R0210" {
		t.Fatalf("template %q", TemplateIDFromClone(id))
	}
	if TemplateIDFromClone("R0201") != "R0201" {
		t.Fatal("uncloned id should pass through")
	}
}
