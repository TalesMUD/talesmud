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

func roomActionMatches(game def.GameCtrl, message *messages.Message) bool {
	if game == nil || message == nil || message.Character == nil || message.Character.CurrentRoomID == "" {
		return false
	}
	room, err := game.GetFacade().RoomsService().FindByID(message.Character.CurrentRoomID)
	if err != nil || room == nil || room.Actions == nil {
		return false
	}
	for _, action := range *room.Actions {
		if commandEquals(message.Data, action.Name) {
			return true
		}
	}
	return false
}
