package messages

import (
	"github.com/talesmud/talesmud/pkg/entities"
	e "github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
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
	Character      *characters.Character `json:"character"`
	XPForNextLevel int32                 `json:"xpForNextLevel"`
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

// AudienceType type
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

// GetAudience ,,,
func (m MessageResponse) GetAudience() AudienceType {
	return m.Audience
}

// GetAudienceID ,,,
func (m MessageResponse) GetAudienceID() string {
	return m.AudienceID
}

// GetOriginID ,,,
func (m MessageResponse) GetOriginID() string {
	return m.OriginID
}

// GetMessage ,,,
func (m MessageResponse) GetMessage() string {
	return m.Message
}

// MessageResponder ...
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

// NewMultiResponse ...
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
	ID            string `json:"id"`
	Name          string `json:"name"`
	DisplayName   string `json:"displayName"` // includes #1, #2 suffix for duplicates
	IsEnemy       bool   `json:"isEnemy"`
	IsMerchant    bool   `json:"isMerchant"`
	IsQuestGiver  bool   `json:"isQuestGiver"`
	HasDialog     bool   `json:"hasDialog"`
	HasIdleDialog bool   `json:"hasIdleDialog"`
	CurrentHP     int32  `json:"currentHp,omitempty"`
	MaxHP         int32  `json:"maxHp,omitempty"`
	Level         int32  `json:"level,omitempty"`
	State         string `json:"state"` // idle, patrol, combat, dead
}

// RoomPlayer represents player character data sent to the frontend for UI rendering
type RoomPlayer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	IsYou bool   `json:"isYou"`
}

// EnterRoomMessage ... Define our message object
type EnterRoomMessage struct {
	MessageResponse
	Room    rooms.Room      `json:"room"`
	NPCs    []RoomNPC       `json:"npcs"`
	Items   []util.RoomItem `json:"items"`
	Players []RoomPlayer    `json:"players"`
}

// RoomPresenceMessage refreshes the online player list for a room without
// re-rendering the room description or entities.
type RoomPresenceMessage struct {
	MessageResponse
	RoomID  string       `json:"roomId"`
	Players []RoomPlayer `json:"players"`
}

// NewEnterRoomMessage creates a new enter room message.
// char is optional — if provided, CopyOnPickup items already collected by this character are hidden.
func NewEnterRoomMessage(room *rooms.Room, user *entities.User, game def.GameCtrl, char *characters.Character) *EnterRoomMessage {
	// Get NPC data for frontend rendering
	roomNPCs := util.GetRoomNPCs(room, game)
	npcs := make([]RoomNPC, len(roomNPCs))
	for i, n := range roomNPCs {
		npcs[i] = RoomNPC{
			ID:            n.ID,
			Name:          n.Name,
			DisplayName:   n.DisplayName,
			IsEnemy:       n.IsEnemy,
			IsMerchant:    n.IsMerchant,
			IsQuestGiver:  n.IsQuestGiver,
			HasDialog:     n.HasDialog,
			HasIdleDialog: n.HasIdleDialog,
			CurrentHP:     n.CurrentHP,
			MaxHP:         n.MaxHP,
			Level:         n.Level,
			State:         n.State,
		}
	}

	// Get item data for frontend rendering (filtered by character)
	roomItems := util.GetRoomItems(room, game, char)

	// Get player character data for frontend rendering
	roomChars := util.GetRoomCharacters(room, user, game)
	players := make([]RoomPlayer, len(roomChars))
	for i, c := range roomChars {
		players[i] = RoomPlayer{
			ID:    c.ID,
			Name:  c.Name,
			IsYou: c.IsYou,
		}
	}

	return &EnterRoomMessage{
		MessageResponse: MessageResponse{
			Audience: MessageAudienceOrigin,
			Type:     MessageTypeEnterRoom,
			Message:  util.CreateRoomDescription(room, user, game),
		},
		Room:    *room,
		NPCs:    npcs,
		Items:   roomItems,
		Players: players,
	}
}

// NewRoomUpdateMessage creates a silent room update that refreshes exits/items/NPCs/players
// on the client without re-rendering the full room description text.
func NewRoomUpdateMessage(room *rooms.Room, user *entities.User, game def.GameCtrl, char *characters.Character) *EnterRoomMessage {
	roomNPCs := util.GetRoomNPCs(room, game)
	npcs := make([]RoomNPC, len(roomNPCs))
	for i, n := range roomNPCs {
		npcs[i] = RoomNPC{
			ID:            n.ID,
			Name:          n.Name,
			DisplayName:   n.DisplayName,
			IsEnemy:       n.IsEnemy,
			IsMerchant:    n.IsMerchant,
			IsQuestGiver:  n.IsQuestGiver,
			HasDialog:     n.HasDialog,
			HasIdleDialog: n.HasIdleDialog,
			CurrentHP:     n.CurrentHP,
			MaxHP:         n.MaxHP,
			Level:         n.Level,
			State:         n.State,
		}
	}

	roomItems := util.GetRoomItems(room, game, char)

	// Get player character data for frontend rendering
	roomChars := util.GetRoomCharacters(room, user, game)
	players := make([]RoomPlayer, len(roomChars))
	for i, c := range roomChars {
		players[i] = RoomPlayer{
			ID:    c.ID,
			Name:  c.Name,
			IsYou: c.IsYou,
		}
	}

	return &EnterRoomMessage{
		MessageResponse: MessageResponse{
			Audience: MessageAudienceOrigin,
			Type:     MessageTypeRoomUpdate,
		},
		Room:    *room,
		NPCs:    npcs,
		Items:   roomItems,
		Players: players,
	}
}

// NewRoomPresenceMessage creates a silent player-presence update for all clients
// in a room. IsYou is intentionally false; the client marks the current
// character locally because this message is broadcast to multiple users.
func NewRoomPresenceMessage(room *rooms.Room, game def.GameCtrl) *RoomPresenceMessage {
	roomChars := util.GetRoomPresenceCharacters(room, game)
	players := make([]RoomPlayer, len(roomChars))
	for i, c := range roomChars {
		players[i] = RoomPlayer{
			ID:    c.ID,
			Name:  c.Name,
			IsYou: false,
		}
	}

	return &RoomPresenceMessage{
		MessageResponse: MessageResponse{
			Audience:   MessageAudienceRoom,
			AudienceID: room.ID,
			Type:       MessageTypeRoomPresence,
		},
		RoomID:  room.ID,
		Players: players,
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

// InventoryUpdateMessage sends the full inventory state to the client
type InventoryUpdateMessage struct {
	MessageResponse
	Inventory     interface{} `json:"inventory"`
	EquippedItems interface{} `json:"equippedItems"`
	Gold          int64       `json:"gold"`
}

// NewInventoryUpdateMessage creates an inventory update from the current character state
func NewInventoryUpdateMessage(msg *Message) *InventoryUpdateMessage {
	if msg == nil || msg.Character == nil {
		return nil
	}
	ch := msg.Character
	return &InventoryUpdateMessage{
		MessageResponse: MessageResponse{
			Audience:   MessageAudienceOrigin,
			AudienceID: msg.FromUser.ID,
			Type:       MessageTypeInventoryUpdate,
		},
		Inventory:     ch.Inventory,
		EquippedItems: ch.EquippedItems,
		Gold:          ch.Gold,
	}
}

// CharacterUpdateMessage sends updated character stats to the client (e.g., after combat damage)
type CharacterUpdateMessage struct {
	MessageResponse
	CurrentHitPoints       int32                 `json:"currentHitPoints"`
	MaxHitPoints           int32                 `json:"maxHitPoints"`
	CurrentMana            int32                 `json:"currentMana"`
	MaxMana                int32                 `json:"maxMana"`
	XP                     int32                 `json:"xp"`
	XPForNextLevel         int32                 `json:"xpForNextLevel"`
	Level                  int32                 `json:"level"`
	Gold                   int64                 `json:"gold"`
	InCombat               bool                  `json:"inCombat"`
	Attributes             characters.Attributes `json:"attributes,omitempty"`
	EquippedSkills         []string              `json:"equippedSkills,omitempty"`
	UnspentAttributePoints int32                 `json:"unspentAttributePoints"`
	SpentAttributePoints   map[string]int32      `json:"spentAttributePoints,omitempty"`

	// Derived combat stats (computed from attributes + equipment)
	AttackPower  int32  `json:"attackPower"`
	AttackAttr   string `json:"attackAttr"`
	WeaponDamage int32  `json:"weaponDamage"`
	AttackMod    int32  `json:"attackMod"`
	Defense      int32  `json:"defense"`
	ManaRegen    int32  `json:"manaRegen,omitempty"`
}

// NewCharacterUpdateMessage creates a character update from the current character state
func NewCharacterUpdateMessage(userID string, ch *characters.Character) *CharacterUpdateMessage {
	if ch == nil {
		return nil
	}

	// Compute derived combat stats using class primary attribute
	weaponDmg := ch.GetWeaponDamage()
	atkMod := int32(ch.GetPrimaryAttackMod())
	attackPower := weaponDmg + atkMod
	if attackPower < 1 {
		attackPower = 1
	}

	return &CharacterUpdateMessage{
		MessageResponse: MessageResponse{
			Audience:   MessageAudienceOrigin,
			AudienceID: userID,
			Type:       MessageTypeCharacterUpdate,
		},
		CurrentHitPoints:       ch.CurrentHitPoints,
		MaxHitPoints:           ch.MaxHitPoints,
		CurrentMana:            ch.CurrentMana,
		MaxMana:                ch.MaxMana,
		XP:                     ch.XP,
		XPForNextLevel:         leveling.GetXPRequired(ch.Level + 1),
		Level:                  ch.Level,
		Gold:                   ch.Gold,
		InCombat:               ch.InCombat,
		Attributes:             ch.Attributes,
		EquippedSkills:         ch.EquippedSkills,
		UnspentAttributePoints: ch.UnspentAttributePoints,
		SpentAttributePoints:   ch.SpentAttributePoints,
		AttackPower:            attackPower,
		AttackAttr:             ch.GetPrimaryAttackAttribute(),
		WeaponDamage:           weaponDmg,
		AttackMod:              atkMod,
		Defense:                ch.GetArmorDefense(),
		ManaRegen:              ch.CalculateManaRegen(),
	}
}

// QuestObjectiveProgress represents objective progress sent to the client
type QuestObjectiveProgress struct {
	ObjectiveID string `json:"objectiveId"`
	Description string `json:"description"`
	Current     int32  `json:"current"`
	Required    int32  `json:"required"`
	Completed   bool   `json:"completed"`
}

// QuestUpdateMessage notifies the client of a quest status change
type QuestUpdateMessage struct {
	MessageResponse
	QuestName        string                   `json:"questName"`
	QuestID          string                   `json:"questId"`
	Status           string                   `json:"status"`
	Objectives       []QuestObjectiveProgress `json:"objectives,omitempty"`
	ChangedObjective *QuestObjectiveProgress  `json:"changedObjective,omitempty"`
}

// QuestReward represents quest rewards sent to the client
type QuestReward struct {
	XP              int32    `json:"xp,omitempty"`
	Gold            int64    `json:"gold,omitempty"`
	ItemTemplateIDs []string `json:"itemTemplateIds,omitempty"`
}

// QuestLogEntry represents a quest in the quest log
type QuestLogEntry struct {
	QuestID       string                   `json:"questId"`
	QuestName     string                   `json:"questName"`
	Description   string                   `json:"description,omitempty"`
	Category      string                   `json:"category,omitempty"`
	Level         int32                    `json:"level,omitempty"`
	Status        string                   `json:"status"`
	ReadyToTurnIn bool                     `json:"readyToTurnIn,omitempty"`
	Objectives    []QuestObjectiveProgress `json:"objectives"`
	Rewards       *QuestReward             `json:"rewards,omitempty"`
	AcceptedAt    string                   `json:"acceptedAt,omitempty"`
	CompletedAt   string                   `json:"completedAt,omitempty"`
}

// QuestLogMessage sends the full quest log to the client
type QuestLogMessage struct {
	MessageResponse
	Quests []QuestLogEntry `json:"quests"`
}

// NewQuestUpdateMessage creates a quest update message for the player
func NewQuestUpdateMessage(userID, questID, questName, status string, objectives []QuestObjectiveProgress) *QuestUpdateMessage {
	return &QuestUpdateMessage{
		MessageResponse: MessageResponse{
			Audience:   MessageAudienceOrigin,
			AudienceID: userID,
			Type:       MessageType(status),
			Message:    questName,
		},
		QuestName:  questName,
		QuestID:    questID,
		Status:     status,
		Objectives: objectives,
	}
}

// NewQuestLogMessage creates a quest log message for the player
func NewQuestLogMessage(userID string, entries []QuestLogEntry) *QuestLogMessage {
	return &QuestLogMessage{
		MessageResponse: MessageResponse{
			Audience:   MessageAudienceOrigin,
			AudienceID: userID,
			Type:       MessageTypeQuestLog,
			Message:    "Quest Log",
		},
		Quests: entries,
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
