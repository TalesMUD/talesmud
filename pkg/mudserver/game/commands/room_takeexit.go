package commands

import (
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	m "github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/mudserver/game/util"
)

// TakeExit ... executes scream command
func TakeExit(exit string) RoomCommand {

	return func(room *rooms.Room, game def.GameCtrl, message *messages.Message) bool {

		if exit, ok := room.GetExit(exit); ok {

			// Block traversal of hidden exits the character hasn't revealed
			if exit.Hidden && !message.Character.HasRevealedExit(room.ID, exit.Name) {
				game.SendMessage() <- message.Reply("You don't see an exit in that direction.")
				return true
			}

			characterID := message.Character.ID

			// Interrupt rest when moving
			game.InterruptRest(message.Character)

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
				game.SetUserSessionCharacter(message.FromUser, character)

				// send all players a left room message
				game.SendMessage() <- messages.CharacterLeftRoom{
					MessageResponse: messages.MessageResponse{
						Audience:   m.MessageAudienceRoomWithoutOrigin,
						AudienceID: room.ID,
						OriginID:   characterID,
						Message:    message.Character.Name + " left.",
					},
				}
				game.SendMessage() <- messages.NewRoomPresenceMessage(room, game)

				// send player a message to change room
				enterRoom := messages.NewEnterRoomMessage(util.RoomWithCharacterReveals(next, message.Character), message.FromUser, game, message.Character)
				enterRoom.AudienceID = message.FromUser.ID
				game.SendMessage() <- enterRoom

				// Track quest progress for room entry
				NotifyQuestRoomEnter(game, characterID, message.FromUser.ID, next)

				// TODO: exploration XP disabled for now, re-enable when client XP display is updated
				// GrantExplorationXP(game, message.Character, message.FromUser.ID, next)

				// send all players in new room a joined message
				game.SendMessage() <- messages.CharacterJoinedRoom{
					MessageResponse: messages.MessageResponse{
						Audience:   m.MessageAudienceRoomWithoutOrigin,
						AudienceID: next.ID,
						OriginID:   characterID,
						Message:    message.Character.Name + " entered.",
					},
				}
				game.SendMessage() <- messages.NewRoomPresenceMessage(next, game)

				return true
			}
		}
		return false
	}
}
