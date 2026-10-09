package modules

import (
	"os"
	"path/filepath"
	"testing"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/gamemode"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/scripts"
	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
	"github.com/talesmud/talesmud/pkg/service"
)

func TestCharactersTopSkipsGuestsUntilPersistentEffects(t *testing.T) {
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "top.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	guest, err := facade.UsersService().Create(&entities.User{Entity: entities.NewEntity(), RefID: "guest:1", Nickname: "Guest", IsGuest: true, Role: entities.RolePlayer})
	if err != nil {
		t.Fatal(err)
	}
	hero, err := facade.UsersService().Create(&entities.User{Entity: entities.NewEntity(), RefID: "auth0|hero", Nickname: "Hero", Role: entities.RolePlayer})
	if err != nil {
		t.Fatal(err)
	}
	for _, ch := range []*characters.Character{
		{Entity: entities.NewEntity(), Name: "Visitor", BelongsUser: *traits.BelongsToUser(guest.ID), Level: 9, XP: 1},
		{Entity: entities.NewEntity(), Name: "Hero", BelongsUser: *traits.BelongsToUser(hero.ID), Level: 2, XP: 1},
		{Entity: entities.NewEntity(), Name: "Orphan", BelongsUser: *traits.BelongsToUser("missing-user"), Level: 3, XP: 1},
	} {
		if _, err := facade.CharactersService().Store(ch); err != nil {
			t.Fatal(err)
		}
	}
	runner := luarunner.NewLuaRunner()
	RegisterAllModules(runner)
	runner.SetServices(facade, nil)
	defer runner.Shutdown()

	off := runner.RunWithResult(scripts.Script{
		Name: "top-off", Language: scripts.ScriptLanguageLua,
		Code: `
local rows = tales.characters.top(10, "level")
local names = {}
for i = 1, #rows do names[i] = rows[i].name end
return table.concat(names, ",")
`,
	}, scripts.NewScriptContext())
	if off == nil || !off.Success || off.Result != "Orphan,Hero" {
		t.Fatalf("guest leaked onto the board: %+v", off)
	}
	guestCheck := runner.RunWithResult(scripts.Script{
		Name: "is-guest", Language: scripts.ScriptLanguageLua,
		Code: `if tales.users.isGuest("` + guest.ID + `") and not tales.users.isGuest("` + hero.ID + `") and not tales.users.isGuest("missing-user") then return "ok" end return "bad"`,
	}, scripts.NewScriptContext())
	if guestCheck == nil || guestCheck.Result != "ok" {
		t.Fatalf("isGuest: %+v", guestCheck)
	}

	path := filepath.Join(t.TempDir(), "guests.yaml")
	if err := os.WriteFile(path, []byte("presentation: classic\nauth: auth0\nguests:\n  persistent_effects: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := gamemode.ApplyFile(path); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		reset := filepath.Join(t.TempDir(), "reset.yaml")
		_ = os.WriteFile(reset, []byte("presentation: classic\nauth: auth0\n"), 0o644)
		_ = gamemode.ApplyFile(reset)
	})
	on := runner.RunWithResult(scripts.Script{
		Name: "top-on", Language: scripts.ScriptLanguageLua,
		Code: `
local rows = tales.characters.top(10, "level")
local names = {}
for i = 1, #rows do names[i] = rows[i].name end
return table.concat(names, ",")
`,
	}, scripts.NewScriptContext())
	if on == nil || !on.Success || on.Result != "Visitor,Orphan,Hero" {
		t.Fatalf("persistent effects: %+v", on)
	}
}
