package game

import (
	"sort"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/instances"
	"github.com/talesmud/talesmud/pkg/service"
)

type cloneGroup struct {
	id     string
	source string
	clones []string
}

// OpInstanceCleanup removes instance copies. allEmpty skips copies that still
// hold a player. A specific copy relocates anyone inside, the same way the
// startup sweep does, and then deletes the copy. This cannot be undone.
func (g *Game) OpInstanceCleanup(roomCopyID string, allEmpty bool) (*OpResult, error) {
	if roomCopyID == "" && !allEmpty {
		return nil, opErr(400, "roomCopyId or allEmpty is required")
	}
	if roomCopyID != "" && !instances.IsCloneID(roomCopyID) {
		return nil, opErr(400, "roomCopyId is not an instance copy")
	}
	all, err := loadRoomMap(g.Facade.RoomsService())
	if err != nil {
		return nil, opErr(500, "could not list rooms")
	}
	groups := cloneGroups(all)
	g.addLiveGroups(groups)
	var targets []*cloneGroup
	if roomCopyID != "" {
		id := instances.InstanceID(roomCopyID)
		grp := groups[id]
		if grp == nil {
			grp = groupContaining(groups, roomCopyID)
		}
		if grp == nil {
			return nil, opErr(404, "instance copy not found")
		}
		targets = []*cloneGroup{grp}
	} else {
		for _, grp := range groups {
			if g.groupOccupied(all, grp) {
				continue
			}
			targets = append(targets, grp)
		}
		sort.Slice(targets, func(i, j int) bool { return targets[i].id < targets[j].id })
	}
	startID := service.ResolveStartRoomID(g.Facade.ServerSettingsService(), g.Facade.RoomsService())
	if !roomPresent(all, startID) {
		startID = ""
	}
	moved := 0
	if roomCopyID != "" {
		n, merr := g.relocateGroup(all, targets[0], startID)
		if merr != nil {
			return nil, merr
		}
		moved = n
	}
	roomsDeleted := 0
	npcsDeleted := 0
	spawnersDeleted := 0
	for _, grp := range targets {
		g.dropMemoryNPCs(grp.clones)
		n, s := g.deleteGroupContent(grp)
		npcsDeleted += n
		spawnersDeleted += s
		roomsDeleted += g.deleteGroupRooms(grp)
	}
	summary := opSummary("Removed %d instance rooms.", roomsDeleted)
	if moved > 0 {
		summary = opSummary("Moved %d characters and removed %d instance rooms.", moved, roomsDeleted)
	}
	return &OpResult{
		Summary:    summary,
		Undoable:   false,
		EntityType: "instance",
		EntityID:   firstGroupID(targets, roomCopyID),
		Detail: map[string]interface{}{
			"roomsDeleted":    roomsDeleted,
			"npcsDeleted":     npcsDeleted,
			"spawnersDeleted": spawnersDeleted,
			"charactersMoved": moved,
			"note":            "Instance cleanup cannot be undone.",
		},
	}, nil
}

func firstGroupID(groups []*cloneGroup, fallback string) string {
	if len(groups) == 1 {
		return groups[0].id
	}
	if fallback != "" {
		return instances.InstanceID(fallback)
	}
	return "all-empty"
}

func cloneGroups(all map[string]*rooms.Room) map[string]*cloneGroup {
	out := map[string]*cloneGroup{}
	for id := range all {
		if !instances.IsCloneID(id) {
			continue
		}
		instID := instances.InstanceID(id)
		if instID == "" {
			instID = id
		}
		grp := out[instID]
		if grp == nil {
			grp = &cloneGroup{id: instID, source: instances.TemplateIDFromClone(id)}
			out[instID] = grp
		}
		grp.clones = append(grp.clones, id)
	}
	return out
}

func (g *Game) addLiveGroups(groups map[string]*cloneGroup) {
	if g.RoomInstances == nil || g.RoomInstances.mgr == nil {
		return
	}
	for _, copyRow := range g.RoomInstances.mgr.List() {
		grp := groups[copyRow.ID]
		if grp == nil {
			src := ""
			if len(copyRow.SourceIDs) > 0 {
				src = copyRow.SourceIDs[0]
			}
			grp = &cloneGroup{id: copyRow.ID, source: src}
			groups[copyRow.ID] = grp
		}
		seen := map[string]bool{}
		for _, id := range grp.clones {
			seen[id] = true
		}
		for _, id := range copyRow.CloneIDs {
			if !seen[id] {
				grp.clones = append(grp.clones, id)
			}
		}
	}
}

func groupContaining(groups map[string]*cloneGroup, roomID string) *cloneGroup {
	for _, grp := range groups {
		for _, id := range grp.clones {
			if id == roomID {
				return grp
			}
		}
	}
	return nil
}

func (g *Game) groupOccupied(all map[string]*rooms.Room, grp *cloneGroup) bool {
	set := cloneSet(grp)
	if g.RoomInstances != nil && g.RoomInstances.mgr != nil {
		for _, copyRow := range g.RoomInstances.mgr.List() {
			if copyRow.ID == grp.id && len(copyRow.PlayerIDs) > 0 {
				return true
			}
		}
	}
	if g.Sessions != nil {
		for _, player := range g.Sessions.all() {
			if set[player.RoomID] {
				return true
			}
		}
	}
	for id := range set {
		room := all[id]
		if room != nil && room.Characters != nil && len(*room.Characters) > 0 {
			return true
		}
	}
	chars, err := g.Facade.CharactersService().FindAll()
	if err != nil {
		return true
	}
	for _, ch := range chars {
		if ch != nil && set[ch.CurrentRoomID] {
			return true
		}
	}
	return false
}

func cloneSet(grp *cloneGroup) map[string]bool {
	set := map[string]bool{}
	if grp == nil {
		return set
	}
	for _, id := range grp.clones {
		set[id] = true
	}
	return set
}

func (g *Game) relocateGroup(all map[string]*rooms.Room, grp *cloneGroup, startID string) (int, error) {
	set := cloneSet(grp)
	chars, err := g.Facade.CharactersService().FindAll()
	if err != nil {
		return 0, opErr(500, "could not list characters")
	}
	type plan struct {
		ch   *characters.Character
		dest string
	}
	var plans []plan
	for _, ch := range chars {
		if ch == nil || !set[ch.CurrentRoomID] {
			continue
		}
		fallback := ""
		if roomPresent(all, ch.BoundRoomID) {
			fallback = ch.BoundRoomID
		}
		dest := instances.RelocationDest(all, ch.CurrentRoomID, startID, fallback)
		if dest == "" || all[dest] == nil {
			return 0, opErr(409, "no room to move "+ch.Name+" to")
		}
		plans = append(plans, plan{ch: ch, dest: dest})
	}
	moved := 0
	for _, plan := range plans {
		room, ok := g.RelocateCharacter(plan.ch, plan.ch.BelongsUserID, plan.dest)
		if !ok || room == nil {
			return moved, opErr(409, "could not move "+characterName(plan.ch))
		}
		if !g.characterOnline(plan.ch.ID) {
			g.noteRelocation(plan.ch.ID, room.Name)
		}
		moved++
	}
	return moved, nil
}

func (g *Game) dropMemoryNPCs(roomIDs []string) {
	if g.NPCManager == nil {
		return
	}
	set := map[string]bool{}
	for _, id := range roomIDs {
		set[id] = true
	}
	for _, n := range g.NPCManager.GetAllInstances() {
		if n != nil && set[n.CurrentRoomID] {
			g.NPCManager.RemoveInstance(n.ID)
		}
	}
}

func (g *Game) deleteGroupContent(grp *cloneGroup) (int, int) {
	set := cloneSet(grp)
	npcsDeleted := 0
	if g.Facade.NPCsService() != nil {
		rows, err := g.Facade.NPCsService().FindAll()
		if err == nil {
			for _, row := range rows {
				if row == nil || row.Entity == nil || !instances.IsCloneID(row.ID) {
					continue
				}
				if instances.InstanceID(row.ID) != grp.id && !set[row.CurrentRoomID] && !set[row.SpawnRoomID] {
					continue
				}
				if err := g.Facade.NPCsService().Delete(row.ID); err == nil {
					npcsDeleted++
				}
			}
		}
	}
	spawnersDeleted := 0
	if g.Facade.NPCSpawnersService() != nil {
		rows, err := g.Facade.NPCSpawnersService().FindAll()
		if err == nil {
			for _, row := range rows {
				if row == nil || row.Entity == nil {
					continue
				}
				if !set[row.RoomID] && !(instances.IsCloneID(row.ID) && instances.InstanceID(row.ID) == grp.id) {
					continue
				}
				if err := g.Facade.NPCSpawnersService().Delete(row.ID); err == nil {
					spawnersDeleted++
				}
			}
		}
	}
	return npcsDeleted, spawnersDeleted
}

func (g *Game) deleteGroupRooms(grp *cloneGroup) int {
	if g.RoomInstances != nil && g.RoomInstances.mgr != nil {
		g.RoomInstances.mgr.ForceDestroy(g.Facade.RoomsService(), grp.id)
	}
	n := 0
	for _, id := range grp.clones {
		if err := g.Facade.RoomsService().Delete(id); err == nil {
			n++
		}
	}
	if n == 0 && g.RoomInstances != nil {
		return len(grp.clones)
	}
	return n
}
