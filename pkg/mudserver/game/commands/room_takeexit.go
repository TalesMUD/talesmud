package commands

import (
	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/instances"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	m "github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/mudserver/game/util"
	"github.com/talesmud/talesmud/pkg/worldmap"
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

			destID := exit.Target
			templateDest, _ := game.GetFacade().RoomsService().FindByID(exit.Target)
			if inst := game.GetRoomInstances(); inst != nil && !inst.IsClone(room.ID) &&
				instances.CrossingIntoInstance(room, templateDest, exit) {
				if cloned, ierr := inst.Enter(characterID, room.ID, exit.Target); ierr == nil && cloned != "" {
					destID = cloned
				} else if ierr != nil {
					log.WithError(ierr).WithField("dest", exit.Target).Warn("TakeExit: instance enter failed")
				}
			}

			// find next room
			next, err := game.GetFacade().RoomsService().FindByID(destID)
			if err == nil && next != nil {

				// update old room
				room.RemoveCharacter(characterID)
				game.GetFacade().RoomsService().Update(room.ID, room)

				// remove first to make sure character is not in two rooms at the same time

				// update new room
				next.AddCharacter(characterID)
				game.GetFacade().RoomsService().Update(next.ID, next)

				// Persist room under the per-character lock so an on-enter
				// setFlag cannot last-write-win and drop CurrentRoomID.
				if err := game.GetFacade().CharactersService().Modify(characterID, func(ch *characters.Character) error {
					ch.CurrentRoomID = next.ID
					worldmap.MarkOn(ch, next)
					return nil
				}); err != nil {
					log.WithError(err).WithField("characterID", characterID).Error("TakeExit: failed to persist current room")
				}
				character := message.Character
				if fresh, ferr := game.GetFacade().CharactersService().FindByID(characterID); ferr == nil && fresh != nil {
					character = fresh
					message.Character = fresh
				} else {
					character.CurrentRoomID = next.ID
				}
				game.SetUserSessionCharacter(message.FromUser, character)
				if inst := game.GetRoomInstances(); inst != nil && inst.IsClone(room.ID) {
					inst.NoteLeave(characterID, room.ID, next.ID)
				}
				PushAtlas(game, message.FromUser.ID, character)

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
			log.WithError(err).WithField("target", exit.Target).Warn("TakeExit: destination room not found")
			game.SendMessage() <- message.Reply("You can't go that way.")
			return true
		}
		return false
	}
}
