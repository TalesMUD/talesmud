package npc

import (
	"fmt"
	"strings"
	"time"

	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

// Movement is authored configuration, shared by templates and spawners.
// A spawner override replaces the complete block; omitted timing uses defaults.
type Movement struct {
	Mode                 string   `bson:"mode" json:"mode" yaml:"mode"`
	Route                []string `bson:"route,omitempty" json:"route,omitempty" yaml:"route,omitempty"`
	Loop                 string   `bson:"loop,omitempty" json:"loop,omitempty" yaml:"loop,omitempty"`
	Zones                []string `bson:"zones,omitempty" json:"zones,omitempty" yaml:"zones,omitempty"`
	Rooms                []string `bson:"rooms,omitempty" json:"rooms,omitempty" yaml:"rooms,omitempty"`
	IntervalSeconds      int      `bson:"intervalSeconds,omitempty" json:"intervalSeconds,omitempty" yaml:"intervalSeconds,omitempty"`
	JitterSeconds        int      `bson:"jitterSeconds,omitempty" json:"jitterSeconds,omitempty" yaml:"jitterSeconds,omitempty"`
	PauseOnEngageSeconds int      `bson:"pauseOnEngageSeconds,omitempty" json:"pauseOnEngageSeconds,omitempty" yaml:"pauseOnEngageSeconds,omitempty"`
	LeaveMessage         string   `bson:"leaveMessage,omitempty" json:"leaveMessage,omitempty" yaml:"leaveMessage,omitempty"`
	ArriveMessage        string   `bson:"arriveMessage,omitempty" json:"arriveMessage,omitempty" yaml:"arriveMessage,omitempty"`
}

func (m *Movement) Clone() *Movement {
	if m == nil {
		return nil
	}
	c := *m
	c.Route = append([]string(nil), m.Route...)
	c.Zones = append([]string(nil), m.Zones...)
	c.Rooms = append([]string(nil), m.Rooms...)
	return &c
}

func (m *Movement) Interval() time.Duration {
	if m == nil || m.IntervalSeconds == 0 {
		return 90 * time.Second
	}
	return time.Duration(m.IntervalSeconds) * time.Second
}
func (m *Movement) Pause() time.Duration {
	if m == nil || m.PauseOnEngageSeconds == 0 {
		return 120 * time.Second
	}
	return time.Duration(m.PauseOnEngageSeconds) * time.Second
}
func (m *Movement) LoopMode() string {
	if m.Loop == "" {
		return "loop"
	}
	return m.Loop
}

// EffectiveMovement keeps the original fields useful regardless of FSM state.
func (n *NPC) EffectiveMovement() *Movement {
	if n.Movement != nil {
		return n.Movement
	}
	if len(n.PatrolPath) > 0 {
		return &Movement{Mode: "route", Route: n.PatrolPath}
	}
	if n.WanderRadius > 0 {
		return &Movement{Mode: "roam"}
	}
	return nil
}

func ContainsID(ids []string, id string) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

// MovementRoomAllowed excludes authored instances, runtime copies and room opt-outs.
// Aggro is still governed by the game's normal noAggro check on arrival.
func MovementRoomAllowed(m *Movement, room *rooms.Room) bool {
	if room == nil || room.Entity == nil || strings.Contains(room.ID, "~") || room.IsInstanceTemplate() {
		return false
	}
	for _, tag := range []string{"noNPC", "no-NPC", "noWander", "no-wander"} {
		if room.HasTag(tag) {
			return false
		}
	}
	return (len(m.Zones) == 0 || ContainsID(m.Zones, room.Area)) && (len(m.Rooms) == 0 || ContainsID(m.Rooms, room.ID))
}

// MovementExitAllowed also accepts the legacy params.locked flag. Teleports are
// not walking exits, even when their target is in the same area.
func MovementExitAllowed(e rooms.Exit) bool {
	locked := e.Locked
	if value, ok := e.Params["locked"]; ok {
		locked = locked || value == true || value == "true"
	}
	return e.Target != "" && !e.Hidden && !e.Instance && !locked && e.Type != rooms.RoomExitTypeTeleport && e.Type != "instance"
}

func MovementExit(from *rooms.Room, to string) (rooms.Exit, bool) {
	if from != nil && from.Exits != nil {
		for _, e := range *from.Exits {
			if e.Target == to && MovementExitAllowed(e) {
				return e, true
			}
		}
	}
	return rooms.Exit{}, false
}

// ValidateMovement is used both by content health and runtime loading. Routes
// must be walkable in order, including closure for loops and reverse for pingpong.
func ValidateMovement(m *Movement, world map[string]*rooms.Room) []string {
	if m == nil {
		return nil
	}
	var issues []string
	if m.Mode != "route" && m.Mode != "roam" {
		issues = append(issues, "mode must be route or roam")
	}
	if m.LoopMode() != "loop" && m.LoopMode() != "pingpong" {
		issues = append(issues, "loop must be loop or pingpong")
	}
	for label, value := range map[string]int{"intervalSeconds": m.IntervalSeconds, "jitterSeconds": m.JitterSeconds, "pauseOnEngageSeconds": m.PauseOnEngageSeconds} {
		if value < 0 || value > 86400 {
			issues = append(issues, label+" must be between 0 and 86400")
		}
	}
	for _, zone := range m.Zones {
		found := false
		for _, room := range world {
			if room.Area == zone {
				found = true
				break
			}
		}
		if !found {
			issues = append(issues, "unknown zone "+zone)
		}
	}
	for _, id := range m.Rooms {
		if !MovementRoomAllowed(m, world[id]) {
			issues = append(issues, "whitelist room missing or forbidden: "+id)
		}
	}
	if m.Mode == "roam" && len(m.Route) > 0 {
		issues = append(issues, "roam cannot specify route")
	}
	if m.Mode == "route" {
		if len(m.Route) < 2 {
			issues = append(issues, "route requires at least two rooms")
		}
		seen := map[string]bool{}
		for _, id := range m.Route {
			if seen[id] {
				issues = append(issues, "duplicate route room "+id)
			}
			seen[id] = true
			if !MovementRoomAllowed(m, world[id]) {
				issues = append(issues, "route room missing or forbidden: "+id)
			}
		}
		check := func(a, b string) {
			if _, ok := MovementExit(world[a], b); !ok {
				issues = append(issues, fmt.Sprintf("route step %s -> %s has no allowed exit", a, b))
			}
		}
		for i := 1; i < len(m.Route); i++ {
			check(m.Route[i-1], m.Route[i])
			if m.LoopMode() == "pingpong" {
				check(m.Route[i], m.Route[i-1])
			}
		}
		if len(m.Route) > 1 && m.LoopMode() == "loop" {
			check(m.Route[len(m.Route)-1], m.Route[0])
		}
	}
	return issues
}

// MovementStartRoom resets an invalid stored location to an allowed spawn or
// first route room. An invalid roaming spawn is refused rather than guessed.
func MovementStartRoom(n *NPC, world map[string]*rooms.Room) string {
	m := n.EffectiveMovement()
	if m == nil {
		return n.CurrentRoomID
	}
	allowed := func(id string) bool {
		if !MovementRoomAllowed(m, world[id]) {
			return false
		}
		if m.Mode == "route" {
			return ContainsID(m.Route, id)
		}
		if n.Movement == nil && n.WanderRadius > 0 {
			return id == n.SpawnRoomID
		}
		return true
	}
	if allowed(n.CurrentRoomID) {
		return n.CurrentRoomID
	}
	if allowed(n.SpawnRoomID) {
		return n.SpawnRoomID
	}
	if m.Mode == "route" && len(m.Route) > 0 && allowed(m.Route[0]) {
		return m.Route[0]
	}
	return ""
}
