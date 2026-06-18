package game

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// handleDefaultMessage implements implicit "say" behavior.
// When player input doesn't match any command, treat it as room speech.
// This allows players to type "hello everyone" instead of "say hello everyone".
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

	game.DisconnectUserSession(user.ID)

	if user.LastCharacter == "" {
		return
	}
	character, err := game.Facade.CharactersService().FindByID(user.LastCharacter)
	if err != nil || character == nil {
		log.WithField("characterID", user.LastCharacter).WithError(err).Warn("Could not load character during user quit")
		return
	}
	room, err := game.Facade.RoomsService().FindByID(character.CurrentRoomID)
	if err != nil || room == nil {
		log.WithField("roomID", character.CurrentRoomID).WithError(err).Warn("Could not load room during user quit")
		return
	}

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

// Find the matching character for the user where the message originated
func (game *Game) attachCharacterToMessage(msg *messages.Message) {

	if msg.Character != nil {
		return
	}

	// could be a processed message that got the user removed
	if msg.FromUser == nil || msg.FromUser.LastCharacter == "" {
		return
	}

	if character, err := game.Facade.CharactersService().FindByID(msg.FromUser.LastCharacter); err == nil {
		character.NormalizeAttributeShorts()
		msg.Character = character
	} else {
		log.Error("Couldt not load character for user")
	}
}

func (game *Game) handleUserJoined(user *entities.User) {
	game.ConnectUserSession(user)

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
