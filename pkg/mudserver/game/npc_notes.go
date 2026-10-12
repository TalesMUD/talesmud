package game

import npc "github.com/talesmud/talesmud/pkg/entities/npcs"

// Note records the last operator event for an NPC instance.
func (m *NPCInstanceManager) Note(id, text string) {
	if m == nil || id == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.notes == nil {
		m.notes = map[string]string{}
	}
	m.notes[id] = text
}

// LastNote returns the last operator event for an NPC instance.
func (m *NPCInstanceManager) LastNote(id string) string {
	if m == nil || id == "" {
		return ""
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.notes[id]
}

// RestoreInstance puts a previously removed NPC back under the same id.
func (m *NPCInstanceManager) RestoreInstance(n *npc.NPC) {
	if m == nil || n == nil || n.Entity == nil || n.ID == "" {
		return
	}
	m.mu.Lock()
	m.instances[n.ID] = n
	m.mu.Unlock()
}

// AliveMatching returns a living instance whose id or template id matches.
func (m *NPCInstanceManager) AliveMatching(templateOrID string) *npc.NPC {
	if m == nil || templateOrID == "" {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, inst := range m.instances {
		if inst == nil || inst.IsDead {
			continue
		}
		if inst.ID == templateOrID || inst.TemplateID == templateOrID {
			return inst
		}
	}
	return nil
}

// DeadMatching returns a dead instance whose id or template id matches.
func (m *NPCInstanceManager) DeadMatching(templateOrID string) *npc.NPC {
	if m == nil || templateOrID == "" {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, inst := range m.instances {
		if inst == nil || !inst.IsDead {
			continue
		}
		if inst.ID == templateOrID || inst.TemplateID == templateOrID {
			return inst
		}
	}
	return nil
}
