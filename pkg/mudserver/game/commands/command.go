package commands

import (
	"github.com/atla/owndnd/pkg/mudserver/game/def"
	"github.com/atla/owndnd/pkg/mudserver/game/messages"
)

// Command ... commands
type Command interface {
	Execute(game def.GameCtrl, message *messages.Message) bool
}
