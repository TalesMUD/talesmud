package commands

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	m "github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/mudserver/game/util"
	"github.com/talesmud/talesmud/pkg/service"
	"github.com/talesmud/talesmud/pkg/worldmap"
)

// SelectCharacterCommand ... select a character
type SelectCharacterCommand struct {
}

// Key ...
func (command *SelectCharacterCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute ... executes scream command
func (command *SelectCharacterCommand) Execute(game def.GameCtrl, message *messages.Message) bool {

	parts := strings.Fields(message.Data)
	characterName := strings.Join(parts[1:], " ")

	if characters, err := game.GetFacade().CharactersService().FindByName(characterName); err == nil {

		for _, character := range characters {
			if character.Name == characterName && character.BelongsUserID == message.FromUser.ID {
				// found character to select
				handleCharacterSelected(game, message.FromUser, character)
				return true
			}
		}
	}
	game.SendMessage() <- message.Reply("Could not select character: " + characterName)
	return true
}

func handleCharacterSelected(game def.GameCtrl, user *entities.User, character *characters.Character) {

	// Normalize attribute short names to uppercase (migration for pre-fix characters)
	character.NormalizeAttributeShorts()

	// Ensure mana is initialized for caster classes (migration for pre-mana characters)
	expectedMaxMana := character.CalculateMaxMana()
	if expectedMaxMana > 0 && character.MaxMana == 0 {
		character.MaxMana = expectedMaxMana
		character.CurrentMana = expectedMaxMana
		game.GetFacade().CharactersService().Update(character.ID, character)
	}

	// Retroactive migration: grant distributable attribute points for existing leveled characters
	if character.Level > 1 && character.UnspentAttributePoints == 0 && len(character.SpentAttributePoints) == 0 {
		retroactivePoints := (character.Level - 1) * int32(leveling.PointsPerLevel)
		character.UnspentAttributePoints = retroactivePoints
		game.GetFacade().CharactersService().Update(character.ID, character)
	}

	// handle Character deselection
	if user.LastCharacter != "" && user.LastCharacter != character.ID {
		if character, err := game.GetFacade().CharactersService().FindByID(user.LastCharacter); err == nil {
			if room, err := game.GetFacade().RoomsService().FindByID(character.CurrentRoomID); err == nil {

				// remove character from current room
				// send all players a left room message
				game.SendMessage() <- messages.CharacterLeftRoom{
					MessageResponse: messages.MessageResponse{
						Audience:   m.MessageAudienceRoomWithoutOrigin,
						AudienceID: room.ID,
						OriginID:   character.ID,
						Message:    character.Name + " left.",
					},
				}

				room.RemoveCharacter(character.ID)
				game.GetFacade().RoomsService().Update(room.ID, room)
				game.SendMessage() <- m.NewRoomPresenceMessage(room, game)
			}
		}
	}

	// update player
	game.SetUserSessionCharacter(user, character)

	characterSelected := &messages.CharacterSelected{
		MessageResponse: messages.MessageResponse{
			Audience:   messages.MessageAudienceOrigin,
			AudienceID: user.ID,
			Type:       messages.MessageTypeCharacterSelected,
			Message:    fmt.Sprintf("You are now playing as [%v]", character.Name),
		},
		Character:      character,
		XPForNextLevel: leveling.GetXPRequired(character.Level + 1),
	}

	game.SendMessage() <- characterSelected

	var currentRoom *rooms.Room
	var err error

	if character.CurrentRoomID != "" {
		if currentRoom, err = game.GetFacade().RoomsService().FindByID(character.CurrentRoomID); err != nil {
			log.WithField("room", character.CurrentRoomID).Warn("CurrentRoomID for player not found (room might have been deleted or temporary)")
			// set to ""
			character.CurrentRoomID = ""
		}
	}

	// new character or not part of a room?
	if character.CurrentRoomID == "" {
		facade := game.GetFacade()
		currentRoom = service.ResolveStartRoom(facade.ServerSettingsService(), facade.RoomsService())
		if currentRoom != nil {
			character.CurrentRoomID = currentRoom.ID
			if character.BoundRoomID == "" {
				character.BoundRoomID = currentRoom.ID
			}
			game.GetFacade().CharactersService().Update(character.ID, character)
		}
	}

	if currentRoom == nil {
		log.WithField("character", character.Name).Error("No start room available for character")
		game.SendMessage() <- messages.Reply(user.ID, "The world has no starting room. Ask a creator to set startRoomID.")
		return
	}

	if err := game.GetFacade().CharactersService().Modify(character.ID, func(ch *characters.Character) error {
		ch.CurrentRoomID = currentRoom.ID
		if ch.BoundRoomID == "" {
			ch.BoundRoomID = currentRoom.ID
		}
		worldmap.MarkOn(ch, currentRoom)
		return nil
	}); err != nil {
		log.WithError(err).WithField("characterID", character.ID).Warn("select character: failed to persist discovery")
	} else if fresh, ferr := game.GetFacade().CharactersService().FindByID(character.ID); ferr == nil && fresh != nil {
		character = fresh
		game.SetUserSessionCharacter(user, character)
	}
	PushAtlas(game, user.ID, character)

	// update room // send these state change messages via channel
	currentRoom.AddCharacter(character.ID)
	game.GetFacade().RoomsService().Update(currentRoom.ID, currentRoom)

	enterRoom := m.NewEnterRoomMessage(util.RoomWithCharacterReveals(currentRoom, character), user, game, character)
	enterRoom.AudienceID = user.ID
	game.SendMessage() <- enterRoom

	game.SendMessage() <- messages.CharacterJoinedRoom{
		MessageResponse: messages.MessageResponse{
			Audience:   m.MessageAudienceRoomWithoutOrigin,
			AudienceID: currentRoom.ID,
			OriginID:   character.ID,
			Message:    character.Name + " entered.",
		},
	}
	game.SendMessage() <- m.NewRoomPresenceMessage(currentRoom, game)

	// Send initial character stats (HP, mana, XP, etc.) to the client
	game.SendMessage() <- m.NewCharacterUpdateMessage(user.ID, character)

	// Send initial inventory state to the client
	game.SendMessage() <- &m.InventoryUpdateMessage{
		MessageResponse: m.MessageResponse{
			Audience:   m.MessageAudienceOrigin,
			AudienceID: user.ID,
			Type:       m.MessageTypeInventoryUpdate,
		},
		Inventory:     character.Inventory,
		EquippedItems: character.EquippedItems,
		Gold:          character.Gold,
	}

	game.GetFacade().QuestsService().GrantAutoQuests(character.ID, currentRoom.Area)

	// Send initial quest log to the client
	sendQuestLogToPlayer(game, user.ID, character.ID)

}

// sendQuestLogToPlayer sends the full quest log with enriched quest details
func sendQuestLogToPlayer(game def.GameCtrl, userID, characterID string) {
	progressList, err := game.GetFacade().QuestsService().GetQuestLog(characterID)
	if err != nil {
		return
	}

	entries := buildQuestLogEntries(game, progressList)

	// Send quest log message
	game.SendMessage() <- m.NewQuestLogMessage(userID, entries)
}
