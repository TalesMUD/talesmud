package game

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

// handleNPCUpdates processes all NPC instances for state updates
func (g *Game) handleNPCUpdates() {
	instances := g.NPCManager.GetAllInstances()

	for _, inst := range instances {
		// Skip templates (shouldn't be in manager, but safety check)
		if inst.IsTemplate {
			continue
		}

		// Handle dead NPCs
		if inst.IsDead {
			if inst.IsInstance() {
				// Spawner-created instances: delete the dead instance.
				// The spawner will create a fresh replacement on its next cycle.
				g.NPCManager.RemoveInstance(inst.Entity.ID)
			} else if inst.ShouldRespawn() {
				// Unique/resident NPCs: recycle in place after respawn timer
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
		g.wanderNPC(inst)
	}

	// Check if NPC has idle dialog and should trigger
	if inst.HasIdleDialog() && inst.IdleDialogTimeout > 0 {
		g.triggerIdleDialog(inst)
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

	nextRoomID := nextPatrolRoom(inst.CurrentRoomID, inst.PatrolPath)
	if nextRoomID == "" {
		return
	}
	if _, err := g.Facade.RoomsService().FindByID(nextRoomID); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"npc":  inst.Entity.ID,
			"room": nextRoomID,
		}).Warn("NPC patrol target room not found")
		return
	}
	g.moveNPC(inst, nextRoomID, "patrols")
}

func nextPatrolRoom(currentRoomID string, patrolPath []string) string {
	if len(patrolPath) == 0 {
		return ""
	}
	for i, roomID := range patrolPath {
		if roomID == currentRoomID {
			return patrolPath[(i+1)%len(patrolPath)]
		}
	}
	return patrolPath[0]
}

func (g *Game) wanderNPC(inst *npc.NPC) {
	current, err := g.Facade.RoomsService().FindByID(inst.CurrentRoomID)
	if err != nil || current == nil || current.Exits == nil {
		return
	}

	spawnRoomID := inst.SpawnRoomID
	if spawnRoomID == "" {
		spawnRoomID = inst.CurrentRoomID
	}

	for _, exit := range *current.Exits {
		if exit.Target == "" || exit.Hidden {
			continue
		}
		distance, ok := g.roomDistance(spawnRoomID, exit.Target, inst.WanderRadius)
		if !ok || distance > inst.WanderRadius {
			continue
		}
		g.moveNPC(inst, exit.Target, "wanders")
		return
	}
}

func (g *Game) roomDistance(startRoomID, targetRoomID string, maxDepth int) (int, bool) {
	if startRoomID == targetRoomID {
		return 0, true
	}
	if maxDepth <= 0 {
		return 0, false
	}

	type queuedRoom struct {
		id       string
		distance int
	}
	visited := map[string]bool{startRoomID: true}
	queue := []queuedRoom{{id: startRoomID}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current.distance >= maxDepth {
			continue
		}

		room, err := g.Facade.RoomsService().FindByID(current.id)
		if err != nil || room == nil || room.Exits == nil {
			continue
		}
		for _, exit := range *room.Exits {
			if exit.Target == "" || exit.Hidden || visited[exit.Target] {
				continue
			}
			nextDistance := current.distance + 1
			if exit.Target == targetRoomID {
				return nextDistance, true
			}
			visited[exit.Target] = true
			queue = append(queue, queuedRoom{id: exit.Target, distance: nextDistance})
		}
	}

	return 0, false
}

func (g *Game) moveNPC(inst *npc.NPC, targetRoomID, verb string) {
	if inst.CurrentRoomID == targetRoomID {
		return
	}
	oldRoomID := inst.CurrentRoomID
	if !g.NPCManager.MoveInstance(inst.Entity.ID, targetRoomID) {
		return
	}
	if oldRoomID != "" {
		g.SendMessage() <- messages.MessageResponse{
			Audience:   messages.MessageAudienceRoom,
			AudienceID: oldRoomID,
			Type:       messages.MessageTypeDefault,
			Username:   inst.GetDisplayName(),
			Message:    inst.GetDisplayName() + " leaves.",
		}
	}
	g.SendMessage() <- messages.MessageResponse{
		Audience:   messages.MessageAudienceRoom,
		AudienceID: targetRoomID,
		Type:       messages.MessageTypeDefault,
		Username:   inst.GetDisplayName(),
		Message:    inst.GetDisplayName() + " " + verb + " in.",
	}
}

func (g *Game) triggerIdleDialog(inst *npc.NPC) {
	now := time.Now()
	if !inst.LastIdleDialog.IsZero() && now.Sub(inst.LastIdleDialog) < inst.IdleDialogTimeout {
		return
	}

	dialog, err := g.Facade.DialogsService().FindByID(inst.IdleDialogID)
	if err != nil || dialog == nil {
		log.WithError(err).WithFields(log.Fields{
			"npc":      inst.Entity.ID,
			"dialogID": inst.IdleDialogID,
		}).Warn("NPC idle dialog not found")
		return
	}

	text := dialog.Render(&dialogs.DialogState{
		CurrentDialogID: "main",
		Context: map[string]string{
			"NPC": inst.GetDisplayName(),
		},
	})
	if text == "" {
		return
	}

	g.SendMessage() <- messages.MessageResponse{
		Audience:   messages.MessageAudienceRoom,
		AudienceID: inst.CurrentRoomID,
		Type:       messages.MessageTypeDefault,
		Username:   inst.GetDisplayName(),
		Message:    text,
	}
	g.NPCManager.UpdateInstance(inst.Entity.ID, func(n *npc.NPC) {
		n.LastIdleDialog = now
	})
}

// updateNPC is a helper for individual NPC updates (deprecated, use handleNPCUpdates)
func (g *Game) updateNPC() {
	// Legacy function, kept for compatibility
}
