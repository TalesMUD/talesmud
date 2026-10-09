package inspect

import (
	"strings"
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/scripts"
	"github.com/talesmud/talesmud/pkg/worldindex"
)

func TestRoomInspectExitsRevealersReachAndReferences(t *testing.T) {
	exits := rooms.Exits{
		{Name: "north", Target: "R1"},
		{Name: "down", Target: "R2", Hidden: true},
	}
	back := rooms.Exits{{Name: "south", Target: "R0", Hidden: true}}
	actions := rooms.Actions{{Name: "pull lever", Type: rooms.RoomActionTypeScript, ScriptId: "S2"}}
	itemIDs := rooms.Items{"ITM1", "ITM-gone"}
	residentIDs := rooms.NPCs{"NPC1"}
	snap := worldindex.NewSnapshot()
	snap.StartRoomID = "R0"
	snap.Rooms["R0"] = &rooms.Room{
		Entity: &entities.Entity{ID: "R0"}, Name: "Awakening", Description: "A stone room.",
		Area: "Town", RoomType: "safe", Tags: []string{"start"},
		OnEnterScriptID: "S1", Exits: &exits, Actions: &actions, Items: &itemIDs, NPCs: &residentIDs,
	}
	snap.Rooms["R1"] = &rooms.Room{Entity: &entities.Entity{ID: "R1"}, Name: "Hall", Exits: &back}
	snap.Rooms["R2"] = &rooms.Room{Entity: &entities.Entity{ID: "R2"}, Name: "Cellar"}
	snap.Rooms["R9"] = &rooms.Room{Entity: &entities.Entity{ID: "R9"}, Name: "Island"}
	snap.Scripts["S1"] = &scripts.Script{
		Entity: &entities.Entity{ID: "S1"}, Name: "Open hatch",
		Code: "tales.game.revealExit(\"R0\", \"down\", charID)\ntales.game.revealExit(\"R1\", \"south\", charID)\n",
	}
	snap.Scripts["S2"] = &scripts.Script{Entity: &entities.Entity{ID: "S2"}, Name: "Lever"}
	snap.Items["ITM1"] = &items.Item{Entity: &entities.Entity{ID: "ITM1"}, Name: "Torch", IsTemplate: true}
	snap.NPCs["NPC1"] = &npc.NPC{Entity: &entities.Entity{ID: "NPC1"}, Name: "Guide", Level: 1, SpawnRoomID: "R0"}
	snap.NPCs["ENM1"] = &npc.NPC{
		Entity: &entities.Entity{ID: "ENM1"}, Name: "Rat", Level: 2, IsTemplate: true, SpawnRoomID: "R0",
		EnemyTrait: &npc.EnemyTrait{Difficulty: "easy"},
	}
	snap.Spawners["SPW"] = &npc.NPCSpawner{
		Entity: &entities.Entity{ID: "SPW"}, Name: "Rats", TemplateID: "ENM1", RoomID: "R0",
		MaxInstances: 3, SpawnInterval: time.Minute, RespawnDelay: 5 * time.Minute,
	}
	snap.Spawners["SPW-else"] = &npc.NPCSpawner{
		Entity: &entities.Entity{ID: "SPW-else"}, TemplateID: "ENM1", RoomID: "R2", MaxInstances: 1,
	}
	snap.Quests["Q-visit"] = &quests.Quest{
		Entity: &entities.Entity{ID: "Q-visit"}, Name: "See the start",
		Objectives: []quests.Objective{{ID: "go", Type: quests.ObjectiveVisit, TargetID: "R0", Description: "Enter the chamber"}},
	}
	snap.Quests["Q-kill"] = &quests.Quest{
		Entity: &entities.Entity{ID: "Q-kill"}, Name: "Rats",
		Objectives: []quests.Objective{{ID: "slay", Type: quests.ObjectiveKill, TargetID: "ENM1"}},
	}

	ix := worldindex.Build(snap)
	view, ok := Room(ix, "R0")
	if !ok {
		t.Fatal("missing room")
	}
	if view.Name != "Awakening" || view.Area != "Town" || len(view.Tags) != 1 || view.Tags[0] != "start" {
		t.Fatalf("summary %+v", view)
	}
	if !view.Reach.Reachable || view.Reach.StartRoomID != "R0" || view.Reach.Reason != "" {
		t.Fatalf("reach %+v", view.Reach)
	}
	if len(view.ExitsOut) != 2 || view.ExitsOut[0].Name != "north" || view.ExitsOut[0].RoomName != "Hall" || view.ExitsOut[0].Hidden {
		t.Fatalf("north %+v", view.ExitsOut)
	}
	down := view.ExitsOut[1]
	if !down.Hidden || down.RoomID != "R2" || down.NoRevealer || len(down.RevealedBy) != 1 || down.RevealedBy[0].ID != "S1" {
		t.Fatalf("down %+v", down)
	}
	if len(view.ExitsIn) != 1 || view.ExitsIn[0].Name != "south" || view.ExitsIn[0].RoomName != "Hall" || !view.ExitsIn[0].Hidden {
		t.Fatalf("inbound %+v", view.ExitsIn)
	}
	if view.ExitsIn[0].NoRevealer || view.ExitsIn[0].RevealedBy[0].Name != "Open hatch" {
		t.Fatalf("inbound reveal %+v", view.ExitsIn[0])
	}
	if view.OnEnter == nil || view.OnEnter.ID != "S1" || len(view.Actions) != 1 || view.Actions[0].Script == nil || view.Actions[0].Script.ID != "S2" {
		t.Fatalf("scripts on=%+v actions=%+v", view.OnEnter, view.Actions)
	}
	if len(view.Items) != 2 || view.Items[0].Name != "Torch" || !view.Items[1].Missing {
		t.Fatalf("items %+v", view.Items)
	}
	if len(view.Actors) != 2 {
		t.Fatalf("actors %+v", view.Actors)
	}
	byID := map[string]RoomActor{}
	for _, actor := range view.Actors {
		byID[actor.ID] = actor
	}
	if byID["NPC1"].How != "resident" || byID["ENM1"].How != "spawn" || byID["ENM1"].Kinds[0] != "enemy" {
		t.Fatalf("actors %+v", byID)
	}
	if len(view.Spawners) != 1 || view.Spawners[0].TemplateID != "ENM1" || view.Spawners[0].MaxInstances != 3 || view.Spawners[0].RespawnTime != "5m0s" {
		t.Fatalf("spawners %+v", view.Spawners)
	}
	if len(view.Quests) != 1 || view.Quests[0].Role != "visit" || view.Quests[0].ID != "Q-visit" {
		t.Fatalf("quests %+v", view.Quests)
	}

	cellar, ok := Room(ix, "R2")
	if !ok || !cellar.Reach.Reachable {
		t.Fatalf("hidden exit should open from the on-enter script, reach=%+v ok=%v", cellar.Reach, ok)
	}
	island, ok := Room(ix, "R9")
	if !ok || island.Reach.Reachable || !strings.Contains(island.Reach.Reason, "no inbound") {
		t.Fatalf("island %+v", island.Reach)
	}
	if _, ok := Room(ix, "missing"); ok {
		t.Fatal("missing room")
	}
	if _, ok := Room(nil, "R0"); ok {
		t.Fatal("nil index")
	}
}

func TestParseExitHow(t *testing.T) {
	name, hidden, instance, ok := parseExitHow("instance hidden exit down")
	if !ok || name != "down" || !hidden || !instance {
		t.Fatalf("instance hidden: %q hidden=%v instance=%v ok=%v", name, hidden, instance, ok)
	}
	name, hidden, instance, ok = parseExitHow("exit north")
	if !ok || name != "north" || hidden || instance {
		t.Fatalf("plain: %q hidden=%v instance=%v ok=%v", name, hidden, instance, ok)
	}
	if _, _, _, ok = parseExitHow("on-enter script"); ok {
		t.Fatal("script how is not an exit")
	}
}
