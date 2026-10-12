package game

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/combat"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// AbortCombat ends a fight with no rewards, penalties, or healing.
// Current hit points stay where the fight left them. The caller may be the
// command loop or a test that has not started Run. The combat ticker takes
// the same lock, so this does not run beside a turn.
func (c *CombatController) AbortCombat(instance *combat.CombatInstance) error {
	if c == nil || c.manager == nil || instance == nil {
		return fmt.Errorf("no combat to end")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	current := c.manager.GetInstance(instance.ID)
	if current == nil {
		return fmt.Errorf("that fight already ended")
	}
	if c.engine != nil {
		c.engine.EndCombat(current, combat.CombatStateAborted)
	} else {
		current.State = combat.CombatStateAborted
	}
	if c.game != nil && c.game.NPCManager != nil {
		for _, enemy := range current.Enemies {
			hp := enemy.CurrentHP
			alive := enemy.IsAlive && hp > 0
			c.game.NPCManager.UpdateInstance(enemy.ID, func(n *npc.NPC) {
				n.InCombat = false
				n.CombatInstanceID = ""
				if alive {
					n.CurrentHitPoints = hp
					n.IsDead = false
					n.State = "idle"
				} else {
					n.CurrentHitPoints = 0
					n.IsDead = true
					n.State = "dead"
				}
			})
		}
	}
	if c.game != nil && c.game.Facade != nil {
		for _, player := range current.Players {
			char, err := c.game.Facade.CharactersService().FindByID(player.ID)
			if err != nil || char == nil {
				continue
			}
			char.InCombat = false
			char.CombatInstanceID = ""
			char.CurrentHitPoints = player.CurrentHP
			if err := c.game.Facade.CharactersService().Update(player.ID, char); err != nil {
				log.WithError(err).WithField("characterID", player.ID).Warn("abort combat: failed to clear player")
			}
			c.markCombatGrace(player.ID)
			if c.game.sendMessage == nil || char.BelongsUserID == "" {
				continue
			}
			text := "An operator ended the fight. No rewards or penalties."
			c.game.sendMessage <- messages.NewCombatEndMessage(char.BelongsUserID, text, string(combat.CombatStateAborted))
			if update := messages.NewCharacterUpdateMessage(char.BelongsUserID, char); update != nil {
				c.game.sendMessage <- update
			}
		}
	}
	c.manager.RemoveInstance(current.ID)
	log.WithField("instanceID", current.ID).Info("operator aborted combat")
	return nil
}

// OpEndCombat ends a stuck fight. It is not undoable.
func (g *Game) OpEndCombat(combatID, characterID, npcID string) (*OpResult, error) {
	if combatID == "" && characterID == "" && npcID == "" {
		return nil, opErr(400, "combatInstanceId, characterId, or npcInstanceId is required")
	}
	if g.CombatController == nil {
		return nil, opErr(404, "no fight is running")
	}
	var inst *combat.CombatInstance
	switch {
	case combatID != "":
		inst = g.CombatController.manager.GetInstance(combatID)
	case characterID != "":
		inst = g.CombatController.GetCombatInstance(characterID)
	default:
		inst = g.CombatController.GetCombatInstanceByNPC(npcID)
	}
	if inst == nil {
		return nil, opErr(404, "no fight is running")
	}
	id := inst.ID
	if err := g.CombatController.AbortCombat(inst); err != nil {
		return nil, opErr(409, err.Error())
	}
	return &OpResult{
		Summary:    opSummary("Ended fight %s. No rewards or penalties.", id),
		Undoable:   false,
		EntityType: "combat",
		EntityID:   id,
		Detail: map[string]interface{}{
			"combatInstanceId": id,
			"note":             "Ending a fight cannot be undone.",
		},
	}, nil
}
