package worldmap

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
)

func continentFixture() []*rooms.Room {
	return []*rooms.Room{
		withCoords(testRoom("street", "Gate", "Z02_oldtown", []string{"outdoor", "gate"}, exit("north", "square", false), exit("inside", "inn", false), exit("down", "crypt", true)), 0, 0, 0),
		withCoords(testRoom("square", "Square", "Z02_oldtown", []string{"outdoor"}), 0, -1, 0),
		testRoom("inn", "Inn", "Z02_oldtown", []string{"indoor"}, exit("outside", "street", false), exit("up", "guest", false)),
		withCoords(testRoom("guest", "Guest room", "Z02_oldtown", []string{"indoor"}), 0, 0, 2),
		testRoom("crypt", "Crypt", "Z02_oldtown", []string{"dungeon"}, exit("up", "street", false)),
		withCoords(testRoom("peak", "Peak", "Z09_thornfield_highlands", []string{"outdoor", "mountain"}, exit("south", "ridge", false)), 30, 30, 4),
		withCoords(testRoom("ridge", "Ridge", "Z09_thornfield_highlands", []string{"outdoor", "mountain"}), 30, 31, 3),
		testRoom("forest", "Forest", "unconfigured", []string{"outdoor", "forest"}, exit("east", "street", false)),
	}
}
func TestContinentSurfaceAnchorsAndDepth(t *testing.T) {
	w := Compile(continentFixture())
	for _, id := range []string{"inn", "guest"} {
		p := w.rooms[id]
		anchor := w.rooms["street"]
		if p.layer != "overworld" || p.role != "interior" || p.surfaceID != "street" || p.x != anchor.x || p.y != anchor.y {
			t.Fatalf("interior %s not anchored: %+v", id, p)
		}
	}
	if w.rooms["peak"].layer != "overworld" {
		t.Fatal("positive outdoor elevation must stay on the continent")
	}
	if w.rooms["crypt"].layer != "lower" {
		t.Fatal("crypt must remain Lower")
	}
	if w.rooms["ridge"].x != w.rooms["peak"].x || w.rooms["ridge"].y != w.rooms["peak"].y+1 {
		t.Fatal("zone translation must preserve authored local geometry")
	}
	if w.rooms["forest"].x >= w.rooms["street"].x {
		t.Fatal("unconfigured forest should attach west through its east exit")
	}
}
func TestContinentConnectedDecorativeAndDeterministic(t *testing.T) {
	rs := continentFixture()
	w := Compile(rs)
	cells := map[[2]int]bool{}
	for _, c := range w.landscape {
		cells[[2]int{c.X, c.Y}] = true
	}
	var seed [2]int
	for c := range cells {
		seed = c
		break
	}
	seen := map[[2]int]bool{seed: true}
	q := [][2]int{seed}
	for len(q) > 0 {
		p := q[0]
		q = q[1:]
		for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			next := [2]int{p[0] + d[0], p[1] + d[1]}
			if cells[next] && !seen[next] {
				seen[next] = true
				q = append(q, next)
			}
		}
	}
	if len(seen) != len(cells) {
		t.Fatalf("continent has disconnected ground: %d/%d", len(seen), len(cells))
	}
	for _, p := range w.rooms {
		if p.layer == "overworld" && !cells[[2]int{p.x, p.y}] {
			t.Fatalf("surface %s outside continent", p.id)
		}
	}
	wire, _ := json.Marshal(w.landscape)
	for _, forbidden := range [][]byte{[]byte("street"), []byte("roomId"), []byte("area"), []byte("canTravel")} {
		if bytes.Contains(wire, forbidden) {
			t.Fatalf("decorative ground leaks %s", forbidden)
		}
	}
	ch := &characters.Character{Entity: &entities.Entity{ID: "c"}, CurrentRoom: traits.CurrentRoom{CurrentRoomID: "street"}, DiscoveredRooms: map[string]bool{"street": true, "peak": true}}
	expected, _ := json.Marshal(Reveal(w, ch))
	for i := 0; i < 20; i++ {
		reversed := make([]*rooms.Room, len(rs))
		for j, r := range rs {
			reversed[len(rs)-j-1] = r
		}
		actual, _ := json.Marshal(Reveal(Compile(reversed), ch))
		if !bytes.Equal(expected, actual) {
			t.Fatal("input order changed continent or revealed atlas")
		}
	}
}
func TestContinentFogAndHiddenEntrance(t *testing.T) {
	w := Compile(continentFixture())
	ch := &characters.Character{Entity: &entities.Entity{ID: "c"}, CurrentRoom: traits.CurrentRoom{CurrentRoomID: "inn"}, DiscoveredRooms: map[string]bool{"inn": true}}
	atlas := Reveal(w, ch)
	byID := map[string]Place{}
	for _, p := range atlas.Places {
		byID[p.ID] = p
	}
	if p := byID["street"]; p.Discovered || p.Name != "" || p.Terrain != "fog" {
		t.Fatalf("indoor discovery must not discover street: %+v", p)
	}
	if _, ok := byID["crypt"]; ok {
		t.Fatal("hidden crypt leaked")
	}
	known, fog := 0, 0
	for _, c := range atlas.Landscape {
		if c.Terrain == "fog" {
			fog++
		} else {
			known++
		}
	}
	if known == 0 || fog == 0 {
		t.Fatal("interior discovery should reveal nearby ground but leave other zones fogged")
	}
	ch.CurrentRoomID = "street"
	ch.DiscoveredRooms["street"] = true
	for _, p := range Reveal(w, ch).Places {
		if p.ID == "street" && len(p.Entrances) > 0 {
			t.Fatal("hidden entrance stamp leaked")
		}
	}
	ch.RevealedExits = map[string][]string{"street": {"down"}}
	atlas = Reveal(w, ch)
	for _, p := range atlas.Places {
		if p.ID == "street" && (len(p.Entrances) != 1 || p.Entrances[0] != "crypt") {
			t.Fatal("revealed entrance unavailable")
		}
		if p.ID == "crypt" && (p.Name != "" || p.Discovered || p.Terrain != "fog") {
			t.Fatal("revealed exit must not discover the underground room")
		}
	}
}

func TestDisconnectedSurfaceRoomsDoNotOverlap(t *testing.T) {
	w := Compile([]*rooms.Room{
		testRoom("a", "A", "isolated", []string{"outdoor"}),
		testRoom("b", "B", "isolated", []string{"outdoor"}),
		testRoom("c", "C", "isolated", []string{"outdoor"}),
	})
	seen := map[[2]int]bool{}
	for _, p := range w.rooms {
		key := [2]int{p.x, p.y}
		if seen[key] {
			t.Fatal("disconnected rooms overlap in a single-area world")
		}
		seen[key] = true
	}
}
