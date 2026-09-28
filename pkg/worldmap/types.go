package worldmap

// PlayerMap is the client-facing discovered-world atlas.
// Coordinates are layout units (not pixels). Clients scale and render.
type PlayerMap struct {
	CharacterID   string     `json:"characterId"`
	CurrentRoomID string     `json:"currentRoomId"`
	CurrentLayer  string     `json:"currentLayer"`
	Layers        []Layer    `json:"layers"`
	Places        []Place    `json:"places"`
	Paths         []Path     `json:"paths"`
	Regions       []Region   `json:"regions"`
	Landscape     []LandCell `json:"landscape,omitempty"`
}

// Layer is a vertical slice of the atlas (overworld / lower / upper).
type Layer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"` // overworld | lower | upper
}

// Place is a room (or fog neighbor) in layout space.
type Place struct {
	MapFeatures      []string        `json:"mapFeatures,omitempty"`
	ArtSeed          string          `json:"artSeed,omitempty"`
	UndergroundStyle string          `json:"undergroundStyle,omitempty"`
	ID               string          `json:"id"`
	Name             string          `json:"name,omitempty"`
	Area             string          `json:"area,omitempty"`
	AreaName         string          `json:"areaName,omitempty"`
	Layer            string          `json:"layer"`
	X                float64         `json:"x"`
	Y                float64         `json:"y"`
	Z                int             `json:"z"`
	Biome            string          `json:"biome"`
	Terrain          string          `json:"terrain"`
	MapRole          string          `json:"mapRole,omitempty"`
	SurfaceRoomID    string          `json:"surfaceRoomId,omitempty"`
	Town             bool            `json:"town,omitempty"`
	Entrances        []string        `json:"entrances,omitempty"`
	Kind             string          `json:"kind"`
	Landmark         bool            `json:"landmark,omitempty"`
	Discovered       bool            `json:"discovered"`
	Current          bool            `json:"current,omitempty"`
	CanTravel        bool            `json:"canTravel,omitempty"`
	Tags             []string        `json:"tags,omitempty"`
	Danger           string          `json:"danger,omitempty"` // safe | low | hazard | hostile | uncharted
	Summary          string          `json:"summary,omitempty"`
	Exits            []PlaceExit     `json:"exits,omitempty"`
	Residents        []PlaceResident `json:"residents,omitempty"`
}

// PlaceExit is a visible (or revealed) way out of a discovered room.
type PlaceExit struct {
	Dir      string `json:"dir"`
	To       string `json:"to"`
	ToName   string `json:"toName,omitempty"`
	Hidden   bool   `json:"hidden,omitempty"`
	Vertical bool   `json:"vertical,omitempty"`
}

// PlaceResident is a best-effort NPC/enemy usually found in a room.
type PlaceResident struct {
	Name string `json:"name"`
	Kind string `json:"kind"` // npc | enemy
}

// Path is a walkable (or fog) connection between places.
type Path struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Dir    string `json:"dir"`
	Kind   string `json:"kind"` // road | trail | stair | passage | hidden
	Layer  string `json:"layer"`
	Hidden bool   `json:"hidden,omitempty"`
}

// Region is an organic hull around rooms of one area on one layer.
type Region struct {
	ID     string       `json:"id"`
	Name   string       `json:"name"`
	Layer  string       `json:"layer"`
	Biome  string       `json:"biome"`
	Hull   [][2]float64 `json:"hull"`
	Places []string     `json:"places"`
}

// LandCell is decorative ground. It never identifies a room or permits travel.
type LandCell struct {
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Terrain string `json:"terrain"`
	area    string
}

type placedRoom struct {
	mapFeatures      []string
	artSeed          string
	undergroundStyle string
	id               string
	name             string
	area             string
	areaName         string
	tags             []string
	x, y, z          int
	layer            string
	role             string
	surfaceID        string
	town             bool
	entrances        []string
	terrain          string
	biome            string
	kind             string
	landmark         bool
	canBind          bool
}
