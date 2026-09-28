package instances

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/service"
)

const (
	// DefaultProcTimeout is how long a generated instance lives if the caller
	// does not set one. Authored graph instances are not on this clock.
	DefaultProcTimeout = 30 * time.Minute
	maxProcRooms       = 20
)

// Encounter is one weighted NPC template eligible for a player level band.
// MaxLevel 0 means no upper bound.
type Encounter struct {
	TemplateID string
	MinLevel   int32
	MaxLevel   int32
	Weight     int
}

// ProcSpec describes one generated instance. The manager does not read a
// calendar and does not move party followers.
type ProcSpec struct {
	TemplateIDs  []string
	Count        int
	ReturnRoomID string
	Encounters   []Encounter
	Timeout      time.Duration
	Seed         int64
}

// Spawn is an NPC template to place in a generated room.
type Spawn struct {
	RoomID     string
	TemplateID string
}

// ProcResult is the private line of rooms for one character.
type ProcResult struct {
	EntryRoomID string
	RoomIDs     []string
	Spawns      []Spawn
}

// Generate builds a private line of rooms for one character.
// A second character gets a different copy. The same character cannot hold
// two instances. An empty encounter band spawns nothing.
func (m *Manager) Generate(roomsSvc service.RoomsService, characterID string, playerLevel int32, spec ProcSpec) (ProcResult, error) {
	if m == nil {
		return ProcResult{}, fmt.Errorf("no instance manager")
	}
	if characterID == "" {
		return ProcResult{}, fmt.Errorf("missing character")
	}
	if spec.Count < 1 || spec.Count > maxProcRooms {
		return ProcResult{}, fmt.Errorf("room count must be 1..%d", maxProcRooms)
	}
	if len(spec.TemplateIDs) == 0 {
		return ProcResult{}, fmt.Errorf("no room templates")
	}
	if spec.ReturnRoomID == "" {
		return ProcResult{}, fmt.Errorf("missing return room")
	}
	if _, err := roomsSvc.FindByID(spec.ReturnRoomID); err != nil {
		return ProcResult{}, fmt.Errorf("return room: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.byCharacter[characterID] != "" {
		return ProcResult{}, fmt.Errorf("character already has an instance")
	}

	seed := spec.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	rng := rand.New(rand.NewSource(seed))

	timeout := spec.Timeout
	if timeout <= 0 {
		timeout = DefaultProcTimeout
	}
	instID := newID()
	inst := &Instance{
		ID:         instID,
		HubRoomID:  spec.ReturnRoomID,
		Occupants:  map[string]bool{characterID: true},
		Clones:     map[string]string{},
		Procedural: true,
		ExpiresAt:  time.Now().Add(timeout),
	}

	cloneIDs := make([]string, spec.Count)
	templates := make([]*rooms.Room, spec.Count)
	for i := 0; i < spec.Count; i++ {
		tid := spec.TemplateIDs[rng.Intn(len(spec.TemplateIDs))]
		src, err := roomsSvc.FindByID(tid)
		if err != nil || src == nil {
			return ProcResult{}, fmt.Errorf("room template %s: %w", tid, err)
		}
		templates[i] = src
		slot := fmt.Sprintf("slot-%d", i)
		cloneIDs[i] = CloneID(slot, instID)
		inst.Clones[slot] = cloneIDs[i]
		inst.CloneOrder = append(inst.CloneOrder, slot)
	}

	var created []string
	var spawns []Spawn
	for i := 0; i < spec.Count; i++ {
		exits := procExits(i, spec.Count, cloneIDs, spec.ReturnRoomID)
		clone := buildProcRoom(templates[i], cloneIDs[i], exits)
		if _, err := roomsSvc.Import(clone); err != nil {
			m.rollbackLocked(roomsSvc, created)
			return ProcResult{}, fmt.Errorf("import clone: %w", err)
		}
		created = append(created, clone.ID)
		m.byClone[clone.ID] = instID
		if npcID, ok := pickEncounter(rng, spec.Encounters, playerLevel); ok {
			spawns = append(spawns, Spawn{RoomID: clone.ID, TemplateID: npcID})
		}
	}

	m.instances[instID] = inst
	m.byCharacter[characterID] = instID
	return ProcResult{EntryRoomID: cloneIDs[0], RoomIDs: cloneIDs, Spawns: spawns}, nil
}

// Expire destroys procedural instances whose timeout has passed.
// Authored graph instances are left alone. Returns deleted clone room ids.
func (m *Manager) Expire(roomsSvc service.RoomsService, now time.Time) []string {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var deleted []string
	for id, inst := range m.instances {
		if inst == nil || !inst.Procedural || inst.ExpiresAt.IsZero() || now.Before(inst.ExpiresAt) {
			continue
		}
		for cid := range inst.Occupants {
			delete(m.byCharacter, cid)
		}
		deleted = append(deleted, m.destroyLocked(roomsSvc, inst)...)
		delete(m.instances, id)
	}
	return deleted
}

func (m *Manager) rollbackLocked(roomsSvc service.RoomsService, ids []string) {
	for _, id := range ids {
		_ = roomsSvc.Delete(id)
		delete(m.byClone, id)
	}
}

func procExits(i, n int, cloneIDs []string, returnRoom string) rooms.Exits {
	var exits rooms.Exits
	if i > 0 {
		exits = append(exits, blockedExit("south", cloneIDs[i-1]))
	}
	if i+1 < n {
		exits = append(exits, blockedExit("north", cloneIDs[i+1]))
	}
	exits = append(exits, blockedExit("out", returnRoom))
	return exits
}

func blockedExit(name, target string) rooms.Exit {
	return rooms.Exit{
		Name:     name,
		Type:     rooms.RoomExitTypeDirection,
		Target:   target,
		Instance: true,
	}
}

func buildProcRoom(src *rooms.Room, cloneID string, exits rooms.Exits) *rooms.Room {
	out := &rooms.Room{
		Entity:          &entities.Entity{ID: cloneID},
		LookAt:          src.LookAt,
		Name:            src.Name,
		Description:     src.Description,
		Area:            src.Area,
		AreaType:        src.AreaType,
		Tags:            append([]string{"procedural"}, src.Tags...),
		OnEnterScriptID: src.OnEnterScriptID,
	}
	chars := rooms.Characters{}
	out.Characters = &chars
	out.Exits = &exits
	return out
}

func pickEncounter(rng *rand.Rand, list []Encounter, level int32) (string, bool) {
	total := 0
	var eligible []Encounter
	for _, enc := range list {
		if enc.TemplateID == "" || enc.Weight <= 0 {
			continue
		}
		if level < enc.MinLevel {
			continue
		}
		if enc.MaxLevel > 0 && level > enc.MaxLevel {
			continue
		}
		eligible = append(eligible, enc)
		total += enc.Weight
	}
	if total == 0 {
		return "", false
	}
	roll := rng.Intn(total)
	for _, enc := range eligible {
		if roll < enc.Weight {
			return enc.TemplateID, true
		}
		roll -= enc.Weight
	}
	return eligible[len(eligible)-1].TemplateID, true
}
