package combat

import (
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/combat"
)

// Manager manages in-memory combat instances
type Manager struct {
	mu sync.RWMutex

	// instances maps combat instance ID to CombatInstance
	instances map[string]*combat.CombatInstance

	// playerToCombat maps character ID to combat instance ID
	playerToCombat map[string]string

	// npcToCombat maps NPC ID to combat instance ID
	npcToCombat map[string]string
}

// NewManager creates a new combat manager
func NewManager() *Manager {
	return &Manager{
		instances:      make(map[string]*combat.CombatInstance),
		playerToCombat: make(map[string]string),
		npcToCombat:    make(map[string]string),
	}
}

// CreateInstance creates a new combat instance for the given room
func (m *Manager) CreateInstance(originRoomID string) *combat.CombatInstance {
	instance := combat.NewCombatInstance(originRoomID)

	m.mu.Lock()
	m.instances[instance.ID] = instance
	m.mu.Unlock()

	log.WithFields(log.Fields{
		"instanceID": instance.ID,
		"roomID":     originRoomID,
	}).Info("Created combat instance")

	return instance
}

// GetInstance returns a combat instance by ID
func (m *Manager) GetInstance(id string) *combat.CombatInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.instances[id]
}

// GetInstanceByPlayerID returns the combat instance a player is in
func (m *Manager) GetInstanceByPlayerID(characterID string) *combat.CombatInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instanceID, ok := m.playerToCombat[characterID]
	if !ok {
		return nil
	}
	return m.instances[instanceID]
}

// GetInstanceByNPCID returns the combat instance an NPC is in
func (m *Manager) GetInstanceByNPCID(npcID string) *combat.CombatInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instanceID, ok := m.npcToCombat[npcID]
	if !ok {
		return nil
	}
	return m.instances[instanceID]
}

// RegisterPlayer adds a player to a combat instance's lookup map
func (m *Manager) RegisterPlayer(characterID string, instanceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.playerToCombat[characterID] = instanceID
}

// RegisterNPC adds an NPC to a combat instance's lookup map
func (m *Manager) RegisterNPC(npcID string, instanceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.npcToCombat[npcID] = instanceID
}

// UnregisterPlayer removes a player from the combat lookup map
func (m *Manager) UnregisterPlayer(characterID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.playerToCombat, characterID)
}

// UnregisterNPC removes an NPC from the combat lookup map
func (m *Manager) UnregisterNPC(npcID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.npcToCombat, npcID)
}

// RemoveInstance removes a combat instance completely
func (m *Manager) RemoveInstance(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, ok := m.instances[id]
	if !ok {
		return
	}

	// Clean up player mappings
	for _, player := range instance.Players {
		delete(m.playerToCombat, player.ID)
	}

	// Clean up NPC mappings
	for _, enemy := range instance.Enemies {
		delete(m.npcToCombat, enemy.ID)
	}

	delete(m.instances, id)

	log.WithField("instanceID", id).Info("Removed combat instance")
}

// GetAllInstances returns all active combat instances
func (m *Manager) GetAllInstances() []*combat.CombatInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*combat.CombatInstance, 0, len(m.instances))
	for _, inst := range m.instances {
		result = append(result, inst)
	}
	return result
}

// GetActiveInstances returns all instances in active state
func (m *Manager) GetActiveInstances() []*combat.CombatInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*combat.CombatInstance
	for _, inst := range m.instances {
		if inst.State == combat.CombatStateActive {
			result = append(result, inst)
		}
	}
	return result
}

// IsPlayerInCombat checks if a player is currently in a combat instance
func (m *Manager) IsPlayerInCombat(characterID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.playerToCombat[characterID]
	return ok
}

// IsNPCInCombat checks if an NPC is currently in a combat instance
func (m *Manager) IsNPCInCombat(npcID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.npcToCombat[npcID]
	return ok
}

// GetInstancesInRoom returns all combat instances originating from a room
func (m *Manager) GetInstancesInRoom(roomID string) []*combat.CombatInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*combat.CombatInstance
	for _, inst := range m.instances {
		if inst.OriginRoomID == roomID {
			result = append(result, inst)
		}
	}
	return result
}

// UpdateInstance updates a combat instance using a callback function
func (m *Manager) UpdateInstance(id string, updater func(*combat.CombatInstance)) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	inst, ok := m.instances[id]
	if !ok {
		return false
	}

	updater(inst)
	return true
}

// GetInstanceCount returns the total number of active combat instances
func (m *Manager) GetInstanceCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.instances)
}
