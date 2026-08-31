package instances

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/service"
)

// Instance is one private copy of a cellar graph for a party of occupants.
type Instance struct {
	ID         string
	HubRoomID  string
	Occupants  map[string]bool
	Clones     map[string]string // template room ID -> clone room ID
	CloneOrder []string
}

// Manager tracks live cellar instances. Destroyed when empty.
type Manager struct {
	mu           sync.Mutex
	instances    map[string]*Instance
	byClone      map[string]string // clone room -> instance ID
	byCharacter  map[string]string // character ID -> instance ID
	byHubDestKey map[string]string // unused reserved
}

// NewManager creates an empty instance manager.
func NewManager() *Manager {
	return &Manager{
		instances:    map[string]*Instance{},
		byClone:      map[string]string{},
		byCharacter:  map[string]string{},
		byHubDestKey: map[string]string{},
	}
}

// Enter clones dest+reachable rooms (not hub) for this character and returns the clone dest ID.
func (m *Manager) Enter(roomsSvc service.RoomsService, characterID, hubID, destID string) (string, error) {
	if characterID == "" || destID == "" {
		return "", fmt.Errorf("missing character or dest")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if instID := m.byCharacter[characterID]; instID != "" {
		if inst := m.instances[instID]; inst != nil {
			if clone, ok := inst.Clones[destID]; ok {
				inst.Occupants[characterID] = true
				return clone, nil
			}
		}
	}

	allList, err := roomsSvc.FindAll()
	if err != nil {
		return "", err
	}
	all := map[string]*rooms.Room{}
	for _, r := range allList {
		if r != nil && r.Entity != nil {
			all[r.ID] = r
		}
	}
	graph := CollectGraph(all, hubID, destID)
	if len(graph) == 0 {
		return "", fmt.Errorf("no instance graph from %s", destID)
	}

	instID := newID()
	inst := &Instance{
		ID:        instID,
		HubRoomID: hubID,
		Occupants: map[string]bool{characterID: true},
		Clones:    map[string]string{},
	}
	for _, tid := range graph {
		inst.Clones[tid] = CloneID(tid, instID)
		inst.CloneOrder = append(inst.CloneOrder, tid)
	}

	for _, tid := range graph {
		src := all[tid]
		if src == nil {
			return "", fmt.Errorf("missing template room %s", tid)
		}
		clone := cloneRoom(src, inst.Clones[tid], hubID, inst.Clones)
		if _, err := roomsSvc.Import(clone); err != nil {
			return "", fmt.Errorf("import clone %s: %w", clone.ID, err)
		}
		m.byClone[clone.ID] = instID
	}

	m.instances[instID] = inst
	m.byCharacter[characterID] = instID
	return inst.Clones[destID], nil
}

// NoteLeave records that a character left an instance room. Destroys the
// instance when no occupants remain. Returns deleted clone room IDs.
func (m *Manager) NoteLeave(roomsSvc service.RoomsService, characterID, fromRoomID, toRoomID string) []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	instID := m.byClone[fromRoomID]
	if instID == "" {
		return nil
	}
	inst := m.instances[instID]
	if inst == nil {
		return nil
	}
	if m.byClone[toRoomID] == instID {
		return nil
	}
	delete(inst.Occupants, characterID)
	delete(m.byCharacter, characterID)
	if len(inst.Occupants) > 0 {
		return nil
	}
	return m.destroyLocked(roomsSvc, inst)
}

// DestroyCharacterInstance removes a character from whatever instance they occupy.
func (m *Manager) DestroyCharacterInstance(roomsSvc service.RoomsService, characterID string) []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	instID := m.byCharacter[characterID]
	if instID == "" {
		return nil
	}
	inst := m.instances[instID]
	if inst == nil {
		return nil
	}
	delete(inst.Occupants, characterID)
	delete(m.byCharacter, characterID)
	if len(inst.Occupants) == 0 {
		return m.destroyLocked(roomsSvc, inst)
	}
	return nil
}

// IsClone reports whether roomID is a live instance copy.
func (m *Manager) IsClone(roomID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.byClone[roomID]
	return ok
}

// ClonesForCharacter returns template->clone ids for the character's instance.
func (m *Manager) ClonesForCharacter(characterID string) map[string]string {
	m.mu.Lock()
	defer m.mu.Unlock()
	instID := m.byCharacter[characterID]
	if instID == "" {
		return nil
	}
	inst := m.instances[instID]
	if inst == nil {
		return nil
	}
	out := make(map[string]string, len(inst.Clones))
	for k, v := range inst.Clones {
		out[k] = v
	}
	return out
}

func (m *Manager) destroyLocked(roomsSvc service.RoomsService, inst *Instance) []string {
	deleted := make([]string, 0, len(inst.CloneOrder))
	for _, tid := range inst.CloneOrder {
		cid := inst.Clones[tid]
		_ = roomsSvc.Delete(cid)
		delete(m.byClone, cid)
		deleted = append(deleted, cid)
	}
	delete(m.instances, inst.ID)
	return deleted
}

func cloneRoom(src *rooms.Room, cloneID, hubID string, clones map[string]string) *rooms.Room {
	out := &rooms.Room{
		Entity:          &entities.Entity{ID: cloneID},
		LookAt:          src.LookAt,
		Name:            src.Name,
		Description:     src.Description,
		RoomType:        src.RoomType,
		Area:            src.Area,
		AreaType:        src.AreaType,
		Tags:            append([]string(nil), src.Tags...),
		OnEnterScriptID: src.OnEnterScriptID,
		CanBind:         src.CanBind,
		Meta:            src.Meta,
		Coords:          src.Coords,
	}
	chars := rooms.Characters{}
	out.Characters = &chars
	if src.Items != nil {
		items := append(rooms.Items(nil), *src.Items...)
		out.Items = &items
	}
	if src.Actions != nil {
		actions := append(rooms.Actions(nil), *src.Actions...)
		out.Actions = &actions
	}
	if src.Exits != nil {
		exits := make(rooms.Exits, 0, len(*src.Exits))
		for _, ex := range *src.Exits {
			mapped := ex
			if mapped.Target != hubID {
				if cid, ok := clones[mapped.Target]; ok {
					mapped.Target = cid
				}
			}
			mapped.Instance = false
			exits = append(exits, mapped)
		}
		out.Exits = &exits
	}
	return out
}

func newID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
