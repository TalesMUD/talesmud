package modules

import (
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/talesmud/talesmud/pkg/repository"
	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
)

// RegisterRoomsModule registers the tales.rooms module
func RegisterRoomsModule(L *lua.LState, runner *luarunner.LuaRunner) int {
	mod := L.NewTable()

	// tales.rooms.get(id) - Get room by ID
	mod.RawSetString("get", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		room, err := facade.RoomsService().FindByID(id)
		if err != nil || room == nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, room))
		return 1
	}))

	// tales.rooms.findByName(name) - Find rooms by name
	mod.RawSetString("findByName", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		rooms, err := facade.RoomsService().FindAllWithQuery(repository.RoomsQuery{Name: name})
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, rooms))
		return 1
	}))

	// tales.rooms.findByArea(area) - Find rooms by area
	mod.RawSetString("findByArea", L.NewFunction(func(L *lua.LState) int {
		area := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		rooms, err := facade.RoomsService().FindAllWithQuery(repository.RoomsQuery{Area: area})
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, rooms))
		return 1
	}))

	// tales.rooms.getAll() - Get all rooms
	mod.RawSetString("getAll", L.NewFunction(func(L *lua.LState) int {
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		rooms, err := facade.RoomsService().FindAll()
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, rooms))
		return 1
	}))

	// tales.rooms.getCharacters(roomID) - Get characters in a room
	mod.RawSetString("getCharacters", L.NewFunction(func(L *lua.LState) int {
		roomID := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		room, err := facade.RoomsService().FindByID(roomID)
		if err != nil || room == nil || room.Characters == nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, *room.Characters))
		return 1
	}))

	// tales.rooms.getNPCs(roomID) - Get NPCs in a room
	mod.RawSetString("getNPCs", L.NewFunction(func(L *lua.LState) int {
		roomID := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		room, err := facade.RoomsService().FindByID(roomID)
		if err != nil || room == nil || room.NPCs == nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, *room.NPCs))
		return 1
	}))

	// tales.rooms.getItems(roomID) - Get items in a room
	mod.RawSetString("getItems", L.NewFunction(func(L *lua.LState) int {
		roomID := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		room, err := facade.RoomsService().FindByID(roomID)
		if err != nil || room == nil || room.Items == nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, *room.Items))
		return 1
	}))

	L.Push(mod)
	return 1
}
