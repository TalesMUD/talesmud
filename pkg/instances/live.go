package instances

import (
	"sort"
	"time"

	"github.com/talesmud/talesmud/pkg/service"
)

// Copy is one live instance for the operator list.
type Copy struct {
	ID        string
	HubRoomID string
	SourceIDs []string
	CloneIDs  []string
	PlayerIDs []string
	Created   time.Time
}

// List returns the in-memory instances. Order is by instance id.
func (m *Manager) List() []Copy {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Copy, 0, len(m.instances))
	for _, inst := range m.instances {
		if inst == nil {
			continue
		}
		row := Copy{
			ID:        inst.ID,
			HubRoomID: inst.HubRoomID,
			Created:   inst.Created,
		}
		for src, clone := range inst.Clones {
			row.SourceIDs = append(row.SourceIDs, src)
			row.CloneIDs = append(row.CloneIDs, clone)
		}
		sort.Strings(row.SourceIDs)
		sort.Strings(row.CloneIDs)
		for id := range inst.Occupants {
			row.PlayerIDs = append(row.PlayerIDs, id)
		}
		sort.Strings(row.PlayerIDs)
		out = append(out, row)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// ForceDestroy deletes an instance's clone rooms after the caller has moved players out.
// id may be the instance id or any clone room id. Missing instances are a no-op.
func (m *Manager) ForceDestroy(roomsSvc service.RoomsService, id string) []string {
	if m == nil || id == "" {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	inst := m.instances[id]
	if inst == nil {
		if instID := m.byClone[id]; instID != "" {
			inst = m.instances[instID]
		}
	}
	if inst == nil {
		return nil
	}
	for cid := range inst.Occupants {
		delete(m.byCharacter, cid)
	}
	inst.Occupants = map[string]bool{}
	return m.destroyLocked(roomsSvc, inst)
}
