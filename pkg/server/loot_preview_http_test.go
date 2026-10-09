package server_test

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/server"
	"github.com/talesmud/talesmud/pkg/server/handler"
	"github.com/talesmud/talesmud/pkg/service"
	"github.com/talesmud/talesmud/pkg/worldindex"
)

func newLootPreviewHTTP(t *testing.T) (*gin.Engine, service.Facade) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	client, err := sqlite.Open(filepath.Join(t.TempDir(), "loot.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	h := &handler.LootTablesHandler{
		Service: facade.LootTablesService(),
		Facade:  facade,
	}
	r := gin.New()
	api := r.Group("/api")
	api.Use(func(c *gin.Context) {
		role := c.GetHeader("X-Role")
		if role == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set("user", &entities.User{
			Entity:   &entities.Entity{ID: "user-1"},
			Nickname: "Tester",
			Role:     role,
		})
		c.Next()
	})
	preview := api.Group("")
	preview.Use(server.CreatorMiddleware())
	preview.POST("loot-tables/:id/roll", h.RollLootPreview)
	return r, facade
}

func TestLootPreviewGuardsAndDeterminism(t *testing.T) {
	r, facade := newLootPreviewHTTP(t)
	if _, err := facade.ItemsService().Import(&items.Item{
		Entity: &entities.Entity{ID: "GEM"}, Name: "Gem", IsTemplate: true,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := facade.LootTablesService().Import(&items.LootTable{
		Entity: &entities.Entity{ID: "LT1"},
		Name:   "Gems",
		Entries: []items.LootEntry{
			{ItemTemplateID: "GEM", DropChance: 0, MinQuantity: 1, MaxQuantity: 3, Guaranteed: true},
			{ItemTemplateID: "CROWN", DropChance: 1, MinQuantity: 1, MaxQuantity: 1, BossOnly: true},
		},
	}); err != nil {
		t.Fatal(err)
	}
	before, err := facade.ItemsService().FindAll(repository.ItemsQuery{})
	if err != nil {
		t.Fatal(err)
	}

	rec := doJSON(r, http.MethodPost, "/api/loot-tables/LT1/roll?n=25&seed=7", "", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous: %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(r, http.MethodPost, "/api/loot-tables/LT1/roll?n=25&seed=7", entities.RolePlayer, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("player: %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(r, http.MethodPost, "/api/loot-tables/missing/roll?n=1&seed=1", entities.RoleCreator, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing: %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(r, http.MethodPost, "/api/loot-tables/LT1/roll?n=0&seed=1", entities.RoleCreator, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("n: %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(r, http.MethodPost, "/api/loot-tables/LT1/roll?n=10001&seed=1", entities.RoleAdmin, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("cap: %d %s", rec.Code, rec.Body.String())
	}

	load := func() (worldindex.Snapshot, error) { return worldindex.NewSnapshot(), nil }
	if _, err := worldindex.Live.Index(load); err != nil {
		t.Fatal(err)
	}
	loads := worldindex.Live.Loads()

	rec = doJSON(r, http.MethodPost, "/api/loot-tables/LT1/roll?n=25&seed=7", entities.RoleCreator, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("creator: %d %s", rec.Code, rec.Body.String())
	}
	again := doJSON(r, http.MethodPost, "/api/loot-tables/LT1/roll?n=25&seed=7&boss=1", entities.RoleAdmin, nil)
	if again.Code != http.StatusOK {
		t.Fatalf("admin: %d %s", again.Code, again.Body.String())
	}
	same := doJSON(r, http.MethodPost, "/api/loot-tables/LT1/roll?n=25&seed=7", entities.RoleCreator, nil)
	if same.Body.String() != rec.Body.String() {
		t.Fatalf("seed diverged\n%s\n%s", rec.Body.String(), same.Body.String())
	}

	var body service.LootPreview
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.N != 25 || body.Seed != 7 || body.Boss || len(body.Drops) != 2 {
		t.Fatalf("body = %+v", body)
	}
	if body.Drops[0].ItemTemplateID != "GEM" || body.Drops[0].Name != "Gem" || body.Drops[0].Rolls != 25 || body.Drops[0].Frequency != 1 {
		t.Fatalf("gem = %+v", body.Drops[0])
	}
	if body.Drops[1].Rolls != 0 {
		t.Fatalf("boss-only dropped: %+v", body.Drops[1])
	}
	var bossed service.LootPreview
	if err := json.Unmarshal(again.Body.Bytes(), &bossed); err != nil {
		t.Fatal(err)
	}
	if !bossed.Boss || bossed.Drops[1].Rolls != 25 {
		t.Fatalf("boss preview = %+v", bossed)
	}

	after, err := facade.ItemsService().FindAll(repository.ItemsQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) {
		t.Fatalf("items %d -> %d", len(before), len(after))
	}
	table, err := facade.LootTablesService().FindByID("LT1")
	if err != nil || table == nil || len(table.Entries) != 2 || table.Entries[0].DropChance != 0 {
		t.Fatalf("table changed: %+v %v", table, err)
	}
	rows, err := facade.AuditService().List("", "", 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("audit rows = %+v", rows)
	}
	if _, err := worldindex.Live.Index(load); err != nil {
		t.Fatal(err)
	}
	if worldindex.Live.Loads() != loads {
		t.Fatalf("preview reloaded search: %d -> %d", loads, worldindex.Live.Loads())
	}
}
