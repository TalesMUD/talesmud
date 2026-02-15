package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/skills"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// SkillShortcutCommand handles numeric skill shortcuts (1, 2, 3, 4) in combat
type SkillShortcutCommand struct{}

// Key returns the command key matcher
func (command *SkillShortcutCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute handles the numeric shortcut command
func (command *SkillShortcutCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		return false
	}

	combatEngine := game.GetCombatEngine()
	if combatEngine == nil || !combatEngine.IsPlayerInCombat(message.Character.Entity.ID) {
		return false // Not in combat — let other commands handle it
	}

	parts := strings.Fields(message.Data)
	if len(parts) == 0 {
		return false
	}

	slotNum, err := strconv.Atoi(parts[0])
	if err != nil || slotNum < 1 {
		return false
	}

	equipped := message.Character.EquippedSkills
	if slotNum > len(equipped) {
		game.SendMessage() <- message.Reply(fmt.Sprintf("No skill in slot %d. You have %d skill(s) equipped.", slotNum, len(equipped)))
		return true
	}

	skillID := equipped[slotNum-1]
	skill := skills.SkillByID(skillID)
	if skill == nil {
		game.SendMessage() <- message.Reply(fmt.Sprintf("Skill in slot %d not found.", slotNum))
		return true
	}

	// Resolve target
	targetID := ""
	if skill.Target == skills.TargetEnemy || skill.Target == skills.TargetAllEnemies {
		instance := combatEngine.GetCombatInstance(message.Character.Entity.ID)
		if instance != nil {
			// Check if a target name was provided after the number
			if len(parts) > 1 {
				targetNameLower := strings.ToLower(strings.Join(parts[1:], " "))
				for _, enemy := range instance.Enemies {
					if enemy.IsAlive && strings.Contains(strings.ToLower(enemy.Name), targetNameLower) {
						targetID = enemy.ID
						break
					}
				}
				if targetID == "" {
					game.SendMessage() <- message.Reply(fmt.Sprintf("No living enemy matching '%s'.", strings.Join(parts[1:], " ")))
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
	combatEngine.QueuePlayerSkill(message.Character.Entity.ID, skillID, targetID)

	// Confirmation message
	msg := fmt.Sprintf("You prepare to use %s", skill.Name)
	if skill.ResourceType == skills.ResourceMana {
		msg += fmt.Sprintf(" (%d mana)", skill.ManaCost)
	}
	msg += " on your next turn."
	game.SendMessage() <- message.Reply(msg)

	return true
}
