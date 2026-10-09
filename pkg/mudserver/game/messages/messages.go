package messages

import (
	e "github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
)

// UserJoined is an in-process event. The user pointer is not a client payload.
type UserJoined struct {
	User *e.User `json:"-"`
}

// UserQuit is an in-process event. The user pointer is not a client payload.
type UserQuit struct {
	User *e.User `json:"-"`
}

// Message is the in-process command envelope. It is not written to the socket.
// The json tags keep a mistaken WriteJSON from emitting the account or the character.
type Message struct {
	FromUser  *e.User               `json:"-"`
	Character *characters.Character `json:"-"`

	Data string `json:"data,omitempty"`

	// NoticeGen stamps command replies for the text client. Zero on classic traffic.
	NoticeGen uint64 `json:"-"`
}

// Reply o a message
func (msg *Message) Reply(message string) MessageResponse {
	rsp := Reply(msg.FromUser.ID, message)
	if msg != nil {
		rsp.NoticeGen = msg.NoticeGen
	}
	return rsp
}

// NewMessage ... creates a new message
func NewMessage(fromUser *e.User, data string) *Message {
	return &Message{
		FromUser: fromUser,
		Data:     data,
	}
}
