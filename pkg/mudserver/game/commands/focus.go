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
	query := strings.TrimSpace(strings.TrimPrefix(message.Data, strings.Fields(message.Data)[0]))
	if query == "" {
		game.SendMessage() <- message.Reply("Focus whom? Usage: focus <enemy ID or name>")
		return true
	}
	var targetID, targetName string
	for _, enemy := range instance.Enemies {
		if enemy.IsAlive && strings.EqualFold(enemy.ID, query) {
			targetID, targetName = enemy.ID, enemy.Name
			break
		}
	}
	if targetID == "" {
		for _, enemy := range instance.Enemies {
			if enemy.IsAlive && strings.Contains(strings.ToLower(enemy.Name), strings.ToLower(query)) {
				targetID, targetName = enemy.ID, enemy.Name
				break
			}
		}
	}
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
