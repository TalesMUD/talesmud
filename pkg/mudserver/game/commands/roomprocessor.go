package commands

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/hoisie/mustache"
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

// Process handles room-based commands
func (roomProcessor *RoomProcessor) Process(game def.GameCtrl, message *messages.Message) bool {

	if message.Character == nil || message.Character.CurrentRoomID == "" {
		return false
	}

	if room, err := game.GetFacade().RoomsService().FindByID(message.Character.CurrentRoomID); err == nil {
		// First, check if this is a dialog selection (number input during conversation)
		// This takes priority over other room commands
		if DialogSelectCommand(room, game, message) {
			return true
		}

		parts := strings.Fields(message.Data)

		if len(parts) > 0 {
			var key = parts[0]
			if command, ok := roomProcessor.commands[key]; ok {

				log.Println("Found command " + key + " executing...")
				return command(room, game, message)

			} else if command, ok := roomProcessor.matchesDynamicCommand(key, room, message); ok {
				// not handled by static command handlers, check dynamic conditions suchs as actions and custom commands
				log.Println("Found dynamic command " + key + " executing...")
				return command(room, game, message)
			}

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
	roomProcessor.RegisterCommand(Display, "r", "room")
}

func (roomProcessor *RoomProcessor) matchesDynamicCommand(key string, room *rooms.Room, message *messages.Message) (RoomCommand, bool) {

	for _, exit := range *room.Exits {
		if strings.HasPrefix(message.Data, exit.Name) {
			// custom exit
			return TakeExit(exit.Name), true
		}
	}

	// create context for actions
	context := map[string]string{
		"characterName": message.Character.Name,
		"roomName":      room.Name,
		"roomArea":      room.Area,
	}

	for _, action := range *room.Actions {
		// support "longer" command inputs as custom action triggers: e.g. "move rocks"
		if strings.HasPrefix(message.Data, action.Name) {

			desc := mustache.Render(action.Description, context)

			switch action.Type {

			case rooms.RoomActionTypeResponse:
				return func(room *rooms.Room, game def.GameCtrl, message *messages.Message) bool {

					game.SendMessage() <- message.Reply(desc)
					return true
				}, true

			case rooms.RoomActionTypeRoomResponse:
				return func(room *rooms.Room, game def.GameCtrl, message *messages.Message) bool {

					actionResponseToRoom := message.Reply(desc)
					actionResponseToRoom.AudienceID = room.ID
					actionResponseToRoom.Audience = messages.MessageAudienceRoom

					game.SendMessage() <- actionResponseToRoom
					return true
				}, true

			case rooms.RoomActionTypeScript:
				fallthrough
			default:
				log.WithField("type", action.Type).WithField("name", action.Name).Error("matched action name but unsupported or empty action type ")
			}

		}
	}
	return nil, false
}
