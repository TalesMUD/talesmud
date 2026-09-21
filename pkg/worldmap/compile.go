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
// optional coords. Authored Coords are pinned. Remaining rooms layout per
// area using compass exits, then area clusters are packed with a gap so
// zones do not bleed. Positions do not depend on who has explored; Reveal
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
	for _, id := range ids {
		if known[id] {
			continue
		}
		r := src[id]
		ug := hasTag(r.Tags, "underground")
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
	occupied := map[cell]string{}
	placed := map[string]bool{}

	placeAt := func(id string, x, y, z int, occ map[cell]string) {
		pr := w.rooms[id]
		if _, taken := occ[cell{z, x, y}]; taken {
			x, y = spiralEmpty(occ, z, x, y)
		}
		pr.x, pr.y, pr.z = x, y, z
		occ[cell{z, x, y}] = id
		placed[id] = true
	}

	// Authored Coords win. Later graph placement must not shove these.
	for _, id := range ids {
		r := src[id]
		if r.Coords == nil {
			continue
		}
		pr := w.rooms[id]
		pr.locked = true
		placeAt(id, int(r.Coords.X), int(r.Coords.Y), pr.z, occupied)
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
		occ := occupied
		if !anchored {
			// Isolated origin so this zone does not spiral into another zone.
			occ = map[cell]string{}
			seed := pickSeedIn(src, gids)
			placeAt(seed, 0, 0, w.rooms[seed].z, occ)
		}
		walkCluster(w, gids, placed, occ)
	}

	for _, id := range ids {
		if placed[id] {
			continue
		}
		placeAt(id, 0, 0, w.rooms[id].z, occupied)
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
		if _, taken := occ[cell{z, x, y}]; taken {
			x, y = spiralEmpty(occ, z, x, y)
		}
		pr.x, pr.y, pr.z = x, y, z
		occ[cell{z, x, y}] = id
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

const areaGap = 4

func translateGroup(w *World, ids []string, dx, dy int) {
	if dx == 0 && dy == 0 {
		return
	}
	for _, id := range ids {
		w.rooms[id].x += dx
		w.rooms[id].y += dy
	}
}

func groupAnchored(w *World, ids []string) bool {
	for _, id := range ids {
		if w.rooms[id].locked {
			return true
		}
	}
	return false
}

func separateAreas(w *World) {
	groups := map[string][]string{}
	for id, pr := range w.rooms {
		groups[pr.area] = append(groups[pr.area], id)
	}
	for area := range groups {
		sort.Strings(groups[area])
	}

	placedArea := map[string]bool{}
	for area, ids := range groups {
		if groupAnchored(w, ids) {
			placedArea[area] = true
		}
	}

	type link struct {
		fromArea, toArea, fromID, toID string
		off                            vec
	}
	var links []link
	for _, e := range w.edges {
		off, ok := offsetFor(e.dir)
		if !ok || off.z != 0 {
			continue
		}
		from, to := w.rooms[e.from], w.rooms[e.to]
		if from == nil || to == nil || from.area == to.area {
			continue
		}
		links = append(links, link{from.area, to.area, e.from, e.to, off})
	}

	seed := pickAreaSeed(w, groups)
	if seed != "" {
		placedArea[seed] = true
	}
	queue := make([]string, 0, len(placedArea))
	for area := range placedArea {
		queue = append(queue, area)
	}
	sort.Strings(queue)

	for len(queue) > 0 {
		area := queue[0]
		queue = queue[1:]
		for _, ln := range links {
			var destArea string
			var fromID, toID string
			var off vec
			if ln.fromArea == area && !placedArea[ln.toArea] {
				destArea, fromID, toID, off = ln.toArea, ln.fromID, ln.toID, ln.off
			} else if ln.toArea == area && !placedArea[ln.fromArea] {
				destArea, fromID, toID, off = ln.fromArea, ln.toID, ln.fromID, vec{-ln.off.x, -ln.off.y, 0}
			} else {
				continue
			}
			alignAreaByExit(w, groups, fromID, toID, off)
			pushAreaOut(w, groups, destArea, placedArea, off)
			placedArea[destArea] = true
			queue = append(queue, destArea)
		}
	}

	leftover := make([]string, 0)
	for area := range groups {
		if !placedArea[area] {
			leftover = append(leftover, area)
		}
	}
	sort.Strings(leftover)
	settled := make([]bbox, 0, len(groups))
	for area := range placedArea {
		settled = append(settled, boundsOf(w, groups[area]))
	}
	cursorX, cursorY, rowH := 0, 0, 0
	if len(settled) > 0 {
		maxX, minY := settled[0].maxX, settled[0].minY
		for _, b := range settled[1:] {
			if b.maxX > maxX {
				maxX = b.maxX
			}
			if b.minY < minY {
				minY = b.minY
			}
		}
		cursorX = maxX + areaGap + 1
		cursorY = minY
	}
	for _, area := range leftover {
		ids := groups[area]
		b := boundsOf(w, ids)
		wdt, hgt := b.width(), b.height()
		if cursorX > 0 && cursorX+wdt > 18 && rowH > 0 {
			cursorY += rowH + areaGap
			cursorX = 0
			rowH = 0
		}
		dx := cursorX - b.minX
		dy := cursorY - b.minY
		trial := b.shifted(dx, dy)
		for _, prev := range settled {
			if trial.overlaps(prev, areaGap) {
				dx = prev.maxX + areaGap + 1 - b.minX
				trial = b.shifted(dx, dy)
			}
		}
		translateGroup(w, ids, dx, dy)
		nb := boundsOf(w, ids)
		settled = append(settled, nb)
		placedArea[area] = true
		cursorX = nb.maxX + areaGap + 1
		if hgt > rowH {
			rowH = hgt
		}
	}
}

func pickAreaSeed(w *World, groups map[string][]string) string {
	for _, pr := range w.rooms {
		if hasTag(pr.tags, "starting_room") && !hasTag(pr.tags, "underground") {
			return pr.area
		}
	}
	for _, pr := range w.rooms {
		if hasTag(pr.tags, "entry_point") && hasTag(pr.tags, "outdoor") {
			return pr.area
		}
	}
	if _, ok := groups["Z01_meadows_forest_path"]; ok {
		return "Z01_meadows_forest_path"
	}
	areas := make([]string, 0, len(groups))
	for a := range groups {
		areas = append(areas, a)
	}
	sort.Strings(areas)
	if len(areas) == 0 {
		return ""
	}
	return areas[0]
}

func alignAreaByExit(w *World, groups map[string][]string, fromID, toID string, off vec) {
	from, to := w.rooms[fromID], w.rooms[toID]
	if from == nil || to == nil {
		return
	}
	step := areaGap + 1
	tx := from.x + off.x*step
	ty := from.y + off.y*step
	translateGroup(w, groups[to.area], tx-to.x, ty-to.y)
}

func pushAreaOut(w *World, groups map[string][]string, area string, placed map[string]bool, off vec) {
	ids := groups[area]
	if len(ids) == 0 {
		return
	}
	if off.x == 0 && off.y == 0 {
		off = vec{0, -1, 0}
	}
	for guard := 0; guard < 48; guard++ {
		b := boundsOf(w, ids)
		hit := false
		for other := range placed {
			if other == area {
				continue
			}
			ob := boundsOf(w, groups[other])
			if b.overlaps(ob, areaGap) {
				translateGroup(w, ids, off.x*(areaGap+1), off.y*(areaGap+1))
				hit = true
				break
			}
		}
		if !hit {
			return
		}
	}
}
