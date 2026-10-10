package server

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

// TestGuestTokenDoesNotBindRealUser is the SSHA-07 regression.
// A guest HMAC whose uid is a normal account must not become that account.
func TestGuestTokenDoesNotBindRealUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prev := localSessions
	t.Cleanup(func() { localSessions = prev })
	localSessions = nil
	const secret = "ssha07-test-secret"
	t.Setenv("GUEST_SECRET", secret)

	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "ssha07.db"))
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
	user, err := facade.UsersService().Create(&entities.User{
		RefID:    "auth0|real-account",
		Username: "realaccount",
		Nickname: "Real",
		Role:     entities.RolePlayer,
	})
	if err != nil || user == nil || user.IsGuest || user.ID == "" {
		t.Fatalf("user %+v %v", user, err)
	}

	r := gin.New()
	r.GET("/who", AuthMiddleware(facade), func(c *gin.Context) {
		got, _ := c.Get("user")
		loaded, _ := got.(*entities.User)
		if loaded == nil {
			c.Status(http.StatusNoContent)
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": loaded.ID, "guest": loaded.IsGuest, "ref": loaded.RefID})
	})

	forged := signGuestUID(t, secret, user.ID)
	rec := getWho(r, forged)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("forged uid status %d body %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), user.ID) || strings.Contains(rec.Body.String(), user.RefID) {
		t.Fatalf("forged token identified the real account: %s", rec.Body.String())
	}

	guestTok, err := facade.GuestService().CreateGuestSessionPick("203.0.113.20", "", "")
	if err != nil {
		t.Fatal(err)
	}
	guestRec := getWho(r, guestTok)
	if guestRec.Code != http.StatusOK || !strings.Contains(guestRec.Body.String(), `"guest":true`) {
		t.Fatalf("real guest %d %s", guestRec.Code, guestRec.Body.String())
	}
}

func signGuestUID(t *testing.T, secret, uid string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   "guest:forged",
		"uid":   uid,
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Add(-time.Minute).Unix(),
		"guest": true,
	})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func getWho(r http.Handler, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/who", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}
