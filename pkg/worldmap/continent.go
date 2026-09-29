package worldmap

import (
	"embed"
	"encoding/json"
	"math"
	"sort"
)

//go:embed map_layout.json
var layoutFS embed.FS

type zoneLayout struct {
	Center [2]int
	Ground string
	Town   bool
}
type layoutConfig struct {
	Gap          int
	CoastPadding int
	Zones        map[string]zoneLayout
}

var layoutMapping = func() layoutConfig {
	b, err := layoutFS.ReadFile("map_layout.json")
	if err != nil {
		panic(err)
	}
	var c layoutConfig
	if err = json.Unmarshal(b, &c); err != nil {
		panic(err)
	}
	if c.Gap < 1 || c.CoastPadding < 1 {
		panic("invalid map layout config")
	}
	return c
}()

func areaGroups(w *World, ids []string) map[string][]string {
	groups := map[string][]string{}
	for _, id := range ids {
		p := w.rooms[id]
		groups[p.area] = append(groups[p.area], id)
	}
	return groups
}
func surfaceBounds(w *World, ids []string) (bbox, string) {
	var surface, lower, upper []string
	for _, id := range ids {
		p := w.rooms[id]
		if p.role == "interior" && p.surfaceID != id {
			continue
		}
		switch p.layer {
		case "overworld":
			surface = append(surface, id)
		case "lower":
			lower = append(lower, id)
		default:
			upper = append(upper, id)
		}
	}
	if len(surface) > 0 {
		return boundsOf(w, surface), "overworld"
	}
	if len(lower) > 0 {
		return boundsOf(w, lower), "lower"
	}
	if len(upper) > 0 {
		return boundsOf(w, upper), "upper"
	}
	return boundsOf(w, ids), "overworld"
}

// packContinent translates area-local geometry, preserving its internal compass
// layout. Compact configured centers disambiguate graph cycles; unknown areas
// attach by their compass exits, then by a small shelf beside settled zones.
func packContinent(w *World, ids []string) {
	groups := areaGroups(w, ids)
	if len(groups) < 2 {
		return
	}
	areas := make([]string, 0, len(groups))
	for a := range groups {
		areas = append(areas, a)
	}
	sort.Slice(areas, func(i, j int) bool {
		a, aok := layoutMapping.Zones[areas[i]]
		b, bok := layoutMapping.Zones[areas[j]]
		if aok != bok {
			return aok
		}
		da, db := a.Center[0]*a.Center[0]+a.Center[1]*a.Center[1], b.Center[0]*b.Center[0]+b.Center[1]*b.Center[1]
		if da != db {
			return da < db
		}
		return areas[i] < areas[j]
	})
	type settledZone struct {
		area, layer string
		box         bbox
	}
	settled := []settledZone{}
	placed := map[string]bool{}
	for _, area := range areas {
		gids := groups[area]
		b, layer := surfaceBounds(w, gids)
		dx, dy := 0, 0
		if cfg, ok := layoutMapping.Zones[area]; ok {
			dx = cfg.Center[0] - (b.minX+b.maxX)/2
			dy = cfg.Center[1] - (b.minY+b.maxY)/2
		} else {
			linked := false
			for _, e := range w.edges {
				a, c := w.rooms[e.from], w.rooms[e.to]
				off, ok := offsetFor(e.dir)
				if a == nil || c == nil || !ok || off.z != 0 {
					continue
				}
				if c.area == area && placed[a.area] {
					dx = a.x + off.x*(layoutMapping.Gap+1) - c.x
					dy = a.y + off.y*(layoutMapping.Gap+1) - c.y
					linked = true
					break
				}
				if a.area == area && placed[c.area] {
					dx = c.x - off.x*(layoutMapping.Gap+1) - a.x
					dy = c.y - off.y*(layoutMapping.Gap+1) - a.y
					linked = true
					break
				}
			}
			if !linked && len(settled) > 0 {
				last := settled[len(settled)-1].box
				dx = last.maxX + layoutMapping.Gap + 1 - b.minX
				dy = last.minY - b.minY
			}
		}
		collision := func(tx, ty int) bool {
			trial := b.shifted(tx, ty)
			for _, prev := range settled {
				if prev.layer == layer && trial.overlaps(prev.box, layoutMapping.Gap-1) {
					return true
				}
			}
			return false
		}
		// Nearest free translation; deterministic row order and a bounded search.
		if collision(dx, dy) {
			found := false
			for radius := 1; radius < 96 && !found; radius++ {
				for yy := -radius; yy <= radius && !found; yy++ {
					for xx := -radius; xx <= radius; xx++ {
						if abs(xx) != radius && abs(yy) != radius {
							continue
						}
						if !collision(dx+xx, dy+yy) {
							dx += xx
							dy += yy
							found = true
							break
						}
					}
				}
			}
		}
		translateGroup(w, gids, dx, dy)
		settled = append(settled, settledZone{area, layer, b.shifted(dx, dy)})
		placed[area] = true
	}
	// Basement floor numbers are collapsed into the Lower projection. Prevent
	// overlaps across zones without altering stored room coordinates or exits.
	occupied := map[cell]string{}
	for _, id := range ids {
		p := w.rooms[id]
		if p.role == "interior" {
			continue
		}
		plane := layoutPlane(p)
		if _, taken := occupied[cell{plane, p.x, p.y}]; taken {
			p.x, p.y = spiralEmpty(occupied, plane, p.x, p.y)
		}
		occupied[cell{plane, p.x, p.y}] = id
	}
}

type groundPatch struct {
	area, terrain string
	x, y, rx, ry  float64
}

func naturalTerrain(key string) string {
	switch key {
	case "forest", "swamp", "mountain", "snow", "desert":
		return key
	default:
		return "grassland"
	}
}
func landscapeNoise(x, y int) float64 {
	v := uint32(int64(x)*374761393 + int64(y)*668265263)
	v = (v ^ (v >> 13)) * 1274126177
	return float64(v&1023)/1023 - .5
}
func distanceToSegment(x, y, ax, ay, bx, by float64) float64 {
	dx, dy := bx-ax, by-ay
	t := 0.0
	if dx*dx+dy*dy > 0 {
		t = math.Max(0, math.Min(1, ((x-ax)*dx+(y-ay)*dy)/(dx*dx+dy*dy)))
	}
	return math.Hypot(x-ax-t*dx, y-ay-t*dy)
}

// buildLandscape makes anonymous decorative ground around surface zones and
// broad land bridges along zone connections. No cell can be a travel target.
func buildLandscape(w *World, ids []string) []LandCell {
	groups := areaGroups(w, ids)
	areas := make([]string, 0, len(groups))
	for a := range groups {
		areas = append(areas, a)
	}
	sort.Strings(areas)
	patches := []groundPatch{}
	byArea := map[string]int{}
	for _, area := range areas {
		var surface []string
		for _, id := range groups[area] {
			p := w.rooms[id]
			if p.layer == "overworld" && (p.role != "interior" || p.surfaceID == id) {
				surface = append(surface, id)
			}
		}
		if len(surface) == 0 {
			continue
		}
		b := boundsOf(w, surface)
		cfg := layoutMapping.Zones[area]
		ground := cfg.Ground
		if ground == "" {
			ground = naturalTerrain(terrainMapping.Areas[area])
		}
		byArea[area] = len(patches)
		patches = append(patches, groundPatch{area, ground, float64(b.minX+b.maxX) / 2, float64(b.minY+b.maxY) / 2, float64(b.width())/2 + float64(layoutMapping.CoastPadding), float64(b.height())/2 + float64(layoutMapping.CoastPadding)})
	}
	if len(patches) == 0 {
		return nil
	}
	type bridge struct{ a, b int }
	bridges := []bridge{}
	seen := map[[2]int]bool{}
	addBridge := func(a, b int) {
		if a == b {
			return
		}
		if a > b {
			a, b = b, a
		}
		key := [2]int{a, b}
		if !seen[key] {
			seen[key] = true
			bridges = append(bridges, bridge{a, b})
		}
	}
	for _, e := range w.edges {
		a, b := w.rooms[e.from], w.rooms[e.to]
		if a == nil || b == nil || a.area == b.area || e.hidden {
			continue
		}
		ia, aok := byArea[a.area]
		ib, bok := byArea[b.area]
		if aok && bok {
			addBridge(ia, ib)
		}
	}
	// A spanning tree joins disconnected authored islands too; it adds ground,
	// never invented room exits or roads. All decorative land is one continent.
	joined := map[int]bool{0: true}
	for len(joined) < len(patches) {
		best := math.Inf(1)
		ba, bb := 0, 0
		for a := 0; a < len(patches); a++ {
			if !joined[a] {
				continue
			}
			for b := 0; b < len(patches); b++ {
				if joined[b] {
					continue
				}
				d := math.Hypot(patches[a].x-patches[b].x, patches[a].y-patches[b].y)
				if d < best {
					best, ba, bb = d, a, b
				}
			}
		}
		addBridge(ba, bb)
		joined[bb] = true
	}
	minX, maxX, minY, maxY := math.Inf(1), math.Inf(-1), math.Inf(1), math.Inf(-1)
	for _, p := range patches {
		minX = math.Min(minX, p.x-p.rx-1)
		maxX = math.Max(maxX, p.x+p.rx+1)
		minY = math.Min(minY, p.y-p.ry-1)
		maxY = math.Max(maxY, p.y+p.ry+1)
	}
	result := []LandCell{}
	for y := int(math.Floor(minY)); y <= int(math.Ceil(maxY)); y++ {
		for x := int(math.Floor(minX)); x <= int(math.Ceil(maxX)); x++ {
			inside := false
			nearest, best := 0, math.Inf(1)
			for i, p := range patches {
				nx, ny := (float64(x)-p.x)/p.rx, (float64(y)-p.y)/p.ry
				d := nx*nx + ny*ny
				coast := 1 + 0.11*math.Sin(float64(x)*.6) + 0.08*math.Cos(float64(y)*.7) + 0.04*landscapeNoise(x, y)
				if d < coast {
					inside = true
				}
				if d < best {
					best, nearest = d, i
				}
			}
			if !inside {
				for _, b := range bridges {
					a, c := patches[b.a], patches[b.b]
					if distanceToSegment(float64(x), float64(y), a.x, a.y, c.x, c.y) < 3.2+.35*math.Sin(float64(x+y)*.45) {
						inside = true
						break
					}
				}
			}
			if !inside {
				continue
			}
			p := patches[nearest]
			result = append(result, LandCell{X: x, Y: y, Terrain: p.terrain, area: p.area})
		}
	}
	// Actual outdoor water, fields, and woodland become local patches. Streets,
	// interiors, and cave floors never color the overworld filler as crate tiles.
	for i := range result {
		c := &result[i]
	roomLoop:
		for _, id := range ids {
			p := w.rooms[id]
			if p.role != "surface" || p.layer != "overworld" {
				continue
			}
			d := abs(c.X-p.x) + abs(c.Y-p.y)
			if d > 1 {
				continue
			}

			if d == 0 {
				c.area = p.area
				switch p.terrain {
				case "water", "shore", "farmland", "forest", "swamp":
					c.Terrain = p.terrain
					// The room's own water/biome beats neighboring vegetation.
					// In particular, Creek Crossing must remain bridgeable water.
					break roomLoop
				}
			}
			if d == 1 && p.area == c.area {
				switch p.terrain {
				case "farmland", "forest", "swamp":
					c.Terrain = p.terrain
				}
			}
		}
	}
	return result
}
