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

			log.Println("Found command " + key + " executing...")
			return val.Execute(game, message)
		}
	}

	return false

}

func (commandProcessor *CommandProcessor) registerCommands() {

	commandProcessor.RegisterCommand(&ScreamCommand{}, "scream through the room", "scream")
	commandProcessor.RegisterCommand(&ShrugCommand{}, "shrug emote", "shrug")
	commandProcessor.RegisterCommand(&SelectCharacterCommand{}, "select a character, use: sc [charactername]", "sc", "selectcharacter")
	commandProcessor.RegisterCommand(&ListCharactersCommand{}, "list all your characters", "lc", "listcharacters")
	commandProcessor.RegisterCommand(&HelpCommand{processor: commandProcessor}, "are you really asking?", "h", "help")
	commandProcessor.RegisterCommand(&WhoCommand{}, "list all online players", "who")
	commandProcessor.RegisterCommand(&InventoryCommand{}, "Display your inventory", "inventory", "i")
	commandProcessor.RegisterCommand(&NewCharacterCommand{}, "Createa new character", "newcharacter", "nc")

}
