package commands

import (
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	m "github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// Display ... executes scream command
func Display(room *rooms.Room, game def.GameCtrl, message *m.Message) bool {

	enterRoom := m.NewEnterRoomMessage(room, message.FromUser, game)
	enterRoom.AudienceID = message.FromUser.ID
	game.SendMessage() <- enterRoom
	return true
}
