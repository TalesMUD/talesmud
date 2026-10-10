package server

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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

func TestSSHDeviceLoginLinkAndGuards(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prev := localSessions
	t.Cleanup(func() { localSessions = prev })

	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "ssh.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	if _, err := facade.RoomsService().Store(&rooms.Room{
		Entity:      &entities.Entity{ID: "R0001"},
		Name:        "Harbor",
		Description: "SSH-FLOW-HARBOR stands quiet.",
		LookAt:      traits.LookAt{Detail: "SSH-FLOW-HARBOR stands quiet."},
	}); err != nil {
		t.Fatal(err)
	}
	auth := authlocal.New(client.DB(), facade.UsersService(), "test-secret", nil)
	UseLocalAuth(auth)
	token, _, err := auth.Register("sshuser", "sshuser@example.com", "password1")
	if err != nil {
		t.Fatal(err)
	}
	user, err := facade.UsersService().FindByUsername("sshuser")
	if err != nil || user == nil {
		t.Fatal(err)
	}
	signer := testSigner(t)
	fp := ssh.FingerprintSHA256(signer.PublicKey())
	keys, err := sshkeys.Open(client.DB())
	if err != nil {
		t.Fatal(err)
	}
	pending := sshkeys.NewPending(10 * time.Minute)
	devices := devicecode.New(devicecode.Config{})
	api := handler.NewSSHAPI()
	r := gin.New()
	protected := r.Group("/api/")
	protected.Use(AuthMiddleware(facade))
	protected.GET("ssh/keys", api.ListKeys)
	protected.POST("ssh/keys", api.AddKey)
	protected.DELETE("ssh/keys/:id", api.DeleteKey)
	protected.POST("ssh/device/lookup", api.Lookup)
	protected.POST("ssh/device/confirm", api.Confirm)
	protected.POST("ssh/device/deny", api.Deny)
	r.GET("/activate", handler.Activate(true))
	srv := httptest.NewServer(r)
	defer srv.Close()
	api.Set(handler.SSHSettings{
		Keys:          keys,
		Devices:       devices,
		ActivateURL:   srv.URL + "/activate",
		KeysEnabled:   true,
		DeviceEnabled: true,
		WebManage:     true,
		MaxKeys:       10,
	})

	mud := mudserver.New(facade)
	mud.Run()
	gate, err := sshgate.Listen(gamemode.SSHConfig{
		Enabled:     true,
		Listen:      "127.0.0.1:0",
		HostKeyPath: filepath.Join(t.TempDir(), "host_ed25519"),
		PublicHost:  "127.0.0.1",
		Guest:       gamemode.SSHGuestConfig{Enabled: true},
		Keys:        gamemode.SSHKeysConfig{Enabled: true, MaxPerAccount: 10},
		Device:      gamemode.SSHDeviceConfig{Enabled: true, ActivateURL: srv.URL + "/activate"},
		IdleTimeout: gamemode.Duration(30 * time.Minute),
	}, sshgate.Deps{
		Mud:     mud,
		Guests:  facade.GuestService(),
		Users:   facade.UsersService(),
		Keys:    keys,
		Pending: pending,
		Devices: devices,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer gate.Close()

	hook := &sshLogHook{}
	log.AddHook(hook)
	defer hook.stop()

	conn := dialPlayer(t, gate.Addr(), signer, true)
	session, stdin, out := playerShell(t, conn)
	text := out.wait(t, "Code: ", 15*time.Second)
	match := regexp.MustCompile(`Code: ([BCDFGHJKLMNPQRSTVWXZ]{4}-[BCDFGHJKLMNPQRSTVWXZ]{4})`).FindStringSubmatch(text)
	if match == nil {
		t.Fatalf("no code in %q", text)
	}
	code := match[1]
	if strings.Contains(text, fp) {
		t.Fatal("terminal showed the full fingerprint")
	}
	body := postAuth(t, srv.URL+"/api/ssh/device/lookup", token, map[string]string{"user_code": code}, "")
	if body["mode"] != "mud" || body["ip"] != "127.0.0.1" {
		t.Fatalf("lookup %#v", body)
	}
	gotKey, _ := body["key"].(string)
	if gotKey != sshgate.MaskFingerprint(fp) || strings.Contains(string(mustJSON(body)), fp) {
		t.Fatalf("lookup key %#v", body["key"])
	}
	csrf, _ := body["csrf"].(string)
	if csrf == "" {
		t.Fatal("missing csrf")
	}
	bad := postStatus(t, srv.URL+"/api/ssh/device/confirm", token, map[string]string{"user_code": code, "csrf": "nope"}, "")
	if bad != http.StatusNotFound {
		t.Fatalf("bad csrf %d", bad)
	}
	evil := postStatus(t, srv.URL+"/api/ssh/device/confirm", token, map[string]string{"user_code": code, "csrf": csrf}, "https://evil.example")
	if evil != http.StatusForbidden {
		t.Fatalf("evil origin %d", evil)
	}
	ok := postStatus(t, srv.URL+"/api/ssh/device/confirm", token, map[string]string{"user_code": code, "csrf": csrf}, srv.URL)
	if ok != http.StatusOK {
		t.Fatalf("confirm %d", ok)
	}
	again := postStatus(t, srv.URL+"/api/ssh/device/confirm", token, map[string]string{"user_code": code, "csrf": csrf}, "")
	if again != http.StatusNotFound {
		t.Fatalf("reuse %d", again)
	}
	out.wait(t, "Remember this computer?", 15*time.Second)
	if _, err := io.WriteString(stdin, "y"); err != nil {
		t.Fatal(err)
	}
	out.wait(t, "Reconnect once to finish", 15*time.Second)
	rows, err := keys.List(user.RefID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("key stored before the verified confirm: %+v", rows)
	}
	_ = session.Close()
	_ = conn.Close()

	conn = dialPlayer(t, gate.Addr(), signer, true)
	session, stdin, out = playerShell(t, conn)
	out.wait(t, "Only press Y if that is your account.", 15*time.Second)
	rows, _ = keys.List(user.RefID)
	if len(rows) != 0 {
		t.Fatal("key stored before the second yes")
	}
	if _, err := io.WriteString(stdin, "y"); err != nil {
		t.Fatal(err)
	}
	out.wait(t, "Key linked.", 15*time.Second)
	rows, err = keys.List(user.RefID)
	if err != nil || len(rows) != 1 || rows[0].CreatedVia != "device" {
		t.Fatalf("linked %+v %v", rows, err)
	}
	_ = session.Close()
	_ = conn.Close()

	conn = dialPlayer(t, gate.Addr(), signer, false)
	session, _, out = playerShell(t, conn)
	direct := out.wait(t, "Connected to", 15*time.Second)
	if strings.Contains(direct, "Only press Y") || strings.Contains(direct, "Code:") {
		t.Fatalf("direct login asked again %q", direct)
	}
	_ = session.Close()
	_ = conn.Close()

	waitOffline(t, facade, user.ID)
	if err := facade.UsersService().BanUser(user.ID); err != nil {
		t.Fatal(err)
	}
	if banned, err := facade.UsersService().FindByID(user.ID); err != nil || banned == nil || !banned.IsBanned {
		t.Fatalf("ban did not stick %+v %v", banned, err)
	}
	if _, err := dialPlayerErr(gate.Addr(), signer, false); err == nil {
		t.Fatal("banned key was accepted")
	}
	if postStatus(t, srv.URL+"/api/ssh/device/lookup", token, map[string]string{"user_code": "BCDF-GHJK"}, "") != http.StatusForbidden {
		t.Fatal("banned confirm was allowed")
	}
	if err := facade.UsersService().UnbanUser(user.ID); err != nil {
		t.Fatal(err)
	}

	if err := keys.Delete(user.RefID, rows[0].ID); err != nil {
		t.Fatal(err)
	}
	conn = dialPlayer(t, gate.Addr(), signer, true)
	session, _, out = playerShell(t, conn)
	out.wait(t, "Code: ", 15*time.Second)
	_ = session.Close()
	_ = conn.Close()

	conn = dialPlayer(t, gate.Addr(), signer, true)
	session, _, out = playerShell(t, conn)
	out.wait(t, "Code: ", 15*time.Second)
	devices.SetClock(func() time.Time { return time.Now().Add(time.Hour) })
	out.wait(t, "code not found or expired", 15*time.Second)
	devices.SetClock(time.Now)
	_ = session.Close()
	_ = conn.Close()

	guestTok, err := facade.GuestService().CreateGuestSessionPick("198.51.100.10", "", "")
	if err != nil {
		t.Fatal(err)
	}
	_, guestCode, _, err := devices.Begin("198.51.100.9", "mud", "test", []string{fp})
	if err != nil {
		t.Fatal(err)
	}
	if postStatus(t, srv.URL+"/api/ssh/device/lookup", guestTok, map[string]string{"user_code": guestCode}, "") != http.StatusForbidden {
		t.Fatal("guest lookup was allowed")
	}
	priv := postAuth(t, srv.URL+"/api/ssh/keys", token, map[string]string{"public_key": "-----BEGIN PRIVATE KEY-----\nsecret\n", "label": "nope"}, "")
	if _, ok := priv["error"]; !ok || strings.Contains(mustJSON(priv), "PRIVATE") {
		t.Fatalf("private key response %#v", priv)
	}
	line := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(testSigner(t).PublicKey())))
	added := postAuth(t, srv.URL+"/api/ssh/keys", token, map[string]string{"public_key": line, "label": "laptop"}, "")
	if added["created_via"] != "web" || added["id"] == "" {
		t.Fatalf("web add %#v", added)
	}
	api.Set(handler.SSHSettings{Keys: keys, Devices: devices, ActivateURL: srv.URL + "/activate", KeysEnabled: true, DeviceEnabled: true, WebManage: false, MaxKeys: 10})
	if postStatus(t, srv.URL+"/api/ssh/keys", token, map[string]string{"public_key": line, "label": "x"}, "") != http.StatusForbidden {
		t.Fatal("web manage off still accepted a key")
	}

	logged := hook.text()
	for _, secret := range []string{code, fp, csrf, token, "PRIVATE", "ssh-ed25519"} {
		if strings.Contains(logged, secret) {
			t.Fatalf("ssh log contained %q in %s", secret, logged)
		}
	}
}

func waitOffline(t *testing.T, facade service.Facade, id string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		user, err := facade.UsersService().FindByID(id)
		if err == nil && user != nil && !user.IsOnline {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func testSigner(t *testing.T) ssh.Signer {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	return signer
}

func dialPlayer(t *testing.T, addr string, signer ssh.Signer, interactive bool) *ssh.Client {
	t.Helper()
	conn, err := dialPlayerErr(addr, signer, interactive)
	if err != nil {
		t.Fatal(err)
	}
	return conn
}

func dialPlayerErr(addr string, signer ssh.Signer, interactive bool) (*ssh.Client, error) {
	methods := []ssh.AuthMethod{ssh.PublicKeys(signer)}
	if interactive {
		methods = append(methods, ssh.KeyboardInteractive(func(string, string, []string, []bool) ([]string, error) {
			return nil, nil
		}))
	}
	return ssh.Dial("tcp", addr, &ssh.ClientConfig{
		User:            "player",
		Auth:            methods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	})
}

func playerShell(t *testing.T, conn *ssh.Client) (*ssh.Session, io.WriteCloser, *flowCollect) {
	t.Helper()
	session, err := conn.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	if err := session.RequestPty("xterm", 24, 80, ssh.TerminalModes{}); err != nil {
		t.Fatal(err)
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdin, err := session.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Shell(); err != nil {
		t.Fatal(err)
	}
	return session, stdin, startFlow(stdout)
}

type flowCollect struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func startFlow(r io.Reader) *flowCollect {
	c := &flowCollect{}
	go func() {
		tmp := make([]byte, 512)
		for {
			n, err := r.Read(tmp)
			if n > 0 {
				c.mu.Lock()
				c.buf.Write(tmp[:n])
				c.mu.Unlock()
			}
			if err != nil {
				return
			}
		}
	}()
	return c
}

func (c *flowCollect) wait(t *testing.T, needle string, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		c.mu.Lock()
		text := c.buf.String()
		c.mu.Unlock()
		if strings.Contains(text, needle) {
			return text
		}
		time.Sleep(20 * time.Millisecond)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	t.Fatalf("missing %q in %q", needle, c.buf.String())
	return ""
}

func postAuth(t *testing.T, url, token string, body map[string]string, origin string) map[string]any {
	t.Helper()
	raw := doPost(t, url, token, body, origin)
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("json %s: %v", raw, err)
	}
	return out
}

func postStatus(t *testing.T, url, token string, body map[string]string, origin string) int {
	t.Helper()
	req := authReq(t, url, token, body, origin)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, res.Body)
	return res.StatusCode
}

func doPost(t *testing.T, url, token string, body map[string]string, origin string) []byte {
	t.Helper()
	req := authReq(t, url, token, body, origin)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 && !strings.Contains(url, "/api/ssh/keys") {
		t.Fatalf("%s %d %s", url, res.StatusCode, raw)
	}
	return raw
}

func authReq(t *testing.T, url, token string, body map[string]string, origin string) *http.Request {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	return req
}

func mustJSON(v any) string {
	raw, _ := json.Marshal(v)
	return string(raw)
}

type sshLogHook struct {
	mu  sync.Mutex
	buf bytes.Buffer
	off bool
}

func (h *sshLogHook) Levels() []log.Level { return log.AllLevels }

func (h *sshLogHook) Fire(e *log.Entry) error {
	if h == nil || e == nil || !strings.HasPrefix(e.Message, "ssh ") {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.off {
		return nil
	}
	h.buf.WriteString(e.Message)
	h.buf.WriteByte(' ')
	for k, v := range e.Data {
		h.buf.WriteString(k)
		h.buf.WriteByte('=')
		h.buf.WriteString(strings.TrimSpace(strings.ReplaceAll(toString(v), "\n", " ")))
		h.buf.WriteByte(' ')
	}
	return nil
}

func (h *sshLogHook) text() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.buf.String()
}

func (h *sshLogHook) stop() {
	h.mu.Lock()
	h.off = true
	h.mu.Unlock()
}

func toString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case error:
		return t.Error()
	default:
		raw, _ := json.Marshal(t)
		return string(raw)
	}
}
