package contenthealth

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

func TestMapOverlaysLevelsAggroAndSpawners(t *testing.T) {
	guard := overlayNPC("GUARD", "Guard", "R0", 2, false)
	brute := overlayNPC("BRUTE", "Brute", "R0", 6, true)
	rat := overlayNPC("RAT", "Rat", "", 8, true)
	rat.RespawnTime = 2 * time.Minute
	boss := overlayNPC("BOSS", "Boss", "", 30, true)
	boss.RespawnTime = time.Minute
	delay := 30 * time.Minute
	override := 10 * time.Minute
	world := World{
		Rooms: []*rooms.Room{
			overlayArt(overlayArea(testRoom("R0", "Square"), "Town"), "square"),
			overlayArea(testRoom("R1", "Lane"), "Town"),
			overlayArea(testRoom("R2", "Nest"), "Wild"),
			testRoom("R3", "Pit"),
			overlayArt(testRoom("R4", "Quiet"), "  "),
		},
		NPCs: []*npc.NPC{guard, brute, rat, boss},
		Spawners: []*npc.NPCSpawner{
			{Entity: &entities.Entity{ID: "SP-RAT"}, Name: "Rat hole", TemplateID: "RAT", RoomID: "R2", MaxInstances: 3, RespawnTimeOverride: &override},
			{Entity: &entities.Entity{ID: "SP-BOSS"}, TemplateID: "BOSS", RoomID: "R3", MaxInstances: 1, RespawnDelay: delay},
			{Entity: &entities.Entity{ID: "SP-PLAIN"}, TemplateID: "RAT", RoomID: "R4", InitialCount: 2},
		},
	}
	residents := rooms.NPCs{"GUARD"}
	world.Rooms[0].NPCs = &residents

	view := MapOverlayView(world, OverlayLive{}, false)
	byID := overlayByID(t, view.Rooms)
	if view.Admin || view.Live {
		t.Fatalf("flags = admin %v live %v", view.Admin, view.Live)
	}
	square := byID["R0"]
	if square.LevelMin != 2 || square.LevelMax != 6 || square.LevelBand != "1" || square.LevelSource != "room" || square.Aggro != 1 || square.MissingArt {
		t.Fatalf("square = %+v", square)
	}
	lane := byID["R1"]
	if lane.LevelMin != 2 || lane.LevelMax != 6 || lane.LevelBand != "1" || lane.LevelSource != "zone" || !lane.MissingArt {
		t.Fatalf("lane = %+v", lane)
	}
	nest := byID["R2"]
	if nest.LevelMin != 8 || nest.LevelMax != 8 || nest.LevelBand != "2" || nest.LevelSource != "room" || nest.Aggro != 3 {
		t.Fatalf("nest = %+v", nest)
	}
	if len(nest.Spawners) != 1 || nest.Spawners[0].RespawnTime != "10m0s" || nest.Spawners[0].Name != "Rat hole" {
		t.Fatalf("nest spawners = %+v", nest.Spawners)
	}
	pit := byID["R3"]
	if pit.LevelBand != "6" || pit.Aggro != 1 || len(pit.Spawners) != 1 || pit.Spawners[0].RespawnTime != "30m0s" || pit.Spawners[0].Name != "Boss" {
		t.Fatalf("pit = %+v", pit)
	}
	quiet := byID["R4"]
	if !quiet.MissingArt || quiet.Aggro != 2 || len(quiet.Spawners) != 1 || quiet.Spawners[0].RespawnTime != "2m0s" {
		t.Fatalf("quiet = %+v", quiet)
	}
	if _, ok := byID["NOPE"]; ok {
		t.Fatal("unknown room was emitted")
	}
}

func TestMapOverlaysQuestsArtAndCopies(t *testing.T) {
	guard := overlayNPC("GUARD", "Guard", "R0", 4, false)
	rat := overlayNPC("RAT", "Rat", "", 12, true)
	world := World{
		Rooms: []*rooms.Room{
			overlayArt(testRoom("R0", "Square"), "square"),
			testRoom("R1", "Cellar"),
			testRoom("R2", "Nest"),
		},
		NPCs: []*npc.NPC{guard, rat},
		Spawners: []*npc.NPCSpawner{
			{Entity: &entities.Entity{ID: "SP1"}, TemplateID: "RAT", RoomID: "R2", MaxInstances: 1},
		},
		Quests: []*quests.Quest{
			{Entity: &entities.Entity{ID: "Q2"}, Name: "Rats", Objectives: []quests.Objective{
				{ID: "k", Type: quests.ObjectiveKill, TargetID: "RAT"},
			}},
			{Entity: &entities.Entity{ID: "Q1"}, Name: "Visit", Objectives: []quests.Objective{
				{ID: "v", Type: quests.ObjectiveVisit, TargetID: "R1"},
				{ID: "t", Type: quests.ObjectiveTalk, TargetID: "GUARD"},
			}},
			{Entity: &entities.Entity{ID: "Q3"}, Name: "Parcel", Objectives: []quests.Objective{
				{ID: "d", Type: quests.ObjectiveDeliver, TargetID: "PARCEL", DeliverToNPCID: "GUARD"},
			}},
		},
	}
	view := MapOverlayView(world, OverlayLive{
		Available: true,
		CloneIDs:  []string{"R2~a", "R2~b", "R2~a", "R2", "nope", "R9~z"},
	}, false)
	byID := overlayByID(t, view.Rooms)
	if !view.Live {
		t.Fatal("live flag")
	}
	if join(byID["R0"].Quests) != "Q1,Q3" || byID["R0"].MissingArt {
		t.Fatalf("square quests = %+v", byID["R0"])
	}
	if join(byID["R1"].Quests) != "Q1" || !byID["R1"].MissingArt {
		t.Fatalf("cellar = %+v", byID["R1"])
	}
	if join(byID["R2"].Quests) != "Q2" || byID["R2"].Copies != 2 {
		t.Fatalf("nest = %+v", byID["R2"])
	}
	if _, ok := byID["R2~a"]; ok {
		t.Fatal("a clone with nobody in it should stay off the overlay")
	}
	if _, ok := byID["R9"]; ok {
		t.Fatal("unknown template should not gain a copy count")
	}
}

func TestMapOverlaysPlayersHideNamesFromCreators(t *testing.T) {
	world := World{Rooms: []*rooms.Room{
		overlayArt(testRoom("R0", "Square"), "square"),
		testRoom("R2", "Nest"),
	}}
	live := OverlayLive{
		Available: true,
		Players: []OverlayPlayer{
			{ID: "c1", Name: "Ada", RoomID: "R0"},
			{ID: "c1", Name: "Ada", RoomID: "R0"},
			{ID: "c2", Name: "Bo", RoomID: "R2~a"},
			{ID: "", Name: "", RoomID: ""},
		},
		CloneIDs: []string{"R2~a"},
	}

	creator := MapOverlayView(world, live, false)
	body, err := json.Marshal(creator)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "Ada") || strings.Contains(string(body), "playerNames") || strings.Contains(string(body), "Bo") {
		t.Fatalf("creator payload leaked names: %s", body)
	}
	byID := overlayByID(t, creator.Rooms)
	if byID["R0"].Players != 1 || byID["R2~a"].Players != 1 || byID["R2"].Players != 0 || byID["R2"].Copies != 1 {
		t.Fatalf("creator rows = %+v", byID)
	}

	admin := MapOverlayView(world, live, true)
	if !admin.Admin {
		t.Fatal("admin flag")
	}
	adminBody, err := json.Marshal(admin)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(adminBody), "Ada") || !strings.Contains(string(adminBody), "Bo") {
		t.Fatalf("admin payload = %s", adminBody)
	}
	byID = overlayByID(t, admin.Rooms)
	if join(byID["R0"].PlayerNames) != "Ada" || join(byID["R2~a"].PlayerNames) != "Bo" {
		t.Fatalf("admin names = %+v", byID)
	}
}

func TestLevelBandMidpoint(t *testing.T) {
	cases := []struct {
		min, max int32
		want     string
	}{
		{4, 4, "1"},
		{2, 6, "1"},
		{5, 5, "2"},
		{1, 20, "3"},
		{19, 19, "4"},
		{20, 29, "5"},
		{30, 30, "6"},
	}
	for _, tc := range cases {
		if got := levelBand(tc.min, tc.max); got != tc.want {
			t.Fatalf("levelBand(%d,%d) = %s, want %s", tc.min, tc.max, got, tc.want)
		}
	}
}

func overlayNPC(id, name, room string, level int32, aggro bool) *npc.NPC {
	n := &npc.NPC{
		Entity:      &entities.Entity{ID: id},
		Name:        name,
		Level:       level,
		SpawnRoomID: room,
	}
	if aggro {
		n.EnemyTrait = &npc.EnemyTrait{AggroOnSight: true}
	}
	return n
}

func overlayArea(room *rooms.Room, area string) *rooms.Room {
	room.Area = area
	return room
}

func overlayArt(room *rooms.Room, background string) *rooms.Room {
	room.Meta = &struct {
		Mood       string `bson:"mood,omitempty" json:"mood,omitempty"`
		Background string `bson:"background,omitempty" json:"background,omitempty"`
	}{Background: background}
	return room
}

func overlayByID(t *testing.T, rooms []OverlayRoom) map[string]OverlayRoom {
	t.Helper()
	out := map[string]OverlayRoom{}
	var prev string
	for _, room := range rooms {
		if _, ok := out[room.ID]; ok {
			t.Fatalf("duplicate room %s", room.ID)
		}
		if prev != "" && room.ID < prev {
			t.Fatalf("rooms not sorted at %s", room.ID)
		}
		prev = room.ID
		out[room.ID] = room
	}
	return out
}

func join(ids []string) string {
	return strings.Join(ids, ",")
}
