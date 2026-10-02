package worldmap

import (
	"encoding/json"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"testing"
)

func TestTerrainClassificationPrecedence(t *testing.T) {
	cases := []struct {
		name     string
		r        rooms.Room
		want     string
		fallback bool
	}{
		{"structured type wins", rooms.Room{RoomType: "snow", Tags: []string{"forest"}, Area: "Z02_sample_town"}, "snow", false},
		{"area type", rooms.Room{AreaType: "desert"}, "desert", false},
		{"tags beat zone", rooms.Room{Tags: []string{"farmland", "outdoor"}, Area: "Z04_sample_woods"}, "farmland", false},
		{"tag order independent", rooms.Room{Tags: []string{"dungeon", "flooded"}}, "water", false},
		{"inn floor", rooms.Room{Name: "The Roadside Inn", Tags: []string{"indoor", "inn"}, Area: "Z02_sample_town"}, "interior", false},
		{"keep name", rooms.Room{Name: "Sample Keep", Area: "Z02_sample_town"}, "castle", false},
		{"whole words", rooms.Room{Name: "Innkeeper's Lane", Area: "Z02_sample_town"}, "city", false},
		{"underground before forest zone", rooms.Room{Tags: []string{"underground"}, Area: "Z04_sample_woods"}, "dungeon", false},
		{"area fallback", rooms.Room{Area: "Z03_sample_marsh"}, "swamp", false},
		{"description last resort", rooms.Room{Name: "Silent Expanse", Description: "A glacier covers the ground."}, "snow", false},
		{"unknown default", rooms.Room{Name: "Unclassified Expanse"}, "grassland", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, def := ClassifyTerrain(&c.r)
			if got != c.want || def != c.fallback {
				t.Fatalf("got %s/%v; want %s/%v", got, def, c.want, c.fallback)
			}
		})
	}
}
func TestTerrainRevealDoesNotRevealFogArt(t *testing.T) {
	a := &rooms.Room{Entity: &entities.Entity{ID: "A"}, Name: "Town Square", Area: "Z02_sample_town", Exits: &rooms.Exits{{Name: "north", Target: "B"}}}
	b := &rooms.Room{Entity: &entities.Entity{ID: "B"}, Name: "Secret Glacier", RoomType: "snow", Area: "Z02_sample_town"}
	ch := &characters.Character{Entity: &entities.Entity{ID: "test"}, DiscoveredRooms: map[string]bool{"A": true}}
	ch.CurrentRoomID = "A"
	atlas := Reveal(Compile([]*rooms.Room{a, b}), ch)
	for _, p := range atlas.Places {
		if p.ID == "B" {
			if p.Terrain != "fog" || p.Discovered {
				t.Fatalf("fog exposed: %+v", p)
			}
			blob, _ := json.Marshal(p)
			var wire map[string]interface{}
			_ = json.Unmarshal(blob, &wire)
			if wire["terrain"] != "fog" {
				t.Fatal(string(blob))
			}
		}
	}
	ch.DiscoveredRooms["B"] = true
	atlas = Reveal(Compile([]*rooms.Room{a, b}), ch)
	for _, p := range atlas.Places {
		if p.ID == "B" && p.Terrain != "snow" {
			t.Fatalf("charted terrain missing: %+v", p)
		}
	}
}
