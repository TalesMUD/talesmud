package worldmap

import (
	"sort"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

// undergroundRoom uses explicit depth and below-ground context. An outdoor
// dungeon entrance remains surface ground; a hillside's positive Z is elevation.
func undergroundRoom(r *rooms.Room) bool {
	if r.Coords != nil && r.Coords.Z < 0 {
		return true
	}
	if hasTag(r.Tags, "underground") {
		return true
	}
	if hasTag(r.Tags, "outdoor") {
		return false
	}
	for _, tag := range []string{"dungeon", "cave", "sewer", "cellar", "crypt", "mine", "tunnel", "undercity", "catacomb"} {
		if hasTag(r.Tags, tag) {
			return true
		}
	}
	for _, word := range []string{"cellar", "crypt", "sewer", "catacombs", "burrow"} {
		if containsTerrainWord(r.Name, word) {
			return true
		}
	}
	return false
}
func containsWord(text, word string) bool { return strings.Contains(" "+text+" ", " "+word+" ") }

func containsTerrainWord(text, word string) bool {
	return containsWord(terrainText(text), terrainText(word))
}

func assignSurface(w *World, src map[string]*rooms.Room, ids []string) {
	for _, id := range ids {
		p, r := w.rooms[id], src[id]
		p.layer = layerID(p.z)
		switch {
		case undergroundRoom(r):
			p.layer = "lower"
			p.role = "underground"
		case hasTag(r.Tags, "outdoor"):
			p.layer = "overworld"
			p.role = "surface"
		case hasTag(r.Tags, "indoor") || p.terrain == "interior":
			p.layer = "overworld"
			p.role = "interior"
		default:
			p.role = "surface"
		}
		p.town = layoutMapping.Zones[p.area].Town || hasTag(p.tags, "village") || hasTag(p.tags, "town")
	}
	// Anchor interiors through the room graph, preferring the nearest outdoor
	// room. Lexical tie breaks keep identical snapshots stable across requests.
	adj := map[string][]string{}
	for _, e := range w.edges {
		a, b := w.rooms[e.from], w.rooms[e.to]
		if a == nil || b == nil || a.area != b.area {
			continue
		}
		adj[e.from] = append(adj[e.from], e.to)
		adj[e.to] = append(adj[e.to], e.from)
	}
	for id := range adj {
		sort.Strings(adj[id])
	}
	for _, id := range ids {
		p := w.rooms[id]
		if p.role != "interior" {
			continue
		}
		seen := map[string]bool{id: true}
		queue := []string{id}
		for len(queue) > 0 && p.surfaceID == "" {
			at := queue[0]
			queue = queue[1:]
			for _, next := range adj[at] {
				if seen[next] {
					continue
				}
				seen[next] = true
				q := w.rooms[next]
				if q.layer != "overworld" {
					continue
				}
				if q.role == "surface" {
					p.surfaceID = next
					break
				}
				queue = append(queue, next)
			}
		}
		// Isolated above-ground buildings remain selectable as one surface group.
		if p.surfaceID == "" {
			for _, other := range ids {
				q := w.rooms[other]
				if q.area == p.area && q.role == "surface" && q.layer == "overworld" {
					p.surfaceID = other
					break
				}
			}
		}
		if p.surfaceID == "" {
			p.surfaceID = id
		}
	}
	for _, e := range w.edges {
		a, b := w.rooms[e.from], w.rooms[e.to]
		if a == nil || b == nil || a.layer != "overworld" || b.layer != "lower" {
			continue
		}
		a.entrances = append(a.entrances, e.to)
	}
	for _, id := range ids {
		sort.Strings(w.rooms[id].entrances)
	}
}
