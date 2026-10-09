package contenthealth

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

func TestMapReachabilityView(t *testing.T) {
	cellar := testRoom("R2", "Cellar")
	cellar.Tags = []string{"instance"}
	world := World{
		StartRoomID: "R0",
		Rooms: []*rooms.Room{
			testRoom("R0", "Start", rooms.Exit{Name: "north", Target: "R1"}),
			testRoom("R1", "Street", rooms.Exit{Name: "south", Target: "R0"}),
			cellar,
		},
	}

	view, ok := MapReachabilityView(world, "")
	if !ok {
		t.Fatal("expected a view")
	}
	if view.StartRoomID != "R0" {
		t.Fatalf("start = %q", view.StartRoomID)
	}
	if view.Islands == nil || view.Rooms == nil {
		t.Fatal("nil slices")
	}
	if len(view.Islands) != 1 || len(view.Islands[0].RoomIDs) != 1 || view.Islands[0].RoomIDs[0] != "R2" {
		t.Fatalf("islands = %+v", view.Islands)
	}
	if view.Islands[0].Reason == "" {
		t.Fatal("island reason is empty")
	}
	byID := map[string]MapRoom{}
	for _, room := range view.Rooms {
		byID[room.ID] = room
	}
	if !byID["R0"].Reachable || !byID["R1"].Reachable || byID["R1"].Instance {
		t.Fatalf("street = %+v start = %+v", byID["R1"], byID["R0"])
	}
	if byID["R2"].Reachable || !byID["R2"].Instance {
		t.Fatalf("cellar = %+v", byID["R2"])
	}

	fromCellar, ok := MapReachabilityView(world, " R2 ")
	if !ok || fromCellar.StartRoomID != "R2" {
		t.Fatalf("from cellar ok=%v start=%q", ok, fromCellar.StartRoomID)
	}
	byID = map[string]MapRoom{}
	for _, room := range fromCellar.Rooms {
		byID[room.ID] = room
	}
	if !byID["R2"].Reachable || byID["R0"].Reachable {
		t.Fatalf("from cellar rooms = %+v", byID)
	}

	if _, ok := MapReachabilityView(world, "NOPE"); ok {
		t.Fatal("missing start room should fail")
	}
}
