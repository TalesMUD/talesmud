package commands

import (
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// Look executes the look command, allowing players to observe their surroundings or specific objects
func Look(room *rooms.Room, game def.GameCtrl, message *messages.Message) bool {
	parts := strings.Fields(message.Data)

	// Handle looking at the room (no specific target)
	if len(parts) == 1 {
		return lookAtRoom(room, game, message)
	}

	// Handle looking at a specific target
	target := strings.Join(parts[1:], " ")
	return lookAtTarget(room, game, message, target)
}

// lookAtRoom handles the case when a player looks at the room without specifying a target
func lookAtRoom(room *rooms.Room, game def.GameCtrl, message *messages.Message) bool {
	var result string

	if room.Detail != "" {
		result = "You look around...\n" + room.Detail
	} else {
		result = "You look around... nothing else to see here."
	}

	// TODO: Add information about visible objects and NPCs in the room

	game.SendMessage() <- message.Reply(result)
	return true
}

// lookAtTarget handles the case when a player looks at a specific object or NPC
func lookAtTarget(room *rooms.Room, game def.GameCtrl, message *messages.Message, target string) bool {
	// TODO: Implement logic to find and describe the target object or NPC
	// This could include:
	// 1. Searching for objects in the room
	// 2. Searching for NPCs in the room
	// 3. Checking for special keywords (e.g., "north", "south" for exits)

	game.SendMessage() <- message.Reply("You don't see anything special about " + target + ".")
	return true
}
