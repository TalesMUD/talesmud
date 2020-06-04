package entities

import "go.mongodb.org/mongo-driver/bson/primitive"

//Entity ...
type Entity struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
}

// NewEntity ...
func NewEntity() *Entity {
	return &Entity{
		ID: primitive.NewObjectID(),
	}
}
