package game

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/ruleset"
	"github.com/talesmud/talesmud/pkg/service"
)

// ApplySessionStart runs the pending dawn pass when a character enters play
// or reconnects. It refills configured resources and clears a combat flag
// whose instance is already gone. It does not relocate and it does not
// apply a defeat penalty.
func (g *Game) ApplySessionStart(char *characters.Character) {
	if g == nil || g.Facade == nil || char == nil || char.ID == "" {
		return
	}
	if ruleset.ApplyNewDay(char, time.Now()) {
		if err := g.Facade.CharactersService().Update(char.ID, char); err != nil {
			log.WithError(err).WithField("characterID", char.ID).Warn("session start: failed to persist new day")
		}
	}
	if g.Resources != nil {
		for _, allowance := range ruleset.ResourceAllowances() {
			if _, _, err := g.Resources.Get(char.ID, allowance.Key); err != nil {
				log.WithError(err).WithField("key", allowance.Key).Warn("session start: resource refill failed")
			}
		}
	}
	g.clearOrphanCombat(char)
}

// EnsureLivingRoom moves a character whose saved room no longer exists
// to the bind room, or the start room when the bind is also gone.
func (g *Game) EnsureLivingRoom(char *characters.Character) {
	if g == nil || g.Facade == nil || char == nil || char.ID == "" {
		return
	}
	if char.CurrentRoomID != "" {
		if room, err := g.Facade.RoomsService().FindByID(char.CurrentRoomID); err == nil && room != nil {
			return
		}
	}
	dest := g.fallbackRoom(char)
	if dest == "" || dest == char.CurrentRoomID {
		return
	}
	g.RelocateCharacter(char, char.BelongsUserID, dest)
}

// ReleaseToSafety ends a live fight without a defeat penalty. A character
// standing in an instance, or in a real room while the profile asks for a
// move, is placed in the instance exit or the configured safe room.
func (g *Game) ReleaseToSafety(characterID string) {
	if g == nil || g.Facade == nil || characterID == "" {
		return
	}
	char, err := g.Facade.CharactersService().FindByID(characterID)
	if err != nil || char == nil {
		return
	}
	inCombat := char.InCombat
	if g.CombatController != nil && g.CombatController.IsPlayerInCombat(char.ID) {
		inCombat = true
		g.CombatController.EndCombatForPlayer(char.ID)
		if fresh, ferr := g.Facade.CharactersService().FindByID(char.ID); ferr == nil && fresh != nil {
			char = fresh
		}
	}
	if char.InCombat {
		char.InCombat = false
		char.CombatInstanceID = ""
		_ = g.Facade.CharactersService().Update(char.ID, char)
	}
	inClone := g.RoomInstances != nil && char.CurrentRoomID != "" && g.RoomInstances.IsClone(char.CurrentRoomID)
	if !inClone && !inCombat {
		return
	}
	if !inClone && ruleset.SafeRoom() == ruleset.SafeStay {
		return
	}
	dest := ""
	if inClone && g.RoomInstances != nil {
		dest = g.RoomInstances.ReturnRoom(char.ID)
		if dest != "" {
			if room, err := g.Facade.RoomsService().FindByID(dest); err != nil || room == nil {
				dest = ""
			}
		}
	}
	if dest == "" {
		switch ruleset.SafeRoom() {
		case ruleset.SafeStart:
			dest = service.ResolveStartRoomID(g.Facade.ServerSettingsService(), g.Facade.RoomsService())
		default:
			dest = g.fallbackRoom(char)
		}
	}
	if dest == "" || dest == char.CurrentRoomID {
		return
	}
	g.RelocateCharacter(char, char.BelongsUserID, dest)
}

func (g *Game) clearOrphanCombat(char *characters.Character) {
	if char == nil || !char.InCombat {
		return
	}
	if g.CombatController != nil && g.CombatController.IsPlayerInCombat(char.ID) {
		return
	}
	char.InCombat = false
	char.CombatInstanceID = ""
	if err := g.Facade.CharactersService().Update(char.ID, char); err != nil {
		log.WithError(err).WithField("characterID", char.ID).Warn("session start: failed to clear combat flag")
	}
}

func (g *Game) fallbackRoom(char *characters.Character) string {
	if char != nil && char.BoundRoomID != "" {
		if room, err := g.Facade.RoomsService().FindByID(char.BoundRoomID); err == nil && room != nil {
			return char.BoundRoomID
		}
	}
	return service.ResolveStartRoomID(g.Facade.ServerSettingsService(), g.Facade.RoomsService())
}
