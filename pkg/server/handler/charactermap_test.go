package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/worldmap"
)

func TestGetCharacterMapRejectsOtherPlayers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	facade := testFacade(t)
	owner := testUser("owner-user", "owner-ref", entities.RolePlayer)
	other := testUser("other-user", "other-ref", entities.RolePlayer)
	for _, user := range []*entities.User{owner, other} {
		if _, err := facade.UsersService().Create(user); err != nil {
			t.Fatalf("create user: %v", err)
		}
	}
	character := &characters.Character{
		Entity:      &entities.Entity{ID: "char-map"},
		Name:        "Mapper",
		BelongsUser: *traits.BelongsToUser(owner.ID),
	}
	if _, err := facade.CharactersService().Store(character); err != nil {
		t.Fatalf("store character: %v", err)
	}

	h := &CharacterMapHandler{
		Characters: facade.CharactersService(),
		Rooms:      facade.RoomsService(),
	}
	params := gin.Params{{Key: "id", Value: character.ID}}
	rec := performHandlerRequest(http.MethodGet, "/api/characters/"+character.ID+"/map", nil, other, params, h.GetCharacterMap)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGetCharacterMapReturnsDiscoveredAtlas(t *testing.T) {
	gin.SetMode(gin.TestMode)
	facade := testFacade(t)
	owner := testUser("owner-user", "owner-ref", entities.RolePlayer)
	if _, err := facade.UsersService().Create(owner); err != nil {
		t.Fatalf("create user: %v", err)
	}

	exitsA := rooms.Exits{
		{Name: "north", Target: "R0102", Type: rooms.RoomExitTypeDirection},
	}
	exitsB := rooms.Exits{
		{Name: "south", Target: "R0101", Type: rooms.RoomExitTypeDirection},
	}
	for _, room := range []*rooms.Room{
		{
			Entity: &entities.Entity{ID: "R0101"}, Name: "Meadow", Area: "Z01_meadows_forest_path",
			Tags: []string{"outdoor", "starting_room"}, Exits: &exitsA,
		},
		{
			Entity: &entities.Entity{ID: "R0102"}, Name: "Field", Area: "Z01_meadows_forest_path",
			Tags: []string{"outdoor"}, Exits: &exitsB,
		},
	} {
		if _, err := facade.RoomsService().Store(room); err != nil {
			t.Fatalf("store room %s: %v", room.ID, err)
		}
	}

	character := &characters.Character{
		Entity:          &entities.Entity{ID: "char-map"},
		Name:            "Mapper",
		BelongsUser:     *traits.BelongsToUser(owner.ID),
		CurrentRoom:     traits.CurrentRoom{CurrentRoomID: "R0101"},
		DiscoveredRooms: map[string]bool{"R0101": true},
	}
	if _, err := facade.CharactersService().Store(character); err != nil {
		t.Fatalf("store character: %v", err)
	}

	h := &CharacterMapHandler{
		Characters: facade.CharactersService(),
		Rooms:      facade.RoomsService(),
	}
	params := gin.Params{{Key: "id", Value: character.ID}}
	rec := performHandlerRequest(http.MethodGet, "/api/characters/"+character.ID+"/map", nil, owner, params, h.GetCharacterMap)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var atlas worldmap.PlayerMap
	if err := json.Unmarshal(rec.Body.Bytes(), &atlas); err != nil {
		t.Fatalf("decode atlas: %v", err)
	}
	if atlas.CurrentRoomID != "R0101" || atlas.CurrentLayer == "" {
		t.Fatalf("unexpected atlas header: %+v", atlas)
	}
	foundCurrent, foundFog := false, false
	for _, p := range atlas.Places {
		if p.ID == "R0101" && p.Discovered && p.Current {
			foundCurrent = true
		}
		if p.ID == "R0102" && !p.Discovered {
			foundFog = true
		}
	}
	if !foundCurrent || !foundFog {
		t.Fatalf("expected current meadow and fog field, places=%+v", atlas.Places)
	}
}
