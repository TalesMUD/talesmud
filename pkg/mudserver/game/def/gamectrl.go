package def

import (
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// GameCtrl def
// interface for commands package to communicate back to game instance
type GameCtrl interface {
	OnMessageReceived() chan *messages.Message
	SendMessage(msg interface{})

	//CreateRoomDescription(room *rooms.Room) string
}
