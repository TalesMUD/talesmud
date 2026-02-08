package modules

import (
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
)

// RegisterQuestsModule registers the tales.quests module
func RegisterQuestsModule(L *lua.LState, runner *luarunner.LuaRunner) int {
	mod := L.NewTable()

	// tales.quests.get(questID) - Get quest definition by ID
	mod.RawSetString("get", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		quest, err := facade.QuestsService().FindByID(id)
		if err != nil || quest == nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, quest))
		return 1
	}))

	// tales.quests.getAll() - Get all quest definitions
	mod.RawSetString("getAll", L.NewFunction(func(L *lua.LState) int {
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		quests, err := facade.QuestsService().FindAll()
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, quests))
		return 1
	}))

	// tales.quests.accept(characterID, questID) - Accept a quest
	mod.RawSetString("accept", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		questID := L.CheckString(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		_, err := facade.QuestsService().AcceptQuest(characterID, questID)
		L.Push(lua.LBool(err == nil))
		return 1
	}))

	// tales.quests.complete(characterID, questID) - Force-complete a quest
	mod.RawSetString("complete", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		questID := L.CheckString(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		_, err := facade.QuestsService().CompleteQuest(characterID, questID)
		L.Push(lua.LBool(err == nil))
		return 1
	}))

	// tales.quests.getProgress(characterID, questID) - Get quest progress
	mod.RawSetString("getProgress", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		questID := L.CheckString(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		progress, err := facade.QuestsService().GetProgress(characterID, questID)
		if err != nil || progress == nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, progress))
		return 1
	}))

	// tales.quests.setProgress(characterID, questID, objectiveID, amount) - Set objective progress
	mod.RawSetString("setProgress", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		questID := L.CheckString(2)
		objectiveID := L.CheckString(3)
		amount := int32(L.CheckInt(4))
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		_, err := facade.QuestsService().IncrementObjective(characterID, questID, objectiveID, amount)
		L.Push(lua.LBool(err == nil))
		return 1
	}))

	// tales.quests.isActive(characterID, questID) - Check if quest is active
	mod.RawSetString("isActive", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		questID := L.CheckString(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		progress, err := facade.QuestsService().GetProgress(characterID, questID)
		if err != nil || progress == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		L.Push(lua.LBool(progress.Status == "active"))
		return 1
	}))

	// tales.quests.isCompleted(characterID, questID) - Check if quest is completed
	mod.RawSetString("isCompleted", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		questID := L.CheckString(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		progress, err := facade.QuestsService().GetProgress(characterID, questID)
		if err != nil || progress == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		L.Push(lua.LBool(progress.Status == "completed"))
		return 1
	}))

	// tales.quests.grantQuest(characterID, questID) - Grant quest (bypass source)
	mod.RawSetString("grantQuest", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		questID := L.CheckString(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		_, err := facade.QuestsService().AcceptQuest(characterID, questID)
		L.Push(lua.LBool(err == nil))
		return 1
	}))

	// tales.quests.abandon(characterID, questID) - Abandon a quest
	mod.RawSetString("abandon", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		questID := L.CheckString(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		err := facade.QuestsService().AbandonQuest(characterID, questID)
		L.Push(lua.LBool(err == nil))
		return 1
	}))

	// tales.quests.getQuestLog(characterID) - Get full quest log
	mod.RawSetString("getQuestLog", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		progressList, err := facade.QuestsService().GetQuestLog(characterID)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, progressList))
		return 1
	}))

	L.Push(mod)
	return 1
}
