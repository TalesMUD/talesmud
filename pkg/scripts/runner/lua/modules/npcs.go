package modules

import (
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
)

// RegisterNPCsModule registers the tales.npcs module
func RegisterNPCsModule(L *lua.LState, runner *luarunner.LuaRunner) int {
	mod := L.NewTable()

	// tales.npcs.get(id) - Get NPC by ID
	mod.RawSetString("get", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		npc, err := facade.NPCsService().FindByID(id)
		if err != nil || npc == nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, npc))
		return 1
	}))

	// tales.npcs.findByName(name) - Find NPCs by name
	mod.RawSetString("findByName", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		npcs, err := facade.NPCsService().FindByName(name)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, npcs))
		return 1
	}))

	// tales.npcs.findInRoom(roomID) - Find NPCs in a room
	mod.RawSetString("findInRoom", L.NewFunction(func(L *lua.LState) int {
		roomID := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		npcs, err := facade.NPCsService().FindByRoom(roomID)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, npcs))
		return 1
	}))

	// tales.npcs.getAll() - Get all NPCs
	mod.RawSetString("getAll", L.NewFunction(func(L *lua.LState) int {
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		npcs, err := facade.NPCsService().FindAll()
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, npcs))
		return 1
	}))

	// tales.npcs.damage(id, amount) - Damage an NPC
	mod.RawSetString("damage", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		amount := L.CheckInt(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		npc, err := facade.NPCsService().FindByID(id)
		if err != nil || npc == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		npc.CurrentHitPoints -= int32(amount)
		if npc.CurrentHitPoints < 0 {
			npc.CurrentHitPoints = 0
		}

		err = facade.NPCsService().Update(id, npc)
		L.Push(lua.LBool(err == nil))
		return 1
	}))

	// tales.npcs.heal(id, amount) - Heal an NPC
	mod.RawSetString("heal", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		amount := L.CheckInt(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		npc, err := facade.NPCsService().FindByID(id)
		if err != nil || npc == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		npc.CurrentHitPoints += int32(amount)
		if npc.CurrentHitPoints > npc.MaxHitPoints {
			npc.CurrentHitPoints = npc.MaxHitPoints
		}

		err = facade.NPCsService().Update(id, npc)
		L.Push(lua.LBool(err == nil))
		return 1
	}))

	// tales.npcs.moveTo(id, roomID) - Move NPC to room
	mod.RawSetString("moveTo", L.NewFunction(func(L *lua.LState) int {
		npcID := L.CheckString(1)
		roomID := L.CheckString(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		npc, err := facade.NPCsService().FindByID(npcID)
		if err != nil || npc == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		// Update NPC's current room
		npc.CurrentRoomID = roomID
		err = facade.NPCsService().Update(npcID, npc)
		L.Push(lua.LBool(err == nil))
		return 1
	}))

	// tales.npcs.isDead(id) - Check if NPC is dead
	mod.RawSetString("isDead", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		npc, err := facade.NPCsService().FindByID(id)
		if err != nil || npc == nil {
			L.Push(lua.LBool(true)) // Treat missing NPC as dead
			return 1
		}

		L.Push(lua.LBool(npc.CurrentHitPoints <= 0))
		return 1
	}))

	// tales.npcs.isEnemy(id) - Check if NPC is an enemy
	mod.RawSetString("isEnemy", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		npc, err := facade.NPCsService().FindByID(id)
		if err != nil || npc == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		L.Push(lua.LBool(npc.IsEnemy()))
		return 1
	}))

	// tales.npcs.isMerchant(id) - Check if NPC is a merchant
	mod.RawSetString("isMerchant", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		npc, err := facade.NPCsService().FindByID(id)
		if err != nil || npc == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		L.Push(lua.LBool(npc.IsMerchant()))
		return 1
	}))

	// tales.npcs.delete(id) - Delete an NPC
	mod.RawSetString("delete", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		err := facade.NPCsService().Delete(id)
		L.Push(lua.LBool(err == nil))
		return 1
	}))

	L.Push(mod)
	return 1
}
