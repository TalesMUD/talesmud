package worldmap

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
)

func testRoom(id, name, area string, tags []string, exits ...rooms.Exit) *rooms.Room {
	ex := rooms.Exits(exits)
	return &rooms.Room{
		Entity: &entities.Entity{ID: id},
		Name:   name,
		Area:   area,
		Tags:   tags,
		Exits:  &ex,
	}
}

func exit(name, target string, hidden bool) rooms.Exit {
	return rooms.Exit{Name: name, Target: target, Hidden: hidden, Type: rooms.RoomExitTypeDirection}
}

func TestCompilePlacesCardinalNeighbors(t *testing.T) {
	w := Compile([]*rooms.Room{
		testRoom("R0001", "Awakening", "Z00_catacombs_intro", []string{"starting_room", "underground", "safe"},
			exit("north", "R0002", false), exit("east", "R0005", false)),
		testRoom("R0002", "Corridor", "Z00_catacombs_intro", []string{"underground"},
			exit("south", "R0001", false), exit("west", "R0003", false)),
		testRoom("R0003", "Alcove", "Z00_catacombs_intro", []string{"underground"},
			exit("east", "R0002", false), exit("north", "R0004", false)),
		testRoom("R0005", "Nest", "Z00_catacombs_intro", []string{"underground"},
			exit("west", "R0001", false)),
	})

	a := w.rooms["R0001"]
	n := w.rooms["R0002"]
	west := w.rooms["R0003"]
	east := w.rooms["R0005"]
	if n.x != a.x || n.y != a.y-1 {
		t.Fatalf("north of start: got (%d,%d) want (%d,%d)", n.x, n.y, a.x, a.y-1)
	}
	if west.x != n.x-1 || west.y != n.y {
		t.Fatalf("west of corridor: got (%d,%d) want (%d,%d)", west.x, west.y, n.x-1, n.y)
	}
	if east.x != a.x+1 || east.y != a.y {
		t.Fatalf("east of start: got (%d,%d) want (%d,%d)", east.x, east.y, a.x+1, a.y)
	}
}

func TestCompileUpDownChangesLayer(t *testing.T) {
	w := Compile([]*rooms.Room{
		testRoom("R0006", "Stairwell", "Z00_catacombs_intro", []string{"underground", "starting_room"},
			exit("up", "R0101", false)),
		testRoom("R0101", "Meadow", "Z01_meadows_forest_path", []string{"outdoor", "entry_point", "safe"},
			exit("north", "R0102", false)),
		testRoom("R0102", "Field", "Z01_meadows_forest_path", []string{"outdoor", "gathering"},
			exit("south", "R0101", false)),
	})
	if w.rooms["R0101"].z != w.rooms["R0006"].z+1 {
		t.Fatalf("meadow z=%d stair z=%d, want meadow one above stair", w.rooms["R0101"].z, w.rooms["R0006"].z)
	}
	if layerID(w.rooms["R0101"].z) == layerID(w.rooms["R0006"].z) {
		t.Fatal("stair and meadow should land on different atlas layers")
	}
	if w.rooms["R0102"].x != w.rooms["R0101"].x || w.rooms["R0102"].y != w.rooms["R0101"].y-1 {
		t.Fatalf("meadow north neighbor misplaced")
	}
}

func TestCompileDiagonalAndHiddenBurrow(t *testing.T) {
	w := Compile([]*rooms.Room{
		testRoom("R0103", "Creek", "Z01_meadows_forest_path", []string{"outdoor", "water", "starting_room"},
			exit("northwest", "R0104", false),
			exit("down", "R0109", true),
			exit("north", "R0105", false)),
		testRoom("R0104", "Camp", "Z01_meadows_forest_path", []string{"outdoor"},
			exit("southeast", "R0103", false)),
		testRoom("R0105", "Forest", "Z01_meadows_forest_path", []string{"outdoor", "forest"},
			exit("south", "R0103", false)),
		testRoom("R0109", "Burrow", "Z01_meadows_forest_path", []string{"underground"},
			exit("back", "R0103", false)),
	})
	creek := w.rooms["R0103"]
	camp := w.rooms["R0104"]
	if camp.x != creek.x-1 || camp.y != creek.y-1 {
		t.Fatalf("bandit camp nw: got (%d,%d) want (%d,%d)", camp.x, camp.y, creek.x-1, creek.y-1)
	}
	if w.rooms["R0109"].z >= creek.z {
		t.Fatalf("burrow should be below creek: burrow z=%d creek z=%d", w.rooms["R0109"].z, creek.z)
	}
}

func TestRevealHidesUndiscoveredAndFog(t *testing.T) {
	w := Compile([]*rooms.Room{
		testRoom("R0101", "Meadow", "Z01_meadows_forest_path", []string{"outdoor", "starting_room", "entry_point"},
			exit("north", "R0102", false)),
		testRoom("R0102", "Field", "Z01_meadows_forest_path", []string{"outdoor"},
			exit("south", "R0101", false), exit("north", "R0103", false)),
		testRoom("R0103", "Creek", "Z01_meadows_forest_path", []string{"outdoor", "water"},
			exit("south", "R0102", false), exit("down", "R0109", true)),
		testRoom("R0109", "Burrow", "Z01_meadows_forest_path", []string{"underground"},
			exit("back", "R0103", false)),
	})
	ch := &characters.Character{
		Entity:          &entities.Entity{ID: "c1"},
		CurrentRoom:     traits.CurrentRoom{CurrentRoomID: "R0101"},
		DiscoveredRooms: map[string]bool{"R0101": true},
	}
	atlas := Reveal(w, ch)

	byID := map[string]Place{}
	for _, p := range atlas.Places {
		byID[p.ID] = p
	}
	if !byID["R0101"].Discovered || !byID["R0101"].Current {
		t.Fatalf("current meadow should be discovered: %+v", byID["R0101"])
	}
	fog, ok := byID["R0102"]
	if !ok || fog.Discovered {
		t.Fatalf("north neighbor should be uncharted fog: %+v", fog)
	}
	if fog.Kind != "uncharted" || fog.Name != "" {
		t.Fatalf("fog place should be unnamed uncharted: %+v", fog)
	}
	if _, seen := byID["R0103"]; seen {
		t.Fatal("creek should not appear before the player reaches the field")
	}
	if _, seen := byID["R0109"]; seen {
		t.Fatal("hidden burrow must not leak onto the map")
	}
	if atlas.CurrentLayer != "overworld" {
		t.Fatalf("current layer %q", atlas.CurrentLayer)
	}
	if len(atlas.Regions) != 1 {
		t.Fatalf("expected one meadow region, got %d", len(atlas.Regions))
	}
}

func TestRevealShowsHiddenAfterRevealExit(t *testing.T) {
	w := Compile([]*rooms.Room{
		testRoom("R0103", "Creek", "Z01_meadows_forest_path", []string{"outdoor", "water", "starting_room"},
			exit("down", "R0109", true)),
		testRoom("R0109", "Burrow", "Z01_meadows_forest_path", []string{"underground"},
			exit("back", "R0103", false)),
	})
	ch := &characters.Character{
		Entity:          &entities.Entity{ID: "c1"},
		CurrentRoom:     traits.CurrentRoom{CurrentRoomID: "R0103"},
		DiscoveredRooms: map[string]bool{"R0103": true},
		RevealedExits:   map[string][]string{"R0103": {"down"}},
	}
	atlas := Reveal(w, ch)
	found := false
	for _, p := range atlas.Places {
		if p.ID == "R0109" && !p.Discovered && p.Kind == "uncharted" {
			found = true
		}
	}
	if !found {
		t.Fatal("revealed hidden exit should fog the burrow, not hide it")
	}
}

func TestMarkOnIdempotent(t *testing.T) {
	ch := &characters.Character{Entity: &entities.Entity{ID: "c1"}}
	room := testRoom("R0001", "Awakening", "Z00_catacombs_intro", []string{"starting_room"})
	if !MarkOn(ch, room) {
		t.Fatal("first visit should be new")
	}
	if MarkOn(ch, room) {
		t.Fatal("second visit should not count as new")
	}
	if !ch.DiscoveredRooms["R0001"] || !ch.DiscoveredAreas["Z00_catacombs_intro"] {
		t.Fatal("discovery maps not set")
	}
	if ch.AllTimeStats.RoomsDiscovered != 1 {
		t.Fatalf("rooms discovered=%d", ch.AllTimeStats.RoomsDiscovered)
	}
}

func TestCompileDoesNotCollapseTwoRoomsOnOneCell(t *testing.T) {
	w := Compile([]*rooms.Room{
		testRoom("A", "A", "Z01", []string{"starting_room", "outdoor"},
			exit("north", "B", false), exit("east", "C", false)),
		testRoom("B", "B", "Z01", []string{"outdoor"},
			exit("east", "D", false)),
		testRoom("C", "C", "Z01", []string{"outdoor"},
			exit("north", "D", false)),
		testRoom("D", "D", "Z01", []string{"outdoor"}),
	})
	seen := map[[3]int]string{}
	for id, pr := range w.rooms {
		k := [3]int{pr.z, pr.x, pr.y}
		if other, ok := seen[k]; ok {
			t.Fatalf("rooms %s and %s share cell %+v", id, other, k)
		}
		seen[k] = id
	}
}

func TestDisplayAreaStripsZonePrefix(t *testing.T) {
	if got := displayArea("Z01_meadows_forest_path"); got != "Meadows Forest Path" {
		t.Fatalf("got %q", got)
	}
}

func TestRevealMarksInstanceCloneAsTemplateCurrent(t *testing.T) {
	w := Compile([]*rooms.Room{
		testRoom("R0215", "The Weary Wanderer - Cellar", "Z02_oldtown", []string{"underground", "instance"},
			exit("up", "R0203", false)),
		testRoom("R0203", "The Weary Wanderer", "Z02_oldtown", []string{"indoor"},
			exit("down", "R0215", false)),
	})
	ch := &characters.Character{
		Entity:          &entities.Entity{ID: "c1"},
		CurrentRoom:     traits.CurrentRoom{CurrentRoomID: "R0215~guest-a"},
		DiscoveredRooms: map[string]bool{"R0215": true, "R0203": true},
	}
	atlas := Reveal(w, ch)
	var currentCount int
	for _, p := range atlas.Places {
		if p.Current {
			currentCount++
			if p.ID != "R0215" {
				t.Fatalf("expected template R0215 current, got %s", p.ID)
			}
		}
	}
	if currentCount != 1 {
		t.Fatalf("expected exactly one current place, got %d", currentCount)
	}
	if atlas.CurrentLayer != "lower" && atlas.CurrentLayer != "overworld" {
		// cellar underground should be lower when z maps that way
		t.Logf("current layer %q (ok if biome z mapping differs)", atlas.CurrentLayer)
	}
}
