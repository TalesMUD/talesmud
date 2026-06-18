package game

import (
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

type npcMovementEvent struct {
	NPCID    string
	FromRoom string
	ToRoom   string
}

// handleNPCUpdates processes all NPC instances for state updates
func (g *Game) handleNPCUpdates() {
	instances := g.NPCManager.GetAllInstances()
	movements := make([]npcMovementEvent, 0)

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
			if movement, moved := g.updateIdleNPC(inst); moved {
				movements = append(movements, movement)
			}
		case "patrol":
			if movement, moved := g.updatePatrolNPC(inst); moved {
				movements = append(movements, movement)
			}
		case "combat":
			// Future: combat logic in separate PRD
		case "fleeing":
			// Future: flee logic
		}
	}

	g.broadcastNPCMovements(movements)
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
func (g *Game) updateIdleNPC(inst *npc.NPC) (npcMovementEvent, bool) {
	// Future: check aggro radius for nearby players.
	g.triggerIdleDialog(inst)

	if inst.WanderRadius > 0 {
		if nextRoomID := g.nextWanderRoom(inst); nextRoomID != "" && nextRoomID != inst.CurrentRoomID {
			return g.moveNPCInstance(inst, nextRoomID)
		}
	}

	return npcMovementEvent{}, false
}

// updatePatrolNPC handles patrol state behavior
func (g *Game) updatePatrolNPC(inst *npc.NPC) (npcMovementEvent, bool) {
	if len(inst.PatrolPath) == 0 {
		// No patrol path, switch to idle
		g.NPCManager.UpdateInstance(inst.Entity.ID, func(n *npc.NPC) {
			n.State = "idle"
		})
		return npcMovementEvent{}, false
	}

	if nextRoomID := nextPatrolRoom(inst); nextRoomID != "" && nextRoomID != inst.CurrentRoomID {
		return g.moveNPCInstance(inst, nextRoomID)
	}
	return npcMovementEvent{}, false
}

func (g *Game) moveNPCInstance(inst *npc.NPC, roomID string) (npcMovementEvent, bool) {
	if inst == nil || inst.Entity == nil || roomID == "" {
		return npcMovementEvent{}, false
	}
	fromRoomID := inst.CurrentRoomID
	if fromRoomID == roomID {
		return npcMovementEvent{}, false
	}
	if _, err := g.Facade.RoomsService().FindByID(roomID); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"npc":  inst.Entity.ID,
			"room": roomID,
		}).Warn("NPC movement target room not found")
		return npcMovementEvent{}, false
	}
	if !g.NPCManager.MoveInstance(inst.Entity.ID, roomID) {
		return npcMovementEvent{}, false
	}
	log.WithFields(log.Fields{
		"npc":  inst.GetDisplayName(),
		"from": fromRoomID,
		"to":   roomID,
	}).Debug("NPC moved")
	return npcMovementEvent{NPCID: inst.Entity.ID, FromRoom: fromRoomID, ToRoom: roomID}, true
}

func nextPatrolRoom(inst *npc.NPC) string {
	if inst == nil || len(inst.PatrolPath) == 0 {
		return ""
	}
	for i, roomID := range inst.PatrolPath {
		if roomID == inst.CurrentRoomID {
			return inst.PatrolPath[(i+1)%len(inst.PatrolPath)]
		}
	}
	return inst.PatrolPath[0]
}

func (g *Game) nextWanderRoom(inst *npc.NPC) string {
	if inst == nil || inst.CurrentRoomID == "" || inst.WanderRadius <= 0 {
		return ""
	}
	spawnRoomID := inst.SpawnRoomID
	if spawnRoomID == "" {
		spawnRoomID = inst.CurrentRoomID
	}
	roomsByID, err := g.loadRoomsByID()
	if err != nil {
		log.WithError(err).Warn("Could not load rooms for NPC wandering")
		return ""
	}
	return nextWanderRoom(inst, roomsByID, spawnRoomID)
}

func nextWanderRoom(inst *npc.NPC, roomsByID map[string]*rooms.Room, spawnRoomID string) string {
	if inst == nil || inst.WanderRadius <= 0 {
		return ""
	}
	current := roomsByID[inst.CurrentRoomID]
	if current == nil {
		return ""
	}

	candidates := visibleExitTargets(current)
	for _, candidate := range candidates {
		if candidate == inst.CurrentRoomID {
			continue
		}
		if distanceWithinRadius(roomsByID, spawnRoomID, candidate, inst.WanderRadius) {
			return candidate
		}
	}
	return ""
}

func visibleExitTargets(room *rooms.Room) []string {
	if room == nil || room.Exits == nil {
		return []string{}
	}
	targets := make([]string, 0, len(*room.Exits))
	for _, exit := range *room.Exits {
		if exit.Hidden || exit.Target == "" {
			continue
		}
		targets = append(targets, exit.Target)
	}
	sort.Strings(targets)
	return targets
}

func distanceWithinRadius(roomsByID map[string]*rooms.Room, fromRoomID, toRoomID string, radius int) bool {
	if radius < 0 {
		return false
	}
	if fromRoomID == toRoomID {
		return true
	}
	type queueItem struct {
		roomID string
		depth  int
	}
	visited := map[string]bool{fromRoomID: true}
	queue := []queueItem{{roomID: fromRoomID, depth: 0}}

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]
		if item.depth >= radius {
			continue
		}
		room := roomsByID[item.roomID]
		for _, target := range visibleExitTargets(room) {
			if visited[target] {
				continue
			}
			if target == toRoomID {
				return true
			}
			visited[target] = true
			queue = append(queue, queueItem{roomID: target, depth: item.depth + 1})
		}
	}
	return false
}

func (g *Game) loadRoomsByID() (map[string]*rooms.Room, error) {
	allRooms, err := g.Facade.RoomsService().FindAll()
	if err != nil {
		return nil, err
	}
	roomsByID := make(map[string]*rooms.Room, len(allRooms))
	for _, room := range allRooms {
		if room != nil {
			roomsByID[room.ID] = room
		}
	}
	return roomsByID, nil
}

func (g *Game) triggerIdleDialog(inst *npc.NPC) {
	if inst == nil || !inst.HasIdleDialog() || inst.IdleDialogTimeout <= 0 {
		return
	}
	key := "npc_idle_dialog:" + inst.Entity.ID
	if value, ok := GetState().Get(key); ok {
		if last, ok := value.(time.Time); ok && time.Since(last) < inst.IdleDialogTimeout {
			return
		}
	}
	dialog, err := g.Facade.DialogsService().FindByID(inst.IdleDialogID)
	if err != nil || dialog == nil {
		log.WithError(err).WithFields(log.Fields{
			"npc":    inst.Entity.ID,
			"dialog": inst.IdleDialogID,
		}).Warn("NPC idle dialog not found")
		return
	}
	GetState().Set(key, time.Now())
	g.SendMessage() <- messages.MessageResponse{
		Audience:   messages.MessageAudienceRoom,
		AudienceID: inst.CurrentRoomID,
		Type:       messages.MessageTypeDefault,
		Username:   inst.GetDisplayName(),
		Message:    dialog.GetText(),
	}
}

func (g *Game) broadcastNPCMovements(movements []npcMovementEvent) {
	if len(movements) == 0 {
		return
	}
	roomIDs := make(map[string]bool)
	for _, movement := range movements {
		if movement.FromRoom != "" {
			roomIDs[movement.FromRoom] = true
		}
		if movement.ToRoom != "" {
			roomIDs[movement.ToRoom] = true
		}
	}
	for roomID := range roomIDs {
		g.sendRoomUpdateToOccupants(roomID)
	}
}

func (g *Game) sendRoomUpdateToOccupants(roomID string) {
	room, err := g.Facade.RoomsService().FindByID(roomID)
	if err != nil || room == nil || room.Characters == nil {
		return
	}
	for _, characterID := range *room.Characters {
		character, err := g.Facade.CharactersService().FindByID(characterID)
		if err != nil || character == nil {
			continue
		}
		user, err := g.Facade.UsersService().FindByID(character.BelongsUserID)
		if err != nil || user == nil || !user.IsOnline || user.LastCharacter != character.ID {
			continue
		}
		update := messages.NewRoomUpdateMessage(room, user, g, character)
		update.AudienceID = user.ID
		g.SendMessage() <- update
	}
}

// updateNPC is a helper for individual NPC updates (deprecated, use handleNPCUpdates)
func (g *Game) updateNPC() {
	// Legacy function, kept for compatibility
}
