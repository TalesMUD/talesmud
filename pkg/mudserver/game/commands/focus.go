package commands

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// FocusCommand changes automatic attack and hostile skill aim without spending a turn.
type FocusCommand struct{}

func (*FocusCommand) Key() CommandKey { return &StartsWithCommandKey{} }

func (*FocusCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("Select a character first.")
		return true
	}
	engine := game.GetCombatEngine()
	if !isInActiveCombat(game, message.Character, engine) {
		game.SendMessage() <- message.Reply("You can only focus an enemy in combat.")
		return true
	}
	instance := engine.GetCombatInstance(message.Character.ID)
	if instance == nil {
		return true
	}
	parts := strings.Fields(message.Data)
	query := ""
	if len(parts) > 1 {
		query = strings.Join(parts[1:], " ")
	}
	if query == "" {
		game.SendMessage() <- message.Reply("Focus whom? Usage: focus <enemy ID or name>")
		return true
	}
	roomID := ""
	if message.Character != nil {
		roomID = message.Character.CurrentRoomID
	}
	targetID, targetName := resolveInCombatTarget(game, roomID, instance.Enemies, query)
	if targetID == "" {
		game.SendMessage() <- message.Reply("That enemy is not alive in this fight.")
		return true
	}
	if player := instance.GetPlayerByID(message.Character.ID); player != nil && player.AutoAttackTargetID == targetID {
		return true
	}
	engine.SetAutoAttackTarget(message.Character.ID, targetID)
	game.SendMessage() <- messages.NewCombatActionMessage(message.FromUser.ID,
		fmt.Sprintf("You focus %s.", targetName),
		messages.CombatActionMessage{ActorID: message.Character.ID, TargetID: targetID, Action: "focus"})
	return true
}
