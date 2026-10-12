package server

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/talesmud/talesmud/pkg/authsession"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/devicecode"
	e "github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	mud "github.com/talesmud/talesmud/pkg/mudserver"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/server/handler"
	"github.com/talesmud/talesmud/pkg/service"
)

type deviceHTTP struct {
	srv      *httptest.Server
	facade   service.Facade
	identity string
}

// Uses ephemeral SQLite and a local JWKS server, never production auth.
func newDeviceHTTP(t *testing.T, cfg devicecode.Config) *deviceHTTP {
	t.Helper()
	gin.SetMode(gin.TestMode)
	prev := localSessions
	localSessions = nil
	clearJWKSCache()
	t.Cleanup(func() { localSessions = prev; clearJWKSCache() })
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	cert := &x509.Certificate{SerialNumber: big.NewInt(1), NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature}
	der, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	const kid = "device-test"
	jwksSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(gin.H{"keys": []gin.H{{"kty": "RSA", "kid": kid, "use": "sig", "x5c": []string{base64.StdEncoding.EncodeToString(der)}}}})
	}))
	t.Cleanup(jwksSrv.Close)
	const audience = "http://device.test/api"
	const issuer = "https://device-test.invalid/"
	t.Setenv("AUTH0_WK_JWKS", jwksSrv.URL)
	t.Setenv("AUTH0_AUDIENCE", audience)
	t.Setenv("AUTH0_DOMAIN", issuer)
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "device.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	sessions := authsession.New("device-test-session-secret", facade.UsersService())
	devices := devicecode.New(cfg)
	api := &handler.DeviceAPI{Devices: devices, Issue: sessions.Issue}
	sshAPI := handler.NewSSHAPI()
	r := gin.New()
	_ = r.SetTrustedProxies(nil)
	r.POST("/api/device/start", api.Start)
	r.POST("/api/device/poll", api.Poll)
	protected := r.Group("/api/", AuthMiddleware(facade, sessions))
	tickets := NewTicketStore()
	protected.POST("ws-ticket", tickets.IssueWSTicket)
	protected.POST("ssh/device/lookup", sshAPI.Lookup)
	protected.POST("ssh/device/confirm", sshAPI.Confirm)
	protected.POST("ssh/device/deny", sshAPI.Deny)
	protected.GET("who", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ref": c.GetString("userid")}) })
	csh := &handler.CharactersHandler{Service: facade.CharactersService()}
	protected.GET("my-characters", csh.GetMyCharacters)
	r.GET("/activate", handler.Activate(false))
	mudServer := mud.New(facade, nil)
	r.GET("/ws", WSAuthMiddleware(facade, tickets, sessions), mudServer.HandleConnections)
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	api.ActivateURL = srv.URL + "/activate"
	sshAPI.Set(handler.SSHSettings{Devices: devices, ActivateURL: api.ActivateURL, DeviceEnabled: true})
	return &deviceHTTP{srv, facade, signAuth0(t, key, kid, audience, issuer, "auth0|device-account")}
}

func (h *deviceHTTP) request(t *testing.T, method, path, token, origin string, body any) (int, map[string]json.RawMessage) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req, err := http.NewRequest(method, h.srv.URL+path, &buf)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	resp, err := h.srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasPrefix(path, "/api/device/") && resp.Header.Get("Cache-Control") != "no-store" {
		t.Fatal("secret response is cacheable")
	}
	var result map[string]json.RawMessage
	if len(data) > 0 {
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("status %d: malformed response", resp.StatusCode)
		}
	}
	return resp.StatusCode, result
}

func jsonString(t *testing.T, body map[string]json.RawMessage, key string) string {
	t.Helper()
	var value string
	if err := json.Unmarshal(body[key], &value); err != nil {
		t.Fatalf("missing or invalid %s", key)
	}
	return value
}

func (h *deviceHTTP) start(t *testing.T) (string, string) {
	t.Helper()
	status, body := h.request(t, http.MethodPost, "/api/device/start", "", "", gin.H{"client": "test-terminal", "client_version": "0.1", "proto": 1})
	if status != http.StatusOK {
		t.Fatalf("start: %d", status)
	}
	code, secret := jsonString(t, body, "user_code"), jsonString(t, body, "device_code")
	if code == secret || len(secret) != 64 {
		t.Fatal("invalid secret separation")
	}
	if jsonString(t, body, "verification_uri") != h.srv.URL+"/activate" {
		t.Fatal("incorrect activation URI")
	}
	complete, err := url.Parse(jsonString(t, body, "verification_uri_complete"))
	if err != nil || complete.Query().Get("code") != code || complete.Path != "/activate" {
		t.Fatal("incorrect complete activation URI")
	}
	if string(body["interval"]) != "5" {
		t.Fatal("incorrect poll interval")
	}
	return code, secret
}

func (h *deviceHTTP) decide(t *testing.T, code, action string) {
	t.Helper()
	status, body := h.request(t, http.MethodPost, "/api/ssh/device/lookup", h.identity, h.srv.URL, gin.H{"user_code": code})
	if status != http.StatusOK {
		t.Fatalf("lookup: %d", status)
	}
	if jsonString(t, body, "client_kind") != "app" || jsonString(t, body, "client") != "test-terminal" {
		t.Fatal("incorrect client metadata")
	}
	csrf := jsonString(t, body, "csrf")
	status, _ = h.request(t, http.MethodPost, "/api/ssh/device/"+action, h.identity, h.srv.URL, gin.H{"user_code": code, "csrf": csrf})
	if status != http.StatusOK {
		t.Fatalf("%s: %d", action, status)
	}
}

func TestPublicDeviceLoginHTTPAndWebSocket(t *testing.T) {
	h := newDeviceHTTP(t, devicecode.Config{})
	code, secret := h.start(t)
	status, body := h.request(t, http.MethodPost, "/api/device/poll", "", "", gin.H{"device_code": secret})
	if status != 400 || jsonString(t, body, "error") != "authorization_pending" {
		t.Fatal("pending poll contract")
	}
	_, body = h.request(t, http.MethodPost, "/api/ssh/device/lookup", h.identity, h.srv.URL, gin.H{"user_code": code})
	csrf := jsonString(t, body, "csrf")
	for _, origin := range []string{"", "https://evil.invalid"} {
		status, _ := h.request(t, http.MethodPost, "/api/ssh/device/confirm", h.identity, origin, gin.H{"user_code": code, "csrf": csrf})
		if status != 403 {
			t.Fatal("confirm accepted missing or wrong origin")
		}
	}
	status, _ = h.request(t, http.MethodPost, "/api/ssh/device/confirm", h.identity, h.srv.URL, gin.H{"user_code": code, "csrf": "incorrect"})
	if status != 404 {
		t.Fatal("confirm accepted wrong CSRF")
	}
	h.decide(t, code, "confirm")
	_, body = h.request(t, http.MethodPost, "/api/device/poll", "", "", gin.H{"device_code": secret})
	if jsonString(t, body, "error") != "slow_down" {
		t.Fatal("fast poll not slowed down")
	}
	// The fast second poll increased the interval from five to ten seconds.
	time.Sleep(10*time.Second + 20*time.Millisecond)
	status, body = h.request(t, http.MethodPost, "/api/device/poll", "", "", gin.H{"device_code": secret})
	if status != 200 || string(body["expires_in"]) != "1800" {
		t.Fatal("confirmed poll did not return session")
	}
	token := jsonString(t, body, "token")
	if token == "auth0|device-account" || !authsession.IsIssuer(token) {
		t.Fatal("poll did not issue signed server session")
	}
	if _, err := h.facade.GuestService().ValidateGuestToken(token); err == nil {
		t.Fatal("account token is a guest HMAC")
	}
	status, body = h.request(t, http.MethodPost, "/api/device/poll", "", "", gin.H{"device_code": secret})
	if status != 400 || jsonString(t, body, "error") != "expired_token" {
		t.Fatal("poll token consumed twice")
	}
	status, body = h.request(t, http.MethodGet, "/api/who", token, "", nil)
	if status != 200 || jsonString(t, body, "ref") != "auth0|device-account" {
		t.Fatal("middleware selected wrong account")
	}
	user, err := h.facade.UsersService().FindByRefID("auth0|device-account")
	if err != nil {
		t.Fatal(err)
	}
	char := &characters.Character{Entity: e.NewEntity(), Name: "Device player", BelongsUser: traits.BelongsUser{BelongsUserID: user.ID}}
	if _, err := h.facade.CharactersService().Store(char); err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodGet, h.srv.URL+"/api/my-characters", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := h.srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var chars []*characters.Character
	err = json.NewDecoder(resp.Body).Decode(&chars)
	resp.Body.Close()
	if err != nil || resp.StatusCode != 200 || len(chars) != 1 || chars[0].ID != char.ID {
		t.Fatal("account token failed character HTTP route")
	}
	conn, resp, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(h.srv.URL, "http")+"/ws", http.Header{"Authorization": []string{"Bearer " + token}})
	if err != nil {
		t.Fatalf("account token websocket: %v", err)
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatal("websocket did not upgrade")
	}
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	var message map[string]any
	if err := conn.ReadJSON(&message); err != nil {
		t.Fatalf("MUD welcome: %v", err)
	}
	_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	_ = conn.Close()
	t.Log("local HTTP start/confirm/poll, account character route and real MUD /ws upgrade verified using temporary SQLite and test-signed Auth0 identity")
}

func TestPublicDeviceDenyExpiryGuestAndLimits(t *testing.T) {
	h := newDeviceHTTP(t, devicecode.Config{TTL: 2 * time.Second, MaxPendingPerIP: 1, StartsPerIP: 3, PollsPerIP: 3})
	code, secret := h.start(t)
	status, _ := h.request(t, http.MethodPost, "/api/device/start", "", "", gin.H{"client": "test", "proto": 1})
	if status != 429 {
		t.Fatal("start pending cap bypassed")
	}
	guestToken, err := h.facade.GuestService().CreateGuestSession("198.51.100.99")
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range []string{"lookup", "confirm", "deny"} {
		status, _ = h.request(t, http.MethodPost, "/api/ssh/device/"+action, guestToken, h.srv.URL, gin.H{"user_code": code, "csrf": "anything"})
		if status != 403 {
			t.Fatalf("guest %s: %d", action, status)
		}
	}
	h.decide(t, code, "deny")
	status, body := h.request(t, http.MethodPost, "/api/device/poll", "", "", gin.H{"device_code": secret})
	if status != 400 || jsonString(t, body, "error") != "access_denied" {
		t.Fatal("deny contract")
	}
	_, secret = h.start(t)
	time.Sleep(2050 * time.Millisecond)
	_, body = h.request(t, http.MethodPost, "/api/device/poll", "", "", gin.H{"device_code": secret})
	if jsonString(t, body, "error") != "expired_token" {
		t.Fatal("expiry contract")
	}
	_, body = h.request(t, http.MethodPost, "/api/device/poll", "", "", gin.H{"device_code": "incorrect"})
	if jsonString(t, body, "error") != "expired_token" {
		t.Fatal("bad secret response differs from expiry")
	}
	_, body = h.request(t, http.MethodPost, "/api/device/poll", "", "", gin.H{"device_code": "another"})
	if jsonString(t, body, "error") != "slow_down" {
		t.Fatal("poll IP cap bypassed")
	}
	status, _ = h.request(t, http.MethodPost, "/api/device/start", "", "", gin.H{"client": "test", "proto": 1})
	if status != 429 {
		t.Fatal("start rate limit bypassed after expiry")
	}
}

func TestSessionRejectedByAuthMiddleware(t *testing.T) {
	h := newDeviceHTTP(t, devicecode.Config{})
	user, err := h.facade.UsersService().FindOrCreateNewUser("auth0|session-invalid")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"expired", "missing exp", "wrong audience", "wrong issuer", "wrong signature", "bare ref"} {
		t.Run(name, func(t *testing.T) {
			claims := jwt.MapClaims{"sub": user.ID, "iss": authsession.Issuer, "aud": authsession.Audience, "kind": "device-session", "sv": 0, "exp": time.Now().Add(time.Minute).Unix()}
			secret := "device-test-session-secret"
			switch name {
			case "expired":
				claims["exp"] = time.Now().Add(-time.Second).Unix()
			case "missing exp":
				delete(claims, "exp")
			case "wrong audience":
				claims["aud"] = "wrong"
			case "wrong issuer":
				claims["iss"] = "wrong"
			case "wrong signature":
				secret = "wrong"
			}
			token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
			if err != nil {
				t.Fatal(err)
			}
			if name == "bare ref" {
				token = user.RefID
			}
			status, _ := h.request(t, http.MethodGet, "/api/who", token, "", nil)
			if status != 401 {
				t.Fatalf("invalid session: %d", status)
			}
		})
	}
}
