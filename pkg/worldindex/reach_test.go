package worldindex

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/scripts"
)

func room(id string, exits ...rooms.Exit) *rooms.Room {
	ex := rooms.Exits(exits)
	return &rooms.Room{Entity: &entities.Entity{ID: id}, Name: id, Exits: &ex}
}

func TestReachabilityOpensCurrentRoomReveal(t *testing.T) {
	actions := rooms.Actions{{Name: "PRESS", Type: rooms.RoomActionTypeScript, ScriptId: "S1"}}
	start := room("R0", rooms.Exit{Name: "north", Target: "R1"})
	gate := room("R1", rooms.Exit{Name: "north", Target: "R2", Hidden: true})
	gate.Actions = &actions
	beyond := room("R2", rooms.Exit{Name: "south", Target: "R1"})
	snap := NewSnapshot()
	snap.StartRoomID = "R0"
	snap.Rooms["R0"] = start
	snap.Rooms["R1"] = gate
	snap.Rooms["R2"] = beyond
	snap.Scripts["S1"] = &scripts.Script{
		Entity: &entities.Entity{ID: "S1"},
		Name:   "Puzzle",
		Code:   "tales.game.revealExit(ctx.roomID, \"north\", charID)\n",
	}
	got := Build(snap).Reachability()
	if !got.Reachable["R2"] {
		t.Fatalf("current-room reveal should open the hidden exit, reachable=%v", got.Reachable)
	}
}

func TestReachabilityHiddenExitOpensWhenScriptIsInvoked(t *testing.T) {
	cellar := room("R1", rooms.Exit{Name: "up", Target: "R0"})
	actions := rooms.Actions{{Name: "EXAMINE", Type: rooms.RoomActionTypeScript, ScriptId: "S1"}}
	start := room("R0", rooms.Exit{Name: "down", Target: "R1", Hidden: true})
	start.Actions = &actions
	snap := NewSnapshot()
	snap.StartRoomID = "R0"
	snap.Rooms["R0"] = start
	snap.Rooms["R1"] = cellar
	snap.Scripts["S1"] = &scripts.Script{
		Entity: &entities.Entity{ID: "S1"},
		Name:   "Examine",
		Code:   "tales.game.revealExit(\"R0\", \"down\", charID)\n",
	}
	ix := Build(snap)
	got := ix.Reachability()
	if !got.Reachable["R1"] {
		t.Fatalf("hidden exit should open, reachable=%v islands=%+v", got.Reachable, got.Islands)
	}
}

func TestReachabilityThreeCutOffIslands(t *testing.T) {
	snap := NewSnapshot()
	snap.StartRoomID = "R0"
	link := func(id, back string) *rooms.Room {
		return room(id, rooms.Exit{Name: "out", Target: back})
	}
	r0 := room("R0", rooms.Exit{Name: "east", Target: "R1"}, rooms.Exit{Name: "west", Target: "R2"})
	r1 := room("R1", rooms.Exit{Name: "west", Target: "R0"})
	actions := rooms.Actions{{Name: "EXAMINE BARRELS", Type: rooms.RoomActionTypeScript, ScriptId: "S1"}}
	r1.Actions = &actions
	r2 := room("R2", rooms.Exit{Name: "west", Target: "R0"})
	rug := rooms.Actions{{Name: "EXAMINE RUG", Type: rooms.RoomActionTypeScript, ScriptId: "S2"}}
	r2.Actions = &rug
	snap.Rooms["R0"] = r0
	snap.Rooms["R1"] = r1
	snap.Rooms["R2"] = r2
	snap.Rooms["A1"] = link("A1", "R1")
	snap.Rooms["A2"] = room("A2", rooms.Exit{Name: "back", Target: "A1"})
	snap.Rooms["B1"] = link("B1", "R2")
	snap.Rooms["C1"] = link("C1", "R0")
	snap.Rooms["C2"] = room("C2", rooms.Exit{Name: "back", Target: "C1"})
	snap.Scripts["S1"] = &scripts.Script{Entity: &entities.Entity{ID: "S1"}, Name: "barrels", Code: "local roomID = \"R1\"\ntales.game.revealExit(roomID, \"deeper\", charID)\n"}
	snap.Scripts["S2"] = &scripts.Script{Entity: &entities.Entity{ID: "S2"}, Name: "rug", Code: "local roomID = \"R2\"\ntales.game.revealExit(roomID, \"down\", charID)\n"}
	snap.Scripts["S3"] = &scripts.Script{Entity: &entities.Entity{ID: "S3"}, Name: "hatch", Code: "local roomID = \"R0\"\ntales.game.revealExit(roomID, \"down\", charID)\n"}

	got := Build(snap).Reachability()
	for _, id := range []string{"R0", "R1", "R2"} {
		if !got.Reachable[id] {
			t.Fatalf("%s should be reachable", id)
		}
	}
	for _, id := range []string{"A1", "A2", "B1", "C1", "C2"} {
		if got.Reachable[id] {
			t.Fatalf("%s should be unreachable", id)
		}
	}
	if len(got.Islands) != 3 {
		t.Fatalf("islands = %+v", got.Islands)
	}
	byRoom := map[string]string{}
	for _, island := range got.Islands {
		for _, id := range island.RoomIDs {
			byRoom[id] = island.Reason
		}
	}
	if !strings.Contains(byRoom["A1"], "revealExit target 'deeper' missing on R1 (S1)") {
		t.Fatalf("barrel island reason %q", byRoom["A1"])
	}
	if !strings.Contains(byRoom["B1"], "revealExit target 'down' missing on R2 (S2)") {
		t.Fatalf("vault island reason %q", byRoom["B1"])
	}
	if !strings.Contains(byRoom["C1"], "S3 never invoked") || !strings.Contains(byRoom["C1"], "R0 no 'down'") {
		t.Fatalf("hatch island reason %q", byRoom["C1"])
	}
}

func TestUnreferencedScriptAndInboundExit(t *testing.T) {
	snap := NewSnapshot()
	snap.StartRoomID = "R0"
	snap.Rooms["R0"] = room("R0")
	snap.Rooms["R9"] = room("R9")
	snap.Scripts["DEAD"] = &scripts.Script{Entity: &entities.Entity{ID: "DEAD"}, Name: "dead", Code: "tales.game.giveItem(charID, \"NOPE\")\n"}
	ix := Build(snap)
	if got := ix.Inbound(KindScript, "DEAD"); len(got) != 0 {
		t.Fatalf("dead script inbound = %+v", got)
	}
	islands := ix.Reachability().Islands
	if len(islands) != 1 || islands[0].Reason != "no inbound exit" {
		t.Fatalf("islands = %+v", islands)
	}
	if give := ix.Outbound(KindScript, "DEAD"); len(give) != 1 || give[0].ToID != "NOPE" || give[0].How != "giveItem" {
		t.Fatalf("give edge = %+v", give)
	}
}

func TestQuestAndBossEdges(t *testing.T) {
	snap := NewSnapshot()
	snap.StartRoomID = "R0"
	snap.Rooms["R0"] = room("R0", rooms.Exit{Name: "north", Target: "R1"})
	snap.Rooms["R1"] = room("R1", rooms.Exit{Name: "south", Target: "R0"})
	snap.NPCs["BOSS"] = &npc.NPC{
		Entity:      &entities.Entity{ID: "BOSS"},
		Name:        "Boss",
		IsTemplate:  false,
		SpawnRoomID: "R1",
		EnemyTrait:  &npc.EnemyTrait{Difficulty: "boss", LootTableID: "LT1"},
	}
	snap.NPCs["RAT"] = &npc.NPC{
		Entity:     &entities.Entity{ID: "RAT"},
		Name:       "Rat",
		IsTemplate: true,
		EnemyTrait: &npc.EnemyTrait{Difficulty: "normal"},
	}
	snap.Spawners["SP1"] = &npc.NPCSpawner{Entity: &entities.Entity{ID: "SP1"}, TemplateID: "RAT", RoomID: "R1"}
	snap.Quests["Q1"] = &quests.Quest{
		Entity: &entities.Entity{ID: "Q1"},
		Name:   "Visit",
		Source: quests.QuestSource{Type: "npc", NPCID: "BOSS"},
		Objectives: []quests.Objective{{
			ID: "v", Type: quests.ObjectiveVisit, TargetID: "R9",
		}},
	}
	snap.Rooms["R9"] = room("R9", rooms.Exit{Name: "out", Target: "R0"})
	ix := Build(snap)
	reach := ix.Reachability()
	if reach.Reachable["R9"] {
		t.Fatal("R9 should be cut off")
	}
	if !ix.npcReachable(snap.NPCs["RAT"], reach.Reachable, reach.InvokedScripts) {
		t.Fatal("rat spawner room is reachable")
	}
	if !ix.npcReachable(snap.NPCs["BOSS"], reach.Reachable, nil) {
		t.Fatal("boss spawn room is reachable")
	}
}
