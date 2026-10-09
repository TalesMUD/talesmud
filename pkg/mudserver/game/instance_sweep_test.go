package game

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/service"
)

func importRoom(t *testing.T, facade service.Facade, id, name string, tags []string, exits rooms.Exits) {
	t.Helper()
	ex := exits
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity: &entities.Entity{ID: id},
		Name:   name,
		Tags:   tags,
		Exits:  &ex,
	}); err != nil {
		t.Fatalf("import room %s: %v", id, err)
	}
}

func importChar(t *testing.T, facade service.Facade, id, name, roomID, bound string, inCombat bool) {
	t.Helper()
	ch := &characters.Character{
		Entity:      &entities.Entity{ID: id},
		Name:        name,
		BelongsUser: *traits.BelongsToUser("user-" + id),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: roomID},
		BoundRoomID: bound,
	}
	if inCombat {
		ch.InCombat = true
		ch.CombatInstanceID = "stale-fight"
	}
	if _, err := facade.CharactersService().Import(ch); err != nil {
		t.Fatalf("import character %s: %v", id, err)
	}
}

func TestInstanceSweepRelocatesCopiesAndIsIdempotent(t *testing.T) {
	g, facade := newNPCTestGame(t)
	importRoom(t, facade, "R0001", "Awakening Chamber", nil, nil)
	importRoom(t, facade, "R0201", "Sewer Grate", nil, rooms.Exits{
		{Name: "down", Target: "R0210", Type: "instance", Instance: true},
	})
	importRoom(t, facade, "R0202", "Lane", nil, rooms.Exits{{Name: "east", Target: "R0201"}})
	importRoom(t, facade, "R0210", "Cellar Mouth", []string{"instance"}, rooms.Exits{
		{Name: "up", Target: "R0201"},
		{Name: "east", Target: "R0211"},
	})
	importRoom(t, facade, "R0211", "Cellar Deep", []string{"instance"}, rooms.Exits{
		{Name: "west", Target: "R0210"},
	})
	importRoom(t, facade, "R0210~abcd", "Cellar Mouth", nil, rooms.Exits{
		{Name: "up", Target: "R0201"},
		{Name: "east", Target: "R0211~abcd"},
	})
	importRoom(t, facade, "R0211~abcd", "Cellar Deep", nil, rooms.Exits{
		{Name: "west", Target: "R0210~abcd"},
	})
	importChar(t, facade, "diver", "Diver", "R0211~abcd", "R0202", true)
	importChar(t, facade, "lost", "Lost", "R4040", "R0202", true)
	importChar(t, facade, "home", "Home", "R0001", "R0001", true)
	importChar(t, facade, "blank", "Blank", "", "R0001", false)

	if _, err := facade.NPCsService().Import(&npc.NPC{
		Entity:      &entities.Entity{ID: "rat~R0210~abcd"},
		Name:        "Copy Rat",
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "R0210~abcd"},
		SpawnRoomID: "R0210~abcd",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := facade.NPCsService().Import(&npc.NPC{
		Entity:      &entities.Entity{ID: "guard"},
		Name:        "Guard",
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "R0001"},
		SpawnRoomID: "R0001",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := facade.NPCSpawnersService().Import(&npc.NPCSpawner{
		Entity:     &entities.Entity{ID: "sp-copy"},
		RoomID:     "R0210~abcd",
		TemplateID: "rat",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := facade.NPCSpawnersService().Import(&npc.NPCSpawner{
		Entity:     &entities.Entity{ID: "sp-town"},
		RoomID:     "R0001",
		TemplateID: "guard",
	}); err != nil {
		t.Fatal(err)
	}

	stats := g.SweepInstanceRooms()
	if stats.RoomsDeleted != 2 || stats.NPCsDeleted != 1 || stats.SpawnersDeleted != 1 || stats.CharactersMoved != 2 {
		t.Fatalf("stats = %+v", stats)
	}
	for _, id := range []string{"R0001", "R0201", "R0202", "R0210", "R0211"} {
		if _, err := facade.RoomsService().FindByID(id); err != nil {
			t.Fatalf("authored room %s was removed: %v", id, err)
		}
	}
	for _, id := range []string{"R0210~abcd", "R0211~abcd"} {
		if _, err := facade.RoomsService().FindByID(id); err == nil {
			t.Fatalf("copy %s still exists", id)
		}
	}
	if _, err := facade.NPCsService().FindByID("rat~R0210~abcd"); err == nil {
		t.Fatal("npc copy still exists")
	}
	if _, err := facade.NPCsService().FindByID("guard"); err != nil {
		t.Fatalf("resident npc removed: %v", err)
	}
	if _, err := facade.NPCSpawnersService().FindByID("sp-copy"); err == nil {
		t.Fatal("copy spawner still exists")
	}
	if _, err := facade.NPCSpawnersService().FindByID("sp-town"); err != nil {
		t.Fatalf("town spawner removed: %v", err)
	}

	diver, err := facade.CharactersService().FindByID("diver")
	if err != nil {
		t.Fatal(err)
	}
	if diver.CurrentRoomID != "R0201" || diver.InCombat || diver.CombatInstanceID != "" {
		t.Fatalf("diver room=%s combat=%v instance=%q", diver.CurrentRoomID, diver.InCombat, diver.CombatInstanceID)
	}
	lost, err := facade.CharactersService().FindByID("lost")
	if err != nil {
		t.Fatal(err)
	}
	if lost.CurrentRoomID != "R0001" || lost.InCombat {
		t.Fatalf("lost room=%s combat=%v", lost.CurrentRoomID, lost.InCombat)
	}
	home, err := facade.CharactersService().FindByID("home")
	if err != nil {
		t.Fatal(err)
	}
	if home.CurrentRoomID != "R0001" || !home.InCombat || home.CombatInstanceID != "stale-fight" {
		t.Fatalf("home should stay put with its flag, room=%s combat=%v id=%q", home.CurrentRoomID, home.InCombat, home.CombatInstanceID)
	}
	if blank := mustChar(t, facade, "blank"); blank.CurrentRoomID != "" {
		t.Fatalf("empty room was filled with %s", blank.CurrentRoomID)
	}
	hub, err := facade.RoomsService().FindByID("R0201")
	if err != nil {
		t.Fatal(err)
	}
	if !hub.IsCharacterInRoom("diver") {
		t.Fatal("diver was not placed in the hub")
	}

	if got := g.TakeRelocationNotice("diver"); got != "You find yourself back at Sewer Grate." {
		t.Fatalf("diver line %q", got)
	}
	if got := g.TakeRelocationNotice("lost"); got != "You find yourself back at Awakening Chamber." {
		t.Fatalf("lost line %q", got)
	}
	if got := g.TakeRelocationNotice("diver"); got != "" {
		t.Fatalf("login line delivered twice: %q", got)
	}
	if got := g.TakeRelocationNotice("home"); got != "" {
		t.Fatalf("home got a line: %q", got)
	}

	again := g.SweepInstanceRooms()
	if again.RoomsDeleted != 0 || again.NPCsDeleted != 0 || again.SpawnersDeleted != 0 || again.CharactersMoved != 0 {
		t.Fatalf("second sweep = %+v", again)
	}
	if got := g.TakeRelocationNotice("diver"); got != "" {
		t.Fatalf("second sweep rearmed the line: %q", got)
	}
}

func TestInstanceSweepUsesProceduralReturnExit(t *testing.T) {
	g, facade := newNPCTestGame(t)
	importRoom(t, facade, "R0001", "Awakening Chamber", nil, nil)
	importRoom(t, facade, "R0108", "Creek", nil, nil)
	importRoom(t, facade, "slot-0~deadbeef", "Trail", nil, rooms.Exits{
		{Name: "out", Target: "R0108"},
	})
	importChar(t, facade, "walker", "Walker", "slot-0~deadbeef", "R0108", false)

	stats := g.SweepInstanceRooms()
	if stats.RoomsDeleted != 1 || stats.CharactersMoved != 1 {
		t.Fatalf("stats = %+v", stats)
	}
	walker, err := facade.CharactersService().FindByID("walker")
	if err != nil {
		t.Fatal(err)
	}
	if walker.CurrentRoomID != "R0108" {
		t.Fatalf("room=%s, want the return exit", walker.CurrentRoomID)
	}
	if _, err := facade.RoomsService().FindByID("slot-0~deadbeef"); err == nil {
		t.Fatal("procedural copy still exists")
	}
}

func TestEnsureLivingRoomLeavesLiveCopyAndMovesOrphan(t *testing.T) {
	g, facade := newNPCTestGame(t)
	importRoom(t, facade, "R0201", "Sewer Grate", nil, rooms.Exits{
		{Name: "down", Target: "R0210", Type: "instance", Instance: true},
	})
	importRoom(t, facade, "R0210", "Cellar Mouth", []string{"instance"}, rooms.Exits{
		{Name: "up", Target: "R0201"},
	})
	importRoom(t, facade, "haven", "Haven", nil, nil)

	live, err := g.RoomInstances.Enter("hero", "R0201", "R0210")
	if err != nil {
		t.Fatal(err)
	}
	importChar(t, facade, "hero", "Hero", live, "haven", false)
	g.EnsureLivingRoom(mustChar(t, facade, "hero"))
	if got := mustChar(t, facade, "hero").CurrentRoomID; got != live {
		t.Fatalf("live copy relocated to %s", got)
	}
	if line := g.TakeRelocationNotice("hero"); line != "" {
		t.Fatalf("live copy announced %q", line)
	}

	importRoom(t, facade, "R0210~orphan", "Cellar Mouth", nil, rooms.Exits{
		{Name: "up", Target: "R0201"},
	})
	importChar(t, facade, "ghost", "Ghost", "R0210~orphan", "haven", true)
	g.EnsureLivingRoom(mustChar(t, facade, "ghost"))
	ghost := mustChar(t, facade, "ghost")
	if ghost.CurrentRoomID != "R0201" || ghost.InCombat {
		t.Fatalf("orphan room=%s combat=%v", ghost.CurrentRoomID, ghost.InCombat)
	}
	if line := g.TakeRelocationNotice("ghost"); line != "You find yourself back at Sewer Grate." {
		t.Fatalf("orphan line %q", line)
	}
	if line := g.TakeRelocationNotice("ghost"); line != "" {
		t.Fatalf("orphan line twice %q", line)
	}
}

func mustChar(t *testing.T, facade service.Facade, id string) *characters.Character {
	t.Helper()
	ch, err := facade.CharactersService().FindByID(id)
	if err != nil || ch == nil {
		t.Fatalf("character %s: %v", id, err)
	}
	return ch
}
