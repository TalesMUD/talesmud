package util

import (
	"github.com/atla/owndnd/pkg/entities/rooms"
)

// CreateRoomDescription ...
func CreateRoomDescription(room *rooms.Room) string {
	description := "[" + room.Name + "]\n"
	description += room.Description + "\n"
	description += "\n"
	description += "The visible exits are:\n"

	for _, exit := range room.Exits {
		description += "[" + exit.Name + "] " + exit.Description + "\n"
	}

	return description
}

//RemoveCharacter ...
func RemoveStringFromSlice(slice []string, inst string) []string {

	for i, elem := range slice {
		if elem == inst {
			if i == len(slice)-1 {
				return append(slice[:i-1])
			}
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
