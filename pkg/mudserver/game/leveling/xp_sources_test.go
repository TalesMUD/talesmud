package leveling

import "testing"

func TestExplorationXPNewRoom(t *testing.T) {
	discoveredRooms := map[string]bool{}
	discoveredAreas := map[string]bool{"some-area": true}

	xp, isNewRoom, isNewArea := CalculateExplorationXP("room-1", "some-area", discoveredRooms, discoveredAreas)

	if !isNewRoom {
		t.Error("Expected isNewRoom to be true")
	}
	if isNewArea {
		t.Error("Expected isNewArea to be false (area already discovered)")
	}
	if xp != ExplorationXPPerRoom {
		t.Errorf("Expected %d XP for normal room, got %d", ExplorationXPPerRoom, xp)
	}
}

func TestExplorationXPNewArea(t *testing.T) {
	discoveredRooms := map[string]bool{}
	discoveredAreas := map[string]bool{}

	xp, isNewRoom, isNewArea := CalculateExplorationXP("room-1", "new-area", discoveredRooms, discoveredAreas)

	if !isNewRoom {
		t.Error("Expected isNewRoom to be true")
	}
	if !isNewArea {
		t.Error("Expected isNewArea to be true")
	}
	if xp != ExplorationXPNewArea {
		t.Errorf("Expected %d XP for new area, got %d", ExplorationXPNewArea, xp)
	}
}

func TestExplorationXPAlreadyDiscovered(t *testing.T) {
	discoveredRooms := map[string]bool{"room-1": true}
	discoveredAreas := map[string]bool{}

	xp, isNewRoom, isNewArea := CalculateExplorationXP("room-1", "some-area", discoveredRooms, discoveredAreas)

	if isNewRoom {
		t.Error("Expected isNewRoom to be false")
	}
	if isNewArea {
		t.Error("Expected isNewArea to be false")
	}
	if xp != 0 {
		t.Errorf("Expected 0 XP for already-discovered room, got %d", xp)
	}
}

func TestExplorationXPNilMaps(t *testing.T) {
	// Nil maps simulate a new character with no discovery tracking yet
	xp, isNewRoom, isNewArea := CalculateExplorationXP("room-1", "some-area", nil, nil)

	if !isNewRoom {
		t.Error("Expected isNewRoom to be true with nil maps")
	}
	if !isNewArea {
		t.Error("Expected isNewArea to be true with nil maps")
	}
	if xp != ExplorationXPNewArea {
		t.Errorf("Expected %d XP for new area with nil maps, got %d", ExplorationXPNewArea, xp)
	}
}

func TestExplorationXPNoAreaName(t *testing.T) {
	// Room with no area should get normal room XP
	xp, isNewRoom, isNewArea := CalculateExplorationXP("room-1", "", nil, nil)

	if !isNewRoom {
		t.Error("Expected isNewRoom to be true")
	}
	if isNewArea {
		t.Error("Expected isNewArea to be false for room with no area")
	}
	if xp != ExplorationXPPerRoom {
		t.Errorf("Expected %d XP for room with no area, got %d", ExplorationXPPerRoom, xp)
	}
}
