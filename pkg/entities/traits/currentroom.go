package traits

//CurrentRoom ...
type CurrentRoom struct {
	CurrentRoomID string `bson:"currentRoom,omitempty" json:"currentRoom,omitempty"`
}

// IsInCurrentRoom ...
func IsInCurrentRoom(roomID string) *CurrentRoom {
	return &CurrentRoom{
		CurrentRoomID: roomID,
	}
}
