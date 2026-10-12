package commands

import (
	"strings"

	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// NPC engagement is optional for other GameCtrl implementations.
type npcEngagementCtrl interface {
	EngageNPC(characterID, npcID, kind string)
	CloseNPCSession(characterID, kind string)
	NPCSessionOpen(characterID, npcID, kind string) bool
}

func engageNPC(game def.GameCtrl, message *messages.Message, id, kind string) {
	if message == nil || message.Character == nil {
		return
	}
	if ctrl, ok := game.GetNPCInstanceManager().(npcEngagementCtrl); ok {
		ctrl.EngageNPC(message.Character.ID, id, kind)
	}
}
func closeNPCSession(game def.GameCtrl, message *messages.Message, kind string) {
	if message == nil || message.Character == nil {
		return
	}
	if ctrl, ok := game.GetNPCInstanceManager().(npcEngagementCtrl); ok {
		ctrl.CloseNPCSession(message.Character.ID, kind)
	}
}
func npcSessionOpen(game def.GameCtrl, characterID, id, kind string) bool {
	ctrl, ok := game.GetNPCInstanceManager().(npcEngagementCtrl)
	return ok && ctrl.NPCSessionOpen(characterID, id, kind)
}

// CloseNPCSessionCommand is emitted when a client dismisses an overlay.
type CloseNPCSessionCommand struct{}

func (*CloseNPCSessionCommand) Key() CommandKey { return &ExactCommandKey{} }
func (*CloseNPCSessionCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		return true
	}
	kind := "shop"
	if strings.EqualFold(message.Data, "dialogclose") {
		kind = "dialog"
		convs, _ := game.GetFacade().ConversationsService().FindAllForCharacter(message.Character.ID)
		for _, conv := range convs {
			game.GetFacade().ConversationsService().ResetConversation(conv)
		}
		if message.FromUser != nil {
			game.SendMessage() <- messages.NewDialogEndMessage(message.FromUser.ID, "", "")
		}
	}
	closeNPCSession(game, message, kind)
	return true
}

// TargetNPCCommand supports selecting an NPC before committing an action.
type TargetNPCCommand struct{}

func (*TargetNPCCommand) Key() CommandKey { return &StartsWithCommandKey{} }
func (*TargetNPCCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		return true
	}
	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		game.SendMessage() <- message.Reply("Target whom?")
		return true
	}
	if n := game.GetNPCInstanceManager().FindInstanceByNameInRoom(message.Character.CurrentRoomID, strings.Join(parts[1:], " ")); n != nil {
		engageNPC(game, message, n.ID, "")
		game.SendMessage() <- message.Reply("You target " + n.Name + ".")
	} else {
		game.SendMessage() <- message.Reply("There is no such target here.")
	}
	return true
}
