package commands

import (
	"log"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// RoomProcessor ... handles room based commands
type RoomProcessor struct {
	commands map[string]RoomCommand
}

//RoomCommand ...
type RoomCommand func(room *rooms.Room, game def.GameCtrl, message *messages.Message) bool

// NewRoomProcessor .. creates a new room processor
func NewRoomProcessor() *RoomProcessor {
	var roomProcessor = &RoomProcessor{
		commands: make(map[string]RoomCommand),
	}
	// only once?
	roomProcessor.registerCommands()
	return roomProcessor
}

// RegisterCommand ... register
func (roomProcessor *RoomProcessor) RegisterCommand(command RoomCommand, keys ...string) {

	for _, key := range keys {
		roomProcessor.commands[key] = command
	}
}

// Process ...asd
func (roomProcessor *RoomProcessor) Process(game def.GameCtrl, message *messages.Message) bool {

	if message.Character == nil || message.Character.CurrentRoomID == "" {
		return false
	}

	if room, err := game.GetFacade().RoomsService().FindByID(message.Character.CurrentRoomID); err == nil {
		parts := strings.Fields(message.Data)

		if len(parts) > 0 {
			var key = parts[0]
			if command, ok := roomProcessor.commands[key]; ok {

				log.Println("Found command " + key + " executing...")
				return command(room, game, message)
			}

			// not handled by static command handlers, check dynamic conditions suchs as actions and custom commands
			//TODO: implement actions
		}
	}
	// message has no room?

	return false
}

func (roomProcessor *RoomProcessor) registerCommands() {

	roomProcessor.RegisterCommand(TakeExit("north"), "n", "north")
	roomProcessor.RegisterCommand(TakeExit("south"), "s", "south")
	roomProcessor.RegisterCommand(TakeExit("east"), "e", "east")
	roomProcessor.RegisterCommand(TakeExit("west"), "w", "west")
	roomProcessor.RegisterCommand(Look, "l", "look")

}
