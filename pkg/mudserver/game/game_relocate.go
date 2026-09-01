package game

import (
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	m "github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/mudserver/game/util"
	"github.com/talesmud/talesmud/pkg/worldmap"
)

// RelocateCharacter moves a character to destRoomID, updates room membership,
// and notifies the player and affected rooms. Returns the destination room on success.
func (g *Game) RelocateCharacter(char *characters.Character, userID, destRoomID string) (*rooms.Room, bool) {
	if char == nil || destRoomID == "" || char.CurrentRoomID == destRoomID {
		return nil, false
	}

	dest, err := g.Facade.RoomsService().FindByID(destRoomID)
	if err != nil || dest == nil {
		return nil, false
	}

	oldRoomID := char.CurrentRoomID
	var oldRoom *rooms.Room
	if oldRoomID != "" {
		oldRoom, _ = g.Facade.RoomsService().FindByID(oldRoomID)
		if oldRoom != nil {
			oldRoom.RemoveCharacter(char.ID)
			if uerr := g.Facade.RoomsService().Update(oldRoom.ID, oldRoom); uerr != nil {
				log.WithError(uerr).WithField("roomID", oldRoom.ID).Warn("RelocateCharacter: failed to update old room")
			}
			if inst := g.GetRoomInstances(); inst != nil && inst.IsClone(oldRoomID) {
				inst.NoteLeave(char.ID, oldRoomID, destRoomID)
			}
		}
	}

	dest.AddCharacter(char.ID)
	if uerr := g.Facade.RoomsService().Update(dest.ID, dest); uerr != nil {
		log.WithError(uerr).WithField("roomID", dest.ID).Warn("RelocateCharacter: failed to update destination room")
	}

	if err := g.Facade.CharactersService().Modify(char.ID, func(ch *characters.Character) error {
		ch.CurrentRoomID = destRoomID
		worldmap.MarkOn(ch, dest)
		return nil
	}); err != nil {
		log.WithError(err).WithField("characterID", char.ID).Error("RelocateCharacter: failed to persist current room")
		return nil, false
	}
	char.CurrentRoomID = destRoomID

	user, _ := g.Facade.UsersService().FindByID(userID)
	if user == nil {
		return dest, true
	}

	if oldRoom != nil {
		g.sendMessage <- messages.CharacterLeftRoom{
			MessageResponse: messages.MessageResponse{
				Audience:   m.MessageAudienceRoomWithoutOrigin,
				AudienceID: oldRoom.ID,
				OriginID:   char.ID,
				Message:    char.Name + " left.",
			},
		}
		g.sendMessage <- messages.NewRoomPresenceMessage(oldRoom, g)
	}

	enterRoom := messages.NewEnterRoomMessage(util.RoomWithCharacterReveals(dest, char), user, g, char)
	enterRoom.AudienceID = user.ID
	g.sendMessage <- enterRoom

	g.sendMessage <- messages.CharacterJoinedRoom{
		MessageResponse: messages.MessageResponse{
			Audience:   m.MessageAudienceRoomWithoutOrigin,
			AudienceID: dest.ID,
			OriginID:   char.ID,
			Message:    char.Name + " entered.",
		},
	}
	g.sendMessage <- messages.NewRoomPresenceMessage(dest, g)

	return dest, true
}
