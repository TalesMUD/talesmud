package commands

import (
	"strings"

	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func normalizeCommand(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}

func commandEquals(input, name string) bool {
	return normalizeCommand(input) == normalizeCommand(name)
}

// actionArgs reports whether input invokes a room action. A one-word action
// also accepts a trailing argument ("deposit 20"). The argument is empty on
// an exact match. Multi-word action names stay exact so "examine moons"
// is not treated as "examine" plus an argument.
func actionArgs(input, name string) (string, bool) {
	in := normalizeCommand(input)
	nm := normalizeCommand(name)
	if in == "" || nm == "" {
		return "", false
	}
	if in == nm {
		return "", true
	}
	if strings.Contains(nm, " ") || len([]rune(nm)) < 2 {
		return "", false
	}
	prefix := nm + " "
	if strings.HasPrefix(in, prefix) {
		return strings.TrimSpace(in[len(prefix):]), true
	}
	return "", false
}

func roomActionMatches(game def.GameCtrl, message *messages.Message) bool {
	if game == nil || message == nil || message.Character == nil || message.Character.CurrentRoomID == "" {
		return false
	}
	room, err := game.GetFacade().RoomsService().FindByID(message.Character.CurrentRoomID)
	if err != nil || room == nil || room.Actions == nil {
		return false
	}
	for _, action := range *room.Actions {
		if _, ok := actionArgs(message.Data, action.Name); ok {
			return true
		}
	}
	return false
}
