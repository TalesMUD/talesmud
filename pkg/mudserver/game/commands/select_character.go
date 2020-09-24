package commands

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	m "github.com/talesmud/talesmud/pkg/mudserver/game/messages"
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
			}
		}
	}

	// update player
	user.LastCharacter = character.ID
	game.GetFacade().UsersService().Update(user.RefID, user)

	characterSelected := &messages.CharacterSelected{
		MessageResponse: messages.MessageResponse{
			Audience:   messages.MessageAudienceOrigin,
			AudienceID: user.ID,
			Type:       messages.MessageTypeCharacterSelected,
			Message:    fmt.Sprintf("You are now playing as [%v]", character.Name),
		},
		Character: character,
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
		// find a random room to start in or get starting room
		rooms, _ := game.GetFacade().RoomsService().FindAll()

		if len(rooms) > 0 {
			// TOOD make this random or select a starting room
			currentRoom = rooms[0]

			//TODO: send this as message
			character.CurrentRoomID = currentRoom.ID
			game.GetFacade().CharactersService().Update(character.ID, character)

		}
	}

	// update room // send these state change messages via channel
	currentRoom.AddCharacter(character.ID)
	game.GetFacade().RoomsService().Update(currentRoom.ID, currentRoom)

	enterRoom := m.NewEnterRoomMessage(currentRoom, user, game)
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

}
