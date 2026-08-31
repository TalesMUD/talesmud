package instances

import (
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

// CollectGraph returns dest plus every room reachable from dest without
// walking back into hubID. That is the private cellar copy.
func CollectGraph(all map[string]*rooms.Room, hubID, destID string) []string {
	if destID == "" || destID == hubID {
		return nil
	}
	seen := map[string]bool{hubID: true}
	queue := []string{destID}
	var out []string
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
		room := all[id]
		if room == nil || room.Exits == nil {
			continue
		}
		for _, ex := range *room.Exits {
			if ex.Target == "" || seen[ex.Target] {
				continue
			}
			queue = append(queue, ex.Target)
		}
	}
	return out
}

// IsInstanceEntrance reports whether taking this exit should spawn a private copy.
func IsInstanceEntrance(ex rooms.Exit) bool {
	if ex.Instance {
		return true
	}
	return strings.EqualFold(string(ex.Type), "instance")
}

// CloneID builds a per-instance room id.
func CloneID(templateID, instanceID string) string {
	return templateID + "~" + instanceID
}

// TemplateIDFromClone reverses CloneID. Uncloned ids are returned as-is.
func TemplateIDFromClone(roomID string) string {
	if i := strings.LastIndex(roomID, "~"); i > 0 {
		return roomID[:i]
	}
	return roomID
}
