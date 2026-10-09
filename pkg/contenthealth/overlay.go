package contenthealth

import (
	"sort"
	"strings"

	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/worldindex"
)

// OverlayPlayer is one online character the map can place.
// MapOverlayView keeps the name only when the caller is an admin.
type OverlayPlayer struct {
	ID     string
	Name   string
	RoomID string
}

// OverlayLive is a read of the running world.
// Available is false when that read did not happen.
// CloneIDs are live instance room ids (they contain "~").
type OverlayLive struct {
	Available bool
	Players   []OverlayPlayer
	CloneIDs  []string
}

// OverlaySpawner is one content spawner drawn on its room.
type OverlaySpawner struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	TemplateID  string `json:"templateId,omitempty"`
	RespawnTime string `json:"respawnTime,omitempty"`
}

// OverlayRoom is one map room. Empty signals are omitted.
// PlayerNames is set only for an admin. A creator still gets Players.
type OverlayRoom struct {
	ID          string           `json:"id"`
	LevelMin    int32            `json:"levelMin,omitempty"`
	LevelMax    int32            `json:"levelMax,omitempty"`
	LevelBand   string           `json:"levelBand,omitempty"`
	LevelSource string           `json:"levelSource,omitempty"`
	Aggro       int              `json:"aggro,omitempty"`
	Spawners    []OverlaySpawner `json:"spawners,omitempty"`
	Players     int              `json:"players,omitempty"`
	PlayerNames []string         `json:"playerNames,omitempty"`
	Quests      []string         `json:"quests,omitempty"`
	MissingArt  bool             `json:"missingArt,omitempty"`
	Copies      int              `json:"copies,omitempty"`
}

// MapOverlays is GET /api/world/overlays.
// Live is false when the game read was skipped. Static layers still return.
type MapOverlays struct {
	Admin bool          `json:"admin"`
	Live  bool          `json:"live"`
	Rooms []OverlayRoom `json:"rooms"`
}

type levelSpan struct {
	min int32
	max int32
	ok  bool
}

// MapOverlayView builds the world-map layers from the content snapshot and a live read.
// includePlayers keeps character names. Creators pass false and still receive counts.
// A room is omitted when every layer is empty. A live clone with players is included
// even though the content snapshot leaves "~" rooms out.
func MapOverlayView(world World, live OverlayLive, includePlayers bool) MapOverlays {
	snap := world.Snapshot()
	ix := worldindex.Build(snap)
	placed, bySpawner := overlayPlacement(snap)

	rows := map[string]*OverlayRoom{}
	own := map[string]levelSpan{}
	for id, room := range snap.Rooms {
		if id == "" || room == nil {
			continue
		}
		row := &OverlayRoom{ID: id}
		if missingArt(room) {
			row.MissingArt = true
		}
		span := levelSpan{}
		seen := map[string]bool{}
		for _, n := range placed[id] {
			addNPCLevel(&span, seen, n)
		}
		for _, spawner := range bySpawner[id] {
			addNPCLevel(&span, seen, snap.NPCs[spawner.TemplateID])
		}
		if span.ok {
			own[id] = span
			applyLevel(row, span, "room")
		}
		row.Aggro = aggroCount(placed[id], bySpawner[id], snap.NPCs)
		row.Spawners = spawnerRows(bySpawner[id], snap.NPCs)
		rows[id] = row
	}

	zones := map[string]levelSpan{}
	for id, span := range own {
		room := snap.Rooms[id]
		if room == nil {
			continue
		}
		area := strings.TrimSpace(room.Area)
		if area == "" {
			continue
		}
		zone := zones[area]
		addSpan(&zone, span)
		zones[area] = zone
	}
	for id, row := range rows {
		if row.LevelBand != "" {
			continue
		}
		room := snap.Rooms[id]
		if room == nil {
			continue
		}
		area := strings.TrimSpace(room.Area)
		if area == "" {
			continue
		}
		span := zones[area]
		if !span.ok {
			continue
		}
		applyLevel(row, span, "zone")
	}

	for roomID, quests := range questRooms(ix) {
		row := ensureRow(rows, roomID)
		row.Quests = questList(quests)
	}

	playerSeen := map[string]map[string]bool{}
	for _, player := range live.Players {
		roomID := strings.TrimSpace(player.RoomID)
		if roomID == "" {
			continue
		}
		id := strings.TrimSpace(player.ID)
		if id != "" {
			if playerSeen[roomID] == nil {
				playerSeen[roomID] = map[string]bool{}
			}
			if playerSeen[roomID][id] {
				continue
			}
			playerSeen[roomID][id] = true
		}
		row := ensureRow(rows, roomID)
		row.Players++
		if !includePlayers {
			continue
		}
		label := strings.TrimSpace(player.Name)
		if label == "" {
			label = id
		}
		if label != "" {
			row.PlayerNames = append(row.PlayerNames, label)
		}
	}
	for _, row := range rows {
		if len(row.PlayerNames) > 0 {
			sort.Strings(row.PlayerNames)
		}
	}

	clones := map[string]bool{}
	for _, id := range live.CloneIDs {
		if clones[id] {
			continue
		}
		clones[id] = true
		template, ok := templateOfClone(id)
		if !ok || snap.Rooms[template] == nil {
			continue
		}
		ensureRow(rows, template).Copies++
	}

	out := make([]OverlayRoom, 0, len(rows))
	for _, row := range rows {
		if row == nil || overlayEmpty(*row) {
			continue
		}
		out = append(out, *row)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	if out == nil {
		out = []OverlayRoom{}
	}
	return MapOverlays{Admin: includePlayers, Live: live.Available, Rooms: out}
}

func overlayEmpty(row OverlayRoom) bool {
	return row.LevelBand == "" && row.Aggro == 0 && len(row.Spawners) == 0 && row.Players == 0 && len(row.Quests) == 0 && !row.MissingArt && row.Copies == 0
}

func ensureRow(rows map[string]*OverlayRoom, id string) *OverlayRoom {
	if row := rows[id]; row != nil {
		return row
	}
	row := &OverlayRoom{ID: id}
	rows[id] = row
	return row
}

func overlayPlacement(snap worldindex.Snapshot) (map[string][]*npc.NPC, map[string][]*npc.NPCSpawner) {
	placed := map[string][]*npc.NPC{}
	seen := map[string]map[string]bool{}
	add := func(roomID string, n *npc.NPC) {
		if roomID == "" || n == nil || n.ID == "" || snap.Rooms[roomID] == nil {
			return
		}
		if seen[roomID] == nil {
			seen[roomID] = map[string]bool{}
		}
		if seen[roomID][n.ID] {
			return
		}
		seen[roomID][n.ID] = true
		placed[roomID] = append(placed[roomID], n)
	}
	for id, room := range snap.Rooms {
		if room == nil || room.NPCs == nil {
			continue
		}
		for _, npcID := range *room.NPCs {
			add(id, snap.NPCs[npcID])
		}
	}
	for _, n := range snap.NPCs {
		if n == nil {
			continue
		}
		add(n.SpawnRoomID, n)
	}
	bySpawner := map[string][]*npc.NPCSpawner{}
	for _, spawner := range snap.Spawners {
		if spawner == nil || spawner.ID == "" || snap.Rooms[spawner.RoomID] == nil {
			continue
		}
		bySpawner[spawner.RoomID] = append(bySpawner[spawner.RoomID], spawner)
	}
	for _, list := range bySpawner {
		sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	}
	return placed, bySpawner
}

func addNPCLevel(span *levelSpan, seen map[string]bool, n *npc.NPC) {
	if n == nil || n.ID == "" || seen[n.ID] {
		return
	}
	seen[n.ID] = true
	if !span.ok {
		span.min, span.max, span.ok = n.Level, n.Level, true
		return
	}
	if n.Level < span.min {
		span.min = n.Level
	}
	if n.Level > span.max {
		span.max = n.Level
	}
}

func addSpan(dst *levelSpan, src levelSpan) {
	if !src.ok {
		return
	}
	if !dst.ok {
		*dst = src
		return
	}
	if src.min < dst.min {
		dst.min = src.min
	}
	if src.max > dst.max {
		dst.max = src.max
	}
}

func applyLevel(row *OverlayRoom, span levelSpan, source string) {
	row.LevelMin = span.min
	row.LevelMax = span.max
	row.LevelBand = levelBand(span.min, span.max)
	row.LevelSource = source
}

// levelBand colors the midpoint of the room or zone range.
// 1 is 0–4, 2 is 5–9, 3 is 10–14, 4 is 15–19, 5 is 20–29, 6 is 30 and above.
func levelBand(min, max int32) string {
	mid := min + (max-min)/2
	switch {
	case mid < 5:
		return "1"
	case mid < 10:
		return "2"
	case mid < 15:
		return "3"
	case mid < 20:
		return "4"
	case mid < 30:
		return "5"
	default:
		return "6"
	}
}

func aggroCount(placed []*npc.NPC, spawners []*npc.NPCSpawner, templates map[string]*npc.NPC) int {
	spawned := map[string]bool{}
	for _, spawner := range spawners {
		if spawner != nil && spawner.TemplateID != "" {
			spawned[spawner.TemplateID] = true
		}
	}
	total := 0
	seen := map[string]bool{}
	for _, n := range placed {
		if n == nil || n.ID == "" || seen[n.ID] || spawned[n.ID] {
			continue
		}
		seen[n.ID] = true
		if n.EnemyTrait != nil && n.EnemyTrait.AggroOnSight {
			total++
		}
	}
	for _, spawner := range spawners {
		if spawner == nil {
			continue
		}
		n := templates[spawner.TemplateID]
		if n == nil || n.EnemyTrait == nil || !n.EnemyTrait.AggroOnSight {
			continue
		}
		count := spawner.MaxInstances
		if count <= 0 {
			count = spawner.InitialCount
		}
		if count <= 0 {
			count = 1
		}
		total += count
	}
	return total
}

func spawnerRows(spawners []*npc.NPCSpawner, templates map[string]*npc.NPC) []OverlaySpawner {
	if len(spawners) == 0 {
		return nil
	}
	out := make([]OverlaySpawner, 0, len(spawners))
	for _, spawner := range spawners {
		if spawner == nil || spawner.ID == "" {
			continue
		}
		n := templates[spawner.TemplateID]
		name := strings.TrimSpace(spawner.Name)
		if name == "" && n != nil {
			name = n.Name
		}
		row := OverlaySpawner{
			ID:          spawner.ID,
			Name:        name,
			TemplateID:  spawner.TemplateID,
			RespawnTime: respawnLabel(spawner, n),
		}
		out = append(out, row)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// respawnLabel matches the inspector: override, then delay, then the template.
func respawnLabel(spawner *npc.NPCSpawner, n *npc.NPC) string {
	if spawner == nil {
		return ""
	}
	if spawner.RespawnTimeOverride != nil && *spawner.RespawnTimeOverride > 0 {
		return spawner.RespawnTimeOverride.String()
	}
	if spawner.RespawnDelay > 0 {
		return spawner.RespawnDelay.String()
	}
	if n != nil && n.RespawnTime > 0 {
		return n.RespawnTime.String()
	}
	return ""
}

func questRooms(ix *worldindex.Index) map[string]map[string]bool {
	out := map[string]map[string]bool{}
	if ix == nil {
		return out
	}
	snap := ix.Snapshot()
	add := func(roomID, questID string) {
		if roomID == "" || questID == "" || snap.Rooms[roomID] == nil {
			return
		}
		if out[roomID] == nil {
			out[roomID] = map[string]bool{}
		}
		out[roomID][questID] = true
	}
	for id := range snap.Quests {
		for _, edge := range ix.Outbound(worldindex.KindQuest, id) {
			switch edge.How {
			case "quest objective visit":
				if edge.ToType == worldindex.KindRoom {
					add(edge.ToID, id)
				}
			case "quest objective kill", "quest objective talk", "quest objective deliver":
				if edge.ToType != worldindex.KindNPC {
					continue
				}
				for _, roomID := range npcRooms(ix, edge.ToID) {
					add(roomID, id)
				}
			}
		}
	}
	return out
}

func npcRooms(ix *worldindex.Index, npcID string) []string {
	if ix == nil || npcID == "" {
		return nil
	}
	snap := ix.Snapshot()
	seen := map[string]bool{}
	var out []string
	add := func(id string) {
		if id == "" || seen[id] || snap.Rooms[id] == nil {
			return
		}
		seen[id] = true
		out = append(out, id)
	}
	for _, edge := range ix.Outbound(worldindex.KindNPC, npcID) {
		if edge.ToType == worldindex.KindRoom && (edge.How == "spawn room" || edge.How == "current room") {
			add(edge.ToID)
		}
	}
	for _, edge := range ix.Inbound(worldindex.KindNPC, npcID) {
		if edge.FromType == worldindex.KindRoom && edge.How == "room resident" {
			add(edge.FromID)
		}
		if edge.FromType != worldindex.KindSpawner || edge.How != "spawner template" {
			continue
		}
		for _, roomEdge := range ix.Outbound(worldindex.KindSpawner, edge.FromID) {
			if roomEdge.ToType == worldindex.KindRoom && roomEdge.How == "spawner room" {
				add(roomEdge.ToID)
			}
		}
	}
	return out
}

func questList(set map[string]bool) []string {
	if len(set) == 0 {
		return nil
	}
	out := make([]string, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

func missingArt(room *rooms.Room) bool {
	if room == nil || room.Meta == nil {
		return true
	}
	return strings.TrimSpace(room.Meta.Background) == ""
}

func templateOfClone(id string) (string, bool) {
	i := strings.LastIndex(id, "~")
	if i <= 0 || i >= len(id)-1 {
		return "", false
	}
	return id[:i], true
}
