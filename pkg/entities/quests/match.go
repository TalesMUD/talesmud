package quests

import "strings"

// KillTargetMatches reports whether a kill objective should advance for this NPC.
// Matching is by templateId, NPC id (including clone prefixes), or display name.
// Name matching covers the common authoring mismatch of targetId ENM0008 vs
// a spawned "Sewer Rat" whose runtime id is a UUID or clone id.
func KillTargetMatches(obj Objective, templateID, npcID, npcName string) bool {
	if obj.Type != ObjectiveKill {
		return false
	}
	target := strings.TrimSpace(obj.TargetID)
	if target == "" {
		return false
	}
	for _, candidate := range killIDCandidates(templateID, npcID) {
		if candidate == target {
			return true
		}
	}
	name := strings.TrimSpace(npcName)
	if name == "" {
		return false
	}
	if strings.EqualFold(name, target) {
		return true
	}
	if tn := strings.TrimSpace(obj.TargetName); tn != "" && strings.EqualFold(name, tn) {
		return true
	}
	return false
}

// RoomIDMatches is true when the player is in targetID or a private clone of it
// (clone ids look like "R0215~a1b2c3d4").
func RoomIDMatches(targetID, roomID string) bool {
	targetID = strings.TrimSpace(targetID)
	roomID = strings.TrimSpace(roomID)
	if targetID == "" || roomID == "" {
		return false
	}
	if roomID == targetID {
		return true
	}
	if i := strings.LastIndex(roomID, "~"); i > 0 && roomID[:i] == targetID {
		return true
	}
	return false
}

// NPCTemplateID prefers an explicit template id, otherwise the uncloned NPC id.
func NPCTemplateID(templateID, npcID string) string {
	if id := strings.TrimSpace(templateID); id != "" {
		return id
	}
	id := strings.TrimSpace(npcID)
	if i := strings.Index(id, "~"); i > 0 {
		return id[:i]
	}
	return id
}

func killIDCandidates(templateID, npcID string) []string {
	seen := map[string]bool{}
	var out []string
	add := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			return
		}
		seen[id] = true
		out = append(out, id)
	}
	add(templateID)
	add(npcID)
	if i := strings.Index(npcID, "~"); i > 0 {
		add(npcID[:i])
	}
	return out
}
