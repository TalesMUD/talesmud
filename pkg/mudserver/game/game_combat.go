package game

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	combatpkg "github.com/talesmud/talesmud/pkg/mudserver/game/combat"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/mudserver/game/util"
)

// CombatController wraps the combat engine and implements CombatEngineCtrl interface
type CombatController struct {
	manager *combatpkg.Manager
	engine  *combatpkg.Engine
	game    *Game
}

// NewCombatController creates a new combat controller
func NewCombatController(game *Game) *CombatController {
	manager := combatpkg.NewManager()
	engine := combatpkg.NewEngine(manager, nil) // Uses default config

	return &CombatController{
		manager: manager,
		engine:  engine,
		game:    game,
	}
}

// IsPlayerInCombat checks if a player is currently in combat
func (c *CombatController) IsPlayerInCombat(characterID string) bool {
	return c.manager.IsPlayerInCombat(characterID)
}

// IsNPCInCombat checks if an NPC is currently in combat
func (c *CombatController) IsNPCInCombat(npcID string) bool {
	return c.manager.IsNPCInCombat(npcID)
}

// GetCombatInstance returns the combat instance a player is in
func (c *CombatController) GetCombatInstance(characterID string) *combat.CombatInstance {
	return c.manager.GetInstanceByPlayerID(characterID)
}

// InitiateCombat starts combat between players and enemies
func (c *CombatController) InitiateCombat(roomID string, players []*characters.Character, enemies []*npc.NPC) *combat.CombatInstance {
	return c.engine.InitiateCombat(roomID, players, enemies)
}

// ProcessPlayerAttack handles a player attacking a target in combat
func (c *CombatController) ProcessPlayerAttack(characterID, targetID string) (message string, combatEnded bool, endState combat.CombatState) {
	instance := c.manager.GetInstanceByPlayerID(characterID)
	if instance == nil {
		return "You are not in combat.", false, combat.CombatStateActive
	}

	result := c.engine.ProcessAttack(instance, characterID, targetID)
	message = result.Message

	// Advance turn
	c.engine.NextTurn(instance)

	// Check if combat ended
	endState = c.engine.CheckCombatEnd(instance)
	if endState != combat.CombatStateActive {
		c.engine.EndCombat(instance, endState)
		combatEnded = true
		return message, combatEnded, endState
	}

	// Process NPC turns if any
	c.processNPCTurns(instance)

	// Check again after NPC turns
	endState = c.engine.CheckCombatEnd(instance)
	if endState != combat.CombatStateActive {
		c.engine.EndCombat(instance, endState)
		combatEnded = true
	}

	return message, combatEnded, endState
}

// ProcessPlayerDefend handles a player defending
func (c *CombatController) ProcessPlayerDefend(characterID string) (message string, combatEnded bool, endState combat.CombatState) {
	instance := c.manager.GetInstanceByPlayerID(characterID)
	if instance == nil {
		return "You are not in combat.", false, combat.CombatStateActive
	}

	result := c.engine.ProcessDefend(instance, characterID)
	message = result.Message

	// Advance turn
	c.engine.NextTurn(instance)

	// Check if combat ended (unlikely from defend, but possible)
	endState = c.engine.CheckCombatEnd(instance)
	if endState != combat.CombatStateActive {
		c.engine.EndCombat(instance, endState)
		combatEnded = true
		return message, combatEnded, endState
	}

	// Process NPC turns if any
	c.processNPCTurns(instance)

	// Check again after NPC turns
	endState = c.engine.CheckCombatEnd(instance)
	if endState != combat.CombatStateActive {
		c.engine.EndCombat(instance, endState)
		combatEnded = true
	}

	return message, combatEnded, endState
}

// ProcessPlayerFlee handles a player attempting to flee
func (c *CombatController) ProcessPlayerFlee(characterID string) (success bool, message string, combatEnded bool, endState combat.CombatState) {
	instance := c.manager.GetInstanceByPlayerID(characterID)
	if instance == nil {
		return false, "You are not in combat.", false, combat.CombatStateActive
	}

	result := c.engine.ProcessFlee(instance, characterID)
	success = result.Success
	message = result.Message

	// Advance turn (even on failed flee)
	c.engine.NextTurn(instance)

	// Check if combat ended (all players fled)
	endState = c.engine.CheckCombatEnd(instance)
	if endState != combat.CombatStateActive {
		c.engine.EndCombat(instance, endState)
		combatEnded = true
		return success, message, combatEnded, endState
	}

	// If flee failed, process NPC turns
	if !success {
		c.processNPCTurns(instance)

		// Check again after NPC turns
		endState = c.engine.CheckCombatEnd(instance)
		if endState != combat.CombatStateActive {
			c.engine.EndCombat(instance, endState)
			combatEnded = true
		}
	}

	return success, message, combatEnded, endState
}

// ProcessPlayerSkill handles a player using a skill in combat
func (c *CombatController) ProcessPlayerSkill(characterID, skillID, targetID string) (message string, combatEnded bool, endState combat.CombatState) {
	instance := c.manager.GetInstanceByPlayerID(characterID)
	if instance == nil {
		return "You are not in combat.", false, combat.CombatStateActive
	}

	result := c.engine.ProcessSkill(instance, characterID, skillID, targetID)
	message = strings.Join(result.Messages, "\n")

	if !result.Success {
		return message, false, combat.CombatStateActive
	}

	// Advance turn
	c.engine.NextTurn(instance)

	// Check if combat ended
	endState = c.engine.CheckCombatEnd(instance)
	if endState != combat.CombatStateActive {
		c.engine.EndCombat(instance, endState)
		combatEnded = true
		return
	}

	// Process NPC turns if any
	c.processNPCTurns(instance)

	// Check again after NPC turns
	endState = c.engine.CheckCombatEnd(instance)
	if endState != combat.CombatStateActive {
		c.engine.EndCombat(instance, endState)
		combatEnded = true
	}

	return
}

// GetCombatStatus returns a formatted status string for the combat
func (c *CombatController) GetCombatStatus(characterID string) string {
	instance := c.manager.GetInstanceByPlayerID(characterID)
	if instance == nil {
		return "You are not in combat."
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\n═══════════════ COMBAT STATUS - Round %d ═══════════════\n\n", instance.Round))

	// Show players
	sb.WriteString("YOUR PARTY:\n")
	for _, player := range instance.Players {
		marker := "  "
		if player.ID == characterID {
			marker = "► "
		}
		hpBar := createHPBar(player.CurrentHP, player.MaxHP)
		status := ""
		if !player.IsAlive {
			status = " [DEAD]"
		} else if player.HasFled {
			status = " [FLED]"
		}
		sb.WriteString(fmt.Sprintf("%s%-16s %s %d/%d HP%s\n", marker, player.Name, hpBar, player.CurrentHP, player.MaxHP, status))
		if player.MaxMana > 0 {
			manaBar := createManaBar(player.CurrentMana, player.MaxMana)
			sb.WriteString(fmt.Sprintf("  %-16s %s %d/%d MP\n", "", manaBar, player.CurrentMana, player.MaxMana))
		}
		if len(player.StatusEffects) > 0 {
			var effects []string
			for _, se := range player.StatusEffects {
				effects = append(effects, fmt.Sprintf("%s(%d)", se.Name, se.Duration))
			}
			sb.WriteString(fmt.Sprintf("  %-16s Effects: %s\n", "", strings.Join(effects, ", ")))
		}
	}

	sb.WriteString("\nENEMIES:\n")
	for _, enemy := range instance.Enemies {
		hpBar := createHPBar(enemy.CurrentHP, enemy.MaxHP)
		status := ""
		if !enemy.IsAlive {
			status = " [DEAD]"
		}
		sb.WriteString(fmt.Sprintf("  %-16s %s %d/%d HP%s\n", enemy.Name, hpBar, enemy.CurrentHP, enemy.MaxHP, status))
		if len(enemy.StatusEffects) > 0 {
			var effects []string
			for _, se := range enemy.StatusEffects {
				effects = append(effects, fmt.Sprintf("%s(%d)", se.Name, se.Duration))
			}
			sb.WriteString(fmt.Sprintf("  %-16s Effects: %s\n", "", strings.Join(effects, ", ")))
		}
	}

	// Show turn order
	sb.WriteString("\nTURN ORDER:\n")
	for i, combatant := range instance.TurnOrder {
		if !combatant.IsAlive || combatant.HasFled {
			continue
		}
		marker := "  "
		if i == instance.CurrentTurnIdx {
			marker = "► "
		}
		sb.WriteString(fmt.Sprintf("%s%d. %s (%d)\n", marker, i+1, combatant.Name, combatant.Initiative))
	}

	// Show auto-attack target and queued action for the player
	player := instance.GetPlayerByID(characterID)
	if player != nil {
		if player.AutoAttackTargetID != "" {
			target := instance.GetCombatantByID(player.AutoAttackTargetID)
			if target != nil && target.IsAlive {
				sb.WriteString(fmt.Sprintf("\nAuto-attacking: %s (%d/%d HP)", target.Name, target.CurrentHP, target.MaxHP))
			}
		}
		if player.QueuedAction != "" {
			queuedInfo := string(player.QueuedAction)
			if player.QueuedAction == combat.CombatActionAttack && player.QueuedTargetID != "" {
				target := instance.GetCombatantByID(player.QueuedTargetID)
				if target != nil {
					queuedInfo = fmt.Sprintf("attack %s", target.Name)
				}
			}
			sb.WriteString(fmt.Sprintf("\nQueued action: %s", queuedInfo))
		}
	}

	sb.WriteString("\n\nCommands: attack <target> | defend | flee | cast <skill> [target] | status")
	sb.WriteString("\n═══════════════════════════════════════════════════════")

	return sb.String()
}

// EndCombatForPlayer removes a player from combat (cleanup on disconnect, etc.)
func (c *CombatController) EndCombatForPlayer(characterID string) {
	instance := c.manager.GetInstanceByPlayerID(characterID)
	if instance == nil {
		return
	}

	// Remove the instance
	c.manager.RemoveInstance(instance.ID)
}

// processNPCTurns handles NPC turns in combat until it's a player's turn
func (c *CombatController) processNPCTurns(instance *combat.CombatInstance) {
	maxNPCTurns := 10 // Safety limit to prevent infinite loops

	for i := 0; i < maxNPCTurns; i++ {
		current := instance.GetCurrentTurnCombatant()
		if current == nil {
			break
		}

		// If it's a player's turn, stop processing NPC turns
		if current.Type == combat.CombatantTypePlayer {
			break
		}

		// It's an NPC's turn
		npcEntity := c.game.NPCManager.GetInstance(current.ID)

		// Determine action
		action, targetID := c.engine.GetNPCAIAction(instance, current, npcEntity)

		var actionMsg string
		switch action {
		case combat.CombatActionAttack:
			if targetID != "" {
				result := c.engine.ProcessAttack(instance, current.ID, targetID)
				actionMsg = result.Message

				// Send message to players in combat
				c.notifyPlayersInCombat(instance, actionMsg)

				// If a player died, sync their HP
				if result.TargetDied {
					target := instance.GetCombatantByID(targetID)
					if target != nil && target.Type == combat.CombatantTypePlayer {
						c.syncPlayerHP(targetID, 0)
					}
				}
			}

		case combat.CombatActionDefend:
			result := c.engine.ProcessDefend(instance, current.ID)
			actionMsg = result.Message
			c.notifyPlayersInCombat(instance, actionMsg)

		case combat.CombatActionFlee:
			result := c.engine.ProcessFlee(instance, current.ID)
			actionMsg = result.Message
			c.notifyPlayersInCombat(instance, actionMsg)
		}

		// Advance turn
		c.engine.NextTurn(instance)

		// Check if combat ended
		endState := c.engine.CheckCombatEnd(instance)
		if endState != combat.CombatStateActive {
			break
		}
	}
}

// notifyPlayersInCombat sends a message to all players in the combat instance
func (c *CombatController) notifyPlayersInCombat(instance *combat.CombatInstance, message string) {
	for _, player := range instance.Players {
		if player.IsAlive && !player.HasFled {
			// Find the user for this character
			char, err := c.game.Facade.CharactersService().FindByID(player.ID)
			if err != nil {
				continue
			}
			c.game.sendMessage <- messages.MessageResponse{
				Audience:   messages.MessageAudienceUser,
				AudienceID: char.BelongsUserID,
				Type:       messages.MessageTypeCombatAction,
				Message:    message,
			}
		}
	}
}

// notifyAllPlayersInInstance sends a message to all players regardless of alive/fled status
func (c *CombatController) notifyAllPlayersInInstance(instance *combat.CombatInstance, message string) {
	for _, player := range instance.Players {
		char, err := c.game.Facade.CharactersService().FindByID(player.ID)
		if err != nil {
			continue
		}
		c.game.sendMessage <- messages.MessageResponse{
			Audience:   messages.MessageAudienceUser,
			AudienceID: char.BelongsUserID,
			Type:       messages.MessageTypeCombatEnd,
			Message:    message,
		}
	}
}

// sendPlayerCharacterUpdate sends updated character stats to all players in the combat instance
func (c *CombatController) sendPlayerCharacterUpdate(instance *combat.CombatInstance) {
	for _, player := range instance.Players {
		char, err := c.game.Facade.CharactersService().FindByID(player.ID)
		if err != nil {
			continue
		}
		// Reflect combat HP and mana in the update
		char.CurrentHitPoints = player.CurrentHP
		char.CurrentMana = player.CurrentMana
		char.InCombat = true

		update := messages.NewCharacterUpdateMessage(char.BelongsUserID, char)
		if update != nil {
			c.game.sendMessage <- update
		}
	}
}

// syncPlayerHP updates a player's HP in the database
func (c *CombatController) syncPlayerHP(characterID string, hp int32) {
	char, err := c.game.Facade.CharactersService().FindByID(characterID)
	if err != nil {
		return
	}
	char.CurrentHitPoints = hp
	c.game.Facade.CharactersService().Update(characterID, char)
}

// createHPBar creates a visual HP bar
func createHPBar(current, max int32) string {
	if max <= 0 {
		return "[░░░░░░░░░░]"
	}
	ratio := float64(current) / float64(max)
	filled := int(ratio * 10)
	if filled > 10 {
		filled = 10
	}
	if filled < 0 {
		filled = 0
	}

	bar := "["
	for i := 0; i < 10; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	bar += "]"
	return bar
}

// createManaBar creates a visual mana bar
func createManaBar(current, max int32) string {
	if max <= 0 {
		return "[░░░░░░░░░░]"
	}
	ratio := float64(current) / float64(max)
	filled := int(ratio * 10)
	if filled > 10 {
		filled = 10
	}
	if filled < 0 {
		filled = 0
	}

	bar := "["
	for i := 0; i < 10; i++ {
		if i < filled {
			bar += "▓"
		} else {
			bar += "░"
		}
	}
	bar += "]"
	return bar
}

// Update handles combat updates (called from game loop)
func (c *CombatController) Update() {
	instances := c.manager.GetActiveInstances()

	for _, instance := range instances {
		// Process all turns continuously (both NPC and player)
		c.processAllTurns(instance)

		// Check for global combat timeout
		if time.Since(instance.CreatedAt).Minutes() >= float64(c.engine.Config.CombatTimeoutMinutes) {
			c.engine.EndCombat(instance, combat.CombatStateTimeout)
			c.notifyPlayersInCombat(instance, "Combat has timed out due to inactivity.")
			c.cleanupCombatInstance(instance, combat.CombatStateTimeout)
		}
	}
}

// processAllTurns processes all combatant turns (NPC and player) in sequence
func (c *CombatController) processAllTurns(instance *combat.CombatInstance) {
	maxTurns := len(instance.TurnOrder) + 2 // Safety limit per tick

	for i := 0; i < maxTurns; i++ {
		current := instance.GetCurrentTurnCombatant()
		if current == nil {
			break
		}

		// Process status effects at start of turn (DoTs, HoTs, stun check)
		logLenBefore := len(instance.Log)
		stunned := c.engine.ProcessStatusEffects(instance, current)

		// Notify players of status effect messages
		for j := logLenBefore; j < len(instance.Log); j++ {
			if instance.Log[j].Message != "" {
				c.notifyPlayersInCombat(instance, instance.Log[j].Message)
			}
		}

		// Re-fetch current (may have been modified by status effects)
		current = instance.GetCurrentTurnCombatant()
		if current == nil || !current.IsAlive {
			endState := c.engine.CheckCombatEnd(instance)
			if endState != combat.CombatStateActive {
				c.engine.EndCombat(instance, endState)
				c.cleanupCombatInstance(instance, endState)
				return
			}
			c.engine.NextTurn(instance)
			continue
		}

		if stunned {
			c.sendPlayerCharacterUpdate(instance)
			c.engine.NextTurn(instance)
			continue
		}

		if current.Type == combat.CombatantTypeNPC {
			// Process NPC turn
			npcEntity := c.game.NPCManager.GetInstance(current.ID)
			action, targetID := c.engine.GetNPCAIAction(instance, current, npcEntity)

			switch action {
			case combat.CombatActionAttack:
				if targetID != "" {
					result := c.engine.ProcessAttack(instance, current.ID, targetID)
					c.notifyPlayersInCombat(instance, result.Message)
					if result.TargetDied {
						target := instance.GetCombatantByID(targetID)
						if target != nil && target.Type == combat.CombatantTypePlayer {
							c.syncPlayerHP(targetID, 0)
						}
					}
				}
			case combat.CombatActionDefend:
				result := c.engine.ProcessDefend(instance, current.ID)
				c.notifyPlayersInCombat(instance, result.Message)
			case combat.CombatActionFlee:
				result := c.engine.ProcessFlee(instance, current.ID)
				c.notifyPlayersInCombat(instance, result.Message)
			}
		} else {
			// Process player turn via auto-attack
			c.processPlayerAutoAttack(instance, current)
		}

		// Send updated character stats to all players after each action
		c.sendPlayerCharacterUpdate(instance)

		// Advance turn
		c.engine.NextTurn(instance)

		// Check if combat ended
		endState := c.engine.CheckCombatEnd(instance)
		if endState != combat.CombatStateActive {
			c.engine.EndCombat(instance, endState)
			c.cleanupCombatInstance(instance, endState)
			return
		}
	}
}

// cleanupCombatInstance cleans up after combat ends, processes rewards, and notifies players
func (c *CombatController) cleanupCombatInstance(instance *combat.CombatInstance, endState combat.CombatState) {

	// Process combat end based on state
	switch endState {
	case combat.CombatStateVictory:
		c.processCombatVictory(instance)
	case combat.CombatStateDefeat:
		c.processCombatDefeat(instance)
	case combat.CombatStateFled:
		c.notifyAllPlayersInInstance(instance, "\n═══════════════════════════════════════════════════\n              ESCAPED\n═══════════════════════════════════════════════════\n\nYou have fled from combat!\n═══════════════════════════════════════════════════")
	case combat.CombatStateTimeout:
		// Timeout message already sent in Update()
	}

	// Clear combat state from players
	for _, player := range instance.Players {
		char, err := c.game.Facade.CharactersService().FindByID(player.ID)
		if err != nil {
			continue
		}
		char.InCombat = false
		char.CombatInstanceID = ""

		// Sync HP and mana
		char.CurrentHitPoints = player.CurrentHP
		char.CurrentMana = player.CurrentMana

		c.game.Facade.CharactersService().Update(player.ID, char)

		// Send updated stats to client (combat ended, final HP/XP/Gold)
		update := messages.NewCharacterUpdateMessage(char.BelongsUserID, char)
		if update != nil {
			c.game.sendMessage <- update
		}
	}

	// Clear combat state from NPCs
	for _, enemy := range instance.Enemies {
		c.game.NPCManager.UpdateInstance(enemy.ID, func(n *npc.NPC) {
			n.InCombat = false
			n.CombatInstanceID = ""
			n.CurrentHitPoints = enemy.CurrentHP
			if enemy.IsAlive {
				n.State = "idle"
			} else {
				n.IsDead = true
				n.State = "dead"
				n.DeathTime = time.Now()
			}
		})
	}

	// Remove the instance
	c.manager.RemoveInstance(instance.ID)

	// Send a room refresh to all players so dead NPCs disappear from the UI
	if room, err := c.game.Facade.RoomsService().FindByID(instance.OriginRoomID); err == nil && room != nil {
		for _, player := range instance.Players {
			char, err := c.game.Facade.CharactersService().FindByID(player.ID)
			if err != nil {
				continue
			}
			user, err := c.game.Facade.UsersService().FindByID(char.BelongsUserID)
			if err != nil {
				continue
			}
			enterRoom := messages.NewEnterRoomMessage(util.RoomWithCharacterReveals(room, char), user, c.game, char)
			enterRoom.AudienceID = user.ID
			c.game.sendMessage <- enterRoom
		}
	}

	log.WithFields(log.Fields{
		"instanceID": instance.ID,
		"endState":   endState,
	}).Info("Combat instance cleaned up")
}

// processCombatVictory handles XP, gold, loot rewards and sends the victory message
func (c *CombatController) processCombatVictory(instance *combat.CombatInstance) {
	var totalXP int64
	var totalGold int64
	var allLootItems []string
	var enemyNames []string

	// Get the room for loot drops
	room, roomErr := c.game.Facade.RoomsService().FindByID(instance.OriginRoomID)

	for _, enemy := range instance.Enemies {
		if enemy.IsAlive {
			continue
		}
		enemyNames = append(enemyNames, enemy.Name)

		npcData := c.game.NPCManager.GetInstance(enemy.ID)
		if npcData == nil || npcData.EnemyTrait == nil {
			continue
		}

		// Use configured XP or calculate from NPC level
		xpReward := npcData.EnemyTrait.XPReward
		if xpReward == 0 {
			xpReward = leveling.CalculateEnemyXPReward(npcData.Level)
		}
		totalXP += xpReward

		// Roll gold - use configured GoldDrop range or calculate from level/difficulty
		goldRange := npcData.EnemyTrait.GoldDrop
		if goldRange.Max > 0 {
			if goldRange.Max > goldRange.Min {
				totalGold += int64(goldRange.Min) + int64(rand.Intn(int(goldRange.Max-goldRange.Min+1)))
			} else {
				totalGold += int64(goldRange.Min)
			}
		} else {
			totalGold += leveling.RollEnemyGold(npcData.Level, npcData.EnemyTrait.Difficulty, rand.Intn)
		}

		// Process loot drops (items placed in room)
		if roomErr == nil && room != nil {
			// Use first living player's level for level-gated drops
			var killerLevel int32 = 1
			livingPlayers := instance.GetLivingPlayers()
			if len(livingPlayers) > 0 {
				char, err := c.game.Facade.CharactersService().FindByID(livingPlayers[0].ID)
				if err == nil && char != nil {
					killerLevel = char.Level
				}
			}

			lootResult, err := DropLootFromNPC(c.game.Facade, npcData, room, killerLevel)
			if err == nil && lootResult != nil {
				for _, item := range lootResult.Items {
					if item.Stackable && item.Quantity > 1 {
						allLootItems = append(allLootItems, fmt.Sprintf("%s (x%d)", item.Name, item.Quantity))
					} else {
						allLootItems = append(allLootItems, item.Name)
					}
				}
			}
		}
	}

	// Track quest progress for NPC kills
	livingPlayers := instance.GetLivingPlayers()
	for _, enemy := range instance.Enemies {
		if enemy.IsAlive {
			continue
		}
		npcData := c.game.NPCManager.GetInstance(enemy.ID)
		if npcData == nil {
			continue
		}
		for _, player := range livingPlayers {
			char, err := c.game.Facade.CharactersService().FindByID(player.ID)
			if err != nil || char == nil {
				continue
			}
			if c.game.QuestTracker != nil {
				c.game.QuestTracker.OnNPCKilled(char.ID, char.BelongsUserID, npcData)
			}
		}
	}

	// Split rewards among living players
	numLiving := int64(len(livingPlayers))
	if numLiving == 0 {
		numLiving = 1
	}
	xpPerPlayer := totalXP / numLiving
	goldPerPlayer := totalGold / numLiving

	// Build victory message
	var sb strings.Builder
	sb.WriteString("\n═══════════════════════════════════════════════════\n")
	sb.WriteString("              VICTORY!\n")
	sb.WriteString("═══════════════════════════════════════════════════\n\n")

	for _, name := range enemyNames {
		sb.WriteString(fmt.Sprintf("Defeated: %s\n", name))
	}

	sb.WriteString(fmt.Sprintf("\nREWARDS:\n"))
	if xpPerPlayer > 0 {
		sb.WriteString(fmt.Sprintf("  + %d XP\n", xpPerPlayer))
	}
	if goldPerPlayer > 0 {
		sb.WriteString(fmt.Sprintf("  + %d Gold\n", goldPerPlayer))
	}
	if len(allLootItems) > 0 {
		sb.WriteString("\nLOOT DROPPED:\n")
		for _, itemName := range allLootItems {
			sb.WriteString(fmt.Sprintf("  - %s\n", itemName))
		}
	}
	if xpPerPlayer == 0 && goldPerPlayer == 0 && len(allLootItems) == 0 {
		sb.WriteString("  (none)\n")
	}

	sb.WriteString("\n═══════════════════════════════════════════════════")

	// Award rewards and notify each living player
	for _, player := range livingPlayers {
		char, err := c.game.Facade.CharactersService().FindByID(player.ID)
		if err != nil {
			continue
		}

		char.XP += int32(xpPerPlayer)
		char.Gold += goldPerPlayer

		// Check for level-up
		if levelsGained, _ := leveling.CheckLevelUp(char); levelsGained > 0 {
			// Apply level-up stat changes
			result := leveling.ApplyLevelUp(char, levelsGained)

			// Save updated character
			c.game.Facade.CharactersService().Update(player.ID, char)

			// Send victory message first
			c.game.sendMessage <- messages.MessageResponse{
				Audience:   messages.MessageAudienceUser,
				AudienceID: char.BelongsUserID,
				Type:       messages.MessageTypeCombatEnd,
				Message:    sb.String(),
			}

			// Send level-up notification
			c.game.sendMessage <- messages.MessageResponse{
				Audience:   messages.MessageAudienceUser,
				AudienceID: char.BelongsUserID,
				Type:       messages.MessageTypeLevelUp,
				Message:    result.Message,
			}

			// Send updated character stats to client
			update := messages.NewCharacterUpdateMessage(char.BelongsUserID, char)
			c.game.sendMessage <- update
		} else {
			// No level-up, just save character with updated XP and gold
			c.game.Facade.CharactersService().Update(player.ID, char)

			// Send victory message
			c.game.sendMessage <- messages.MessageResponse{
				Audience:   messages.MessageAudienceUser,
				AudienceID: char.BelongsUserID,
				Type:       messages.MessageTypeCombatEnd,
				Message:    sb.String(),
			}
		}

		// Send inventory update so UI reflects new gold
		c.game.sendMessage <- messages.InventoryUpdateMessage{
			MessageResponse: messages.MessageResponse{
				Audience:   messages.MessageAudienceUser,
				AudienceID: char.BelongsUserID,
				Type:       messages.MessageTypeInventoryUpdate,
			},
			Inventory:     char.Inventory,
			EquippedItems: char.EquippedItems,
			Gold:          char.Gold,
		}
	}
}

// processCombatDefeat handles death penalties and sends the defeat message
func (c *CombatController) processCombatDefeat(instance *combat.CombatInstance) {
	for _, player := range instance.Players {
		char, err := c.game.Facade.CharactersService().FindByID(player.ID)
		if err != nil {
			continue
		}

		var sb strings.Builder
		sb.WriteString("\n═══════════════════════════════════════════════════\n")
		sb.WriteString("              DEFEAT\n")
		sb.WriteString("═══════════════════════════════════════════════════\n\n")
		sb.WriteString("You have been defeated!\n\n")

		// Gold loss penalty (10%)
		goldLoss := int64(float64(char.Gold) * 0.10)
		if goldLoss > 0 {
			char.Gold -= goldLoss
			sb.WriteString(fmt.Sprintf("PENALTY: Lost %d gold\n", goldLoss))
		}

		// Respawn with 50% HP
		char.CurrentHitPoints = char.MaxHitPoints / 2
		if char.CurrentHitPoints < 1 {
			char.CurrentHitPoints = 1
		}

		// Update the combatant HP so the later cleanup sync uses the right value
		for i := range instance.Players {
			if instance.Players[i].ID == player.ID {
				instance.Players[i].CurrentHP = char.CurrentHitPoints
				break
			}
		}

		sb.WriteString(fmt.Sprintf("\nYou awaken with %d/%d HP.\n", char.CurrentHitPoints, char.MaxHitPoints))
		sb.WriteString("═══════════════════════════════════════════════════")

		c.game.Facade.CharactersService().Update(player.ID, char)

		c.game.sendMessage <- messages.MessageResponse{
			Audience:   messages.MessageAudienceUser,
			AudienceID: char.BelongsUserID,
			Type:       messages.MessageTypeCombatEnd,
			Message:    sb.String(),
		}
	}
}

// QueuePlayerAction queues an action for a player's next auto-attack turn
func (c *CombatController) QueuePlayerAction(characterID string, action combat.CombatAction, targetID string) {
	instance := c.manager.GetInstanceByPlayerID(characterID)
	if instance == nil {
		return
	}

	player := instance.GetPlayerByID(characterID)
	if player == nil {
		return
	}

	player.QueuedAction = action
	player.QueuedTargetID = targetID
	c.engine.UpdateCombatant(instance, player)
}

// QueuePlayerSkill queues a skill for a player's next turn
func (c *CombatController) QueuePlayerSkill(characterID, skillID, targetID string) {
	instance := c.manager.GetInstanceByPlayerID(characterID)
	if instance == nil {
		return
	}

	player := instance.GetPlayerByID(characterID)
	if player == nil {
		return
	}

	player.QueuedAction = combat.CombatActionSkill
	player.QueuedSkillID = skillID
	player.QueuedTargetID = targetID
	c.engine.UpdateCombatant(instance, player)
}

// SetAutoAttackTarget sets the persistent auto-attack target for a player
func (c *CombatController) SetAutoAttackTarget(characterID string, targetID string) {
	instance := c.manager.GetInstanceByPlayerID(characterID)
	if instance == nil {
		return
	}

	player := instance.GetPlayerByID(characterID)
	if player == nil {
		return
	}

	player.AutoAttackTargetID = targetID
	c.engine.UpdateCombatant(instance, player)
}

// processPlayerAutoAttack processes a player's turn automatically
func (c *CombatController) processPlayerAutoAttack(instance *combat.CombatInstance, player *combat.CombatantRef) {
	if player == nil || !player.IsAlive || player.HasFled {
		return
	}

	// Check for queued action
	if player.QueuedAction != "" {
		switch player.QueuedAction {
		case combat.CombatActionFlee:
			result := c.engine.ProcessFlee(instance, player.ID)
			c.notifyPlayersInCombat(instance, result.Message)

		case combat.CombatActionDefend:
			result := c.engine.ProcessDefend(instance, player.ID)
			c.notifyPlayersInCombat(instance, result.Message)

		case combat.CombatActionSkill:
			skillResult := c.engine.ProcessSkill(instance, player.ID, player.QueuedSkillID, player.QueuedTargetID)
			for _, msg := range skillResult.Messages {
				c.notifyPlayersInCombat(instance, msg)
			}
			for _, diedID := range skillResult.TargetsDied {
				target := instance.GetCombatantByID(diedID)
				if target != nil && target.Type == combat.CombatantTypePlayer {
					c.syncPlayerHP(diedID, 0)
				}
			}

		case combat.CombatActionAttack:
			targetID := player.QueuedTargetID
			if targetID != "" {
				// Validate target is alive
				target := instance.GetCombatantByID(targetID)
				if target != nil && target.IsAlive {
					result := c.engine.ProcessAttack(instance, player.ID, targetID)
					c.notifyPlayersInCombat(instance, result.Message)
					// Update persistent auto-attack target
					player.AutoAttackTargetID = targetID
					c.engine.UpdateCombatant(instance, player)
					if result.TargetDied {
						if target.Type == combat.CombatantTypePlayer {
							c.syncPlayerHP(targetID, 0)
						}
					}
				} else {
					// Queued target is dead/invalid, fall through to auto-attack
					c.doAutoAttack(instance, player)
				}
			} else {
				c.doAutoAttack(instance, player)
			}
		}

		// Clear the queued action
		// Re-fetch player since it may have been updated
		playerRef := instance.GetPlayerByID(player.ID)
		if playerRef != nil {
			playerRef.QueuedAction = ""
			playerRef.QueuedTargetID = ""
			playerRef.QueuedSkillID = ""
			c.engine.UpdateCombatant(instance, playerRef)
		}
		return
	}

	// No queued action - auto-attack
	c.doAutoAttack(instance, player)
}

// doAutoAttack performs the default auto-attack for a player
func (c *CombatController) doAutoAttack(instance *combat.CombatInstance, player *combat.CombatantRef) {
	targetID := player.AutoAttackTargetID

	// Validate auto-attack target
	if targetID != "" {
		target := instance.GetCombatantByID(targetID)
		if target == nil || !target.IsAlive {
			targetID = "" // Target is dead/invalid, pick a new one
		}
	}

	// If no valid target, pick first living enemy
	if targetID == "" {
		livingEnemies := instance.GetLivingEnemies()
		if len(livingEnemies) > 0 {
			targetID = livingEnemies[0].ID
			player.AutoAttackTargetID = targetID
			c.engine.UpdateCombatant(instance, player)
		}
	}

	if targetID == "" {
		return // No enemies to attack
	}

	result := c.engine.ProcessAttack(instance, player.ID, targetID)
	c.notifyPlayersInCombat(instance, result.Message)

	if result.TargetDied {
		target := instance.GetCombatantByID(targetID)
		if target != nil && target.Type == combat.CombatantTypePlayer {
			c.syncPlayerHP(targetID, 0)
		}
	}
}

// Ensure CombatController implements CombatEngineCtrl
var _ def.CombatEngineCtrl = (*CombatController)(nil)
