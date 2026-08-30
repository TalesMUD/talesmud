package modules

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/mudserver/game/util"
	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
	"github.com/talesmud/talesmud/pkg/worldmap"
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

	// tales.game.hasItem(characterID, itemID) - Check if character has an item (by ID or TemplateID)
	mod.RawSetString("hasItem", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		itemID := L.CheckString(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		character, err := facade.CharactersService().FindByID(characterID)
		if err != nil || character == nil {
			logrus.WithField("characterID", characterID).WithError(err).Warn("[Script] hasItem: character not found")
			L.Push(lua.LBool(false))
			return 1
		}

		// Check by item ID first
		if character.Inventory.FindItemByID(itemID) != nil {
			L.Push(lua.LBool(true))
			return 1
		}

		// Also check by TemplateID (e.g., "ITM0001" references the template)
		for _, item := range character.Inventory.Items {
			if item.TemplateID == itemID {
				L.Push(lua.LBool(true))
				return 1
			}
		}

		L.Push(lua.LBool(false))
		return 1
	}))

	// tales.game.getFlag(characterID, flagName) - Get a game flag from a character
	mod.RawSetString("getFlag", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		flagName := L.CheckString(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNil)
			return 1
		}

		character, err := facade.CharactersService().FindByID(characterID)
		if err != nil || character == nil {
			logrus.WithField("characterID", characterID).WithError(err).Warn("[Script] getFlag: character not found")
			L.Push(lua.LNil)
			return 1
		}

		if character.Flags == nil {
			L.Push(lua.LNil)
			return 1
		}

		val, ok := character.Flags[flagName]
		if !ok {
			L.Push(lua.LNil)
			return 1
		}

		L.Push(luar.New(L, val))
		return 1
	}))

	// tales.game.setFlag(characterID, flagName, value) - Set a game flag on a character (persisted)
	mod.RawSetString("setFlag", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		flagName := L.CheckString(2)
		value := L.Get(3)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		err := facade.CharactersService().Modify(characterID, func(character *characters.Character) error {
			if character.Flags == nil {
				character.Flags = make(map[string]interface{})
			}
			switch v := value.(type) {
			case lua.LBool:
				character.Flags[flagName] = bool(v)
			case lua.LNumber:
				character.Flags[flagName] = float64(v)
			case lua.LString:
				character.Flags[flagName] = string(v)
			case *lua.LNilType:
				delete(character.Flags, flagName)
			default:
				character.Flags[flagName] = fmt.Sprintf("%v", v)
			}
			return nil
		})
		if err != nil {
			logrus.WithField("characterID", characterID).WithError(err).Warn("[Script] setFlag: failed to persist character")
			L.Push(lua.LBool(false))
			return 1
		}

		L.Push(lua.LBool(true))
		return 1
	}))

	// tales.game.revealExit(roomID, exitName, characterID) - Reveal a hidden exit for a specific character
	// The exit stays globally hidden; only the specified character can see and traverse it.
	mod.RawSetString("revealExit", L.NewFunction(func(L *lua.LState) int {
		roomID := L.CheckString(1)
		exitName := L.CheckString(2)
		characterID := L.CheckString(3)
		facade := runner.GetFacade()
		game := runner.GetGame()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		// Validate the exit exists on the room
		room, err := facade.RoomsService().FindByID(roomID)
		if err != nil || room == nil {
			logrus.WithField("roomID", roomID).WithError(err).Warn("[Script] revealExit: room not found")
			L.Push(lua.LBool(false))
			return 1
		}

		if room.Exits == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		found := false
		for _, exit := range *room.Exits {
			if strings.EqualFold(exit.Name, exitName) {
				found = true
				break
			}
		}

		if !found {
			logrus.WithField("roomID", roomID).WithField("exit", exitName).Warn("[Script] revealExit: exit not found")
			L.Push(lua.LBool(false))
			return 1
		}

		// Track the reveal on the character (not the room)
		character, err := facade.CharactersService().FindByID(characterID)
		if err != nil || character == nil {
			logrus.WithField("characterID", characterID).WithError(err).Warn("[Script] revealExit: character not found")
			L.Push(lua.LBool(false))
			return 1
		}

		// Skip if already revealed (avoid unnecessary DB write and room update)
		if character.HasRevealedExit(roomID, exitName) {
			L.Push(lua.LBool(true))
			return 1
		}

		err = facade.CharactersService().Modify(characterID, func(ch *characters.Character) error {
			ch.RevealExit(roomID, exitName)
			character = ch
			return nil
		})
		if err != nil {
			logrus.WithField("characterID", characterID).WithError(err).Warn("[Script] revealExit: failed to persist character")
			L.Push(lua.LBool(false))
			return 1
		}

		// Send a silent room update so the client learns about the new exit
		// without re-rendering the full room description (which would bury
		// any narrative messages the script already sent).
		if game != nil {
			charUser, err := facade.UsersService().FindByID(character.BelongsUserID)
			if err == nil && charUser != nil {
				roomView := util.RoomWithCharacterReveals(room, character)
				roomUpdate := messages.NewRoomUpdateMessage(roomView, charUser, game, character)
				roomUpdate.Audience = messages.MessageAudienceUser
				roomUpdate.AudienceID = charUser.ID
				game.SendMessage() <- roomUpdate
				if rs, rerr := facade.RoomsService().FindAll(); rerr == nil {
					atlas := worldmap.Reveal(worldmap.Compile(rs), character)
					game.SendMessage() <- messages.NewAtlasMessage(charUser.ID, atlas)
				}
			}
		}

		L.Push(lua.LBool(true))
		return 1
	}))

	// tales.game.hasRevealedExit(characterID, roomID, exitName) - Check if character has revealed a hidden exit
	mod.RawSetString("hasRevealedExit", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		roomID := L.CheckString(2)
		exitName := L.CheckString(3)
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

		L.Push(lua.LBool(character.HasRevealedExit(roomID, exitName)))
		return 1
	}))

	// tales.game.giveItem(characterID, templateID) - Create item from template and add to inventory
	mod.RawSetString("giveItem", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		templateID := L.CheckString(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		item, err := facade.ItemsService().CreateInstanceFromTemplate(templateID)
		if err != nil || item == nil {
			logrus.WithField("templateID", templateID).WithError(err).Warn("[Script] giveItem: failed to create item from template")
			L.Push(lua.LBool(false))
			return 1
		}

		err = facade.CharactersService().Modify(characterID, func(character *characters.Character) error {
			return character.Inventory.AddItem(item)
		})
		if err != nil {
			logrus.WithField("characterID", characterID).WithField("item", templateID).WithError(err).Warn("[Script] giveItem: failed to persist character")
			L.Push(lua.LBool(false))
			return 1
		}

		if game := runner.GetGame(); game != nil {
			if qt := game.GetQuestTracker(); qt != nil {
				userID := ""
				if ch, ferr := facade.CharactersService().FindByID(characterID); ferr == nil && ch != nil {
					userID = ch.BelongsUserID
				}
				qt.OnItemPickup(characterID, userID, item)
			}
		}

		L.Push(lua.LBool(true))
		return 1
	}))

	// tales.game.hasEquipped(characterID, slotName) - Check if character has an item equipped in a slot
	mod.RawSetString("hasEquipped", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		slotName := L.CheckString(2)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		character, err := facade.CharactersService().FindByID(characterID)
		if err != nil || character == nil {
			logrus.WithField("characterID", characterID).WithError(err).Warn("[Script] hasEquipped: character not found")
			L.Push(lua.LBool(false))
			return 1
		}

		if character.EquippedItems == nil {
			L.Push(lua.LBool(false))
			return 1
		}

		slot := items.ItemSlot(slotName)
		item, ok := character.EquippedItems[slot]
		L.Push(lua.LBool(ok && item != nil))
		return 1
	}))

	// tales.game.hasCollectedItem(characterID, templateID) - Check if character collected a CopyOnPickup item
	mod.RawSetString("hasCollectedItem", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		templateID := L.CheckString(2)
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

		L.Push(lua.LBool(character.HasCollectedCopyItem(templateID)))
		return 1
	}))

	// tales.game.resetCollectedItem(characterID, templateID) - Reset a collected CopyOnPickup item flag
	mod.RawSetString("resetCollectedItem", L.NewFunction(func(L *lua.LState) int {
		characterID := L.CheckString(1)
		templateID := L.CheckString(2)
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

		if character.Flags != nil {
			delete(character.Flags, "collected_item:"+templateID)
		}
		facade.CharactersService().Update(characterID, character)

		L.Push(lua.LBool(true))
		return 1
	}))

	L.Push(mod)
	return 1
}
