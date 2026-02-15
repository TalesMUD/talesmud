package game

import (
	"log"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

const (
	passiveRegenPercent    = 0.02  // 2% per tick (out of combat)
	restingRegenPercent    = 0.10  // 10% per tick (resting)
	combatHPRegenPercent   = 0.005 // 0.5% per tick (in combat) — 1/4 of passive
	combatManaRegenPercent = 0.01  // 1% per tick (in combat) — 1/5 of passive
)

// handleRegenerationUpdates processes HP regeneration for all online players.
// Called every 10 seconds by the regeneration ticker.
func (g *Game) handleRegenerationUpdates() {
	// Get all online users
	onlineUsers, err := g.GetFacade().UsersService().FindAllOnline()
	if err != nil || len(onlineUsers) == 0 {
		return
	}

	for _, user := range onlineUsers {
		if user.LastCharacter == "" {
			continue
		}

		// Load character
		char, err := g.GetFacade().CharactersService().FindByID(user.LastCharacter)
		if err != nil || char == nil {
			continue
		}

		// Check if fully recovered (HP and mana)
		hpFull := char.CurrentHitPoints >= char.MaxHitPoints
		manaFull := char.MaxMana <= 0 || char.CurrentMana >= char.MaxMana
		if char.CurrentHitPoints <= 0 || (hpFull && manaFull) {
			if hpFull && manaFull && isResting(char) {
				g.clearRestingState(char, user.ID)
				g.SendMessage() <- messages.Reply(user.ID, "You are now fully rested and recovered.")
			}
			continue
		}

		// Clear resting state if in combat
		if char.InCombat && isResting(char) {
			g.clearRestingState(char, user.ID)
		}

		// Calculate and apply HP regeneration (reduced in combat)
		regenAmount := g.calculateRegenAmount(char)

		// Calculate mana regeneration (reduced in combat)
		manaRegenAmount := g.calculateManaRegenAmount(char)

		if regenAmount > 0 || manaRegenAmount > 0 {
			g.applyRegeneration(char, user.ID, regenAmount, manaRegenAmount)
		}
	}
}

// calculateRegenAmount determines how much HP to regenerate based on character state.
func (g *Game) calculateRegenAmount(char *characters.Character) int32 {
	var regenPercent float64

	if char.InCombat {
		regenPercent = combatHPRegenPercent
	} else if isResting(char) {
		regenPercent = restingRegenPercent
	} else {
		regenPercent = passiveRegenPercent
	}

	// Calculate amount (minimum 1 HP per tick)
	amount := int32(float64(char.MaxHitPoints) * regenPercent)
	if amount < 1 {
		amount = 1
	}

	return amount
}

// calculateManaRegenAmount determines how much mana to regenerate based on character state.
func (g *Game) calculateManaRegenAmount(char *characters.Character) int32 {
	if char.MaxMana <= 0 || char.CurrentMana >= char.MaxMana {
		return 0
	}

	var regenPercent float64
	if char.InCombat {
		regenPercent = combatManaRegenPercent
	} else if isResting(char) {
		regenPercent = 0.15 // 15% per tick while resting
	} else {
		regenPercent = 0.05 // 5% per tick out of combat
	}

	amount := int32(float64(char.MaxMana) * regenPercent)
	if amount < 1 {
		amount = 1
	}
	return amount
}

// applyRegeneration applies HP and mana regeneration to a character.
func (g *Game) applyRegeneration(char *characters.Character, userID string, hpAmount int32, manaAmount int32) {
	// Add HP and cap at MaxHP
	if hpAmount > 0 {
		char.CurrentHitPoints += hpAmount
		if char.CurrentHitPoints > char.MaxHitPoints {
			char.CurrentHitPoints = char.MaxHitPoints
		}
	}

	// Add mana and cap at MaxMana
	if manaAmount > 0 {
		char.CurrentMana += manaAmount
		if char.CurrentMana > char.MaxMana {
			char.CurrentMana = char.MaxMana
		}
	}

	// Save character
	if err := g.GetFacade().CharactersService().Update(char.ID, char); err != nil {
		log.Printf("Error saving character %s during regeneration: %v", char.ID, err)
		return
	}

	// Send character update message
	g.SendMessage() <- messages.NewCharacterUpdateMessage(userID, char)

	// If resting and now fully recovered, send completion message
	hpFull := char.CurrentHitPoints >= char.MaxHitPoints
	manaFull := char.MaxMana <= 0 || char.CurrentMana >= char.MaxMana
	if hpFull && manaFull && isResting(char) {
		g.clearRestingState(char, userID)
		g.SendMessage() <- messages.Reply(userID, "You are now fully rested and recovered.")
	}
}

// clearRestingState removes the resting flag from a character.
func (g *Game) clearRestingState(char *characters.Character, userID string) {
	if char.Flags == nil {
		return
	}

	delete(char.Flags, "resting")

	// Save character
	if err := g.GetFacade().CharactersService().Update(char.ID, char); err != nil {
		log.Printf("Error clearing resting state for character %s: %v", char.ID, err)
	}
}

// InterruptRest stops a character from resting (called by other commands).
func (g *Game) InterruptRest(char *characters.Character) {
	if !isResting(char) {
		return
	}

	delete(char.Flags, "resting")

	// Save character
	if err := g.GetFacade().CharactersService().Update(char.ID, char); err != nil {
		log.Printf("Error interrupting rest for character %s: %v", char.ID, err)
	}
}

// isResting checks if a character is currently resting.
func isResting(char *characters.Character) bool {
	if char.Flags == nil {
		return false
	}

	resting, ok := char.Flags["resting"]
	if !ok {
		return false
	}

	restingBool, ok := resting.(bool)
	return ok && restingBool
}
