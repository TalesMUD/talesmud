package commands

import (
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// Command ... commands
type Command interface {
	Execute(game def.GameCtrl, message *messages.Message) bool
}
