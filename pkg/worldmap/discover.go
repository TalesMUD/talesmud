package worldmap

import (
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

// MarkOn records that the character has entered this room.
// Returns true when the room is newly discovered.
// Does not award exploration XP — that path stays independently gated.
func MarkOn(ch *characters.Character, room *rooms.Room) bool {
	if ch == nil || room == nil || room.Entity == nil || room.ID == "" {
		return false
	}
	if ch.DiscoveredRooms == nil {
		ch.DiscoveredRooms = map[string]bool{}
	}
	if ch.DiscoveredRooms[room.ID] {
		return false
	}
	ch.DiscoveredRooms[room.ID] = true
	ch.AllTimeStats.RoomsDiscovered++
	if room.Area != "" {
		if ch.DiscoveredAreas == nil {
			ch.DiscoveredAreas = map[string]bool{}
		}
		ch.DiscoveredAreas[room.Area] = true
	}
	return true
}
