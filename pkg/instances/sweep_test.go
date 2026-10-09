package instances

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

func tagRoom(id string, tags []string, exits ...rooms.Exit) *rooms.Room {
	r := room(id, exits...)
	r.Tags = tags
	r.Name = id
	return r
}

func TestIsCloneID(t *testing.T) {
	if !IsCloneID(CloneID("R0210", "abcd")) {
		t.Fatal("copy id")
	}
	if IsCloneID("R0210") || IsCloneID("") {
		t.Fatal("authored ids are not copies")
	}
	if InstanceID(CloneID("R0210", "abcd")) != "abcd" {
		t.Fatal("suffix")
	}
	if InstanceID("R0210") != "" {
		t.Fatal("plain id has no suffix")
	}
}

func TestEntranceWalksBackToHub(t *testing.T) {
	all := map[string]*rooms.Room{
		"R0001": tagRoom("R0001", nil),
		"R0201": tagRoom("R0201", nil, rooms.Exit{Name: "down", Target: "R0210", Type: "instance", Instance: true}),
		"R0202": tagRoom("R0202", nil, rooms.Exit{Name: "east", Target: "R0201"}),
		"R0210": tagRoom("R0210", []string{"instance"},
			rooms.Exit{Name: "up", Target: "R0201"},
			rooms.Exit{Name: "east", Target: "R0211"}),
		"R0211":      tagRoom("R0211", []string{"instance"}, rooms.Exit{Name: "west", Target: "R0210"}),
		"R0211~abcd": tagRoom("R0211~abcd", nil, rooms.Exit{Name: "west", Target: "R0210~abcd"}),
	}
	if got := EntranceForTemplate(all, "R0211"); got != "R0201" {
		t.Fatalf("deep room entrance = %q", got)
	}
	if got := EntranceForTemplate(all, "R0210"); got != "R0201" {
		t.Fatalf("mouth entrance = %q", got)
	}
	if got := EntranceForTemplate(all, "R0211~abcd"); got != "" {
		t.Fatalf("clone id is not a template, got %q", got)
	}
	if got := EntranceForTemplate(all, "missing"); got != "" {
		t.Fatalf("missing template got %q", got)
	}
}

func TestRelocationOrder(t *testing.T) {
	all := map[string]*rooms.Room{
		"R0001":       tagRoom("R0001", nil),
		"R0108":       tagRoom("R0108", nil),
		"R0201":       tagRoom("R0201", nil, rooms.Exit{Name: "down", Target: "R0210", Type: "instance"}),
		"R0210":       tagRoom("R0210", []string{"instance"}, rooms.Exit{Name: "up", Target: "R0201"}),
		"R0210~abcd":  tagRoom("R0210~abcd", nil, rooms.Exit{Name: "up", Target: "R0201"}),
		"slot-0~dead": tagRoom("slot-0~dead", nil, rooms.Exit{Name: "out", Target: "R0108"}),
	}
	if got := RelocationDest(all, "R0210~abcd", "R0001", "R0108"); got != "R0201" {
		t.Fatalf("copy should use the hub, got %q", got)
	}
	if got := RelocationDest(all, "slot-0~dead", "R0001", "R0108"); got != "R0108" {
		t.Fatalf("procedural copy should use the return exit, got %q", got)
	}
	if got := RelocationDest(all, "gone", "R0001", "R0108"); got != "R0001" {
		t.Fatalf("missing room should use the start room, got %q", got)
	}
	if got := RelocationDest(all, "gone", "", "R0108"); got != "R0108" {
		t.Fatalf("missing start should use the fallback, got %q", got)
	}
	if got := RelocationDest(all, "gone", "R0210~abcd", "nope"); got != "" {
		t.Fatalf("clone and missing anchors are not destinations, got %q", got)
	}
}
