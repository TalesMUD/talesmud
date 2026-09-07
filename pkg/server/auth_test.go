package server

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
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

func TestAuthMiddlewareRejectsAccessTokenQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(testFacade(t)))
	r.GET("/api/user", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/api/user?access_token=secret", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestWSAuthMiddlewareTicketSingleUse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	facade := testFacade(t)
	user := &entities.User{
		Entity:   &entities.Entity{ID: "user-1"},
		RefID:    "auth0|user-1",
		Nickname: "tester",
		Role:     entities.RolePlayer,
	}
	if _, err := facade.UsersService().Create(user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	tickets := NewTicketStore()
	ticket, _, err := tickets.Issue(user)
	if err != nil {
		t.Fatalf("issue ticket: %v", err)
	}

	r := gin.New()
	r.Use(WSAuthMiddleware(facade, tickets))
	r.GET("/ws", func(c *gin.Context) {
		got, _ := c.Get("user")
		if got.(*entities.User).ID != user.ID {
			t.Errorf("user mismatch")
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ws?ticket="+ticket, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("first use expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/ws?ticket="+ticket, nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("reuse expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestWSAuthMiddlewareRejectsAccessTokenQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(WSAuthMiddleware(testFacade(t), NewTicketStore()))
	r.GET("/ws", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/ws?access_token=secret", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestTicketExpires(t *testing.T) {
	tickets := NewTicketStore()
	user := &entities.User{Entity: &entities.Entity{ID: "u1"}}
	tickets.mu.Lock()
	tickets.tickets["old"] = ticketEntry{userID: user.ID, exp: time.Now().Add(-time.Second)}
	tickets.mu.Unlock()
	if _, ok := tickets.Consume("old"); ok {
		t.Fatal("expired ticket should not consume")
	}
}

func TestCreatorMiddlewareRejectsPlayers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user", &entities.User{
			Entity: &entities.Entity{ID: "p1"},
			Role:   entities.RolePlayer,
		})
		c.Next()
	})
	r.Use(CreatorMiddleware())
	r.GET("/api/scripts", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/api/scripts", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestTrustedProxiesDefault(t *testing.T) {
	t.Setenv("TRUSTED_PROXIES", "")
	got := trustedProxies()
	if !containsString(got, "127.0.0.1") || !containsString(got, "::1") {
		t.Fatalf("unexpected default proxies %v", got)
	}
}
