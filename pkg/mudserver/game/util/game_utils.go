package util

import (
	"strconv"

	"github.com/talesmud/talesmud/pkg/entities"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
)

// CreateRoomDescription ...
func CreateRoomDescription(room *rooms.Room, user *entities.User, game def.GameCtrl) string {
	description := "\n[" + room.Name + "]\n"
	description += room.Description

	// Characters - only show online players
	if room.Characters != nil && len(*room.Characters) > 0 {
		onlineChars := []string{}

		for _, char := range *room.Characters {
			if character, err := game.GetFacade().CharactersService().FindByID(char); err == nil {
				// Check if the character's owner is online
				if charUser, err := game.GetFacade().UsersService().FindByID(character.BelongsUserID); err == nil {
					// Only show if user is online AND this is their active character
					if charUser.IsOnline && charUser.LastCharacter == character.ID {
						charName := character.Name
						if character.ID == user.LastCharacter {
							charName += "(you)"
						}
						onlineChars = append(onlineChars, charName)
					}
				}
			}
		}

		if len(onlineChars) > 0 {
			description += "\n"
			charResult := "- In the room: "
			for i, name := range onlineChars {
				if i > 0 {
					charResult += ", "
				}
				charResult += name
			}
			description += charResult
		}
	}

	// NPCs in the room from the instance manager (spawned NPCs are in memory, not database)
	npcs := game.GetNPCInstanceManager().GetInstancesInRoom(room.ID)
	if len(npcs) > 0 {
		var enemies []string
		var friendlyNPCs []string

		// Build display names with numbers for duplicates
		displayNames := buildNPCDisplayNames(npcs)

		for _, n := range npcs {
			displayName := displayNames[n.Entity.ID]
			if n.IsEnemy() {
				enemies = append(enemies, displayName)
			} else {
				friendlyNPCs = append(friendlyNPCs, displayName)
			}
		}

		if len(enemies) > 0 {
			description += "\n- Enemies: "
			for i, name := range enemies {
				if i > 0 {
					description += ", "
				}
				description += name
			}
		}

		if len(friendlyNPCs) > 0 {
			description += "\n- NPCs: "
			for i, name := range friendlyNPCs {
				if i > 0 {
					description += ", "
				}
				description += name
			}
		}
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

// buildNPCDisplayNames creates display names for NPCs, adding numbers when duplicates exist
// Returns a map of NPC ID -> display name (e.g., "Rat #1", "Rat #2" or just "Rat" if unique)
func buildNPCDisplayNames(npcs []*npc.NPC) map[string]string {
	result := make(map[string]string)

	// Count NPCs by base name
	nameCounts := make(map[string]int)
	for _, n := range npcs {
		nameCounts[n.Name]++
	}

	// Track current index for each name
	nameIndex := make(map[string]int)

	// Assign display names
	for _, n := range npcs {
		baseName := n.Name
		if nameCounts[baseName] > 1 {
			// Multiple NPCs with same name - add number
			nameIndex[baseName]++
			result[n.Entity.ID] = baseName + "#" + strconv.Itoa(nameIndex[baseName])
		} else {
			// Unique name - no number needed
			result[n.Entity.ID] = baseName
		}
	}

	return result
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
