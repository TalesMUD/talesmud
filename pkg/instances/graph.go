package instances

import (
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

// Marker separates a template or slot id from a per-instance copy.
// Authored room ids do not contain it. Live copies do: "<template>~<instance>".
const Marker = "~"

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

// CrossingIntoInstance is true when leaving a shared/hub room into an instance template.
func CrossingIntoInstance(from, dest *rooms.Room, ex rooms.Exit) bool {
	if dest == nil {
		return false
	}
	if from != nil && (from.IsInstanceTemplate() || IsCloneID(from.ID)) {
		return false
	}
	if IsInstanceEntrance(ex) {
		return true
	}
	return dest.IsInstanceTemplate()
}

// CloneID builds a per-instance room id.
func CloneID(templateID, instanceID string) string {
	return templateID + Marker + instanceID
}

// IsCloneID reports whether id is an instance copy.
// A tagged template without the marker is not a copy.
func IsCloneID(id string) bool {
	return strings.Contains(id, Marker)
}

// TemplateIDFromClone reverses CloneID. Uncloned ids are returned as-is.
func TemplateIDFromClone(roomID string) string {
	if i := strings.LastIndex(roomID, Marker); i > 0 {
		return roomID[:i]
	}
	return roomID
}

// InstanceID is the per-copy suffix after the last marker.
// It is empty when id is not a copy.
func InstanceID(id string) string {
	i := strings.LastIndex(id, Marker)
	if i <= 0 || i >= len(id)-1 {
		return ""
	}
	return id[i+1:]
}
