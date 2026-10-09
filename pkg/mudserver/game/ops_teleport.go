package game

import (
	"github.com/talesmud/talesmud/pkg/entities/characters"
)

// OpTeleport moves a character through the same path a player walk uses.
// Online characters receive leave and enter messages. Offline characters
// keep the new room on their saved row. A fight blocks the move unless force
// ends that fight first, with no rewards.
func (g *Game) OpTeleport(characterID, roomID string, force bool) (*OpResult, error) {
	if characterID == "" || roomID == "" {
		return nil, opErr(400, "characterId and roomId are required")
	}
	char, err := g.Facade.CharactersService().FindByID(characterID)
	if err != nil || char == nil {
		return nil, opErr(404, "character not found")
	}
	dest, err := g.Facade.RoomsService().FindByID(roomID)
	if err != nil || dest == nil {
		return nil, opErr(404, "room not found")
	}
	inCombat := char.InCombat
	if g.CombatController != nil && g.CombatController.IsPlayerInCombat(char.ID) {
		inCombat = true
	}
	if inCombat && !force {
		return nil, opErr(409, "character is in combat; pass force to end the fight and move them")
	}
	ended := false
	if inCombat && force && g.CombatController != nil {
		if inst := g.CombatController.GetCombatInstance(char.ID); inst != nil {
			if err := g.CombatController.AbortCombat(inst); err != nil {
				return nil, opErr(409, err.Error())
			}
			ended = true
		}
		char, err = g.Facade.CharactersService().FindByID(characterID)
		if err != nil || char == nil {
			return nil, opErr(404, "character not found")
		}
	}
	from := char.CurrentRoomID
	if from == roomID {
		return nil, opErr(400, "character is already in that room")
	}
	moved, ok := g.RelocateCharacter(char, char.BelongsUserID, roomID)
	if !ok || moved == nil {
		return nil, opErr(400, "could not move the character")
	}
	name := char.Name
	if name == "" {
		name = char.ID
	}
	roomName := moved.Name
	if roomName == "" {
		roomName = moved.ID
	}
	summary := opSummary("Teleported %s to %s.", name, roomName)
	if ended {
		summary = opSummary("Ended the fight, then teleported %s to %s.", name, roomName)
	}
	return &OpResult{
		Summary:    summary,
		Undoable:   true,
		EntityType: "character",
		EntityID:   char.ID,
		Before:     jsonRaw(map[string]string{"roomId": from}),
		After:      jsonRaw(map[string]string{"roomId": roomID}),
		Inverse: inverseOf("teleport", map[string]interface{}{
			"characterId":  char.ID,
			"roomId":       from,
			"expectRoomId": roomID,
		}),
		Detail: map[string]interface{}{
			"characterId": char.ID,
			"fromRoomId":  from,
			"roomId":      roomID,
			"endedCombat": ended,
		},
	}, nil
}

func (g *Game) characterInCombat(char *characters.Character) bool {
	if char == nil {
		return false
	}
	if g != nil && g.CombatController != nil && g.CombatController.IsPlayerInCombat(char.ID) {
		return true
	}
	return char.InCombat
}
