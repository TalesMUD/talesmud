package modules

import (
	"github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"

	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
)

// RegisterCombatModule registers tales.combat for content-authored combat procs.
func RegisterCombatModule(L *lua.LState, runner *luarunner.LuaRunner) int {
	mod := L.NewTable()

	// tales.combat.applyDot(attackerID, targetID, effectID, name, damage, duration) -> bool
	// Refreshes duration when the same effectID is reapplied (no stack spam).
	mod.RawSetString("applyDot", L.NewFunction(func(L *lua.LState) int {
		attackerID := L.CheckString(1)
		targetID := L.CheckString(2)
		effectID := L.CheckString(3)
		name := L.OptString(4, effectID)
		damage := int32(L.CheckInt(5))
		duration := L.CheckInt(6)
		game := runner.GetGame()
		if game == nil {
			L.Push(lua.LBool(false))
			return 1
		}
		engine := game.GetCombatEngine()
		if engine == nil {
			L.Push(lua.LBool(false))
			return 1
		}
		ok := engine.ApplyCombatDot(attackerID, targetID, effectID, name, damage, duration)
		if !ok {
			logrus.WithFields(logrus.Fields{
				"attackerID": attackerID,
				"targetID":   targetID,
				"effectID":   effectID,
			}).Debug("[Script] combat.applyDot failed")
		}
		L.Push(lua.LBool(ok))
		return 1
	}))

	L.Push(mod)
	return 1
}
