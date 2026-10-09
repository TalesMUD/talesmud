package modules

import (
	"strings"

	lua "github.com/yuin/gopher-lua"

	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
)

// RegisterUsersModule registers tales.users. isGuest is the pack-side check
// for guests.persistent_effects. A missing user is not a guest.
func RegisterUsersModule(L *lua.LState, runner *luarunner.LuaRunner) int {
	mod := L.NewTable()
	mod.RawSetString("isGuest", L.NewFunction(func(L *lua.LState) int {
		id := ""
		if L.GetTop() >= 1 && L.Get(1).Type() == lua.LTString {
			id = strings.TrimSpace(L.CheckString(1))
		}
		L.Push(lua.LBool(userIsGuest(runner, id)))
		return 1
	}))
	L.Push(mod)
	return 1
}

func userIsGuest(runner *luarunner.LuaRunner, id string) bool {
	if runner == nil || id == "" {
		return false
	}
	facade := runner.GetFacade()
	if facade == nil || facade.UsersService() == nil {
		return false
	}
	user, err := facade.UsersService().FindByID(id)
	if err != nil || user == nil {
		return false
	}
	return user.IsGuest
}
