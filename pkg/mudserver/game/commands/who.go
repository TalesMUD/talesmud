package commands

import (
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

	result := "List of all online players:\n"

	if users, err := game.GetFacade().UsersService().FindAllOnline(); err == nil {
		for i, user := range users {
			if i > 0 {
				result += ", "
			}
			if character, err := game.GetFacade().CharactersService().FindByID(user.LastCharacter); err == nil {
				result += character.Name
			}
		}
	}

	game.SendMessage() <- messages.Reply(message.FromUser.ID, result)
	return true
}
