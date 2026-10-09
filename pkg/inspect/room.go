package inspect

import (
	"sort"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/worldindex"
)

// ExitView is one exit into or out of the inspected room.
// RoomID is the other room: the target for an outbound exit, the source for an inbound exit.
type ExitView struct {
	Name       string `json:"name"`
	RoomID     string `json:"roomId,omitempty"`
	RoomName   string `json:"roomName,omitempty"`
	Missing    bool   `json:"missing,omitempty"`
	Hidden     bool   `json:"hidden,omitempty"`
	Instance   bool   `json:"instance,omitempty"`
	NoRevealer bool   `json:"noRevealer,omitempty"`
	RevealedBy []Link `json:"revealedBy,omitempty"`
}

// RoomAction is one authored room action.
type RoomAction struct {
	Name   string `json:"name,omitempty"`
	Type   string `json:"type,omitempty"`
	Script *Link  `json:"script,omitempty"`
}

// RoomActor is a content NPC placed in this room, not a live instance.
type RoomActor struct {
	ID      string   `json:"id"`
	Name    string   `json:"name,omitempty"`
	Missing bool     `json:"missing,omitempty"`
	How     string   `json:"how"`
	Level   int32    `json:"level,omitempty"`
	Kinds   []string `json:"kinds,omitempty"`
}

// ReachInfo is worldindex reachability for this room from the start room.
type ReachInfo struct {
	StartRoomID string `json:"startRoomId,omitempty"`
	Reachable   bool   `json:"reachable"`
	Reason      string `json:"reason,omitempty"`
}

// RoomView is the static half of the room inspector.
// RequestedID is set when the caller asked for an instance copy and the view is its template.
type RoomView struct {
	ID          string        `json:"id"`
	RequestedID string        `json:"requestedId,omitempty"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Area        string        `json:"area,omitempty"`
	RoomType    string        `json:"roomType,omitempty"`
	Tags        []string      `json:"tags"`
	Reach       ReachInfo     `json:"reach"`
	ExitsOut    []ExitView    `json:"exitsOut"`
	ExitsIn     []ExitView    `json:"exitsIn"`
	Actors      []RoomActor   `json:"actors"`
	Spawners    []SpawnerView `json:"spawners"`
	Items       []Link        `json:"items"`
	OnEnter     *Link         `json:"onEnter,omitempty"`
	Actions     []RoomAction  `json:"actions"`
	Quests      []QuestHit    `json:"quests"`
}

// Room builds the static inspector for one content room.
// Reachability is worldindex.Index.Reachability. Hidden-exit revealers use RevealTargets.
func Room(ix *worldindex.Index, id string) (RoomView, bool) {
	id = strings.TrimSpace(id)
	if ix == nil || id == "" || !ix.Has(worldindex.KindRoom, id) {
		return RoomView{}, false
	}
	room := ix.Snapshot().Rooms[id]
	if room == nil {
		return RoomView{}, false
	}
	view := RoomView{
		ID:          room.ID,
		Name:        room.Name,
		Description: room.Description,
		Area:        room.Area,
		RoomType:    room.RoomType,
		Tags:        append([]string{}, room.Tags...),
		Reach:       reachFor(ix, id),
		ExitsOut:    exitsOut(ix, room),
		ExitsIn:     exitsIn(ix, id),
		Actors:      actorsFor(ix, room),
		Spawners:    spawnersInRoom(ix, id),
		Items:       roomItems(ix, room),
		OnEnter:     link(ix, worldindex.KindScript, room.OnEnterScriptID),
		Actions:     actionsFor(ix, room),
		Quests:      roomQuests(ix, id),
	}
	if view.ID == "" {
		view.ID = id
	}
	if view.Tags == nil {
		view.Tags = []string{}
	}
	return view, true
}

func reachFor(ix *worldindex.Index, id string) ReachInfo {
	report := ix.Reachability()
	info := ReachInfo{StartRoomID: report.StartRoomID, Reachable: report.Reachable[id]}
	if info.Reachable {
		return info
	}
	for _, island := range report.Islands {
		for _, roomID := range island.RoomIDs {
			if roomID == id {
				info.Reason = island.Reason
				return info
			}
		}
	}
	if info.StartRoomID == "" {
		info.Reason = "no start room"
	}
	return info
}

func exitsOut(ix *worldindex.Index, room *rooms.Room) []ExitView {
	out := []ExitView{}
	if room == nil || room.Exits == nil {
		return out
	}
	for _, exit := range *room.Exits {
		out = append(out, exitView(ix, room.ID, exit))
	}
	return out
}

func exitView(ix *worldindex.Index, fromRoom string, exit rooms.Exit) ExitView {
	view := ExitView{
		Name:     exit.Name,
		Hidden:   exit.Hidden,
		Instance: exit.Instance || strings.EqualFold(string(exit.Type), "instance"),
	}
	if other := link(ix, worldindex.KindRoom, exit.Target); other != nil {
		view.RoomID = other.ID
		view.RoomName = other.Name
		view.Missing = other.Missing
	}
	if exit.Hidden {
		view.RevealedBy = revealers(ix, fromRoom, exit.Name)
		view.NoRevealer = len(view.RevealedBy) == 0
	}
	return view
}

func exitsIn(ix *worldindex.Index, roomID string) []ExitView {
	out := []ExitView{}
	for _, edge := range ix.Inbound(worldindex.KindRoom, roomID) {
		name, hidden, instance, ok := parseExitHow(edge.How)
		if !ok || edge.FromType != worldindex.KindRoom {
			continue
		}
		view := ExitView{Name: name, Hidden: hidden, Instance: instance}
		if from := link(ix, worldindex.KindRoom, edge.FromID); from != nil {
			view.RoomID = from.ID
			view.RoomName = from.Name
			view.Missing = from.Missing
		}
		if hidden {
			view.RevealedBy = revealers(ix, edge.FromID, name)
			view.NoRevealer = len(view.RevealedBy) == 0
		}
		out = append(out, view)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].RoomID != out[j].RoomID {
			return out[i].RoomID < out[j].RoomID
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func parseExitHow(how string) (name string, hidden, instance, ok bool) {
	how = strings.TrimSpace(how)
	if strings.HasPrefix(how, "instance ") {
		instance = true
		how = strings.TrimPrefix(how, "instance ")
	}
	switch {
	case strings.HasPrefix(how, "hidden exit "):
		return strings.TrimPrefix(how, "hidden exit "), true, instance, true
	case strings.HasPrefix(how, "exit "):
		return strings.TrimPrefix(how, "exit "), false, instance, true
	default:
		return "", false, false, false
	}
}

func revealers(ix *worldindex.Index, roomID, exitName string) []Link {
	out := []Link{}
	if ix == nil || roomID == "" || exitName == "" {
		return out
	}
	seen := map[string]bool{}
	for _, script := range ix.Snapshot().Scripts {
		if script == nil || script.Entity == nil || seen[script.ID] {
			continue
		}
		for _, edge := range ix.Outbound(worldindex.KindScript, script.ID) {
			if edge.How != "revealExit" || !strings.EqualFold(edge.ExitName, exitName) {
				continue
			}
			for _, target := range ix.RevealTargets(edge) {
				if target != roomID {
					continue
				}
				seen[script.ID] = true
				if item := link(ix, worldindex.KindScript, script.ID); item != nil {
					out = append(out, *item)
				}
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func actorsFor(ix *worldindex.Index, room *rooms.Room) []RoomActor {
	out := []RoomActor{}
	seen := map[string]bool{}
	if room.NPCs != nil {
		for _, id := range *room.NPCs {
			if id == "" || seen[id] {
				continue
			}
			seen[id] = true
			out = append(out, actor(ix, id, "resident"))
		}
	}
	for id, n := range ix.Snapshot().NPCs {
		if n == nil || n.SpawnRoomID != room.ID || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, actor(ix, id, "spawn"))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func actor(ix *worldindex.Index, id, how string) RoomActor {
	row := RoomActor{ID: id, How: how}
	n := ix.Snapshot().NPCs[id]
	if n == nil {
		row.Missing = true
		return row
	}
	row.Name = n.Name
	row.Level = n.Level
	row.Kinds = npcKinds(n)
	return row
}

func spawnersInRoom(ix *worldindex.Index, roomID string) []SpawnerView {
	out := []SpawnerView{}
	for id, spawner := range ix.Snapshot().Spawners {
		if spawner == nil || spawner.RoomID != roomID {
			continue
		}
		if spawner.ID != "" {
			id = spawner.ID
		}
		n := ix.Snapshot().NPCs[spawner.TemplateID]
		row := SpawnerView{
			ID:            id,
			Name:          spawner.Name,
			RoomID:        spawner.RoomID,
			MaxInstances:  spawner.MaxInstances,
			SpawnInterval: durationString(spawner.SpawnInterval),
		}
		if tpl := link(ix, worldindex.KindNPC, spawner.TemplateID); tpl != nil {
			row.TemplateID = tpl.ID
			row.TemplateName = tpl.Name
			row.TemplateMissing = tpl.Missing
			if row.Name == "" {
				row.Name = tpl.Name
			}
		}
		if room := link(ix, worldindex.KindRoom, spawner.RoomID); room != nil {
			row.RoomName = room.Name
			row.RoomMissing = room.Missing
		}
		row.RespawnTime, row.RespawnSource = spawnerRespawn(spawner, n)
		out = append(out, row)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func roomItems(ix *worldindex.Index, room *rooms.Room) []Link {
	out := []Link{}
	if room.Items == nil {
		return out
	}
	for _, id := range *room.Items {
		if item := link(ix, worldindex.KindItem, id); item != nil {
			out = append(out, *item)
		}
	}
	return out
}

func actionsFor(ix *worldindex.Index, room *rooms.Room) []RoomAction {
	out := []RoomAction{}
	if room.Actions == nil {
		return out
	}
	for _, action := range *room.Actions {
		out = append(out, RoomAction{
			Name:   action.Name,
			Type:   string(action.Type),
			Script: link(ix, worldindex.KindScript, action.ScriptId),
		})
	}
	return out
}

func roomQuests(ix *worldindex.Index, roomID string) []QuestHit {
	out := []QuestHit{}
	for _, quest := range ix.Snapshot().Quests {
		if quest == nil || quest.Entity == nil {
			continue
		}
		name := ix.Name(worldindex.KindQuest, quest.ID)
		if name == quest.ID {
			name = quest.Name
		}
		for _, objective := range quest.Objectives {
			if objective.Type != quests.ObjectiveVisit || strings.TrimSpace(objective.TargetID) != roomID {
				continue
			}
			out = append(out, QuestHit{
				ID:          quest.ID,
				Name:        name,
				Role:        "visit",
				ObjectiveID: objective.ID,
				Description: objective.Description,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ID != out[j].ID {
			return out[i].ID < out[j].ID
		}
		return out[i].ObjectiveID < out[j].ObjectiveID
	})
	return out
}
