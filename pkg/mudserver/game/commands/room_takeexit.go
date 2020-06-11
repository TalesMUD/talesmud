package commands

import (
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	m "github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// TakeExit ... executes scream command
func TakeExit(exit string) RoomCommand {

	return func(room *rooms.Room, game def.GameCtrl, message *messages.Message) bool {

		if exit, ok := room.GetExit(exit); ok {

			characterID := message.Character.ID

			// find next room
			if next, err := game.GetFacade().RoomsService().FindByID(exit.Target); err == nil {

				// update old room
				room.RemoveCharacter(characterID)
				game.GetFacade().RoomsService().Update(room.ID, room)

				// remove first to make sure character is not in two rooms at the same time

				// update new room
				next.AddCharacter(characterID)
				game.GetFacade().RoomsService().Update(next.ID, next)

				// update player
				character := message.Character
				character.CurrentRoomID = next.ID
				game.GetFacade().CharactersService().Update(character.ID, character)

				// send all players a left room message
				game.SendMessage() <- messages.CharacterLeftRoom{
					MessageResponse: messages.MessageResponse{
						Audience:   m.MessageAudienceRoomWithoutOrigin,
						AudienceID: room.ID,
						OriginID:   characterID,
						Message:    message.Character.Name + " left.",
					},
				}

				// send player a message to change room
				enterRoom := messages.NewEnterRoomMessage(next, message.FromUser, game)
				enterRoom.AudienceID = message.FromUser.ID
				game.SendMessage() <- enterRoom

				// send all players in new room a joined message
				game.SendMessage() <- messages.CharacterJoinedRoom{
					MessageResponse: messages.MessageResponse{
						Audience:   m.MessageAudienceRoomWithoutOrigin,
						AudienceID: next.ID,
						OriginID:   characterID,
						Message:    message.Character.Name + " entered.",
					},
				}

				return true
			}
		}
		return false
	}
}
