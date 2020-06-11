package messages

import (
	e "github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
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

// Reply o a message
func (msg *Message) Reply(message string) MessageResponse {
	return Reply(msg.FromUser.ID, message)
}

// NewMessage ... creates a new message
func NewMessage(fromUser *e.User, data string) *Message {
	return &Message{
		FromUser: fromUser,
		Data:     data,
	}
}
