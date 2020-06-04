package commands

import (
	"log"
	"strings"

	"github.com/atla/owndnd/pkg/mudserver/game/def"
	"github.com/atla/owndnd/pkg/mudserver/game/messages"
)

// CommandProcessor ... global user struct to control logins
type CommandProcessor struct {
	commands map[string]Command
}

// NewCommandProcessor .. creates a new command processor
func NewCommandProcessor() *CommandProcessor {
	var commandProcessor = &CommandProcessor{
		commands: make(map[string]Command),
	}
	// only once?
	commandProcessor.registerCommands()
	return commandProcessor
}

// RegisterCommand ... register
func (commandProcessor *CommandProcessor) RegisterCommand(key string, command Command) {
	commandProcessor.commands[key] = command
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

	commandProcessor.RegisterCommand("scream", &ScreamCommand{})
	commandProcessor.RegisterCommand("shrug", &ShrugCommand{})

}
