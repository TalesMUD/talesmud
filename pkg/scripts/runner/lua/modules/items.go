package modules

import (
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
)

// RegisterItemsModule registers the tales.items module
func RegisterItemsModule(L *lua.LState, runner *luarunner.LuaRunner) int {
	mod := L.NewTable()

	// tales.items.get(id) - Get item by ID
	mod.RawSetString("get", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		item, err := facade.ItemsService().Items().FindByID(id)
		if err != nil || item == nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, item))
		return 1
	}))

	// tales.items.findByName(name) - Find items by name
	mod.RawSetString("findByName", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		items, err := facade.ItemsService().Items().FindByName(name)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, items))
		return 1
	}))

	// tales.items.getTemplate(id) - Get item template by ID
	mod.RawSetString("getTemplate", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		template, err := facade.ItemsService().ItemTemplates().FindByID(id)
		if err != nil || template == nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, template))
		return 1
	}))

	// tales.items.findTemplates(name) - Find item templates by name
	mod.RawSetString("findTemplates", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		templates, err := facade.ItemsService().ItemTemplates().FindByName(name)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, templates))
		return 1
	}))

	// tales.items.createFromTemplate(templateID) - Create item from template
	mod.RawSetString("createFromTemplate", L.NewFunction(func(L *lua.LState) int {
		templateID := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		item, err := facade.ItemsService().CreateItemFromTemplate(templateID)
		if err != nil || item == nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, item))
		return 1
	}))

	// tales.items.store(item) - Store/save an item
	mod.RawSetString("store", L.NewFunction(func(L *lua.LState) int {
		// Get the item table from Lua and convert back
		// For now, just return false as this requires more complex conversion
		L.Push(lua.LBool(false))
		return 1
	}))

	// tales.items.delete(id) - Delete an item
	mod.RawSetString("delete", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		err := facade.ItemsService().Items().Delete(id)
		L.Push(lua.LBool(err == nil))
		return 1
	}))

	L.Push(mod)
	return 1
}
