package game

import (
	"encoding/json"

	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
)

// OpNPCHeal sets a living NPC to full hit points. The inverse restores the previous total.
func (g *Game) OpNPCHeal(instanceID string) (*OpResult, error) {
	if instanceID == "" {
		return nil, opErr(400, "npcInstanceId is required")
	}
	if g.NPCManager == nil {
		return nil, opErr(503, "npc manager is not running")
	}
	inst := g.NPCManager.GetInstance(instanceID)
	if inst == nil {
		return nil, opErr(404, "npc instance not found")
	}
	if inst.IsDead || inst.CurrentHitPoints <= 0 {
		return nil, opErr(409, "npc is dead")
	}
	before := inst.CurrentHitPoints
	maxHP := inst.MaxHitPoints
	if maxHP < before {
		maxHP = before
	}
	g.NPCManager.UpdateInstance(instanceID, func(n *npc.NPC) {
		n.CurrentHitPoints = maxHP
		n.IsDead = false
	})
	g.syncNPCCombatHP(instanceID, maxHP)
	fresh := g.NPCManager.GetInstance(instanceID)
	g.persistNPCIfStored(fresh)
	g.NPCManager.Note(instanceID, "healed to full")
	name := npcName(fresh)
	return &OpResult{
		Summary:    opSummary("Healed %s to full.", name),
		Undoable:   true,
		EntityType: "npc",
		EntityID:   instanceID,
		Before:     jsonRaw(map[string]int32{"hp": before}),
		After:      jsonRaw(map[string]int32{"hp": maxHP}),
		Inverse: inverseOf("restore-hp", map[string]interface{}{
			"npcInstanceId": instanceID,
			"hp":            before,
			"expectHp":      maxHP,
		}),
		Detail: map[string]interface{}{
			"npcInstanceId": instanceID,
			"hp":            maxHP,
			"previousHp":    before,
		},
	}, nil
}

// OpNPCRespawn brings a unique or template NPC into its spawn room when none is alive.
func (g *Game) OpNPCRespawn(templateID, roomID string) (*OpResult, error) {
	if templateID == "" {
		return nil, opErr(400, "npcTemplateId is required")
	}
	if g.NPCManager == nil {
		return nil, opErr(503, "npc manager is not running")
	}
	if alive := g.NPCManager.AliveMatching(templateID); alive != nil {
		return nil, opErr(409, "npc is already alive")
	}
	if dead := g.NPCManager.DeadMatching(templateID); dead != nil {
		return g.respawnExisting(dead, roomID)
	}
	stored, err := g.Facade.NPCsService().FindByID(templateID)
	if err == nil && stored != nil && !stored.IsTemplate {
		if !stored.IsDead && stored.CurrentHitPoints > 0 {
			g.NPCManager.RegisterExistingNPC(stored, firstRoom(roomID, stored.CurrentRoomID, stored.SpawnRoomID))
			g.NPCManager.Note(stored.ID, "registered living resident")
			return &OpResult{
				Summary:    opSummary("%s was already alive and is now in the running world.", npcName(stored)),
				Undoable:   false,
				EntityType: "npc",
				EntityID:   stored.ID,
				Detail: map[string]interface{}{
					"npcInstanceId": stored.ID,
					"note":          "The resident was already alive. Nothing was changed to undo.",
				},
			}, nil
		}
		g.NPCManager.RegisterExistingNPC(stored, firstRoom(roomID, stored.SpawnRoomID, stored.CurrentRoomID))
		return g.respawnExisting(g.NPCManager.GetInstance(stored.ID), roomID)
	}
	if stored == nil || !stored.IsTemplate {
		if err != nil || stored == nil {
			return nil, opErr(404, "npc template not found")
		}
	}
	spawnRoom := roomID
	if spawnRoom == "" && stored != nil {
		spawnRoom = stored.SpawnRoomID
	}
	if spawnRoom == "" {
		return nil, opErr(400, "roomId is required")
	}
	if _, rerr := g.Facade.RoomsService().FindByID(spawnRoom); rerr != nil {
		return nil, opErr(404, "room not found")
	}
	created, serr := g.NPCManager.SpawnInstanceDirect(templateID, spawnRoom)
	if serr != nil || created == nil {
		return nil, opErr(400, "could not spawn the npc")
	}
	g.NPCManager.Note(created.ID, "respawned")
	return &OpResult{
		Summary:    opSummary("Respawned %s in %s.", npcName(created), spawnRoom),
		Undoable:   true,
		EntityType: "npc",
		EntityID:   created.ID,
		After:      jsonRaw(npcSnap(created)),
		Inverse: inverseOf("npc-despawn", map[string]interface{}{
			"npcInstanceId": created.ID,
			"expectRoomId":  created.CurrentRoomID,
			"expectHp":      created.CurrentHitPoints,
			"expectAlive":   true,
		}),
		Detail: map[string]interface{}{
			"npcInstanceId": created.ID,
			"roomId":        created.CurrentRoomID,
			"templateId":    templateID,
		},
	}, nil
}

func (g *Game) respawnExisting(dead *npc.NPC, roomID string) (*OpResult, error) {
	if dead == nil || dead.Entity == nil {
		return nil, opErr(404, "npc instance not found")
	}
	before := cloneNPC(dead)
	room := firstRoom(roomID, dead.SpawnRoomID, dead.CurrentRoomID)
	if room == "" {
		return nil, opErr(400, "roomId is required")
	}
	if _, err := g.Facade.RoomsService().FindByID(room); err != nil {
		return nil, opErr(404, "room not found")
	}
	if !g.NPCManager.RespawnInstance(dead.ID) {
		return nil, opErr(409, "npc could not be respawned")
	}
	if roomID != "" || dead.CurrentRoomID != room {
		g.NPCManager.UpdateInstance(dead.ID, func(n *npc.NPC) {
			n.CurrentRoomID = room
			if n.SpawnRoomID == "" {
				n.SpawnRoomID = room
			}
		})
	}
	fresh := g.NPCManager.GetInstance(dead.ID)
	g.persistNPCIfStored(fresh)
	g.NPCManager.Note(dead.ID, "respawned")
	return &OpResult{
		Summary:    opSummary("Respawned %s in %s.", npcName(fresh), room),
		Undoable:   true,
		EntityType: "npc",
		EntityID:   dead.ID,
		Before:     jsonRaw(before),
		After:      jsonRaw(npcSnap(fresh)),
		Inverse: inverseOf("restore-npc", map[string]interface{}{
			"npc":           before,
			"expectRoomId":  fresh.CurrentRoomID,
			"expectHp":      fresh.CurrentHitPoints,
			"expectAlive":   !fresh.IsDead,
			"npcInstanceId": fresh.ID,
		}),
		Detail: map[string]interface{}{
			"npcInstanceId": fresh.ID,
			"roomId":        room,
		},
	}, nil
}

// OpNPCDespawn removes a living or dead instance from the running world.
// A fight must be ended first. A persisted unique is marked dead and kept as content.
func (g *Game) OpNPCDespawn(instanceID string) (*OpResult, error) {
	if instanceID == "" {
		return nil, opErr(400, "npcInstanceId is required")
	}
	if g.NPCManager == nil {
		return nil, opErr(503, "npc manager is not running")
	}
	inst := g.NPCManager.GetInstance(instanceID)
	if inst == nil {
		return nil, opErr(404, "npc instance not found")
	}
	if g.CombatController != nil && g.CombatController.IsNPCInCombat(instanceID) {
		return nil, opErr(409, "end the fight first")
	}
	before := cloneNPC(inst)
	g.NPCManager.RemoveInstance(instanceID)
	if stored, err := g.Facade.NPCsService().FindByID(instanceID); err == nil && stored != nil && !stored.IsTemplate {
		stored.IsDead = true
		stored.CurrentHitPoints = 0
		stored.InCombat = false
		stored.CombatInstanceID = ""
		stored.State = "dead"
		_ = g.Facade.NPCsService().Update(instanceID, stored)
	}
	g.NPCManager.Note(instanceID, "despawned")
	return &OpResult{
		Summary:    opSummary("Despawned %s.", npcName(before)),
		Undoable:   true,
		EntityType: "npc",
		EntityID:   instanceID,
		Before:     jsonRaw(before),
		Inverse: inverseOf("restore-npc", map[string]interface{}{
			"npc":           before,
			"expectMissing": true,
			"npcInstanceId": instanceID,
		}),
		Detail: map[string]interface{}{"npcInstanceId": instanceID},
	}, nil
}

func (g *Game) syncNPCCombatHP(instanceID string, hp int32) {
	if g.CombatController == nil {
		return
	}
	inst := g.CombatController.GetCombatInstanceByNPC(instanceID)
	if inst == nil {
		return
	}
	if ref := inst.GetCombatantByID(instanceID); ref != nil {
		ref.CurrentHP = hp
		if hp > 0 {
			ref.IsAlive = true
		}
		if ref.MaxHP < hp {
			ref.MaxHP = hp
		}
	}
}

func (g *Game) persistNPCIfStored(n *npc.NPC) {
	if g == nil || g.Facade == nil || n == nil || n.Entity == nil || n.ID == "" || n.IsTemplate {
		return
	}
	stored, err := g.Facade.NPCsService().FindByID(n.ID)
	if err != nil || stored == nil {
		return
	}
	_ = g.Facade.NPCsService().Update(n.ID, n)
}

func npcName(n *npc.NPC) string {
	if n == nil {
		return "NPC"
	}
	if n.Name != "" {
		return n.Name
	}
	return n.ID
}

func npcSnap(n *npc.NPC) map[string]interface{} {
	if n == nil {
		return nil
	}
	return map[string]interface{}{
		"id":     n.ID,
		"roomId": n.CurrentRoomID,
		"hp":     n.CurrentHitPoints,
		"alive":  !n.IsDead,
	}
}

func cloneNPC(n *npc.NPC) *npc.NPC {
	if n == nil {
		return nil
	}
	raw, err := json.Marshal(n)
	if err != nil {
		return nil
	}
	var out npc.NPC
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return &out
}

func firstRoom(ids ...string) string {
	for _, id := range ids {
		if id != "" {
			return id
		}
	}
	return ""
}
