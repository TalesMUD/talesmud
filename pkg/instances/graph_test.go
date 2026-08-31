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

func TestCollectGraphFollowsHiddenCellarWing(t *testing.T) {
	all := map[string]*rooms.Room{
		"R0203": room("R0203", rooms.Exit{Name: "down", Target: "R0215", Type: "direction"}),
		"R0215": room("R0215",
			rooms.Exit{Name: "up", Target: "R0203"},
			rooms.Exit{Name: "deeper", Target: "R0230", Hidden: true}),
		"R0230": room("R0230",
			rooms.Exit{Name: "back", Target: "R0215"},
			rooms.Exit{Name: "deeper", Target: "R0231"}),
		"R0231": room("R0231", rooms.Exit{Name: "back", Target: "R0230"}),
		"R0205": room("R0205", rooms.Exit{Name: "north", Target: "R0203"}),
	}
	got := CollectGraph(all, "R0203", "R0215")
	set := map[string]bool{}
	for _, id := range got {
		set[id] = true
	}
	if !set["R0215"] || !set["R0230"] || !set["R0231"] {
		t.Fatalf("expected inn cellar + hidden wing, got %v", got)
	}
	if set["R0203"] || set["R0205"] {
		t.Fatalf("hub and town must stay shared, got %v", got)
	}
}

func TestCrossingIntoInstanceUsesRoomTags(t *testing.T) {
	hub := room("R0203", rooms.Exit{Name: "down", Target: "R0215", Type: "direction"})
	hub.Tags = []string{"shared", "inn"}
	cellar := room("R0215", rooms.Exit{Name: "up", Target: "R0203"})
	cellar.Tags = []string{"instance", "cellar"}
	down, _ := hub.GetExit("down")
	if !CrossingIntoInstance(hub, cellar, down) {
		t.Fatal("down from shared inn into tagged cellar must instance")
	}
	up, _ := cellar.GetExit("up")
	if CrossingIntoInstance(cellar, hub, up) {
		t.Fatal("leaving the cellar must not spawn another instance")
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
