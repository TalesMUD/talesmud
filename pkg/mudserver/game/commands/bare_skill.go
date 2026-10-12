package commands

import (
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/skills"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// BareSkill handles a line that is exactly the name of one of the character's
// equipped skills, optionally followed by a target ("aimed shot rat").
// In a fight it casts the skill. Out of a fight it explains why nothing
// happened instead of saying the skill name to the room.
// It returns false when the line is not a skill name.
func BareSkill(game def.GameCtrl, message *messages.Message) bool {
	if game == nil || message == nil || message.Character == nil {
		return false
	}
	text := strings.ToLower(strings.Join(strings.Fields(message.Data), " "))
	if text == "" {
		return false
	}
	matched := ""
	for _, sid := range message.Character.EquippedSkills {
		skill := skills.SkillByID(sid)
		if skill == nil || skill.Name == "" {
			continue
		}
		name := strings.ToLower(skill.Name)
		if text == name || strings.HasPrefix(text, name+" ") {
			if len(name) > len(matched) {
				matched = skill.Name
			}
		}
	}
	if matched == "" {
		return false
	}
	if !isInActiveCombat(game, message.Character, game.GetCombatEngine()) {
		game.SendMessage() <- message.Reply("You're not in combat. " + matched + " can only be used in a fight.")
		return true
	}
	cast := *message
	cast.Data = "cast " + strings.TrimSpace(message.Data)
	return (&CastCommand{}).Execute(game, &cast)
}
