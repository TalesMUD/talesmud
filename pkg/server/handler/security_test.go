package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/scripts/runner"
	"github.com/talesmud/talesmud/pkg/service"
)

func testFacade(t *testing.T) service.Facade {
	t.Helper()

	client, err := sqlite.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	return service.NewFacade(repository.NewSQLiteFactory(client), runner.NewMultiRunner())
}

func testUser(id, refID, role string) *entities.User {
	return &entities.User{
		Entity:   &entities.Entity{ID: id},
		RefID:    refID,
		Nickname: id,
		Role:     role,
	}
}

func performHandlerRequest(method, path string, body interface{}, user *entities.User, params gin.Params, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	var payload *bytes.Reader
	if body == nil {
		payload = bytes.NewReader(nil)
	} else if raw, ok := body.(string); ok {
		payload = bytes.NewReader([]byte(raw))
	} else {
		data, _ := json.Marshal(body)
		payload = bytes.NewReader(data)
	}

	req := httptest.NewRequest(method, path, payload)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req
	ctx.Params = params
	if user != nil {
		ctx.Set("user", user)
		ctx.Set("userid", user.RefID)
	}
	handler(ctx)
	return rec
}

func TestImportValidatesJSONBeforeDroppingData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	facade := testFacade(t)
	room := &rooms.Room{
		Entity: &entities.Entity{ID: "room-1"},
		Name:   "Preserved Room",
	}
	if _, err := facade.RoomsService().Store(room); err != nil {
		t.Fatalf("store room: %v", err)
	}

	h := &ExportHandler{
		RoomsService:      facade.RoomsService(),
		CharactersService: facade.CharactersService(),
		UserService:       facade.UsersService(),
		ItemsService:      facade.ItemsService(),
		ScriptService:     facade.ScriptsService(),
		NPCsService:       facade.NPCsService(),
		DialogsService:    facade.DialogsService(),
		PartiesService:    facade.PartiesService(),
	}

	rec := performHandlerRequest(http.MethodPost, "/admin/import", "{", nil, nil, h.Import)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", rec.Code)
	}

	if _, err := facade.RoomsService().FindByID(room.ID); err != nil {
		t.Fatalf("room was dropped before import body was validated: %v", err)
	}
}

func TestUpdateUserUsesAuthenticatedRefIDAndPreservesRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	facade := testFacade(t)

	user := testUser("user-1", "auth0|user-1", entities.RolePlayer)
	user.IsNewUser = true
	if _, err := facade.UsersService().Create(user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	h := &UsersHandler{Service: facade.UsersService()}
	rec := performHandlerRequest(
		http.MethodPut,
		"/api/user",
		gin.H{
			"name":      "Marcus",
			"nickname":  "Marcus",
			"isNewUser": false,
			"role":      entities.RoleAdmin,
		},
		user,
		nil,
		h.UpdateUser,
	)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	stored, err := facade.UsersService().FindByRefID(user.RefID)
	if err != nil {
		t.Fatalf("load updated user: %v", err)
	}
	if stored.Name != "Marcus" || stored.Nickname != "Marcus" {
		t.Fatalf("expected profile fields to update, got name=%q nickname=%q", stored.Name, stored.Nickname)
	}
	if stored.IsNewUser {
		t.Fatal("expected isNewUser to be cleared")
	}
	if stored.Role != entities.RolePlayer {
		t.Fatalf("expected role to stay %q, got %q", entities.RolePlayer, stored.Role)
	}
}

func TestCharacterHandlersRejectCrossUserAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	facade := testFacade(t)
	owner := testUser("owner-user", "owner-ref", entities.RolePlayer)
	other := testUser("other-user", "other-ref", entities.RolePlayer)
	for _, user := range []*entities.User{owner, other} {
		if _, err := facade.UsersService().Create(user); err != nil {
			t.Fatalf("create user %s: %v", user.ID, err)
		}
	}

	character := &characters.Character{
		Entity:           &entities.Entity{ID: "char-1"},
		Name:             "Owner Character",
		BelongsUser:      *traits.BelongsToUser(owner.ID),
		MaxHitPoints:     10,
		CurrentHitPoints: 10,
	}
	if _, err := facade.CharactersService().Store(character); err != nil {
		t.Fatalf("store character: %v", err)
	}

	h := &CharactersHandler{Service: facade.CharactersService()}
	params := gin.Params{{Key: "id", Value: character.ID}}

	tests := []struct {
		name    string
		method  string
		body    interface{}
		handler gin.HandlerFunc
	}{
		{name: "get", method: http.MethodGet, handler: h.GetCharacterByID},
		{name: "update", method: http.MethodPut, body: character, handler: h.UpdateCharacterByID},
		{name: "delete", method: http.MethodDelete, handler: h.DeleteCharacterByID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := performHandlerRequest(tt.method, "/api/characters/"+character.ID, tt.body, other, params, tt.handler)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestGetCharactersReturnsOnlyOwnCharactersForPlayers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	facade := testFacade(t)
	owner := testUser("owner-user", "owner-ref", entities.RolePlayer)
	other := testUser("other-user", "other-ref", entities.RolePlayer)

	owned := &characters.Character{
		Entity:      &entities.Entity{ID: "owned-char"},
		Name:        "Owned Character",
		BelongsUser: *traits.BelongsToUser(owner.ID),
	}
	foreign := &characters.Character{
		Entity:      &entities.Entity{ID: "foreign-char"},
		Name:        "Foreign Character",
		BelongsUser: *traits.BelongsToUser(other.ID),
	}
	if _, err := facade.CharactersService().Store(owned); err != nil {
		t.Fatalf("store owned character: %v", err)
	}
	if _, err := facade.CharactersService().Store(foreign); err != nil {
		t.Fatalf("store foreign character: %v", err)
	}

	h := &CharactersHandler{Service: facade.CharactersService()}
	rec := performHandlerRequest(http.MethodGet, "/api/characters", nil, owner, nil, h.GetCharacters)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var got []characters.Character
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 1 || got[0].ID != owned.ID {
		t.Fatalf("expected only owned character %q, got %#v", owned.ID, got)
	}
}

func TestPostCharacterRejectsCreatingForAnotherUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	facade := testFacade(t)
	owner := testUser("owner-user", "owner-ref", entities.RolePlayer)
	other := testUser("other-user", "other-ref", entities.RolePlayer)

	body := &characters.Character{
		Entity:      &entities.Entity{ID: "char-2"},
		Name:        "Forged Character",
		BelongsUser: *traits.BelongsToUser(owner.ID),
	}

	h := &CharactersHandler{Service: facade.CharactersService()}
	rec := performHandlerRequest(http.MethodPost, "/api/characters", body, other, nil, h.PostCharacter)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestQuestProgressHandlersRejectCrossUserCharacterAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	facade := testFacade(t)
	owner := testUser("owner-user", "owner-ref", entities.RolePlayer)
	other := testUser("other-user", "other-ref", entities.RolePlayer)

	character := &characters.Character{
		Entity:      &entities.Entity{ID: "char-quest"},
		Name:        "Quest Owner",
		BelongsUser: *traits.BelongsToUser(owner.ID),
	}
	if _, err := facade.CharactersService().Store(character); err != nil {
		t.Fatalf("store character: %v", err)
	}
	quest := &quests.Quest{
		Entity:      &entities.Entity{ID: "quest-1"},
		Name:        "Simple Quest",
		Description: "A simple quest",
		Source:      quests.QuestSource{Type: "auto"},
		Objectives: []quests.Objective{
			{
				ID:          "visit-room",
				Type:        quests.ObjectiveVisit,
				Description: "Visit the test room",
				TargetID:    "room-quest",
			},
		},
	}
	if _, err := facade.RoomsService().Store(&rooms.Room{Entity: &entities.Entity{ID: "room-quest"}, Name: "Quest Room"}); err != nil {
		t.Fatalf("store quest room: %v", err)
	}
	if _, err := facade.QuestsService().Store(quest); err != nil {
		t.Fatalf("store quest: %v", err)
	}
	if _, err := facade.QuestsService().AcceptQuest(character.ID, quest.ID); err != nil {
		t.Fatalf("accept quest setup: %v", err)
	}

	h := &QuestsHandler{
		Service:           facade.QuestsService(),
		CharactersService: facade.CharactersService(),
	}
	params := gin.Params{
		{Key: "characterId", Value: character.ID},
		{Key: "questId", Value: quest.ID},
	}

	tests := []struct {
		name    string
		method  string
		path    string
		handler gin.HandlerFunc
	}{
		{name: "log", method: http.MethodGet, path: "/api/quest-progress/" + character.ID, handler: h.GetQuestLog},
		{name: "accept", method: http.MethodPost, path: "/api/quest-progress/" + character.ID + "/accept/" + quest.ID, handler: h.AcceptQuest},
		{name: "abandon", method: http.MethodPost, path: "/api/quest-progress/" + character.ID + "/abandon/" + quest.ID, handler: h.AbandonQuest},
		{name: "complete", method: http.MethodPost, path: "/api/quest-progress/" + character.ID + "/complete/" + quest.ID, handler: h.CompleteQuest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := performHandlerRequest(tt.method, tt.path, nil, other, params, tt.handler)
			if rec.Code != http.StatusForbidden {
				t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}
