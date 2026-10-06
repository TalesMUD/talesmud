package combat

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/combat"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
	"github.com/talesmud/talesmud/pkg/portraits"
)

// CombatConfig holds global combat configuration
type CombatConfig struct {
	TurnTimeoutSeconds       int     // Legacy absolute turn timeout (Default: 60); prefer DecisionWindowSeconds for player action wait
	DecisionWindowSeconds    int     // Player decision window before auto-attack (Default: 5)
	TurnBeatMs               int     // Authored windup/beat before next turn may resolve (Default: 1000)
	ReactionMs               int     // Post-resolve reaction pause (Default: 400)
	AFKAutoFleeAfterTurns    int     // Default: 3
	DeathGoldLossPercent     float64 // Default: 0.10 (10%)
	DeathRespawnHPPercent    float64 // Default: 0.50 (50%)
	FleeBaseChance           float64 // Default: 0.50 (50%)
	FleeDEXBonus             float64 // Per DEX point bonus (Default: 0.02)
	DefendBonusPercent       float64 // Default: 0.50 (50% defense boost)
	CriticalHitChance        float64 // Default: 0.05 (5%)
	CriticalHitMultiplier    float64 // Default: 2.0
	CombatTimeoutMinutes     int     // Absolute max combat length (Default: 15)
	IdleCombatTimeoutMinutes int     // Soft release after no action (Default: 5)
}

// DefaultConfig returns the default combat configuration
func DefaultConfig() *CombatConfig {
	return &CombatConfig{
		TurnTimeoutSeconds:       60,
		DecisionWindowSeconds:    5,
		TurnBeatMs:               1000,
		ReactionMs:               400,
		AFKAutoFleeAfterTurns:    3,
		DeathGoldLossPercent:     0.10,
		DeathRespawnHPPercent:    0.50,
		FleeBaseChance:           0.50,
		FleeDEXBonus:             0.02,
		DefendBonusPercent:       0.50,
		CriticalHitChance:        0.05,
		CriticalHitMultiplier:    2.0,
		CombatTimeoutMinutes:     15,
		IdleCombatTimeoutMinutes: 5,
	}
}

// BeatBudget returns the authored pause after a resolved action before the next turn may start.
func (c *CombatConfig) BeatBudget() time.Duration {
	if c == nil {
		return 1400 * time.Millisecond
	}
	ms := c.TurnBeatMs + c.ReactionMs
	if ms <= 0 {
		ms = 1400
	}
	return time.Duration(ms) * time.Millisecond
}

// DecisionWindow returns the player decision window duration.
func (c *CombatConfig) DecisionWindow() time.Duration {
	if c == nil || c.DecisionWindowSeconds <= 0 {
		return 5 * time.Second
	}
	return time.Duration(c.DecisionWindowSeconds) * time.Second
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

	// Calculate attack power from weapon + class primary attribute modifier
	attackPower := char.GetWeaponDamage() + int32(char.GetPrimaryAttackMod())
	if attackPower < 1 {
		attackPower = 1
	}

	ref := combat.CombatantRef{
		ID:          char.Entity.ID,
		Type:        combat.CombatantTypePlayer,
		Name:        char.Name,
		Initiative:  0, // Will be rolled
		IsAlive:     true,
		HasFled:     false,
		Level:       char.Level,
		ClassID:     char.Class.ID,
		Portrait:    portraits.ForPlayer(char),
		MaxHP:       char.MaxHitPoints,
		CurrentHP:   char.CurrentHitPoints,
		AttackPower: attackPower,
		Defense:     defense,
		STRMod:      char.GetPrimaryAttackMod(), // Uses class primary attribute (STR/DEX/INT/WIS)
		DEXMod:      char.GetDEXMod(),
		CONMod:      char.GetCONMod(),
		INTMod:      char.GetINTMod(),
		WISMod:      char.GetWISMod(),
		MaxMana:     char.MaxMana,
		CurrentMana: char.CurrentMana,
		ManaRegen:   char.CalculateManaRegen(),
	}
	if balance.IsWard(ref.ClassID) {
		ref.Grit = balance.OpeningGrit(ref.ClassID)
		ref.ClassID = "ward"
	}
	// Kit charges are armed by the skill button, not at combat start.
	bolt, rig := balance.BoltRigCharges(char.Class.ID)
	ref.BoltLeft = bolt
	ref.RigLeft = rig
	ref.RaceID = characters.CanonicalRaceID(char.Race.ID)
	if len(char.EquippedSkills) > 0 {
		ref.EquippedSkills = make([]string, len(char.EquippedSkills))
		copy(ref.EquippedSkills, char.EquippedSkills)
		ref.SkillCooldowns = make(map[string]int)
	}
	snapshotWeaponOnHit(&ref, char)
	return ref
}

// CreateCombatantFromNPC creates a CombatantRef from an NPC
func (e *Engine) CreateCombatantFromNPC(n *npc.NPC) combat.CombatantRef {
	var attackPower int32 = 1
	var defense int32 = 0
	var dexMod int = 0

	difficulty := ""
	if n.EnemyTrait != nil {
		attackPower = n.EnemyTrait.AttackPower
		defense = n.EnemyTrait.Defense
		difficulty = n.EnemyTrait.Difficulty
	}

	// Use level as a rough approximation for DEX modifier if not specified
	dexMod = int(n.Level) / 4

	ref := combat.CombatantRef{
		ID:          n.Entity.ID,
		Type:        combat.CombatantTypeNPC,
		Name:        n.GetDisplayName(),
		Portrait:    portraits.ForNPC(n),
		TemplateID:  n.TemplateID,
		Initiative:  0, // Will be rolled
		IsAlive:     true,
		HasFled:     false,
		Level:       n.Level,
		Difficulty:  difficulty,
		MaxHP:       n.MaxHitPoints,
		CurrentHP:   n.CurrentHitPoints,
		AttackPower: attackPower,
		Defense:     defense,
		STRMod:      int(n.Level) / 4, // Approximation
		DEXMod:      dexMod,
		CONMod:      int(n.Level) / 4, // Approximation
	}
	phases := balance.BossPhases(difficulty)
	if len(phases) > 0 {
		ref.BossPhase, ref.BossPhaseLabel, ref.BossPhaseCount = 1, phases[0].Label, len(phases)
	}
	return ref
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

// JoinCombat adds a character to an existing active combat instance.
// Preserves whose turn it currently is when rebuilding initiative order.
func (e *Engine) JoinCombat(instance *combat.CombatInstance, char *characters.Character) bool {
	if e == nil || instance == nil || char == nil || char.Entity == nil {
		return false
	}
	if instance.State != combat.CombatStateActive {
		return false
	}
	if instance.GetPlayerByID(char.Entity.ID) != nil {
		return false
	}
	if e.Manager != nil && e.Manager.IsPlayerInCombat(char.Entity.ID) {
		return false
	}

	currentID := ""
	if instance.CurrentTurnIdx >= 0 && instance.CurrentTurnIdx < len(instance.TurnOrder) {
		currentID = instance.TurnOrder[instance.CurrentTurnIdx].ID
	}

	combatant := e.CreateCombatantFromCharacter(char)
	e.RollInitiative(&combatant)
	instance.Players = append(instance.Players, combatant)
	if e.Manager != nil {
		e.Manager.RegisterPlayer(char.Entity.ID, instance.ID)
	}

	e.BuildTurnOrder(instance)
	if currentID != "" {
		for i, c := range instance.TurnOrder {
			if c.ID == currentID {
				instance.CurrentTurnIdx = i
				break
			}
		}
	}

	instance.AddLogEntry(combat.CombatLogEntry{
		ActorID:   char.Entity.ID,
		ActorName: char.Name,
		Action:    combat.CombatActionAttack,
		Message:   fmt.Sprintf("%s joins the fight!", char.Name),
	})

	log.WithFields(log.Fields{
		"instanceID":  instance.ID,
		"characterID": char.Entity.ID,
		"players":     len(instance.Players),
	}).Info("Player joined combat")

	return true
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
	Hit        bool
	Critical   bool
	Miss       bool
	Damage     int32
	TargetDied bool
	Roll       int
	ToHit      int
	TargetAC   int
	Message    string
}

// ProcessAttack handles an attack from attacker to target
func (e *Engine) ProcessAttack(instance *combat.CombatInstance, attackerID, targetID string) AttackResult {
	attacker := instance.GetCombatantByID(attackerID)
	if attacker == nil {
		return AttackResult{Miss: true, Message: "Invalid attacker or target"}
	}
	swings := balance.ClassSwings(attacker.ClassID)
	if swings < 1 {
		swings = 1
	}
	var last AttackResult
	var parts []string
	var total int32
	anyHit := false
	for i := 0; i < swings; i++ {
		last = e.processAttackSwingMult(instance, attackerID, targetID, 1, true)
		if last.Message != "" {
			parts = append(parts, last.Message)
		}
		total += last.Damage
		if last.Hit {
			anyHit = true
		}
		if last.TargetDied {
			break
		}
		attacker = instance.GetCombatantByID(attackerID)
		if attacker == nil || !attacker.IsAlive {
			break
		}
		target := instance.GetCombatantByID(targetID)
		if target == nil || !target.IsAlive {
			break
		}
	}
	last.Damage = total
	last.Hit = anyHit
	last.Miss = !anyHit
	if len(parts) > 0 {
		last.Message = strings.Join(parts, " ")
	}
	return last
}

func (e *Engine) processAttackSwing(instance *combat.CombatInstance, attackerID, targetID string) AttackResult {
	return e.processAttackSwingMult(instance, attackerID, targetID, 1, true)
}

func (e *Engine) processAttackSwingMult(instance *combat.CombatInstance, attackerID, targetID string, mult float64, logIt bool) AttackResult {
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

	redirected := false
	if guard := e.guardRedirect(instance, attacker, target); guard != nil {
		target = guard
		targetID = guard.ID
		redirected = true
	} else if stand := e.standRedirect(instance, attacker, target); stand != nil {
		target = stand
		targetID = stand.ID
	}

	// Smoke: the target misses their next swing. One swing, then it is gone.
	if target.SmokeMiss {
		target.SmokeMiss = false
		e.UpdateCombatant(instance, target)
		smoked := AttackResult{Miss: true, Message: fmt.Sprintf("%s swings at %s and misses. The smoke holds.", attacker.Name, target.Name)}
		if logIt {
			instance.AddLogEntry(combat.CombatLogEntry{
				ActorID:    attacker.ID,
				ActorName:  attacker.Name,
				Action:     combat.CombatActionAttack,
				TargetID:   target.ID,
				TargetName: target.Name,
				Result:     "miss",
				Message:    smoked.Message,
			})
		}
		return smoked
	}

	// Check dodge from status effects (e.g. Evasion)
	if dodgeChance := getDodgeBuffPercent(target); dodgeChance > 0 {
		dodgeRoll := rand.Intn(100) + 1
		if dodgeRoll <= int(dodgeChance*100) {
			dodgeResult := AttackResult{
				Miss:    true,
				Message: fmt.Sprintf("%s attacks %s but %s dodges!", attacker.Name, target.Name, target.Name),
			}
			instance.AddLogEntry(combat.CombatLogEntry{
				ActorID:    attacker.ID,
				ActorName:  attacker.Name,
				Action:     combat.CombatActionAttack,
				TargetID:   target.ID,
				TargetName: target.Name,
				Result:     "dodged",
				Message:    dodgeResult.Message,
			})
			return dodgeResult
		}
	}

	// Roll to hit: 1d20 + STR modifier + level-gap bonus.
	// Natural 1 always misses. Natural 20 always hits.
	mods := balance.LevelGapModifiers(attacker.Level, target.Level)
	roll := rand.Intn(20) + 1
	baseCrit := 0.05
	if e.Config != nil && e.Config.CriticalHitChance > 0 {
		baseCrit = e.Config.CriticalHitChance
	}
	critChance := baseCrit + mods.CritChanceDelta
	if critChance < 0 {
		critChance = 0
	}
	if critChance > 0.95 {
		critChance = 0.95
	}
	targetAC := 10 + int(target.Defense) + int(target.DefenseBonus)
	hit, crit, toHit := ResolveAttackRoll(roll, attacker.STRMod, mods.HitBonus, targetAC, critChance, baseCrit, rand.Float64())

	result := AttackResult{
		Roll:     roll,
		ToHit:    toHit,
		TargetAC: targetAC,
		Hit:      hit,
		Critical: crit,
		Miss:     !hit,
	}

	if !hit {
		if roll == 1 {
			result.Message = fmt.Sprintf("%s swings wildly at %s but completely misses!",
				attacker.Name, target.Name)
		} else if mods.HitBonus != 0 {
			result.Message = fmt.Sprintf("%s attacks %s but misses! (Roll: %d + %d %+d lvl = %d vs AC %d)",
				attacker.Name, target.Name, roll, attacker.STRMod, mods.HitBonus, toHit, targetAC)
		} else {
			result.Message = fmt.Sprintf("%s attacks %s but misses! (Roll: %d + %d = %d vs AC %d)",
				attacker.Name, target.Name, roll, attacker.STRMod, toHit, targetAC)
		}
		return result
	}

	// Calculate damage
	result.Damage = e.CalculateDamage(attacker, target, result.Critical)
	if mult > 0 && mult != 1 {
		scaled := int32(math.Round(float64(result.Damage) * mult))
		if scaled < 1 {
			scaled = 1
		}
		result.Damage = scaled
	} else if result.Hit && attacker.Level <= 1 && balance.IsWard(attacker.ClassID) && result.Damage > 0 {
		result.Damage += balance.StarterSwing(attacker.ClassID)
	}

	// Once per fight, the next landed blow is halved.
	braced := false
	if target.BraceLeft > 0 && result.Damage > 0 {
		target.BraceLeft--
		halved := int32(math.Round(float64(result.Damage) * 0.5))
		if halved < 1 {
			halved = 1
		}
		result.Damage = halved
		braced = true
		e.UpdateCombatant(instance, target)
	}

	// Mana shield absorption
	if shield := hasManaShield(target); shield != nil && shield.Value > 0 {
		absorbed := result.Damage
		if absorbed > shield.Value {
			absorbed = shield.Value
		}
		shield.Value -= absorbed
		result.Damage -= absorbed
		if shield.Value <= 0 {
			removeStatusEffectByID(target, shield.ID)
		}
	}

	// Apply damage to target
	target.CurrentHP -= result.Damage
	if target.CurrentHP <= 0 {
		target.CurrentHP = 0
		target.IsAlive = false
		result.TargetDied = true
	}

	// Update the target in the instance
	e.UpdateCombatant(instance, target)
	wardNote := e.applyWardSoak(instance, attacker, target, result.Damage, redirected, true)

	// Build message
	if result.Critical {
		result.Message = fmt.Sprintf("CRITICAL HIT! %s strikes %s for %d damage!",
			attacker.Name, target.Name, result.Damage)
	} else if mods.HitBonus != 0 {
		result.Message = fmt.Sprintf("%s hits %s for %d damage. (Roll: %d + %d %+d lvl = %d vs AC %d)",
			attacker.Name, target.Name, result.Damage, roll, attacker.STRMod, mods.HitBonus, toHit, targetAC)
	} else {
		result.Message = fmt.Sprintf("%s hits %s for %d damage. (Roll: %d + %d = %d vs AC %d)",
			attacker.Name, target.Name, result.Damage, roll, attacker.STRMod, toHit, targetAC)
	}

	if target.GlyphCut > 0 && result.Damage > 0 {
		result.Damage -= target.GlyphCut
		if result.Damage < 0 {
			result.Damage = 0
		}
		target.GlyphCut = 0
		e.UpdateCombatant(instance, target)
	}

	if braced {
		result.Message = "You brace. " + result.Message
	}
	if wardNote != "" {
		result.Message += " " + wardNote
	}
	if attacker.Enraged {
		result.Message = "Enraged! " + result.Message
	}
	if result.TargetDied {
		result.Message += fmt.Sprintf(" %s has been defeated!", target.Name)
	} else {
		result.Message += fmt.Sprintf(" (%d/%d HP)", target.CurrentHP, target.MaxHP)
	}
	e.refreshEnrage(instance, target)

	// Content-authored weapon on-hit DoT (refresh duration; no stack spam).
	if msg := e.applyWeaponOnHitDot(instance, attacker, target); msg != "" {
		result.Message += " " + msg
	}
	// Add to combat log
	logResult := "hit"
	if result.Critical {
		logResult = "critical"
	}
	if logIt {
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
	}

	e.applyScrapReturn(instance, attacker, target, result.Damage, result.Hit)
	return result
}

// CalculateDamage computes damage from attacker to target, accounting for status effect buffs
func (e *Engine) CalculateDamage(attacker, target *combat.CombatantRef, critical bool) int32 {
	// Base damage = AttackPower (includes weapon damage + STR for players)
	baseDamage := attacker.AttackPower

	// Apply attack buffs from status effects
	if atkBuff := getAttackBuffPercent(attacker); atkBuff != 0 {
		baseDamage = int32(float64(baseDamage) * (1 + atkBuff))
		if baseDamage < 1 {
			baseDamage = 1
		}
	}

	// Defense reduction with status effect buffs
	effectiveDefense := target.Defense
	if defBuff := getDefenseBuffPercent(target); defBuff != 0 {
		effectiveDefense = int32(float64(effectiveDefense) * (1 + defBuff))
		if effectiveDefense < 0 {
			effectiveDefense = 0
		}
	}
	reduction := effectiveDefense / 2

	// Final damage (minimum 1)
	damage := baseDamage - reduction
	if damage < 1 {
		damage = 1
	}

	// Level gap scales damage after defense and before the crit multiplier.
	// Equal levels leave the pre-gap number unchanged.
	damage = balance.ScaleDamage(attacker.Level, target.Level, damage)
	damage = balance.ScaleClassDamage(attacker.ClassID, target.ClassID, attacker.Level, target.Level, damage)
	damage = balance.ApplyRacialWeaponBonus(attacker.RaceID, attacker.WeaponSubType, damage)
	if attacker.HobbleRounds > 0 {
		hobbled := int32(math.Round(float64(damage) * 0.80))
		if hobbled < 1 {
			hobbled = 1
		}
		damage = hobbled
	}
	if attacker.Type == combat.CombatantTypeNPC {
		damage = balance.ScaleBossPhaseDamage(damage, attacker.Difficulty, attacker.BossPhase, attacker.Enraged)
	}

	// Critical hit doubles damage (or uses CriticalHitMultiplier when it is not 2).
	if critical {
		mult := 2.0
		if e.Config != nil && e.Config.CriticalHitMultiplier > 0 {
			mult = e.Config.CriticalHitMultiplier
		}
		if mult == 2 {
			damage *= 2
		} else {
			damage = int32(math.Round(float64(damage) * mult))
			if damage < 1 {
				damage = 1
			}
		}
	}

	return damage
}

// ResolveAttackRoll decides hit and crit for one d20 attack.
// extraRoll is in [0,1) and is used only for crit chances that are not a pure natural 20.
// roll 1 always misses. roll 20 always hits. critChance is the absolute crit chance
// (base + level-gap delta), already clamped by the caller into a sane range.
func ResolveAttackRoll(roll, strMod, hitBonus, targetAC int, critChance, baseCrit, extraRoll float64) (hit bool, crit bool, toHit int) {
	if baseCrit <= 0 {
		baseCrit = 0.05
	}
	if critChance < 0 {
		critChance = 0
	}
	if extraRoll < 0 {
		extraRoll = 0
	}
	toHit = roll + strMod + hitBonus
	if roll == 1 {
		return false, false, toHit
	}
	if roll == 20 {
		isCrit := true
		if critChance < baseCrit {
			isCrit = extraRoll < critChance/baseCrit
		}
		return true, isCrit, toHit
	}
	if toHit < targetAC {
		return false, false, toHit
	}
	if critChance > baseCrit && extraRoll < (critChance-baseCrit) {
		return true, true, toHit
	}
	return true, false, toHit
}

// UpdateCombatant updates a combatant's data in both the player/enemy list and turn order
func (e *Engine) UpdateCombatant(instance *combat.CombatInstance, updated *combat.CombatantRef) {
	e.refreshBossPhase(instance, updated)
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

		// Round-start processing: tick cooldowns and mana regen
		e.ProcessRoundStart(instance)

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

	// Idle soft-release (no resolved action for N minutes)
	if e.idleTimedOut(instance) {
		return combat.CombatStateTimeout
	}

	// Absolute max combat length backstop
	if time.Since(instance.CreatedAt).Minutes() >= float64(e.Config.CombatTimeoutMinutes) {
		return combat.CombatStateTimeout
	}

	return combat.CombatStateActive
}

// idleTimedOut reports whether combat has had no resolved action for IdleCombatTimeoutMinutes.
func (e *Engine) idleTimedOut(instance *combat.CombatInstance) bool {
	if e == nil || e.Config == nil || instance == nil {
		return false
	}
	mins := e.Config.IdleCombatTimeoutMinutes
	if mins <= 0 {
		return false
	}
	anchor := instance.LastActionAt
	if anchor.IsZero() {
		anchor = instance.CreatedAt
	}
	if anchor.IsZero() {
		return false
	}
	return time.Since(anchor).Minutes() >= float64(mins)
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

// NPCAttackStep is one enemy attack action: a wind-up, or the hit that follows it.
type NPCAttackStep struct {
	Telegraph bool
	Ability   string
	Message   string
	Attack    AttackResult
}

// StepNPCAttack spends a boss or elite action on a telegraph, or resolves the hit.
// A stored wind-up counts down and lands when it reaches zero. Enrage skips new wind-ups.
func (e *Engine) StepNPCAttack(instance *combat.CombatInstance, actorID, targetID string) NPCAttackStep {
	if instance == nil {
		return NPCAttackStep{Message: "Invalid combat"}
	}
	actor := instance.GetCombatantByID(actorID)
	if actor == nil {
		return NPCAttackStep{Message: "Invalid attacker"}
	}
	e.UpdateCombatant(instance, actor)
	e.refreshEnrage(instance, actor)

	if actor.TelegraphTurns > 0 {
		actor.TelegraphTurns--
		ability := actor.TelegraphAbility
		if actor.TelegraphTurns > 0 {
			e.UpdateCombatant(instance, actor)
			msg := fmt.Sprintf("%s is still winding up %s!", actor.Name, ability)
			instance.AddLogEntry(combat.CombatLogEntry{
				ActorID:   actor.ID,
				ActorName: actor.Name,
				Action:    combat.CombatActionAttack,
				Result:    "telegraph",
				Message:   msg,
			})
			return NPCAttackStep{Telegraph: true, Ability: ability, Message: msg}
		}
		actor.TelegraphAbility = ""
		e.UpdateCombatant(instance, actor)
		result := e.ProcessAttack(instance, actorID, targetID)
		if ability != "" && result.Message != "" {
			result.Message = ability + " lands! " + result.Message
		}
		return NPCAttackStep{Ability: ability, Message: result.Message, Attack: result}
	}

	turns := balance.BossTelegraphTurns(actor.Difficulty, actor.Enraged)
	if turns > 0 && targetID != "" {
		label := balance.BossPhaseTelegraphLabel(actor.Difficulty, actor.BossPhase)
		actor.TelegraphTurns = turns
		actor.TelegraphAbility = label
		e.UpdateCombatant(instance, actor)
		msg := fmt.Sprintf("%s is winding up %s!", actor.Name, label)
		instance.AddLogEntry(combat.CombatLogEntry{
			ActorID:   actor.ID,
			ActorName: actor.Name,
			Action:    combat.CombatActionAttack,
			Result:    "telegraph",
			Message:   msg,
		})
		return NPCAttackStep{Telegraph: true, Ability: label, Message: msg}
	}

	result := e.ProcessAttack(instance, actorID, targetID)
	return NPCAttackStep{Message: result.Message, Attack: result}
}

// refreshBossPhase runs on every stat update, including skills and status effects.
// Healing never reverses a phase; lethal hits do not announce a new chapter.
func (e *Engine) refreshBossPhase(instance *combat.CombatInstance, actor *combat.CombatantRef) {
	if instance == nil || actor == nil || actor.Type != combat.CombatantTypeNPC || !actor.IsAlive || actor.HasFled || actor.CurrentHP <= 0 || actor.MaxHP <= 0 {
		return
	}
	phases := balance.BossPhases(actor.Difficulty)
	if len(phases) == 0 {
		return
	}
	actor.BossPhaseCount = len(phases)
	if actor.BossPhase == 0 {
		actor.BossPhase, actor.BossPhaseLabel = 1, phases[0].Label
	}
	for actor.BossPhase < len(phases) {
		next := phases[actor.BossPhase]
		if float64(actor.CurrentHP)/float64(actor.MaxHP) > next.BelowHP {
			break
		}
		actor.BossPhase++
		actor.BossPhaseLabel = next.Label
		instance.AddLogEntry(combat.CombatLogEntry{
			ActorID: actor.ID, ActorName: actor.Name, Action: combat.CombatActionAttack,
			Result: "phase-enter", Message: fmt.Sprintf("%s enters phase %d: %s!", actor.Name, actor.BossPhase, next.Label),
		})
	}
}

func (e *Engine) refreshEnrage(instance *combat.CombatInstance, actor *combat.CombatantRef) {
	if e == nil || instance == nil || actor == nil || actor.Type != combat.CombatantTypeNPC || !actor.IsAlive || actor.CurrentHP <= 0 {
		return
	}
	if !balance.ShouldEnrage(actor.Difficulty, instance.Round, actor.CurrentHP, actor.MaxHP, actor.Enraged) {
		return
	}
	actor.Enraged = true
	if balance.BossTelegraphTurns(actor.Difficulty, true) == 0 {
		actor.TelegraphTurns = 0
		actor.TelegraphAbility = ""
	}
	e.UpdateCombatant(instance, actor)
	instance.AddLogEntry(combat.CombatLogEntry{
		ActorID:   actor.ID,
		ActorName: actor.Name,
		Action:    combat.CombatActionAttack,
		Result:    "enrage",
		Message:   fmt.Sprintf("%s becomes enraged!", actor.Name),
	})
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

// snapshotWeaponOnHit copies main-hand on-hit script id and optional attribute DoT
// onto the combatant. Game-set names/numbers stay in content attrs / Lua scripts.
func snapshotWeaponOnHit(ref *combat.CombatantRef, char *characters.Character) {
	if ref == nil || char == nil || char.EquippedItems == nil {
		return
	}
	weapon := char.EquippedItems[items.ItemSlotMainHand]
	if weapon == nil {
		return
	}
	ref.WeaponSubType = string(weapon.SubType)
	ref.OnHitScriptID = weapon.OnHitScriptID
	if ref.OnHitScriptID == "" && weapon.TemplateID != "" {
		// Instance may have dropped script id; TemplateID alone is not enough here
		// without a service lookup — content should stamp onHitScriptId onto templates
		// so CreateInstanceFromTemplate copies it. Attributes still drive the DoT.
	}
	dot := parseOnHitDotAttrs(weapon.Attributes)
	if dot.Active {
		ref.OnHitDot = dot
	}
}

func parseOnHitDotAttrs(attrs map[string]interface{}) combat.OnHitDot {
	var out combat.OnHitDot
	if attrs == nil {
		return out
	}
	// Nested map form: on_hit_dot: {id, name, damage, duration, message}
	if raw, ok := attrs["on_hit_dot"]; ok {
		if m, ok := raw.(map[string]interface{}); ok {
			out.ID = attrString(m["id"])
			out.Name = attrString(m["name"])
			out.Damage = attrInt32(m["damage"])
			out.Duration = int(attrInt32(m["duration"]))
			out.Message = attrString(m["message"])
		}
	}
	// Flat form fallback
	if out.ID == "" {
		out.ID = attrString(attrs["on_hit_dot_id"])
	}
	if out.Name == "" {
		out.Name = attrString(attrs["on_hit_dot_name"])
	}
	if out.Damage == 0 {
		out.Damage = attrInt32(attrs["on_hit_dot_damage"])
	}
	if out.Duration == 0 {
		out.Duration = int(attrInt32(attrs["on_hit_dot_duration"]))
	}
	if out.Message == "" {
		out.Message = attrString(attrs["on_hit_dot_message"])
	}
	if out.ID == "" && out.Name != "" {
		out.ID = strings.ToLower(strings.ReplaceAll(out.Name, " ", "_"))
	}
	if out.Name == "" {
		out.Name = out.ID
	}
	if out.Damage > 0 && out.Duration > 0 && out.ID != "" {
		out.Active = true
	}
	return out
}

func attrString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return fmt.Sprintf("%v", t)
	}
}

func attrInt32(v interface{}) int32 {
	if v == nil {
		return 0
	}
	switch t := v.(type) {
	case int32:
		return t
	case int:
		return int32(t)
	case int64:
		return int32(t)
	case float64:
		return int32(t)
	case float32:
		return int32(t)
	case string:
		n, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return 0
		}
		return int32(n)
	default:
		return 0
	}
}

func (e *Engine) applyInscribe(instance *combat.CombatInstance, attacker, target *combat.CombatantRef) string {
	if e == nil || instance == nil || attacker == nil || target == nil || !target.IsAlive {
		return ""
	}
	if !balance.IsRunecaster(attacker.ClassID) {
		return ""
	}
	se := combat.StatusEffect{
		ID:       uuid.New().String(),
		SkillID:  "inscribe",
		Name:     "Inscribe",
		Type:     "dot",
		Value:    4,
		Duration: 3,
		SourceID: attacker.ID,
	}
	e.applyStatusEffect(instance, target, se)
	e.UpdateCombatant(instance, target)
	return "You inscribe them. The rune burns (4/round, 3 rounds)."
}

func (e *Engine) applyWeaponOnHitDot(instance *combat.CombatInstance, attacker, target *combat.CombatantRef) string {
	if e == nil || instance == nil || attacker == nil || target == nil || !target.IsAlive {
		return ""
	}
	dot := attacker.OnHitDot
	if !dot.Active || dot.Damage < 1 || dot.Duration < 1 {
		return ""
	}
	se := combat.StatusEffect{
		ID:       uuid.New().String(),
		SkillID:  "onhit:" + dot.ID,
		Name:     dot.Name,
		Type:     "dot",
		Value:    dot.Damage,
		Duration: dot.Duration,
		SourceID: attacker.ID,
	}
	e.applyStatusEffect(instance, target, se)
	e.UpdateCombatant(instance, target)
	if dot.Message != "" {
		return dot.Message
	}
	return fmt.Sprintf("%s sears %s (%d/tick, %d rounds).", dot.Name, target.Name, dot.Damage, dot.Duration)
}

// ApplyStatusEffectFromScript exposes status-effect application for Lua on-hit scripts.
func (e *Engine) ApplyStatusEffectFromScript(instance *combat.CombatInstance, targetID string, se combat.StatusEffect) bool {
	if e == nil || instance == nil || targetID == "" {
		return false
	}
	target := instance.GetCombatantByID(targetID)
	if target == nil || !target.IsAlive {
		return false
	}
	if se.ID == "" {
		se.ID = uuid.New().String()
	}
	e.applyStatusEffect(instance, target, se)
	e.UpdateCombatant(instance, target)
	return true
}
