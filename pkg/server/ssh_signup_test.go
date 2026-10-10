package server

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"

	"github.com/talesmud/talesmud/pkg/authlocal"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/devicecode"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/gamemode"
	"github.com/talesmud/talesmud/pkg/mudserver"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/server/handler"
	"github.com/talesmud/talesmud/pkg/service"
	"github.com/talesmud/talesmud/pkg/sshgate"
	"github.com/talesmud/talesmud/pkg/sshkeys"
)

func TestSSHRegisterThenConfirm(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prev := localSessions
	t.Cleanup(func() { localSessions = prev })
	applyMode(t, "presentation: classic\nauth: local\nssh:\n  signup:\n    enabled: true\n  guest:\n    enabled: false\n")

	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "signup.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	if _, err := facade.RoomsService().Store(&rooms.Room{
		Entity:      &entities.Entity{ID: "R0001"},
		Name:        "Harbor",
		Description: "SSH-SIGNUP-HARBOR stands quiet.",
		LookAt:      traits.LookAt{Detail: "SSH-SIGNUP-HARBOR stands quiet."},
	}); err != nil {
		t.Fatal(err)
	}
	auth := authlocal.New(client.DB(), facade.UsersService(), "test-secret", nil)
	UseLocalAuth(auth)
	if _, _, err := auth.Register("no", "no@example.com", "password1"); !errors.Is(err, authlocal.ErrUsername) {
		t.Fatalf("short name %v", err)
	}
	if _, _, err := auth.Register("bad-name", "bad@example.com", "password1"); err == nil {
		t.Fatal("hyphen accepted")
	}
	if _, _, err := auth.Register("newssh", "newssh@example.com", "short"); !errors.Is(err, authlocal.ErrPassword) {
		t.Fatalf("short password %v", err)
	}
	if _, err := facade.UsersService().FindByUsername("newssh"); err == nil {
		t.Fatal("rejected registration created a user")
	}

	keys, err := sshkeys.Open(client.DB())
	if err != nil {
		t.Fatal(err)
	}
	devices := devicecode.New(devicecode.Config{})
	api := handler.NewSSHAPI()
	r := gin.New()
	r.POST("/api/auth/register", (&handler.LocalAuthHandler{Auth: auth}).Register)
	protected := r.Group("/api/")
	protected.Use(AuthMiddleware(facade))
	protected.POST("ssh/device/lookup", api.Lookup)
	protected.POST("ssh/device/confirm", api.Confirm)
	r.GET("/activate", handler.Activate(true))
	srv := httptest.NewServer(r)
	defer srv.Close()
	api.Set(handler.SSHSettings{
		Keys: keys, Devices: devices, ActivateURL: srv.URL + "/activate",
		KeysEnabled: true, DeviceEnabled: true, WebManage: false, MaxKeys: 10,
	})
	page := httptest.NewRecorder()
	r.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/activate?code=BCDF-GHJK&signup=1", nil))
	if !strings.Contains(page.Body.String(), "Create account") || !strings.Contains(page.Body.String(), "Sign in") {
		t.Fatal("activate page did not show both forms")
	}

	mud := mudserver.New(facade)
	mud.Run()
	gate, err := sshgate.Listen(gamemode.SSHConfig{
		Enabled: true, Listen: "127.0.0.1:0",
		HostKeyPath: filepath.Join(t.TempDir(), "host_ed25519"),
		Keys:        gamemode.SSHKeysConfig{Enabled: true, MaxPerAccount: 10},
		Device:      gamemode.SSHDeviceConfig{Enabled: true, ActivateURL: srv.URL + "/activate"},
		Signup:      gamemode.SSHSignupConfig{Enabled: true},
		IdleTimeout: gamemode.Duration(30 * time.Minute),
	}, sshgate.Deps{Mud: mud, Users: facade.UsersService(), Keys: keys, Devices: devices})
	if err != nil {
		t.Fatal(err)
	}
	defer gate.Close()

	signer := testSigner(t)
	fp := ssh.FingerprintSHA256(signer.PublicKey())
	conn := dialPlayer(t, gate.Addr(), signer, true)
	session, stdin, out := playerShell(t, conn)
	text := out.wait(t, "Code: ", 15*time.Second)
	if !strings.Contains(text, "[N] New player") || !strings.Contains(text, "signup=1") {
		t.Fatalf("lobby %q", text)
	}
	if _, err := facade.UsersService().FindByUsername("newssh"); err == nil {
		t.Fatal("ssh lobby created an account")
	}
	match := regexp.MustCompile(`Code: ([BCDFGHJKLMNPQRSTVWXZ]{4}-[BCDFGHJKLMNPQRSTVWXZ]{4})`).FindStringSubmatch(text)
	if match == nil {
		t.Fatalf("no code in %q", text)
	}
	code := match[1]
	reg := postAuth(t, srv.URL+"/api/auth/register", "", map[string]string{
		"username": "newssh", "email": "newssh@example.com", "password": "password1",
	}, "")
	token, _ := reg["token"].(string)
	if token == "" {
		t.Fatalf("register %#v", reg)
	}
	if _, err := facade.UsersService().FindByUsername("newssh"); err != nil {
		t.Fatal(err)
	}
	user, err := facade.UsersService().FindByUsername("newssh")
	if err != nil || user == nil {
		t.Fatal(err)
	}
	lookup := postAuth(t, srv.URL+"/api/ssh/device/lookup", token, map[string]string{"user_code": code}, srv.URL)
	csrf, _ := lookup["csrf"].(string)
	if csrf == "" || lookup["ip"] != "127.0.0.1" {
		t.Fatalf("lookup %#v", lookup)
	}
	if postStatus(t, srv.URL+"/api/ssh/device/confirm", token, map[string]string{"user_code": code, "csrf": csrf}, "") != http.StatusForbidden {
		t.Fatal("missing origin")
	}
	if postStatus(t, srv.URL+"/api/ssh/device/confirm", token, map[string]string{"user_code": code, "csrf": csrf}, "https://evil.example") != http.StatusForbidden {
		t.Fatal("evil origin")
	}
	if postStatus(t, srv.URL+"/api/ssh/device/confirm", token, map[string]string{"user_code": code, "csrf": "nope"}, srv.URL) != http.StatusNotFound {
		t.Fatal("bad csrf")
	}
	if postStatus(t, srv.URL+"/api/ssh/device/confirm", token, map[string]string{"user_code": code, "csrf": csrf}, srv.URL) != http.StatusOK {
		t.Fatal("confirm")
	}
	out.wait(t, "Remember this computer?", 15*time.Second)
	if _, err := io.WriteString(stdin, "y"); err != nil {
		t.Fatal(err)
	}
	linked := out.wait(t, "Key linked.", 15*time.Second)
	if strings.Contains(linked, "Reconnect once") {
		t.Fatal("signup link asked to reconnect")
	}
	rows, err := keys.List(user.RefID)
	if err != nil || len(rows) != 1 || rows[0].CreatedVia != "device" || rows[0].Fingerprint != fp {
		t.Fatalf("linked %+v %v", rows, err)
	}
	_ = session.Close()
	_ = conn.Close()

	conn = dialPlayer(t, gate.Addr(), signer, false)
	session, _, out = playerShell(t, conn)
	direct := out.wait(t, "Connected to", 15*time.Second)
	if strings.Contains(direct, "Code:") || strings.Contains(direct, "Only press Y") {
		t.Fatalf("key login asked again %q", direct)
	}
	_ = session.Close()
	_ = conn.Close()
}

func TestSSHSignupCodeExpires(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prev := localSessions
	t.Cleanup(func() { localSessions = prev })
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "expire.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	auth := authlocal.New(client.DB(), facade.UsersService(), "test-secret", nil)
	UseLocalAuth(auth)
	token, _, err := auth.Register("expireu", "expireu@example.com", "password1")
	if err != nil {
		t.Fatal(err)
	}
	user, err := facade.UsersService().FindByUsername("expireu")
	if err != nil {
		t.Fatal(err)
	}
	devices := devicecode.New(devicecode.Config{TTL: 200 * time.Millisecond})
	_, code, _, err := devices.Begin("203.0.113.9", "door", "test", nil)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(350 * time.Millisecond)
	api := handler.NewSSHAPI()
	r := gin.New()
	protected := r.Group("/api/")
	protected.Use(AuthMiddleware(facade))
	protected.POST("ssh/device/lookup", api.Lookup)
	protected.POST("ssh/device/confirm", api.Confirm)
	srv := httptest.NewServer(r)
	defer srv.Close()
	api.Set(handler.SSHSettings{
		Devices: devices, ActivateURL: srv.URL + "/activate", DeviceEnabled: true,
	})
	if postStatus(t, srv.URL+"/api/ssh/device/lookup", token, map[string]string{"user_code": code}, srv.URL) != http.StatusNotFound {
		t.Fatal("expired code was found")
	}
	if postStatus(t, srv.URL+"/api/ssh/device/confirm", token, map[string]string{"user_code": code, "csrf": "stale"}, srv.URL) != http.StatusNotFound {
		t.Fatal("expired code was confirmed")
	}
	if _, _, _, err := devices.Begin("2001:db8:1:2::5", "door", "test", nil); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := devices.Begin("2001:db8:1:2::6", "door", "test", nil); err != nil {
		t.Fatal(err)
	}
	limited := devicecode.New(devicecode.Config{MaxPendingPerIP: 1})
	if _, _, _, err := limited.Begin("203.0.113.8", "door", "test", nil); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := limited.Begin("::ffff:203.0.113.8", "door", "test", nil); !errors.Is(err, devicecode.ErrLimited) {
		t.Fatalf("mapped address did not share the cap: %v", err)
	}
	if _, err := facade.UsersService().FindByUsername("nobody"); err == nil {
		t.Fatal("a capped signup created a user")
	}
	_ = user
}

func applyMode(t *testing.T, body string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "mode.yaml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
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
}
