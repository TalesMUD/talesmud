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

	now := time.Now()
	if spawnerRespawnDelayed(spawner.RespawnDelay, state.LastDeathTime, now) {
		return
	}
	if !spawnerShouldSpawn(aliveCount, spawner.InitialCount, spawner.MaxInstances, spawner.SpawnInterval, state.LastSpawnTime, now) {
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
	if instance != nil && instance.Entity != nil {
		g.NoteNPCAppeared(instance.Entity.ID, instance.CurrentRoomID)
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

// spawnerShouldSpawn keeps InitialCount filled immediately (so a second
// guest still finds the tutorial rat). SpawnInterval only throttles extra
// instances between InitialCount and MaxInstances.
func spawnerShouldSpawn(alive, initial, max int, interval time.Duration, lastSpawn, now time.Time) bool {
	if max <= 0 {
		max = 1
	}
	if alive >= max {
		return false
	}
	if initial > 0 && alive < initial {
		return true
	}
	if interval <= 0 {
		return true
	}
	return now.Sub(lastSpawn) >= interval
}

// spawnerRespawnDelayed is true while a spawner with a RespawnDelay is still
// inside that window after its last instance death. It never blocks a spawner
// that has not lost an instance yet (server start fills InitialCount).
func spawnerRespawnDelayed(delay time.Duration, lastDeath, now time.Time) bool {
	if delay <= 0 || lastDeath.IsZero() {
		return false
	}
	return now.Sub(lastDeath) < delay
}
