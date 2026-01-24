package commands

import (
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	m "github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// CharacterCommand displays character stats and info
type CharacterCommand struct {
}

// Key returns the command key matcher
func (command *CharacterCommand) Key() CommandKey { return &ExactCommandKey{} }

// Execute displays character stats
func (command *CharacterCommand) Execute(game def.GameCtrl, message *m.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	char := message.Character
	var sb strings.Builder

	// Header
	sb.WriteString("=== ")
	sb.WriteString(char.Name)
	sb.WriteString(" ===\n")

	// Basic info
	sb.WriteString("Race: ")
	sb.WriteString(char.Race.Name)
	sb.WriteString(" | Class: ")
	sb.WriteString(char.Class.Name)
	sb.WriteString("\n")

	sb.WriteString("Level: ")
	sb.WriteString(itoa(int(char.Level)))
	sb.WriteString(" | XP: ")
	sb.WriteString(itoa(int(char.XP)))
	sb.WriteString("\n\n")

	// Vitals
	sb.WriteString("[Vitals]\n")
	sb.WriteString("  HP: ")
	sb.WriteString(itoa(int(char.CurrentHitPoints)))
	sb.WriteString("/")
	sb.WriteString(itoa(int(char.MaxHitPoints)))
	sb.WriteString("\n")

	// Gold
	sb.WriteString("  Gold: ")
	sb.WriteString(itoa64(char.Gold))
	sb.WriteString("\n\n")

	// Attributes
	sb.WriteString("[Attributes]\n")
	if len(char.Attributes) > 0 {
		for _, attr := range char.Attributes {
			sb.WriteString("  ")
			sb.WriteString(attr.Short)
			sb.WriteString(": ")
			sb.WriteString(itoa(int(attr.Value)))
			mod := (attr.Value - 10) / 2
			if mod >= 0 {
				sb.WriteString(" (+")
			} else {
				sb.WriteString(" (")
			}
			sb.WriteString(itoa(int(mod)))
			sb.WriteString(")\n")
		}
	} else {
		sb.WriteString("  (No attributes defined)\n")
	}

	// Combat stats
	sb.WriteString("\n[Combat]\n")
	weaponDmg := char.GetWeaponDamage()
	armorDef := char.GetArmorDefense()
	sb.WriteString("  Weapon Damage: ")
	sb.WriteString(itoa(int(weaponDmg)))
	if weaponDmg == 1 {
		sb.WriteString(" (Unarmed)")
	}
	sb.WriteString("\n")
	sb.WriteString("  Armor Defense: ")
	sb.WriteString(itoa(int(armorDef)))
	sb.WriteString("\n")
	if char.InCombat {
		sb.WriteString("  Status: IN COMBAT\n")
	}

	// Equipment
	sb.WriteString("\n[Equipment]\n")
	equipSlots := []struct {
		slot items.ItemSlot
		name string
	}{
		{items.ItemSlotHead, "Head"},
		{items.ItemSlotNeck, "Neck"},
		{items.ItemSlotChest, "Chest"},
		{items.ItemSlotHands, "Hands"},
		{items.ItemSlotLegs, "Legs"},
		{items.ItemSlotBoots, "Boots"},
		{items.ItemSlotMainHand, "Main Hand"},
		{items.ItemSlotOffHand, "Off Hand"},
		{items.ItemSlotRing1, "Ring 1"},
		{items.ItemSlotRing2, "Ring 2"},
	}

	hasEquipment := false
	for _, es := range equipSlots {
		if item, ok := char.EquippedItems[es.slot]; ok && item != nil {
			hasEquipment = true
			sb.WriteString("  ")
			sb.WriteString(es.name)
			sb.WriteString(": ")
			sb.WriteString(item.Name)

			// Show quality for non-normal items
			if item.Quality != "" && item.Quality != items.ItemQualityNormal {
				sb.WriteString(" [")
				sb.WriteString(strings.ToUpper(string(item.Quality)))
				sb.WriteString("]")
			}

			// Show key stats
			if dmg, ok := item.Attributes["damage"]; ok {
				sb.WriteString(" (Dmg: ")
				sb.WriteString(formatAttrValue(dmg))
				sb.WriteString(")")
			}
			if def, ok := item.Attributes["defense"]; ok {
				sb.WriteString(" (Def: ")
				sb.WriteString(formatAttrValue(def))
				sb.WriteString(")")
			}
			if arm, ok := item.Attributes["armor"]; ok {
				sb.WriteString(" (Armor: ")
				sb.WriteString(formatAttrValue(arm))
				sb.WriteString(")")
			}
			sb.WriteString("\n")
		}
	}

	if !hasEquipment {
		sb.WriteString("  (Nothing equipped)\n")
	}

	game.SendMessage() <- message.Reply(sb.String())
	return true
}

// formatAttrValue converts an attribute value to string
func formatAttrValue(val interface{}) string {
	switch v := val.(type) {
	case float64:
		return itoa(int(v))
	case int:
		return itoa(v)
	case int32:
		return itoa(int(v))
	case int64:
		return itoa64(v)
	default:
		return "?"
	}
}
