package def

import (
	"github.com/talesmud/talesmud/pkg/service"
)

// GameCtrl def
// interface for commands package to communicate back to game instance
type GameCtrl interface {

	// Used to pass messages as events inside the mud server, e.g. translate a command into other user messages etc.
	OnMessageReceived() chan interface{}
	// used to send replies/messages to users, origin or rooms, or global
	SendMessage() chan interface{}
	GetFacade() service.Facade
}
