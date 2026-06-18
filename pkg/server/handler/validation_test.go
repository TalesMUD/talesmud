package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

func TestValidationHandlerRejectsUnknownEntityType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Params = gin.Params{{Key: "entityType", Value: "bad"}}

	h := &ValidationHandler{}
	h.ValidateEntity(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestRoomSaveRejectsBrokenExit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	facade := testFacade(t)
	exits := rooms.Exits{{Name: "north", Target: "missing-room"}}
	room := &rooms.Room{
		Entity: &entities.Entity{ID: "room-a"},
		Name:   "Broken Room",
		Exits:  &exits,
	}

	h := &RoomsHandler{
		Service: facade.RoomsService(),
		Facade:  facade,
	}

	rec := performHandlerRequest(http.MethodPost, "/api/rooms", room, nil, nil, h.PostRoom)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}
}
