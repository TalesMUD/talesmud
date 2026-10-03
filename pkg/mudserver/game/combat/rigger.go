package combat

import (
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	entcombat "github.com/talesmud/talesmud/pkg/entities/combat"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
)

func isPoisonStatus(se entcombat.StatusEffect) bool {
	blob := strings.ToLower(se.SkillID + " " + se.Name + " " + se.Stat)
	return strings.Contains(blob, "poison")
}

// applyScrapReturn deals the landed hit back to its attacker, then breaks the scrap.
// The original hit is already logged. This adds one return line and does not absorb.
func (e *Engine) applyScrapReturn(instance *entcombat.CombatInstance, attacker, target *entcombat.CombatantRef, damage int32, hit bool) {
	if e == nil || instance == nil || attacker == nil || target == nil || !hit || damage <= 0 || !target.ScrapArmed {
		return
	}
	target.ScrapArmed = false
	attacker.CurrentHP -= damage
	if attacker.CurrentHP <= 0 {
		attacker.CurrentHP = 0
		attacker.IsAlive = false
	}
	e.UpdateCombatant(instance, attacker)
	if target.ID != attacker.ID {
		e.UpdateCombatant(instance, target)
	}
	instance.AddLogEntry(entcombat.CombatLogEntry{
		ActorID:    attacker.ID,
		ActorName:  attacker.Name,
		Action:     entcombat.CombatActionBolt,
		TargetID:   target.ID,
		TargetName: target.Name,
		Result:     "scrap",
		Damage:     damage,
		Message:    fmt.Sprintf("Scrap returns %d to %s. The scrap breaks.", damage, attacker.Name),
	})
}

// ProcessBolt arms scrap on someone in this fight, including the rigger.
// It does not swing. A second bolt in the same fight does not arm another charge.
func (e *Engine) ProcessBolt(instance *entcombat.CombatInstance, actorID, targetID string) string {
	if e == nil || instance == nil {
		return ""
	}
	actor := instance.GetCombatantByID(actorID)
	if actor == nil || !balance.IsRigger(actor.ClassID) || actor.BoltLeft <= 0 {
		return ""
	}
	target := livingBoltTarget(instance, actor, targetID)
	if target == nil {
		return ""
	}
	actor.BoltLeft = 0
	target.ScrapArmed = true
	e.UpdateCombatant(instance, actor)
	if target.ID != actor.ID {
		e.UpdateCombatant(instance, target)
	}
	msg := fmt.Sprintf("You bolt scrap onto %s.", target.Name)
	instance.AddLogEntry(entcombat.CombatLogEntry{
		ActorID:    actor.ID,
		ActorName:  actor.Name,
		Action:     entcombat.CombatActionBolt,
		TargetID:   target.ID,
		TargetName: target.Name,
		Result:     "armed",
		Message:    msg,
	})
	return msg
}

func livingBoltTarget(instance *entcombat.CombatInstance, actor *entcombat.CombatantRef, targetID string) *entcombat.CombatantRef {
	targetID = strings.TrimSpace(targetID)
	if targetID == "" && actor != nil {
		targetID = actor.AutoAttackTargetID
	}
	if targetID != "" {
		if target := instance.GetCombatantByID(targetID); target != nil && target.IsAlive {
			return target
		}
		lower := strings.ToLower(targetID)
		for _, list := range [][]entcombat.CombatantRef{instance.Players, instance.Enemies} {
			for i := range list {
				c := &list[i]
				if c.IsAlive && strings.Contains(strings.ToLower(c.Name), lower) {
					return instance.GetCombatantByID(c.ID)
				}
			}
		}
	}
	if actor != nil {
		if enemies := instance.GetLivingEnemies(); len(enemies) > 0 {
			return enemies[0]
		}
		if actor.IsAlive {
			return actor
		}
	}
	return nil
}

// ProcessRig drops one turret in the fight's room and forfeits the swing.
// The turret hits once immediately and once more on the next round start, then despawns.
func (e *Engine) ProcessRig(instance *entcombat.CombatInstance, actorID string) string {
	if e == nil || instance == nil {
		return ""
	}
	actor := instance.GetCombatantByID(actorID)
	if actor == nil || !balance.IsRigger(actor.ClassID) || actor.RigLeft <= 0 || instance.Rig != nil {
		return ""
	}
	actor.RigLeft = 0
	e.UpdateCombatant(instance, actor)
	instance.Rig = &entcombat.RigTurret{
		ID:         uuid.New().String(),
		OwnerID:    actor.ID,
		Name:       "Rig",
		RoomID:     instance.OriginRoomID,
		RoundsLeft: 2,
		Follows:    false,
		Mult:       0.50,
	}
	msg := "You drop a rig. It stays in the room."
	instance.AddLogEntry(entcombat.CombatLogEntry{
		ActorID:   actor.ID,
		ActorName: actor.Name,
		Action:    entcombat.CombatActionRig,
		Result:    "drop",
		Message:   msg,
	})
	if hit := e.tickRig(instance); hit != "" {
		return msg + "\n" + hit
	}
	return msg
}

// tickRig is one turret round. It does not move the turret or add a follower.
func (e *Engine) tickRig(instance *entcombat.CombatInstance) string {
	if e == nil || instance == nil || instance.Rig == nil {
		return ""
	}
	rig := instance.Rig
	if rig.Follows {
		rig.Follows = false
	}
	if rig.RoomID == "" {
		rig.RoomID = instance.OriginRoomID
	}
	if rig.RoundsLeft <= 0 {
		instance.Rig = nil
		return ""
	}
	rig.RoundsLeft--
	last := rig.RoundsLeft <= 0
	owner := instance.GetCombatantByID(rig.OwnerID)
	target := rigStrikeTarget(instance, owner)
	var (
		msg                  string
		dmg                  int32
		targetID, targetName string
	)
	if owner == nil || target == nil {
		if last {
			msg = "The rig falls apart."
		} else {
			msg = "The rig waits. Nothing is in reach."
		}
	} else {
		swing := e.CalculateDamage(owner, target, false)
		mult := rig.Mult
		if mult <= 0 {
			mult = 0.50
		}
		dmg = int32(math.Round(float64(swing) * mult))
		if dmg < 1 {
			dmg = 1
		}
		target.CurrentHP -= dmg
		if target.CurrentHP <= 0 {
			target.CurrentHP = 0
			target.IsAlive = false
		}
		e.UpdateCombatant(instance, target)
		targetID = target.ID
		targetName = target.Name
		if last {
			msg = fmt.Sprintf("The rig hits %s for %d. The rig falls apart.", target.Name, dmg)
		} else {
			msg = fmt.Sprintf("The rig hits %s for %d.", target.Name, dmg)
		}
	}
	instance.AddLogEntry(entcombat.CombatLogEntry{
		ActorID:    rig.ID,
		ActorName:  rig.Name,
		Action:     entcombat.CombatActionRig,
		TargetID:   targetID,
		TargetName: targetName,
		Result:     "hit",
		Damage:     dmg,
		Message:    msg,
	})
	if last {
		instance.Rig = nil
	}
	return msg
}

func rigStrikeTarget(instance *entcombat.CombatInstance, owner *entcombat.CombatantRef) *entcombat.CombatantRef {
	if owner != nil && owner.AutoAttackTargetID != "" {
		if t := instance.GetCombatantByID(owner.AutoAttackTargetID); t != nil && t.IsAlive && t.ID != owner.ID {
			return t
		}
	}
	for _, enemy := range instance.GetLivingEnemies() {
		if owner == nil || enemy.ID != owner.ID {
			return enemy
		}
	}
	return nil
}
