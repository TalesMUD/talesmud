package game

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func (game *Game) handleDefaultMessage(message *messages.Message) {

	user := ""

	if message.FromUser != nil {
		user = message.FromUser.Nickname
		if message.Character != nil {
			user = message.Character.Name
		}
	}

	out := messages.NewRoomBasedMessage(user, message.Data)

	if message.Character != nil {
		out.AudienceID = message.Character.CurrentRoomID
	}

	game.SendMessage() <- out
}

func (game *Game) handleUserQuit(user *entities.User) {

	log.Info("Handle User Quit " + user.Nickname)

	// set user offline
	user.IsOnline = false
	game.Facade.UsersService().Update(user.RefID, user)

	character, _ := game.Facade.CharactersService().FindByID(user.LastCharacter)
	room, _ := game.Facade.RoomsService().FindByID(character.CurrentRoomID)

	//TOOD: move update to queue
	room.RemoveCharacter(character.ID)
	game.Facade.RoomsService().Update(room.ID, room)

	game.SendMessage() <- messages.CharacterLeftRoom{
		MessageResponse: messages.MessageResponse{
			Audience:   messages.MessageAudienceRoomWithoutOrigin,
			OriginID:   character.ID,
			AudienceID: character.CurrentRoomID,
			Message:    character.Name + " left.",
		},
	}
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

		if chars, err := game.Facade.CharactersService().FindAllForUser(user.ID); err == nil {

			// take first character for now
			// TODO: let the player choose?
			if len(chars) > 0 {
				user.LastCharacter = chars[0].ID
				user.LastSeen = time.Now()
				user.IsOnline = true
				//TODO: send updates via message queue?
				game.Facade.UsersService().Update(user.RefID, user)
			}
		} else {
			// player has no character yet, respnd with createCharacter Message
			game.SendMessage() <- messages.NewCreateCharacterMessage(user.ID)
			return
		}
	}

	if character, err := game.Facade.CharactersService().FindByID(user.LastCharacter); err != nil {

		log.WithField("user", user.Name).Error("Could not select character for user")
		// player character may be broken, let the user create a new one
		//game.SendMessage(messages.NewCreateCharacterMessage(user.ID))
		// send list characters command
		game.onMessageReceived <- messages.NewMessage(user, "lc")
	} else {

		// send message as userwould do it
		selectCharacterMsg := messages.NewMessage(user, "selectcharacter "+character.Name)
		game.OnMessageReceived() <- selectCharacterMsg
	}
}
