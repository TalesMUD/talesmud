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
	rooms     map[string]*placedRoom
	edges     []edge
	landscape []LandCell
}

// Compile builds a stable atlas of the whole world from room exits and
// optional area-local coords. Compact zone offsets preserve local geometry on
// one decorative continent; interiors project onto exterior anchors. Positions
// do not depend on who has explored; Reveal applies fog of war on top.
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
		terrain, _ := ClassifyTerrain(r)
		features, seed := ClassifyArt(r)
		pr := &placedRoom{
			terrain:          terrain,
			mapFeatures:      features,
			artSeed:          seed,
			undergroundStyle: undergroundStyle(r),
			id:               r.ID,
			name:             r.Name,
			area:             r.Area,
			areaName:         displayArea(r.Area),
			tags:             append([]string(nil), r.Tags...),
			canBind:          r.CanBind,
			biome:            inferBiome(r.Area, r.Tags),
			kind:             inferKind(r.Tags, r.CanBind, inferBiome(r.Area, r.Tags)),
			landmark:         isLandmark(r.Tags, r.CanBind),
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
	sort.Slice(w.edges, func(i, j int) bool {
		a, b := w.edges[i], w.edges[j]
		if a.from != b.from {
			return a.from < b.from
		}
		if a.dir != b.dir {
			return a.dir < b.dir
		}
		return a.to < b.to
	})

	assignZ(w, src, ids)
	assignSurface(w, src, ids)
	placeXY(w, src, ids)
	packContinent(w, ids)
	for _, id := range ids {
		p := w.rooms[id]
		if p.role == "interior" && p.surfaceID != id {
			if q := w.rooms[p.surfaceID]; q != nil {
				p.x, p.y = q.x, q.y
			}
		}
	}
	w.landscape = buildLandscape(w, ids)
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
	for _, id := range ids {
		if known[id] {
			continue
		}
		r := src[id]
		ug := undergroundRoom(r)
		if ug {
			w.rooms[id].z = -1
			known[id] = true
			continue
		}
		if hasTag(r.Tags, "indoor") {
			w.rooms[id].z = 0
			known[id] = true
			continue
		}
		if hasTag(r.Tags, "outdoor") && !ug {
			w.rooms[id].z = 0
			known[id] = true
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

func sameCluster(a, b *placedRoom) bool {
	if a == nil || b == nil {
		return false
	}
	return a.area == b.area
}

func placeXY(w *World, src map[string]*rooms.Room, ids []string) {
	occupiedByArea := map[string]map[cell]string{}
	placed := map[string]bool{}

	placeAt := func(id string, x, y, z int, occ map[cell]string) {
		pr := w.rooms[id]
		plane := layoutPlane(pr)
		if _, taken := occ[cell{plane, x, y}]; taken {
			x, y = spiralEmpty(occ, plane, x, y)
		}
		pr.x, pr.y, pr.z = x, y, z
		occ[cell{plane, x, y}] = id
		placed[id] = true
	}

	// Authored area-local Coords take precedence over inferred graph positions.
	for _, id := range ids {
		r := src[id]
		if r.Coords == nil {
			continue
		}
		pr := w.rooms[id]
		if occupiedByArea[pr.area] == nil {
			occupiedByArea[pr.area] = map[cell]string{}
		}
		placeAt(id, int(r.Coords.X), int(r.Coords.Y), pr.z, occupiedByArea[pr.area])
	}

	groups := map[string][]string{}
	for _, id := range ids {
		groups[w.rooms[id].area] = append(groups[w.rooms[id].area], id)
	}
	areas := make([]string, 0, len(groups))
	for area := range groups {
		areas = append(areas, area)
		sort.Strings(groups[area])
	}
	sort.Strings(areas)

	for _, area := range areas {
		gids := groups[area]
		anchored := false
		for _, id := range gids {
			if placed[id] {
				anchored = true
				break
			}
		}
		occ := occupiedByArea[area]
		if occ == nil {
			occ = map[cell]string{}
			occupiedByArea[area] = occ
		}
		if !anchored {
			// Each area has its own occupancy map and local origin.
			seed := pickSeedIn(src, gids)
			placeAt(seed, 0, 0, w.rooms[seed].z, occ)
		}
		walkCluster(w, gids, placed, occ)
	}

	for _, id := range ids {
		if placed[id] {
			continue
		}
		p := w.rooms[id]
		if occupiedByArea[p.area] == nil {
			occupiedByArea[p.area] = map[cell]string{}
		}
		placeAt(id, 0, 0, p.z, occupiedByArea[p.area])
	}
}

func walkCluster(w *World, gids []string, placed map[string]bool, occ map[cell]string) {
	inGroup := map[string]bool{}
	queue := make([]string, 0, len(gids))
	for _, id := range gids {
		inGroup[id] = true
		if placed[id] {
			queue = append(queue, id)
		}
	}
	sort.Strings(queue)

	placeAt := func(id string, x, y, z int) {
		pr := w.rooms[id]
		plane := layoutPlane(pr)
		if _, taken := occ[cell{plane, x, y}]; taken {
			x, y = spiralEmpty(occ, plane, x, y)
		}
		pr.x, pr.y, pr.z = x, y, z
		occ[cell{plane, x, y}] = id
		placed[id] = true
	}

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		from := w.rooms[id]
		for _, e := range w.edges {
			if e.from != id {
				continue
			}
			dest := w.rooms[e.to]
			if dest == nil || placed[e.to] || !inGroup[e.to] || !sameCluster(from, dest) {
				continue
			}
			off, ok := offsetFor(e.dir)
			if !ok {
				continue
			}
			placeAt(e.to, from.x+off.x, from.y+off.y, dest.z)
			queue = append(queue, e.to)
		}
		for _, e := range w.edges {
			if e.to != id {
				continue
			}
			srcRoom := w.rooms[e.from]
			if srcRoom == nil || placed[e.from] || !inGroup[e.from] || !sameCluster(from, srcRoom) {
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

	for _, e := range w.edges {
		if !inGroup[e.from] || !inGroup[e.to] {
			continue
		}
		if placed[e.to] || w.rooms[e.to] == nil || !placed[e.from] {
			continue
		}
		if _, ok := offsetFor(e.dir); ok {
			continue
		}
		if !sameCluster(w.rooms[e.from], w.rooms[e.to]) {
			continue
		}
		from := w.rooms[e.from]
		placeAt(e.to, from.x+1, from.y, w.rooms[e.to].z)
	}
}

func spiralEmpty(occupied map[cell]string, z, x, y int) (int, int) {
	if _, taken := occupied[cell{z, x, y}]; !taken {
		return x, y
	}
	for r := 1; r <= 32; r++ {
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
	return x + 33, y
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func pickSeed(src map[string]*rooms.Room, ids []string) string {
	return pickSeedIn(src, ids)
}

func pickSeedIn(src map[string]*rooms.Room, ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	for _, id := range ids {
		if hasTag(src[id].Tags, "starting_room") {
			return id
		}
	}
	for _, id := range ids {
		if hasTag(src[id].Tags, "entry_point") {
			return id
		}
	}
	for _, id := range ids {
		if id == "R0001" {
			return id
		}
	}
	return ids[0]
}

type bbox struct{ minX, minY, maxX, maxY int }

func boundsOf(w *World, ids []string) bbox {
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

func (b bbox) shifted(dx, dy int) bbox {
	return bbox{b.minX + dx, b.minY + dy, b.maxX + dx, b.maxY + dy}
}

func (b bbox) overlaps(o bbox, pad int) bool {
	return b.minX <= o.maxX+pad && b.maxX+pad >= o.minX &&
		b.minY <= o.maxY+pad && b.maxY+pad >= o.minY
}

func (b bbox) width() int  { return b.maxX - b.minX + 1 }
func (b bbox) height() int { return b.maxY - b.minY + 1 }

func translateGroup(w *World, ids []string, dx, dy int) {
	for _, id := range ids {
		p := w.rooms[id]
		p.x += dx
		p.y += dy
	}
}

// Interiors occupy a separate placement plane until attached to their surface.
func layoutPlane(p *placedRoom) int {
	if p.role == "interior" {
		return 2
	}
	if p.layer == "lower" {
		return -1
	}
	if p.layer == "upper" {
		return 1
	}
	return 0
}
