package game

import (
	log "github.com/sirupsen/logrus"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
)

// handleNPCUpdates processes all NPC instances for state updates
func (g *Game) handleNPCUpdates() {
	instances := g.NPCManager.GetAllInstances()

	for _, inst := range instances {
		// Skip templates (shouldn't be in manager, but safety check)
		if inst.IsTemplate {
			continue
		}

		// Handle dead NPCs - check respawn
		if inst.IsDead {
			if inst.ShouldRespawn() {
				g.respawnNPC(inst)
			}
			continue
		}

		// Update NPC based on state
		switch inst.State {
		case "idle":
			g.updateIdleNPC(inst)
		case "patrol":
			g.updatePatrolNPC(inst)
		case "combat":
			// Future: combat logic in separate PRD
		case "fleeing":
			// Future: flee logic
		}
	}
}

// respawnNPC resets an NPC to alive state at its spawn room
func (g *Game) respawnNPC(inst *npc.NPC) {
	success := g.NPCManager.RespawnInstance(inst.Entity.ID)
	if success {
		log.WithFields(log.Fields{
			"npc":  inst.GetDisplayName(),
			"room": inst.SpawnRoomID,
		}).Debug("NPC respawned")
	}
}

// updateIdleNPC handles idle state behavior
func (g *Game) updateIdleNPC(inst *npc.NPC) {
	// Future: check aggro radius for nearby players
	// For now, just handle wandering if configured
	if inst.WanderRadius > 0 {
		// TODO: implement wandering behavior
		// This would pick a random adjacent room within wander radius
	}

	// Check if NPC has idle dialog and should trigger
	if inst.HasIdleDialog() && inst.IdleDialogTimeout > 0 {
		// TODO: implement idle dialog triggering
	}
}

// updatePatrolNPC handles patrol state behavior
func (g *Game) updatePatrolNPC(inst *npc.NPC) {
	if len(inst.PatrolPath) == 0 {
		// No patrol path, switch to idle
		g.NPCManager.UpdateInstance(inst.Entity.ID, func(n *npc.NPC) {
			n.State = "idle"
		})
		return
	}

	// TODO: implement patrol path following
	// This would track current patrol index and move to next room
}

// updateNPC is a helper for individual NPC updates (deprecated, use handleNPCUpdates)
func (g *Game) updateNPC() {
	// Legacy function, kept for compatibility
}
