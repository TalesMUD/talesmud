package messages

import (
	e "github.com/atla/owndnd/pkg/entities"
	"github.com/atla/owndnd/pkg/entities/characters"
	"github.com/atla/owndnd/pkg/entities/rooms"
	"github.com/atla/owndnd/pkg/mudserver/game/util"
)

// CharacterJoinedRoom ... asdd
type CharacterJoinedRoom struct {
	MessageResponse
}

// CharacterLeftRoom ... asdd
type CharacterLeftRoom struct {
	MessageResponse
}

// CharacterSelected ...
type CharacterSelected struct {
	MessageResponse
	Character *characters.Character `json:"character"`
}

// NewUserQuit ... creates a new User Joined event
func NewUserQuit(user *e.User) *UserQuit {
	return &UserQuit{
		User: user,
	}
}

// NewUserJoined ... creates a new User Joined event
func NewUserJoined(user *e.User) *UserJoined {
	return &UserJoined{
		User: user,
	}
}

//AudienceType type
type AudienceType int

const (
	MessageAudienceOrigin = iota + 1
	MessageAudienceUser
	MessageAudienceRoom
	MessageAudienceGlobal
	MessageAudienceSystem
)

// MessageResponse ... Define our message object
type MessageResponse struct {
	Audience   AudienceType `json:"-"`
	AudienceID string       `json:"-"`

	Type     string `json:"type"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

//GetAudience ,,,
func (m MessageResponse) GetAudience() AudienceType {
	return m.Audience
}

//GetAudienceID ,,,
func (m MessageResponse) GetAudienceID() string {
	return m.AudienceID
}

//GetMessage ,,,
func (m MessageResponse) GetMessage() string {
	return m.Message
}

//MessageResponder ...
type MessageResponder interface {
	GetAudience() AudienceType
	GetAudienceID() string
	GetMessage() string
}

// MultiResponse ...
type MultiResponse struct {
	Responses []MessageResponse
}

//NewMultiResponse ...
func NewMultiResponse(responses ...MessageResponse) MultiResponse {
	mr := MultiResponse{
		Responses: []MessageResponse{},
	}
	for _, rsp := range responses {
		mr.Responses = append(mr.Responses, rsp)
	}
	return mr
}

// EnterRoomMessage ... Define our message object
type EnterRoomMessage struct {
	MessageResponse
	Room rooms.Room `json:"room"`
}

//NewEnterRoomMessage ...
func NewEnterRoomMessage(room *rooms.Room) *EnterRoomMessage {
	return &EnterRoomMessage{
		MessageResponse: MessageResponse{
			Audience: MessageAudienceOrigin,
			Type:     "enterRoom",
			Message:  util.CreateRoomDescription(room),
		},
		Room: *room,
	}
}

// NewOutgoingMessage ... creates a new Websocket message
func NewOutgoingMessage(user string, message string) MessageResponse {
	return MessageResponse{
		// default
		Audience: MessageAudienceRoom,
		Type:     "message",
		Message:  message,
		Username: user,
	}

}
