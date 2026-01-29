package combat

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
)

// CombatConfig holds global combat configuration
type CombatConfig struct {
	TurnTimeoutSeconds    int     // Default: 60
	AFKAutoFleeAfterTurns int     // Default: 3
	DeathGoldLossPercent  float64 // Default: 0.10 (10%)
	DeathRespawnHPPercent float64 // Default: 0.50 (50%)
	FleeBaseChance        float64 // Default: 0.50 (50%)
	FleeDEXBonus          float64 // Per DEX point bonus (Default: 0.02)
	DefendBonusPercent    float64 // Default: 0.50 (50% defense boost)
	CriticalHitChance     float64 // Default: 0.05 (5%)
	CriticalHitMultiplier float64 // Default: 2.0
	CombatTimeoutMinutes  int     // Default: 30
}

// DefaultConfig returns the default combat configuration
func DefaultConfig() *CombatConfig {
	return &CombatConfig{
		TurnTimeoutSeconds:    60,
		AFKAutoFleeAfterTurns: 3,
		DeathGoldLossPercent:  0.10,
		DeathRespawnHPPercent: 0.50,
		FleeBaseChance:        0.50,
		FleeDEXBonus:          0.02,
		DefendBonusPercent:    0.50,
		CriticalHitChance:     0.05,
		CriticalHitMultiplier: 2.0,
		CombatTimeoutMinutes:  30,
	}
}

// Engine handles combat logic and calculations
type Engine struct {
	Config  *CombatConfig
	Manager *Manager
}

// NewEngine creates a new combat engine
func NewEngine(manager *Manager, config *CombatConfig) *Engine {
	if config == nil {
		config = DefaultConfig()
	}
	return &Engine{
		Config:  config,
		Manager: manager,
	}
}

// CreateCombatantFromCharacter creates a CombatantRef from a Character
func (e *Engine) CreateCombatantFromCharacter(char *characters.Character) combat.CombatantRef {
	// Calculate defense from equipment
	defense := char.GetArmorDefense()

	// Calculate attack power from weapon
	attackPower := char.GetWeaponDamage() + int32(char.GetSTRMod())
	if attackPower < 1 {
		attackPower = 1
	}

	return combat.CombatantRef{
		ID:          char.Entity.ID,
		Type:        combat.CombatantTypePlayer,
		Name:        char.Name,
		Initiative:  0, // Will be rolled
		IsAlive:     true,
		HasFled:     false,
		MaxHP:       char.MaxHitPoints,
		CurrentHP:   char.CurrentHitPoints,
		AttackPower: attackPower,
		Defense:     defense,
		STRMod:      char.GetSTRMod(),
		DEXMod:      char.GetDEXMod(),
		CONMod:      char.GetCONMod(),
	}
}

// CreateCombatantFromNPC creates a CombatantRef from an NPC
func (e *Engine) CreateCombatantFromNPC(n *npc.NPC) combat.CombatantRef {
	var attackPower int32 = 1
	var defense int32 = 0
	var dexMod int = 0

	if n.EnemyTrait != nil {
		attackPower = n.EnemyTrait.AttackPower
		defense = n.EnemyTrait.Defense
	}

	// Use level as a rough approximation for DEX modifier if not specified
	dexMod = int(n.Level) / 4

	return combat.CombatantRef{
		ID:          n.Entity.ID,
		Type:        combat.CombatantTypeNPC,
		Name:        n.GetDisplayName(),
		Initiative:  0, // Will be rolled
		IsAlive:     true,
		HasFled:     false,
		MaxHP:       n.MaxHitPoints,
		CurrentHP:   n.CurrentHitPoints,
		AttackPower: attackPower,
		Defense:     defense,
		STRMod:      int(n.Level) / 4, // Approximation
		DEXMod:      dexMod,
		CONMod:      int(n.Level) / 4, // Approximation
	}
}

// RollInitiative rolls initiative (1d20 + DEX modifier) for a combatant
func (e *Engine) RollInitiative(c *combat.CombatantRef) int {
	roll := rand.Intn(20) + 1 // 1d20
	initiative := roll + c.DEXMod
	c.Initiative = initiative
	return initiative
}

// InitiateCombat creates a new combat instance with the given participants
func (e *Engine) InitiateCombat(roomID string, players []*characters.Character, enemies []*npc.NPC) *combat.CombatInstance {
	instance := e.Manager.CreateInstance(roomID)

	// Add players
	for _, char := range players {
		combatant := e.CreateCombatantFromCharacter(char)
		e.RollInitiative(&combatant)
		instance.Players = append(instance.Players, combatant)
		e.Manager.RegisterPlayer(char.Entity.ID, instance.ID)
	}

	// Add enemies
	for _, enemy := range enemies {
		combatant := e.CreateCombatantFromNPC(enemy)
		e.RollInitiative(&combatant)
		instance.Enemies = append(instance.Enemies, combatant)
		e.Manager.RegisterNPC(enemy.Entity.ID, instance.ID)
	}

	// Build turn order (all combatants sorted by initiative, highest first)
	e.BuildTurnOrder(instance)

	// Set state to active
	instance.State = combat.CombatStateActive
	instance.TurnStartTime = time.Now()
	instance.Round = 1

	// Log combat start
	instance.AddLogEntry(combat.CombatLogEntry{
		Action:  combat.CombatActionAttack, // Using attack as placeholder for "combat started"
		Message: fmt.Sprintf("Combat begins! Round %d", instance.Round),
	})

	log.WithFields(log.Fields{
		"instanceID": instance.ID,
		"players":    len(instance.Players),
		"enemies":    len(instance.Enemies),
		"roomID":     roomID,
	}).Info("Combat initiated")

	return instance
}

// BuildTurnOrder creates the turn order from all living combatants sorted by initiative
func (e *Engine) BuildTurnOrder(instance *combat.CombatInstance) {
	instance.TurnOrder = make([]combat.CombatantRef, 0)

	// Add living players
	for _, p := range instance.Players {
		if p.IsAlive && !p.HasFled {
			instance.TurnOrder = append(instance.TurnOrder, p)
		}
	}

	// Add living enemies
	for _, enemy := range instance.Enemies {
		if enemy.IsAlive {
			instance.TurnOrder = append(instance.TurnOrder, enemy)
		}
	}

	// Sort by initiative (highest first)
	sort.Slice(instance.TurnOrder, func(i, j int) bool {
		return instance.TurnOrder[i].Initiative > instance.TurnOrder[j].Initiative
	})

	instance.CurrentTurnIdx = 0
}

// AttackResult contains the result of an attack action
type AttackResult struct {
	Hit         bool
	Critical    bool
	Miss        bool
	Damage      int32
	TargetDied  bool
	Roll        int
	ToHit       int
	TargetAC    int
	Message     string
}

// ProcessAttack handles an attack from attacker to target
func (e *Engine) ProcessAttack(instance *combat.CombatInstance, attackerID, targetID string) AttackResult {
	attacker := instance.GetCombatantByID(attackerID)
	target := instance.GetCombatantByID(targetID)

	if attacker == nil || target == nil {
		return AttackResult{Miss: true, Message: "Invalid attacker or target"}
	}

	if !attacker.IsAlive {
		return AttackResult{Miss: true, Message: "Attacker is dead"}
	}

	if !target.IsAlive {
		return AttackResult{Miss: true, Message: "Target is already dead"}
	}

	// Roll to hit: 1d20 + STR modifier
	roll := rand.Intn(20) + 1
	toHit := roll + attacker.STRMod

	// Target AC = 10 + Defense + DefenseBonus
	targetAC := 10 + int(target.Defense) + int(target.DefenseBonus)

	result := AttackResult{
		Roll:     roll,
		ToHit:    toHit,
		TargetAC: targetAC,
	}

	// Critical hit on natural 20
	if roll == 20 {
		result.Hit = true
		result.Critical = true
	} else if roll == 1 {
		// Critical miss on natural 1
		result.Miss = true
		result.Message = fmt.Sprintf("%s swings wildly at %s but completely misses!",
			attacker.Name, target.Name)
		return result
	} else if toHit >= targetAC {
		result.Hit = true
	} else {
		result.Miss = true
		result.Message = fmt.Sprintf("%s attacks %s but misses! (Roll: %d + %d = %d vs AC %d)",
			attacker.Name, target.Name, roll, attacker.STRMod, toHit, targetAC)
		return result
	}

	// Calculate damage
	result.Damage = e.CalculateDamage(attacker, target, result.Critical)

	// Apply damage to target
	target.CurrentHP -= result.Damage
	if target.CurrentHP <= 0 {
		target.CurrentHP = 0
		target.IsAlive = false
		result.TargetDied = true
	}

	// Update the target in the instance
	e.UpdateCombatant(instance, target)

	// Build message
	if result.Critical {
		result.Message = fmt.Sprintf("CRITICAL HIT! %s strikes %s for %d damage!",
			attacker.Name, target.Name, result.Damage)
	} else {
		result.Message = fmt.Sprintf("%s hits %s for %d damage. (Roll: %d + %d = %d vs AC %d)",
			attacker.Name, target.Name, result.Damage, roll, attacker.STRMod, toHit, targetAC)
	}

	if result.TargetDied {
		result.Message += fmt.Sprintf(" %s has been defeated!", target.Name)
	} else {
		result.Message += fmt.Sprintf(" (%d/%d HP)", target.CurrentHP, target.MaxHP)
	}

	// Add to combat log
	logResult := "hit"
	if result.Critical {
		logResult = "critical"
	}
	instance.AddLogEntry(combat.CombatLogEntry{
		ActorID:    attacker.ID,
		ActorName:  attacker.Name,
		Action:     combat.CombatActionAttack,
		TargetID:   target.ID,
		TargetName: target.Name,
		Result:     logResult,
		Damage:     result.Damage,
		Message:    result.Message,
	})

	return result
}

// CalculateDamage computes damage from attacker to target
func (e *Engine) CalculateDamage(attacker, target *combat.CombatantRef, critical bool) int32 {
	// Base damage = AttackPower (includes weapon damage + STR for players)
	baseDamage := attacker.AttackPower

	// Defense reduction = target defense / 2
	reduction := target.Defense / 2

	// Final damage (minimum 1)
	damage := baseDamage - reduction
	if damage < 1 {
		damage = 1
	}

	// Critical hit doubles damage
	if critical {
		damage *= 2
	}

	return damage
}

// UpdateCombatant updates a combatant's data in both the player/enemy list and turn order
func (e *Engine) UpdateCombatant(instance *combat.CombatInstance, updated *combat.CombatantRef) {
	// Update in players or enemies list
	for i := range instance.Players {
		if instance.Players[i].ID == updated.ID {
			instance.Players[i] = *updated
			break
		}
	}
	for i := range instance.Enemies {
		if instance.Enemies[i].ID == updated.ID {
			instance.Enemies[i] = *updated
			break
		}
	}

	// Update in turn order
	instance.UpdateCombatantInTurnOrder(updated.ID)
}

// DefendResult contains the result of a defend action
type DefendResult struct {
	DefenseBonus int32
	Message      string
}

// ProcessDefend handles a defend action
func (e *Engine) ProcessDefend(instance *combat.CombatInstance, defenderID string) DefendResult {
	defender := instance.GetCombatantByID(defenderID)
	if defender == nil {
		return DefendResult{Message: "Invalid defender"}
	}

	// Calculate defense bonus (50% of current defense, minimum 2)
	bonus := int32(float64(defender.Defense) * e.Config.DefendBonusPercent)
	if bonus < 2 {
		bonus = 2
	}

	defender.DefenseBonus = bonus
	e.UpdateCombatant(instance, defender)

	result := DefendResult{
		DefenseBonus: bonus,
		Message:      fmt.Sprintf("%s takes a defensive stance! (+%d defense until next turn)", defender.Name, bonus),
	}

	instance.AddLogEntry(combat.CombatLogEntry{
		ActorID:   defender.ID,
		ActorName: defender.Name,
		Action:    combat.CombatActionDefend,
		Result:    "defended",
		Message:   result.Message,
	})

	return result
}

// FleeResult contains the result of a flee attempt
type FleeResult struct {
	Success bool
	Roll    int
	Chance  int
	Message string
}

// ProcessFlee handles a flee attempt
func (e *Engine) ProcessFlee(instance *combat.CombatInstance, fleeingID string) FleeResult {
	fleeing := instance.GetCombatantByID(fleeingID)
	if fleeing == nil {
		return FleeResult{Success: false, Message: "Invalid combatant"}
	}

	// Calculate flee chance: base + DEX bonus
	chance := e.Config.FleeBaseChance + (float64(fleeing.DEXMod) * e.Config.FleeDEXBonus)
	if chance > 0.95 {
		chance = 0.95 // Cap at 95%
	}
	if chance < 0.10 {
		chance = 0.10 // Minimum 10%
	}

	chancePercent := int(chance * 100)
	roll := rand.Intn(100) + 1 // 1-100

	result := FleeResult{
		Roll:   roll,
		Chance: chancePercent,
	}

	if roll <= chancePercent {
		// Success
		result.Success = true
		fleeing.HasFled = true
		e.UpdateCombatant(instance, fleeing)

		result.Message = fmt.Sprintf("%s successfully flees from combat! (Roll: %d <= %d%%)",
			fleeing.Name, roll, chancePercent)
	} else {
		// Failure - lose turn
		result.Success = false
		result.Message = fmt.Sprintf("%s tries to flee but fails! The enemies block the escape. (Roll: %d > %d%%)",
			fleeing.Name, roll, chancePercent)
	}

	instance.AddLogEntry(combat.CombatLogEntry{
		ActorID:   fleeing.ID,
		ActorName: fleeing.Name,
		Action:    combat.CombatActionFlee,
		Result:    map[bool]string{true: "fled", false: "blocked"}[result.Success],
		Message:   result.Message,
	})

	return result
}

// NextTurn advances to the next turn, handling round progression
func (e *Engine) NextTurn(instance *combat.CombatInstance) *combat.CombatantRef {
	// Clear defense bonus from the combatant whose turn just ended
	current := instance.GetCurrentTurnCombatant()
	if current != nil {
		current.DefenseBonus = 0
		e.UpdateCombatant(instance, current)
	}

	// Advance turn index
	instance.CurrentTurnIdx++

	// Check if we've completed a round
	if instance.CurrentTurnIdx >= len(instance.TurnOrder) {
		instance.Round++
		// Rebuild turn order (in case combatants died/fled)
		e.BuildTurnOrder(instance)

		if len(instance.TurnOrder) == 0 {
			return nil
		}

		instance.AddLogEntry(combat.CombatLogEntry{
			Message: fmt.Sprintf("--- Round %d ---", instance.Round),
		})
	}

	// Skip dead or fled combatants
	for instance.CurrentTurnIdx < len(instance.TurnOrder) {
		current := &instance.TurnOrder[instance.CurrentTurnIdx]
		if current.IsAlive && !current.HasFled {
			break
		}
		instance.CurrentTurnIdx++
	}

	// Check again if we've run out of combatants
	if instance.CurrentTurnIdx >= len(instance.TurnOrder) {
		return nil
	}

	// Reset turn timer
	instance.TurnStartTime = time.Now()
	instance.LastActionAt = time.Now()

	return instance.GetCurrentTurnCombatant()
}

// CheckCombatEnd checks if combat should end and returns the new state
func (e *Engine) CheckCombatEnd(instance *combat.CombatInstance) combat.CombatState {
	if instance.AllEnemiesDead() {
		return combat.CombatStateVictory
	}

	if instance.AllPlayersDead() {
		return combat.CombatStateDefeat
	}

	if instance.AllPlayersFled() {
		return combat.CombatStateFled
	}

	// Check for global timeout
	if time.Since(instance.CreatedAt).Minutes() >= float64(e.Config.CombatTimeoutMinutes) {
		return combat.CombatStateTimeout
	}

	return combat.CombatStateActive
}

// EndCombat finalizes a combat instance with the given result
func (e *Engine) EndCombat(instance *combat.CombatInstance, result combat.CombatState) {
	instance.State = result

	instance.AddLogEntry(combat.CombatLogEntry{
		Message: fmt.Sprintf("Combat ended: %s", result),
	})

	log.WithFields(log.Fields{
		"instanceID": instance.ID,
		"result":     result,
		"rounds":     instance.Round,
	}).Info("Combat ended")
}

// GetNPCAIAction determines what action an NPC should take
func (e *Engine) GetNPCAIAction(instance *combat.CombatInstance, npcCombatant *combat.CombatantRef, npcEntity *npc.NPC) (action combat.CombatAction, targetID string) {
	// Check flee threshold
	if npcEntity != nil && npcEntity.EnemyTrait != nil && npcEntity.EnemyTrait.FleeThreshold > 0 {
		hpPercent := float64(npcCombatant.CurrentHP) / float64(npcCombatant.MaxHP)
		if hpPercent <= npcEntity.EnemyTrait.FleeThreshold {
			return combat.CombatActionFlee, ""
		}
	}

	// Default: attack the player with lowest HP percentage
	var target *combat.CombatantRef
	var lowestHPPercent float64 = 2.0 // Higher than 100%

	for i := range instance.Players {
		p := &instance.Players[i]
		if p.IsAlive && !p.HasFled {
			hpPercent := float64(p.CurrentHP) / float64(p.MaxHP)
			if hpPercent < lowestHPPercent {
				lowestHPPercent = hpPercent
				target = p
			}
		}
	}

	if target != nil {
		return combat.CombatActionAttack, target.ID
	}

	// No valid target (shouldn't happen in active combat)
	return combat.CombatActionDefend, ""
}

