package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/ruleset"
	"github.com/talesmud/talesmud/pkg/service"
)

func TestLeaderboardShapeLimitAndPrivacy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	if err := ruleset.LoadBytes([]byte("ranking:\n  cache_seconds: 60\n")); err != nil {
		t.Fatal(err)
	}

	client, err := sqlite.Open(filepath.Join(t.TempDir(), "board.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	h := &LeaderboardHandler{
		Characters: facade.CharactersService(),
		Users:      facade.UsersService(),
		Now:        func() time.Time { return time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC) },
	}
	seedLeaderboard(t, facade)

	r := gin.New()
	r.GET("/api/leaderboard", h.Get)

	body := getLeaderboard(t, r, "/api/leaderboard?self=bea&limit=1")
	if body["limit"].(float64) != 1 {
		t.Fatalf("limit %#v", body["limit"])
	}
	rows := body["rows"].([]any)
	if len(rows) != 1 {
		t.Fatalf("rows %#v", rows)
	}
	top := rows[0].(map[string]any)
	if top["name"] != "Ced" || top["rank"].(float64) != 1 || top["level"].(float64) != 9 {
		t.Fatalf("top %#v", top)
	}
	you := body["you"].(map[string]any)
	if you["name"] != "Bea" || you["rank"].(float64) != 3 || you["class"] != "Rogue" {
		t.Fatalf("you %#v", you)
	}
	assertNoPrivateLeaderboardFields(t, body)
	raw := mustLeaderboardRaw(t, r, "/api/leaderboard?self=bea&limit=1")
	if strings.Contains(raw, "secret@example.com") || strings.Contains(raw, "belongsUser") || strings.Contains(raw, "u-play") {
		t.Fatalf("private data in %s", raw)
	}

	clamped := getLeaderboard(t, r, "/api/leaderboard?limit=500")
	if clamped["limit"].(float64) != 100 {
		t.Fatalf("clamp %#v", clamped["limit"])
	}
	if len(clamped["rows"].([]any)) != 4 {
		t.Fatalf("eligible rows %#v", clamped["rows"])
	}
	names := make([]string, 0)
	for _, row := range clamped["rows"].([]any) {
		names = append(names, row.(map[string]any)["name"].(string))
	}
	for _, banned := range []string{"Ghost", "Root", "Orphan"} {
		for _, name := range names {
			if name == banned {
				t.Fatalf("excluded %s in %#v", banned, names)
			}
		}
	}

	// Cache holds the first snapshot. A new high-level character stays hidden until the TTL passes.
	seedCharacter(t, facade, "u-play", "Zoe", 40, time.Time{})
	cached := getLeaderboard(t, r, "/api/leaderboard")
	if cached["limit"].(float64) != 25 {
		t.Fatalf("default limit %#v", cached["limit"])
	}
	for _, row := range cached["rows"].([]any) {
		if row.(map[string]any)["name"] == "Zoe" {
			t.Fatal("cache miss")
		}
	}
	if err := ruleset.LoadBytes([]byte("ranking:\n  cache_seconds: 0\n")); err != nil {
		t.Fatal(err)
	}
	fresh := getLeaderboard(t, r, "/api/leaderboard?limit=0")
	if fresh["limit"].(float64) != 1 {
		t.Fatalf("zero clamp %#v", fresh["limit"])
	}
	if fresh["rows"].([]any)[0].(map[string]any)["name"] != "Zoe" {
		t.Fatalf("fresh %#v", fresh["rows"])
	}
	garbage := getLeaderboard(t, r, "/api/leaderboard?limit=nope")
	if garbage["limit"].(float64) != 25 {
		t.Fatalf("garbage %#v", garbage["limit"])
	}
}

func seedLeaderboard(t *testing.T, facade service.Facade) {
	t.Helper()
	users := []*entities.User{
		{Entity: &entities.Entity{ID: "u-play"}, Role: entities.RolePlayer, Email: "secret@example.com", IsOnline: false},
		{Entity: &entities.Entity{ID: "u-guest"}, Role: entities.RolePlayer, IsGuest: true, Email: "guest@example.com"},
		{Entity: &entities.Entity{ID: "u-admin"}, Role: entities.RoleAdmin, Email: "admin@example.com"},
		{Entity: &entities.Entity{ID: "u-maker"}, Role: entities.RoleCreator},
	}
	for _, user := range users {
		if _, err := facade.UsersService().Import(user); err != nil {
			t.Fatal(err)
		}
	}
	early := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	seedCharacter(t, facade, "u-play", "Ced", 9, early)
	seedGeared(t, facade, "u-play", "Ada", 5, early.Add(time.Hour))
	seedCharacter(t, facade, "u-play", "Bea", 5, early.Add(2*time.Hour))
	seedCharacter(t, facade, "u-maker", "Maker", 2, early)
	seedCharacter(t, facade, "u-guest", "Ghost", 99, early)
	seedCharacter(t, facade, "u-admin", "Root", 98, early)
	seedCharacter(t, facade, "missing", "Orphan", 50, early)
}

func seedCharacter(t *testing.T, facade service.Facade, userID, name string, level int32, created time.Time) {
	t.Helper()
	character := &characters.Character{
		Entity:      &entities.Entity{ID: strings.ToLower(name)},
		BelongsUser: traits.BelongsUser{BelongsUserID: userID},
		Name:        name,
		Level:       level,
		Created:     created,
		Class:       characters.Class{ID: "rogue", Name: "Rogue"},
		Race:        characters.Race{ID: "human", Name: "Human"},
	}
	if _, err := facade.CharactersService().Import(character); err != nil {
		t.Fatal(err)
	}
}

func seedGeared(t *testing.T, facade service.Facade, userID, name string, level int32, created time.Time) {
	t.Helper()
	character := &characters.Character{
		Entity:      &entities.Entity{ID: strings.ToLower(name)},
		BelongsUser: traits.BelongsUser{BelongsUserID: userID},
		Name:        name,
		Level:       level,
		Created:     created,
		Class:       characters.Class{ID: "warrior", Name: "Warrior"},
		Race:        characters.Race{ID: "dwarf", Name: "Dwarf"},
		EquippedItems: map[items.ItemSlot]*items.Item{
			items.ItemSlotMainHand: {
				Entity: &entities.Entity{ID: "blade"}, Level: 2, Quality: items.ItemQualityRare,
				Attributes: map[string]any{"damage": 3},
			},
		},
	}
	if _, err := facade.CharactersService().Import(character); err != nil {
		t.Fatal(err)
	}
}

func getLeaderboard(t *testing.T, r *gin.Engine, path string) map[string]any {
	t.Helper()
	raw := mustLeaderboardRaw(t, r, path)
	var body map[string]any
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatal(err)
	}
	return body
}

func mustLeaderboardRaw(t *testing.T, r *gin.Engine, path string) string {
	t.Helper()
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("%s %d %s", path, rec.Code, rec.Body.String())
	}
	return rec.Body.String()
}

func assertNoPrivateLeaderboardFields(t *testing.T, v any) {
	t.Helper()
	switch node := v.(type) {
	case map[string]any:
		for key, child := range node {
			switch strings.ToLower(key) {
			case "email", "userid", "user_id", "belongsuser", "accountid", "refid", "password", "passwordhash", "token", "id":
				t.Fatalf("private field %s", key)
			}
			assertNoPrivateLeaderboardFields(t, child)
		}
	case []any:
		for _, child := range node {
			assertNoPrivateLeaderboardFields(t, child)
		}
	}
}
