package worldmap

import "strings"

type vec struct{ x, y, z int }

var directionOffsets = map[string]vec{
	"north":     {0, -1, 0},
	"south":     {0, 1, 0},
	"east":      {1, 0, 0},
	"west":      {-1, 0, 0},
	"northeast": {1, -1, 0},
	"northwest": {-1, -1, 0},
	"southeast": {1, 1, 0},
	"southwest": {-1, 1, 0},
	"ne":        {1, -1, 0},
	"nw":        {-1, -1, 0},
	"se":        {1, 1, 0},
	"sw":        {-1, 1, 0},
	"up":        {0, 0, 1},
	"down":      {0, 0, -1},
}

var directionAliases = map[string]string{
	"n":         "north",
	"s":         "south",
	"e":         "east",
	"w":         "west",
	"u":         "up",
	"d":         "down",
	"upward":    "up",
	"upwards":   "up",
	"ascend":    "up",
	"downward":  "down",
	"downwards": "down",
	"descend":   "down",
}

func normalizeDir(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	if alias, ok := directionAliases[n]; ok {
		return alias
	}
	return n
}

func offsetFor(name string) (vec, bool) {
	d := normalizeDir(name)
	off, ok := directionOffsets[d]
	return off, ok
}

func isVertical(name string) bool {
	d := normalizeDir(name)
	return d == "up" || d == "down"
}

func isCompass(name string) bool {
	off, ok := offsetFor(name)
	return ok && off.z == 0
}
