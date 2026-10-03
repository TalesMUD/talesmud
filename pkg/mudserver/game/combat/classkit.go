package combat

import (
	"fmt"
	"math"

	entcombat "github.com/talesmud/talesmud/pkg/entities/combat"
	"github.com/talesmud/talesmud/pkg/entities/skills"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
)

func (e *Engine) tickKitRounds(instance *entcombat.CombatInstance, c *entcombat.CombatantRef) {
	if c == nil {
		return
	}
	changed := false
	if c.StandRounds > 0 {
		c.StandRounds--
		changed = true
	}
	if c.HobbleRounds > 0 {
		c.HobbleRounds--
		changed = true
	}
	if changed {
		e.UpdateCombatant(instance, c)
	}
}

// standRedirect sends a hit aimed at someone else onto a living Stand.
// The stander's own swings are not redirected onto them.
func (e *Engine) standRedirect(instance *entcombat.CombatInstance, attacker, target *entcombat.CombatantRef) *entcombat.CombatantRef {
	if instance == nil || target == nil {
		return nil
	}
	for _, list := range [][]entcombat.CombatantRef{instance.Players, instance.Enemies} {
		for i := range list {
			c := &list[i]
			if c.StandRounds <= 0 || !c.IsAlive || c.HasFled || c.ID == target.ID {
				continue
			}
			if attacker != nil && c.ID == attacker.ID {
				continue
			}
			return instance.GetCombatantByID(c.ID)
		}
	}
	return nil
}

func kitAlreadySpent(c *entcombat.CombatantRef, id string) bool {
	return c != nil && c.KitSpent != nil && c.KitSpent[id]
}

func markKitSpent(c *entcombat.CombatantRef, id string) {
	if c.KitSpent == nil {
		c.KitSpent = map[string]bool{}
	}
	c.KitSpent[id] = true
}

func (e *Engine) kitLog(instance *entcombat.CombatInstance, actor, target *entcombat.CombatantRef, msg string, dmg int32) {
	entry := entcombat.CombatLogEntry{
		ActorID:   actor.ID,
		ActorName: actor.Name,
		Action:    entcombat.CombatActionSkill,
		Result:    "cast",
		Damage:    dmg,
		Message:   msg,
	}
	if target != nil {
		entry.TargetID = target.ID
		entry.TargetName = target.Name
	}
	instance.AddLogEntry(entry)
}

func kitFail(name, msg string) SkillResult {
	return SkillResult{Success: false, SkillName: name, Messages: []string{msg}}
}

func (e *Engine) processClassKit(instance *entcombat.CombatInstance, caster *entcombat.CombatantRef, skill *skills.Skill, targetID string) SkillResult {
	if e == nil || instance == nil || caster == nil || skill == nil {
		return kitFail("", "The skill does nothing.")
	}
	if !skill.AllowsClass(caster.ClassID) {
		return kitFail(skill.Name, fmt.Sprintf("%s is not your skill.", skill.Name))
	}
	id := skill.Entity.ID
	if skill.OncePerFight && kitAlreadySpent(caster, id) {
		return kitFail(skill.Name, fmt.Sprintf("%s is spent for this fight.", skill.Name))
	}
	if !skill.OncePerFight && skill.CooldownRounds > 0 && caster.SkillCooldowns != nil {
		if cd := caster.SkillCooldowns[id]; cd > 0 {
			return kitFail(skill.Name, fmt.Sprintf("%s is on cooldown (%d rounds remaining)", skill.Name, cd))
		}
	}

	pay := func() {
		if skill.OncePerFight {
			markKitSpent(caster, id)
		} else if skill.CooldownRounds > 0 {
			if caster.SkillCooldowns == nil {
				caster.SkillCooldowns = map[string]int{}
			}
			caster.SkillCooldowns[id] = skill.CooldownRounds
		}
		e.UpdateCombatant(instance, caster)
	}

	switch skill.Kit {
	case skills.KitBrace:
		pay()
		caster = instance.GetCombatantByID(caster.ID)
		caster.BraceLeft = 1
		e.UpdateCombatant(instance, caster)
		msg := "You brace. The next hit is halved."
		e.kitLog(instance, caster, caster, msg, 0)
		return SkillResult{Success: true, SkillName: skill.Name, KeepsSwing: true, Messages: []string{msg}}

	case skills.KitPin:
		target := livingKitTarget(instance, caster, targetID)
		if target == nil {
			return kitFail(skill.Name, "Nobody to pin.")
		}
		caster.PinLeft = 1
		e.UpdateCombatant(instance, caster)
		if !e.ArmPin(instance, caster.ID, target.ID) {
			caster = instance.GetCombatantByID(caster.ID)
			caster.PinLeft = 0
			e.UpdateCombatant(instance, caster)
			return kitFail(skill.Name, "The pin does not catch.")
		}
		pay()
		caster = instance.GetCombatantByID(caster.ID)
		msg := fmt.Sprintf("You pin %s. The next time they leave, they stay.", target.Name)
		e.kitLog(instance, caster, target, msg, 0)
		return SkillResult{Success: true, SkillName: skill.Name, KeepsSwing: true, Messages: []string{msg}}

	case skills.KitSlip:
		pay()
		caster = instance.GetCombatantByID(caster.ID)
		caster.HasFled = true
		e.UpdateCombatant(instance, caster)
		msg := "You slip out."
		e.kitLog(instance, caster, nil, msg, 0)
		return SkillResult{Success: true, SkillName: skill.Name, SlipMove: true, Messages: []string{msg}}

	case skills.KitSmoke:
		target := livingKitTarget(instance, caster, targetID)
		if target == nil {
			return kitFail(skill.Name, "Nobody to smoke.")
		}
		pay()
		target = instance.GetCombatantByID(target.ID)
		target.SmokeMiss = true
		e.UpdateCombatant(instance, target)
		caster = instance.GetCombatantByID(caster.ID)
		msg := fmt.Sprintf("You throw smoke. %s misses their next swing.", target.Name)
		e.kitLog(instance, caster, target, msg, 0)
		return SkillResult{Success: true, SkillName: skill.Name, Messages: []string{msg}}

	case skills.KitStand:
		pay()
		caster = instance.GetCombatantByID(caster.ID)
		caster.StandRounds = 2
		e.UpdateCombatant(instance, caster)
		msg := "You stand. For two rounds, hits aimed at others hit you."
		e.kitLog(instance, caster, caster, msg, 0)
		return SkillResult{Success: true, SkillName: skill.Name, Messages: []string{msg}}

	case skills.KitGlyph:
		pay()
		caster = instance.GetCombatantByID(caster.ID)
		caster.GlyphCut = 4
		e.UpdateCombatant(instance, caster)
		msg := "You glyph. The next hit on you is cut by 4."
		e.kitLog(instance, caster, caster, msg, 0)
		return SkillResult{Success: true, SkillName: skill.Name, Messages: []string{msg}}

	case skills.KitHobble:
		target := livingKitTarget(instance, caster, targetID)
		if target == nil {
			return kitFail(skill.Name, "Nobody to hobble.")
		}
		pay()
		target = instance.GetCombatantByID(target.ID)
		target.HobbleRounds = 2
		e.UpdateCombatant(instance, target)
		caster = instance.GetCombatantByID(caster.ID)
		msg := fmt.Sprintf("You hobble %s. Their hits are lighter for two rounds.", target.Name)
		e.kitLog(instance, caster, target, msg, 0)
		return SkillResult{Success: true, SkillName: skill.Name, Messages: []string{msg}}

	case skills.KitInscribe:
		target := livingKitTarget(instance, caster, targetID)
		if target == nil {
			return kitFail(skill.Name, "Nobody to inscribe.")
		}
		msg := e.applyInscribe(instance, caster, target)
		if msg == "" {
			return kitFail(skill.Name, "The rune does not catch.")
		}
		pay()
		caster = instance.GetCombatantByID(caster.ID)
		e.kitLog(instance, caster, target, msg, 0)
		return SkillResult{Success: true, SkillName: skill.Name, Messages: []string{msg}}

	case skills.KitSlam:
		target := livingKitTarget(instance, caster, targetID)
		if target == nil {
			return kitFail(skill.Name, "Nobody to slam.")
		}
		pay()
		swing := e.processAttackSwingMult(instance, caster.ID, target.ID, skill.SwingMult, false)
		target = instance.GetCombatantByID(target.ID)
		name := targetID
		if target != nil {
			name = target.Name
		}
		msg := fmt.Sprintf("You slam %s for %d.", name, swing.Damage)
		if swing.TargetDied {
			msg += fmt.Sprintf(" %s has been defeated!", name)
		}
		caster = instance.GetCombatantByID(caster.ID)
		e.kitLog(instance, caster, target, msg, swing.Damage)
		res := SkillResult{Success: true, SkillName: skill.Name, TotalDamage: swing.Damage, HitsLanded: 0, Messages: []string{msg}}
		if swing.Hit {
			res.HitsLanded = 1
		}
		if swing.TargetDied && target != nil {
			res.TargetsDied = []string{target.ID}
		}
		return res

	case skills.KitNick:
		target := livingKitTarget(instance, caster, targetID)
		if target == nil {
			return kitFail(skill.Name, "Nobody to nick.")
		}
		pay()
		swings := balance.ClassSwings(caster.ClassID)
		if swings < 1 {
			swings = 1
		}
		var total int32
		var died bool
		var hit bool
		for i := 0; i < swings; i++ {
			swing := e.processAttackSwingMult(instance, caster.ID, target.ID, 1, false)
			total += swing.Damage
			hit = hit || swing.Hit
			if swing.TargetDied {
				died = true
				break
			}
		}
		if !died {
			extra := e.processAttackSwingMult(instance, caster.ID, target.ID, skill.SwingMult, false)
			total += extra.Damage
			hit = hit || extra.Hit
			died = extra.TargetDied
		}
		target = instance.GetCombatantByID(target.ID)
		name := "them"
		if target != nil {
			name = target.Name
		}
		msg := fmt.Sprintf("You nick %s for %d.", name, total)
		if died {
			msg += fmt.Sprintf(" %s has been defeated!", name)
		}
		caster = instance.GetCombatantByID(caster.ID)
		e.kitLog(instance, caster, target, msg, total)
		res := SkillResult{Success: true, SkillName: skill.Name, TotalDamage: total, Messages: []string{msg}}
		if hit {
			res.HitsLanded = 1
		}
		if died && target != nil {
			res.TargetsDied = []string{target.ID}
		}
		return res

	case skills.KitSear:
		target := livingKitTarget(instance, caster, targetID)
		if target == nil {
			return kitFail(skill.Name, "Nobody to sear.")
		}
		pay()
		caster = instance.GetCombatantByID(caster.ID)
		target = instance.GetCombatantByID(target.ID)
		dmg := e.CalculateDamage(caster, target, false)
		mult := skill.SwingMult
		if mult <= 0 {
			mult = 1.80
		}
		dmg = int32(math.Round(float64(dmg) * mult))
		if dmg < 1 {
			dmg = 1
		}
		target.CurrentHP -= dmg
		died := false
		if target.CurrentHP <= 0 {
			target.CurrentHP = 0
			target.IsAlive = false
			died = true
		}
		e.UpdateCombatant(instance, target)
		msg := fmt.Sprintf("You sear %s for %d.", target.Name, dmg)
		if died {
			msg += fmt.Sprintf(" %s has been defeated!", target.Name)
		}
		e.kitLog(instance, caster, target, msg, dmg)
		res := SkillResult{Success: true, SkillName: skill.Name, TotalDamage: dmg, HitsLanded: 1, Messages: []string{msg}}
		if died {
			res.TargetsDied = []string{target.ID}
		}
		return res

	case skills.KitReel:
		target := fledKitTarget(instance, caster, targetID)
		if target == nil {
			if explicitStillHere(instance, targetID) {
				return kitFail(skill.Name, "They have not left.")
			}
			return kitFail(skill.Name, "Nobody has left.")
		}
		pay()
		target.HasFled = false
		e.UpdateCombatant(instance, target)
		caster = instance.GetCombatantByID(caster.ID)
		msg := fmt.Sprintf("You reel %s back into the fight.", target.Name)
		e.kitLog(instance, caster, target, msg, 0)
		return SkillResult{Success: true, SkillName: skill.Name, ReelID: target.ID, Messages: []string{msg}}

	case skills.KitBolt:
		if caster.BoltLeft <= 0 {
			return kitFail(skill.Name, "Bolt is spent for this fight.")
		}
		msg := e.ProcessBolt(instance, caster.ID, targetID)
		if msg == "" {
			return kitFail(skill.Name, "The scrap does nothing.")
		}
		pay()
		return SkillResult{Success: true, SkillName: skill.Name, Messages: []string{msg}}

	case skills.KitRig:
		if caster.RigLeft <= 0 || instance.Rig != nil {
			return kitFail(skill.Name, "Rig is spent for this fight.")
		}
		msg := e.ProcessRig(instance, caster.ID)
		if msg == "" {
			return kitFail(skill.Name, "The rig does nothing.")
		}
		pay()
		return SkillResult{Success: true, SkillName: skill.Name, Messages: []string{msg}}

	case skills.KitOverload:
		if instance.Rig == nil || instance.Rig.RoundsLeft <= 0 || instance.Rig.OwnerID != caster.ID {
			return kitFail(skill.Name, "There is no turret to overload.")
		}
		pay()
		instance.Rig.Mult = 0.80
		caster = instance.GetCombatantByID(caster.ID)
		msg := "You overload the rig. Its remaining hits hit harder."
		e.kitLog(instance, caster, nil, msg, 0)
		return SkillResult{Success: true, SkillName: skill.Name, Messages: []string{msg}}
	default:
		return kitFail(skill.Name, "The skill does nothing.")
	}
}

func livingKitTarget(instance *entcombat.CombatInstance, caster *entcombat.CombatantRef, targetID string) *entcombat.CombatantRef {
	if instance == nil || caster == nil {
		return nil
	}
	if targetID != "" {
		if t := instance.GetCombatantByID(targetID); t != nil && t.IsAlive && !t.HasFled && t.ID != caster.ID {
			return t
		}
	}
	if caster.AutoAttackTargetID != "" {
		if t := instance.GetCombatantByID(caster.AutoAttackTargetID); t != nil && t.IsAlive && !t.HasFled && t.ID != caster.ID {
			return t
		}
	}
	for _, enemy := range instance.GetLivingEnemies() {
		if enemy.IsAlive && !enemy.HasFled && enemy.ID != caster.ID {
			return enemy
		}
	}
	for _, p := range instance.GetLivingPlayers() {
		if p.IsAlive && !p.HasFled && p.ID != caster.ID {
			return p
		}
	}
	return nil
}

func explicitStillHere(instance *entcombat.CombatInstance, targetID string) bool {
	if targetID == "" || instance == nil {
		return false
	}
	t := instance.GetCombatantByID(targetID)
	return t != nil && t.IsAlive && !t.HasFled
}

func fledKitTarget(instance *entcombat.CombatInstance, caster *entcombat.CombatantRef, targetID string) *entcombat.CombatantRef {
	if instance == nil || caster == nil {
		return nil
	}
	if targetID != "" {
		t := instance.GetCombatantByID(targetID)
		if t != nil && t.IsAlive && t.HasFled && t.ID != caster.ID {
			return t
		}
		return nil
	}
	for _, list := range [][]entcombat.CombatantRef{instance.Enemies, instance.Players} {
		for i := range list {
			c := &list[i]
			if c.IsAlive && c.HasFled && c.ID != caster.ID {
				return instance.GetCombatantByID(c.ID)
			}
		}
	}
	return nil
}
