package mudserver

import "github.com/talesmud/talesmud/pkg/entities"

// SessionHook owns a connected player's input when a presentation mode
// replaces the classic room command loop. A nil hook keeps classic play.
type SessionHook interface {
	Active() bool
	OnConnect(user *entities.User, send func(any))
	OnInput(user *entities.User, text string, send func(any)) bool
	OnDisconnect(user *entities.User)
	// OnNotice records one player-visible line and redraws. Callers may
	// invoke it off the message-drain goroutine.
	OnNotice(user *entities.User, text string, send func(any))
}
