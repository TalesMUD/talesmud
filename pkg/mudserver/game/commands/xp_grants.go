package commands

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// GrantExplorationXP awards XP for discovering a new room.
// Does nothing if the room was already discovered by this character.
func GrantExplorationXP(game def.GameCtrl, char *characters.Character, userID string, room *rooms.Room) {
	if char == nil || room == nil {
		return
	}

	xp, isNewRoom, isNewArea := leveling.CalculateExplorationXP(
		room.ID, room.Area, char.DiscoveredRooms, char.DiscoveredAreas,
	)

	if !isNewRoom {
		return
	}

	// Initialize maps if nil (first discovery for this character)
	if char.DiscoveredRooms == nil {
		char.DiscoveredRooms = make(map[string]bool)
	}
	if char.DiscoveredAreas == nil {
		char.DiscoveredAreas = make(map[string]bool)
	}

	// Mark room as discovered
	char.DiscoveredRooms[room.ID] = true
	char.AllTimeStats.RoomsDiscovered++

	// Mark area as discovered if new
	if isNewArea && room.Area != "" {
		char.DiscoveredAreas[room.Area] = true
	}

	// Award XP
	char.XP += xp

	// Build notification message
	var msg string
	if isNewArea {
		msg = fmt.Sprintf("You discovered a new area: %s! (+%d XP)", room.Area, xp)
	} else {
		msg = fmt.Sprintf("You discovered: %s (+%d XP)", room.Name, xp)
	}

	// Check for level-up
	if levelsGained, _ := leveling.CheckLevelUp(char); levelsGained > 0 {
		result := leveling.ApplyLevelUp(char, levelsGained)

		// Save character with all updates
		game.GetFacade().CharactersService().Update(char.ID, char)

		// Send discovery notification
		game.SendMessage() <- messages.MessageResponse{
			Audience:   messages.MessageAudienceUser,
			AudienceID: userID,
			Type:       messages.MessageTypeDefault,
			Message:    msg,
		}

		// Send level-up notification
		game.SendMessage() <- messages.MessageResponse{
			Audience:   messages.MessageAudienceUser,
			AudienceID: userID,
			Type:       messages.MessageTypeLevelUp,
			Message:    result.Message,
		}

		// Send updated character stats
		game.SendMessage() <- messages.NewCharacterUpdateMessage(userID, char)
	} else {
		// No level-up, just save and notify
		game.GetFacade().CharactersService().Update(char.ID, char)

		// Send discovery notification
		game.SendMessage() <- messages.MessageResponse{
			Audience:   messages.MessageAudienceUser,
			AudienceID: userID,
			Type:       messages.MessageTypeDefault,
			Message:    msg,
		}

		// Send updated character stats (XP changed)
		game.SendMessage() <- messages.NewCharacterUpdateMessage(userID, char)
	}

	log.WithFields(log.Fields{
		"characterID": char.ID,
		"roomID":      room.ID,
		"roomName":    room.Name,
		"xpAwarded":   xp,
		"isNewArea":   isNewArea,
	}).Info("Awarded exploration XP")
}
