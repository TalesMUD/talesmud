package worldmap

import (
	"sort"

	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

type edge struct {
	from, to, dir string
	hidden        bool
}

type World struct {
	rooms map[string]*placedRoom
	edges []edge
}

// Compile builds a stable atlas of the whole world from room exits and
// optional coords. Positions do not depend on who has explored; Reveal
// applies fog of war on top.
func Compile(rs []*rooms.Room) *World {
	w := &World{
		rooms: make(map[string]*placedRoom, len(rs)),
		edges: make([]edge, 0, len(rs)*2),
	}
	ids := make([]string, 0, len(rs))
	src := make(map[string]*rooms.Room, len(rs))
	for _, r := range rs {
		if r == nil || r.Entity == nil || r.ID == "" {
			continue
		}
		src[r.ID] = r
		ids = append(ids, r.ID)
		pr := &placedRoom{
			id:       r.ID,
			name:     r.Name,
			area:     r.Area,
			areaName: displayArea(r.Area),
			tags:     append([]string(nil), r.Tags...),
			canBind:  r.CanBind,
			biome:    inferBiome(r.Area, r.Tags),
			kind:     inferKind(r.Tags, r.CanBind, inferBiome(r.Area, r.Tags)),
			landmark: isLandmark(r.Tags, r.CanBind),
		}
		w.rooms[r.ID] = pr
		if r.Exits == nil {
			continue
		}
		for _, ex := range *r.Exits {
			if ex.Target == "" {
				continue
			}
			w.edges = append(w.edges, edge{
				from: r.ID, to: ex.Target, dir: normalizeDir(ex.Name), hidden: ex.Hidden,
			})
		}
	}
	sort.Strings(ids)

	assignZ(w, src, ids)
	placeXY(w, src, ids)
	separateAreas(w)
	return w
}

func assignZ(w *World, src map[string]*rooms.Room, ids []string) {
	known := map[string]bool{}
	for _, id := range ids {
		r := src[id]
		if r.Coords != nil {
			w.rooms[id].z = int(r.Coords.Z)
			known[id] = true
		}
	}
	if len(known) == 0 {
		for _, id := range ids {
			r := src[id]
			if hasTag(r.Tags, "outdoor") && !hasTag(r.Tags, "underground") {
				w.rooms[id].z = 0
				known[id] = true
			}
		}
	}
	if len(known) == 0 {
		seed := pickSeed(src, ids)
		if seed != "" {
			w.rooms[seed].z = 0
			known[seed] = true
		}
	}

	changed := true
	for changed {
		changed = false
		for _, e := range w.edges {
			off, ok := offsetFor(e.dir)
			if !ok || off.z == 0 {
				continue
			}
			if known[e.from] && !known[e.to] && w.rooms[e.to] != nil {
				w.rooms[e.to].z = w.rooms[e.from].z + off.z
				known[e.to] = true
				changed = true
			}
			if known[e.to] && !known[e.from] && w.rooms[e.from] != nil {
				w.rooms[e.from].z = w.rooms[e.to].z - off.z
				known[e.from] = true
				changed = true
			}
		}
	}

	for _, id := range ids {
		if known[id] {
			continue
		}
		if hasTag(src[id].Tags, "underground") {
			w.rooms[id].z = -1
		} else {
			w.rooms[id].z = 0
		}
	}
}

type cell struct{ z, x, y int }

func placeXY(w *World, src map[string]*rooms.Room, ids []string) {
	occupied := map[cell]string{}
	placed := map[string]bool{}

	occupy := func(id string) {
		pr := w.rooms[id]
		occupied[cell{pr.z, pr.x, pr.y}] = id
		placed[id] = true
	}

	placeAt := func(id string, x, y, z int) {
		pr := w.rooms[id]
		if _, taken := occupied[cell{z, x, y}]; taken {
			x, y = spiralEmpty(occupied, z, x, y)
		}
		pr.x, pr.y, pr.z = x, y, z
		occupy(id)
	}

	for _, id := range ids {
		r := src[id]
		if r.Coords == nil {
			continue
		}
		placeAt(id, int(r.Coords.X), int(r.Coords.Y), w.rooms[id].z)
	}

	seed := pickSeed(src, ids)
	if seed != "" && !placed[seed] {
		placeAt(seed, 0, 0, w.rooms[seed].z)
	}

	queue := make([]string, 0, len(ids))
	for id := range placed {
		queue = append(queue, id)
	}
	sort.Strings(queue)

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		from := w.rooms[id]
		for _, e := range w.edges {
			if e.from != id {
				continue
			}
			dest := w.rooms[e.to]
			if dest == nil || placed[e.to] {
				continue
			}
			off, ok := offsetFor(e.dir)
			if !ok {
				continue
			}
			placeAt(e.to, from.x+off.x, from.y+off.y, dest.z)
			queue = append(queue, e.to)
		}
		// Reverse compass: if someone points at us, they sit in the opposite cell.
		for _, e := range w.edges {
			if e.to != id {
				continue
			}
			srcRoom := w.rooms[e.from]
			if srcRoom == nil || placed[e.from] {
				continue
			}
			off, ok := offsetFor(e.dir)
			if !ok {
				continue
			}
			placeAt(e.from, from.x-off.x, from.y-off.y, srcRoom.z)
			queue = append(queue, e.from)
		}
	}

	// Named / portal exits: park the destination next to the source.
	for _, e := range w.edges {
		if placed[e.to] || w.rooms[e.to] == nil || !placed[e.from] {
			continue
		}
		if _, ok := offsetFor(e.dir); ok {
			continue
		}
		from := w.rooms[e.from]
		placeAt(e.to, from.x+1, from.y, w.rooms[e.to].z)
	}

	for _, id := range ids {
		if placed[id] {
			continue
		}
		placeAt(id, 0, 0, w.rooms[id].z)
	}
}

func spiralEmpty(occupied map[cell]string, z, x, y int) (int, int) {
	if _, taken := occupied[cell{z, x, y}]; !taken {
		return x, y
	}
	for r := 1; r <= 12; r++ {
		for dx := -r; dx <= r; dx++ {
			for dy := -r; dy <= r; dy++ {
				if abs(dx) != r && abs(dy) != r {
					continue
				}
				nx, ny := x+dx, y+dy
				if _, taken := occupied[cell{z, nx, ny}]; !taken {
					return nx, ny
				}
			}
		}
	}
	return x + 13, y
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func pickSeed(src map[string]*rooms.Room, ids []string) string {
	for _, id := range ids {
		if hasTag(src[id].Tags, "starting_room") {
			return id
		}
	}
	if _, ok := src["R0001"]; ok {
		return "R0001"
	}
	if len(ids) == 0 {
		return ""
	}
	return ids[0]
}

func separateAreas(w *World) {
	type bbox struct{ minX, minY, maxX, maxY int }
	type areaKey struct {
		area string
		z    int
	}

	groups := map[areaKey][]string{}
	for id, pr := range w.rooms {
		k := areaKey{pr.area, pr.z}
		groups[k] = append(groups[k], id)
	}
	keys := make([]areaKey, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
		sort.Strings(groups[k])
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].z != keys[j].z {
			return keys[i].z < keys[j].z
		}
		return keys[i].area < keys[j].area
	})

	boundsOf := func(ids []string) bbox {
		pr := w.rooms[ids[0]]
		b := bbox{pr.x, pr.y, pr.x, pr.y}
		for _, id := range ids[1:] {
			p := w.rooms[id]
			if p.x < b.minX {
				b.minX = p.x
			}
			if p.y < b.minY {
				b.minY = p.y
			}
			if p.x > b.maxX {
				b.maxX = p.x
			}
			if p.y > b.maxY {
				b.maxY = p.y
			}
		}
		return b
	}
	overlaps := func(a, b bbox, pad int) bool {
		return a.minX <= b.maxX+pad && a.maxX+pad >= b.minX &&
			a.minY <= b.maxY+pad && a.maxY+pad >= b.minY
	}

	settled := make([]areaKey, 0, len(keys))
	for _, k := range keys {
		ids := groups[k]
		if k.area == "" || len(ids) == 0 {
			settled = append(settled, k)
			continue
		}
		b := boundsOf(ids)
		dx, dy := 0, 0
		for _, prev := range settled {
			if prev.z != k.z {
				continue
			}
			pb := boundsOf(groups[prev])
			shifted := bbox{b.minX + dx, b.minY + dy, b.maxX + dx, b.maxY + dy}
			guard := 0
			for overlaps(shifted, pb, 1) && guard < 20 {
				cx := (shifted.minX + shifted.maxX) / 2
				cy := (shifted.minY + shifted.maxY) / 2
				pcx := (pb.minX + pb.maxX) / 2
				pcy := (pb.minY + pb.maxY) / 2
				ox := cx - pcx
				oy := cy - pcy
				if ox == 0 && oy == 0 {
					ox = 1
				}
				if abs(ox) >= abs(oy) {
					if ox >= 0 {
						dx++
					} else {
						dx--
					}
				} else {
					if oy >= 0 {
						dy++
					} else {
						dy--
					}
				}
				shifted = bbox{b.minX + dx, b.minY + dy, b.maxX + dx, b.maxY + dy}
				guard++
			}
		}
		if dx != 0 || dy != 0 {
			for _, id := range ids {
				w.rooms[id].x += dx
				w.rooms[id].y += dy
			}
		}
		settled = append(settled, k)
	}
}
