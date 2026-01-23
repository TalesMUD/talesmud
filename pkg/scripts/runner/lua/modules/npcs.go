package modules

import (
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
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

	// === Template and Instance Functions ===

	// tales.npcs.getTemplates() - Get all NPC templates
	mod.RawSetString("getTemplates", L.NewFunction(func(L *lua.LState) int {
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		templates, err := facade.NPCsService().FindAllTemplates()
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, templates))
		return 1
	}))

	// tales.npcs.isTemplate(id) - Check if NPC is a template
	mod.RawSetString("isTemplate", L.NewFunction(func(L *lua.LState) int {
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

		L.Push(lua.LBool(npc.IsTemplate))
		return 1
	}))

	// tales.npcs.spawnFromTemplate(templateId, roomId) - Spawn instance from template
	mod.RawSetString("spawnFromTemplate", L.NewFunction(func(L *lua.LState) int {
		templateID := L.CheckString(1)
		roomID := L.CheckString(2)

		game := runner.GetGame()
		if game == nil {
			L.Push(lua.LNil)
			return 1
		}

		npcMgr := game.GetNPCInstanceManager()
		if npcMgr == nil {
			L.Push(lua.LNil)
			return 1
		}

		instance, err := npcMgr.SpawnInstanceDirect(templateID, roomID)
		if err != nil || instance == nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, instance))
		return 1
	}))

	// tales.npcs.getInstance(id) - Get instance from memory (not DB)
	mod.RawSetString("getInstance", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)

		game := runner.GetGame()
		if game == nil {
			L.Push(lua.LNil)
			return 1
		}

		npcMgr := game.GetNPCInstanceManager()
		if npcMgr == nil {
			L.Push(lua.LNil)
			return 1
		}

		instance := npcMgr.GetInstance(id)
		if instance == nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, instance))
		return 1
	}))

	// tales.npcs.getInstancesInRoom(roomId) - Get all instances in room
	mod.RawSetString("getInstancesInRoom", L.NewFunction(func(L *lua.LState) int {
		roomID := L.CheckString(1)

		game := runner.GetGame()
		if game == nil {
			L.Push(lua.LNil)
			return 1
		}

		npcMgr := game.GetNPCInstanceManager()
		if npcMgr == nil {
			L.Push(lua.LNil)
			return 1
		}

		instances := npcMgr.GetInstancesInRoom(roomID)
		L.Push(luar.New(L, instances))
		return 1
	}))

	// tales.npcs.kill(id) - Kill an NPC instance
	mod.RawSetString("kill", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)

		game := runner.GetGame()
		if game == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		npcMgr := game.GetNPCInstanceManager()
		if npcMgr == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		success := npcMgr.KillInstance(id)
		L.Push(lua.LBool(success))
		return 1
	}))

	// tales.npcs.setState(id, state) - Set NPC instance state
	mod.RawSetString("setState", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		state := L.CheckString(2)

		game := runner.GetGame()
		if game == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		npcMgr := game.GetNPCInstanceManager()
		if npcMgr == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		success := npcMgr.UpdateInstance(id, func(n *npc.NPC) {
			n.State = state
		})
		L.Push(lua.LBool(success))
		return 1
	}))

	// tales.npcs.getState(id) - Get NPC state
	mod.RawSetString("getState", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)

		// First try instance manager
		game := runner.GetGame()
		if game != nil {
			npcMgr := game.GetNPCInstanceManager()
			if npcMgr != nil {
				instance := npcMgr.GetInstance(id)
				if instance != nil {
					L.Push(lua.LString(instance.State))
					return 1
				}
			}
		}

		// Fall back to persisted NPC
		facade := runner.GetFacade()
		if facade != nil {
			npc, err := facade.NPCsService().FindByID(id)
			if err == nil && npc != nil {
				L.Push(lua.LString(npc.State))
				return 1
			}
		}

		L.Push(lua.LString(""))
		return 1
	}))

	// tales.npcs.damageInstance(id, amount) - Damage an instance
	mod.RawSetString("damageInstance", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		amount := L.CheckInt(2)

		game := runner.GetGame()
		if game == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		npcMgr := game.GetNPCInstanceManager()
		if npcMgr == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		died := npcMgr.DamageInstance(id, int32(amount))
		L.Push(lua.LBool(died))
		return 1
	}))

	// tales.npcs.healInstance(id, amount) - Heal an instance
	mod.RawSetString("healInstance", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		amount := L.CheckInt(2)

		game := runner.GetGame()
		if game == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		npcMgr := game.GetNPCInstanceManager()
		if npcMgr == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		success := npcMgr.HealInstance(id, int32(amount))
		L.Push(lua.LBool(success))
		return 1
	}))

	// tales.npcs.moveInstance(id, roomId) - Move an instance to a room
	mod.RawSetString("moveInstance", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		roomID := L.CheckString(2)

		game := runner.GetGame()
		if game == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		npcMgr := game.GetNPCInstanceManager()
		if npcMgr == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		success := npcMgr.MoveInstance(id, roomID)
		L.Push(lua.LBool(success))
		return 1
	}))

	L.Push(mod)
	return 1
}
