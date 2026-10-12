package npc

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

func TestValidateMovementRequiresRealRouteIncludingClosure(t *testing.T) {
	ab := rooms.Exits{{Name: "east", Target: "b"}}
	ba := rooms.Exits{{Name: "west", Target: "a"}}
	world := map[string]*rooms.Room{"a": {Entity: &entities.Entity{ID: "a"}, Area: "zone", Exits: &ab}, "b": {Entity: &entities.Entity{ID: "b"}, Area: "zone", Exits: &ba}}
	m := &Movement{Mode: "route", Route: []string{"a", "b"}, Zones: []string{"zone"}}
	if issues := ValidateMovement(m, world); len(issues) > 0 {
		t.Fatal(issues)
	}
	ba[0].Hidden = true
	for _, loop := range []string{"loop", "pingpong"} {
		m.Loop = loop
		if len(ValidateMovement(m, world)) == 0 {
			t.Fatal("missing closure/reverse not checked")
		}
	}
	ba[0].Hidden = false
	for _, cfg := range []*Movement{{Mode: "bad"}, {Mode: "route", Route: []string{"a"}}, {Mode: "roam", Loop: "bad"}, {Mode: "roam", IntervalSeconds: -1}, {Mode: "roam", Zones: []string{"missing"}}, {Mode: "roam", Rooms: []string{"missing"}}, {Mode: "route", Route: []string{"a", "a"}}, {Mode: "route", Route: []string{"a", "b"}, Zones: []string{"elsewhere"}}} {
		if len(ValidateMovement(cfg, world)) == 0 {
			t.Fatalf("accepted %#v", cfg)
		}
	}
}
func TestMovementCopyDoesNotAliasTemplate(t *testing.T) {
	src := &Movement{Mode: "route", Route: []string{"a", "b"}, Rooms: []string{"a"}, Zones: []string{"z"}}
	cp := src.Clone()
	cp.Route[0] = "changed"
	cp.Rooms[0] = "changed"
	cp.Zones[0] = "changed"
	if src.Route[0] != "a" || src.Rooms[0] != "a" || src.Zones[0] != "z" {
		t.Fatal("instance mutated template")
	}
}
