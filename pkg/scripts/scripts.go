package scripts

import "github.com/talesmud/talesmud/pkg/entities"

// Script ...
type Script struct {
	//ID          string `bson:"id,omitempty" json:"id"`
	*entities.Entity `bson:",inline"`

	Name        string `bson:"name,omitempty" json:"name"`
	Description string `bson:"description,omitempty" json:"description"`
	Code        string `bson:"code,omitempty" json:"code"`
}
