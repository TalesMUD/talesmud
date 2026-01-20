package modules

import (
	"github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"

	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
)

// RegisterGameModule registers the tales.game module
func RegisterGameModule(L *lua.LState, runner *luarunner.LuaRunner) int {
	mod := L.NewTable()

	// tales.game.msgToRoom(roomID, message) - Send message to all players in room
	mod.RawSetString("msgToRoom", L.NewFunction(func(L *lua.LState) int {
		roomID := L.CheckString(1)
		message := L.CheckString(2)
		game := runner.GetGame()
		facade := runner.GetFacade()
		if game == nil || facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		// Validate room exists to avoid "success but nothing happened" confusion.
		if _, err := facade.RoomsService().FindByID(roomID); err != nil {
			logrus.WithField("roomID", roomID).WithError(err).Warn("[Script] msgToRoom: room not found")
			L.Push(lua.LBool(false))
			return 1
		}

		msg := messages.NewRoomBasedMessage("SYSTEM", message)
		msg.Audience = messages.MessageAudienceRoom
		msg.AudienceID = roomID
		game.SendMessage() <- msg

		L.Push(lua.LBool(true))
		return 1
	}))

	// tales.game.msgToCharacter(characterID, message) - Send message to specific character
	mod.RawSetString("msgToCharacter", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		message := L.CheckString(2)
		game := runner.GetGame()
		facade := runner.GetFacade()
		if game == nil || facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		// Find the character to get their user ID
		character, err := facade.CharactersService().FindByID(characterID)
		if err != nil || character == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		msg := messages.NewRoomBasedMessage("SYSTEM", message)
		msg.Audience = messages.MessageAudienceUser
		msg.AudienceID = character.BelongsUserID
		game.SendMessage() <- msg

		L.Push(lua.LBool(true))
		return 1
	}))

	// tales.game.msgToUser(userID, message) - Send message to specific user
	mod.RawSetString("msgToUser", L.NewFunction(func(L *lua.LState) int {
		userID := L.CheckString(1)
		message := L.CheckString(2)
		game := runner.GetGame()
		if game == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		msg := messages.NewRoomBasedMessage("SYSTEM", message)
		msg.Audience = messages.MessageAudienceUser
		msg.AudienceID = userID
		game.SendMessage() <- msg

		L.Push(lua.LBool(true))
		return 1
	}))

	// tales.game.broadcast(message) - Send message to all connected players
	mod.RawSetString("broadcast", L.NewFunction(func(L *lua.LState) int {
		message := L.CheckString(1)
		game := runner.GetGame()
		if game == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		msg := messages.NewRoomBasedMessage("SYSTEM", message)
		msg.Audience = messages.MessageAudienceGlobal
		game.SendMessage() <- msg

		L.Push(lua.LBool(true))
		return 1
	}))

	// tales.game.msgToRoomExcept(roomID, message, excludeCharacterID) - Send to room except one player
	mod.RawSetString("msgToRoomExcept", L.NewFunction(func(L *lua.LState) int {
		roomID := L.CheckString(1)
		message := L.CheckString(2)
		excludeID := L.CheckString(3)
		game := runner.GetGame()
		facade := runner.GetFacade()
		if game == nil || facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		// Validate room exists
		if _, err := facade.RoomsService().FindByID(roomID); err != nil {
			logrus.WithField("roomID", roomID).WithError(err).Warn("[Script] msgToRoomExcept: room not found")
			L.Push(lua.LBool(false))
			return 1
		}

		// Get the user ID to exclude
		character, err := facade.CharactersService().FindByID(excludeID)
		if err != nil || character == nil {
			// If we can't find the character, just send to everyone
			msg := messages.NewRoomBasedMessage("SYSTEM", message)
			msg.Audience = messages.MessageAudienceRoom
			msg.AudienceID = roomID
			game.SendMessage() <- msg
			L.Push(lua.LBool(true))
			return 1
		}

		msg := messages.NewRoomBasedMessage("SYSTEM", message)
		msg.Audience = messages.MessageAudienceRoomWithoutOrigin
		msg.AudienceID = roomID
		msg.OriginID = character.BelongsUserID
		game.SendMessage() <- msg

		L.Push(lua.LBool(true))
		return 1
	}))

	// tales.game.log(level, message) - Log a message
	mod.RawSetString("log", L.NewFunction(func(L *lua.LState) int {
		level := L.CheckString(1)
		message := L.CheckString(2)

		switch level {
		case "debug":
			logrus.Debug("[Script] " + message)
		case "info":
			logrus.Info("[Script] " + message)
		case "warn":
			logrus.Warn("[Script] " + message)
		case "error":
			logrus.Error("[Script] " + message)
		default:
			logrus.Info("[Script] " + message)
		}

		return 0
	}))

	L.Push(mod)
	return 1
}
