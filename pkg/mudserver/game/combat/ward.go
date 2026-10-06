package combat

import (
	"fmt"
	"strings"

	entcombat "github.com/talesmud/talesmud/pkg/entities/combat"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
)

func (e *Engine) tickWardGuard(c *entcombat.CombatantRef) bool {
	if c == nil || c.GuardRounds <= 0 {
		return false
	}
	c.GuardRounds--
	if c.GuardRounds <= 0 {
		c.GuardCharges = 0
		c.GuardSelf = false
		c.GuardTargetID = ""
	}
	return true
}

// guardRedirect sends the next hit aimed at a guarded ally onto that Ward.
// A self guard does not move the hit. The Ward's own swing is not redirected.
func (e *Engine) guardRedirect(instance *entcombat.CombatInstance, attacker, target *entcombat.CombatantRef) *entcombat.CombatantRef {
	if instance == nil || target == nil {
		return nil
	}
	for _, list := range [][]entcombat.CombatantRef{instance.Players, instance.Enemies} {
		for i := range list {
			c := &list[i]
			if c.GuardCharges <= 0 || c.GuardRounds <= 0 || c.GuardSelf || c.GuardTargetID == "" {
				continue
			}
			if !c.IsAlive || c.HasFled || c.GuardTargetID != target.ID || c.ID == target.ID {
				continue
			}
			if attacker != nil && c.ID == attacker.ID {
				continue
			}
			if !balance.IsWard(c.ClassID) {
				continue
			}
			return instance.GetCombatantByID(c.ID)
		}
	}
	return nil
}

// applyWardSoak grants Grit for damage taken and throws a fraction of a hit back.
// Retaliate uses Grit from before this hit. Self-guard grants two Grit and does not reduce damage.
// redirected means this hit was moved off a guarded ally, which spends the guard.
func (e *Engine) applyWardSoak(instance *entcombat.CombatInstance, attacker, target *entcombat.CombatantRef, damage int32, redirected, retaliate bool) string {
	if e == nil || instance == nil || target == nil || damage <= 0 || !balance.IsWard(target.ClassID) {
		return ""
	}
	note := ""
	if retaliate {
		back := balance.WardRetaliateDamage(damage, target.Grit)
		if back > 0 && attacker != nil && attacker.ID != target.ID && attacker.IsAlive {
			attacker.CurrentHP -= back
			if attacker.CurrentHP <= 0 {
				attacker.CurrentHP = 0
				attacker.IsAlive = false
			}
			e.UpdateCombatant(instance, attacker)
			note = fmt.Sprintf("Grit throws %d back at %s.", back, attacker.Name)
			if !attacker.IsAlive {
				note += fmt.Sprintf(" %s has been defeated!", attacker.Name)
			}
			instance.AddLogEntry(entcombat.CombatLogEntry{
				ActorID:    target.ID,
				ActorName:  target.Name,
				Action:     entcombat.CombatActionAttack,
				TargetID:   attacker.ID,
				TargetName: attacker.Name,
				Result:     "grit",
				Damage:     back,
				Message:    note,
			})
		}
	}
	selfSoak := target.GuardSelf && target.GuardCharges > 0 && target.GuardRounds > 0
	if selfSoak || (redirected && target.GuardCharges > 0) {
		target.GuardCharges = 0
		target.GuardRounds = 0
		target.GuardSelf = false
		target.GuardTargetID = ""
	}
	target.Grit = balance.WardGritAfter(target.Grit, selfSoak)
	e.UpdateCombatant(instance, target)
	return note
}

func livingWardAllies(instance *entcombat.CombatInstance, caster *entcombat.CombatantRef) []*entcombat.CombatantRef {
	if instance == nil || caster == nil {
		return nil
	}
	list := instance.Players
	if caster.Type != entcombat.CombatantTypePlayer {
		list = instance.Enemies
	}
	out := make([]*entcombat.CombatantRef, 0)
	for i := range list {
		c := &list[i]
		if c.IsAlive && !c.HasFled && c.ID != caster.ID {
			if live := instance.GetCombatantByID(c.ID); live != nil {
				out = append(out, live)
			}
		}
	}
	return out
}

// resolveWardGuard picks an ally, or self when that ally is missing or self is chosen.
// An explicit name that matches nobody fails instead of guessing.
func resolveWardGuard(instance *entcombat.CombatInstance, caster *entcombat.CombatantRef, targetID string) (ally *entcombat.CombatantRef, self bool, ok bool) {
	if instance == nil || caster == nil {
		return nil, false, false
	}
	token := strings.TrimSpace(targetID)
	if token == "" {
		if allies := livingWardAllies(instance, caster); len(allies) > 0 {
			return allies[0], false, true
		}
		return nil, true, true
	}
	if strings.EqualFold(token, "self") || strings.EqualFold(token, "me") || token == caster.ID {
		return nil, true, true
	}
	if t := instance.GetCombatantByID(token); t != nil && t.IsAlive && !t.HasFled {
		if t.ID == caster.ID {
			return nil, true, true
		}
		for _, ally := range livingWardAllies(instance, caster) {
			if ally.ID == t.ID {
				return ally, false, true
			}
		}
		return nil, false, false
	}
	lower := strings.ToLower(token)
	for _, ally := range livingWardAllies(instance, caster) {
		if strings.Contains(strings.ToLower(ally.Name), lower) {
			return ally, false, true
		}
	}
	return nil, false, false
}
