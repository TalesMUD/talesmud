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

	// DialogID references the main interactive dialog for this NPC (stored in dialogs collection)
	DialogID string `bson:"dialogID,omitempty" json:"dialogID,omitempty"`

	// IdleDialogID references an optional idle dialog that triggers automatically
	IdleDialogID      string        `bson:"idleDialogID,omitempty" json:"idleDialogID,omitempty"`
	IdleDialogTimeout time.Duration `bson:"idleDialogTimeout,omitempty" json:"idleDialogTimeout,omitempty"`

	// Traits for specialized behaviors
	EnemyTrait    *EnemyTrait    `bson:"enemyTrait,omitempty" json:"enemyTrait,omitempty"`
	MerchantTrait *MerchantTrait `bson:"merchantTrait,omitempty" json:"merchantTrait,omitempty"`

	Created time.Time `bson:"created" json:"created,omitempty"`
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
