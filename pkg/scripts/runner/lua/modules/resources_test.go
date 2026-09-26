package modules

import (
	"path/filepath"
	"testing"
	"time"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/resources"
	"github.com/talesmud/talesmud/pkg/scripts"
	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
)

func TestLuaResourcesUnknownAndConsume(t *testing.T) {
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "lua-resources.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	store, err := resources.New(client.DB())
	if err != nil {
		t.Fatal(err)
	}
	store.SetNow(func() time.Time { return time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC) })
	store.Configure([]resources.Allowance{{
		Key: "gatherings", Amount: 2, Reset: resources.ResetCalendar, Timezone: "UTC",
	}})

	runner := luarunner.NewLuaRunner()
	RegisterAllModules(runner)
	runner.SetResourceStore(store)
	defer runner.Shutdown()

	code := `
local allowance, remaining, ok = tales.resources.get("c1", "gatherings")
local spent, spentOK = tales.resources.consume("c1", "gatherings", 1)
local missingRem, missingOK = tales.resources.consume("c1", "nope", 1)
local overRem, overOK = tales.resources.consume("c1", "gatherings", 5)
return {allowance, remaining, ok, spent, spentOK, missingRem, missingOK, overRem, overOK}
`
	result := runner.RunWithResult(scripts.Script{
		Name:     "resources",
		Language: scripts.ScriptLanguageLua,
		Code:     code,
	}, scripts.NewScriptContext())
	if result == nil || !result.Success {
		t.Fatalf("script failed: %+v", result)
	}
	row, ok := result.Result.([]interface{})
	if !ok || len(row) != 9 {
		t.Fatalf("result = %#v", result.Result)
	}
	want := []interface{}{
		float64(2), float64(2), true,
		float64(1), true,
		float64(0), false,
		float64(1), false,
	}
	for i := range want {
		if row[i] != want[i] {
			t.Fatalf("index %d = %#v, want %#v (full %#v)", i, row[i], want[i], row)
		}
	}
	// Exhausted consume must not create a row for the unknown key or drop the real balance.
	left, found, err := store.Get("c1", "gatherings")
	if err != nil || !found || left.Remaining != 1 {
		t.Fatalf("stored %+v found=%v err=%v", left, found, err)
	}
	if _, found, err := store.Get("c1", "nope"); err != nil || found {
		t.Fatalf("unknown key became configured: found=%v err=%v", found, err)
	}
}

func TestLuaResourcesWithoutStore(t *testing.T) {
	runner := luarunner.NewLuaRunner()
	RegisterAllModules(runner)
	defer runner.Shutdown()
	result := runner.RunWithResult(scripts.Script{
		Name:     "resources-nil",
		Language: scripts.ScriptLanguageLua,
		Code: `
local allowance, remaining, ok = tales.resources.get("c", "gatherings")
local left, consumed = tales.resources.consume("c", "gatherings", 1)
return {allowance, remaining, ok, left, consumed}
`,
	}, scripts.NewScriptContext())
	if result == nil || !result.Success {
		t.Fatalf("script failed: %+v", result)
	}
	row, ok := result.Result.([]interface{})
	if !ok || len(row) != 5 {
		t.Fatalf("result = %#v", result.Result)
	}
	if row[0] != float64(0) || row[1] != float64(0) || row[2] != false || row[3] != float64(0) || row[4] != false {
		t.Fatalf("nil store result = %#v", row)
	}
}
