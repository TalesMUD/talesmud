package contenthealth

import (
	"sort"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/worldindex"
)

const (
	verdictGreen = "green"
	verdictAmber = "amber"
	verdictRed   = "red"
)

// Place is one world site that can satisfy a quest step.
type Place struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	Name      string `json:"name,omitempty"`
	How       string `json:"how,omitempty"`
	NodeID    string `json:"nodeId,omitempty"`
	Reachable bool   `json:"reachable"`
	Detail    string `json:"detail,omitempty"`
}

type objectiveEval struct {
	Index    int
	ID       string
	Type     string
	Verdict  string
	Reason   string
	Problems []string
	Places   []Place
}

type questEval struct {
	ID         string
	Eligible   bool
	Objectives []objectiveEval
}

type questReport struct {
	Obtainable map[string]bool
	Eligible   map[string]bool
	ByID       map[string]*questEval
}

// evaluateQuests decides which quests can complete and which reward items that unlocks.
// A reward item counts only after its quest is completable and that quest's prerequisites
// can be finished before any quest that needs the item. The scan repeats until it stops
// changing. A cycle never becomes eligible.
func evaluateQuests(ix *worldindex.Index, reach worldindex.Reachability) questReport {
	report := questReport{
		Obtainable: map[string]bool{},
		Eligible:   map[string]bool{},
		ByID:       map[string]*questEval{},
	}
	if ix == nil {
		return report
	}
	for id, ok := range ix.Obtainable(reach) {
		if ok && id != "" {
			report.Obtainable[id] = true
		}
	}
	snap := ix.Snapshot()
	ids := sortedIDs(snap.Quests)
	limit := len(ids) + 1
	for iter := 0; iter < limit; iter++ {
		changed := false
		for _, id := range ids {
			if report.Eligible[id] {
				continue
			}
			quest := snap.Quests[id]
			if quest == nil || !prereqsEligible(quest, snap.Quests, report.Eligible) {
				continue
			}
			if questBlocked(ix, reach, report.Obtainable, quest) {
				continue
			}
			report.Eligible[id] = true
			changed = true
			for _, itemID := range quest.Rewards.ItemTemplateIDs {
				if itemID != "" {
					report.Obtainable[itemID] = true
				}
			}
		}
		if !changed {
			break
		}
	}
	for _, id := range ids {
		quest := snap.Quests[id]
		if quest == nil {
			continue
		}
		ev := &questEval{ID: id, Eligible: report.Eligible[id]}
		for i, objective := range quest.Objectives {
			problems := objectiveProblems(ix, reach, report.Obtainable, objective)
			ev.Objectives = append(ev.Objectives, decorateObjective(ix, reach, report, id, i, objective, problems))
		}
		report.ByID[id] = ev
	}
	return report
}

func questBlocked(ix *worldindex.Index, reach worldindex.Reachability, obtainable map[string]bool, quest *quests.Quest) bool {
	for _, objective := range quest.Objectives {
		if len(objectiveProblems(ix, reach, obtainable, objective)) > 0 {
			return true
		}
	}
	return false
}

func prereqsEligible(quest *quests.Quest, all map[string]*quests.Quest, eligible map[string]bool) bool {
	if quest == nil {
		return false
	}
	for _, req := range quest.RequiredQuestIDs {
		req = strings.TrimSpace(req)
		if req == "" {
			continue
		}
		if all[req] == nil || !eligible[req] {
			return false
		}
	}
	return true
}

// requiresQuest reports whether questID's prerequisite chain includes target.
func requiresQuest(all map[string]*quests.Quest, questID, target string) bool {
	if questID == "" || target == "" {
		return false
	}
	seen := map[string]bool{}
	var walk func(string) bool
	walk = func(id string) bool {
		if id == "" || seen[id] {
			return false
		}
		seen[id] = true
		quest := all[id]
		if quest == nil {
			return false
		}
		for _, req := range quest.RequiredQuestIDs {
			req = strings.TrimSpace(req)
			if req == target || walk(req) {
				return true
			}
		}
		return false
	}
	return walk(questID)
}

func objectiveProblems(ix *worldindex.Index, reach worldindex.Reachability, obtainable map[string]bool, objective quests.Objective) []string {
	if ix == nil {
		return nil
	}
	snap := ix.Snapshot()
	switch objective.Type {
	case quests.ObjectiveVisit:
		if objective.TargetID == "" || snap.Rooms[objective.TargetID] == nil {
			return nil
		}
		if !reach.Reachable[objective.TargetID] {
			return []string{"visit room " + objective.TargetID + " is unreachable"}
		}
	case quests.ObjectiveKill:
		if objective.TargetID == "" || snap.NPCs[objective.TargetID] == nil {
			return nil
		}
		if !ix.CanMeet(snap.NPCs[objective.TargetID], reach) {
			return []string{"kill target " + objective.TargetID + " is never spawned in a reachable room"}
		}
	case quests.ObjectiveCollect:
		if objective.TargetID == "" || snap.Items[objective.TargetID] == nil {
			return nil
		}
		if !obtainable[objective.TargetID] {
			return []string{"collect item " + objective.TargetID + " is not obtainable"}
		}
	case quests.ObjectiveDeliver:
		var problems []string
		if objective.TargetID != "" && snap.Items[objective.TargetID] != nil && !obtainable[objective.TargetID] {
			problems = append(problems, "deliver item "+objective.TargetID+" is not obtainable")
		}
		if objective.DeliverToNPCID != "" && snap.NPCs[objective.DeliverToNPCID] != nil && !ix.CanMeet(snap.NPCs[objective.DeliverToNPCID], reach) {
			problems = append(problems, "deliver NPC "+objective.DeliverToNPCID+" is not reachable")
		}
		return problems
	case quests.ObjectiveTalk:
		if objective.TargetID == "" || snap.NPCs[objective.TargetID] == nil {
			return nil
		}
		if !ix.CanMeet(snap.NPCs[objective.TargetID], reach) {
			return []string{"talk NPC " + objective.TargetID + " is not reachable"}
		}
	}
	return nil
}

func decorateObjective(ix *worldindex.Index, reach worldindex.Reachability, report questReport, questID string, index int, objective quests.Objective, problems []string) objectiveEval {
	ev := objectiveEval{
		Index:    index,
		ID:       objective.ID,
		Type:     string(objective.Type),
		Problems: problems,
		Places:   placesForObjective(ix, reach, report, questID, objective),
	}
	if len(problems) > 0 {
		ev.Verdict = verdictRed
		ev.Reason = strings.Join(problems, "; ")
		return ev
	}
	if reason := amberReason(ix, objective); reason != "" {
		ev.Verdict = verdictAmber
		ev.Reason = reason
		return ev
	}
	ev.Verdict = verdictGreen
	ev.Reason = "can be satisfied"
	return ev
}

func amberReason(ix *worldindex.Index, objective quests.Objective) string {
	snap := ix.Snapshot()
	var notes []string
	add := func(note string) {
		if note != "" {
			notes = append(notes, note)
		}
	}
	switch objective.Type {
	case quests.ObjectiveCustom:
		if objective.CheckScriptID != "" && snap.Scripts[objective.CheckScriptID] == nil {
			add("check script " + objective.CheckScriptID + " is not in the world")
		}
		add("custom objective is not checked statically")
	case quests.ObjectiveVisit:
		add(missingTarget("room", objective.TargetID, snap.Rooms[objective.TargetID] != nil))
	case quests.ObjectiveKill:
		add(missingTarget("NPC", objective.TargetID, snap.NPCs[objective.TargetID] != nil))
	case quests.ObjectiveCollect:
		add(missingTarget("item", objective.TargetID, snap.Items[objective.TargetID] != nil))
	case quests.ObjectiveDeliver:
		if objective.TargetID != "" && snap.Items[objective.TargetID] == nil {
			add("item " + objective.TargetID + " is not in the world")
		}
		if objective.DeliverToNPCID != "" && snap.NPCs[objective.DeliverToNPCID] == nil {
			add("NPC " + objective.DeliverToNPCID + " is not in the world")
		}
		if objective.TargetID == "" && objective.DeliverToNPCID == "" {
			add("deliver objective has no item or NPC")
		}
	case quests.ObjectiveTalk:
		add(missingTarget("NPC", objective.TargetID, objective.TargetID == "" || snap.NPCs[objective.TargetID] != nil))
		if objective.DialogNodeID != "" && objective.TargetID != "" && snap.NPCs[objective.TargetID] != nil && !npcHasDialogNode(snap, snap.NPCs[objective.TargetID], objective.DialogNodeID) {
			add("dialog node " + objective.DialogNodeID + " is not on a dialog this NPC uses")
		}
	default:
		if objective.Type != "" {
			add("objective type " + string(objective.Type) + " is not checked statically")
		}
	}
	return strings.Join(notes, "; ")
}

func missingTarget(kind, id string, exists bool) string {
	if id == "" {
		return kind + " target is not set"
	}
	if !exists {
		return kind + " " + id + " is not in the world"
	}
	return ""
}

func (ev *questEval) verdict() (string, string) {
	if ev == nil || len(ev.Objectives) == 0 {
		return verdictGreen, "can complete"
	}
	verdict := verdictGreen
	var reds, ambers []string
	for _, obj := range ev.Objectives {
		switch obj.Verdict {
		case verdictRed:
			verdict = verdictRed
			if obj.Reason != "" {
				reds = append(reds, obj.Reason)
			}
		case verdictAmber:
			if verdict != verdictRed {
				verdict = verdictAmber
			}
			if obj.Reason != "" {
				ambers = append(ambers, obj.Reason)
			}
		}
	}
	switch verdict {
	case verdictRed:
		return verdictRed, strings.Join(reds, "; ")
	case verdictAmber:
		return verdictAmber, strings.Join(ambers, "; ")
	default:
		return verdictGreen, "can complete"
	}
}

func placesForObjective(ix *worldindex.Index, reach worldindex.Reachability, report questReport, questID string, objective quests.Objective) []Place {
	seen := map[string]bool{}
	var places []Place
	add := func(p Place) {
		addPlace(&places, seen, p)
	}
	switch objective.Type {
	case quests.ObjectiveVisit:
		add(roomPlace(ix, reach, objective.TargetID, "visit target"))
	case quests.ObjectiveKill:
		for _, place := range npcPlaces(ix, reach, objective.TargetID) {
			add(place)
		}
	case quests.ObjectiveCollect:
		for _, place := range itemPlaces(ix, reach, report, questID, objective.TargetID) {
			add(place)
		}
	case quests.ObjectiveDeliver:
		for _, place := range itemPlaces(ix, reach, report, questID, objective.TargetID) {
			add(place)
		}
		for _, place := range npcPlaces(ix, reach, objective.DeliverToNPCID) {
			add(place)
		}
	case quests.ObjectiveTalk:
		for _, place := range npcPlaces(ix, reach, objective.TargetID) {
			add(place)
		}
		for _, place := range npcDialogNodePlaces(ix, reach, objective.TargetID, objective.DialogNodeID) {
			add(place)
		}
		for _, place := range questDialogPlaces(ix, reach, questID, "progress") {
			add(place)
		}
	}
	if objective.CheckScriptID != "" {
		add(scriptPlace(ix, reach, objective.CheckScriptID, "objective script"))
	}
	if objective.Type == quests.ObjectiveCustom && objective.CheckScriptID == "" {
		add(Place{Type: "script", How: "custom check", Detail: "no check script"})
	}
	sortPlaces(places)
	return places
}

func addPlace(dst *[]Place, seen map[string]bool, p Place) {
	if p.Type == "" && p.ID == "" {
		return
	}
	key := p.Type + "\x00" + p.ID + "\x00" + p.How + "\x00" + p.NodeID
	if seen[key] {
		return
	}
	seen[key] = true
	*dst = append(*dst, p)
}

func sortPlaces(places []Place) {
	sort.Slice(places, func(i, j int) bool {
		if places[i].Type != places[j].Type {
			return places[i].Type < places[j].Type
		}
		if places[i].ID != places[j].ID {
			return places[i].ID < places[j].ID
		}
		if places[i].NodeID != places[j].NodeID {
			return places[i].NodeID < places[j].NodeID
		}
		return places[i].How < places[j].How
	})
}

func roomPlace(ix *worldindex.Index, reach worldindex.Reachability, roomID, how string) Place {
	if roomID == "" {
		return Place{}
	}
	snap := ix.Snapshot()
	exists := snap.Rooms[roomID] != nil
	return Place{
		Type:      "room",
		ID:        roomID,
		Name:      ix.Name(worldindex.KindRoom, roomID),
		How:       how,
		Reachable: exists && reach.Reachable[roomID],
	}
}

func scriptPlace(ix *worldindex.Index, reach worldindex.Reachability, scriptID, how string) Place {
	if scriptID == "" {
		return Place{}
	}
	exists := ix.Snapshot().Scripts[scriptID] != nil
	reachable := reach.InvokedScripts[scriptID]
	if how == "objective script" {
		reachable = exists
	}
	return Place{
		Type:      "script",
		ID:        scriptID,
		Name:      ix.Name(worldindex.KindScript, scriptID),
		How:       how,
		Reachable: reachable,
	}
}

func npcPlaces(ix *worldindex.Index, reach worldindex.Reachability, npcID string) []Place {
	if npcID == "" {
		return nil
	}
	snap := ix.Snapshot()
	n := snap.NPCs[npcID]
	var places []Place
	seen := map[string]bool{}
	add := func(p Place) { addPlace(&places, seen, p) }
	meet := n != nil && ix.CanMeet(n, reach)
	add(Place{
		Type:      "npc",
		ID:        npcID,
		Name:      ix.Name(worldindex.KindNPC, npcID),
		How:       "target",
		Reachable: meet,
	})
	if n == nil {
		return places
	}
	if n.SpawnRoomID != "" {
		add(roomPlace(ix, reach, n.SpawnRoomID, "spawn room"))
	}
	if n.CurrentRoomID != "" && n.CurrentRoomID != n.SpawnRoomID {
		add(roomPlace(ix, reach, n.CurrentRoomID, "current room"))
	}
	for _, edge := range ix.Inbound(worldindex.KindNPC, npcID) {
		switch {
		case edge.FromType == worldindex.KindSpawner && edge.How == "spawner template":
			if spawner := snap.Spawners[edge.FromID]; spawner != nil && spawner.RoomID != "" {
				add(roomPlace(ix, reach, spawner.RoomID, "spawner "+edge.FromID))
			}
		case edge.FromType == worldindex.KindScript && (edge.How == "summon" || edge.How == "spawnFromTemplate"):
			add(Place{
				Type:      "script",
				ID:        edge.FromID,
				Name:      ix.Name(worldindex.KindScript, edge.FromID),
				How:       edge.How,
				Reachable: reach.InvokedScripts[edge.FromID],
			})
		case edge.FromType == worldindex.KindRoom && edge.How == "room resident":
			add(roomPlace(ix, reach, edge.FromID, "room resident"))
		}
	}
	return places
}

func itemPlaces(ix *worldindex.Index, reach worldindex.Reachability, report questReport, dependentQuest, itemID string) []Place {
	if itemID == "" {
		return nil
	}
	snap := ix.Snapshot()
	var places []Place
	seen := map[string]bool{}
	add := func(p Place) { addPlace(&places, seen, p) }
	for _, edge := range ix.Inbound(worldindex.KindItem, itemID) {
		switch edge.FromType {
		case worldindex.KindRoom:
			if edge.How == "room item" {
				add(roomPlace(ix, reach, edge.FromID, "room item"))
			}
		case worldindex.KindNPC:
			n := snap.NPCs[edge.FromID]
			add(Place{
				Type:      "npc",
				ID:        edge.FromID,
				Name:      ix.Name(worldindex.KindNPC, edge.FromID),
				How:       edge.How,
				Reachable: n != nil && ix.CanMeet(n, reach),
			})
		case worldindex.KindLootTable:
			for _, npcEdge := range ix.Inbound(worldindex.KindLootTable, edge.FromID) {
				if npcEdge.FromType != worldindex.KindNPC {
					continue
				}
				n := snap.NPCs[npcEdge.FromID]
				add(Place{
					Type:      "npc",
					ID:        npcEdge.FromID,
					Name:      ix.Name(worldindex.KindNPC, npcEdge.FromID),
					How:       edge.How,
					Reachable: n != nil && ix.CanMeet(n, reach),
					Detail:    edge.FromID,
				})
			}
		case worldindex.KindScript:
			if edge.How == "giveItem" {
				add(Place{
					Type:      "script",
					ID:        edge.FromID,
					Name:      ix.Name(worldindex.KindScript, edge.FromID),
					How:       "giveItem",
					Reachable: reach.InvokedScripts[edge.FromID],
				})
			}
		case worldindex.KindQuest:
			if edge.How != "quest reward" {
				continue
			}
			precedes := report.Eligible[edge.FromID] && !requiresQuest(snap.Quests, edge.FromID, dependentQuest)
			detail := ""
			if !precedes {
				if requiresQuest(snap.Quests, edge.FromID, dependentQuest) {
					detail = "prerequisites cannot precede this quest"
				} else if !report.Eligible[edge.FromID] {
					detail = "reward quest cannot complete"
				}
			}
			add(Place{
				Type:      "quest",
				ID:        edge.FromID,
				Name:      ix.Name(worldindex.KindQuest, edge.FromID),
				How:       "quest reward",
				Reachable: precedes,
				Detail:    detail,
			})
		case worldindex.KindCharacterTemplate:
			add(Place{
				Type:      "characterTemplate",
				ID:        edge.FromID,
				Name:      ix.Name(worldindex.KindCharacterTemplate, edge.FromID),
				How:       edge.How,
				Reachable: false,
				Detail:    "starting kit is not a world source",
			})
		}
	}
	return places
}

func npcDialogNodePlaces(ix *worldindex.Index, reach worldindex.Reachability, npcID, nodeID string) []Place {
	if npcID == "" || nodeID == "" {
		return nil
	}
	snap := ix.Snapshot()
	n := snap.NPCs[npcID]
	if n == nil {
		return nil
	}
	var ids []string
	if n.DialogID != "" {
		ids = append(ids, n.DialogID)
	}
	if n.IdleDialogID != "" && n.IdleDialogID != n.DialogID {
		ids = append(ids, n.IdleDialogID)
	}
	meet := ix.CanMeet(n, reach)
	var places []Place
	for _, id := range ids {
		dialog := snap.Dialogs[id]
		found := dialog != nil && dialog.FindDialog(nodeID) != nil
		detail := ""
		if !found {
			detail = "node not found"
		}
		places = append(places, Place{
			Type:      "dialog",
			ID:        id,
			Name:      ix.Name(worldindex.KindDialog, id),
			How:       "dialog node",
			NodeID:    nodeID,
			Reachable: meet && found,
			Detail:    detail,
		})
	}
	return places
}

func questDialogPlaces(ix *worldindex.Index, reach worldindex.Reachability, questID, action string) []Place {
	if ix == nil || questID == "" {
		return nil
	}
	var places []Place
	seen := map[string]bool{}
	for _, ref := range dialogQuestRefs(ix.Snapshot(), questID) {
		if action != "" && !strings.EqualFold(ref.action, action) {
			continue
		}
		addPlace(&places, seen, Place{
			Type:      "dialog",
			ID:        ref.dialogID,
			Name:      ix.Name(worldindex.KindDialog, ref.dialogID),
			How:       dialogHow(ref.action),
			NodeID:    ref.nodeID,
			Reachable: dialogMeetable(ix, reach, ref.dialogID),
			Detail:    trimText(ref.text),
		})
	}
	return places
}

type dialogQuestRef struct {
	dialogID string
	nodeID   string
	action   string
	text     string
}

func dialogQuestRefs(snap worldindex.Snapshot, questID string) []dialogQuestRef {
	var out []dialogQuestRef
	for _, id := range sortedIDs(snap.Dialogs) {
		walkDialogQuest(id, snap.Dialogs[id], questID, &out)
	}
	return out
}

func walkDialogQuest(dialogID string, node *dialogs.Dialog, questID string, out *[]dialogQuestRef) {
	if node == nil {
		return
	}
	if node.QuestID == questID {
		*out = append(*out, dialogQuestRef{
			dialogID: dialogID,
			nodeID:   node.NodeID,
			action:   node.Action,
			text:     node.Text,
		})
	}
	for _, option := range node.Options {
		walkDialogQuest(dialogID, option, questID, out)
	}
	if node.Answer != nil {
		walkDialogQuest(dialogID, node.Answer, questID, out)
	}
}

func dialogMeetable(ix *worldindex.Index, reach worldindex.Reachability, dialogID string) bool {
	for _, edge := range ix.Inbound(worldindex.KindDialog, dialogID) {
		if edge.FromType != worldindex.KindNPC {
			continue
		}
		n := ix.Snapshot().NPCs[edge.FromID]
		if n != nil && ix.CanMeet(n, reach) {
			return true
		}
	}
	return false
}

func dialogHow(action string) string {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "accept":
		return "accepts quest"
	case "complete":
		return "turns in quest"
	case "progress":
		return "progresses quest"
	default:
		if action == "" {
			return "dialog quest link"
		}
		return action
	}
}

func npcHasDialogNode(snap worldindex.Snapshot, n *npc.NPC, nodeID string) bool {
	if n == nil || nodeID == "" {
		return false
	}
	for _, id := range []string{n.DialogID, n.IdleDialogID} {
		if id == "" {
			continue
		}
		dialog := snap.Dialogs[id]
		if dialog != nil && dialog.FindDialog(nodeID) != nil {
			return true
		}
	}
	return false
}

func trimText(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	if len(text) > 80 {
		return text[:77] + "..."
	}
	return text
}

func worseVerdict(current, next string) string {
	if verdictRank(next) > verdictRank(current) {
		return next
	}
	return current
}

func verdictRank(verdict string) int {
	switch verdict {
	case verdictRed:
		return 3
	case verdictAmber:
		return 2
	default:
		return 1
	}
}
