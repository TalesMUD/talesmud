package util

import (
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
)

// CreateRoomDescription ...
func CreateRoomDescription(room *rooms.Room, user *entities.User, game def.GameCtrl) string {
	description := "\n[" + room.Name + "]\n"
	description += room.Description

	// Characters
	if len(*room.Characters) > 0 {
		description += "\n"
		charResult := "- In the room: "

		for i, char := range *room.Characters {
			if i > 0 {
				charResult += ", "
			}

			if character, err := game.GetFacade().CharactersService().FindByID(char); err == nil {
				charResult += character.Name

				if character.ID == user.LastCharacter {
					charResult += "(you)"
				}
			}
		}

		description += charResult
	}
	// Exits
	description += "\n"
	description += "- The visible exits are:\n"

	for _, exit := range *room.Exits {
		if !exit.Hidden {
			description += " + [" + exit.Name + "] " + exit.Description + "\n"
		}
	}

	return description
}

//RemoveStringFromSlice ...
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
