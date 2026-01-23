package def

import (
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/service"
)

// NPCInstanceCtrl provides access to the NPC instance manager for scripts
type NPCInstanceCtrl interface {
	// GetInstance returns an NPC instance by ID
	GetInstance(id string) *npc.NPC
	// GetInstancesInRoom returns all alive NPC instances in a room
	GetInstancesInRoom(roomID string) []*npc.NPC
	// GetAllInstances returns all NPC instances
	GetAllInstances() []*npc.NPC
	// SpawnInstanceDirect spawns an NPC from a template (not via spawner)
	SpawnInstanceDirect(templateID, roomID string) (*npc.NPC, error)
	// KillInstance marks an NPC instance as dead
	KillInstance(id string) bool
	// DamageInstance applies damage and returns true if the NPC died
	DamageInstance(id string, amount int32) bool
	// HealInstance restores HP to an instance
	HealInstance(id string, amount int32) bool
	// MoveInstance moves an instance to a new room
	MoveInstance(id, roomID string) bool
	// UpdateInstance updates an instance using a callback
	UpdateInstance(id string, updater func(*npc.NPC)) bool
	// FindInstanceByNameInRoom finds an instance by name in a room
	FindInstanceByNameInRoom(roomID, name string) *npc.NPC
}

// GameCtrl def
// interface for commands package to communicate back to game instance
type GameCtrl interface {

	// Used to pass messages as events inside the mud server, e.g. translate a command into other user messages etc.
	OnMessageReceived() chan interface{}
	// used to send replies/messages to users, origin or rooms, or global
	SendMessage() chan interface{}
	GetFacade() service.Facade
	// GetNPCInstanceManager returns the NPC instance controller
	GetNPCInstanceManager() NPCInstanceCtrl
}
