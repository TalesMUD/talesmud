package modules

import (
	"time"

	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"

	"github.com/talesmud/talesmud/pkg/instances"
	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
)

// RegisterInstancesModule registers tales.instances.
// generate builds a private room line. It does not move the character,
// spend a resource, or pull followers. The caller script does those.
func RegisterInstancesModule(L *lua.LState, runner *luarunner.LuaRunner) int {
	mod := L.NewTable()

	// tales.instances.generate(characterID, playerLevel, spec) -> entryRoomID, err
	mod.RawSetString("generate", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		level := int32(L.CheckInt(2))
		spec, err := procSpecFromLua(L.CheckTable(3))
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		game := runner.GetGame()
		if game == nil || game.GetRoomInstances() == nil {
			L.Push(lua.LNil)
			L.Push(lua.LString("instances unavailable"))
			return 2
		}
		res, err := game.GetRoomInstances().Generate(characterID, level, spec)
		if err != nil {
			log.WithError(err).WithField("characterID", characterID).Warn("generate instance failed")
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(res.EntryRoomID))
		L.Push(lua.LNil)
		return 2
	}))

	L.Push(mod)
	return 1
}

func procSpecFromLua(tbl *lua.LTable) (instances.ProcSpec, error) {
	spec := instances.ProcSpec{Seed: int64(lua.LVAsNumber(tbl.RawGetString("seed")))}
	if n := int(lua.LVAsNumber(tbl.RawGetString("count"))); n > 0 {
		spec.Count = n
	}
	if secs := time.Duration(lua.LVAsNumber(tbl.RawGetString("timeout"))); secs > 0 {
		spec.Timeout = secs * time.Second
	}
	if ret, ok := tbl.RawGetString("returnRoom").(lua.LString); ok {
		spec.ReturnRoomID = string(ret)
	}
	if roomsTbl, ok := tbl.RawGetString("templates").(*lua.LTable); ok {
		roomsTbl.ForEach(func(_, value lua.LValue) {
			if s, ok := value.(lua.LString); ok && s != "" {
				spec.TemplateIDs = append(spec.TemplateIDs, string(s))
			}
		})
	}
	if encTbl, ok := tbl.RawGetString("encounters").(*lua.LTable); ok {
		encTbl.ForEach(func(_, value lua.LValue) {
			row, ok := value.(*lua.LTable)
			if !ok {
				return
			}
			id, _ := row.RawGetString("id").(lua.LString)
			spec.Encounters = append(spec.Encounters, instances.Encounter{
				TemplateID: string(id),
				MinLevel:   int32(lua.LVAsNumber(row.RawGetString("minLevel"))),
				MaxLevel:   int32(lua.LVAsNumber(row.RawGetString("maxLevel"))),
				Weight:     int(lua.LVAsNumber(row.RawGetString("weight"))),
			})
		})
	}
	return spec, nil
}
