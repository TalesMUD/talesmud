package worldmap

// PlayerMap is the client-facing discovered-world atlas.
// Coordinates are layout units (not pixels). Clients scale and render.
type PlayerMap struct {
	CharacterID   string   `json:"characterId"`
	CurrentRoomID string   `json:"currentRoomId"`
	CurrentLayer  string   `json:"currentLayer"`
	Layers        []Layer  `json:"layers"`
	Places        []Place  `json:"places"`
	Paths         []Path   `json:"paths"`
	Regions       []Region `json:"regions"`
}

// Layer is a vertical slice of the atlas (overworld / lower / upper).
type Layer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"` // overworld | lower | upper
}

// Place is a room (or fog neighbor) in layout space.
type Place struct {
	ID         string   `json:"id"`
	Name       string   `json:"name,omitempty"`
	Area       string   `json:"area,omitempty"`
	AreaName   string   `json:"areaName,omitempty"`
	Layer      string   `json:"layer"`
	X          float64  `json:"x"`
	Y          float64  `json:"y"`
	Z          int      `json:"z"`
	Biome      string   `json:"biome"`
	Kind       string   `json:"kind"`
	Landmark   bool     `json:"landmark,omitempty"`
	Discovered bool     `json:"discovered"`
	Current    bool     `json:"current,omitempty"`
	CanTravel  bool     `json:"canTravel,omitempty"`
	Tags       []string `json:"tags,omitempty"`
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

type placedRoom struct {
	id       string
	name     string
	area     string
	areaName string
	tags     []string
	x, y, z  int
	biome    string
	kind     string
	landmark bool
	canBind  bool
}
