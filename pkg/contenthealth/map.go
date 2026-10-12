package contenthealth

import (
	"sort"
	"strings"

	"github.com/talesmud/talesmud/pkg/worldindex"
)

// MapRoom is one content room on the world-map reachability layer.
type MapRoom struct {
	ID        string `json:"id"`
	Reachable bool   `json:"reachable"`
	Instance  bool   `json:"instance"`
}

// MapIsland is one set of rooms the start room cannot reach.
type MapIsland struct {
	RoomIDs []string `json:"roomIds"`
	Reason  string   `json:"reason"`
}

// MapReachability is the world-map layer. It uses the same walk as content health.
type MapReachability struct {
	StartRoomID string      `json:"startRoomId"`
	Rooms       []MapRoom   `json:"rooms"`
	Islands     []MapIsland `json:"islands"`
}

// MapReachabilityView builds the reachable set and the unreachable islands.
// from overrides the start room. An unknown from returns false.
// An empty from uses the world's start room, then the engine default.
func MapReachabilityView(world World, from string) (MapReachability, bool) {
	snap := world.indexSnapshot()
	from = strings.TrimSpace(from)
	if from != "" {
		if snap.Rooms[from] == nil {
			return MapReachability{}, false
		}
		snap.StartRoomID = from
	}
	reach := worldindex.Build(snap).Reachability()

	islands := make([]MapIsland, 0, len(reach.Islands))
	for _, island := range reach.Islands {
		ids := island.RoomIDs
		if ids == nil {
			ids = []string{}
		}
		islands = append(islands, MapIsland{RoomIDs: ids, Reason: island.Reason})
	}

	ids := make([]string, 0, len(snap.Rooms))
	for id := range snap.Rooms {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	rooms := make([]MapRoom, 0, len(ids))
	for _, id := range ids {
		room := snap.Rooms[id]
		rooms = append(rooms, MapRoom{
			ID:        id,
			Reachable: reach.Reachable[id],
			Instance:  room != nil && room.IsInstanceTemplate(),
		})
	}
	return MapReachability{
		StartRoomID: reach.StartRoomID,
		Rooms:       rooms,
		Islands:     islands,
	}, true
}
