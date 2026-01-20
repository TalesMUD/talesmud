package entities

import (
	"github.com/google/uuid"
)

// Entity is the base type for all entities with a unique ID.
type Entity struct {
	ID string `json:"id"`
}

// NewEntity creates a new entity with a generated UUID.
func NewEntity() *Entity {
	return &Entity{
		ID: uuid.New().String(),
	}
}
