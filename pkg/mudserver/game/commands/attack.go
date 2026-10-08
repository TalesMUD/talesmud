package commands

import (
	"fmt"
	"strings"
	"sync"

	"github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/ruleset"
)

// threatArmed remembers a character who was warned about a specific enemy.
// The next attack on that enemy, or attack!, engages. In-memory only.
var threatArmed sync.Map

func attackIsForced(data string) bool {
	parts := strings.Fields(data)
	if len(parts) == 0 {
		return false
	}
	switch strings.ToLower(parts[0]) {
	case "attack!", "a!", "hit!":
		return true
	default:
		return false
	}
}

func threatWarnKey(charID, npcID string) string {
	return charID + "\x00" + npcID
}

// refuseOverlevelWarning blocks the first attack on an orange-or-worse enemy.
// attack! and a second attack on the same enemy proceed. Returns true if refused.
func refuseOverlevelWarning(game def.GameCtrl, message *messages.Message, target *npc.NPC, force bool) bool {
	if game == nil || message == nil || message.Character == nil || target == nil || target.Entity == nil {
		return false
	}
	if !target.IsEnemy() {
		return false
	}
	tier := balance.ThreatTier(message.Character.Level, target.Level)
	if !balance.ThreatNeedsWarning(tier) {
		return false
	}
	key := threatWarnKey(message.Character.ID, target.Entity.ID)
	if force {
		threatArmed.Delete(key)
		return false
	}
	if _, armed := threatArmed.Load(key); armed {
		threatArmed.Delete(key)
		return false
	}
	threatArmed.Store(key, true)
	name := target.GetDisplayName()
	if name == "" {
		name = target.Name
	}
	game.SendMessage() <- message.Reply(fmt.Sprintf("%s is much stronger than you. Type attack! or attack again to engage.", name))
	return true
}

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

	// Interrupt rest when attacking
	game.InterruptRest(message.Character)

	// Parse target from command: "attack goblin" or "a goblin"
	parts := strings.Fields(message.Data)
	targetName := ""
	if len(parts) >= 2 {
		targetName = strings.Join(parts[1:], " ")
	}

	// Check if player is already in combat
	if isInActiveCombat(game, message.Character, combatEngine) {
		// If the fight's origin room no longer matches the player's room,
		// the combat is orphaned (player walked away before the leave-guard).
		// Drop it so attack can target enemies actually present here.
		instance := combatEngine.GetCombatInstance(message.Character.ID)
		if instance != nil && instance.OriginRoomID != "" &&
			instance.OriginRoomID != message.Character.CurrentRoomID {
			combatEngine.EndCombatForPlayer(message.Character.ID)
			clearStaleCombatState(game, message.Character)
			combatEngine.ClearCombatGrace(message.Character.ID)
			game.SendMessage() <- message.Reply("Your previous fight is too far away — you break off.")
			// Fall through to initiate combat against the room target below.
		} else {
			// Player is in combat - this is an in-combat attack
			return command.handleInCombatAttack(game, message, combatEngine, targetName)
		}
	}

	// Not in combat. first_hostile picks the first hostile; ask does not.
	if targetName == "" {
		if ruleset.BareAttack() == ruleset.BareAttackFirst {
			targetName = firstHostileName(game, message.Character.CurrentRoomID)
		}
		if targetName == "" {
			game.SendMessage() <- message.Reply("Attack whom? Usage: attack <target>")
			return true
		}
	}

	// Player is not in combat - try to initiate combat
	return command.handleInitiateCombat(game, message, combatEngine, targetName, attackIsForced(message.Data))
}

// handleInitiateCombat handles attacking an NPC to start combat
func (command *AttackCommand) handleInitiateCombat(game def.GameCtrl, message *messages.Message, combatEngine def.CombatEngineCtrl, targetName string, force bool) bool {
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

	// Check if NPC is already in combat — same-room allies can join that fight
	if combatEngine.IsNPCInCombat(target.Entity.ID) {
		return command.handleJoinCombat(game, message, combatEngine, target, force)
	}

	// Post-combat grace: stop blender of sequential 1v1s
	if combatEngine.CombatGraceActive(message.Character.ID) {
		game.SendMessage() <- message.Reply("You catch your breath... (too soon to fight again)")
		return true
	}

	if refuseOverlevelWarning(game, message, target, force) {
		return true
	}

	// Same path as aggro-on-sight. Swarm pack, onAggro, combatStart, and the party nudge
	// live on BeginEngagement. The breath window and overlevel warning stay here.
	if combatEngine.BeginEngagement(message.Character.CurrentRoomID, message.Character, message.FromUser.ID, target, false) == nil {
		game.SendMessage() <- message.Reply("Failed to initiate combat.")
	}
	return true
}

func firstHostileName(game def.GameCtrl, roomID string) string {
	if game == nil || roomID == "" {
		return ""
	}
	mgr := game.GetNPCInstanceManager()
	if mgr == nil {
		return ""
	}
	for _, n := range mgr.GetInstancesInRoom(roomID) {
		if n == nil || n.IsDead || !n.IsEnemy() || strings.TrimSpace(n.Name) == "" {
			continue
		}
		return n.Name
	}
	return ""
}

// handleJoinCombat adds a same-room player to an existing fight against this NPC.
func (command *AttackCommand) handleJoinCombat(game def.GameCtrl, message *messages.Message, combatEngine def.CombatEngineCtrl, target *npc.NPC, force bool) bool {
	instance := combatEngine.GetCombatInstanceByNPC(target.Entity.ID)
	if instance == nil || instance.State != combat.CombatStateActive {
		game.SendMessage() <- message.Reply(fmt.Sprintf("%s is already in combat with someone else!", target.Name))
		return true
	}

	// Must be in the fight's origin room
	if instance.OriginRoomID == "" || instance.OriginRoomID != message.Character.CurrentRoomID {
		game.SendMessage() <- message.Reply(fmt.Sprintf("%s is already in combat with someone else!", target.Name))
		return true
	}

	// Already in a different combat
	if combatEngine.IsPlayerInCombat(message.Character.ID) {
		game.SendMessage() <- message.Reply("You are already in a different fight.")
		return true
	}

	// Already in this instance (shouldn't happen — isInActiveCombat would have caught it)
	if instance.GetPlayerByID(message.Character.ID) != nil {
		game.SendMessage() <- message.Reply("You are already in this fight.")
		return true
	}

	if refuseOverlevelWarning(game, message, target, force) {
		return true
	}

	if !combatEngine.JoinCombat(instance, message.Character) {
		game.SendMessage() <- message.Reply(fmt.Sprintf("Could not join the fight against %s.", target.Name))
		return true
	}

	message.Character.InCombat = true
	message.Character.CombatInstanceID = instance.ID
	_ = game.GetFacade().CharactersService().Update(message.Character.ID, message.Character)

	startMsg := fmt.Sprintf("\n%s\n%s\n\n",
		"═══════════════════════════════════════════════════",
		"              YOU JOIN THE FIGHT!")
	startMsg += fmt.Sprintf("You leap into the fray against %s!\n\n", target.Name)
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

	start := messages.NewCombatStartMessage(
		message.FromUser.ID,
		startMsg,
		combatViews(instance.Enemies, message.Character.Level),
		combatViews(instance.Players, message.Character.Level),
	)
	start.TargetID = target.Entity.ID
	game.SendMessage() <- start

	// Existing fighters need the new roster before the next action resolves.
	for _, fighter := range instance.Players {
		if fighter.ID == message.Character.ID || !fighter.IsAlive || fighter.HasFled {
			continue
		}
		ch, err := game.GetFacade().CharactersService().FindByID(fighter.ID)
		if err != nil || ch == nil || ch.BelongsUserID == "" {
			continue
		}
		roster := append(combatViews(instance.Players, fighter.Level), combatViews(instance.Enemies, fighter.Level)...)
		game.SendMessage() <- messages.NewCombatActionMessage(ch.BelongsUserID,
			fmt.Sprintf("%s joins the fight!", message.Character.Name),
			messages.CombatActionMessage{ActorID: message.Character.ID, ActorName: message.Character.Name, Action: "join", Combatants: roster})
	}

	combatEngine.SetAutoAttackTarget(message.Character.Entity.ID, target.Entity.ID)
	game.SendMessage() <- message.Reply("\nCombat is automatic. Commands: attack <target> (switch target) | defend | flee | status")

	roomMsg := messages.MessageResponse{
		Audience:   messages.MessageAudienceRoomWithoutOrigin,
		AudienceID: message.Character.CurrentRoomID,
		OriginID:   message.FromUser.ID,
		Type:       messages.MessageTypeDefault,
		Message:    fmt.Sprintf("%s joins the fight against %s!", message.Character.Name, target.Name),
	}
	game.SendMessage() <- roomMsg

	return true
}

func combatViews(refs []combat.CombatantRef, viewerLevel int32) []messages.CombatantView {
	out := make([]messages.CombatantView, 0, len(refs))
	for _, r := range refs {
		view := messages.CombatantView{
			ID: r.ID, Type: string(r.Type), Name: r.Name, Portrait: r.Portrait, HP: r.CurrentHP, MaxHP: r.MaxHP,
			Mana: r.CurrentMana, MaxMana: r.MaxMana, ClassID: r.ClassID, IsAlive: r.IsAlive, HasFled: r.HasFled, Level: r.Level,
			Telegraph: r.TelegraphAbility, Enraged: r.Enraged,
			BossPhase: r.BossPhase, BossPhaseLabel: r.BossPhaseLabel, BossPhaseCount: r.BossPhaseCount,
		}
		if r.Type == combat.CombatantTypeNPC {
			view.Threat = balance.ThreatTier(viewerLevel, r.Level)
		}
		out = append(out, view)
	}
	return out
}

// handleInCombatAttack handles an attack action during combat (queues target switch)
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

	// Get list of living enemies
	var livingEnemies []struct {
		id    string
		name  string
		hp    int32
		maxHP int32
	}
	for _, enemy := range instance.Enemies {
		if enemy.IsAlive {
			livingEnemies = append(livingEnemies, struct {
				id    string
				name  string
				hp    int32
				maxHP int32
			}{enemy.ID, enemy.Name, enemy.CurrentHP, enemy.MaxHP})
		}
	}

	if targetName == "" {
		targetID := ""
		if player := instance.GetPlayerByID(message.Character.Entity.ID); player != nil && player.AutoAttackTargetID != "" {
			for _, enemy := range livingEnemies {
				if enemy.id == player.AutoAttackTargetID {
					targetID = enemy.id
					break
				}
			}
		}
		if targetID == "" && len(livingEnemies) > 0 {
			targetID = livingEnemies[0].id
		}
		if targetID == "" {
			game.SendMessage() <- message.Reply("No enemies to attack!")
			return true
		}
		combatEngine.SetAutoAttackTarget(message.Character.Entity.ID, targetID)
		combatEngine.QueuePlayerAction(message.Character.Entity.ID, combat.CombatActionAttack, targetID)
		name := targetID
		for _, enemy := range livingEnemies {
			if enemy.id == targetID {
				name = enemy.name
				break
			}
		}
		game.SendMessage() <- message.Reply(fmt.Sprintf("You attack %s.", name))
		return true
	}

	// Resolve by ID, UI label (Name#N), or partial name — same rules as initiate.
	roomID := ""
	if message.Character != nil {
		roomID = message.Character.CurrentRoomID
	}
	targetID, _ := resolveInCombatTarget(game, roomID, instance.Enemies, targetName)

	if targetID == "" {
		labels := livingEnemyDisplayLabels(instance.Enemies)
		var targets []string
		for _, e := range livingEnemies {
			label := labels[e.id]
			if label == "" {
				label = e.name
			}
			targets = append(targets, fmt.Sprintf("%s (%d/%d HP)", label, e.hp, e.maxHP))
		}
		game.SendMessage() <- message.Reply(fmt.Sprintf("Invalid target '%s'. Available targets: %s", targetName, strings.Join(targets, ", ")))
		return true
	}

	// Queue the target switch - it will take effect on the player's next turn
	combatEngine.SetAutoAttackTarget(message.Character.Entity.ID, targetID)
	combatEngine.QueuePlayerAction(message.Character.Entity.ID, combat.CombatActionAttack, targetID)

	// Find the target name for the message
	targetDisplayName := targetName
	for _, e := range livingEnemies {
		if e.id == targetID {
			targetDisplayName = e.name
			break
		}
	}
	game.SendMessage() <- message.Reply(fmt.Sprintf("You switch your focus to %s.", targetDisplayName))
	game.SendMessage() <- messages.NewCombatActionMessage(message.FromUser.ID, "", messages.CombatActionMessage{
		ActorID: message.Character.ID, TargetID: targetID, Action: "focus",
	})

	return true
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
	if !isInActiveCombat(game, message.Character, combatEngine) {
		game.SendMessage() <- message.Reply("You are not in combat.")
		return true
	}

	// Queue defend for next turn
	combatEngine.QueuePlayerAction(message.Character.Entity.ID, combat.CombatActionDefend, "")
	game.SendMessage() <- message.Reply("You prepare to defend on your next turn.")

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
	if !isInActiveCombat(game, message.Character, combatEngine) {
		game.SendMessage() <- message.Reply("You are not in combat. There's nothing to flee from.")
		return true
	}

	// Queue flee for next turn
	combatEngine.QueuePlayerAction(message.Character.Entity.ID, combat.CombatActionFlee, "")
	game.SendMessage() <- message.Reply("You prepare to flee on your next turn.")

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
	if !isInActiveCombat(game, message.Character, combatEngine) {
		game.SendMessage() <- message.Reply("You are not in combat.")
		return true
	}

	status := combatEngine.GetCombatStatus(message.Character.Entity.ID)
	game.SendMessage() <- message.Reply(status)

	return true
}
