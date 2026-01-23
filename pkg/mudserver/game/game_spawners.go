package game

import (
	"time"

	log "github.com/sirupsen/logrus"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
)

// handleSpawnerUpdates processes all spawners and spawns new instances as needed
func (g *Game) handleSpawnerUpdates() {
	spawners, err := g.Facade.NPCSpawnersService().FindAll()
	if err != nil {
		log.WithError(err).Error("Failed to load spawners for update")
		return
	}

	for _, spawner := range spawners {
		g.updateSpawner(spawner)
	}
}

// updateSpawner checks if a spawner should spawn new instances
func (g *Game) updateSpawner(spawner *npc.NPCSpawner) {
	// Ensure spawner state exists
	state := g.NPCManager.EnsureSpawnerState(spawner.ID)

	// Clean up dead instances from active list
	g.NPCManager.CleanupDeadFromSpawner(spawner.ID)

	// Count current alive instances
	aliveCount := g.NPCManager.CountAliveForSpawner(spawner.ID)

	// Check if we're at max capacity
	if aliveCount >= spawner.MaxInstances {
		return
	}

	// Check spawn interval
	if time.Since(state.LastSpawnTime) < spawner.SpawnInterval {
		return
	}

	// Spawn new instance
	instance, err := g.NPCManager.SpawnInstance(spawner)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"spawner":  spawner.ID,
			"template": spawner.TemplateID,
		}).Warn("Failed to spawn NPC from spawner")
		return
	}

	log.WithFields(log.Fields{
		"spawner":  spawner.ID,
		"instance": instance.Entity.ID,
		"name":     instance.GetTargetName(),
		"room":     spawner.RoomID,
		"count":    aliveCount + 1,
		"max":      spawner.MaxInstances,
	}).Debug("Spawner created new NPC instance")
}
