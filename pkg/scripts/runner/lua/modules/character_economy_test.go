package modules

import (
	"fmt"
	"path/filepath"
	"testing"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/ruleset"
	"github.com/talesmud/talesmud/pkg/scripts"
	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
	"github.com/talesmud/talesmud/pkg/service"
)

func TestLuaGoldBindAndApplyLevels(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	ruleset.SetLevelUpMode(ruleset.ModeTrainer)

	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "economy.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	if _, err := facade.RoomsService().Store(&rooms.Room{
		Entity: &entities.Entity{ID: "room-1"},
		Name:   "Shrine",
	}); err != nil {
		t.Fatal(err)
	}
	need := leveling.GetXPRequired(2)
	created, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "c1"},
		Name:             "Novice",
		BelongsUser:      *traits.BelongsToUser("user-1"),
		Level:            1,
		XP:               need,
		Gold:             10,
		MaxHitPoints:     20,
		CurrentHitPoints: 20,
		Class:            characters.ClassWarrior,
		Attributes: []characters.Attribute{
			{Short: "STR", Value: 10},
			{Short: "DEX", Value: 10},
			{Short: "INT", Value: 10},
			{Short: "WIS", Value: 10},
			{Short: "STA", Value: 10},
		},
		Inventory: items.Inventory{Size: 10},
	})
	if err != nil {
		t.Fatal(err)
	}
	id := created.ID

	runner := luarunner.NewLuaRunner()
	RegisterAllModules(runner)
	runner.SetServices(facade, nil)
	defer runner.Shutdown()

	result := runner.RunWithResult(scripts.Script{
		Name:     "economy",
		Language: scripts.ScriptLanguageLua,
		Code: fmt.Sprintf(`
local paid = tales.characters.addGold(%q, -3)
local refused = tales.characters.addGold(%q, -100)
local bound = tales.characters.setBind(%q, "room-1")
local missing = tales.characters.setBind(%q, "nope")
local cleared = tales.characters.setBind(%q, "")
local levels = tales.characters.applyLevels(%q)
local again = tales.characters.applyLevels(%q)
local reset = tales.characters.setProgress(%q, 1, 0, 25)
local capped = tales.characters.setProgress(%q, 99, -5, 25)
return {paid, refused, bound, missing, cleared, levels, again, reset, capped}
`, id, id, id, id, id, id, id, id, id),
	}, scripts.NewScriptContext())
	if result == nil || !result.Success {
		t.Fatalf("script: %+v", result)
	}
	row, ok := result.Result.([]interface{})
	if !ok || len(row) != 9 {
		t.Fatalf("result %#v", result.Result)
	}
	want := []interface{}{true, false, true, false, true, float64(1), float64(0), true, true}
	for i := range want {
		if row[i] != want[i] {
			t.Fatalf("index %d = %#v want %#v full %#v", i, row[i], want[i], row)
		}
	}
	stored, err := facade.CharactersService().FindByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Gold != 7 || stored.BoundRoomID != "" || stored.Level != 50 || stored.XP != 0 || stored.MaxHitPoints != 25 {
		t.Fatalf("gold=%d bind=%q level=%d xp=%d hp=%d", stored.Gold, stored.BoundRoomID, stored.Level, stored.XP, stored.MaxHitPoints)
	}
	if stored.Class != characters.ClassWarrior {
		t.Fatalf("class changed: %+v", stored.Class)
	}
}
