package game

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/service"
)

// SpawnerState tracks runtime state for an NPC spawner
type SpawnerState struct {
	// ActiveInstances contains IDs of currently alive instances managed by this spawner
	ActiveInstances []string
	// LastSpawnTime is when the last instance was spawned
	LastSpawnTime time.Time
}

// NPCInstanceManager manages in-memory NPC instances
// Instances are spawned from templates and exist only in memory (not persisted)
type NPCInstanceManager struct {
	mu sync.RWMutex

	// instances maps instance ID to NPC
	instances map[string]*npc.NPC

	// spawnerState tracks runtime state per spawner ID
	spawnerState map[string]*SpawnerState

	facade service.Facade
}

// NewNPCInstanceManager creates a new NPC instance manager
func NewNPCInstanceManager(facade service.Facade) *NPCInstanceManager {
	return &NPCInstanceManager{
		instances:    make(map[string]*npc.NPC),
		spawnerState: make(map[string]*SpawnerState),
		facade:       facade,
	}
}

// Initialize loads all spawners and creates their initial instances
func (m *NPCInstanceManager) Initialize() error {
	// First, load spawners
	spawners, err := m.facade.NPCSpawnersService().FindAll()
	if err != nil {
		return err
	}

	log.WithField("count", len(spawners)).Info("Initializing NPC spawners")

	for _, spawner := range spawners {
		m.spawnerState[spawner.ID] = &SpawnerState{
			ActiveInstances: make([]string, 0),
			LastSpawnTime:   time.Now(),
		}

		// Spawn initial count
		for i := 0; i < spawner.InitialCount && i < spawner.MaxInstances; i++ {
			if _, err := m.SpawnInstance(spawner); err != nil {
				log.WithError(err).WithField("spawner", spawner.ID).Warn("Failed to spawn initial NPC")
			}
		}
	}

	// Then, load residents (unique NPCs assigned directly to rooms)
	if err := m.initializeResidents(); err != nil {
		log.WithError(err).Warn("Failed to initialize room residents")
	}

	log.WithField("instances", len(m.instances)).Info("NPC instance manager initialized")
	return nil
}

// initializeResidents loads unique NPCs that are directly assigned to rooms
func (m *NPCInstanceManager) initializeResidents() error {
	rooms, err := m.facade.RoomsService().FindAll()
	if err != nil {
		return err
	}

	residentCount := 0
	for _, room := range rooms {
		npcIDs := room.GetNPCIDs()
		for _, npcID := range npcIDs {
			// Load the NPC from the database
			npcData, err := m.facade.NPCsService().FindByID(npcID)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"npcID":  npcID,
					"roomID": room.ID,
				}).Warn("Failed to load resident NPC")
				continue
			}

			// Skip if it's a template (shouldn't happen, but safety check)
			if npcData.IsTemplate {
				log.WithFields(log.Fields{
					"npcID":  npcID,
					"roomID": room.ID,
				}).Warn("Skipping template NPC assigned as resident")
				continue
			}

			// Register the NPC
			m.RegisterExistingNPC(npcData, room.ID)
			residentCount++
		}
	}

	if residentCount > 0 {
		log.WithField("count", residentCount).Info("Loaded resident NPCs from rooms")
	}

	return nil
}

// SpawnInstance creates a new NPC instance from a spawner's template
func (m *NPCInstanceManager) SpawnInstance(spawner *npc.NPCSpawner) (*npc.NPC, error) {
	instance, err := m.facade.NPCsService().SpawnFromTemplate(spawner.TemplateID, spawner.RoomID)
	if err != nil {
		return nil, err
	}

	// Override respawn time if specified in spawner
	if spawner.RespawnTimeOverride != nil {
		instance.RespawnTime = *spawner.RespawnTimeOverride
	}

	m.mu.Lock()
	m.instances[instance.Entity.ID] = instance

	state := m.spawnerState[spawner.ID]
	if state != nil {
		state.ActiveInstances = append(state.ActiveInstances, instance.Entity.ID)
		state.LastSpawnTime = time.Now()
	}
	m.mu.Unlock()

	log.WithFields(log.Fields{
		"instance": instance.Entity.ID,
		"name":     instance.GetTargetName(),
		"template": spawner.TemplateID,
		"room":     spawner.RoomID,
	}).Info("Spawned NPC instance")

	return instance, nil
}

// SpawnInstanceDirect creates a new NPC instance from a template ID directly (not via spawner)
func (m *NPCInstanceManager) SpawnInstanceDirect(templateID, roomID string) (*npc.NPC, error) {
	instance, err := m.facade.NPCsService().SpawnFromTemplate(templateID, roomID)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.instances[instance.Entity.ID] = instance
	m.mu.Unlock()

	log.WithFields(log.Fields{
		"instance": instance.Entity.ID,
		"name":     instance.GetTargetName(),
		"template": templateID,
		"room":     roomID,
	}).Info("Spawned NPC instance directly")

	return instance, nil
}

// RegisterExistingNPC registers a unique NPC (resident) that already exists in the database
// This is used for non-template NPCs that are directly assigned to rooms
func (m *NPCInstanceManager) RegisterExistingNPC(n *npc.NPC, roomID string) {
	if n == nil || n.Entity == nil {
		return
	}

	// Set the current room ID
	n.CurrentRoomID = roomID
	n.SpawnRoomID = roomID

	// Ensure NPC is in a valid state
	if n.CurrentHitPoints <= 0 && !n.IsDead {
		n.CurrentHitPoints = n.MaxHitPoints
	}
	if n.State == "" {
		n.State = "idle"
	}

	m.mu.Lock()
	m.instances[n.Entity.ID] = n
	m.mu.Unlock()

	log.WithFields(log.Fields{
		"id":   n.Entity.ID,
		"name": n.GetDisplayName(),
		"room": roomID,
	}).Debug("Registered resident NPC")
}

// GetInstance returns an instance by ID
func (m *NPCInstanceManager) GetInstance(id string) *npc.NPC {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.instances[id]
}

// GetInstancesInRoom returns all alive instances in a room
func (m *NPCInstanceManager) GetInstancesInRoom(roomID string) []*npc.NPC {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*npc.NPC
	for _, inst := range m.instances {
		if inst.CurrentRoomID == roomID && !inst.IsDead {
			result = append(result, inst)
		}
	}
	return result
}

// GetAllInstances returns all instances (including dead ones awaiting respawn)
func (m *NPCInstanceManager) GetAllInstances() []*npc.NPC {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*npc.NPC, 0, len(m.instances))
	for _, inst := range m.instances {
		result = append(result, inst)
	}
	return result
}

// GetAliveInstances returns only alive instances
func (m *NPCInstanceManager) GetAliveInstances() []*npc.NPC {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*npc.NPC
	for _, inst := range m.instances {
		if !inst.IsDead {
			result = append(result, inst)
		}
	}
	return result
}

// KillInstance marks an instance as dead
func (m *NPCInstanceManager) KillInstance(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	inst, ok := m.instances[id]
	if !ok {
		return false
	}

	inst.IsDead = true
	inst.DeathTime = time.Now()
	inst.State = "dead"
	inst.CurrentHitPoints = 0

	log.WithFields(log.Fields{
		"instance": id,
		"name":     inst.GetDisplayName(),
	}).Info("NPC instance killed")

	return true
}

// RespawnInstance resets an instance to alive state
func (m *NPCInstanceManager) RespawnInstance(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	inst, ok := m.instances[id]
	if !ok {
		return false
	}

	inst.IsDead = false
	inst.CurrentHitPoints = inst.MaxHitPoints
	inst.State = "idle"
	inst.CurrentRoomID = inst.SpawnRoomID
	inst.Updated = time.Now()

	log.WithFields(log.Fields{
		"instance": id,
		"name":     inst.GetDisplayName(),
		"room":     inst.SpawnRoomID,
	}).Info("NPC instance respawned")

	return true
}

// RemoveInstance completely removes an instance from the manager
func (m *NPCInstanceManager) RemoveInstance(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.instances, id)

	// Remove from spawner state
	for _, state := range m.spawnerState {
		for i, instID := range state.ActiveInstances {
			if instID == id {
				state.ActiveInstances = append(state.ActiveInstances[:i], state.ActiveInstances[i+1:]...)
				break
			}
		}
	}
}

// UpdateInstance updates an instance's state using a callback function
func (m *NPCInstanceManager) UpdateInstance(id string, updater func(*npc.NPC)) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	inst, ok := m.instances[id]
	if !ok {
		return false
	}

	updater(inst)
	inst.Updated = time.Now()
	return true
}

// DamageInstance applies damage to an instance and returns true if it died
func (m *NPCInstanceManager) DamageInstance(id string, amount int32) (died bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	inst, ok := m.instances[id]
	if !ok || inst.IsDead {
		return false
	}

	inst.CurrentHitPoints -= amount
	inst.Updated = time.Now()

	if inst.CurrentHitPoints <= 0 {
		inst.CurrentHitPoints = 0
		inst.IsDead = true
		inst.DeathTime = time.Now()
		inst.State = "dead"
		return true
	}
	return false
}

// HealInstance restores health to an instance
func (m *NPCInstanceManager) HealInstance(id string, amount int32) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	inst, ok := m.instances[id]
	if !ok || inst.IsDead {
		return false
	}

	inst.CurrentHitPoints += amount
	if inst.CurrentHitPoints > inst.MaxHitPoints {
		inst.CurrentHitPoints = inst.MaxHitPoints
	}
	inst.Updated = time.Now()
	return true
}

// MoveInstance moves an instance to a new room
func (m *NPCInstanceManager) MoveInstance(id, roomID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	inst, ok := m.instances[id]
	if !ok || inst.IsDead {
		return false
	}

	inst.CurrentRoomID = roomID
	inst.Updated = time.Now()
	return true
}

// GetSpawnerState returns the runtime state for a spawner
func (m *NPCInstanceManager) GetSpawnerState(spawnerID string) *SpawnerState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.spawnerState[spawnerID]
}

// EnsureSpawnerState creates spawner state if it doesn't exist
func (m *NPCInstanceManager) EnsureSpawnerState(spawnerID string) *SpawnerState {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.spawnerState[spawnerID]
	if !ok {
		state = &SpawnerState{
			ActiveInstances: make([]string, 0),
			LastSpawnTime:   time.Now(),
		}
		m.spawnerState[spawnerID] = state
	}
	return state
}

// CleanupDeadFromSpawner removes dead instances from a spawner's active list
func (m *NPCInstanceManager) CleanupDeadFromSpawner(spawnerID string) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	state := m.spawnerState[spawnerID]
	if state == nil {
		return 0
	}

	alive := make([]string, 0)
	removed := 0
	for _, id := range state.ActiveInstances {
		inst := m.instances[id]
		if inst != nil && !inst.IsDead {
			alive = append(alive, id)
		} else {
			removed++
		}
	}
	state.ActiveInstances = alive
	return removed
}

// CountAliveForSpawner returns the count of alive instances for a spawner
func (m *NPCInstanceManager) CountAliveForSpawner(spawnerID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state := m.spawnerState[spawnerID]
	if state == nil {
		return 0
	}

	count := 0
	for _, id := range state.ActiveInstances {
		inst := m.instances[id]
		if inst != nil && !inst.IsDead {
			count++
		}
	}
	return count
}

// FindInstanceByNameInRoom finds an instance by name match in a room
// Supports numbered format like "Rat #2" for targeting specific instances
func (m *NPCInstanceManager) FindInstanceByNameInRoom(roomID, name string) *npc.NPC {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get all alive instances in the room
	var roomInstances []*npc.NPC
	for _, inst := range m.instances {
		if inst.CurrentRoomID == roomID && !inst.IsDead {
			roomInstances = append(roomInstances, inst)
		}
	}

	if len(roomInstances) == 0 {
		return nil
	}

	// Check if name contains a number suffix like "Rat#2" or "rat#1"
	nameLower := toLower(name)
	if idx := lastIndexOf(nameLower, "#"); idx != -1 {
		baseName := name[:idx]
		numStr := name[idx+1:]
		if num := parsePositiveInt(numStr); num > 0 {
			// Find the nth instance with this base name
			count := 0
			for _, inst := range roomInstances {
				if containsIgnoreCase(inst.Name, baseName) {
					count++
					if count == num {
						return inst
					}
				}
			}
			return nil // Number out of range
		}
	}

	// Fall back to partial name match (returns first match)
	for _, inst := range roomInstances {
		if containsIgnoreCase(inst.GetDisplayName(), name) ||
			containsIgnoreCase(inst.GetTargetName(), name) {
			return inst
		}
	}
	return nil
}

// lastIndexOf returns the last index of substr in s, or -1 if not found
func lastIndexOf(s, substr string) int {
	for i := len(s) - len(substr); i >= 0; i-- {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// parsePositiveInt parses a positive integer, returns 0 on failure
func parsePositiveInt(s string) int {
	if len(s) == 0 {
		return 0
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && containsLower(toLower(s), toLower(substr))
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
