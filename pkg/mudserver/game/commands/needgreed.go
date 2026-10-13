package commands

import (
	"strings"

	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// NeedCommand records Need on the open party loot roll.
type NeedCommand struct{}

// Execute handles: need
func (c *NeedCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	return recordLootRoll(game, message, "need")
}

// Key returns matching keys
func (c *NeedCommand) Key() CommandKey { return &ExactCommandKey{} }

// GreedCommand records Greed on the open party loot roll.
type GreedCommand struct{}

// Execute handles: greed
func (c *GreedCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	return recordLootRoll(game, message, "greed")
}

// Key returns matching keys
func (c *GreedCommand) Key() CommandKey { return &ExactCommandKey{} }

// PassCommand records Pass on the open party loot roll.
type PassCommand struct{}

// Execute handles: pass
func (c *PassCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	return recordLootRoll(game, message, "pass")
}

// Key returns matching keys
func (c *PassCommand) Key() CommandKey { return &ExactCommandKey{} }

// LootRollCommand handles: loot need|greed|pass
type LootRollCommand struct{}

// Execute handles loot need/greed/pass aliases.
func (c *LootRollCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message == nil || message.FromUser == nil || message.Character == nil {
		return true
	}
	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Usage: loot need|greed|pass")
		return true
	}
	choice := strings.ToLower(parts[1])
	switch choice {
	case "need", "greed", "pass":
		return recordLootRoll(game, message, choice)
	default:
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "Usage: loot need|greed|pass")
		return true
	}
}

// Key returns matching keys
func (c *LootRollCommand) Key() CommandKey { return &StartsWithCommandKey{} }

func recordLootRoll(game def.GameCtrl, message *messages.Message, choice string) bool {
	if message == nil || message.FromUser == nil || message.Character == nil {
		return true
	}
	if !game.HasActiveNeedGreed(message.Character.ID) {
		game.SendMessage() <- messages.Reply(message.FromUser.ID, "There is no loot roll open.")
		return true
	}
	game.RecordNeedGreedChoice(message.Character.ID, message.FromUser.ID, choice)
	return true
}
