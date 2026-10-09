package game

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/instances"
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
// An orphan copy (the id contains the instance marker, and the live manager
// does not know it) uses the startup sweep order: hub, return exit, start
// room, then bind. A live copy is left where it is.
func (g *Game) EnsureLivingRoom(char *characters.Character) {
	if g == nil || g.Facade == nil || char == nil || char.ID == "" {
		return
	}
	// A live copy is a real place to stand. Only an orphan marker, or a
	// missing room, is relocated here.
	orphan := g.orphanInstanceCopy(char.CurrentRoomID)
	if char.CurrentRoomID != "" && !orphan {
		if room, err := g.Facade.RoomsService().FindByID(char.CurrentRoomID); err == nil && room != nil {
			return
		}
	}
	dest := ""
	if orphan {
		if all, err := loadRoomMap(g.Facade.RoomsService()); err == nil {
			start := service.ResolveStartRoomID(g.Facade.ServerSettingsService(), g.Facade.RoomsService())
			if !roomPresent(all, start) {
				start = ""
			}
			fallback := ""
			if roomPresent(all, char.BoundRoomID) {
				fallback = char.BoundRoomID
			}
			dest = instances.RelocationDest(all, char.CurrentRoomID, start, fallback)
		}
	}
	if dest == "" {
		dest = g.fallbackRoom(char)
	}
	if dest == "" || dest == char.CurrentRoomID {
		return
	}
	g.clearOrphanCombat(char)
	moved, ok := g.RelocateCharacter(char, char.BelongsUserID, dest)
	if !ok || !orphan || moved == nil {
		return
	}
	name := moved.Name
	if name == "" {
		name = moved.ID
	}
	g.noteRelocation(char.ID, name)
}

// inInstanceCopy is true for a live manager clone and for any id that still
// carries the instance marker after the manager was emptied.
func (g *Game) inInstanceCopy(roomID string) bool {
	if instances.IsCloneID(roomID) {
		return true
	}
	return g != nil && g.RoomInstances != nil && roomID != "" && g.RoomInstances.IsClone(roomID)
}

// orphanInstanceCopy is a saved copy the live manager does not know about.
func (g *Game) orphanInstanceCopy(roomID string) bool {
	if !instances.IsCloneID(roomID) {
		return false
	}
	if g != nil && g.RoomInstances != nil && g.RoomInstances.IsClone(roomID) {
		return false
	}
	return true
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
	inClone := g.inInstanceCopy(char.CurrentRoomID)
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
