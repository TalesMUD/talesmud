package game

import (
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

func needsUpdate(room *rooms.Room) bool {

	if len(room.Characters) > 0 {
		return true
	}

	return false

}

func (g *Game) updateRoom(room *rooms.Room) {

	g.removeOfflineCharacters(room)

	// execute scripts in the room
}

func (g *Game) removeOfflineCharacters(room *rooms.Room) {

	needsUpdate := false

	for _, char := range room.Characters {
		remove := false

		if character, err := g.Facade.CharactersService().FindByID(char); err == nil {
			// check if character is still in this room
			if character.CurrentRoomID != room.ID {
				remove = true
			} else {
				if user, err := g.Facade.UsersService().FindByID(character.BelongsUserID); err == nil {
					//check if user is logged in with another character
					if user.LastCharacter != char {
						remove = true
					} else if user.LastCharacter == char && !user.IsOnline {
						// check if player is offline
						remove = true
					}
				}
			}
		} else {
			// error finding character? remove it from the room
			remove = true
		}

		if remove {
			room.RemoveCharacter(char)
			needsUpdate = true
			logrus.New().WithField("character", char).Info("Removed character from room")

		}

	}

	if needsUpdate {
		logrus.Info("Updating room because character was removed")
		g.Facade.RoomsService().Update(room.ID, room)
	}
}

func (g *Game) handleRoomUpdates() {

	if allRooms, err := g.Facade.RoomsService().FindAll(); err == nil {

		for _, room := range allRooms {
			if needsUpdate(room) {
				go g.updateRoom(room)
			}
		}
	} else {
		log.WithError(err).Error("Could nto update rooms")
	}

}
