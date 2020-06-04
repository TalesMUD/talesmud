package characters

import (
	"time"

	"github.com/atla/owndnd/pkg/entities"
	"github.com/atla/owndnd/pkg/entities/traits"
)

//Race type
type Race int

const (
	rHuman Race = iota + 1
	rDwarf
	rElve
)

func (cr Race) String() string {
	return [...]string{"human", "dwarf", "elve"}[cr]
}

//Class type
type Class int

const (
	cWarrior Class = iota + 1
	cWizard
	cRanger
)

func (cr Class) String() string {
	return [...]string{"warrior", "wzard", "ranger"}[cr]
}

// Attribute data
type Attribute struct {
	Name  string `json:"name"`
	Short string `json:"short"`
	Value int32  `json:"value"`
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
	ArmorClass       int32 `json:"armorClass"`

	Created    time.Time    `bson:"created" json:"created,omitempty"`
	Attributes []*Attribute `bson:"attributes" json:"attributes,omitempty"`

	PersonalityTraits string `json:"personalityTraits,omitempty"`

	// complex character fields
	Inventory Inventory `json:"inventory"`
}
