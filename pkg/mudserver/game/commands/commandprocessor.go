package commands

import (
	"log"
	"strings"

	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// CommandProcessor ... global user struct to control logins
type CommandProcessor struct {
	commands map[string]Command
	Help     map[string]string
}

// NewCommandProcessor .. creates a new command processor
func NewCommandProcessor() *CommandProcessor {
	var commandProcessor = &CommandProcessor{
		commands: make(map[string]Command),
		Help:     make(map[string]string),
	}
	// only once?
	commandProcessor.registerCommands()
	return commandProcessor
}

// RegisterCommand ... register
func (commandProcessor *CommandProcessor) RegisterCommand(command Command, desc string, keys ...string) {

	cmds := "["
	for i, key := range keys {
		if i > 0 {
			cmds += ", "
		}
		cmds += key
		commandProcessor.commands[key] = command
	}

	cmds += "]"

	commandProcessor.Help[cmds] = desc
}

// Process ...asd
func (commandProcessor *CommandProcessor) Process(game def.GameCtrl, message *messages.Message) bool {

	parts := strings.Fields(message.Data)

	if len(parts) > 0 {
		var key = parts[0]
		if val, ok := commandProcessor.commands[key]; ok {

			// support for commands without parameters to enable input like "i did find something" but still support the command "i" for inventory
			if val.Key() != nil && val.Key().Matches(key, message.Data) == false {
				return false
			}

			log.Println("Found command " + key + " executing...")
			return val.Execute(game, message)
		}
	}

	return false

}

func (commandProcessor *CommandProcessor) registerCommands() {

	commandProcessor.RegisterCommand(&ScreamCommand{}, "Scream through the room", "scream")
	commandProcessor.RegisterCommand(&ShrugCommand{}, "Shrug emote", "shrug")
	commandProcessor.RegisterCommand(&SelectCharacterCommand{}, "Select a character, use: sc [charactername]", "sc", "selectcharacter")
	commandProcessor.RegisterCommand(&ListCharactersCommand{}, "List all your characters", "lc", "listcharacters")
	commandProcessor.RegisterCommand(&HelpCommand{processor: commandProcessor}, "Are you really asking?", "h", "help")
	commandProcessor.RegisterCommand(&WhoCommand{}, "List all online players", "who")
	commandProcessor.RegisterCommand(&InventoryCommand{}, "Display your inventory", "inventory", "i")
	commandProcessor.RegisterCommand(&NewCharacterCommand{}, "Create a new character", "newcharacter", "nc")

}
