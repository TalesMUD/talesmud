package commands

import (
	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	m "github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/mudserver/game/util"
)

// Display ... executes scream command
func Display(room *rooms.Room, game def.GameCtrl, message *m.Message) bool {

	charID := ""
	if message.Character != nil {
		charID = message.Character.ID
	}
	log.WithFields(log.Fields{
		"userId":      message.FromUser.ID,
		"nickname":    message.FromUser.Nickname,
		"characterId": charID,
		"roomId":      room.ID,
	}).Info("enterRoom")

	enterRoom := m.NewEnterRoomMessage(util.RoomWithCharacterReveals(room, message.Character), message.FromUser, game, message.Character)
	enterRoom.AudienceID = message.FromUser.ID
	game.SendMessage() <- enterRoom
	return true
}
