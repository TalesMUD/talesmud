package game

import (
	"log"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/ruleset"
)

// handleRegenerationUpdates processes regeneration for online players.
// The ticker fires once a second on the server clock, not per character.
// A tick where no pool is due returns before any character is loaded.
// Housekeeping (skip the dead, clear a full rest, clear rest in a fight)
// runs on the ticks that do load characters. With the default 10 second
// intervals that is the same cadence as before. Disabling every pool also
// skips that pass. applyRegeneration and InterruptRest still clear the flag.
func (g *Game) handleRegenerationUpdates() {
	if g == nil {
		return
	}
	g.regenTick++
	tick := g.regenTick
	profile := ruleset.Regen()
	if !anyPoolDue(tick, profile) {
		return
	}

	onlinePlayers := g.GetOnlinePlayers()
	if len(onlinePlayers) == 0 {
		return
	}

	for _, player := range onlinePlayers {
		if player.CharacterID == "" {
			continue
		}

		char, err := g.GetFacade().CharactersService().FindByID(player.CharacterID)
		if err != nil || char == nil {
			continue
		}

		hpFull := char.CurrentHitPoints >= char.MaxHitPoints
		manaFull := char.MaxMana <= 0 || char.CurrentMana >= char.MaxMana
		if char.CurrentHitPoints <= 0 || (hpFull && manaFull) {
			if hpFull && manaFull && isResting(char) {
				g.clearRestingState(char, player.UserID)
				g.SendMessage() <- messages.Reply(player.UserID, "You are now fully rested and recovered.")
			}
			continue
		}
		if char.InCombat && isResting(char) {
			g.clearRestingState(char, player.UserID)
		}

		hpAmount, manaAmount := regenAmounts(char, tick, profile)
		if hpAmount > 0 || manaAmount > 0 {
			g.applyRegeneration(char, player.UserID, hpAmount, manaAmount)
		}
	}
}

// poolDue reports whether this pool grants points on tick.
// Inactive pools are never due, even when the interval is 1.
func poolDue(tick uint64, p ruleset.RegenPolicy) bool {
	if !p.Active() || p.IntervalSeconds < 1 {
		return false
	}
	return tick%uint64(p.IntervalSeconds) == 0
}

// anyPoolDue is false when none of the six pools grant points on tick.
func anyPoolDue(tick uint64, profile ruleset.RegenProfile) bool {
	pairs := []ruleset.RegenPair{profile.OutOfCombat, profile.Resting, profile.InCombat}
	for _, pair := range pairs {
		if poolDue(tick, pair.HP) || poolDue(tick, pair.Mana) {
			return true
		}
	}
	return false
}

// regenAmounts is the HP and mana gained on this tick.
// In combat uses the in-combat pools, even when the resting flag is set.
// Otherwise the resting flag selects the resting pools, and the rest use
// the out-of-combat pools. A full HP pool contributes nothing.
func regenAmounts(char *characters.Character, tick uint64, profile ruleset.RegenProfile) (hp, mana int32) {
	if char == nil {
		return 0, 0
	}
	var pair ruleset.RegenPair
	switch {
	case char.InCombat:
		pair = profile.InCombat
	case isResting(char):
		pair = profile.Resting
	default:
		pair = profile.OutOfCombat
	}
	if poolDue(tick, pair.HP) && char.CurrentHitPoints < char.MaxHitPoints {
		hp = policyAmount(char.MaxHitPoints, pair.HP)
	}
	if poolDue(tick, pair.Mana) && char.MaxMana > 0 && char.CurrentMana < char.MaxMana {
		mana = policyAmount(char.MaxMana, pair.Mana)
	}
	return hp, mana
}

func policyAmount(max int32, policy ruleset.RegenPolicy) int32 {
	if !policy.Active() {
		return 0
	}
	amount := int32(float64(max)*(policy.Percent/100)) + policy.Flat
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
	if userID != "" {
		g.SendMessage() <- messages.NewCharacterUpdateMessage(userID, char)
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
	if char.BelongsUserID != "" {
		g.SendMessage() <- messages.NewCharacterUpdateMessage(char.BelongsUserID, char)
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
