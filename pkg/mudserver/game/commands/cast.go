package commands

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/skills"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// CastCommand handles using skills/spells in combat
type CastCommand struct{}

// Key returns the command key matcher
func (command *CastCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute handles the cast command
func (command *CastCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	combatEngine := game.GetCombatEngine()
	if combatEngine == nil || !combatEngine.IsPlayerInCombat(message.Character.Entity.ID) {
		game.SendMessage() <- message.Reply("You can only use skills in combat.")
		return true
	}

	// Parse: cast <skill_name> [target]
	parts := strings.Fields(message.Data)
	if len(parts) < 2 {
		// Show equipped skills
		command.showEquippedSkills(game, message)
		return true
	}

	// Try to match skill name (may be multi-word like "power strike")
	// Strategy: try longest match first, then shorter
	skillName := ""
	targetName := ""
	equippedSkills := message.Character.EquippedSkills

	// Try matching from longest possible skill name to shortest
	for end := len(parts); end >= 2; end-- {
		candidate := strings.Join(parts[1:end], " ")
		candidateLower := strings.ToLower(candidate)

		for _, sid := range equippedSkills {
			skill := skills.SkillByID(sid)
			if skill != nil && strings.ToLower(skill.Name) == candidateLower {
				skillName = sid
				if end < len(parts) {
					targetName = strings.Join(parts[end:], " ")
				}
				break
			}
		}
		if skillName != "" {
			break
		}
	}

	// Also try partial match if no exact match found
	if skillName == "" {
		candidate := strings.ToLower(strings.Join(parts[1:], " "))
		for _, sid := range equippedSkills {
			skill := skills.SkillByID(sid)
			if skill != nil && strings.Contains(strings.ToLower(skill.Name), candidate) {
				skillName = sid
				break
			}
		}
		// If still no match, try just the first word after "cast"
		if skillName == "" {
			candidate = strings.ToLower(parts[1])
			for _, sid := range equippedSkills {
				skill := skills.SkillByID(sid)
				if skill != nil && strings.Contains(strings.ToLower(skill.Name), candidate) {
					skillName = sid
					if len(parts) > 2 {
						targetName = strings.Join(parts[2:], " ")
					}
					break
				}
			}
		}
	}

	if skillName == "" {
		game.SendMessage() <- message.Reply(fmt.Sprintf("You don't have a skill matching '%s' equipped. Use 'skills' to see your skills.", strings.Join(parts[1:], " ")))
		return true
	}

	skill := skills.SkillByID(skillName)

	// Resolve target
	targetID := ""
	if skill.Target == skills.TargetEnemy || skill.Target == skills.TargetAllEnemies {
		instance := combatEngine.GetCombatInstance(message.Character.Entity.ID)
		if instance != nil {
			if targetName != "" {
				// Find enemy by name
				targetNameLower := strings.ToLower(targetName)
				for _, enemy := range instance.Enemies {
					if enemy.IsAlive && strings.Contains(strings.ToLower(enemy.Name), targetNameLower) {
						targetID = enemy.ID
						break
					}
				}
				if targetID == "" {
					game.SendMessage() <- message.Reply(fmt.Sprintf("No living enemy matching '%s'.", targetName))
					return true
				}
			} else {
				// Use auto-attack target or first living enemy
				player := instance.GetPlayerByID(message.Character.Entity.ID)
				if player != nil && player.AutoAttackTargetID != "" {
					target := instance.GetCombatantByID(player.AutoAttackTargetID)
					if target != nil && target.IsAlive {
						targetID = player.AutoAttackTargetID
					}
				}
				if targetID == "" {
					livingEnemies := instance.GetLivingEnemies()
					if len(livingEnemies) > 0 {
						targetID = livingEnemies[0].ID
					}
				}
			}
		}
	}

	// Queue the skill
	combatEngine.QueuePlayerSkill(message.Character.Entity.ID, skillName, targetID)

	// Build confirmation message
	msg := fmt.Sprintf("You prepare to use %s", skill.Name)
	if skill.ResourceType == skills.ResourceMana {
		msg += fmt.Sprintf(" (%d mana)", skill.ManaCost)
	}
	msg += " on your next turn."
	game.SendMessage() <- message.Reply(msg)

	return true
}

func (command *CastCommand) showEquippedSkills(game def.GameCtrl, message *messages.Message) {
	equipped := message.Character.EquippedSkills
	if len(equipped) == 0 {
		game.SendMessage() <- message.Reply("You have no skills equipped. Use 'skills' to see available skills.")
		return
	}

	var sb strings.Builder
	sb.WriteString("Equipped skills:\n")
	for i, sid := range equipped {
		skill := skills.SkillByID(sid)
		if skill == nil {
			continue
		}
		cost := ""
		if skill.ResourceType == skills.ResourceMana {
			cost = fmt.Sprintf("%d mana", skill.ManaCost)
		} else {
			cost = fmt.Sprintf("%d round CD", skill.CooldownRounds)
		}
		sb.WriteString(fmt.Sprintf("  %d. %s - %s (%s)\n", i+1, skill.Name, skill.Description, cost))
	}
	sb.WriteString("\nUsage: cast <skill name> [target] — or just type the slot number (1-4)")

	game.SendMessage() <- message.Reply(sb.String())
}
