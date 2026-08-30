package rooms

import "testing"

func TestGetExitMatchesCaseInsensitively(t *testing.T) {
	exits := Exits{{Name: "north", Target: "R0004"}, {Name: "down", Target: "R0109", Hidden: true}}
	room := &Room{Exits: &exits}
	got, ok := room.GetExit("NORTH")
	if !ok || got.Target != "R0004" {
		t.Fatalf("expected north -> R0004, ok=%v got=%#v", ok, got)
	}
	down, ok := room.GetExit("DOWN")
	if !ok || down.Target != "R0109" || !down.Hidden {
		t.Fatalf("expected hidden down -> R0109, ok=%v got=%#v", ok, down)
	}
	if _, ok := room.GetExit("west"); ok {
		t.Fatal("west should not match")
	}
}

func TestRoomsFromJSONStringParsesArray(t *testing.T) {
	got, err := RoomsFromJSONString(`[{"id":"room-1","name":"Start"},{"id":"room-2","name":"North"}]`)
	if err != nil {
		t.Fatalf("RoomsFromJSONString returned error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 rooms, got %d", len(got))
	}
	if got[0].ID != "room-1" || got[0].Name != "Start" {
		t.Fatalf("first room not populated: %#v", got[0])
	}
	if got[1].ID != "room-2" || got[1].Name != "North" {
		t.Fatalf("second room not populated: %#v", got[1])
	}
}
