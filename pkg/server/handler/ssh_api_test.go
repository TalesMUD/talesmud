package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	e "github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/gamemode"
)

func TestSSHRoutesClosedUntilConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := NewSSHAPI()
	r := gin.New()
	r.GET("/keys", func(c *gin.Context) {
		c.Set("user", &e.User{RefID: "local:a"})
		api.ListKeys(c)
	})
	r.POST("/lookup", func(c *gin.Context) {
		c.Set("user", &e.User{RefID: "local:a"})
		api.Lookup(c)
	})
	req := httptest.NewRequest(http.MethodGet, "/keys", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("keys %d", rec.Code)
	}
	req = httptest.NewRequest(http.MethodPost, "/lookup", strings.NewReader(`{"user_code":"BCDF-GHJK"}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("lookup %d %s", rec.Code, rec.Body.String())
	}
}

func TestOriginAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/ssh/device/confirm", nil)
	if originAllowed(req, "http://127.0.0.1:8031/activate") {
		t.Fatal("missing origin")
	}
	req.Header.Set("Origin", "https://evil.example")
	if originAllowed(req, "http://127.0.0.1:8031/activate") {
		t.Fatal("mismatched origin")
	}
	if originAllowed(req, "") {
		t.Fatal("origin without activate url")
	}
	req.Header.Set("Origin", "http://127.0.0.1:8031")
	if !originAllowed(req, "http://127.0.0.1:8031/activate") {
		t.Fatal("same origin")
	}
	req.Header.Del("Origin")
	req.Header.Set("Referer", "http://127.0.0.1:8031/activate?code=BCDF-GHJK")
	if originAllowed(req, "http://127.0.0.1:8031/activate") {
		t.Fatal("referer without origin")
	}
}

func TestActivatePageIsLocalAndPrefills(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/activate", Activate(true))
	r.GET("/off", Activate(false))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/off?code=BCDF-GHJK", nil))
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/play/?activate=BCDF-GHJK" {
		t.Fatalf("non-local %d %s", rec.Code, rec.Header().Get("Location"))
	}
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/activate?code=BCDF-GHJK", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("activate %d", rec.Code)
	}
	if rec.Header().Get("Cache-Control") != "no-store" || rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatalf("headers %v", rec.Header())
	}
	if !strings.Contains(rec.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'") {
		t.Fatalf("csp %s", rec.Header().Get("Content-Security-Policy"))
	}
	body := rec.Body.String()
	if strings.Contains(body, "Aethermoor") || strings.Contains(strings.ToLower(body), "lair of") {
		t.Fatal("activate page names a world")
	}
	if strings.Count(body, "lookupCode()") != 2 || strings.Contains(body, "lookupCode();") {
		t.Fatalf("lookup calls %d", strings.Count(body, "lookupCode()"))
	}
	if !strings.Contains(body, `params.get("code")`) {
		t.Fatal("code is not prefilled")
	}
	_, _, tokenKey := gamemode.ClientPage()
	if !strings.Contains(body, tokenKey) {
		t.Fatalf("token key %s missing", tokenKey)
	}
}
