package modules

import (
	"path/filepath"
	"testing"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/scripts"
	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
	"github.com/talesmud/talesmud/pkg/service"
)

func TestCharactersTopOrdersByLevelThenXP(t *testing.T) {
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "top.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	for _, ch := range []*characters.Character{
		{Entity: &entities.Entity{ID: "low"}, Name: "Low", BelongsUser: *traits.BelongsToUser("u1"), Level: 2, XP: 50},
		{Entity: &entities.Entity{ID: "high"}, Name: "High", BelongsUser: *traits.BelongsToUser("u2"), Level: 4, XP: 10},
		{Entity: &entities.Entity{ID: "mid"}, Name: "Mid", BelongsUser: *traits.BelongsToUser("u3"), Level: 4, XP: 80},
	} {
		if _, err := facade.CharactersService().Store(ch); err != nil {
			t.Fatal(err)
		}
	}
	runner := luarunner.NewLuaRunner()
	RegisterAllModules(runner)
	runner.SetServices(facade, nil)
	defer runner.Shutdown()
	result := runner.RunWithResult(scripts.Script{
		Name:     "top",
		Language: scripts.ScriptLanguageLua,
		Code: `
local rows = tales.characters.top(2, "level")
return rows[1].name .. "," .. rows[2].name
`,
	}, scripts.NewScriptContext())
	if result == nil || !result.Success {
		t.Fatalf("script: %+v", result)
	}
	if result.Result != "Mid,High" {
		t.Fatalf("top = %#v", result.Result)
	}
}
