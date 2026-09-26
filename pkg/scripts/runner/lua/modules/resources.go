package modules

import (
	"errors"

	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"

	"github.com/talesmud/talesmud/pkg/resources"
	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
)

// RegisterResourcesModule registers tales.resources.
// get and consume answer an unconfigured key with ok=false and do not create a balance.
func RegisterResourcesModule(L *lua.LState, runner *luarunner.LuaRunner) int {
	mod := L.NewTable()

	// tales.resources.get(characterID, key) -> allowance, remaining, ok
	mod.RawSetString("get", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		key := L.CheckString(2)
		store := runner.ResourceStore()
		if store == nil {
			pushBalance(L, 0, 0, false)
			return 3
		}
		res, ok, err := store.Get(characterID, key)
		if err != nil || !ok {
			if err != nil {
				log.WithError(err).WithField("key", key).Warn("resource get failed")
			}
			pushBalance(L, 0, 0, false)
			return 3
		}
		pushBalance(L, res.Allowance, res.Remaining, true)
		return 3
	}))

	// tales.resources.consume(characterID, key, n) -> remaining, ok
	mod.RawSetString("consume", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		key := L.CheckString(2)
		n := L.CheckInt(3)
		store := runner.ResourceStore()
		if store == nil {
			L.Push(lua.LNumber(0))
			L.Push(lua.LBool(false))
			return 2
		}
		res, err := store.Consume(characterID, key, n)
		if err != nil {
			if !errors.Is(err, resources.ErrUnknown) && !errors.Is(err, resources.ErrExhausted) {
				log.WithError(err).WithField("key", key).Warn("resource consume failed")
			}
			L.Push(lua.LNumber(res.Remaining))
			L.Push(lua.LBool(false))
			return 2
		}
		L.Push(lua.LNumber(res.Remaining))
		L.Push(lua.LBool(true))
		return 2
	}))

	L.Push(mod)
	return 1
}

func pushBalance(L *lua.LState, allowance, remaining int, ok bool) {
	L.Push(lua.LNumber(allowance))
	L.Push(lua.LNumber(remaining))
	L.Push(lua.LBool(ok))
}
