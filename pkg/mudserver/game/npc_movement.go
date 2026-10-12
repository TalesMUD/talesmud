package game

import (
	"math/rand"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

func nextMoveAt(now time.Time, cfg *npc.Movement) time.Time {
	jitter := 0
	if cfg.JitterSeconds > 0 && cfg.JitterSeconds <= 86400 {
		jitter = rand.Intn(cfg.JitterSeconds + 1)
	}
	return now.Add(cfg.Interval() + time.Duration(jitter)*time.Second)
}

// initializeMovement runs before registration, never under the manager lock.
// Live instances are not saved: residents and spawns restart from authored rooms.
func (m *NPCInstanceManager) initializeMovement(n *npc.NPC) {
	cfg := n.EffectiveMovement()
	if cfg == nil {
		return
	}
	all, err := m.facade.RoomsService().FindAll()
	if err != nil {
		return
	}
	world := map[string]*rooms.Room{}
	for _, room := range all {
		if room != nil {
			world[room.ID] = room
		}
	}
	if start := npc.MovementStartRoom(n, world); start != "" {
		n.CurrentRoomID = start
	}
	if issues := npc.ValidateMovement(cfg, world); len(issues) > 0 {
		log.WithFields(log.Fields{"npc": n.ID, "issues": issues}).Warn("NPC movement invalid; movement skipped")
	}
	n.NextMoveAt = nextMoveAt(time.Now(), cfg)
	n.PausedUntil = time.Time{}
	n.RouteDirection = 1
}

// EngageNPC refreshes a short pause or opens an indefinite same-room session.
// The optional interface lets command test doubles remain compatible.
func (m *NPCInstanceManager) EngageNPC(characterID, npcID, kind string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	session := kind == "dialog" || kind == "shop"
	if session {
		// Replacing a moving NPC's overlay with a stationary NPC also releases it.
		for _, oldID := range m.engagements[characterID] {
			if old := m.instances[oldID]; old != nil && old.EffectiveMovement() != nil {
				old.PausedUntil = time.Now().Add(old.EffectiveMovement().Pause())
			}
		}
		delete(m.engagements, characterID)
	}
	n := m.instances[npcID]
	if n == nil || n.EffectiveMovement() == nil {
		return
	}
	n.PausedUntil = time.Now().Add(n.EffectiveMovement().Pause())
	if !session {
		return
	}
	if m.engagements == nil {
		m.engagements = map[string]map[string]string{}
	}
	m.engagements[characterID] = map[string]string{kind: npcID}
}
func (m *NPCInstanceManager) CloseNPCSession(characterID, kind string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for session, id := range m.engagements[characterID] {
		if kind != "" && kind != session {
			continue
		}
		if n := m.instances[id]; n != nil && n.EffectiveMovement() != nil {
			n.PausedUntil = time.Now().Add(n.EffectiveMovement().Pause())
		}
		delete(m.engagements[characterID], session)
	}
	if len(m.engagements[characterID]) == 0 {
		delete(m.engagements, characterID)
	}
}
func (m *NPCInstanceManager) NPCSessionOpen(characterID, npcID, kind string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.engagements[characterID][kind] == npcID
}
func (m *NPCInstanceManager) releaseNPCSessionsOutsideRoom(characterID, roomID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for kind, id := range m.engagements[characterID] {
		if n := m.instances[id]; n == nil || n.CurrentRoomID != roomID {
			if n != nil && n.EffectiveMovement() != nil {
				n.PausedUntil = time.Now().Add(n.EffectiveMovement().Pause())
			}
			delete(m.engagements[characterID], kind)
		}
	}
	if len(m.engagements[characterID]) == 0 {
		delete(m.engagements, characterID)
	}
}

// updateTimedNPC is called on the command loop, so engagement and movement
// cannot interleave between resolving an NPC and opening its dialog or shop.
func (g *Game) updateTimedNPC(n *npc.NPC, now time.Time) (npcMovementEvent, bool) {
	cfg := n.EffectiveMovement()
	if cfg == nil || n.IsDead || n.State == "dead" || n.InCombat || (n.MaxHitPoints > 0 && n.CurrentHitPoints <= 0) {
		return npcMovementEvent{}, false
	}
	if g.CombatController != nil && g.CombatController.IsNPCInCombat(n.ID) {
		return npcMovementEvent{}, false
	}
	world, err := g.loadRoomsByID()
	if err != nil {
		return npcMovementEvent{}, false
	}
	if len(npc.ValidateMovement(cfg, world)) > 0 {
		return npcMovementEvent{}, false
	}
	// Do not opt low-level enemies or bosses into wandering aggression. The
	// normal arrival hook remains responsible for noAggro rooms and cooldowns.
	if n.EnemyTrait != nil && n.EnemyTrait.AggroOnSight && (n.Level < 10 || strings.EqualFold(n.EnemyTrait.Difficulty, "boss") || npc.ContainsID(n.Tags, "boss")) {
		return npcMovementEvent{}, false
	}
	if !npc.MovementRoomAllowed(cfg, world[n.CurrentRoomID]) {
		return npcMovementEvent{}, false
	}

	g.NPCManager.mu.Lock()
	engaged := false
	for _, sessions := range g.NPCManager.engagements {
		for _, id := range sessions {
			if id == n.ID {
				engaged = true
			}
		}
	}
	if engaged {
		n.PausedUntil = now.Add(cfg.Pause())
	}
	if n.NextMoveAt.IsZero() {
		n.NextMoveAt = nextMoveAt(now, cfg)
	}
	due := !now.Before(n.NextMoveAt) && !now.Before(n.PausedUntil) && !engaged
	if due {
		n.NextMoveAt = nextMoveAt(now, cfg)
	}
	g.NPCManager.mu.Unlock()
	if !due {
		return npcMovementEvent{}, false
	}

	target := ""
	direction := n.RouteDirection
	if cfg.Mode == "route" {
		for i, id := range cfg.Route {
			if id != n.CurrentRoomID {
				continue
			}
			next := i + 1
			if cfg.LoopMode() == "pingpong" {
				if direction == 0 {
					direction = 1
				}
				if i == len(cfg.Route)-1 {
					direction = -1
				}
				if i == 0 {
					direction = 1
				}
				next = i + direction
			} else {
				next = (i + 1) % len(cfg.Route)
			}
			target = cfg.Route[next]
			break
		}
	} else {
		candidates := []string{}
		for _, id := range visibleExitTargets(world[n.CurrentRoomID]) {
			if id == n.CurrentRoomID || !npc.MovementRoomAllowed(cfg, world[id]) {
				continue
			}
			if n.Movement == nil && n.WanderRadius > 0 && !distanceWithinRadius(world, n.SpawnRoomID, id, n.WanderRadius) {
				continue
			}
			candidates = append(candidates, id)
		}
		if len(candidates) > 0 {
			target = candidates[rand.Intn(len(candidates))]
		}
	}
	if target == "" {
		return npcMovementEvent{}, false
	}
	event, moved := g.moveNPCInstance(n, target)
	if moved {
		g.NPCManager.UpdateInstance(n.ID, func(n *npc.NPC) { n.RouteDirection = direction })
	}
	return event, moved
}

func oppositeDirection(dir string) string {
	switch strings.ToLower(dir) {
	case "n", "north":
		return "south"
	case "s", "south":
		return "north"
	case "e", "east":
		return "west"
	case "w", "west":
		return "east"
	case "ne", "northeast":
		return "southwest"
	case "nw", "northwest":
		return "southeast"
	case "se", "southeast":
		return "northwest"
	case "sw", "southwest":
		return "northeast"
	case "up":
		return "below"
	case "down":
		return "above"
	case "in":
		return "outside"
	case "out":
		return "inside"
	}
	return "road"
}
func movementLine(template, name, dir, fromDir string) string {
	return strings.NewReplacer("{name}", name, "{dir}", dir, "{fromDir}", fromDir).Replace(template)
}

func (m *NPCInstanceManager) movementHeld(id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, sessions := range m.engagements {
		for _, npcID := range sessions {
			if npcID == id {
				return true
			}
		}
	}
	return false
}
