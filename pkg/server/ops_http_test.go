package server_test

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
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/server"
	"github.com/talesmud/talesmud/pkg/server/handler"
	"github.com/talesmud/talesmud/pkg/service"
)

func newOpsHTTP(t *testing.T) (*gin.Engine, *game.Game, service.Facade) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	client, err := sqlite.Open(filepath.Join(t.TempDir(), "ops.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	g := game.New(facade)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		role := c.GetHeader("X-Role")
		if role == "" {
			role = entities.RoleAdmin
		}
		c.Set("user", &entities.User{
			Entity:   &entities.Entity{ID: "admin-1"},
			Nickname: "Admin",
			Role:     role,
		})
		c.Next()
	})
	r.POST("/api/ops/:action", server.AdminMiddleware(), handler.OpsAction(g, facade.AuditService()))
	r.POST("/api/audit/:id/undo", server.AdminMiddleware(), handler.UndoAudit(g, facade.AuditService()))
	r.GET("/api/audit", handler.ListAudit(facade.AuditService()))
	r.PUT("/api/rooms/:id", handler.AuditWrites(facade.AuditService()), func(c *gin.Context) {
		var room rooms.Room
		if err := c.ShouldBindJSON(&room); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if room.Entity == nil {
			room.Entity = &entities.Entity{ID: c.Param("id")}
		}
		room.ID = c.Param("id")
		if err := facade.RoomsService().Update(c.Param("id"), &room); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "updated room"})
	})
	r.DELETE("/api/rooms/:id", handler.AuditWrites(facade.AuditService()), func(c *gin.Context) {
		if err := facade.RoomsService().Delete(c.Param("id")); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted room"})
	})
	return r, g, facade
}

func doJSON(r http.Handler, method, path, role string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if role != "" {
		req.Header.Set("X-Role", role)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestOpsRejectPlayerAndMissingConfirm(t *testing.T) {
	r, _, _ := newOpsHTTP(t)
	actions := []string{
		"teleport", "give-item", "take-item", "npc-heal", "npc-respawn", "npc-despawn",
		"end-combat", "instance-cleanup", "quest-step", "regrant-starter-kit",
	}
	for _, action := range actions {
		rec := doJSON(r, http.MethodPost, "/api/ops/"+action, entities.RolePlayer, map[string]interface{}{"confirm": true})
		if rec.Code != http.StatusForbidden {
			t.Fatalf("%s player: %d %s", action, rec.Code, rec.Body.String())
		}
		rec = doJSON(r, http.MethodPost, "/api/ops/"+action, entities.RoleAdmin, map[string]interface{}{"characterId": "x"})
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s confirm: %d %s", action, rec.Code, rec.Body.String())
		}
	}
}

func TestOpsGiveUndoAndConflictOverHTTP(t *testing.T) {
	r, _, facade := newOpsHTTP(t)
	exits := rooms.Exits{}
	if _, err := facade.RoomsService().Import(&rooms.Room{Entity: &entities.Entity{ID: "room-a"}, Name: "Alpha", Exits: &exits}); err != nil {
		t.Fatal(err)
	}
	if _, err := facade.CharactersService().Import(&characters.Character{
		Entity:      &entities.Entity{ID: "char-1"},
		Name:        "Hero",
		BelongsUser: *traits.BelongsToUser("user-1"),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "room-a"},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := facade.ItemsService().Import(&items.Item{
		Entity: &entities.Entity{ID: "tpl-bread"}, Name: "Bread", IsTemplate: true,
	}); err != nil {
		t.Fatal(err)
	}
	rec := doJSON(r, http.MethodPost, "/api/ops/give-item", entities.RoleAdmin, map[string]interface{}{
		"confirm": true, "characterId": "char-1", "itemTemplateId": "tpl-bread", "quantity": 1,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("give %d %s", rec.Code, rec.Body.String())
	}
	var given struct {
		AuditID  string `json:"auditId"`
		Undoable bool   `json:"undoable"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &given); err != nil || given.AuditID == "" || !given.Undoable {
		t.Fatalf("body %s", rec.Body.String())
	}
	got, _ := facade.CharactersService().FindByID("char-1")
	if got.Inventory.Count() != 1 {
		t.Fatalf("count %d", got.Inventory.Count())
	}
	// Change the character so the inverse no longer matches.
	got.Inventory.Items = nil
	if err := facade.CharactersService().Update(got.ID, got); err != nil {
		t.Fatal(err)
	}
	rec = doJSON(r, http.MethodPost, "/api/audit/"+given.AuditID+"/undo", entities.RoleAdmin, nil)
	if rec.Code != http.StatusConflict {
		t.Fatalf("conflict %d %s", rec.Code, rec.Body.String())
	}
	// Give again and undo cleanly.
	rec = doJSON(r, http.MethodPost, "/api/ops/give-item", entities.RoleAdmin, map[string]interface{}{
		"confirm": true, "characterId": "char-1", "itemTemplateId": "tpl-bread", "quantity": 1,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("second give %d %s", rec.Code, rec.Body.String())
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &given)
	rec = doJSON(r, http.MethodPost, "/api/audit/"+given.AuditID+"/undo", entities.RoleAdmin, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("undo %d %s", rec.Code, rec.Body.String())
	}
	got, _ = facade.CharactersService().FindByID("char-1")
	if got.Inventory.Count() != 0 {
		t.Fatalf("after undo %d", got.Inventory.Count())
	}
}

func TestAuditMiddlewareRestoresRoomUpdateAndDelete(t *testing.T) {
	r, _, facade := newOpsHTTP(t)
	original := &rooms.Room{Entity: &entities.Entity{ID: "room-1"}, Name: "Old"}
	if _, err := facade.RoomsService().Store(original); err != nil {
		t.Fatal(err)
	}
	rec := doJSON(r, http.MethodPut, "/api/rooms/room-1", entities.RoleAdmin, &rooms.Room{
		Entity: &entities.Entity{ID: "room-1"}, Name: "New",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("put %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(r, http.MethodGet, "/api/audit?entityType=rooms&entityId=room-1", "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list %d %s", rec.Code, rec.Body.String())
	}
	var rows []struct {
		ID     string `json:"id"`
		Action string `json:"action"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &rows); err != nil || len(rows) != 1 || rows[0].Action != "update" {
		t.Fatalf("rows %s %v", rec.Body.String(), err)
	}
	rec = doJSON(r, http.MethodPost, "/api/audit/"+rows[0].ID+"/undo", entities.RoleAdmin, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("undo %d %s", rec.Code, rec.Body.String())
	}
	got, err := facade.RoomsService().FindByID("room-1")
	if err != nil || got.Name != "Old" {
		t.Fatalf("restored %+v %v", got, err)
	}

	rec = doJSON(r, http.MethodDelete, "/api/rooms/room-1", entities.RoleAdmin, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(r, http.MethodGet, "/api/audit?entityType=rooms&entityId=room-1&limit=5", "", nil)
	_ = json.Unmarshal(rec.Body.Bytes(), &rows)
	var deleteID string
	for _, row := range rows {
		if row.Action == "delete" {
			deleteID = row.ID
		}
	}
	if deleteID == "" {
		t.Fatalf("no delete row %s", rec.Body.String())
	}
	rec = doJSON(r, http.MethodPost, "/api/audit/"+deleteID+"/undo", entities.RoleAdmin, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("undo delete %d %s", rec.Code, rec.Body.String())
	}
	if _, err := facade.RoomsService().FindByID("room-1"); err != nil {
		t.Fatalf("room not restored: %v", err)
	}
}
