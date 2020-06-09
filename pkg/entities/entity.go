package entities

import (
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//Entity ...
type Entity struct {
	_ID primitive.ObjectID `bson:"_id,omitempty"`
	ID  string             `bson:"id,omitempty" json:"id"`
}

// NewEntity ...
func NewEntity() *Entity {
	return &Entity{
		_ID: primitive.NewObjectID(),
		ID:  uuid.New().String(),
	}
}
