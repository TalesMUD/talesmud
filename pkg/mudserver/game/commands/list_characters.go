package commands

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// ListCharactersCommand represents the command to list a user's characters
type ListCharactersCommand struct{}

// Key returns the command key for the ListCharactersCommand
func (command *ListCharactersCommand) Key() CommandKey {
	return &ExactCommandKey{}
}

// Execute handles the execution of the list characters command
func (command *ListCharactersCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	characters, err := game.GetFacade().CharactersService().FindAllForUser(message.FromUser.ID)
	if err != nil {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Error retrieving characters. Please try again later.")
		return false
	}

	if len(characters) == 0 {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "You have no characters.")
		return true
	}

	// Build the character list
	var characterList []string
	for _, character := range characters {
		characterInfo := fmt.Sprintf("- %s [LVL %d %s %dxp] %s - %s",
			character.Name,
			character.Level,
			character.Class.Name,
			character.XP,
			character.Race.Name,
			character.Description)
		characterList = append(characterList, characterInfo)
	}

	// Compose the final message
	result := fmt.Sprintf("Your Characters:\n%s\n\nTo select a character, use: sc [charactername]",
		strings.Join(characterList, "\n"))

	game.SendMessage() <- messages.Reply(message.FromUser.ID, result)
	return true
}
