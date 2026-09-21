package commands

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/scripts"
)

// GatherCommand helps players forage from world nodes: gather / forage / harvest [target]
// Exact room actions like "GATHER HERBS" still work via the room processor.
type GatherCommand struct{}

// Key ...
func (command *GatherCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute ...
func (command *GatherCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		return true
	}
	room, err := game.GetFacade().RoomsService().FindByID(message.Character.CurrentRoomID)
	if err != nil || room == nil {
		game.SendMessage() <- message.Reply("There is nothing to gather here.")
		return true
	}

	nodes := gatherActions(room)
	parts := strings.Fields(message.Data)
	target := ""
	if len(parts) > 1 {
		target = strings.ToLower(strings.Join(parts[1:], " "))
	}

	if len(nodes) == 0 {
		game.SendMessage() <- message.Reply("You find nothing worth gathering here.")
		return true
	}

	var chosen *rooms.Action
	if target == "" {
		if len(nodes) == 1 {
			chosen = &nodes[0]
		} else {
			names := make([]string, 0, len(nodes))
			for _, a := range nodes {
				names = append(names, a.Name)
			}
			game.SendMessage() <- message.Reply(fmt.Sprintf("Gatherables here: %s. Try: gather <name>", strings.Join(names, ", ")))
			return true
		}
	} else {
		for i := range nodes {
			n := strings.ToLower(nodes[i].Name)
			if n == target || strings.Contains(n, target) || strings.HasSuffix(n, " "+target) {
				chosen = &nodes[i]
				break
			}
			// strip leading gather/forage/harvest verb from action name for matching
			for _, verb := range []string{"gather ", "forage ", "harvest "} {
				if strings.HasPrefix(n, verb) {
					rest := strings.TrimPrefix(n, verb)
					if rest == target || strings.Contains(rest, target) {
						chosen = &nodes[i]
					}
				}
			}
			if chosen != nil {
				break
			}
		}
		if chosen == nil {
			game.SendMessage() <- message.Reply(fmt.Sprintf("No gatherable '%s' here. Try: gather", target))
			return true
		}
	}

	return runGatherAction(game, message, room, *chosen)
}

func gatherActions(room *rooms.Room) []rooms.Action {
	out := make([]rooms.Action, 0)
	if room == nil || room.Actions == nil {
		return out
	}
	for _, a := range *room.Actions {
		n := strings.ToLower(a.Name)
		if strings.HasPrefix(n, "gather") || strings.HasPrefix(n, "forage") || strings.HasPrefix(n, "harvest") {
			out = append(out, a)
		}
	}
	return out
}

func runGatherAction(game def.GameCtrl, message *messages.Message, room *rooms.Room, action rooms.Action) bool {
	switch action.Type {
	case rooms.RoomActionTypeScript:
		if action.ScriptId == "" {
			game.SendMessage() <- message.Reply("Nothing happens.")
			return true
		}
		script, err := game.GetFacade().ScriptsService().FindByID(action.ScriptId)
		if err != nil || script == nil {
			game.SendMessage() <- message.Reply("Nothing happens.")
			return true
		}
		ctx := scripts.NewScriptContext()
		ctx.Set("eventType", "room.action")
		ctx.Set("action", action.Name)
		ctx.Set("room", room)
		ctx.Set("roomID", room.ID)
		ctx.Set("character", message.Character)
		ctx.Set("characterID", message.Character.ID)
		if action.Params != nil {
			ctx.Set("params", action.Params)
		}
		_ = game.GetFacade().Runner().RunWithResult(*script, ctx)
		return true
	case rooms.RoomActionTypeResponse, rooms.RoomActionTypeRoomResponse:
		text := action.Response
		if text == "" {
			text = action.Description
		}
		if text == "" {
			text = "You gather what you can."
		}
		game.SendMessage() <- message.Reply(text)
		return true
	default:
		game.SendMessage() <- message.Reply("You can't gather that.")
		return true
	}
}
