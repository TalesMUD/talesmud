package game

import (
	"log"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/ruleset"
)

const (
	// Combat and resting regeneration stay on a 10 second cadence.
	combatRestRegenInterval = 10

	restingRegenPercent     = 0.10  // 10% HP per tick while resting
	restingManaRegenPercent = 0.15  // 15% mana per tick while resting
	combatHPRegenPercent    = 0.005 // 0.5% HP per tick in combat
	combatManaRegenPercent  = 0.01  // 1% mana per tick in combat
)

// handleRegenerationUpdates processes regeneration for online players.
// The ticker fires once a second. Combat, resting, and the fully-rested
// message run every 10th tick. Passive rates use the ruleset intervals.
// A tick with nothing due returns before any character is loaded.
func (g *Game) handleRegenerationUpdates() {
	if g == nil {
		return
	}
	g.regenTick++
	tick := g.regenTick
	hpPolicy, manaPolicy := ruleset.OutOfCombatRegen()
	combatOrRestDue := cadenceDue(tick, combatRestRegenInterval)
	passiveHPDue := cadenceDue(tick, hpPolicy.IntervalSeconds)
	passiveManaDue := cadenceDue(tick, manaPolicy.IntervalSeconds)
	if !combatOrRestDue && !passiveHPDue && !passiveManaDue {
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

		if combatOrRestDue {
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
		} else if char.CurrentHitPoints <= 0 || char.InCombat || isResting(char) {
			continue
		} else {
			hpFull := char.CurrentHitPoints >= char.MaxHitPoints
			manaFull := char.MaxMana <= 0 || char.CurrentMana >= char.MaxMana
			if hpFull && manaFull {
				continue
			}
		}

		hpAmount, manaAmount := regenAmounts(char, tick, hpPolicy, manaPolicy)
		if hpAmount > 0 || manaAmount > 0 {
			g.applyRegeneration(char, player.UserID, hpAmount, manaAmount)
		}
	}
}

func cadenceDue(tick uint64, intervalSeconds int) bool {
	if intervalSeconds < 1 {
		return false
	}
	return tick%uint64(intervalSeconds) == 0
}

// regenAmounts is the HP and mana gained on this tick.
// Out of combat and not resting, the policies apply on their own intervals.
// In combat and while resting, the fixed 10 second rates apply and the policies do not.
func regenAmounts(char *characters.Character, tick uint64, hpPolicy, manaPolicy ruleset.RegenPolicy) (hp, mana int32) {
	if char == nil {
		return 0, 0
	}
	switch {
	case char.InCombat:
		if !cadenceDue(tick, combatRestRegenInterval) {
			return 0, 0
		}
		return scaledAmount(char.MaxHitPoints, combatHPRegenPercent), manaScaled(char, combatManaRegenPercent)
	case isResting(char):
		if !cadenceDue(tick, combatRestRegenInterval) {
			return 0, 0
		}
		return scaledAmount(char.MaxHitPoints, restingRegenPercent), manaScaled(char, restingManaRegenPercent)
	default:
		if cadenceDue(tick, hpPolicy.IntervalSeconds) {
			hp = policyAmount(char.MaxHitPoints, hpPolicy)
		}
		if cadenceDue(tick, manaPolicy.IntervalSeconds) && char.MaxMana > 0 && char.CurrentMana < char.MaxMana {
			mana = policyAmount(char.MaxMana, manaPolicy)
		}
		return hp, mana
	}
}

func policyAmount(max int32, policy ruleset.RegenPolicy) int32 {
	if !policy.Enabled || (policy.Percent == 0 && policy.Flat == 0) {
		return 0
	}
	amount := int32(float64(max)*(policy.Percent/100)) + policy.Flat
	if (policy.Percent > 0 || policy.Flat > 0) && amount < 1 {
		amount = 1
	}
	return amount
}

func scaledAmount(max int32, rate float64) int32 {
	amount := int32(float64(max) * rate)
	if amount < 1 {
		amount = 1
	}
	return amount
}

func manaScaled(char *characters.Character, rate float64) int32 {
	if char.MaxMana <= 0 || char.CurrentMana >= char.MaxMana {
		return 0
	}
	return scaledAmount(char.MaxMana, rate)
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
