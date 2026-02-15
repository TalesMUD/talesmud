package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
	m "github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// SpendCommand handles distributing attribute points
type SpendCommand struct{}

// Key returns the command key matcher
func (command *SpendCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute handles the spend command
func (command *SpendCommand) Execute(game def.GameCtrl, message *m.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	char := message.Character
	parts := strings.Fields(message.Data)

	// "spend" with no args -> show status
	if len(parts) < 2 {
		return command.showStatus(game, message)
	}

	// Parse arguments - support both "spend dex 1" and "spend 1 dex"
	var attrShort string
	amount := int32(1)

	if len(parts) == 2 {
		// "spend dex" — single arg is the attribute, amount defaults to 1
		attrShort = strings.ToUpper(parts[1])
	} else {
		// Try "spend <attr> <amount>"
		if n, err := strconv.Atoi(parts[2]); err == nil && n > 0 {
			attrShort = strings.ToUpper(parts[1])
			amount = int32(n)
		} else if n, err := strconv.Atoi(parts[1]); err == nil && n > 0 {
			// Try "spend <amount> <attr>"
			attrShort = strings.ToUpper(parts[2])
			amount = int32(n)
		} else {
			game.SendMessage() <- message.Reply("Usage: spend <attribute> [amount]  (e.g., spend dex 3)")
			return true
		}
	}

	// Block if in combat
	combatEngine := game.GetCombatEngine()
	if combatEngine != nil && combatEngine.IsPlayerInCombat(char.Entity.ID) {
		game.SendMessage() <- message.Reply("You cannot allocate attribute points during combat.")
		return true
	}

	// Validate
	errMsg := leveling.ValidateAttributeSpend(char, attrShort, amount)
	if errMsg != "" {
		game.SendMessage() <- message.Reply(errMsg)
		return true
	}

	// Apply
	leveling.ApplyAttributeSpend(char, attrShort, amount)

	// Recalculate mana if INT changed (for caster classes)
	if attrShort == "INT" {
		newMaxMana := char.CalculateMaxMana()
		if newMaxMana > char.MaxMana {
			char.CurrentMana += (newMaxMana - char.MaxMana)
			char.MaxMana = newMaxMana
		}
	}

	// Recalculate max HP if STA changed
	if attrShort == "STA" {
		hpBonus := amount * int32(leveling.STAMultiplier)
		char.MaxHitPoints += hpBonus
		char.CurrentHitPoints += hpBonus
	}

	// Persist
	game.GetFacade().CharactersService().Update(char.ID, char)

	// Send confirmation
	newValue := char.GetAttribute(attrShort)
	newMod := char.GetAttributeModifier(attrShort)
	modStr := fmt.Sprintf("%+d", newMod)
	game.SendMessage() <- message.Reply(
		fmt.Sprintf("Spent %d point(s) on %s. New value: %d (%s). Unspent: %d",
			amount, attrShort, newValue, modStr, char.UnspentAttributePoints))

	// Send character update to frontend
	game.SendMessage() <- m.NewCharacterUpdateMessage(message.FromUser.ID, char)

	return true
}

func (command *SpendCommand) showStatus(game def.GameCtrl, message *m.Message) bool {
	char := message.Character
	caps := leveling.ClassAttributeCaps(char.Class.ID)

	var sb strings.Builder
	sb.WriteString("\n═══════ ATTRIBUTE POINTS ═══════\n\n")
	sb.WriteString(fmt.Sprintf("Unspent Points: %d\n\n", char.UnspentAttributePoints))
	sb.WriteString("ATTRIBUTE  VALUE  SPENT / CAP\n")
	sb.WriteString("─────────  ─────  ──────────\n")

	for _, attr := range char.Attributes {
		spent := int32(0)
		if char.SpentAttributePoints != nil {
			spent = char.SpentAttributePoints[attr.Short]
		}
		cap := caps[attr.Short]
		mod := (attr.Value - 10) / 2
		modStr := fmt.Sprintf("%+d", mod)
		sb.WriteString(fmt.Sprintf("  %-5s    %3d (%s)  %d / %d\n",
			attr.Short, attr.Value, modStr, spent, cap))
	}

	sb.WriteString("\nUsage: spend <attribute> [amount]")
	sb.WriteString("\nExample: spend str 3")
	sb.WriteString("\n═══════════════════════════════\n")

	game.SendMessage() <- message.Reply(sb.String())
	return true
}
