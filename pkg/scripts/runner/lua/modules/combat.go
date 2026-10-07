package modules

import (
	"reflect"

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

	// tales.combat.healNpc(npcID, amount) -> number restored (0 if none)
	mod.RawSetString("healNpc", L.NewFunction(func(L *lua.LState) int {
		npcID := L.CheckString(1)
		amount := int32(L.CheckInt(2))
		game := runner.GetGame()
		if game == nil {
			L.Push(lua.LNumber(0))
			return 1
		}
		engine := game.GetCombatEngine()
		if engine == nil {
			L.Push(lua.LNumber(0))
			return 1
		}
		L.Push(lua.LNumber(engine.HealCombatNPC(npcID, amount)))
		return 1
	}))

	// tales.combat.applyEffect(targetID, effectID) -> bool
	// Only an existing buff or debuff skill id. Damage skills return false.
	mod.RawSetString("applyEffect", L.NewFunction(func(L *lua.LState) int {
		targetID := L.CheckString(1)
		effectID := L.CheckString(2)
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
		L.Push(lua.LBool(engine.ApplyCombatEffect(targetID, effectID)))
		return 1
	}))

	// tales.combat.summon(templateId, count) -> number spawned.
	// Uses the running enemy-hook context. Count clamps to 1..3 and the fight cap.
	// Unknown template, a non-enemy, no fight, or a full cap returns 0.
	mod.RawSetString("summon", L.NewFunction(func(L *lua.LState) int {
		templateID := L.OptString(1, "")
		count := L.OptInt(2, 1)
		game := runner.GetGame()
		if game == nil || templateID == "" {
			logrus.WithField("template", templateID).Warn("summon: no active fight")
			L.Push(lua.LNumber(0))
			return 1
		}
		engine := game.GetCombatEngine()
		if engine == nil {
			logrus.WithField("template", templateID).Warn("summon: no active fight")
			L.Push(lua.LNumber(0))
			return 1
		}
		summonerID := luaHookNPCID(L)
		spawned := engine.SummonCombatAllies(summonerID, templateID, count)
		L.Push(lua.LNumber(spawned))
		return 1
	}))

	L.Push(mod)
	return 1
}

// luaHookNPCID reads ctx.npc.ID from the running enemy hook.
func luaHookNPCID(L *lua.LState) string {
	if L == nil {
		return ""
	}
	ctx := L.GetGlobal("ctx")
	tbl, ok := ctx.(*lua.LTable)
	if !ok {
		return ""
	}
	npc := L.GetField(tbl, "npc")
	if npc == lua.LNil {
		return ""
	}
	if s := luaUserFieldString(npc, "ID"); s != "" {
		return s
	}
	id := L.GetField(npc, "ID")
	if s, ok := id.(lua.LString); ok {
		return string(s)
	}
	return ""
}

func luaUserFieldString(v lua.LValue, field string) string {
	ud, ok := v.(*lua.LUserData)
	if !ok || ud == nil || ud.Value == nil {
		return ""
	}
	rv := reflect.ValueOf(ud.Value)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return ""
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return ""
	}
	f := rv.FieldByName(field)
	if !f.IsValid() || f.Kind() != reflect.String {
		return ""
	}
	return f.String()
}
