package portraits

import (
	"path"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
)

const PublicPath = "/api/portraits"

// FileName is the on-disk / URL filename for an NPC portrait.
// Spawned instances use their template ID so every Sample Wolf shares one face.
func FileName(id, templateID string) string {
	key := strings.TrimSpace(templateID)
	if key == "" {
		key = strings.TrimSpace(id)
	}
	if i := strings.LastIndex(key, "~"); i > 0 {
		key = key[:i]
	}
	if key == "" {
		return ""
	}
	return key + ".png"
}

// URL is the guest-public portrait URL (no Auth0).
func URL(id, templateID string) string {
	name := FileName(id, templateID)
	if name == "" {
		return ""
	}
	return path.Join(PublicPath, name)
}

// ForNPC returns the portrait URL for a runtime NPC.
func ForNPC(n *npc.NPC) string {
	if n == nil || n.Entity == nil {
		return ""
	}
	return URL(n.ID, n.TemplateID)
}

// ForPlayer returns the shared portrait for a supported race/class pair.
func ForPlayer(ch *characters.Character) string {
	if ch == nil {
		return ""
	}
	race := strings.ToLower(strings.TrimSpace(ch.Race.ID))
	class := strings.ToLower(strings.TrimSpace(ch.Class.ID))
	if race == "elve" || race == "elves" {
		race = "elf"
	}
	if class == "wizard" {
		class = "mage"
	}
	if class == "hunter" {
		class = "ranger"
	}
	switch race {
	case "human", "dwarf", "elf":
	default:
		return ""
	}
	switch class {
	case "warrior", "rogue", "mage", "ranger", "cleric", "druid":
	default:
		return ""
	}
	return path.Join(PublicPath, "player-"+race+"-"+class+".png")
}
