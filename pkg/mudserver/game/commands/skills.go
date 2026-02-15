package commands

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/skills"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// SkillsCommand handles skill management (list, equip, unequip)
type SkillsCommand struct{}

// Key returns the command key matcher
func (command *SkillsCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute handles the skills command
func (command *SkillsCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	parts := strings.Fields(message.Data)

	if len(parts) < 2 {
		return command.listSkills(game, message)
	}

	subcommand := strings.ToLower(parts[1])
	switch subcommand {
	case "equip":
		if len(parts) < 3 {
			game.SendMessage() <- message.Reply("Usage: skills equip <skill name>")
			return true
		}
		return command.equipSkill(game, message, strings.Join(parts[2:], " "))
	case "unequip":
		if len(parts) < 3 {
			game.SendMessage() <- message.Reply("Usage: skills unequip <skill name>")
			return true
		}
		return command.unequipSkill(game, message, strings.Join(parts[2:], " "))
	default:
		return command.listSkills(game, message)
	}
}

func (command *SkillsCommand) listSkills(game def.GameCtrl, message *messages.Message) bool {
	char := message.Character
	classID := char.Class.ID
	level := char.Level

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n═══════ SKILLS (%s L%d) ═══════\n\n", classID, level))

	// Show equipped skills
	maxSlots := skills.MaxSkillSlots(classID, level)
	sb.WriteString(fmt.Sprintf("EQUIPPED (%d/%d slots):\n", len(char.EquippedSkills), maxSlots))
	if len(char.EquippedSkills) == 0 {
		sb.WriteString("  (none)\n")
	} else {
		for i, sid := range char.EquippedSkills {
			skill := skills.SkillByID(sid)
			if skill == nil {
				sb.WriteString(fmt.Sprintf("  %d. [unknown: %s]\n", i+1, sid))
				continue
			}
			cost := formatSkillCost(skill)
			sb.WriteString(fmt.Sprintf("  %d. %-20s %s (%s)\n", i+1, skill.Name, skill.Description, cost))
		}
	}

	// Show available (unlocked but not equipped)
	available := skills.AvailableSkills(classID, level)
	equippedSet := make(map[string]bool)
	for _, sid := range char.EquippedSkills {
		equippedSet[sid] = true
	}

	sb.WriteString("\nAVAILABLE:\n")
	hasAvailable := false
	for _, skill := range available {
		if equippedSet[skill.ID] {
			continue
		}
		hasAvailable = true
		cost := formatSkillCost(skill)
		sb.WriteString(fmt.Sprintf("  %-20s L%d  %s (%s)\n", skill.Name, skill.LevelRequired, skill.Description, cost))
	}
	if !hasAvailable {
		sb.WriteString("  (all available skills equipped)\n")
	}

	// Show locked skills
	allClassSkills := skills.SkillsForClass(classID)
	sb.WriteString("\nLOCKED:\n")
	hasLocked := false
	for _, skill := range allClassSkills {
		if skill.LevelRequired <= level {
			continue
		}
		hasLocked = true
		sb.WriteString(fmt.Sprintf("  %-20s Unlocks at L%d\n", skill.Name, skill.LevelRequired))
	}
	if !hasLocked {
		sb.WriteString("  (all skills unlocked)\n")
	}

	sb.WriteString("\nCommands: skills equip <name> | skills unequip <name>")
	sb.WriteString("\n═══════════════════════════════════")

	game.SendMessage() <- message.Reply(sb.String())
	return true
}

func (command *SkillsCommand) equipSkill(game def.GameCtrl, message *messages.Message, skillName string) bool {
	combatEngine := game.GetCombatEngine()
	if combatEngine != nil && combatEngine.IsPlayerInCombat(message.Character.Entity.ID) {
		game.SendMessage() <- message.Reply("You cannot change skills during combat.")
		return true
	}

	char := message.Character
	classID := char.Class.ID
	level := char.Level

	// Check slot limit
	maxSlots := skills.MaxSkillSlots(classID, level)
	if len(char.EquippedSkills) >= maxSlots {
		game.SendMessage() <- message.Reply(fmt.Sprintf("All %d skill slots are full. Unequip a skill first.", maxSlots))
		return true
	}

	// Find the skill by name
	skillNameLower := strings.ToLower(skillName)
	available := skills.AvailableSkills(classID, level)
	var found *skills.Skill
	for i := range available {
		if strings.ToLower(available[i].Name) == skillNameLower ||
			strings.Contains(strings.ToLower(available[i].Name), skillNameLower) {
			found = available[i]
			break
		}
	}

	if found == nil {
		game.SendMessage() <- message.Reply(fmt.Sprintf("No available skill matching '%s'. Use 'skills' to see available skills.", skillName))
		return true
	}

	// Check if already equipped
	for _, sid := range char.EquippedSkills {
		if sid == found.ID {
			game.SendMessage() <- message.Reply(fmt.Sprintf("%s is already equipped.", found.Name))
			return true
		}
	}

	// Equip
	char.EquippedSkills = append(char.EquippedSkills, found.ID)
	game.GetFacade().CharactersService().Update(char.ID, char)

	game.SendMessage() <- message.Reply(fmt.Sprintf("Equipped %s! (%d/%d slots)", found.Name, len(char.EquippedSkills), maxSlots))
	return true
}

func (command *SkillsCommand) unequipSkill(game def.GameCtrl, message *messages.Message, skillName string) bool {
	combatEngine := game.GetCombatEngine()
	if combatEngine != nil && combatEngine.IsPlayerInCombat(message.Character.Entity.ID) {
		game.SendMessage() <- message.Reply("You cannot change skills during combat.")
		return true
	}

	char := message.Character
	classID := char.Class.ID
	level := char.Level

	// Find the equipped skill by name
	skillNameLower := strings.ToLower(skillName)
	foundIdx := -1
	var foundSkill *skills.Skill
	for i, sid := range char.EquippedSkills {
		skill := skills.SkillByID(sid)
		if skill != nil && (strings.ToLower(skill.Name) == skillNameLower ||
			strings.Contains(strings.ToLower(skill.Name), skillNameLower)) {
			foundIdx = i
			foundSkill = skill
			break
		}
	}

	if foundIdx < 0 {
		game.SendMessage() <- message.Reply(fmt.Sprintf("No equipped skill matching '%s'.", skillName))
		return true
	}

	// Remove from equipped
	char.EquippedSkills = append(char.EquippedSkills[:foundIdx], char.EquippedSkills[foundIdx+1:]...)
	game.GetFacade().CharactersService().Update(char.ID, char)

	maxSlots := skills.MaxSkillSlots(classID, level)
	game.SendMessage() <- message.Reply(fmt.Sprintf("Unequipped %s. (%d/%d slots)", foundSkill.Name, len(char.EquippedSkills), maxSlots))
	return true
}

func formatSkillCost(skill *skills.Skill) string {
	if skill.ResourceType == skills.ResourceMana {
		return fmt.Sprintf("%d mana", skill.ManaCost)
	}
	return fmt.Sprintf("%d round CD", skill.CooldownRounds)
}
