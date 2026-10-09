package worldindex

import (
	"fmt"
	"sort"
	"strings"

	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

// Island is one connected set of rooms that cannot be reached from the start room.
type Island struct {
	RoomIDs []string `json:"roomIds"`
	Reason  string   `json:"reason"`
}

// Reachability is a BFS from the start room.
// Hidden exits open only when a revealExit(room, exitName) script is invoked
// from a place the player can already reach: a room action, an on-enter script,
// an item use, or an NPC hook on an NPC the player can meet.
type Reachability struct {
	StartRoomID    string          `json:"startRoomId"`
	Reachable      map[string]bool `json:"reachable"`
	Islands        []Island        `json:"islands"`
	InvokedScripts map[string]bool `json:"invokedScripts,omitempty"`
}

// Reachability computes the reachable set and the unreachable islands.
// Instance-tagged rooms use the same exit graph as every other room.
func (ix *Index) Reachability() Reachability {
	report := Reachability{
		StartRoomID: ix.snap.StartRoomID,
		Reachable:   map[string]bool{},
	}
	if ix == nil {
		return report
	}
	start := ix.snap.StartRoomID
	if start == "" || ix.snap.Rooms[start] == nil {
		report.Islands = ix.islands(report.Reachable, nil)
		return report
	}

	openHidden := map[string]bool{}
	invoked := map[string]bool{}
	reachable := map[string]bool{}
	for {
		next := ix.bfs(start, openHidden)
		changed := len(next) != len(reachable)
		reachable = next
		if ix.markInvoked(reachable, invoked) {
			changed = true
		}
		if ix.openRevealed(reachable, invoked, openHidden) {
			changed = true
		}
		if !changed {
			break
		}
	}
	report.Reachable = reachable
	report.InvokedScripts = invoked
	report.Islands = ix.islands(reachable, invoked)
	return report
}

func (ix *Index) bfs(start string, openHidden map[string]bool) map[string]bool {
	reachable := map[string]bool{}
	if ix.snap.Rooms[start] == nil {
		return reachable
	}
	queue := []string{start}
	reachable[start] = true
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		room := ix.snap.Rooms[id]
		if room == nil || room.Exits == nil {
			continue
		}
		for _, exit := range *room.Exits {
			if exit.Target == "" || ix.snap.Rooms[exit.Target] == nil {
				continue
			}
			if exit.Hidden && !openHidden[hiddenKey(id, exit.Name)] {
				continue
			}
			if reachable[exit.Target] {
				continue
			}
			reachable[exit.Target] = true
			queue = append(queue, exit.Target)
		}
	}
	return reachable
}

func hiddenKey(roomID, exitName string) string {
	return strings.ToLower(strings.TrimSpace(roomID)) + "\x00" + strings.ToLower(strings.TrimSpace(exitName))
}

func (ix *Index) markInvoked(reachable, invoked map[string]bool) bool {
	changed := false
	mark := func(id string) {
		if id == "" || invoked[id] || ix.snap.Scripts[id] == nil {
			return
		}
		invoked[id] = true
		changed = true
	}
	for id, room := range ix.snap.Rooms {
		if !reachable[id] || room == nil {
			continue
		}
		mark(room.OnEnterScriptID)
		if room.Actions == nil {
			continue
		}
		for _, action := range *room.Actions {
			mark(action.ScriptId)
		}
	}
	for _, n := range ix.snap.NPCs {
		if n == nil || n.EnemyTrait == nil || !ix.npcReachable(n, reachable, invoked) {
			continue
		}
		et := n.EnemyTrait
		mark(et.OnAggroScript)
		mark(et.OnDeathScript)
		mark(et.OnFleeScript)
		mark(et.OnLowHealthScript)
	}
	obtainable := ix.obtainableItems(reachable, invoked)
	for id, item := range ix.snap.Items {
		if item == nil || !obtainable[id] {
			continue
		}
		mark(item.OnUseScriptID)
		mark(item.OnHitScriptID)
	}
	for _, quest := range ix.snap.Quests {
		if quest == nil || !ix.questOffered(quest, reachable, obtainable) {
			continue
		}
		mark(quest.OnAcceptScriptID)
		mark(quest.OnCompleteScriptID)
		for _, objective := range quest.Objectives {
			mark(objective.CheckScriptID)
		}
	}
	return changed
}

func (ix *Index) openRevealed(reachable, invoked, openHidden map[string]bool) bool {
	changed := false
	for scriptID := range invoked {
		for _, edge := range ix.Outbound(KindScript, scriptID) {
			if edge.How != "revealExit" || edge.ExitName == "" {
				continue
			}
			for _, roomID := range ix.RevealTargets(edge) {
				// ctx.room is the room the player is standing in, so a current-room
				// reveal only opens an exit on a room that can already run the script.
				if edge.ToID == CurrentRoom && !reachable[roomID] {
					continue
				}
				room := ix.snap.Rooms[roomID]
				if room == nil || !roomHasExit(room, edge.ExitName) {
					continue
				}
				k := hiddenKey(roomID, edge.ExitName)
				if openHidden[k] {
					continue
				}
				exit, ok := room.GetExit(edge.ExitName)
				if !ok || !exit.Hidden {
					continue
				}
				openHidden[k] = true
				changed = true
			}
		}
	}
	return changed
}

// RevealTargets lists rooms a revealExit edge applies to.
// A current-room reveal applies to every room that runs the script.
func (ix *Index) RevealTargets(edge Edge) []string {
	if ix == nil || edge.How != "revealExit" || edge.ExitName == "" {
		return nil
	}
	if edge.ToID == CurrentRoom {
		return ix.scriptRooms(edge.FromID)
	}
	if edge.ToType == KindRoom && edge.ToID != "" && edge.ToID != CurrentRoom {
		return []string{edge.ToID}
	}
	return nil
}

func (ix *Index) scriptRooms(scriptID string) []string {
	seen := map[string]bool{}
	var rooms []string
	for _, edge := range ix.Inbound(KindScript, scriptID) {
		if edge.FromType != KindRoom || edge.FromID == "" || seen[edge.FromID] {
			continue
		}
		seen[edge.FromID] = true
		rooms = append(rooms, edge.FromID)
	}
	sort.Strings(rooms)
	return rooms
}

func roomHasExit(room *rooms.Room, name string) bool {
	_, ok := room.GetExit(name)
	return ok
}

// NPCReachable reports whether a player who can walk the reachable set can meet this NPC.
func (ix *Index) NPCReachable(n *npc.NPC, reachable map[string]bool) bool {
	if ix == nil || n == nil {
		return false
	}
	invoked := map[string]bool{}
	// Callers that already expanded scripts pass reachability only.
	// Summons are handled by npcReachable when the script set is known.
	return ix.npcReachable(n, reachable, invoked)
}

func (ix *Index) npcReachable(n *npc.NPC, reachable, invoked map[string]bool) bool {
	if n == nil {
		return false
	}
	if n.SpawnRoomID != "" && reachable[n.SpawnRoomID] {
		return true
	}
	if n.CurrentRoomID != "" && reachable[n.CurrentRoomID] {
		return true
	}
	for _, spawner := range ix.snap.Spawners {
		if spawner == nil || spawner.TemplateID != n.ID {
			continue
		}
		if reachable[spawner.RoomID] {
			return true
		}
	}
	if n.ID == "" {
		return false
	}
	for _, edge := range ix.Inbound(KindNPC, n.ID) {
		if edge.FromType != KindScript || !invoked[edge.FromID] {
			continue
		}
		if edge.How == "summon" || edge.How == "spawnFromTemplate" {
			return true
		}
	}
	return false
}

// ObtainableItems lists item template ids a player can get from a reachable source.
func (ix *Index) ObtainableItems(reachable map[string]bool) map[string]bool {
	return ix.obtainableItems(reachable, nil)
}

func (ix *Index) obtainableItems(reachable, invoked map[string]bool) map[string]bool {
	out := map[string]bool{}
	for roomID, room := range ix.snap.Rooms {
		if !reachable[roomID] || room == nil || room.Items == nil {
			continue
		}
		for _, itemID := range *room.Items {
			if itemID != "" {
				out[itemID] = true
			}
		}
	}
	reachableNPC := map[string]bool{}
	for id, n := range ix.snap.NPCs {
		if ix.npcReachable(n, reachable, invoked) {
			reachableNPC[id] = true
		}
	}
	lootUsable := map[string]bool{}
	for id, n := range ix.snap.NPCs {
		if !reachableNPC[id] || n == nil || n.EnemyTrait == nil {
			continue
		}
		if n.EnemyTrait.LootTableID != "" {
			lootUsable[n.EnemyTrait.LootTableID] = true
		}
		for _, itemID := range n.EnemyTrait.GuaranteedLoot {
			if itemID != "" {
				out[itemID] = true
			}
		}
	}
	for id, table := range ix.snap.LootTables {
		if !lootUsable[id] || table == nil {
			continue
		}
		for _, entry := range table.Entries {
			if entry.ItemTemplateID != "" {
				out[entry.ItemTemplateID] = true
			}
		}
	}
	for id, n := range ix.snap.NPCs {
		if !reachableNPC[id] || n == nil || n.MerchantTrait == nil {
			continue
		}
		for _, stock := range n.MerchantTrait.Inventory {
			if stock.ItemTemplateID != "" {
				out[stock.ItemTemplateID] = true
			}
		}
	}
	for scriptID := range invoked {
		for _, edge := range ix.Outbound(KindScript, scriptID) {
			if edge.How == "giveItem" && edge.ToType == KindItem && edge.ToID != "" {
				out[edge.ToID] = true
			}
		}
	}
	return out
}

func (ix *Index) questOffered(quest *quests.Quest, reachable, obtainable map[string]bool) bool {
	switch strings.ToLower(quest.Source.Type) {
	case "auto", "script", "":
		if quest.Source.NPCID == "" && quest.Source.ItemID == "" {
			return true
		}
	}
	if quest.Source.NPCID != "" {
		if n := ix.snap.NPCs[quest.Source.NPCID]; n != nil && ix.npcReachable(n, reachable, nil) {
			return true
		}
	}
	if quest.Source.ItemID != "" && obtainable[quest.Source.ItemID] {
		return true
	}
	return false
}

func (ix *Index) dialogReachable(dialogID string, reachable map[string]bool, invoked map[string]bool) bool {
	for _, edge := range ix.Inbound(KindDialog, dialogID) {
		if edge.FromType != KindNPC {
			continue
		}
		if n := ix.snap.NPCs[edge.FromID]; n != nil && ix.npcReachable(n, reachable, invoked) {
			return true
		}
	}
	return false
}

func (ix *Index) islands(reachable, invoked map[string]bool) []Island {
	if invoked == nil {
		invoked = map[string]bool{}
	}
	var unreachable []string
	for id := range ix.snap.Rooms {
		if id == "" || reachable[id] {
			continue
		}
		unreachable = append(unreachable, id)
	}
	sort.Strings(unreachable)
	if len(unreachable) == 0 {
		return nil
	}
	inIsland := map[string]bool{}
	for _, id := range unreachable {
		inIsland[id] = true
	}
	adj := map[string][]string{}
	add := func(a, b string) {
		if a == "" || b == "" || a == b || !inIsland[a] || !inIsland[b] {
			return
		}
		adj[a] = append(adj[a], b)
	}
	for _, id := range unreachable {
		room := ix.snap.Rooms[id]
		if room == nil || room.Exits == nil {
			continue
		}
		for _, exit := range *room.Exits {
			add(id, exit.Target)
			add(exit.Target, id)
		}
	}
	seen := map[string]bool{}
	var islands []Island
	for _, id := range unreachable {
		if seen[id] {
			continue
		}
		var component []string
		queue := []string{id}
		seen[id] = true
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			component = append(component, cur)
			for _, next := range adj[cur] {
				if seen[next] {
					continue
				}
				seen[next] = true
				queue = append(queue, next)
			}
		}
		sort.Strings(component)
		islands = append(islands, Island{
			RoomIDs: component,
			Reason:  ix.islandReason(component, inIsland, reachable, invoked),
		})
	}
	sort.Slice(islands, func(i, j int) bool {
		return islands[i].RoomIDs[0] < islands[j].RoomIDs[0]
	})
	return islands
}

func (ix *Index) islandReason(component []string, inIsland, reachable, invoked map[string]bool) string {
	member := map[string]bool{}
	for _, id := range component {
		member[id] = true
	}
	var notes []string
	seen := map[string]bool{}
	add := func(note string) {
		note = strings.TrimSpace(note)
		if note == "" || seen[note] {
			return
		}
		seen[note] = true
		notes = append(notes, note)
	}

	parents := map[string]bool{}
	for _, id := range component {
		room := ix.snap.Rooms[id]
		if room == nil || room.Exits == nil {
			continue
		}
		for _, exit := range *room.Exits {
			if exit.Target != "" && !member[exit.Target] && ix.snap.Rooms[exit.Target] != nil {
				parents[exit.Target] = true
			}
		}
	}

	for _, script := range ix.snap.Scripts {
		if script == nil {
			continue
		}
		for _, edge := range ix.Outbound(KindScript, script.ID) {
			if edge.How != "revealExit" || edge.ExitName == "" {
				continue
			}
			for _, roomID := range ix.RevealTargets(edge) {
				if !parents[roomID] && !ix.revealPointsAtRoom(roomID, edge.ExitName, member) {
					continue
				}
				room := ix.snap.Rooms[roomID]
				exit, has := rooms.Exit{}, false
				if room != nil {
					exit, has = room.GetExit(edge.ExitName)
				}
				missing := room == nil || !has
				pointsHere := has && member[exit.Target]
				if !missing && !pointsHere {
					continue
				}
				if missing {
					add(fmt.Sprintf("revealExit target '%s' missing on %s (%s)", edge.ExitName, roomID, script.ID))
					add(fmt.Sprintf("%s no '%s'", roomID, edge.ExitName))
				}
				if !ix.scriptReferenced(script.ID) {
					add(fmt.Sprintf("%s never invoked", script.ID))
				} else if !invoked[script.ID] && (missing || (has && exit.Hidden && !pointsHere)) {
					add(fmt.Sprintf("%s never invoked", script.ID))
				} else if has && exit.Hidden && pointsHere && !invoked[script.ID] {
					add(fmt.Sprintf("%s never invoked", script.ID))
				}
			}
		}
	}

	for parentID := range parents {
		room := ix.snap.Rooms[parentID]
		if room == nil || room.Exits == nil || !reachable[parentID] {
			continue
		}
		for _, exit := range *room.Exits {
			if !exit.Hidden || !member[exit.Target] {
				continue
			}
			if !ix.exitHasRevealer(parentID, exit.Name) {
				add(fmt.Sprintf("hidden exit '%s' on %s has no revealer", exit.Name, parentID))
			}
		}
	}

	if len(notes) == 0 {
		return "no inbound exit"
	}
	sort.Strings(notes)
	return strings.Join(notes, "; ")
}

func (ix *Index) revealPointsAtRoom(roomID, exitName string, member map[string]bool) bool {
	room := ix.snap.Rooms[roomID]
	if room == nil {
		return false
	}
	exit, ok := room.GetExit(exitName)
	return ok && member[exit.Target]
}

func (ix *Index) scriptReferenced(scriptID string) bool {
	return len(ix.Inbound(KindScript, scriptID)) > 0
}

func (ix *Index) exitHasRevealer(roomID, exitName string) bool {
	for _, script := range ix.snap.Scripts {
		if script == nil {
			continue
		}
		for _, edge := range ix.Outbound(KindScript, script.ID) {
			if edge.How != "revealExit" || !strings.EqualFold(edge.ExitName, exitName) {
				continue
			}
			for _, target := range ix.RevealTargets(edge) {
				if target == roomID {
					return true
				}
			}
		}
	}
	return false
}

// CanMeet reports whether a player who already reached the given set can meet this NPC,
// including summons from scripts that reachability marked as invoked.
func (ix *Index) CanMeet(n *npc.NPC, reach Reachability) bool {
	if ix == nil || n == nil {
		return false
	}
	return ix.npcReachable(n, reach.Reachable, reach.InvokedScripts)
}

// Obtainable lists item template ids a player can get from the reachable set,
// including giveItem from scripts that reachability marked as invoked.
func (ix *Index) Obtainable(reach Reachability) map[string]bool {
	if ix == nil {
		return map[string]bool{}
	}
	return ix.obtainableItems(reach.Reachable, reach.InvokedScripts)
}

// ScriptReferenced reports whether any other entity points at this script.
func (ix *Index) ScriptReferenced(scriptID string) bool {
	if ix == nil || scriptID == "" {
		return false
	}
	return ix.scriptReferenced(scriptID)
}

// ExitHasRevealer reports whether a script calls revealExit for this exit.
func (ix *Index) ExitHasRevealer(roomID, exitName string) bool {
	if ix == nil {
		return false
	}
	return ix.exitHasRevealer(roomID, exitName)
}
