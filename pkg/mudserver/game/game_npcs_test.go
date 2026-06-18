package game

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
)

func TestNextPatrolRoomAdvancesAndWraps(t *testing.T) {
	inst := &npc.NPC{
		CurrentRoom: traits.CurrentRoom{},
		PatrolPath:  []string{"R1", "R2", "R3"},
	}

	inst.CurrentRoomID = "R1"
	if got := nextPatrolRoom(inst); got != "R2" {
		t.Fatalf("expected R2, got %q", got)
	}

	inst.CurrentRoomID = "R3"
	if got := nextPatrolRoom(inst); got != "R1" {
		t.Fatalf("expected wrap to R1, got %q", got)
	}

	inst.CurrentRoomID = "outside"
	if got := nextPatrolRoom(inst); got != "R1" {
		t.Fatalf("expected first path room when off path, got %q", got)
	}
}

func TestNextWanderRoomUsesVisibleExitWithinRadius(t *testing.T) {
	inst := &npc.NPC{
		CurrentRoom:  traits.CurrentRoom{},
		WanderRadius: 1,
	}
	inst.CurrentRoomID = "spawn"

	roomsByID := map[string]*rooms.Room{
		"spawn": testRoomWithExits("spawn", []rooms.Exit{
			{Name: "east", Target: "near"},
			{Name: "west", Target: "hidden", Hidden: true},
		}),
		"near":   testRoomWithExits("near", nil),
		"hidden": testRoomWithExits("hidden", nil),
	}

	if got := nextWanderRoom(inst, roomsByID, "spawn"); got != "near" {
		t.Fatalf("expected visible neighbor near, got %q", got)
	}
}

func TestNextWanderRoomRejectsExitOutsideRadius(t *testing.T) {
	inst := &npc.NPC{
		CurrentRoom:  traits.CurrentRoom{},
		WanderRadius: 1,
	}
	inst.CurrentRoomID = "middle"

	roomsByID := map[string]*rooms.Room{
		"spawn":  testRoomWithExits("spawn", []rooms.Exit{{Name: "east", Target: "middle"}}),
		"middle": testRoomWithExits("middle", []rooms.Exit{{Name: "east", Target: "far"}}),
		"far":    testRoomWithExits("far", nil),
	}

	if got := nextWanderRoom(inst, roomsByID, "spawn"); got != "" {
		t.Fatalf("expected no move outside radius, got %q", got)
	}
}

func TestDistanceWithinRadius(t *testing.T) {
	roomsByID := map[string]*rooms.Room{
		"R1": testRoomWithExits("R1", []rooms.Exit{{Name: "east", Target: "R2"}}),
		"R2": testRoomWithExits("R2", []rooms.Exit{{Name: "east", Target: "R3"}}),
		"R3": testRoomWithExits("R3", nil),
	}

	if !distanceWithinRadius(roomsByID, "R1", "R2", 1) {
		t.Fatal("expected R2 to be within radius 1")
	}
	if distanceWithinRadius(roomsByID, "R1", "R3", 1) {
		t.Fatal("expected R3 to be outside radius 1")
	}
	if !distanceWithinRadius(roomsByID, "R1", "R3", 2) {
		t.Fatal("expected R3 to be within radius 2")
	}
}

func testRoomWithExits(id string, exits []rooms.Exit) *rooms.Room {
	roomExits := rooms.Exits(exits)
	return &rooms.Room{
		Entity: &entities.Entity{ID: id},
		Name:   id,
		Exits:  &roomExits,
	}
}
