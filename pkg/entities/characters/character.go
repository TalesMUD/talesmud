package characters

import (
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/traits"
)

// Attribute data
type Attribute struct {
	Name  string `json:"name"`
	Short string `json:"short"`
	Value int32  `json:"value"`
}

//Attributes ...
type Attributes []Attribute

//NewAttribute ...
func NewAttribute(name, short string, value int32) Attribute {
	return Attribute{
		Name:  name,
		Short: short,
		Value: value,
	}
}

//Character data
type Character struct {
	*entities.Entity   `bson:",inline"`
	traits.BelongsUser `bson:",inline"`
	traits.CurrentRoom `bson:",inline"`

	Name        string `json:"name"`
	Description string `json:"description"`
	Race        Race   `json:"race"`
	Class       Class  `json:"class"`

	CurrentHitPoints int32 `json:"currentHitPoints"`
	MaxHitPoints     int32 `json:"maxHitPoints"`

	XP    int32 `json:"xp"`
	Level int32 `json:"level"`

	Created    time.Time  `bson:"created" json:"created,omitempty"`
	Attributes Attributes `bson:"attributes" json:"attributes,omitempty"`

	PersonalityTraits string `json:"personalityTraits,omitempty"`

	// complex character fields
	Inventory Inventory `json:"inventory"`
}
