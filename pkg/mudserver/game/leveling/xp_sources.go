package leveling

const (
	// ExplorationXPPerRoom is the XP awarded for discovering a normal room
	ExplorationXPPerRoom int32 = 5
	// ExplorationXPNewArea is the XP awarded for discovering the first room in a new area/zone
	ExplorationXPNewArea int32 = 15
)

// CalculateExplorationXP returns the XP to award for entering a room.
// Returns 0 if the room was already discovered.
// Awards bonus XP (ExplorationXPNewArea) for the first room in a new area.
func CalculateExplorationXP(roomID, areaName string, discoveredRooms, discoveredAreas map[string]bool) (xp int32, isNewRoom bool, isNewArea bool) {
	if discoveredRooms != nil && discoveredRooms[roomID] {
		return 0, false, false
	}

	isNewRoom = true

	// First room in a new area gets bonus XP
	if areaName != "" && (discoveredAreas == nil || !discoveredAreas[areaName]) {
		return ExplorationXPNewArea, true, true
	}

	return ExplorationXPPerRoom, true, false
}
