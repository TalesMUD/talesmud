package def

import (
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/service"
)

// GameCtrl def
// interface for commands package to communicate back to game instance
type GameCtrl interface {
	OnMessageReceived() chan *messages.Message
	SendMessage(msg interface{})
	GetFacade() service.Facade

	//CreateRoomDescription(room *rooms.Room) string
}
