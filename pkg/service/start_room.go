package service

import (
	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

const DefaultStartRoomID = "R0001"

// ResolveStartRoomID returns the room ID new/guest characters should spawn in.
// Order: configured settings.StartRoomID, then R0001. Empty string means neither exists.
func ResolveStartRoomID(settings ServerSettingsService, roomsSvc RoomsService) string {
	candidates := make([]string, 0, 2)
	if settings != nil {
		if s, err := settings.Get(); err == nil && s != nil && s.StartRoomID != "" {
			candidates = append(candidates, s.StartRoomID)
		}
	}
	candidates = append(candidates, DefaultStartRoomID)

	seen := map[string]bool{}
	for _, id := range candidates {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		if roomsSvc == nil {
			return id
		}
		if room, err := roomsSvc.FindByID(id); err == nil && room != nil {
			return id
		}
	}
	return ""
}

// ResolveStartRoom loads the start room entity. Does not fall back to rooms[0].
func ResolveStartRoom(settings ServerSettingsService, roomsSvc RoomsService) *rooms.Room {
	id := ResolveStartRoomID(settings, roomsSvc)
	if id == "" || roomsSvc == nil {
		return nil
	}
	room, err := roomsSvc.FindByID(id)
	if err != nil {
		return nil
	}
	return room
}
