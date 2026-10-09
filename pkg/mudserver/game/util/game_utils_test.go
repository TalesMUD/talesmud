package util

import (
	"reflect"
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
)

// stubRoomGame answers the two GameCtrl calls CreateRoomDescription makes.
// The embedded interfaces stay nil; unused methods are not called.
type stubRoomGame struct{ def.GameCtrl }

func (stubRoomGame) GetRoomPlayers(string, string) []def.OnlinePlayer { return nil }

func (stubRoomGame) GetNPCInstanceManager() def.NPCInstanceCtrl { return stubRoomNPCs{} }

type stubRoomNPCs struct{ def.NPCInstanceCtrl }

func (stubRoomNPCs) GetInstancesInRoom(string) []*npc.NPC { return nil }

func TestCreateRoomDescriptionNilExits(t *testing.T) {
	room := &rooms.Room{
		Entity:      &entities.Entity{ID: "room-nil-exits"},
		Name:        "Bare Cell",
		Description: "Stone and nothing else.",
	}
	got := CreateRoomDescription(room, &entities.User{LastCharacter: "char-1"}, stubRoomGame{})
	if !strings.Contains(got, "[Bare Cell]") || !strings.Contains(got, "Stone and nothing else.") {
		t.Fatalf("room text missing: %q", got)
	}
	if !strings.Contains(got, "- The visible exits are:") {
		t.Fatalf("exit header missing: %q", got)
	}
	if strings.Contains(got, " + [") {
		t.Fatalf("nil exits listed a direction: %q", got)
	}
}

func TestCreateRoomDescriptionListsVisibleExits(t *testing.T) {
	exits := rooms.Exits{
		{Name: "north", Description: "A corridor"},
		{Name: "secret", Description: "A crack", Hidden: true},
	}
	room := &rooms.Room{
		Entity:      &entities.Entity{ID: "room-exits"},
		Name:        "Landing",
		Description: "A landing.",
		Exits:       &exits,
	}
	got := CreateRoomDescription(room, &entities.User{}, stubRoomGame{})
	if !strings.Contains(got, " + [north] A corridor") {
		t.Fatalf("visible exit missing: %q", got)
	}
	if strings.Contains(got, "secret") || strings.Contains(got, "crack") {
		t.Fatalf("hidden exit listed: %q", got)
	}
}

func TestRemoveStringFromSliceRemovesLastElement(t *testing.T) {
	got := RemoveStringFromSlice([]string{"north", "south", "east"}, "east")
	want := []string{"north", "south"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %#v, got %#v", want, got)
	}
}

func TestRemoveStringFromSliceRemovesOnlyMatch(t *testing.T) {
	got := RemoveStringFromSlice([]string{"north", "south", "east"}, "south")
	want := []string{"north", "east"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %#v, got %#v", want, got)
	}
}
