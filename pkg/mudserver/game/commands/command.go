package commands

import (
	"strings"

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
	Key() CommandKey
}

// CommandKey is used to differentiate between commands that have parameters or comands (with single or multiple words) without parameters
type CommandKey interface {
	Matches(key string, cmd string) bool //parts []string
}

// ExactCommandKey ...
type ExactCommandKey struct {
}

// Matches ...
func (k ExactCommandKey) Matches(key string, cmd string) bool {
	return cmd == key
}

// StartsWithCommandKey ...
type StartsWithCommandKey struct {
}

// Matches ...
func (k StartsWithCommandKey) Matches(key string, cmd string) bool {
	return strings.HasPrefix(cmd, key)
}
