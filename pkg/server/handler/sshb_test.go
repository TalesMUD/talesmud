package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/talesmud/talesmud/pkg/authlocal"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

// TestRegisterResponseDoesNotEnumerate is SSHB-02. A new account, a taken
// username, and a taken email share one status and one body.
func TestRegisterResponseDoesNotEnumerate(t *testing.T) {
	prev := authLimiter
	authLimiter = newIPLimiter(12, time.Minute)
	t.Cleanup(func() { authLimiter = prev })

	gin.SetMode(gin.TestMode)
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "enum.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	auth := authlocal.New(client.DB(), facade.UsersService(), "test-secret", nil)
	r := gin.New()
	h := &LocalAuthHandler{Auth: auth}
	r.POST("/api/auth/register", h.Register)
	r.POST("/api/auth/login", h.Login)

	first := postRaw(t, r, "/api/auth/register", `{"username":"ada_ok","email":"ada_ok@example.com","password":"password1"}`)
	userHit := postRaw(t, r, "/api/auth/register", `{"username":"Ada_Ok","email":"other@example.com","password":"password2"}`)
	emailHit := postRaw(t, r, "/api/auth/register", `{"username":"bea_ok","email":"ADA_OK@example.com","password":"password2"}`)
	if first.code != http.StatusOK || userHit.code != first.code || emailHit.code != first.code {
		t.Fatalf("status new %d user %d email %d", first.code, userHit.code, emailHit.code)
	}
	if first.body != userHit.body || first.body != emailHit.body {
		t.Fatalf("bodies differ %q %q %q", first.body, userHit.body, emailHit.body)
	}
	want := `{"ok":true,"message":"Sign in to continue."}`
	if strings.TrimSpace(first.body) != want {
		t.Fatalf("body %s", first.body)
	}
	for _, leaked := range []string{"already", "exists", "token", "ada_ok", "email", "username"} {
		if strings.Contains(strings.ToLower(first.body), leaked) {
			t.Fatalf("body contains %q: %s", leaked, first.body)
		}
	}
	if _, err := facade.UsersService().FindByUsername("ada_ok"); err != nil {
		t.Fatal(err)
	}
	if _, err := facade.UsersService().FindByUsername("bea_ok"); err == nil {
		t.Fatal("email collision created an account")
	}

	badName := postRaw(t, r, "/api/auth/register", `{"username":"bad-name","email":"bad@example.com","password":"password1"}`)
	badMail := postRaw(t, r, "/api/auth/register", `{"username":"goodname","email":"not-an-email","password":"password1"}`)
	badPass := postRaw(t, r, "/api/auth/register", `{"username":"goodname","email":"good@example.com","password":"short"}`)
	for _, got := range []rawReply{badName, badMail, badPass} {
		if got.code != http.StatusBadRequest || got.body == first.body {
			t.Fatalf("validation %d %s", got.code, got.body)
		}
	}
	if strings.Contains(badName.body, "already") || strings.Contains(badMail.body, "already") {
		t.Fatal("validation says an account exists")
	}

	wrong := postRaw(t, r, "/api/auth/login", `{"username":"ada_ok","password":"password2"}`)
	if wrong.code != http.StatusUnauthorized || !strings.Contains(wrong.body, "invalid username or password") {
		t.Fatalf("login %d %s", wrong.code, wrong.body)
	}
	if strings.Contains(wrong.body, "already") || strings.Contains(wrong.body, "Sign in to continue.") {
		t.Fatalf("login body %s", wrong.body)
	}
	right := postRaw(t, r, "/api/auth/login", `{"username":"ada_ok","password":"password1"}`)
	if right.code != http.StatusOK || !strings.Contains(right.body, `"token"`) {
		t.Fatalf("login did not issue a session %d %s", right.code, right.body)
	}
}

type rawReply struct {
	code int
	body string
}

func postRaw(t *testing.T, h http.Handler, path, body string) rawReply {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "203.0.113.77:40000"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	raw, _ := io.ReadAll(rec.Body)
	return rawReply{code: rec.Code, body: string(raw)}
}

// TestAuthLimiterBucketsPrefix is SSHB-03. Register and login share allow,
// and one IPv6 /64 is one bucket.
func TestAuthLimiterBucketsPrefix(t *testing.T) {
	prev := authLimiter
	authLimiter = newIPLimiter(12, time.Minute)
	t.Cleanup(func() { authLimiter = prev })

	for i := 0; i < 12; i++ {
		if !authLimiter.allow("2001:db8:abcd::1") {
			t.Fatalf("attempt %d was limited", i+1)
		}
	}
	if authLimiter.allow("2001:db8:abcd::1") {
		t.Fatal("13th attempt from the same address was allowed")
	}
	if authLimiter.allow("2001:db8:abcd::2") || authLimiter.allow("2001:DB8:ABCD::2") {
		t.Fatal("another address in the same /64 was allowed")
	}
	if !authLimiter.allow("2001:db8:abce::1") {
		t.Fatal("a different /64 was blocked")
	}

	authLimiter = newIPLimiter(12, time.Minute)
	for i := 0; i < 12; i++ {
		if !authLimiter.allow("203.0.113.8") {
			t.Fatalf("v4 attempt %d was limited", i+1)
		}
	}
	if authLimiter.allow("::ffff:203.0.113.8") {
		t.Fatal("IPv4-mapped address was a new bucket")
	}
	if !authLimiter.allow("203.0.113.9") {
		t.Fatal("a different IPv4 address was blocked")
	}

	authLimiter = newIPLimiter(1, time.Minute)
	if !authLimiter.allow("") || authLimiter.allow("   ") {
		t.Fatal("blank addresses did not share a bucket")
	}
	if !authLimiter.allow("not-an-ip") {
		t.Fatal("an unparseable address shared the blank bucket")
	}
}
