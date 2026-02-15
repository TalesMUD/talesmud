package combat

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/talesmud/talesmud/pkg/entities/combat"
	"github.com/talesmud/talesmud/pkg/entities/skills"
)

// --- Attribute and Status Effect Helpers ---

// getAttrMod returns the attribute modifier for a combatant based on attribute name
func getAttrMod(c *combat.CombatantRef, attr string) int {
	switch strings.ToUpper(attr) {
	case "STR":
		return c.STRMod
	case "DEX":
		return c.DEXMod
	case "CON":
		return c.CONMod
	case "INT":
		return c.INTMod
	case "WIS":
		return c.WISMod
	default:
		return 0
	}
}

// getAttackBuffPercent returns total attack buff/debuff percent from status effects
func getAttackBuffPercent(c *combat.CombatantRef) float64 {
	var total float64
	for _, se := range c.StatusEffects {
		if se.Stat == "attack" {
			total += se.Percent
		}
	}
	return total
}

// getDefenseBuffPercent returns total defense buff/debuff percent from status effects
func getDefenseBuffPercent(c *combat.CombatantRef) float64 {
	var total float64
	for _, se := range c.StatusEffects {
		if se.Stat == "defense" {
			total += se.Percent
		}
	}
	return total
}

// getDodgeBuffPercent returns total dodge buff percent from status effects
func getDodgeBuffPercent(c *combat.CombatantRef) float64 {
	var total float64
	for _, se := range c.StatusEffects {
		if se.Stat == "dodge" {
			total += se.Percent
		}
	}
	return total
}

// isStunned returns true if the combatant has an active stun effect
func isStunned(c *combat.CombatantRef) bool {
	for _, se := range c.StatusEffects {
		if se.Type == "stun" {
			return true
		}
	}
	return false
}

// hasManaShield returns the mana shield StatusEffect if active, nil otherwise
func hasManaShield(c *combat.CombatantRef) *combat.StatusEffect {
	for i := range c.StatusEffects {
		if c.StatusEffects[i].Stat == "mana_shield" {
			return &c.StatusEffects[i]
		}
	}
	return nil
}

// removeStatusEffectByID removes a status effect by its ID
func removeStatusEffectByID(c *combat.CombatantRef, id string) {
	effects := make([]combat.StatusEffect, 0, len(c.StatusEffects))
	for _, se := range c.StatusEffects {
		if se.ID != id {
			effects = append(effects, se)
		}
	}
	c.StatusEffects = effects
}

// --- Skill Processing ---

// SkillResult contains the result of using a skill
type SkillResult struct {
	Success     bool
	SkillName   string
	Messages    []string
	TotalDamage int32
	TotalHeal   int32
	TargetsDied []string
}

// ProcessSkill handles a combatant using a skill in combat
func (e *Engine) ProcessSkill(instance *combat.CombatInstance, casterID, skillID, targetID string) SkillResult {
	caster := instance.GetCombatantByID(casterID)
	if caster == nil || !caster.IsAlive {
		return SkillResult{Success: false, Messages: []string{"Invalid or dead caster"}}
	}

	skill := skills.SkillByID(skillID)
	if skill == nil {
		return SkillResult{Success: false, Messages: []string{"Unknown skill"}}
	}

	// Verify skill is equipped
	equipped := false
	for _, sid := range caster.EquippedSkills {
		if sid == skillID {
			equipped = true
			break
		}
	}
	if !equipped {
		return SkillResult{Success: false, Messages: []string{fmt.Sprintf("%s is not equipped", skill.Name)}}
	}

	// Check resource availability
	if skill.ResourceType == skills.ResourceMana {
		if caster.CurrentMana < skill.ManaCost {
			return SkillResult{Success: false, Messages: []string{
				fmt.Sprintf("Not enough mana for %s (need %d, have %d)", skill.Name, skill.ManaCost, caster.CurrentMana),
			}}
		}
	} else if skill.ResourceType == skills.ResourceCooldown {
		if caster.SkillCooldowns != nil {
			if cd, ok := caster.SkillCooldowns[skillID]; ok && cd > 0 {
				return SkillResult{Success: false, Messages: []string{
					fmt.Sprintf("%s is on cooldown (%d rounds remaining)", skill.Name, cd),
				}}
			}
		}
	}

	// Deduct resource
	if skill.ResourceType == skills.ResourceMana {
		caster.CurrentMana -= skill.ManaCost
	} else if skill.ResourceType == skills.ResourceCooldown {
		if caster.SkillCooldowns == nil {
			caster.SkillCooldowns = make(map[string]int)
		}
		caster.SkillCooldowns[skillID] = skill.CooldownRounds
	}

	result := SkillResult{Success: true, SkillName: skill.Name}

	// Calculate scaled power
	attrMod := getAttrMod(caster, skill.ScalingAttr)
	basePower := skill.BasePower + int32(float64(attrMod)*skill.ScalingFactor)
	if basePower < 1 {
		basePower = 1
	}

	// Special: Nature's Gift heals based on max HP
	if skill.ScalingAttr == "maxhp" {
		basePower = int32(float64(caster.MaxHP) * skill.ScalingFactor)
		if basePower < 1 {
			basePower = 1
		}
	}

	// Apply caster's attack buffs to damage skills
	if skill.Effect == skills.EffectDamage || skill.Effect == skills.EffectDot {
		if atkBuff := getAttackBuffPercent(caster); atkBuff != 0 {
			basePower = int32(float64(basePower) * (1 + atkBuff))
			if basePower < 1 {
				basePower = 1
			}
		}
	}

	// Resolve primary effect
	switch skill.Effect {
	case skills.EffectDamage:
		e.resolveSkillDamage(instance, caster, skill, basePower, targetID, &result)
	case skills.EffectHeal:
		e.resolveSkillHeal(instance, caster, skill, basePower, &result)
	case skills.EffectBuff:
		e.resolveSkillBuff(instance, caster, skill, &result)
	case skills.EffectDebuff:
		e.resolveSkillDebuff(instance, caster, skill, targetID, &result)
	case skills.EffectDot:
		e.resolveSkillDot(instance, caster, skill, basePower, targetID, &result)
	case skills.EffectHot:
		e.resolveSkillHot(instance, caster, skill, basePower, &result)
	}

	// Handle secondary effects (e.g. Holy Strike: damage + heal, Pin Down: damage + debuff)
	if skill.SecondaryEffect != "" {
		e.resolveSecondaryEffect(instance, caster, skill, targetID, &result)
	}

	// Special: Berserker Rage also debuffs self defense
	if skillID == "warrior_berserker_rage" {
		e.applyStatusEffect(instance, caster, combat.StatusEffect{
			ID:       uuid.New().String(),
			SkillID:  skillID,
			Name:     "Berserker Rage (Defense Down)",
			Type:     "debuff",
			Stat:     "defense",
			Percent:  -0.25,
			Duration: 3,
			SourceID: caster.ID,
		})
		result.Messages = append(result.Messages, fmt.Sprintf("%s's defense drops by 25%% during the rage!", caster.Name))
	}

	// Special: Shield Bash applies stun
	if skillID == "warrior_shield_bash" && skill.Duration > 0 {
		target := instance.GetCombatantByID(targetID)
		if target != nil && target.IsAlive {
			e.applyStatusEffect(instance, target, combat.StatusEffect{
				ID:       uuid.New().String(),
				SkillID:  skillID,
				Name:     "Stunned",
				Type:     "stun",
				Duration: skill.Duration,
				SourceID: caster.ID,
			})
			result.Messages = append(result.Messages, fmt.Sprintf("%s is stunned for %d round!", target.Name, skill.Duration))
			e.UpdateCombatant(instance, target)
		}
	}

	// Update caster state (mana/cooldowns changed)
	e.UpdateCombatant(instance, caster)

	// Add combat log entries
	for _, msg := range result.Messages {
		instance.AddLogEntry(combat.CombatLogEntry{
			ActorID:   caster.ID,
			ActorName: caster.Name,
			Action:    combat.CombatActionSkill,
			TargetID:  targetID,
			Damage:    result.TotalDamage,
			Message:   msg,
			Result:    "skill",
		})
	}

	return result
}

// --- Skill Effect Resolution ---

func (e *Engine) resolveSkillDamage(instance *combat.CombatInstance, caster *combat.CombatantRef, skill *skills.Skill, basePower int32, targetID string, result *SkillResult) {
	targets := e.getSkillTargets(instance, caster, skill, targetID)

	hitCount := 1
	if skill.HitCount > 0 {
		hitCount = skill.HitCount
	}

	for _, target := range targets {
		totalDmg := int32(0)
		for hit := 0; hit < hitCount; hit++ {
			damage := basePower

			// Apply defense reduction unless skill ignores defense
			if !skill.IgnoresDefense {
				effectiveDefense := target.Defense
				if defBuff := getDefenseBuffPercent(target); defBuff != 0 {
					effectiveDefense = int32(float64(effectiveDefense) * (1 + defBuff))
					if effectiveDefense < 0 {
						effectiveDefense = 0
					}
				}
				damage -= effectiveDefense / 2
			}
			if damage < 1 {
				damage = 1
			}

			// Mana shield absorption
			if shield := hasManaShield(target); shield != nil && shield.Value > 0 {
				absorbed := damage
				if absorbed > shield.Value {
					absorbed = shield.Value
				}
				shield.Value -= absorbed
				damage -= absorbed
				if damage < 0 {
					damage = 0
				}
				if shield.Value <= 0 {
					removeStatusEffectByID(target, shield.ID)
					result.Messages = append(result.Messages, fmt.Sprintf("%s's Mana Shield shatters!", target.Name))
				}
			}

			target.CurrentHP -= damage
			totalDmg += damage
			result.TotalDamage += damage
		}

		if target.CurrentHP <= 0 {
			target.CurrentHP = 0
			target.IsAlive = false
			result.TargetsDied = append(result.TargetsDied, target.ID)
		}
		e.UpdateCombatant(instance, target)

		// Build message
		msg := ""
		if hitCount > 1 {
			msg = fmt.Sprintf("%s uses %s on %s for %d damage (%d hits)!",
				caster.Name, skill.Name, target.Name, totalDmg, hitCount)
		} else {
			msg = fmt.Sprintf("%s casts %s on %s for %d damage!",
				caster.Name, skill.Name, target.Name, totalDmg)
		}
		if !target.IsAlive {
			msg += fmt.Sprintf(" %s has been defeated!", target.Name)
		} else {
			msg += fmt.Sprintf(" (%d/%d HP)", target.CurrentHP, target.MaxHP)
		}
		result.Messages = append(result.Messages, msg)
	}
}

func (e *Engine) resolveSkillHeal(instance *combat.CombatInstance, caster *combat.CombatantRef, skill *skills.Skill, basePower int32, result *SkillResult) {
	healing := basePower
	caster.CurrentHP += healing
	if caster.CurrentHP > caster.MaxHP {
		caster.CurrentHP = caster.MaxHP
	}
	result.TotalHeal += healing
	e.UpdateCombatant(instance, caster)

	result.Messages = append(result.Messages, fmt.Sprintf("%s casts %s, restoring %d HP! (%d/%d HP)",
		caster.Name, skill.Name, healing, caster.CurrentHP, caster.MaxHP))
}

func (e *Engine) resolveSkillBuff(instance *combat.CombatInstance, caster *combat.CombatantRef, skill *skills.Skill, result *SkillResult) {
	se := combat.StatusEffect{
		ID:       uuid.New().String(),
		SkillID:  skill.ID,
		Name:     skill.Name,
		Type:     "buff",
		Stat:     skill.BuffStat,
		Percent:  skill.BuffPercent,
		Duration: skill.Duration,
		SourceID: caster.ID,
	}

	// Mana Shield stores absorption amount in Value field
	if skill.BuffStat == "mana_shield" {
		attrMod := getAttrMod(caster, skill.ScalingAttr)
		se.Value = skill.BasePower + int32(float64(attrMod)*skill.ScalingFactor)
	}

	e.applyStatusEffect(instance, caster, se)
	e.UpdateCombatant(instance, caster)

	if skill.BuffStat == "mana_shield" {
		result.Messages = append(result.Messages, fmt.Sprintf("%s creates a Mana Shield absorbing up to %d damage!",
			caster.Name, se.Value))
	} else {
		percentStr := fmt.Sprintf("+%.0f%%", skill.BuffPercent*100)
		result.Messages = append(result.Messages, fmt.Sprintf("%s casts %s! (%s %s for %d rounds)",
			caster.Name, skill.Name, percentStr, skill.BuffStat, skill.Duration))
	}
}

func (e *Engine) resolveSkillDebuff(instance *combat.CombatInstance, caster *combat.CombatantRef, skill *skills.Skill, targetID string, result *SkillResult) {
	target := instance.GetCombatantByID(targetID)
	if target == nil || !target.IsAlive {
		result.Messages = append(result.Messages, "Invalid target for debuff")
		return
	}

	se := combat.StatusEffect{
		ID:       uuid.New().String(),
		SkillID:  skill.ID,
		Name:     skill.Name,
		Type:     "debuff",
		Stat:     skill.BuffStat,
		Percent:  skill.BuffPercent,
		Duration: skill.Duration,
		SourceID: caster.ID,
	}

	e.applyStatusEffect(instance, target, se)
	e.UpdateCombatant(instance, target)

	percentStr := fmt.Sprintf("%.0f%%", skill.BuffPercent*100)
	result.Messages = append(result.Messages, fmt.Sprintf("%s casts %s on %s! (%s %s for %d rounds)",
		caster.Name, skill.Name, target.Name, percentStr, skill.BuffStat, skill.Duration))
}

func (e *Engine) resolveSkillDot(instance *combat.CombatInstance, caster *combat.CombatantRef, skill *skills.Skill, basePower int32, targetID string, result *SkillResult) {
	target := instance.GetCombatantByID(targetID)
	if target == nil || !target.IsAlive {
		result.Messages = append(result.Messages, "Invalid target for DoT")
		return
	}

	se := combat.StatusEffect{
		ID:       uuid.New().String(),
		SkillID:  skill.ID,
		Name:     skill.Name,
		Type:     "dot",
		Value:    basePower,
		Duration: skill.Duration,
		SourceID: caster.ID,
	}

	e.applyStatusEffect(instance, target, se)
	e.UpdateCombatant(instance, target)

	result.Messages = append(result.Messages, fmt.Sprintf("%s casts %s on %s! (%d damage per round for %d rounds)",
		caster.Name, skill.Name, target.Name, basePower, skill.Duration))
}

func (e *Engine) resolveSkillHot(instance *combat.CombatInstance, caster *combat.CombatantRef, skill *skills.Skill, basePower int32, result *SkillResult) {
	se := combat.StatusEffect{
		ID:       uuid.New().String(),
		SkillID:  skill.ID,
		Name:     skill.Name,
		Type:     "hot",
		Value:    basePower,
		Duration: skill.Duration,
		SourceID: caster.ID,
	}

	e.applyStatusEffect(instance, caster, se)
	e.UpdateCombatant(instance, caster)

	result.Messages = append(result.Messages, fmt.Sprintf("%s casts %s! (Healing %d per round for %d rounds)",
		caster.Name, skill.Name, basePower, skill.Duration))
}

func (e *Engine) resolveSecondaryEffect(instance *combat.CombatInstance, caster *combat.CombatantRef, skill *skills.Skill, targetID string, result *SkillResult) {
	attrMod := getAttrMod(caster, skill.ScalingAttr)
	secondaryPower := skill.SecondaryBasePower + int32(float64(attrMod)*skill.SecondaryScaling)
	if secondaryPower < 1 {
		secondaryPower = 1
	}

	switch skill.SecondaryEffect {
	case skills.EffectHeal:
		if skill.SecondaryTarget == skills.TargetSelf {
			caster.CurrentHP += secondaryPower
			if caster.CurrentHP > caster.MaxHP {
				caster.CurrentHP = caster.MaxHP
			}
			result.TotalHeal += secondaryPower
			e.UpdateCombatant(instance, caster)
			result.Messages = append(result.Messages, fmt.Sprintf("%s is healed for %d HP! (%d/%d HP)",
				caster.Name, secondaryPower, caster.CurrentHP, caster.MaxHP))
		}

	case skills.EffectDebuff:
		target := instance.GetCombatantByID(targetID)
		if target != nil && target.IsAlive {
			se := combat.StatusEffect{
				ID:       uuid.New().String(),
				SkillID:  skill.ID,
				Name:     skill.Name + " (debuff)",
				Type:     "debuff",
				Stat:     skill.BuffStat,
				Percent:  skill.BuffPercent,
				Duration: skill.Duration,
				SourceID: caster.ID,
			}
			e.applyStatusEffect(instance, target, se)
			e.UpdateCombatant(instance, target)
			result.Messages = append(result.Messages, fmt.Sprintf("%s's %s is reduced by %.0f%% for %d rounds!",
				target.Name, skill.BuffStat, -skill.BuffPercent*100, skill.Duration))
		}
	}
}

// --- Status Effect Management ---

// applyStatusEffect adds a status effect to a combatant, refreshing if the same skill+stat already exists
func (e *Engine) applyStatusEffect(instance *combat.CombatInstance, target *combat.CombatantRef, se combat.StatusEffect) {
	// Remove existing effect from same skill+stat (refresh)
	effects := make([]combat.StatusEffect, 0, len(target.StatusEffects))
	for _, existing := range target.StatusEffects {
		if existing.SkillID != se.SkillID || existing.Stat != se.Stat {
			effects = append(effects, existing)
		}
	}
	effects = append(effects, se)
	target.StatusEffects = effects
}

// getSkillTargets returns the target(s) for a skill based on its targeting type
func (e *Engine) getSkillTargets(instance *combat.CombatInstance, caster *combat.CombatantRef, skill *skills.Skill, targetID string) []*combat.CombatantRef {
	switch skill.Target {
	case skills.TargetSelf:
		return []*combat.CombatantRef{caster}
	case skills.TargetAllEnemies:
		if caster.Type == combat.CombatantTypePlayer {
			return instance.GetLivingEnemies()
		}
		return instance.GetLivingPlayers()
	default: // TargetEnemy
		target := instance.GetCombatantByID(targetID)
		if target != nil && target.IsAlive {
			return []*combat.CombatantRef{target}
		}
		// Fallback: pick first living enemy
		if caster.Type == combat.CombatantTypePlayer {
			enemies := instance.GetLivingEnemies()
			if len(enemies) > 0 {
				return []*combat.CombatantRef{enemies[0]}
			}
		} else {
			players := instance.GetLivingPlayers()
			if len(players) > 0 {
				return []*combat.CombatantRef{players[0]}
			}
		}
		return nil
	}
}

// --- Status Effect Tick Processing ---

// ProcessStatusEffects processes all status effects on a combatant at the start of their turn.
// Returns true if the combatant is stunned and should skip their turn.
func (e *Engine) ProcessStatusEffects(instance *combat.CombatInstance, combatant *combat.CombatantRef) bool {
	stunned := false
	remaining := make([]combat.StatusEffect, 0, len(combatant.StatusEffects))

	for i := range combatant.StatusEffects {
		se := &combatant.StatusEffects[i]

		switch se.Type {
		case "stun":
			stunned = true
			instance.AddLogEntry(combat.CombatLogEntry{
				ActorID:   combatant.ID,
				ActorName: combatant.Name,
				Message:   fmt.Sprintf("%s is stunned and cannot act!", combatant.Name),
				Result:    "stunned",
			})

		case "dot":
			combatant.CurrentHP -= se.Value
			if combatant.CurrentHP <= 0 {
				combatant.CurrentHP = 0
				combatant.IsAlive = false
			}
			instance.AddLogEntry(combat.CombatLogEntry{
				ActorID:   combatant.ID,
				ActorName: combatant.Name,
				Damage:    se.Value,
				Message:   fmt.Sprintf("%s takes %d damage from %s! (%d/%d HP)", combatant.Name, se.Value, se.Name, combatant.CurrentHP, combatant.MaxHP),
				Result:    "dot",
			})

		case "hot":
			healing := se.Value
			combatant.CurrentHP += healing
			if combatant.CurrentHP > combatant.MaxHP {
				combatant.CurrentHP = combatant.MaxHP
			}
			instance.AddLogEntry(combat.CombatLogEntry{
				ActorID:   combatant.ID,
				ActorName: combatant.Name,
				Message:   fmt.Sprintf("%s is healed for %d by %s! (%d/%d HP)", combatant.Name, healing, se.Name, combatant.CurrentHP, combatant.MaxHP),
				Result:    "hot",
			})
		}

		// Decrement duration
		se.Duration--
		if se.Duration > 0 {
			remaining = append(remaining, *se)
		} else {
			instance.AddLogEntry(combat.CombatLogEntry{
				ActorID:   combatant.ID,
				ActorName: combatant.Name,
				Message:   fmt.Sprintf("%s's %s effect has worn off.", combatant.Name, se.Name),
				Result:    "effect_expired",
			})
		}
	}

	combatant.StatusEffects = remaining
	e.UpdateCombatant(instance, combatant)
	return stunned
}

// --- Round Start Processing ---

// ProcessManaRegen restores a small amount of mana to a combatant at the start of a new round.
// In-combat regen is halved compared to the base ManaRegen value.
func (e *Engine) ProcessManaRegen(instance *combat.CombatInstance, combatant *combat.CombatantRef) {
	if combatant.MaxMana <= 0 || combatant.ManaRegen <= 0 {
		return
	}
	if combatant.CurrentMana >= combatant.MaxMana {
		return
	}

	// In-combat mana regen is half the base rate (minimum 1)
	regen := combatant.ManaRegen / 2
	if regen < 1 {
		regen = 1
	}

	combatant.CurrentMana += regen
	if combatant.CurrentMana > combatant.MaxMana {
		combatant.CurrentMana = combatant.MaxMana
	}
	e.UpdateCombatant(instance, combatant)
}

// TickCooldowns decrements all skill cooldowns for a combatant at round start
func (e *Engine) TickCooldowns(instance *combat.CombatInstance, combatant *combat.CombatantRef) {
	if combatant.SkillCooldowns == nil {
		return
	}
	for skillID, cd := range combatant.SkillCooldowns {
		if cd > 0 {
			combatant.SkillCooldowns[skillID] = cd - 1
		}
		if combatant.SkillCooldowns[skillID] <= 0 {
			delete(combatant.SkillCooldowns, skillID)
		}
	}
	e.UpdateCombatant(instance, combatant)
}

// ProcessRoundStart handles round-start processing for all living combatants
func (e *Engine) ProcessRoundStart(instance *combat.CombatInstance) {
	for i := range instance.Players {
		p := &instance.Players[i]
		if p.IsAlive && !p.HasFled {
			e.TickCooldowns(instance, p)
			e.ProcessManaRegen(instance, p)
		}
	}
	for i := range instance.Enemies {
		en := &instance.Enemies[i]
		if en.IsAlive {
			e.TickCooldowns(instance, en)
			e.ProcessManaRegen(instance, en)
		}
	}
}

// IsStunned returns true if the combatant has a stun status effect (public method)
func (e *Engine) IsStunned(combatant *combat.CombatantRef) bool {
	return isStunned(combatant)
}
