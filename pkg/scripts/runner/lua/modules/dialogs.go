package modules

import (
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
)

// RegisterDialogsModule registers the tales.dialogs module
func RegisterDialogsModule(L *lua.LState, runner *luarunner.LuaRunner) int {
	mod := L.NewTable()

	// tales.dialogs.get(id) - Get dialog by ID
	mod.RawSetString("get", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		dialog, err := facade.DialogsService().FindByID(id)
		if err != nil || dialog == nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, dialog))
		return 1
	}))

	// tales.dialogs.findByName(name) - Find dialogs by name
	mod.RawSetString("findByName", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		dialogs, err := facade.DialogsService().FindByName(name)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, dialogs))
		return 1
	}))

	// tales.dialogs.getAll() - Get all dialogs
	mod.RawSetString("getAll", L.NewFunction(func(L *lua.LState) int {
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		dialogs, err := facade.DialogsService().FindAll()
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, dialogs))
		return 1
	}))

	// tales.dialogs.getConversation(characterID, targetID) - Get or create conversation
	mod.RawSetString("getConversation", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		targetID := L.CheckString(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		// Find conversations for this character and target
		convs, err := facade.ConversationsService().FindAllForCharacter(characterID)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}

		// Find the one matching the target
		for _, conv := range convs {
			if conv.TargetID == targetID {
				L.Push(luar.New(L, conv))
				return 1
			}
		}

		L.Push(lua.LNil)
		return 1
	}))

	// tales.dialogs.setContext(conversationID, key, value) - Set conversation context
	mod.RawSetString("setContext", L.NewFunction(func(L *lua.LState) int {
		convID := L.CheckString(1)
		key := L.CheckString(2)
		value := L.CheckString(3)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		conv, err := facade.ConversationsService().FindByID(convID)
		if err != nil || conv == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		conv.SetContext(key, value)
		err = facade.ConversationsService().Update(convID, conv)
		L.Push(lua.LBool(err == nil))
		return 1
	}))

	// tales.dialogs.getContext(conversationID, key) - Get conversation context value
	mod.RawSetString("getContext", L.NewFunction(func(L *lua.LState) int {
		convID := L.CheckString(1)
		key := L.CheckString(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		conv, err := facade.ConversationsService().FindByID(convID)
		if err != nil || conv == nil {
			L.Push(lua.LNil)
			return 1
		}

		if conv.Context == nil {
			L.Push(lua.LNil)
			return 1
		}

		if value, ok := conv.Context[key]; ok {
			L.Push(lua.LString(value))
			return 1
		}

		L.Push(lua.LNil)
		return 1
	}))

	// tales.dialogs.hasVisited(conversationID, nodeID) - Check if dialog node was visited
	mod.RawSetString("hasVisited", L.NewFunction(func(L *lua.LState) int {
		convID := L.CheckString(1)
		nodeID := L.CheckString(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		conv, err := facade.ConversationsService().FindByID(convID)
		if err != nil || conv == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		L.Push(lua.LBool(conv.HasVisited(nodeID)))
		return 1
	}))

	// tales.dialogs.getVisitCount(conversationID, nodeID) - Get visit count for dialog node
	mod.RawSetString("getVisitCount", L.NewFunction(func(L *lua.LState) int {
		convID := L.CheckString(1)
		nodeID := L.CheckString(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNumber(0))
			return 1
		}

		conv, err := facade.ConversationsService().FindByID(convID)
		if err != nil || conv == nil {
			L.Push(lua.LNumber(0))
			return 1
		}

		L.Push(lua.LNumber(conv.GetVisitCount(nodeID)))
		return 1
	}))

	L.Push(mod)
	return 1
}
