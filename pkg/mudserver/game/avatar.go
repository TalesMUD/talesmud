package game

import (
	"github.com/atla/owndnd/pkg/entities"
	"github.com/atla/owndnd/pkg/entities/characters"
)

// Avatar ... default active entity that moves in the world
// Avatars can be either controlled by Players/Users or be attached/belong to bots
// Once a user is logged in he automatically gets attached his last used aavatar
type Avatar struct {
	ID        string
	User      *entities.User
	Character *characters.Character
}

// NewAvatar ... creates and returns a new room instance
func NewAvatar() *Avatar {
	return &Avatar{}
}
