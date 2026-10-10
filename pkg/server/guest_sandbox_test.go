package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/talesmud/talesmud/pkg/authlocal"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/devicecode"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/server/handler"
	"github.com/talesmud/talesmud/pkg/service"
	"github.com/talesmud/talesmud/pkg/sshkeys"
)

// TestGuestSandboxDeniesElevation proves a guest token cannot confirm or deny
// a device code, link or revoke a key, open a creator or admin route, or
// change account fields. An admin token still reaches those routes.
func TestGuestSandboxDeniesElevation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prev := localSessions
	t.Cleanup(func() { localSessions = prev })

	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "guestbox.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	if _, err := facade.RoomsService().Store(&rooms.Room{
		Entity:      &entities.Entity{ID: "R0001"},
		Name:        "Harbor",
		Description: "quiet",
		LookAt:      traits.LookAt{Detail: "quiet"},
	}); err != nil {
		t.Fatal(err)
	}
	auth := authlocal.New(client.DB(), facade.UsersService(), "test-secret", nil)
	UseLocalAuth(auth)
	adminTok, _, err := auth.Register("keeper", "keeper@example.com", "password1")
	if err != nil {
		t.Fatal(err)
	}
	admin, err := facade.UsersService().FindByUsername("keeper")
	if err != nil || admin == nil {
		t.Fatal(err)
	}
	admin.Role = entities.RoleAdmin
	if err := facade.UsersService().Update(admin.RefID, admin); err != nil {
		t.Fatal(err)
	}

	guestTok, err := facade.GuestService().CreateGuestSessionPick("203.0.113.8", "", "")
	if err != nil {
		t.Fatal(err)
	}
	guestID, err := facade.GuestService().ValidateGuestToken(guestTok)
	if err != nil {
		t.Fatal(err)
	}
	guest, err := facade.UsersService().FindByID(guestID)
	if err != nil || guest == nil {
		t.Fatal(err)
	}
	if !guest.IsGuest || guest.IsCreator() || guest.IsAdmin() || guest.Role != entities.RolePlayer || !strings.HasPrefix(guest.RefID, "guest:") {
		t.Fatalf("created guest was elevated: %+v", guest)
	}
	nick := guest.Nickname

	keys, err := sshkeys.Open(client.DB())
	if err != nil {
		t.Fatal(err)
	}
	devices := devicecode.New(devicecode.Config{})
	api := handler.NewSSHAPI()
	r := gin.New()
	protected := r.Group("/api/")
	protected.Use(AuthMiddleware(facade))
	protected.PUT("user", (&handler.UsersHandler{Service: facade.UsersService()}).UpdateUser)
	protected.GET("ssh/keys", api.ListKeys)
	protected.POST("ssh/keys", api.AddKey)
	protected.DELETE("ssh/keys/:id", api.DeleteKey)
	protected.POST("ssh/device/confirm", api.Confirm)
	protected.POST("ssh/device/deny", api.Deny)
	creator := protected.Group("creator-probe")
	creator.Use(CreatorMiddleware())
	creator.POST("touch", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	adminAPI := protected.Group("admin-probe")
	adminAPI.Use(AdminMiddleware())
	adminAPI.POST("touch", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	srv := httptest.NewServer(r)
	defer srv.Close()
	api.Set(handler.SSHSettings{
		Keys: keys, Devices: devices, ActivateURL: srv.URL + "/activate",
		KeysEnabled: true, DeviceEnabled: true, WebManage: true, MaxKeys: 10,
	})

	_, code, _, err := devices.Begin("203.0.113.9", "mud", "test", nil)
	if err != nil {
		t.Fatal(err)
	}
	line := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGuestSandboxProbeKeyNotRealData00000 guest"
	checks := []struct {
		method, path, body, want string
	}{
		{http.MethodPost, "/api/ssh/device/confirm", `{"user_code":"` + code + `","csrf":"nope"}`, "account cannot confirm"},
		{http.MethodPost, "/api/ssh/device/deny", `{"user_code":"` + code + `","csrf":"nope"}`, "account cannot confirm"},
		{http.MethodGet, "/api/ssh/keys", "", "account cannot confirm"},
		{http.MethodPost, "/api/ssh/keys", `{"public_key":"` + line + `","label":"nope"}`, "account cannot confirm"},
		{http.MethodDelete, "/api/ssh/keys/not-a-key", "", "account cannot confirm"},
		{http.MethodPost, "/api/creator-probe/touch", "", "Creator access required"},
		{http.MethodPost, "/api/admin-probe/touch", "", "Admin access required"},
		{http.MethodPut, "/api/user", `{"name":"New","email":"new@example.com","nickname":"Admin","picture":"new.png","role":"admin","isGuest":false}`, "guest accounts cannot be changed"},
	}
	for _, check := range checks {
		status, body := callAs(t, check.method, srv.URL+check.path, guestTok, check.body)
		if status != http.StatusForbidden || !strings.Contains(body, check.want) {
			t.Fatalf("%s %s status %d body %s", check.method, check.path, status, body)
		}
	}

	rows, err := keys.List(guest.RefID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("guest linked a key: %+v", rows)
	}
	if _, err := devices.Lookup(code, admin.RefID, "203.0.113.9", func(string) error { return nil }); err != nil {
		t.Fatalf("guest confirm consumed the code: %v", err)
	}
	stored, err := facade.UsersService().FindByID(guest.ID)
	if err != nil || stored == nil {
		t.Fatal(err)
	}
	if !stored.IsGuest || stored.RefID != guest.RefID || stored.Nickname != nick || stored.Role != entities.RolePlayer || stored.IsCreator() || stored.IsAdmin() {
		t.Fatalf("guest account changed: %+v", stored)
	}

	for _, path := range []string{"/api/creator-probe/touch", "/api/admin-probe/touch"} {
		status, body := callAs(t, http.MethodPost, srv.URL+path, adminTok, "")
		if status != http.StatusOK || !strings.Contains(body, `"ok":true`) {
			t.Fatalf("admin %s status %d body %s", path, status, body)
		}
	}
}

func callAs(t *testing.T, method, url, token, raw string) (int, string) {
	t.Helper()
	var body io.Reader
	if raw != "" {
		body = strings.NewReader(raw)
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatal(err)
	}
	if raw != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	got, _ := io.ReadAll(res.Body)
	return res.StatusCode, string(got)
}
