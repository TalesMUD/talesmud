package instances

import (
	"sort"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

// EntranceForTemplate is the shared hub for templateID's instance graph:
// a non-instance room whose exit leads into that graph. Clone rows and
// authored templates are not entrances. An unknown template returns "".
func EntranceForTemplate(all map[string]*rooms.Room, templateID string) string {
	if templateID == "" || IsCloneID(templateID) || all[templateID] == nil {
		return ""
	}
	ids := make([]string, 0, len(all))
	for id := range all {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, hubID := range ids {
		if hubID == templateID || IsCloneID(hubID) {
			continue
		}
		hub := all[hubID]
		if hub == nil || hub.Exits == nil || hub.IsInstanceTemplate() {
			continue
		}
		for _, ex := range *hub.Exits {
			target := ex.Target
			if target == "" || IsCloneID(target) || all[target] == nil {
				continue
			}
			dest := all[target]
			if !IsInstanceEntrance(ex) && !dest.IsInstanceTemplate() {
				continue
			}
			for _, id := range CollectGraph(all, hubID, target) {
				if id == templateID {
					return hubID
				}
			}
		}
	}
	return ""
}

// ReturnRoomForClone is a non-instance exit target saved on a copy, such as
// a procedural "out" exit. EntranceForTemplate wins when the template remains.
func ReturnRoomForClone(all map[string]*rooms.Room, cloneID string) string {
	suffix := InstanceID(cloneID)
	if suffix == "" {
		return ""
	}
	type cand struct {
		name   string
		target string
	}
	var cands []cand
	for id, room := range all {
		if room == nil || room.Exits == nil || !IsCloneID(id) || InstanceID(id) != suffix {
			continue
		}
		for _, ex := range *room.Exits {
			if ex.Target == "" || IsCloneID(ex.Target) || all[ex.Target] == nil {
				continue
			}
			if all[ex.Target].IsInstanceTemplate() {
				continue
			}
			cands = append(cands, cand{name: strings.ToLower(ex.Name), target: ex.Target})
		}
	}
	sort.Slice(cands, func(i, j int) bool {
		iOut := cands[i].name == "out"
		jOut := cands[j].name == "out"
		if iOut != jOut {
			return iOut
		}
		if cands[i].target != cands[j].target {
			return cands[i].target < cands[j].target
		}
		return cands[i].name < cands[j].name
	})
	if len(cands) == 0 {
		return ""
	}
	return cands[0].target
}

// RelocationDest chooses a real room for a character saved in a copy or in
// a room that is gone.
//
// Order: the template group's entrance, else the copy's return exit, else
// startID, else fallbackID. startID and fallbackID are ignored unless they
// name a non-copy room that is present in all.
func RelocationDest(all map[string]*rooms.Room, currentID, startID, fallbackID string) string {
	if IsCloneID(currentID) {
		if id := EntranceForTemplate(all, TemplateIDFromClone(currentID)); roomOK(all, id) {
			return id
		}
		if id := ReturnRoomForClone(all, currentID); roomOK(all, id) {
			return id
		}
	}
	if roomOK(all, startID) {
		return startID
	}
	if roomOK(all, fallbackID) {
		return fallbackID
	}
	return ""
}

func roomOK(all map[string]*rooms.Room, id string) bool {
	if id == "" || IsCloneID(id) {
		return false
	}
	return all[id] != nil
}
