package commands

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// AttackCommand handles attacking NPCs and combat actions
type AttackCommand struct {
}

// Key returns the command key matcher
func (command *AttackCommand) Key() CommandKey { return &StartsWithCommandKey{} }

// Execute handles the attack command
func (command *AttackCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	if message.Character.CurrentRoomID == "" {
		game.SendMessage() <- message.Reply("You are not in a room.")
		return true
	}

	combatEngine := game.GetCombatEngine()
	if combatEngine == nil {
		game.SendMessage() <- message.Reply("Combat system is not available.")
		return true
	}

	// Parse target from command: "attack goblin" or "a goblin"
	parts := strings.Fields(message.Data)
	targetName := ""
	if len(parts) >= 2 {
		targetName = strings.Join(parts[1:], " ")
	}

	// Check if player is already in combat
	if combatEngine.IsPlayerInCombat(message.Character.Entity.ID) {
		// Player is in combat - this is an in-combat attack
		return command.handleInCombatAttack(game, message, combatEngine, targetName)
	}

	// Not in combat - need a target name to initiate
	if targetName == "" {
		game.SendMessage() <- message.Reply("Attack whom? Usage: attack <target>")
		return true
	}

	// Player is not in combat - try to initiate combat
	return command.handleInitiateCombat(game, message, combatEngine, targetName)
}

// handleInitiateCombat handles attacking an NPC to start combat
func (command *AttackCommand) handleInitiateCombat(game def.GameCtrl, message *messages.Message, combatEngine def.CombatEngineCtrl, targetName string) bool {
	npcManager := game.GetNPCInstanceManager()
	if npcManager == nil {
		game.SendMessage() <- message.Reply("Error: NPC system not available.")
		return true
	}

	// Find the target NPC in the room
	target := npcManager.FindInstanceByNameInRoom(message.Character.CurrentRoomID, targetName)
	if target == nil {
		game.SendMessage() <- message.Reply(fmt.Sprintf("There is no '%s' here to attack.", targetName))
		return true
	}

	// Check if the NPC is an enemy (has EnemyTrait)
	if !target.IsEnemy() {
		game.SendMessage() <- message.Reply(fmt.Sprintf("%s is not hostile. You cannot attack them.", target.Name))
		return true
	}

	// Check if NPC is already dead
	if target.IsDead {
		game.SendMessage() <- message.Reply(fmt.Sprintf("%s is already dead.", target.Name))
		return true
	}

	// Check if NPC is already in combat with someone else
	if combatEngine.IsNPCInCombat(target.Entity.ID) {
		game.SendMessage() <- message.Reply(fmt.Sprintf("%s is already in combat with someone else!", target.Name))
		return true
	}

	// Gather all enemies to pull into combat
	enemies := []*npc.NPC{target}

	// If the target has CallForHelp, pull nearby enemies
	if target.EnemyTrait != nil && target.EnemyTrait.CallForHelp {
		nearbyNPCs := npcManager.GetInstancesInRoom(message.Character.CurrentRoomID)
		for _, nearby := range nearbyNPCs {
			if nearby.Entity.ID == target.Entity.ID {
				continue // Skip the original target
			}
			if nearby.IsEnemy() && !nearby.IsDead && !combatEngine.IsNPCInCombat(nearby.Entity.ID) {
				// Check if this enemy also has CallForHelp or AggroOnSight
				if nearby.EnemyTrait != nil && (nearby.EnemyTrait.CallForHelp || nearby.EnemyTrait.AggroOnSight) {
					enemies = append(enemies, nearby)
				}
			}
		}
	}

	// Gather players (just the attacker for now, others can join)
	players := []*characters.Character{message.Character}

	// Initiate combat
	instance := combatEngine.InitiateCombat(message.Character.CurrentRoomID, players, enemies)
	if instance == nil {
		game.SendMessage() <- message.Reply("Failed to initiate combat.")
		return true
	}

	// Update character's combat state in database
	message.Character.InCombat = true
	message.Character.CombatInstanceID = instance.ID
	game.GetFacade().CharactersService().Update(message.Character.ID, message.Character)

	// Update NPC combat states
	for _, enemy := range enemies {
		npcManager.UpdateInstance(enemy.Entity.ID, func(n *npc.NPC) {
			n.InCombat = true
			n.CombatInstanceID = instance.ID
			n.State = "combat"
		})
	}

	// Build combat start message
	var enemyNames []string
	for _, e := range enemies {
		enemyNames = append(enemyNames, e.Name)
	}

	startMsg := fmt.Sprintf("\n%s\n%s\n\n",
		"═══════════════════════════════════════════════════",
		"              COMBAT INITIATED!")
	startMsg += fmt.Sprintf("You attack %s!\n\n", target.Name)

	if len(enemies) > 1 {
		startMsg += fmt.Sprintf("Enemies join the fight: %s\n\n", strings.Join(enemyNames[1:], ", "))
	}

	// Show turn order
	startMsg += "Turn Order:\n"
	for i, combatant := range instance.TurnOrder {
		marker := "  "
		if i == instance.CurrentTurnIdx {
			marker = "► "
		}
		startMsg += fmt.Sprintf("%s%d. %s (Initiative: %d)\n", marker, i+1, combatant.Name, combatant.Initiative)
	}

	startMsg += "\n" + combatEngine.GetCombatStatus(message.Character.Entity.ID)
	startMsg += "\n═══════════════════════════════════════════════════"

	game.SendMessage() <- message.Reply(startMsg)

	// Check if it's the player's turn
	currentTurn := instance.GetCurrentTurnCombatant()
	if currentTurn != nil && currentTurn.ID == message.Character.Entity.ID {
		game.SendMessage() <- message.Reply("\nIt's your turn! Actions: attack <target> | defend | flee | status")
	} else if currentTurn != nil {
		// It's an NPC's turn - the game loop will handle NPC actions
		game.SendMessage() <- message.Reply(fmt.Sprintf("\n%s prepares to act...", currentTurn.Name))
	}

	// Notify other players in the room
	roomMsg := messages.MessageResponse{
		Audience:   messages.MessageAudienceRoomWithoutOrigin,
		AudienceID: message.Character.CurrentRoomID,
		OriginID:   message.FromUser.ID,
		Type:       messages.MessageTypeCombatStart,
		Message:    fmt.Sprintf("%s engages %s in combat!", message.Character.Name, strings.Join(enemyNames, ", ")),
	}
	game.SendMessage() <- roomMsg

	return true
}

// handleInCombatAttack handles an attack action during combat
func (command *AttackCommand) handleInCombatAttack(game def.GameCtrl, message *messages.Message, combatEngine def.CombatEngineCtrl, targetName string) bool {
	instance := combatEngine.GetCombatInstance(message.Character.Entity.ID)
	if instance == nil {
		// Player thinks they're in combat but instance is gone
		message.Character.InCombat = false
		message.Character.CombatInstanceID = ""
		game.GetFacade().CharactersService().Update(message.Character.ID, message.Character)
		game.SendMessage() <- message.Reply("You are not in combat.")
		return true
	}

	// Check if it's the player's turn
	currentTurn := instance.GetCurrentTurnCombatant()
	if currentTurn == nil || currentTurn.ID != message.Character.Entity.ID {
		timeRemaining := instance.GetTurnTimeRemaining()
		if currentTurn != nil {
			game.SendMessage() <- message.Reply(fmt.Sprintf("It's not your turn! Waiting for %s. (%ds remaining)", currentTurn.Name, timeRemaining))
		} else {
			game.SendMessage() <- message.Reply("Combat is in an invalid state.")
		}
		return true
	}

	// Find the target enemy
	var targetID string

	// Get list of living enemies
	var livingEnemies []struct {
		id   string
		name string
		hp   int32
		maxHP int32
	}
	for _, enemy := range instance.Enemies {
		if enemy.IsAlive {
			livingEnemies = append(livingEnemies, struct {
				id   string
				name string
				hp   int32
				maxHP int32
			}{enemy.ID, enemy.Name, enemy.CurrentHP, enemy.MaxHP})
		}
	}

	// Auto-target if only one enemy and no target specified
	if targetName == "" {
		if len(livingEnemies) == 1 {
			targetID = livingEnemies[0].id
		} else if len(livingEnemies) > 1 {
			// Multiple enemies - need to specify target
			var targets []string
			for _, e := range livingEnemies {
				targets = append(targets, fmt.Sprintf("%s (%d/%d HP)", e.name, e.hp, e.maxHP))
			}
			game.SendMessage() <- message.Reply(fmt.Sprintf("Multiple enemies! Specify target: attack <target>\nAvailable: %s", strings.Join(targets, ", ")))
			return true
		} else {
			game.SendMessage() <- message.Reply("No enemies to attack!")
			return true
		}
	} else {
		// Search for target by name
		targetNameLower := strings.ToLower(targetName)
		for _, enemy := range livingEnemies {
			if strings.Contains(strings.ToLower(enemy.name), targetNameLower) {
				targetID = enemy.id
				break
			}
		}
	}

	if targetID == "" {
		// List available targets
		var targets []string
		for _, e := range livingEnemies {
			targets = append(targets, fmt.Sprintf("%s (%d/%d HP)", e.name, e.hp, e.maxHP))
		}
		game.SendMessage() <- message.Reply(fmt.Sprintf("Invalid target '%s'. Available targets: %s", targetName, strings.Join(targets, ", ")))
		return true
	}

	// Process the attack
	resultMsg, combatEnded, endState := combatEngine.ProcessPlayerAttack(message.Character.Entity.ID, targetID)

	game.SendMessage() <- message.Reply(resultMsg)

	if combatEnded {
		command.handleCombatEnd(game, message, combatEngine, endState)
	} else {
		// Show next turn info
		newInstance := combatEngine.GetCombatInstance(message.Character.Entity.ID)
		if newInstance != nil {
			nextTurn := newInstance.GetCurrentTurnCombatant()
			if nextTurn != nil {
				if nextTurn.ID == message.Character.Entity.ID {
					game.SendMessage() <- message.Reply("\nIt's still your turn! Actions: attack <target> | defend | flee | status")
				} else {
					game.SendMessage() <- message.Reply(fmt.Sprintf("\n%s's turn...", nextTurn.Name))
				}
			}
		}
	}

	return true
}

// handleCombatEnd handles the end of combat
func (command *AttackCommand) handleCombatEnd(game def.GameCtrl, message *messages.Message, combatEngine def.CombatEngineCtrl, endState combat.CombatState) {
	// Get the combat instance before cleanup
	instance := combatEngine.GetCombatInstance(message.Character.Entity.ID)

	var endMsg string
	switch endState {
	case combat.CombatStateVictory:
		endMsg = "\n═══════════════════════════════════════════════════\n"
		endMsg += "              VICTORY!\n"
		endMsg += "═══════════════════════════════════════════════════\n\n"

		if instance != nil {
			// Calculate and show rewards
			var totalXP int64 = 0
			var totalGold int64 = 0

			for _, enemy := range instance.Enemies {
				// Get NPC data to access rewards
				npcData := game.GetNPCInstanceManager().GetInstance(enemy.ID)
				if npcData != nil && npcData.EnemyTrait != nil {
					totalXP += npcData.EnemyTrait.XPReward

					// Roll gold
					goldRange := npcData.EnemyTrait.GoldDrop
					if goldRange.Max > goldRange.Min {
						// Random between min and max
						gold := goldRange.Min + (goldRange.Max-goldRange.Min)/2 // Simplified
						totalGold += int64(gold)
					} else {
						totalGold += int64(goldRange.Min)
					}
				}
				endMsg += fmt.Sprintf("Defeated: %s\n", enemy.Name)
			}

			// Award XP and gold to living players
			livingPlayers := instance.GetLivingPlayers()
			if len(livingPlayers) > 0 {
				xpPerPlayer := totalXP / int64(len(livingPlayers))
				goldPerPlayer := totalGold / int64(len(livingPlayers))

				endMsg += fmt.Sprintf("\nREWARDS:\n")
				endMsg += fmt.Sprintf("  Experience: %d XP\n", xpPerPlayer)
				endMsg += fmt.Sprintf("  Gold: %d gold\n", goldPerPlayer)

				// Update character with rewards
				message.Character.XP += int32(xpPerPlayer)
				message.Character.Gold += goldPerPlayer
			}
		}

		endMsg += "\n═══════════════════════════════════════════════════"

	case combat.CombatStateDefeat:
		endMsg = "\n═══════════════════════════════════════════════════\n"
		endMsg += "              DEFEAT\n"
		endMsg += "═══════════════════════════════════════════════════\n\n"
		endMsg += "You have been defeated!\n\n"

		// Calculate gold loss (10%)
		goldLoss := int64(float64(message.Character.Gold) * 0.10)
		if goldLoss > 0 {
			message.Character.Gold -= goldLoss
			endMsg += fmt.Sprintf("PENALTY: Lost %d gold\n", goldLoss)
		}

		// Set HP to 50% and respawn
		message.Character.CurrentHitPoints = message.Character.MaxHitPoints / 2
		if message.Character.CurrentHitPoints < 1 {
			message.Character.CurrentHitPoints = 1
		}

		// Determine respawn location
		respawnRoom := message.Character.BoundRoomID
		if respawnRoom == "" {
			// Use a default starting room (would need to be configured)
			respawnRoom = message.Character.CurrentRoomID // Stay in same room as fallback
		}

		endMsg += fmt.Sprintf("\nYou awaken with %d/%d HP.\n", message.Character.CurrentHitPoints, message.Character.MaxHitPoints)
		endMsg += "═══════════════════════════════════════════════════"

	case combat.CombatStateFled:
		endMsg = "\n═══════════════════════════════════════════════════\n"
		endMsg += "              ESCAPED\n"
		endMsg += "═══════════════════════════════════════════════════\n\n"
		endMsg += "You have fled from combat!\n"
		endMsg += "═══════════════════════════════════════════════════"

	default:
		endMsg = "\nCombat has ended."
	}

	// Clear combat state
	message.Character.InCombat = false
	message.Character.CombatInstanceID = ""
	game.GetFacade().CharactersService().Update(message.Character.ID, message.Character)

	// Clear NPC combat states
	if instance != nil {
		npcManager := game.GetNPCInstanceManager()
		for _, enemy := range instance.Enemies {
			npcManager.UpdateInstance(enemy.ID, func(n *npc.NPC) {
				n.InCombat = false
				n.CombatInstanceID = ""
				if n.CurrentHitPoints > 0 {
					n.State = "idle"
				}
			})
		}
	}

	// End combat in engine
	combatEngine.EndCombatForPlayer(message.Character.Entity.ID)

	game.SendMessage() <- message.Reply(endMsg)
}

// DefendCommand handles the defend action in combat
type DefendCommand struct{}

// Key returns the command key matcher
func (command *DefendCommand) Key() CommandKey { return &ExactCommandKey{} }

// Execute handles the defend command
func (command *DefendCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	combatEngine := game.GetCombatEngine()
	if combatEngine == nil || !combatEngine.IsPlayerInCombat(message.Character.Entity.ID) {
		game.SendMessage() <- message.Reply("You are not in combat.")
		return true
	}

	instance := combatEngine.GetCombatInstance(message.Character.Entity.ID)
	if instance == nil {
		game.SendMessage() <- message.Reply("Combat instance not found.")
		return true
	}

	// Check if it's the player's turn
	currentTurn := instance.GetCurrentTurnCombatant()
	if currentTurn == nil || currentTurn.ID != message.Character.Entity.ID {
		game.SendMessage() <- message.Reply("It's not your turn!")
		return true
	}

	resultMsg, combatEnded, endState := combatEngine.ProcessPlayerDefend(message.Character.Entity.ID)
	game.SendMessage() <- message.Reply(resultMsg)

	if combatEnded {
		attackCmd := &AttackCommand{}
		attackCmd.handleCombatEnd(game, message, combatEngine, endState)
	} else {
		// Show next turn info
		newInstance := combatEngine.GetCombatInstance(message.Character.Entity.ID)
		if newInstance != nil {
			nextTurn := newInstance.GetCurrentTurnCombatant()
			if nextTurn != nil && nextTurn.ID != message.Character.Entity.ID {
				game.SendMessage() <- message.Reply(fmt.Sprintf("\n%s's turn...", nextTurn.Name))
			}
		}
	}

	return true
}

// FleeCommand handles fleeing from combat
type FleeCommand struct{}

// Key returns the command key matcher
func (command *FleeCommand) Key() CommandKey { return &ExactCommandKey{} }

// Execute handles the flee command
func (command *FleeCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	combatEngine := game.GetCombatEngine()
	if combatEngine == nil || !combatEngine.IsPlayerInCombat(message.Character.Entity.ID) {
		game.SendMessage() <- message.Reply("You are not in combat. There's nothing to flee from.")
		return true
	}

	instance := combatEngine.GetCombatInstance(message.Character.Entity.ID)
	if instance == nil {
		game.SendMessage() <- message.Reply("Combat instance not found.")
		return true
	}

	// Check if it's the player's turn
	currentTurn := instance.GetCurrentTurnCombatant()
	if currentTurn == nil || currentTurn.ID != message.Character.Entity.ID {
		game.SendMessage() <- message.Reply("It's not your turn! You can only flee on your turn.")
		return true
	}

	success, resultMsg, combatEnded, endState := combatEngine.ProcessPlayerFlee(message.Character.Entity.ID)
	game.SendMessage() <- message.Reply(resultMsg)

	if combatEnded {
		attackCmd := &AttackCommand{}
		attackCmd.handleCombatEnd(game, message, combatEngine, endState)
	} else if !success {
		// Flee failed, show next turn
		newInstance := combatEngine.GetCombatInstance(message.Character.Entity.ID)
		if newInstance != nil {
			nextTurn := newInstance.GetCurrentTurnCombatant()
			if nextTurn != nil && nextTurn.ID != message.Character.Entity.ID {
				game.SendMessage() <- message.Reply(fmt.Sprintf("\n%s's turn...", nextTurn.Name))
			}
		}
	}

	return true
}

// CombatStatusCommand shows current combat status
type CombatStatusCommand struct{}

// Key returns the command key matcher
func (command *CombatStatusCommand) Key() CommandKey { return &ExactCommandKey{} }

// Execute handles the status command
func (command *CombatStatusCommand) Execute(game def.GameCtrl, message *messages.Message) bool {
	if message.Character == nil {
		game.SendMessage() <- message.Reply("You need to select a character first.")
		return true
	}

	combatEngine := game.GetCombatEngine()
	if combatEngine == nil || !combatEngine.IsPlayerInCombat(message.Character.Entity.ID) {
		game.SendMessage() <- message.Reply("You are not in combat.")
		return true
	}

	status := combatEngine.GetCombatStatus(message.Character.Entity.ID)
	game.SendMessage() <- message.Reply(status)

	return true
}
