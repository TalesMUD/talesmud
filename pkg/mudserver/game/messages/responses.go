package messages

import (
	"github.com/talesmud/talesmud/pkg/entities"
	e "github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/util"
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
	MessageAudienceRoomWithoutOrigin
	MessageAudienceGlobal
	MessageAudienceSystem
)

// MessageResponse ... Define our message object
type MessageResponse struct {
	Audience   AudienceType `json:"-"`
	AudienceID string       `json:"-"`
	OriginID   string       `json:"-"`

	Type     MessageType `json:"type"`
	Username string      `json:"username"`
	Message  string      `json:"message"`
}

//GetAudience ,,,
func (m MessageResponse) GetAudience() AudienceType {
	return m.Audience
}

//GetAudienceID ,,,
func (m MessageResponse) GetAudienceID() string {
	return m.AudienceID
}

//GetOriginID ,,,
func (m MessageResponse) GetOriginID() string {
	return m.OriginID
}

//GetMessage ,,,
func (m MessageResponse) GetMessage() string {
	return m.Message
}

//MessageResponder ...
type MessageResponder interface {
	GetAudience() AudienceType
	GetAudienceID() string
	GetOriginID() string
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

// RoomNPC represents NPC data sent to the frontend for UI rendering
type RoomNPC struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"` // includes #1, #2 suffix for duplicates
	IsEnemy     bool   `json:"isEnemy"`
	IsMerchant  bool   `json:"isMerchant"`
	CurrentHP   int32  `json:"currentHp,omitempty"`
	MaxHP       int32  `json:"maxHp,omitempty"`
	Level       int32  `json:"level,omitempty"`
	State       string `json:"state"` // idle, combat, dead
}

// EnterRoomMessage ... Define our message object
type EnterRoomMessage struct {
	MessageResponse
	Room rooms.Room `json:"room"`
	NPCs []RoomNPC  `json:"npcs"`
}

//NewEnterRoomMessage ...
func NewEnterRoomMessage(room *rooms.Room, user *entities.User, game def.GameCtrl) *EnterRoomMessage {
	// Get NPC data for frontend rendering
	roomNPCs := util.GetRoomNPCs(room, game)
	npcs := make([]RoomNPC, len(roomNPCs))
	for i, n := range roomNPCs {
		npcs[i] = RoomNPC{
			ID:          n.ID,
			Name:        n.Name,
			DisplayName: n.DisplayName,
			IsEnemy:     n.IsEnemy,
			IsMerchant:  n.IsMerchant,
			CurrentHP:   n.CurrentHP,
			MaxHP:       n.MaxHP,
			Level:       n.Level,
			State:       n.State,
		}
	}

	return &EnterRoomMessage{
		MessageResponse: MessageResponse{
			Audience: MessageAudienceOrigin,
			Type:     MessageTypeEnterRoom,
			Message:  util.CreateRoomDescription(room, user, game),
		},
		Room: *room,
		NPCs: npcs,
	}
}

// NewRoomBasedMessage ... creates a new Websocket message
func NewRoomBasedMessage(user string, message string) MessageResponse {
	return MessageResponse{
		// default
		Audience: MessageAudienceRoom,
		Type:     MessageTypeDefault,
		Message:  message,
		Username: user,
	}
}

// Reply ... creates a reply message
func Reply(userID string, message string) MessageResponse {
	return MessageResponse{
		// default
		Audience:   MessageAudienceOrigin,
		AudienceID: userID,
		Type:       MessageTypeDefault,
		Message:    message,
	}
}

// NewCreateCharacterMessage ...
func NewCreateCharacterMessage(user string) MessageResponse {
	return MessageResponse{
		Type:       MessageTypeCreateCharacter,
		Message:    "User has no characters created.",
		Audience:   MessageAudienceOrigin,
		AudienceID: user,
	}
}

// DialogOption represents a player choice in a dialog
type DialogOption struct {
	Index int    `json:"index"` // 1-based index for player input
	Text  string `json:"text"`  // The option text to display
}

// DialogMessage represents a dialog message with NPC text and player options
type DialogMessage struct {
	MessageResponse
	NPCName        string         `json:"npcName"`
	NPCText        string         `json:"npcText"`
	Options        []DialogOption `json:"options"`
	ConversationID string         `json:"conversationID"`
}

// NewDialogMessage creates a new dialog message for the player
func NewDialogMessage(userID, npcName, npcText string, options []DialogOption, conversationID string) *DialogMessage {
	return &DialogMessage{
		MessageResponse: MessageResponse{
			Audience:   MessageAudienceOrigin,
			AudienceID: userID,
			Type:       MessageTypeDialog,
			Message:    npcText,
		},
		NPCName:        npcName,
		NPCText:        npcText,
		Options:        options,
		ConversationID: conversationID,
	}
}

// NewDialogEndMessage creates a message indicating the conversation has ended
func NewDialogEndMessage(userID, npcName, message string) MessageResponse {
	return MessageResponse{
		Audience:   MessageAudienceOrigin,
		AudienceID: userID,
		Type:       MessageTypeDialogEnd,
		Message:    message,
		Username:   npcName,
	}
}
