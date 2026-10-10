package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	if !strings.Contains(body, `id="keyline"`) || strings.Contains(body, `"Key " + (data.key || "none") +`) {
		t.Fatal("confirm page does not give the stored key its own line")
	}
	if !strings.Contains(body, `getElementById("keyline").textContent = "Key " + (data.key || "none")`) {
		t.Fatal("key line is not filled from the lookup")
	}
	if !strings.Contains(body, `params.get("code")`) {
		t.Fatal("code is not prefilled")
	}
	if strings.Contains(body, "addkey") || strings.Contains(body, "/api/ssh/keys\",") {
		t.Fatal("activate page still offers a key paste")
	}
	_, _, tokenKey := gamemode.ClientPage()
	if !strings.Contains(body, tokenKey) {
		t.Fatalf("token key %s missing", tokenKey)
	}
	if strings.Contains(body, "Create account") || strings.Contains(body, "/api/auth/register") {
		t.Fatal("create account is offered while signup is closed")
	}
}

func TestActivateSignupFormWhenOpen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mode.yaml")
	if err := os.WriteFile(path, []byte("presentation: door_tui\nauth: local\nssh:\n  signup:\n    enabled: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := gamemode.ApplyFile(path); err != nil {
		t.Fatal(err)
	}
	off := filepath.Join(t.TempDir(), "off.yaml")
	if err := os.WriteFile(off, []byte("presentation: classic\nauth: auth0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := gamemode.ApplyFile(off); err != nil {
			t.Error(err)
		}
	})
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/activate", Activate(true))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/activate?code=BCDF-GHJK&signup=1", nil))
	body := rec.Body.String()
	if !strings.Contains(body, "Sign in") || !strings.Contains(body, "Create account") || !strings.Contains(body, "/api/auth/register") {
		t.Fatal("signup page is missing a side")
	}
	if strings.Contains(body, "addkey") || strings.Contains(body, "/api/ssh/keys\",") {
		t.Fatal("signup page offers a key paste")
	}
	if strings.Contains(body, "lookupCode();") {
		t.Fatal("register looks the code up")
	}
	if !strings.Contains(body, "Sign in to continue.") {
		t.Fatal("register does not send the player to sign in")
	}
	start := strings.Index(body, `getElementById("register")`)
	end := strings.Index(body, "let csrf")
	if start < 0 || end < start {
		t.Fatal("register script is missing")
	}
	reg := body[start:end]
	if strings.Contains(reg, "localStorage") || strings.Contains(reg, "lookupCode") || strings.Contains(reg, "token") {
		t.Fatal("register stores a session or looks the code up")
	}
	if strings.Contains(body, "decide('/api/ssh/device/confirm');") {
		t.Fatal("confirm is called by itself")
	}
	tail := body[strings.LastIndex(body, "if (token())"):]
	if strings.Contains(tail, "lookupCode") {
		t.Fatal("page load looks up the code")
	}
}

func TestRegisterRateLimit(t *testing.T) {
	prev := authLimiter
	authLimiter = newIPLimiter(12, time.Minute)
	t.Cleanup(func() { authLimiter = prev })
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/auth/register", (&LocalAuthHandler{}).Register)
	for i := 0; i < 12; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(""))
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("try %d status %d", i, rec.Code)
		}
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader("")))
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("limit %d %s", rec.Code, rec.Body.String())
	}
}
