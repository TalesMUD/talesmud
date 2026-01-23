package npc

import (
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
)

// NPCSpawner defines a spawn point that automatically creates NPC instances from a template
type NPCSpawner struct {
	*entities.Entity `bson:",inline"`

	// Configuration
	// Name is an optional human-readable name for the spawner (admin/debug use)
	Name string `json:"name,omitempty"`
	// TemplateID references the NPC template to spawn instances from
	TemplateID string `json:"templateId"`
	// RoomID is where spawned instances will appear
	RoomID string `json:"roomId"`
	// MaxInstances is the maximum number of alive instances at once
	MaxInstances int `json:"maxInstances"`
	// SpawnInterval is the time between spawn attempts when under max
	SpawnInterval time.Duration `json:"spawnInterval"`
	// InitialCount is how many instances to spawn when the world loads
	InitialCount int `json:"initialCount"`

	// Optional Overrides
	// RespawnTimeOverride overrides the template's respawn time if set
	RespawnTimeOverride *time.Duration `json:"respawnTimeOverride,omitempty"`

	// Metadata
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated,omitempty"`
}

// NewNPCSpawner creates a new spawner with a generated ID
func NewNPCSpawner(templateID, roomID string, maxInstances int, spawnInterval time.Duration) *NPCSpawner {
	return &NPCSpawner{
		Entity:        entities.NewEntity(),
		TemplateID:    templateID,
		RoomID:        roomID,
		MaxInstances:  maxInstances,
		SpawnInterval: spawnInterval,
		InitialCount:  1,
		Created:       time.Now(),
	}
}
