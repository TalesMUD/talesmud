package npc

import (
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/traits"
)

//NPC data
type NPC struct {
	*entities.Entity   `bson:",inline"`
	traits.BelongsUser `bson:",inline"` // optionally, e.g. a bot could belog to a user
	traits.CurrentRoom `bson:",inline"`

	Name        string           `json:"name"`
	Description string           `json:"description"`
	Race        characters.Race  `json:"race,omitempty"`
	Class       characters.Class `json:"class,omitempty"`

	CurrentHitPoints int32 `json:"currentHitPoints"`
	MaxHitPoints     int32 `json:"maxHitPoints"`
	Level            int32 `json:"level"`

	EnemyTrait *EnemyTrait `json:"enemyTrait"`

	Created time.Time `bson:"created" json:"created,omitempty"`
	//Attributes Attributes `bson:"attributes" json:"attributes,omitempty"`

}

// IsEnemy ...
func (npc *NPC) IsEnemy() bool {
	return npc.EnemyTrait != nil
}

// Companion??
