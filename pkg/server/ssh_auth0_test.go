package server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/devicecode"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/server/handler"
	"github.com/talesmud/talesmud/pkg/service"
)

func TestSSHAuth0DeviceConfirm(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prev := localSessions
	localSessions = nil
	clearJWKSCache()
	t.Cleanup(func() {
		localSessions = prev
		clearJWKSCache()
	})

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	other, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	const kid = "ssh-test-kid"
	body, err := json.Marshal(map[string]any{
		"keys": []map[string]any{{
			"kty": "RSA",
			"kid": kid,
			"use": "sig",
			"x5c": []string{base64.StdEncoding.EncodeToString(der)},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	jwksSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer jwksSrv.Close()

	const (
		audience = "http://tales.test/api"
		issuer   = "https://ssh-test.invalid/"
		subject  = "auth0|ssh-test"
	)
	t.Setenv("AUTH0_WK_JWKS", jwksSrv.URL)
	t.Setenv("AUTH0_AUDIENCE", audience)
	t.Setenv("AUTH0_DOMAIN", issuer)

	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "auth0.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	devices := devicecode.New(devicecode.Config{})
	id, display, _, err := devices.Begin("203.0.113.8", "mud", "OpenSSH_9.6", []string{"SHA256:abcdefghijklmnopqrstuvwxyz012345"})
	if err != nil {
		t.Fatal(err)
	}
	api := handler.NewSSHAPI()
	r := gin.New()
	protected := r.Group("/api/")
	protected.Use(AuthMiddleware(facade))
	protected.POST("ssh/device/lookup", api.Lookup)
	protected.POST("ssh/device/confirm", api.Confirm)
	api.Set(handler.SSHSettings{
		Devices:       devices,
		ActivateURL:   "http://127.0.0.1/activate",
		DeviceEnabled: true,
	})

	signed := signAuth0(t, key, kid, audience, issuer, subject)
	lookup := postAuth0(t, r, "/api/ssh/device/lookup", signed, `{"user_code":"`+display+`"}`)
	if lookup.Code != http.StatusOK {
		t.Fatalf("lookup %d", lookup.Code)
	}
	var view struct {
		IP   string `json:"ip"`
		Mode string `json:"mode"`
		Key  string `json:"key"`
		CSRF string `json:"csrf"`
	}
	if err := json.Unmarshal(lookup.Body.Bytes(), &view); err != nil {
		t.Fatal(err)
	}
	if view.IP != "203.0.113.8" || view.Mode != "mud" || view.CSRF == "" {
		t.Fatalf("lookup view ip=%q mode=%q csrf empty=%v", view.IP, view.Mode, view.CSRF == "")
	}
	if view.Key == "" || strings.Contains(view.Key, "abcdefghijklmnopqrstuvwxyz012345") {
		t.Fatalf("lookup leaked or dropped the key fingerprint")
	}
	confirm := postAuth0(t, r, "/api/ssh/device/confirm", signed, `{"user_code":"`+display+`","csrf":"`+view.CSRF+`"}`)
	if confirm.Code != http.StatusOK {
		t.Fatalf("confirm %d", confirm.Code)
	}
	ref, ok := devices.Take(id)
	if !ok || ref != subject {
		t.Fatalf("take ok=%v ref=%q", ok, ref)
	}
	user, err := facade.UsersService().FindByRefID(subject)
	if err != nil || user == nil || user.RefID != subject {
		t.Fatalf("user from token: %v", err)
	}

	bad := postAuth0(t, r, "/api/ssh/device/lookup", signAuth0(t, other, kid, audience, issuer, subject), `{"user_code":"`+display+`"}`)
	if bad.Code != http.StatusUnauthorized {
		t.Fatalf("bad signature %d", bad.Code)
	}
}

func signAuth0(t *testing.T, key *rsa.PrivateKey, kid, audience, issuer, subject string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": subject,
		"aud": audience,
		"iss": issuer,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Add(-time.Minute).Unix(),
	})
	token.Header["kid"] = kid
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func postAuth0(t *testing.T, r http.Handler, path, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func clearJWKSCache() {
	jwksCache.Lock()
	jwksCache.keys = nil
	jwksCache.fetchedAt = time.Time{}
	jwksCache.Unlock()
}
