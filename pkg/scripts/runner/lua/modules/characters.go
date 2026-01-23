package modules

import (
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
)

// RegisterCharactersModule registers the tales.characters module
func RegisterCharactersModule(L *lua.LState, runner *luarunner.LuaRunner) int {
	mod := L.NewTable()

	// tales.characters.get(id) - Get character by ID
	mod.RawSetString("get", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		character, err := facade.CharactersService().FindByID(id)
		if err != nil || character == nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, character))
		return 1
	}))

	// tales.characters.findByName(name) - Find characters by name
	mod.RawSetString("findByName", L.NewFunction(func(L *lua.LState) int {
		name := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		characters, err := facade.CharactersService().FindByName(name)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, characters))
		return 1
	}))

	// tales.characters.getAll() - Get all characters
	mod.RawSetString("getAll", L.NewFunction(func(L *lua.LState) int {
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		characters, err := facade.CharactersService().FindAll()
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, characters))
		return 1
	}))

	// tales.characters.getRoom(characterID) - Get the room a character is in
	mod.RawSetString("getRoom", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		character, err := facade.CharactersService().FindByID(characterID)
		if err != nil || character == nil {
			L.Push(lua.LNil)
			return 1
		}

		room, err := facade.RoomsService().FindByID(character.CurrentRoomID)
		if err != nil || room == nil {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, room))
		return 1
	}))

	// tales.characters.damage(id, amount) - Damage a character
	mod.RawSetString("damage", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		amount := L.CheckInt(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		character, err := facade.CharactersService().FindByID(id)
		if err != nil || character == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		character.CurrentHitPoints -= int32(amount)
		if character.CurrentHitPoints < 0 {
			character.CurrentHitPoints = 0
		}

		err = facade.CharactersService().Update(id, character)
		L.Push(lua.LBool(err == nil))
		return 1
	}))

	// tales.characters.heal(id, amount) - Heal a character
	mod.RawSetString("heal", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		amount := L.CheckInt(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		character, err := facade.CharactersService().FindByID(id)
		if err != nil || character == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		character.CurrentHitPoints += int32(amount)
		if character.CurrentHitPoints > character.MaxHitPoints {
			character.CurrentHitPoints = character.MaxHitPoints
		}

		err = facade.CharactersService().Update(id, character)
		L.Push(lua.LBool(err == nil))
		return 1
	}))

	// tales.characters.teleport(id, roomID) - Teleport character to room
	mod.RawSetString("teleport", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		roomID := L.CheckString(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		character, err := facade.CharactersService().FindByID(characterID)
		if err != nil || character == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		// Remove from old room
		if character.CurrentRoomID != "" {
			oldRoom, _ := facade.RoomsService().FindByID(character.CurrentRoomID)
			if oldRoom != nil {
				oldRoom.RemoveCharacter(characterID)
				facade.RoomsService().Update(oldRoom.ID, oldRoom)
			}
		}

		// Add to new room
		newRoom, err := facade.RoomsService().FindByID(roomID)
		if err != nil || newRoom == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		newRoom.AddCharacter(characterID)
		facade.RoomsService().Update(newRoom.ID, newRoom)

		// Update character's current room
		character.CurrentRoomID = roomID
		facade.CharactersService().Update(characterID, character)

		L.Push(lua.LBool(true))
		return 1
	}))

	// tales.characters.giveXP(id, amount) - Give XP to character
	mod.RawSetString("giveXP", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		amount := L.CheckInt(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		character, err := facade.CharactersService().FindByID(id)
		if err != nil || character == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		character.XP += int32(amount)
		err = facade.CharactersService().Update(id, character)
		L.Push(lua.LBool(err == nil))
		return 1
	}))

	L.Push(mod)
	return 1
}
