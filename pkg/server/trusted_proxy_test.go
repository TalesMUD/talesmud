package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/authlocal"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/gamemode"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/scripts/runner"
	"github.com/talesmud/talesmud/pkg/server/handler"
	"github.com/talesmud/talesmud/pkg/service"
)

func TestDirectPeerCannotRotateForwardedForPastLoginLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	facade := service.NewFacade(repository.NewSQLiteFactory(client), runner.NewMultiRunner())
	auth := authlocal.New(client.DB(), facade.UsersService(), "test-secret", &authlocal.CaptureMailer{})

	r := gin.New()
	if err := r.SetTrustedProxies(gamemode.TrustedProxies()); err != nil {
		t.Fatal(err)
	}
	r.POST("/api/auth/login", (&handler.LocalAuthHandler{Auth: auth}).Login)

	peer := "203.0.113.10:40000"
	var last int
	for i := 0; i < 13; i++ {
		body := strings.NewReader(`{"username":"nobody","password":"not-a-real-password"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
		req.RemoteAddr = peer
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", fmt.Sprintf("198.51.100.%d", i+1))
		req.Header.Set("X-Real-IP", fmt.Sprintf("198.51.100.%d", i+1))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		last = rec.Code
		if i < 12 && rec.Code == http.StatusTooManyRequests {
			t.Fatalf("request %d was limited early: %d", i+1, rec.Code)
		}
	}
	if last != http.StatusTooManyRequests {
		t.Fatalf("rotating X-Forwarded-For bypassed the limit, last status %d", last)
	}
}
