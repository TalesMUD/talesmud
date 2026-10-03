package commands

import (
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/combat"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// BoltCommand arms scrap. It forfeits the swing. Rigger only, once per fight.
type BoltCommand struct{}

func (command *BoltCommand) Key() CommandKey { return &StartsWithCommandKey{} }

func (command *BoltCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil || message.Character.Entity == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}
	combatEngine := game.GetCombatEngine()
	if !isInActiveCombat(game, message.Character, combatEngine) {
		game.SendMessage() <- message.Reply("You are not in combat.")
		return true
	}
	parts := strings.Fields(message.Data)
	targetName := ""
	if len(parts) >= 2 {
		targetName = strings.Join(parts[1:], " ")
	}
	targetID := resolveBoltCommandTarget(game, message, combatEngine, targetName)
	combatEngine.QueuePlayerAction(message.Character.Entity.ID, combat.CombatActionBolt, targetID)
	game.SendMessage() <- message.Reply("You prepare to bolt scrap.")
	return true
}

func resolveBoltCommandTarget(game def.GameCtrl, message *messages.Message, combatEngine def.CombatEngineCtrl, targetName string) string {
	if targetName == "" || message.Character == nil || message.Character.Entity == nil {
		return ""
	}
	instance := combatEngine.GetCombatInstance(message.Character.Entity.ID)
	if instance == nil {
		return ""
	}
	if id, _ := resolveInCombatTarget(game, message.Character.CurrentRoomID, instance.Enemies, targetName); id != "" {
		return id
	}
	lower := strings.ToLower(targetName)
	for i := range instance.Players {
		p := &instance.Players[i]
		if p.IsAlive && (strings.EqualFold(p.ID, targetName) || strings.Contains(strings.ToLower(p.Name), lower)) {
			return p.ID
		}
	}
	if strings.EqualFold(targetName, "self") || strings.EqualFold(targetName, "me") {
		return message.Character.Entity.ID
	}
	return ""
}

// RigCommand drops a turret. It forfeits the swing. Rigger only, once per fight.
type RigCommand struct{}

func (command *RigCommand) Key() CommandKey { return &StartsWithCommandKey{} }

func (command *RigCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil || message.Character.Entity == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}
	combatEngine := game.GetCombatEngine()
	if !isInActiveCombat(game, message.Character, combatEngine) {
		game.SendMessage() <- message.Reply("You are not in combat.")
		return true
	}
	combatEngine.QueuePlayerAction(message.Character.Entity.ID, combat.CombatActionRig, "")
	game.SendMessage() <- message.Reply("You prepare to drop a rig.")
	return true
}
