package worldmap

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
)

func TestRoomArtCustomization(t *testing.T) {
	for _, tc := range []struct{ key, name, tag string }{
		{"forge", "Ironhand's Forge", "smithy"}, {"shrine", "Shrine of the Twelve", "temple"}, {"tavern", "The Wanderer", "inn"},
		{"farm", "Wheat Fields", "farmland"}, {"tower", "Guard Post", "gate"}, {"keep", "Castle Courtyard", "fortress"},
		{"ruins", "Broken Walls", "ruined"}, {"graveyard", "Burial Ground", "graveyard"}, {"dock", "River Docks", "pier"},
		{"mine", "Mine Adit", "mine"}, {"magic", "Standing Stone", "arcane"}, {"flowers", "Wildflower Field", "garden"},
	} {
		r := testRoom("room", tc.name, "zone", []string{tc.tag})
		features, seed := ClassifyArt(r)
		if !strings.Contains("|"+strings.Join(features, "|")+"|", "|"+tc.key+"|") {
			t.Errorf("%s missing from %+v", tc.key, features)
		}
		_, again := ClassifyArt(r)
		if seed == "" || seed != again {
			t.Fatal("art seed not deterministic")
		}
		r.Description = "New room customization"
		_, changed := ClassifyArt(r)
		if seed == changed {
			t.Fatal("customization should affect variation seed")
		}
	}
	r := testRoom("r", "Ordinary room", "zone", nil)
	r.RoomType = "smithy"
	features, _ := ClassifyArt(r)
	if features[0] != "forge" {
		t.Fatal("structured room type ignored")
	}
	r.RoomType = ""
	r.Description = "A magical ley line circles the standing stone."
	features, _ = ClassifyArt(r)
	if features[0] != "magic" {
		t.Fatal("descriptive fallback ignored")
	}
	r.Name = "Tavern"
	features, _ = ClassifyArt(r)
	if strings.Contains(strings.Join(features, " "), "magic") {
		t.Fatal("description must not override an explicit room feature")
	}
	r.Name = "Workshop"
	r.Description = ""
	actions := rooms.Actions{{Name: "shop", ScriptId: "private-script", Params: map[string]interface{}{"secret": "not-exported"}}}
	r.Actions = &actions
	features, _ = ClassifyArt(r)
	if len(features) != 1 || features[0] != "shop" {
		t.Fatal("service action name ignored")
	}
}
func TestArtHintsOnlyDiscovered(t *testing.T) {
	rs := []*rooms.Room{
		testRoom("start", "Field", "zone", []string{"outdoor"}, exit("north", "shop", false)),
		testRoom("shop", "Secret Forge", "zone", []string{"indoor", "smithy"}),
	}
	ch := &characters.Character{Entity: &entities.Entity{ID: "player"}, CurrentRoom: traits.CurrentRoom{CurrentRoomID: "start"}}
	atlas := Reveal(Compile(rs), ch)
	for _, p := range atlas.Places {
		if p.ID == "shop" && (len(p.MapFeatures) > 0 || p.ArtSeed != "" || p.UndergroundStyle != "") {
			t.Fatal("unknown room art leaks through fog")
		}
	}
	ch.DiscoveredRooms = map[string]bool{"shop": true}
	atlas = Reveal(Compile(rs), ch)
	for _, p := range atlas.Places {
		if p.ID == "shop" && (p.ArtSeed == "" || len(p.MapFeatures) == 0 || p.MapFeatures[0] != "forge") {
			t.Fatal("charted room art missing")
		}
	}
	b, _ := json.Marshal(atlas)
	if strings.Contains(string(b), "private-script") {
		t.Fatal("scripts must not be atlas art inputs")
	}
}
func TestUndergroundArtStyles(t *testing.T) {
	for _, tc := range []struct{ name, want string }{{"Mine Gallery", "cave"}, {"Crypt Landing", "crypt"}, {"Inn Cellar", "cellar"}, {"Sewer Landing", "sewer"}} {
		r := testRoom("r", tc.name, "zone", []string{"underground"})
		if got := undergroundStyle(r); got != tc.want {
			t.Fatalf("%s style %s want %s", tc.name, got, tc.want)
		}
	}
}
