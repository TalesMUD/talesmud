package rooms

import "testing"

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
