package game

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	combatpkg "github.com/talesmud/talesmud/pkg/mudserver/game/combat"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/mudserver/game/util"
)

const combatBreathGrace = 3 * time.Second

// CombatController wraps the combat engine and implements CombatEngineCtrl interface
type CombatController struct {
	manager    *combatpkg.Manager
	engine     *combatpkg.Engine
	game       *Game
	graceUntil map[string]time.Time // characterID -> grace expiry
}

// NewCombatController creates a new combat controller
func NewCombatController(game *Game) *CombatController {
	manager := combatpkg.NewManager()
	engine := combatpkg.NewEngine(manager, nil) // Uses default config

	return &CombatController{
		manager:    manager,
		engine:     engine,
		game:       game,
		graceUntil: make(map[string]time.Time),
	}
}

// CombatGraceActive reports whether the character is in the post-combat breath window.
func (c *CombatController) CombatGraceActive(characterID string) bool {
	if c == nil || characterID == "" || c.graceUntil == nil {
		return false
	}
	until, ok := c.graceUntil[characterID]
	if !ok {
		return false
	}
	if time.Now().Before(until) {
		return true
	}
	delete(c.graceUntil, characterID)
	return false
}

func (c *CombatController) markCombatGrace(characterID string) {
	if c == nil || characterID == "" {
		return
	}
	if c.graceUntil == nil {
		c.graceUntil = make(map[string]time.Time)
	}
	c.graceUntil[characterID] = time.Now().Add(combatBreathGrace)
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


// playerCombatQueueState snapshots queue + cooldowns + pacing clocks for one player.
func (c *CombatController) playerCombatQueueState(instance *combat.CombatInstance, playerID string) messages.CombatQueueState {
	state := messages.CombatQueueState{}
	if instance == nil {
		return state
	}
	if !instance.NextActionAt.IsZero() {
		state.NextActionAtMs = instance.NextActionAt.UnixMilli()
	}
	if !instance.DecisionDeadline.IsZero() {
		state.DecisionDeadlineMs = instance.DecisionDeadline.UnixMilli()
	}
	player := instance.GetPlayerByID(playerID)
	if player == nil {
		return state
	}
	if player.QueuedAction != "" {
		state.QueuedAction = string(player.QueuedAction)
		state.QueuedSkillID = player.QueuedSkillID
		state.QueuedTargetID = player.QueuedTargetID
	}
	cds := make(map[string]int)
	for k, v := range player.SkillCooldowns {
		if v > 0 {
			cds[k] = v
		}
	}
	state.SkillCooldowns = cds
	return state
}

// emitPlayerQueueUpdate pushes combatStatus so BattleStage can show queued chips / CD overlays immediately.
func (c *CombatController) emitPlayerQueueUpdate(instance *combat.CombatInstance, characterID, prose string) {
	if instance == nil {
		return
	}
	char, err := c.game.Facade.CharactersService().FindByID(characterID)
	if err != nil || char == nil {
		return
	}
	queue := c.playerCombatQueueState(instance, characterID)
	c.game.sendMessage <- messages.NewCombatStatusMessage(char.BelongsUserID, prose, instance.Round, queue)
}

// notifyPlayersInCombat sends a prose combat line to all living players (terminal/console path).
func (c *CombatController) notifyPlayersInCombat(instance *combat.CombatInstance, message string) {
	c.notifyCombatAction(instance, messages.CombatActionMessage{}, message)
}

// notifyCombatAction sends a structured combatAction (plus human Message) to living players.
func (c *CombatController) notifyCombatAction(instance *combat.CombatInstance, action messages.CombatActionMessage, prose string) {
	if prose == "" && action.MessageResponse.Message != "" {
		prose = action.MessageResponse.Message
	}
	if len(action.Combatants) == 0 {
		action.Combatants = combatantViewsFromInstance(instance)
	}
	for _, player := range instance.Players {
		if player.IsAlive && !player.HasFled {
			char, err := c.game.Facade.CharactersService().FindByID(player.ID)
			if err != nil {
				continue
			}
			payload := action
			payload.CombatQueueState = c.playerCombatQueueState(instance, player.ID)
			c.game.sendMessage <- messages.NewCombatActionMessage(char.BelongsUserID, prose, payload)
		}
	}
}

// emitCombatTurn notifies players that a combatant's turn (decision window) has started.
func (c *CombatController) emitCombatTurn(instance *combat.CombatInstance, actor *combat.CombatantRef, deadline time.Time) {
	if actor == nil {
		return
	}
	deadlineMs := int64(0)
	prose := fmt.Sprintf("Round %d — %s's turn.", instance.Round, actor.Name)
	if actor.Type == combat.CombatantTypePlayer {
		deadlineMs = deadline.UnixMilli()
		prose = fmt.Sprintf("Round %d — Your turn, %s! Choose an action (auto-attack in %ds).",
			instance.Round, actor.Name, c.engine.Config.DecisionWindowSeconds)
	}
	for _, player := range instance.Players {
		if !player.IsAlive || player.HasFled {
			continue
		}
		char, err := c.game.Facade.CharactersService().FindByID(player.ID)
		if err != nil {
			continue
		}
		msg := messages.NewCombatTurnMessage(
			char.BelongsUserID, prose, actor.ID, actor.Name, instance.Round, deadlineMs,
		)
		msg.CombatQueueState = c.playerCombatQueueState(instance, player.ID)
		// Prefer the live decision deadline on the turn payload when present.
		if msg.DecisionDeadlineMs == 0 && deadlineMs > 0 {
			msg.DecisionDeadlineMs = deadlineMs
		}
		if deadlineMs > 0 {
			msg.DeadlineMs = deadlineMs
		}
		c.game.sendMessage <- msg
	}
}

func combatantViewsFromInstance(instance *combat.CombatInstance) []messages.CombatantView {
	out := make([]messages.CombatantView, 0, len(instance.Players)+len(instance.Enemies))
	for _, p := range instance.Players {
		out = append(out, messages.CombatantView{ID: p.ID, Name: p.Name, Portrait: p.Portrait, HP: p.CurrentHP, MaxHP: p.MaxHP})
	}
	for _, e := range instance.Enemies {
		out = append(out, messages.CombatantView{ID: e.ID, Name: e.Name, Portrait: e.Portrait, HP: e.CurrentHP, MaxHP: e.MaxHP})
	}
	return out
}

func fxIDForAttack(result combatpkg.AttackResult, targetDied bool) string {
	if targetDied {
		return "death"
	}
	if result.Miss {
		return "miss"
	}
	if result.Critical {
		return "slash"
	}
	if result.Hit {
		return "slash"
	}
	return "miss"
}

func resultStringForAttack(result combatpkg.AttackResult) string {
	if result.Critical {
		return "crit"
	}
	if result.Miss {
		return "miss"
	}
	if result.Hit {
		return "hit"
	}
	return "miss"
}

// notifyAllPlayersInInstance sends a combatEnd to all players regardless of alive/fled status
func (c *CombatController) notifyAllPlayersInInstance(instance *combat.CombatInstance, message, outcome string) {
	for _, player := range instance.Players {
		char, err := c.game.Facade.CharactersService().FindByID(player.ID)
		if err != nil {
			continue
		}
		c.game.sendMessage <- messages.NewCombatEndMessage(char.BelongsUserID, message, outcome)
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
			c.cleanupCombatInstance(instance, combat.CombatStateTimeout)
		}
	}
}

// processAllTurns advances at most one combatant action per call, gated by authored beat budget.
// Player turns open a DecisionWindowSeconds window (combatTurn); timeout → auto-attack.
func (c *CombatController) processAllTurns(instance *combat.CombatInstance) {
	now := time.Now()

	// Pacing gate: wait out previous turn's beat/reaction budget
	if !instance.NextActionAt.IsZero() && now.Before(instance.NextActionAt) {
		return
	}

	current := instance.GetCurrentTurnCombatant()
	if current == nil {
		return
	}

	// Player decision window: announce once, then wait for queue or deadline
	if current.Type == combat.CombatantTypePlayer && current.IsAlive && !current.HasFled {
		player := instance.GetPlayerByID(current.ID)
		hasQueue := player != nil && player.QueuedAction != ""

		if instance.Phase != combat.CombatPhaseWaitingPlayer {
			instance.Phase = combat.CombatPhaseWaitingPlayer
			instance.TurnStartTime = now
			instance.DecisionDeadline = now.Add(c.engine.Config.DecisionWindow())
			c.emitCombatTurn(instance, current, instance.DecisionDeadline)
			// If already queued, resolve on the next eligible tick (small windup via NextActionAt)
			if hasQueue {
				instance.NextActionAt = now.Add(time.Duration(c.engine.Config.TurnBeatMs) * time.Millisecond)
			}
			return
		}

		if !hasQueue && now.Before(instance.DecisionDeadline) {
			return
		}
	}

	// Resolve exactly one turn this Update
	instance.Phase = combat.CombatPhaseResolving

	// Process status effects at start of turn (DoTs, HoTs, stun check)
	logLenBefore := len(instance.Log)
	stunned := c.engine.ProcessStatusEffects(instance, current)
	for j := logLenBefore; j < len(instance.Log); j++ {
		if instance.Log[j].Message != "" {
			c.notifyPlayersInCombat(instance, instance.Log[j].Message)
		}
	}

	current = instance.GetCurrentTurnCombatant()
	if current == nil || !current.IsAlive {
		endState := c.engine.CheckCombatEnd(instance)
		if endState != combat.CombatStateActive {
			c.engine.EndCombat(instance, endState)
			c.cleanupCombatInstance(instance, endState)
			return
		}
		c.finishTurnBeat(instance)
		return
	}

	if stunned {
		c.sendPlayerCharacterUpdate(instance)
		c.finishTurnBeat(instance)
		return
	}

	if current.Type == combat.CombatantTypeNPC {
		c.resolveNPCTurn(instance, current)
	} else {
		c.processPlayerAutoAttack(instance, current)
	}

	c.sendPlayerCharacterUpdate(instance)

	endState := c.engine.CheckCombatEnd(instance)
	if endState != combat.CombatStateActive {
		c.engine.EndCombat(instance, endState)
		c.cleanupCombatInstance(instance, endState)
		return
	}

	c.finishTurnBeat(instance)
}

// finishTurnBeat advances to the next combatant and applies the authored beat budget gate.
func (c *CombatController) finishTurnBeat(instance *combat.CombatInstance) {
	c.engine.NextTurn(instance)
	instance.Phase = combat.CombatPhasePlayingBeat
	instance.NextActionAt = time.Now().Add(c.engine.Config.BeatBudget())
	instance.DecisionDeadline = time.Time{}

	endState := c.engine.CheckCombatEnd(instance)
	if endState != combat.CombatStateActive {
		c.engine.EndCombat(instance, endState)
		c.cleanupCombatInstance(instance, endState)
	}
}

// resolveNPCTurn executes one NPC action and emits structured combatAction events.
func (c *CombatController) resolveNPCTurn(instance *combat.CombatInstance, current *combat.CombatantRef) {
	npcEntity := c.game.NPCManager.GetInstance(current.ID)
	action, targetID := c.engine.GetNPCAIAction(instance, current, npcEntity)

	switch action {
	case combat.CombatActionAttack:
		if targetID != "" {
			result := c.engine.ProcessAttack(instance, current.ID, targetID)
			target := instance.GetCombatantByID(targetID)
			remaining, maxHP := int32(0), int32(0)
			if target != nil {
				remaining, maxHP = target.CurrentHP, target.MaxHP
			}
			c.notifyCombatAction(instance, messages.CombatActionMessage{
				ActorID:     current.ID,
				ActorName:   current.Name,
				TargetID:    targetID,
				Action:      string(combat.CombatActionAttack),
				Result:      resultStringForAttack(result),
				Damage:      result.Damage,
				RemainingHP: remaining,
				MaxHP:       maxHP,
				FxID:        fxIDForAttack(result, result.TargetDied),
			}, result.Message)
			if result.TargetDied {
				if target != nil && target.Type == combat.CombatantTypePlayer {
					c.syncPlayerHP(targetID, 0)
				}
			}
		}
	case combat.CombatActionDefend:
		result := c.engine.ProcessDefend(instance, current.ID)
		c.notifyCombatAction(instance, messages.CombatActionMessage{
			ActorID:   current.ID,
			ActorName: current.Name,
			Action:    string(combat.CombatActionDefend),
			Result:    "defended",
			FxID:      "defend",
		}, result.Message)
	case combat.CombatActionFlee:
		result := c.engine.ProcessFlee(instance, current.ID)
		res := "blocked"
		fx := "flee"
		if result.Success {
			res = "fled"
		}
		c.notifyCombatAction(instance, messages.CombatActionMessage{
			ActorID:   current.ID,
			ActorName: current.Name,
			Action:    string(combat.CombatActionFlee),
			Result:    res,
			FxID:      fx,
		}, result.Message)
	}
}

// cleanupCombatInstance cleans up after combat ends, processes rewards, and notifies players
func (c *CombatController) cleanupCombatInstance(instance *combat.CombatInstance, endState combat.CombatState) {

	// Mark NPCs dead, then refresh the origin room, THEN send victory
	// text. The room update must not be queued after combatEnd or a
	// player who already left will see the fight room again.
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
	c.refreshOriginRoomAfterCombat(instance)

	// Process combat end based on state
	switch endState {
	case combat.CombatStateVictory:
		c.processCombatVictory(instance)
	case combat.CombatStateDefeat:
		c.processCombatDefeat(instance)
	case combat.CombatStateFled:
		c.notifyAllPlayersInInstance(instance, "\n═══════════════════════════════════════════════════\n              ESCAPED\n═══════════════════════════════════════════════════\n\nYou have fled from combat!\n═══════════════════════════════════════════════════", string(combat.CombatStateFled))
	case combat.CombatStateTimeout:
		c.notifyAllPlayersInInstance(instance, "Combat has timed out due to inactivity.", string(combat.CombatStateTimeout))
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
		c.markCombatGrace(player.ID)

		// Send updated stats to client (combat ended, final HP/XP/Gold)
		update := messages.NewCharacterUpdateMessage(char.BelongsUserID, char)
		if update != nil {
			c.game.sendMessage <- update
		}
	}

	// Remove the instance
	c.manager.RemoveInstance(instance.ID)

	log.WithFields(log.Fields{
		"instanceID": instance.ID,
		"endState":   endState,
	}).Info("Combat instance cleaned up")
}

func (c *CombatController) refreshOriginRoomAfterCombat(instance *combat.CombatInstance) {
	if instance == nil || instance.OriginRoomID == "" {
		return
	}
	room, err := c.game.Facade.RoomsService().FindByID(instance.OriginRoomID)
	if err != nil || room == nil {
		return
	}
	for _, player := range instance.Players {
		char, err := c.game.Facade.CharactersService().FindByID(player.ID)
		if err != nil || char == nil {
			continue
		}
		if char.CurrentRoomID != instance.OriginRoomID {
			continue
		}
		user, err := c.game.Facade.UsersService().FindByID(char.BelongsUserID)
		if err != nil {
			continue
		}
		roomUpdate := messages.NewRoomUpdateMessage(util.RoomWithCharacterReveals(room, char), user, c.game, char)
		roomUpdate.AudienceID = user.ID
		c.game.sendMessage <- roomUpdate
	}
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
			// Spawned/clone rats are deleted as soon as they die. Credit the
			// kill from the combatant snapshot so private cellars still count.
			npcData = npcFromCombatant(enemy)
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
			c.game.sendMessage <- messages.NewCombatEndMessage(char.BelongsUserID, sb.String(), string(combat.CombatStateVictory))

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
			c.game.sendMessage <- messages.NewCombatEndMessage(char.BelongsUserID, sb.String(), string(combat.CombatStateVictory))
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

		// XP loss penalty (10%)
		xpLoss := int32(float64(char.XP) * 0.10)
		if xpLoss > 0 {
			char.XP -= xpLoss
			if char.XP < 0 {
				char.XP = 0
			}
			sb.WriteString(fmt.Sprintf("PENALTY: Lost %d experience\n", xpLoss))
		}

		// Gold loss penalty (1 gold)
		if char.Gold > 0 {
			char.Gold -= 1
			sb.WriteString("PENALTY: Lost 1 gold\n")
		}

		if damaged := char.DamageEquippedArmor(); len(damaged) > 0 {
			sb.WriteString("Your armor is battered:\n")
			for _, name := range damaged {
				sb.WriteString("  - ")
				sb.WriteString(name)
				sb.WriteString("\n")
			}
			sb.WriteString("A merchant can repair it.\n")
		}

		// Respawn with 50% HP
		char.CurrentHitPoints = char.MaxHitPoints / 2
		if char.CurrentHitPoints < 1 {
			char.CurrentHitPoints = 1
		}

		if char.BoundRoomID != "" && char.BoundRoomID != char.CurrentRoomID {
			if boundRoom, ok := c.game.RelocateCharacter(char, char.BelongsUserID, char.BoundRoomID); ok {
				sb.WriteString(fmt.Sprintf("\nYou find yourself back at %s.\n", boundRoom.Name))
			}
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

		c.game.sendMessage <- messages.NewCombatEndMessage(char.BelongsUserID, sb.String(), string(combat.CombatStateDefeat))
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
	player.QueuedSkillID = ""
	c.engine.UpdateCombatant(instance, player)
	label := string(action)
	if label == "" {
		label = "action"
	}
	c.emitPlayerQueueUpdate(instance, characterID, fmt.Sprintf("Queued: %s", label))
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
	c.emitPlayerQueueUpdate(instance, characterID, fmt.Sprintf("Queued skill: %s", skillID))
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
			res := "blocked"
			if result.Success {
				res = "fled"
			}
			c.notifyCombatAction(instance, messages.CombatActionMessage{
				ActorID: player.ID, ActorName: player.Name,
				Action: string(combat.CombatActionFlee), Result: res, FxID: "flee",
			}, result.Message)

		case combat.CombatActionDefend:
			result := c.engine.ProcessDefend(instance, player.ID)
			c.notifyCombatAction(instance, messages.CombatActionMessage{
				ActorID: player.ID, ActorName: player.Name,
				Action: string(combat.CombatActionDefend), Result: "defended", FxID: "defend",
			}, result.Message)

		case combat.CombatActionSkill:
			skillResult := c.engine.ProcessSkill(instance, player.ID, player.QueuedSkillID, player.QueuedTargetID)
			targetID := player.QueuedTargetID
			if skillResult.TotalHeal > 0 && targetID == "" {
				targetID = player.ID // self-heal floats on caster
			}
			remaining, maxHP := int32(0), int32(0)
			if targetID != "" {
				if t := instance.GetCombatantByID(targetID); t != nil {
					remaining, maxHP = t.CurrentHP, t.MaxHP
				}
			}
			fx := "cast"
			if len(skillResult.TargetsDied) > 0 {
				fx = "death"
			}
			prose := strings.Join(skillResult.Messages, "\n")
			c.notifyCombatAction(instance, messages.CombatActionMessage{
				ActorID: player.ID, ActorName: player.Name,
				TargetID: targetID, Action: string(combat.CombatActionSkill),
				Result: "cast", Damage: skillResult.TotalDamage, Heal: skillResult.TotalHeal,
				RemainingHP: remaining, MaxHP: maxHP, FxID: fx,
			}, prose)
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
					c.notifyCombatAction(instance, messages.CombatActionMessage{
						ActorID: player.ID, ActorName: player.Name, TargetID: targetID,
						Action: string(combat.CombatActionAttack), Result: resultStringForAttack(result),
						Damage: result.Damage, RemainingHP: target.CurrentHP, MaxHP: target.MaxHP,
						FxID: fxIDForAttack(result, result.TargetDied),
					}, result.Message)
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
	target := instance.GetCombatantByID(targetID)
	remaining, maxHP := int32(0), int32(0)
	if target != nil {
		remaining, maxHP = target.CurrentHP, target.MaxHP
	}
	c.notifyCombatAction(instance, messages.CombatActionMessage{
		ActorID: player.ID, ActorName: player.Name, TargetID: targetID,
		Action: string(combat.CombatActionAttack), Result: resultStringForAttack(result),
		Damage: result.Damage, RemainingHP: remaining, MaxHP: maxHP,
		FxID: fxIDForAttack(result, result.TargetDied),
	}, result.Message)

	if result.TargetDied {
		if target != nil && target.Type == combat.CombatantTypePlayer {
			c.syncPlayerHP(targetID, 0)
		}
	}
}

// Ensure CombatController implements CombatEngineCtrl
var _ def.CombatEngineCtrl = (*CombatController)(nil)

func npcFromCombatant(enemy combat.CombatantRef) *npc.NPC {
	return &npc.NPC{
		Entity:     &entities.Entity{ID: enemy.ID},
		Name:       enemy.Name,
		TemplateID: enemy.TemplateID,
	}
}
