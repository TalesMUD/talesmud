package commands

import (
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// Look ... executes lookat command
func Look(room *rooms.Room, game def.GameCtrl, message *messages.Message) bool {

	parts := strings.Fields(message.Data)

	// look at room
	if len(parts) == 1 {

		result := ""

		if room.Detail != "" {
			result = "You look around...\n"
			result += room.Detail
		} else {
			result = "You look around... nothing else to see here."
		}

		game.SendMessage() <- message.Reply(result)
		return true

	} else {
		// look at object
		// find object
		// look at NPC
	}

	return false
}
