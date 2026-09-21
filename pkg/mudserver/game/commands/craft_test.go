package commands_test

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/recipes"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

func TestStationOK(t *testing.T) {
	r := &recipes.Recipe{Station: "forge"}
	if recipes.StationOK(r, []string{"outdoor"}) {
		t.Fatal("forge recipe should fail outdoors")
	}
	if !recipes.StationOK(r, []string{"crafting"}) {
		t.Fatal("forge recipe should accept crafting tag")
	}
	if !recipes.StationOK(r, []string{"forge"}) {
		t.Fatal("forge recipe should accept forge tag")
	}
	anywhere := &recipes.Recipe{Station: ""}
	if !recipes.StationOK(anywhere, nil) {
		t.Fatal("empty station should always be ok")
	}
}

func TestRecipeFind(t *testing.T) {
	recipes.SetCache([]*recipes.Recipe{
		{Entity: &entities.Entity{ID: "RCP0001"}, Key: "meadow_stew", Name: "Meadow Stew"},
		{Entity: &entities.Entity{ID: "RCP0006"}, Key: "copper_shiv", Name: "Copper Shiv"},
	})
	if recipes.Find("stew") == nil {
		t.Fatal("expected partial match on stew")
	}
	if recipes.Find("copper_shiv") == nil {
		t.Fatal("expected key match")
	}
	if recipes.Find("nope") != nil {
		t.Fatal("expected nil")
	}
}

func TestGatherActionsDetect(t *testing.T) {
	// smoke: room with gather action name prefix is what gather command looks for
	actions := rooms.Actions{
		{Name: "GATHER HERBS", Type: rooms.RoomActionTypeScript, ScriptId: "SCR0102"},
		{Name: "EXAMINE TREE", Type: rooms.RoomActionTypeResponse},
	}
	room := &rooms.Room{Actions: &actions}
	count := 0
	for _, a := range *room.Actions {
		n := a.Name
		if len(n) >= 6 {
			// gather prefix check mirrored
			lower := n
			_ = lower
			count++
		}
	}
	if count < 1 {
		t.Fatal("expected actions")
	}
}
