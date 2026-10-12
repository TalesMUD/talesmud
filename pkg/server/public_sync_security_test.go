package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/authsession"
	"github.com/talesmud/talesmud/pkg/entities"
	mud "github.com/talesmud/talesmud/pkg/mudserver"
)

// Exercise the mounted routes: testing CreatorMiddleware alone cannot catch a
// definition endpoint accidentally mounted in the player group during a sync.
func TestPublicSyncWorldDefinitionsRequireCreator(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const secret = "public-sync-test-secret"
	t.Setenv("GUEST_SECRET", secret)
	facade := testFacade(t)
	sessions := authsession.New(secret, facade.UsersService())
	application := &app{
		Router: gin.New(), Facade: facade, mud: mud.New(facade, nil),
		tickets: NewTicketStore(), sessions: sessions,
	}
	application.setupRoutes()

	definitions := []string{
		"rooms", "rooms-vh", "rooms/missing", "items", "items/missing",
		"scripts", "script-types", "world/graph", "world/rooms-minimal",
		"balance/enemy-scaling", "npcs", "npcs/templates", "npcs/missing",
		"spawners", "spawners/missing", "dialogs", "dialogs/missing",
		"loottables", "loottables/missing", "backgrounds", "settings",
		"quests", "quests/missing", "skills", "skills/missing",
	}
	for _, role := range []string{"player", "guest", "creator", "admin"} {
		t.Run(role, func(t *testing.T) {
			user := &entities.User{
				Entity: entities.NewEntity(), RefID: "auth0|sync-" + role,
				Role: entities.RolePlayer, IsGuest: role == "guest",
				GuestExpiresAt: time.Now().Add(time.Hour),
			}
			if role == "creator" {
				user.Role = entities.RoleCreator
			} else if role == "admin" {
				user.Role = entities.RoleAdmin
			}
			if _, err := facade.UsersService().Create(user); err != nil {
				t.Fatal(err)
			}
			var token string
			if user.IsGuest {
				token = signGuestUID(t, secret, user.ID)
			} else {
				var err error
				token, _, err = sessions.Issue(user.RefID)
				if err != nil {
					t.Fatal(err)
				}
			}
			request := func(method, path string) *httptest.ResponseRecorder {
				req := httptest.NewRequest(method, "/api/"+path, strings.NewReader(`{}`))
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("Content-Type", "application/json")
				rec := httptest.NewRecorder()
				application.Router.ServeHTTP(rec, req)
				return rec
			}
			if got := request(http.MethodGet, "my-characters"); got.Code != http.StatusOK {
				t.Fatalf("own character list: %d %s", got.Code, got.Body.String())
			}
			if user.IsCreator() {
				for _, path := range []string{"rooms", "scripts"} {
					if got := request(http.MethodGet, path); got.Code != http.StatusOK {
						t.Fatalf("creator GET %s: %d %s", path, got.Code, got.Body.String())
					}
				}
				return
			}
			for _, path := range definitions {
				if got := request(http.MethodGet, path); got.Code != http.StatusForbidden {
					t.Errorf("player GET %s: expected 403, got %d", path, got.Code)
				}
			}
			if got := request(http.MethodPost, "characters"); got.Code != http.StatusForbidden {
				t.Errorf("raw character creation: expected 403, got %d", got.Code)
			}
		})
	}
}
