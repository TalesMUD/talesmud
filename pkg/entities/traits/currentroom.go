package traits

//CurrentRoom ...
type CurrentRoom struct {
	CurrentRoomID string `bson:"currentRoomID,omitempty" json:"currentRoomID,omitempty"`
}

// IsInCurrentRoom ...
func IsInCurrentRoom(roomID string) *CurrentRoom {
	return &CurrentRoom{
		CurrentRoomID: roomID,
	}
}
