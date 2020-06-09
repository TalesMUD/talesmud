package skills

import "github.com/talesmud/talesmud/pkg/entities"

//Skill ... type
type Skill struct {
	*entities.Entity `bson:",inline"`

	Name        string `bson:"name,omitempty" json:"name"`
	Description string `bson:"description,omitempty" json:"description"`
}