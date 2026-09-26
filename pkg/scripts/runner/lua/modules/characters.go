package modules

import (
	"errors"
	"strings"

	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/ruleset"
	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
)

var errNotEnoughGold = errors.New("not enough gold")

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

	// tales.characters.addGold(id, delta) - signed gold change. A debit below zero is refused.
	mod.RawSetString("addGold", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		delta := int64(L.CheckNumber(2))
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}
		err := facade.CharactersService().Modify(id, func(character *characters.Character) error {
			if character.Gold+delta < 0 {
				return errNotEnoughGold
			}
			character.Gold += delta
			return nil
		})
		if err != nil {
			if !errors.Is(err, errNotEnoughGold) {
				log.WithError(err).WithField("characterID", id).Warn("addGold failed")
			}
			L.Push(lua.LBool(false))
			return 1
		}
		pushGoldUpdate(runner, id)
		L.Push(lua.LBool(true))
		return 1
	}))

	// tales.characters.setBind(id, roomID) - set or clear the respawn room.
	mod.RawSetString("setBind", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		roomID := strings.TrimSpace(L.CheckString(2))
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}
		if roomID != "" {
			room, err := facade.RoomsService().FindByID(roomID)
			if err != nil || room == nil {
				L.Push(lua.LBool(false))
				return 1
			}
		}
		err := facade.CharactersService().Modify(id, func(character *characters.Character) error {
			character.BoundRoomID = roomID
			return nil
		})
		L.Push(lua.LBool(err == nil))
		return 1
	}))

	// tales.characters.applyLevels(id) - apply levels the current XP can buy.
	mod.RawSetString("applyLevels", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LNumber(0))
			return 1
		}
		var gained int
		var userID, msg string
		err := facade.CharactersService().Modify(id, func(character *characters.Character) error {
			userID = character.BelongsUserID
			result := leveling.ApplyPendingLevels(character)
			if result == nil {
				return nil
			}
			gained = result.LevelsGained
			msg = result.Message
			return nil
		})
		if err != nil {
			log.WithError(err).WithField("characterID", id).Warn("applyLevels failed")
			L.Push(lua.LNumber(0))
			return 1
		}
		if gained > 0 {
			if game := runner.GetGame(); game != nil && msg != "" {
				game.SendMessage() <- messages.MessageResponse{
					Audience:   messages.MessageAudienceUser,
					AudienceID: userID,
					Type:       messages.MessageTypeLevelUp,
					Message:    msg,
				}
			}
			pushGoldUpdate(runner, id)
		}
		L.Push(lua.LNumber(gained))
		return 1
	}))

	// tales.characters.setProgress(id, level, xp [, maxHP])
	// Sets level and XP without touching class, skills, inventory, gold, or flags.
	// Level is clamped to 1..the effective cap. XP below zero becomes zero.
	// A positive maxHP replaces max and current hit points.
	mod.RawSetString("setProgress", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckString(1)
		level := int32(L.CheckInt(2))
		xp := int32(L.CheckInt(3))
		var maxHP int32
		if L.GetTop() >= 4 && L.Get(4).Type() != lua.LTNil {
			maxHP = int32(L.CheckInt(4))
		}
		facade := runner.GetFacade()
		if facade == nil {
			L.Push(lua.LBool(false))
			return 1
		}
		err := facade.CharactersService().Modify(id, func(character *characters.Character) error {
			capLevel := character.GetEffectiveMaxLevel(ruleset.LevelCap())
			if level < 1 {
				level = 1
			}
			if capLevel > 0 && level > capLevel {
				level = capLevel
			}
			if xp < 0 {
				xp = 0
			}
			character.Level = level
			character.XP = xp
			if maxHP > 0 {
				character.MaxHitPoints = maxHP
				character.CurrentHitPoints = maxHP
			}
			return nil
		})
		if err != nil {
			log.WithError(err).WithField("characterID", id).Warn("setProgress failed")
			L.Push(lua.LBool(false))
			return 1
		}
		pushGoldUpdate(runner, id)
		L.Push(lua.LBool(true))
		return 1
	}))

	L.Push(mod)
	return 1
}

func pushGoldUpdate(runner *luarunner.LuaRunner, characterID string) {
	game := runner.GetGame()
	facade := runner.GetFacade()
	if game == nil || facade == nil {
		return
	}
	character, err := facade.CharactersService().FindByID(characterID)
	if err != nil || character == nil {
		return
	}
	if update := messages.NewCharacterUpdateMessage(character.BelongsUserID, character); update != nil {
		game.SendMessage() <- update
	}
	game.SendMessage() <- messages.InventoryUpdateMessage{
		MessageResponse: messages.MessageResponse{
			Audience:   messages.MessageAudienceUser,
			AudienceID: character.BelongsUserID,
			Type:       messages.MessageTypeInventoryUpdate,
		},
		Inventory:     character.Inventory,
		EquippedItems: character.EquippedItems,
		Gold:          character.Gold,
	}
}
