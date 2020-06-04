package game

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	m "github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/mudserver/game/util"
)

func (game *Game) handleDefaultMessage(message *messages.Message) {

	user := ""

	if message.FromUser != nil {
		user = message.FromUser.Nickname
		if message.Character != nil {
			user = message.Character.Name
		}
	}

	out := messages.NewOutgoingMessage(user, message.Data)
	out.AudienceID = message.Character.CurrentRoomID

	game.SendMessage(out)
}

func (game *Game) handleUserQuit(user *entities.User) {

	character, _ := game.Facade.CharactersService().FindByID(user.LastCharacter)
	room, _ := game.Facade.RoomsService().FindByID(character.CurrentRoomID)

	room.Characters = util.RemoveStringFromSlice(room.Characters, character.ID.Hex())
	//TOOD: move update to queue
	game.Facade.RoomsService().Update(room.ID.Hex(), room)

	game.SendMessage(messages.CharacterLeftRoom{
		MessageResponse: messages.MessageResponse{
			Audience:   messages.MessageAudienceGlobal,
			AudienceID: character.CurrentRoomID,
			Message:    character.Name + " left.",
		},
	})
}

func (game *Game) attachCharacterToMessage(msg *messages.Message) {

	if msg.Character != nil {
		return
	}

	// could be a processed message that got the user removed
	if msg.FromUser == nil || msg.FromUser.LastCharacter == "" {
		return
	}

	if character, err := game.Facade.CharactersService().FindByID(msg.FromUser.LastCharacter); err == nil {
		msg.Character = character
	} else {
		log.Error("Couldt not load character for user")
	}
}

func (game *Game) handleUserJoined(user *entities.User) {

	// get active character for user

	if user.LastCharacter == "" {

		if chars, err := game.Facade.CharactersService().FindAllForUser(user.ID.Hex()); err == nil {

			// take first character for now
			// TODO: let the player choose?
			if len(chars) > 0 {
				user.LastCharacter = chars[0].ID.Hex()
				user.LastSeen = time.Now()
				//TODO: send updates via message queue?
				game.Facade.UsersService().Update(user.RefID, user)
			}
		}
	}

	character, _ := game.Facade.CharactersService().FindByID(user.LastCharacter)

	characterSelected := &messages.CharacterSelected{
		MessageResponse: messages.MessageResponse{
			Audience:   messages.MessageAudienceOrigin,
			AudienceID: user.ID.Hex(),
			Type:       "characterSelected",
			Message:    fmt.Sprintf("You are now playing as [%v]", character.Name),
		},
		Character: character,
	}

	game.SendMessage(characterSelected)

	if character, err := game.Facade.CharactersService().FindByID(user.LastCharacter); err == nil {

		var currentRoom *rooms.Room

		// new character or not part of a room?
		if character.CurrentRoomID == "" {
			// find a random room to start in or get starting room
			rooms, _ := game.Facade.RoomsService().FindAll()

			if len(rooms) > 0 {
				// TOOD make this random or select a starting room
				currentRoom = rooms[0]
			}
		} else {
			if currentRoom, err = game.Facade.RoomsService().FindByID(character.CurrentRoomID); err != nil {
				log.WithField("room", character.CurrentRoomID).Warn("CurrentRoomID for player not found")
			}
		}

		// update room // send these state change messages via channel
		currentRoom.Characters = append(currentRoom.Characters, character.ID.Hex())
		game.Facade.RoomsService().Update(currentRoom.ID.Hex(), currentRoom)

		enterRoom := m.NewEnterRoomMessage(currentRoom)
		enterRoom.AudienceID = user.ID.Hex()
		game.SendMessage(enterRoom)

		game.SendMessage(messages.CharacterJoinedRoom{
			MessageResponse: messages.MessageResponse{
				Audience:   m.MessageAudienceRoom,
				AudienceID: currentRoom.ID.Hex(),
				Message:    character.Name + " entered.",
			},
		})
	}
}
