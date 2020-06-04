package messages

import (
	e "github.com/atla/owndnd/pkg/entities"
	"github.com/atla/owndnd/pkg/entities/characters"
)

//UserJoined ... player joined event
type UserJoined struct{ User *e.User }

//UserQuit ... player joined event
type UserQuit struct{ User *e.User }

// Message ... main message container to pass data from e to server and back
type Message struct {
	FromUser  *e.User
	Character *characters.Character

	Data string
}

// NewMessage ... creates a new message
func NewMessage(fromUser *e.User, data string) *Message {
	return &Message{
		FromUser: fromUser,
		Data:     data,
	}
}
