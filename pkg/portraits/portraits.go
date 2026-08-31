package portraits

import (
	"path"
	"strings"

	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
)

const PublicPath = "/api/portraits"

// FileName is the on-disk / URL filename for an NPC portrait.
// Spawned instances use their template ID so every Meadow Wolf shares one face.
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
