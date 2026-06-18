package commands

import (
	"strings"

	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// WhoCommand ... foo
type WhoCommand struct {
}

// Key ...
func (command *WhoCommand) Key() CommandKey { return &ExactCommandKey{} }

// Execute ... executes who command
func (command *WhoCommand) Execute(game def.GameCtrl, message *messages.Message) bool {

	names := []string{}
	for _, player := range game.GetOnlinePlayers() {
		names = append(names, player.CharacterName)
	}

	result := "List of all online players:\n"
	if len(names) > 0 {
		result += strings.Join(names, ", ")
	}

	game.SendMessage() <- messages.Reply(message.FromUser.ID, result)
	return true
}
