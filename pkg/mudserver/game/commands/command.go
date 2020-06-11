package commands

import (
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// Command ... commands
/*
Commands are usually parsed from player input or sent from other actions, commands usually trigger an immediate result to 
the originator, the room or another player. Commands can also invoke system messages that lead to other events or commands
*/
type Command interface {
	Execute(game def.GameCtrl, message *messages.Message) bool
}
