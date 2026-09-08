package worldmap

import (
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
)

// ResidentsFromNPCs groups unique/template NPCs by the room they haunt.
func ResidentsFromNPCs(npcs []*npc.NPC) map[string][]PlaceResident {
	out := map[string][]PlaceResident{}
	seen := map[string]map[string]bool{}
	add := func(room, name, kind string) {
		if room == "" || name == "" {
			return
		}
		if seen[room] == nil {
			seen[room] = map[string]bool{}
		}
		if seen[room][name] {
			return
		}
		seen[room][name] = true
		out[room] = append(out[room], PlaceResident{Name: name, Kind: kind})
	}
	for _, n := range npcs {
		if n == nil || n.IsDead {
			continue
		}
		kind := "npc"
		if n.IsEnemy() {
			kind = "enemy"
		}
		if n.IsTemplate {
			add(n.SpawnRoomID, n.Name, kind)
			continue
		}
		room := n.CurrentRoomID
		if room == "" {
			room = n.SpawnRoomID
		}
		add(room, n.Name, kind)
	}
	return out
}
