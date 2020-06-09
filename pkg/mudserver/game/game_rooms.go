package game

import (
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

	//log.Info("Updating room " + room.Name)

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
