// Package guide chooses one next-step hint from quest progress and the room graph.
// It does not know any particular world pack.
package guide

import (
	"sort"
	"strings"
)

// Hint is one line a new character can act on. Empty Text means show nothing.
type Hint struct {
	Text      string
	QuestName string
	Step      string
	Command   string
	Label     string
}

// World is the slice of game state the hint is allowed to see.
type World struct {
	Dismissed     bool
	CurrentRoomID string
	// Revealed lists hidden exit names a character has unlocked, keyed by room id.
	Revealed map[string][]string
	Quests   []Quest
	Rooms    []Room
	// Places maps a catalog id (NPC template, placed NPC) to room ids.
	Places map[string][]string
	Here   []Actor
}

// Quest is a definition plus this character's progress.
// Status "" means it is not on the character yet.
type Quest struct {
	ID             string
	Name           string
	Level          int32
	Guide          bool
	Status         string
	SourceType     string
	SourceNPCID    string
	SourceNPCName  string
	TurnInAnywhere bool
	TurnInNPCID    string
	TurnInNPCName  string
	Objectives     []Objective
}

// Objective is one step and how far the character is.
type Objective struct {
	Type             string
	Description      string
	TargetID         string
	TargetName       string
	Amount           int32
	Current          int32
	Completed        bool
	Order            int32
	DeliverToNPCID   string
	DeliverToNPCName string
}

// Room is an id plus the exits a character might walk.
type Room struct {
	ID    string
	Exits []Exit
}

// Exit is one way out of a room.
type Exit struct {
	Name   string
	Target string
	Hidden bool
	Locked bool
}

// Actor is a living NPC in the current room.
type Actor struct {
	TemplateID string
	Name       string
}

// Build returns the single next step, or nil when the hint must stay hidden.
func Build(w World) *Hint {
	if w.Dismissed || anyCompleted(w.Quests) {
		return nil
	}
	pool := poolOf(w.Quests)
	if len(pool) == 0 {
		return nil
	}
	order := rank(pool)
	for _, i := range order {
		if h := step(w, pool[i]); h != nil && strings.TrimSpace(h.Text) != "" {
			return h
		}
	}
	return nil
}

func anyCompleted(quests []Quest) bool {
	for _, q := range quests {
		if strings.EqualFold(q.Status, "completed") {
			return true
		}
	}
	return false
}

func poolOf(quests []Quest) []Quest {
	guided := false
	for _, q := range quests {
		if q.Guide && !closed(q) {
			guided = true
			break
		}
	}
	var pool []Quest
	for _, q := range quests {
		if closed(q) {
			continue
		}
		if guided {
			if q.Guide {
				pool = append(pool, q)
			}
			continue
		}
		if strings.EqualFold(q.Status, "active") {
			pool = append(pool, q)
		}
	}
	return pool
}

func closed(q Quest) bool {
	switch strings.ToLower(strings.TrimSpace(q.Status)) {
	case "completed", "failed", "abandoned":
		return true
	default:
		return false
	}
}

func accepted(q Quest) bool {
	return strings.EqualFold(q.Status, "active")
}

func rank(pool []Quest) []int {
	idx := make([]int, len(pool))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(i, j int) bool {
		a, b := pool[idx[i]], pool[idx[j]]
		if ta, tb := tier(a), tier(b); ta != tb {
			return ta < tb
		}
		if a.Level != b.Level {
			return a.Level < b.Level
		}
		return a.ID < b.ID
	})
	return idx
}

func tier(q Quest) int {
	if ready(q) {
		return 0
	}
	if accepted(q) && hasOpenKill(q) {
		return 1
	}
	if !accepted(q) {
		return 2
	}
	return 3
}

func ready(q Quest) bool {
	if !accepted(q) || len(q.Objectives) == 0 {
		return false
	}
	for _, o := range q.Objectives {
		if !o.done() {
			return false
		}
	}
	return true
}

func hasOpenKill(q Quest) bool {
	for _, o := range q.Objectives {
		if !o.done() && strings.EqualFold(o.Type, "kill") {
			return true
		}
	}
	return false
}

func (o Objective) done() bool {
	if o.Completed {
		return true
	}
	need := o.Amount
	if need < 1 {
		need = 1
	}
	return o.Current >= need && o.Current > 0
}

func step(w World, q Quest) *Hint {
	if !accepted(q) && !autoSource(q) {
		return offer(w, q)
	}
	if ready(q) {
		return turnIn(w, q)
	}
	obj := firstOpen(q)
	if obj == nil {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(obj.Type)) {
	case "kill":
		return killStep(w, q, *obj)
	case "visit":
		return visitStep(w, q, *obj)
	case "talk":
		return talkStep(w, q, obj.TargetID, obj.TargetName, obj.Description, "talk")
	case "deliver":
		name := obj.DeliverToNPCName
		if name == "" {
			name = obj.TargetName
		}
		id := obj.DeliverToNPCID
		if id == "" {
			id = obj.TargetID
		}
		return talkStep(w, q, id, name, obj.Description, "head")
	default:
		return textOnly(q, obj.Description, "head")
	}
}

func autoSource(q Quest) bool {
	switch strings.ToLower(strings.TrimSpace(q.SourceType)) {
	case "auto", "script", "":
		return q.SourceNPCID == ""
	default:
		return false
	}
}

func offer(w World, q Quest) *Hint {
	name := displayName(q.SourceNPCName, "the quest giver")
	line := "Talk to " + name + " to accept " + q.Name + "."
	if a := actorHere(w, q.SourceNPCID, name); a != nil {
		return &Hint{Text: "Talk to " + displayName(a.Name, name) + ".", QuestName: q.Name, Step: "accept", Command: "talk " + displayName(a.Name, name), Label: "Talk"}
	}
	if room := nearest(w, q.SourceNPCID); room != "" && room != w.CurrentRoomID {
		if exit, ok := firstStep(w, room); ok {
			return &Hint{Text: headPhrase(exit) + " " + line, QuestName: q.Name, Step: "accept", Command: exit, Label: labelExit(exit)}
		}
	}
	return textOnly(q, line, "accept")
}

func turnIn(w World, q Quest) *Hint {
	if q.TurnInAnywhere || q.TurnInNPCID == "" {
		return &Hint{
			Text:      "Turn in " + q.Name + ".",
			QuestName: q.Name,
			Step:      "turnin",
			Command:   "complete " + q.Name,
			Label:     "Turn in",
		}
	}
	name := displayName(q.TurnInNPCName, "the quest giver")
	if a := actorHere(w, q.TurnInNPCID, name); a != nil {
		who := displayName(a.Name, name)
		return &Hint{Text: "Talk to " + who + " to turn in " + q.Name + ".", QuestName: q.Name, Step: "turnin", Command: "talk " + who, Label: "Talk"}
	}
	line := "Talk to " + name + " to turn in " + q.Name + "."
	if room := nearest(w, q.TurnInNPCID); room != "" && room != w.CurrentRoomID {
		if exit, ok := firstStep(w, room); ok {
			return &Hint{Text: headPhrase(exit) + " " + line, QuestName: q.Name, Step: "turnin", Command: exit, Label: labelExit(exit)}
		}
	}
	return textOnly(q, line, "turnin")
}

func killStep(w World, q Quest, obj Objective) *Hint {
	name := displayName(obj.TargetName, "the enemy")
	if a := actorHere(w, obj.TargetID, name); a != nil {
		who := displayName(obj.TargetName, a.Name)
		return &Hint{Text: "Tap the " + who + " to fight.", QuestName: q.Name, Step: "fight", Command: "attack " + who, Label: "Fight"}
	}
	blurb := obj.Description
	if strings.TrimSpace(blurb) == "" {
		blurb = "Fight the " + name + "."
	}
	if atPlace(w, obj.TargetID) {
		return textOnly(q, blurb, "fight")
	}
	if room := nearest(w, obj.TargetID); room != "" && room != w.CurrentRoomID {
		if exit, ok := firstStep(w, room); ok {
			return &Hint{Text: headPhrase(exit) + " " + blurb, QuestName: q.Name, Step: "head", Command: exit, Label: labelExit(exit)}
		}
	}
	return textOnly(q, blurb, "head")
}

func visitStep(w World, q Quest, obj Objective) *Hint {
	blurb := obj.Description
	if strings.TrimSpace(blurb) == "" {
		blurb = "Keep going."
	}
	if obj.TargetID == "" || obj.TargetID == w.CurrentRoomID {
		return textOnly(q, blurb, "head")
	}
	if exit, ok := firstStep(w, obj.TargetID); ok {
		return &Hint{Text: headPhrase(exit) + " " + blurb, QuestName: q.Name, Step: "head", Command: exit, Label: labelExit(exit)}
	}
	return textOnly(q, blurb, "head")
}

func talkStep(w World, q Quest, npcID, npcName, blurb, fallbackStep string) *Hint {
	name := displayName(npcName, "them")
	if strings.TrimSpace(blurb) == "" {
		blurb = "Talk to " + name + "."
	}
	if a := actorHere(w, npcID, name); a != nil {
		who := displayName(a.Name, name)
		return &Hint{Text: "Talk to " + who + ".", QuestName: q.Name, Step: "talk", Command: "talk " + who, Label: "Talk"}
	}
	if atPlace(w, npcID) {
		return textOnly(q, blurb, "talk")
	}
	if room := nearest(w, npcID); room != "" && room != w.CurrentRoomID {
		if exit, ok := firstStep(w, room); ok {
			return &Hint{Text: headPhrase(exit) + " " + blurb, QuestName: q.Name, Step: "head", Command: exit, Label: labelExit(exit)}
		}
	}
	return textOnly(q, blurb, fallbackStep)
}

func textOnly(q Quest, text, step string) *Hint {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	return &Hint{Text: text, QuestName: q.Name, Step: step}
}

func firstOpen(q Quest) *Objective {
	idx := make([]int, len(q.Objectives))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(i, j int) bool {
		return q.Objectives[idx[i]].Order < q.Objectives[idx[j]].Order
	})
	for _, i := range idx {
		if !q.Objectives[i].done() {
			obj := q.Objectives[i]
			return &obj
		}
	}
	return nil
}

func actorHere(w World, templateID, name string) *Actor {
	name = strings.ToLower(strings.TrimSpace(name))
	for i := range w.Here {
		a := &w.Here[i]
		if templateID != "" && a.TemplateID == templateID {
			return a
		}
		an := strings.ToLower(strings.TrimSpace(a.Name))
		if name != "" && an != "" && (strings.Contains(an, name) || strings.Contains(name, an)) {
			return a
		}
	}
	return nil
}

func atPlace(w World, id string) bool {
	for _, room := range w.Places[id] {
		if room == w.CurrentRoomID {
			return true
		}
	}
	return false
}

func nearest(w World, id string) string {
	rooms := w.Places[id]
	if len(rooms) == 0 {
		return ""
	}
	best := ""
	bestDist := int(^uint(0) >> 1)
	for _, room := range rooms {
		if room == w.CurrentRoomID {
			return room
		}
		d := distance(w, room)
		if d >= 0 && d < bestDist {
			best = room
			bestDist = d
		}
	}
	return best
}

func distance(w World, dest string) int {
	if dest == "" {
		return -1
	}
	if dest == w.CurrentRoomID {
		return 0
	}
	type node struct {
		room string
		dist int
	}
	q := []node{{w.CurrentRoomID, 0}}
	seen := map[string]bool{w.CurrentRoomID: true}
	index := roomIndex(w)
	for len(q) > 0 {
		cur := q[0]
		q = q[1:]
		for _, ex := range index[cur.room] {
			if !walkable(w, cur.room, ex) || seen[ex.Target] {
				continue
			}
			if ex.Target == dest {
				return cur.dist + 1
			}
			seen[ex.Target] = true
			q = append(q, node{ex.Target, cur.dist + 1})
		}
	}
	return -1
}

// firstStep returns the exit name to type from the current room.
func firstStep(w World, dest string) (string, bool) {
	if dest == "" || dest == w.CurrentRoomID {
		return "", false
	}
	type node struct {
		room string
		via  string
	}
	q := []node{{w.CurrentRoomID, ""}}
	seen := map[string]bool{w.CurrentRoomID: true}
	index := roomIndex(w)
	for len(q) > 0 {
		cur := q[0]
		q = q[1:]
		for _, ex := range index[cur.room] {
			if !walkable(w, cur.room, ex) || seen[ex.Target] {
				continue
			}
			via := cur.via
			if via == "" {
				via = ex.Name
			}
			if ex.Target == dest {
				return via, via != ""
			}
			seen[ex.Target] = true
			q = append(q, node{ex.Target, via})
		}
	}
	return "", false
}

func roomIndex(w World) map[string][]Exit {
	out := make(map[string][]Exit, len(w.Rooms))
	for _, room := range w.Rooms {
		out[room.ID] = room.Exits
	}
	return out
}

func walkable(w World, roomID string, ex Exit) bool {
	if ex.Locked || strings.TrimSpace(ex.Name) == "" || strings.TrimSpace(ex.Target) == "" {
		return false
	}
	if !ex.Hidden {
		return true
	}
	for _, name := range w.Revealed[roomID] {
		if strings.EqualFold(name, ex.Name) {
			return true
		}
	}
	return false
}

func headPhrase(exitName string) string {
	n := strings.ToLower(strings.TrimSpace(exitName))
	switch n {
	case "north", "south", "east", "west", "up", "down",
		"northeast", "northwest", "southeast", "southwest",
		"n", "s", "e", "w", "u", "d":
		return "Head " + n + "."
	default:
		if n == "" {
			return "Move on."
		}
		return "Take " + strings.TrimSpace(exitName) + "."
	}
}

func labelExit(exitName string) string {
	n := strings.TrimSpace(exitName)
	if n == "" {
		return "Go"
	}
	return strings.ToUpper(n[:1]) + n[1:]
}

func displayName(name, fallback string) string {
	name = strings.TrimSpace(name)
	if name != "" && !strings.EqualFold(name, "the quest giver") && !strings.EqualFold(name, "them") && !strings.EqualFold(name, "the enemy") {
		return name
	}
	if strings.TrimSpace(fallback) != "" {
		return strings.TrimSpace(fallback)
	}
	return name
}
