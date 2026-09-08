package worldmap

import (
	"sort"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/characters"
)

// Reveal filters a compiled world into the atlas a character may see:
// discovered rooms, fog neighbors through visible exits, and area hulls.
func Reveal(w *World, ch *characters.Character) PlayerMap {
	out := PlayerMap{
		Layers:  []Layer{},
		Places:  []Place{},
		Paths:   []Path{},
		Regions: []Region{},
	}
	if w == nil {
		return out
	}
	discovered := map[string]bool{}
	if ch != nil {
		out.CharacterID = ch.ID
		out.CurrentRoomID = ch.CurrentRoomID
		for id, ok := range ch.DiscoveredRooms {
			if ok {
				discovered[id] = true
			}
		}
		if ch.CurrentRoomID != "" {
			discovered[ch.CurrentRoomID] = true
			if i := strings.Index(ch.CurrentRoomID, "~"); i > 0 {
				discovered[ch.CurrentRoomID[:i]] = true
			}
		}
	}

	include := map[string]bool{}
	fog := map[string]bool{}
	for id := range discovered {
		if w.rooms[id] == nil {
			continue
		}
		include[id] = true
	}

	type pathKey struct{ a, b, dir string }
	seenPath := map[pathKey]bool{}

	for _, e := range w.edges {
		if !discovered[e.from] || w.rooms[e.from] == nil {
			continue
		}
		if e.hidden && (ch == nil || !ch.HasRevealedExit(e.from, e.dir)) {
			continue
		}
		dest := w.rooms[e.to]
		if dest == nil {
			continue
		}
		if !discovered[e.to] {
			fog[e.to] = true
			include[e.to] = true
		}
		pk := pathKey{e.from, e.to, e.dir}
		if seenPath[pk] {
			continue
		}
		seenPath[pk] = true
		from := w.rooms[e.from]
		out.Paths = append(out.Paths, Path{
			From:   e.from,
			To:     e.to,
			Dir:    e.dir,
			Kind:   pathKind(e.dir, from.tags, dest.tags, e.hidden),
			Layer:  layerID(from.z),
			Hidden: e.hidden,
		})
	}

	layerSet := map[string]Layer{}
	regionRooms := map[string][]string{} // area|layer -> ids

	ids := make([]string, 0, len(include))
	for id := range include {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		pr := w.rooms[id]
		lid := layerID(pr.z)
		layerSet[lid] = Layer{ID: lid, Name: layerName(lid), Kind: layerKind(lid)}
		isFog := fog[id] && !discovered[id]
		p := Place{
			ID:         pr.id,
			Area:       pr.area,
			AreaName:   pr.areaName,
			Layer:      lid,
			X:          float64(pr.x),
			Y:          float64(pr.y),
			Z:          pr.z,
			Biome:      pr.biome,
			Kind:       pr.kind,
			Landmark:   pr.landmark && !isFog,
			Discovered: !isFog,
			Current:    ch != nil && placeIsCurrent(id, ch.CurrentRoomID),
			Tags:       pr.tags,
		}
		if isFog {
			p.Name = ""
			p.Kind = "uncharted"
			p.CanTravel = false
			p.Danger = "uncharted"
			p.Summary = "Fog of war. Walk closer to chart this ground."
		} else {
			p.Name = pr.name
			p.CanTravel = ch != nil && !placeIsCurrent(id, ch.CurrentRoomID)
			p.Danger = inferDanger(pr.tags, pr.kind)
			p.Summary = inferSummary(pr.kind, p.Danger)
			p.Exits = collectPlaceExits(w, ch, id, discovered)
			rk := pr.area + "|" + lid
			regionRooms[rk] = append(regionRooms[rk], id)
		}
		out.Places = append(out.Places, p)
		if p.Current {
			out.CurrentLayer = lid
		}
	}

	if out.CurrentLayer == "" {
		out.CurrentLayer = "overworld"
	}

	layerIDs := make([]string, 0, len(layerSet))
	for id := range layerSet {
		layerIDs = append(layerIDs, id)
	}
	sort.Slice(layerIDs, func(i, j int) bool {
		order := map[string]int{"lower": 0, "overworld": 1, "upper": 2}
		return order[layerIDs[i]] < order[layerIDs[j]]
	})
	for _, id := range layerIDs {
		out.Layers = append(out.Layers, layerSet[id])
	}

	regionKeys := make([]string, 0, len(regionRooms))
	for k := range regionRooms {
		regionKeys = append(regionKeys, k)
	}
	sort.Strings(regionKeys)
	for _, rk := range regionKeys {
		rids := regionRooms[rk]
		if len(rids) == 0 {
			continue
		}
		pr := w.rooms[rids[0]]
		pts := make([][2]float64, 0, len(rids))
		for _, id := range rids {
			p := w.rooms[id]
			pts = append(pts, [2]float64{float64(p.x), float64(p.y)})
		}
		out.Regions = append(out.Regions, Region{
			ID:     pr.area + ":" + layerID(pr.z),
			Name:   pr.areaName,
			Layer:  layerID(pr.z),
			Biome:  majorityBiome(w, rids),
			Hull:   convexHull(pts),
			Places: rids,
		})
	}

	return out
}

// placeIsCurrent matches template room ids to private instance clones
// (e.g. R0215~charId → R0215) so exactly one atlas place is marked current.
func placeIsCurrent(placeID, currentRoomID string) bool {
	if currentRoomID == "" {
		return false
	}
	if placeID == currentRoomID {
		return true
	}
	if i := strings.Index(currentRoomID, "~"); i > 0 {
		return placeID == currentRoomID[:i]
	}
	return false
}

func collectPlaceExits(w *World, ch *characters.Character, fromID string, discovered map[string]bool) []PlaceExit {
	out := make([]PlaceExit, 0, 4)
	for _, e := range w.edges {
		if e.from != fromID {
			continue
		}
		if e.hidden && (ch == nil || !ch.HasRevealedExit(e.from, e.dir)) {
			continue
		}
		dest := w.rooms[e.to]
		if dest == nil {
			continue
		}
		px := PlaceExit{
			Dir:      e.dir,
			To:       e.to,
			Hidden:   e.hidden,
			Vertical: isVertical(e.dir),
		}
		if discovered[e.to] {
			px.ToName = dest.name
		}
		out = append(out, px)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Dir == out[j].Dir {
			return out[i].To < out[j].To
		}
		return out[i].Dir < out[j].Dir
	})
	return out
}

func inferDanger(tags []string, kind string) string {
	if hasTag(tags, "safe") || hasTag(tags, "inn") || hasTag(tags, "shrine") || hasTag(tags, "starting_room") {
		return "safe"
	}
	if kind == "water" || hasTag(tags, "water") {
		return "hazard"
	}
	if kind == "dungeon" || hasTag(tags, "underground") || hasTag(tags, "dungeon") || hasTag(tags, "cave") {
		return "hostile"
	}
	return "low"
}

func inferSummary(kind, danger string) string {
	switch danger {
	case "safe":
		return "A haven. Bind, rest, and resupply."
	case "hostile":
		return "Dangerous ground. Watch the dark."
	case "hazard":
		return "Treacherous footing."
	}
	if kind == "settlement" {
		return "Town streets and hearths."
	}
	return "Open country. Chart as you walk."
}

// AttachResidents fills discovered places with best-effort NPC/enemy lists.
func AttachResidents(atlas *PlayerMap, byRoom map[string][]PlaceResident) {
	if atlas == nil || len(byRoom) == 0 {
		return
	}
	for i := range atlas.Places {
		if !atlas.Places[i].Discovered {
			continue
		}
		if rs := byRoom[atlas.Places[i].ID]; len(rs) > 0 {
			atlas.Places[i].Residents = rs
		}
	}
}

func majorityBiome(w *World, ids []string) string {
	counts := map[string]int{}
	best, n := "wild", 0
	for _, id := range ids {
		b := w.rooms[id].biome
		counts[b]++
		if counts[b] > n || (counts[b] == n && b < best) {
			best, n = b, counts[b]
		}
	}
	return best
}
