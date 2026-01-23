package npc

import (
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/traits"
)

// NPC represents a Non-Player Character in the game
type NPC struct {
	*entities.Entity   `bson:",inline"`
	traits.BelongsUser `bson:",inline"` // optionally, e.g. a bot could belong to a user
	traits.CurrentRoom `bson:",inline"`

	Name        string           `bson:"name" json:"name"`
	Description string           `bson:"description" json:"description"`
	Race        characters.Race  `bson:"race,omitempty" json:"race,omitempty"`
	Class       characters.Class `bson:"class,omitempty" json:"class,omitempty"`

	CurrentHitPoints int32 `bson:"currentHitPoints" json:"currentHitPoints"`
	MaxHitPoints     int32 `bson:"maxHitPoints" json:"maxHitPoints"`
	Level            int32 `bson:"level" json:"level"`

	// Template System
	// IsTemplate indicates this NPC is a blueprint for spawning multiple instances
	// When false, this NPC is a unique singleton that cannot be spawned via spawners
	IsTemplate bool `bson:"isTemplate" json:"isTemplate"`
	// TemplateID references the source template for spawned instances
	TemplateID string `bson:"templateId,omitempty" json:"templateId,omitempty"`
	// InstanceSuffix is a unique suffix for spawned instances (e.g., "abc123")
	InstanceSuffix string `bson:"instanceSuffix,omitempty" json:"instanceSuffix,omitempty"`

	// Behavior Configuration
	// SpawnRoomID is the room where this NPC spawns/respawns
	SpawnRoomID string `bson:"spawnRoomId,omitempty" json:"spawnRoomId,omitempty"`
	// RespawnTime is how long after death before respawning (0 = no respawn)
	RespawnTime time.Duration `bson:"respawnTime,omitempty" json:"respawnTime,omitempty"`
	// WanderRadius is how many rooms away from spawn the NPC can wander (0 = stationary)
	WanderRadius int `bson:"wanderRadius,omitempty" json:"wanderRadius,omitempty"`
	// PatrolPath is an ordered list of room IDs for patrol behavior
	PatrolPath []string `bson:"patrolPath,omitempty" json:"patrolPath,omitempty"`

	// State Tracking
	// IsDead indicates the NPC is currently dead and awaiting respawn
	IsDead bool `bson:"isDead" json:"isDead"`
	// DeathTime is when the NPC died (for respawn timing)
	DeathTime time.Time `bson:"deathTime,omitempty" json:"deathTime,omitempty"`
	// State is the FSM state: "idle", "combat", "patrol", "dead", "fleeing"
	State string `bson:"state" json:"state"`

	// DialogID references the main interactive dialog for this NPC (stored in dialogs collection)
	DialogID string `bson:"dialogID,omitempty" json:"dialogID,omitempty"`

	// IdleDialogID references an optional idle dialog that triggers automatically
	IdleDialogID      string        `bson:"idleDialogID,omitempty" json:"idleDialogID,omitempty"`
	IdleDialogTimeout time.Duration `bson:"idleDialogTimeout,omitempty" json:"idleDialogTimeout,omitempty"`

	// Traits for specialized behaviors
	EnemyTrait    *EnemyTrait    `bson:"enemyTrait,omitempty" json:"enemyTrait,omitempty"`
	MerchantTrait *MerchantTrait `bson:"merchantTrait,omitempty" json:"merchantTrait,omitempty"`

	Created time.Time `bson:"created" json:"created,omitempty"`
	Updated time.Time `bson:"updated,omitempty" json:"updated,omitempty"`
}

// IsEnemy returns true if this NPC has enemy behavior
func (npc *NPC) IsEnemy() bool {
	return npc.EnemyTrait != nil
}

// HasDialog returns true if this NPC has an interactive dialog attached
func (npc *NPC) HasDialog() bool {
	return npc.DialogID != ""
}

// HasIdleDialog returns true if this NPC has an idle dialog attached
func (npc *NPC) HasIdleDialog() bool {
	return npc.IdleDialogID != ""
}

// IsMerchant returns true if this NPC can trade with players
func (npc *NPC) IsMerchant() bool {
	return npc.MerchantTrait != nil
}

// IsInstance returns true if this NPC is a spawned instance from a template
func (npc *NPC) IsInstance() bool {
	return npc.TemplateID != "" && npc.InstanceSuffix != ""
}

// GetDisplayName returns the name shown to players
func (npc *NPC) GetDisplayName() string {
	return npc.Name
}

// GetTargetName returns the unique name for targeting commands
// For instances, this includes the suffix to distinguish between multiple spawns
func (npc *NPC) GetTargetName() string {
	if npc.InstanceSuffix != "" {
		return npc.Name + "-" + npc.InstanceSuffix
	}
	return npc.Name
}

// ShouldRespawn returns true if this NPC is dead and ready to respawn
func (npc *NPC) ShouldRespawn() bool {
	if !npc.IsDead || npc.RespawnTime <= 0 {
		return false
	}
	return time.Since(npc.DeathTime) >= npc.RespawnTime
}
