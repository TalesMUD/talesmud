package commands

import (
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// BindCommand handles binding to a respawn location
type BindCommand struct {
}

// Key returns the command key matcher
func (command *BindCommand) Key() CommandKey { return &ExactCommandKey{} }

// Execute handles the bind command
func (command *BindCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	if message.Character.CurrentRoomID == "" {
		game.SendMessage() <- message.Reply("You are not in a room.")
		return true
	}

	// Get the current room
	room, err := game.GetFacade().RoomsService().FindByID(message.Character.CurrentRoomID)
	if err != nil {
		game.SendMessage() <- message.Reply("Error finding room.")
		return true
	}

	// Check if room supports binding
	if !room.CanBind {
		game.SendMessage() <- message.Reply("You cannot bind your soul here. Try an inn or temple.")
		return true
	}

	// Check if already bound here
	if message.Character.BoundRoomID == room.ID {
		game.SendMessage() <- message.Reply("You are already bound to this location.")
		return true
	}

	// Bind to this room
	message.Character.BoundRoomID = room.ID
	if err := game.GetFacade().CharactersService().Update(message.Character.ID, message.Character); err != nil {
		game.SendMessage() <- message.Reply("Error binding to this location.")
		return true
	}

	game.SendMessage() <- message.Reply("You bind your soul to " + room.Name + ". You will respawn here upon death.")
	return true
}
