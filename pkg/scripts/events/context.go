package events

import (
	"time"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

// EventContext contains all the context data for an event
type EventContext struct {
	// Event metadata
	EventType EventType              `json:"eventType"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`

	// Common context
	Room      *rooms.Room            `json:"room,omitempty"`
	Character *characters.Character  `json:"character,omitempty"`
	NPC       *npc.NPC               `json:"npc,omitempty"`
	Item      *items.Item            `json:"item,omitempty"`

	// Movement context
	FromRoom *rooms.Room `json:"fromRoom,omitempty"`
	ToRoom   *rooms.Room `json:"toRoom,omitempty"`

	// Combat context
	Target interface{} `json:"target,omitempty"`
	Damage int32       `json:"damage,omitempty"`
	Killer interface{} `json:"killer,omitempty"`

	// Dialog context
	DialogID       string `json:"dialogId,omitempty"`
	NodeID         string `json:"nodeId,omitempty"`
	OptionSelected int    `json:"optionSelected,omitempty"`
	ConversationID string `json:"conversationId,omitempty"`

	// Quest context
	QuestID     string `json:"questId,omitempty"`
	ObjectiveID string `json:"objectiveId,omitempty"`
	Progress    int    `json:"progress,omitempty"`

	// Action context
	ActionName   string                 `json:"actionName,omitempty"`
	ActionParams map[string]interface{} `json:"actionParams,omitempty"`
}

// NewEventContext creates a new event context
func NewEventContext(eventType EventType) *EventContext {
	return &EventContext{
		EventType: eventType,
		Timestamp: time.Now(),
		Data:      make(map[string]interface{}),
	}
}

// WithRoom adds room context
func (c *EventContext) WithRoom(room *rooms.Room) *EventContext {
	c.Room = room
	return c
}

// WithCharacter adds character context
func (c *EventContext) WithCharacter(character *characters.Character) *EventContext {
	c.Character = character
	return c
}

// WithNPC adds NPC context
func (c *EventContext) WithNPC(n *npc.NPC) *EventContext {
	c.NPC = n
	return c
}

// WithItem adds item context
func (c *EventContext) WithItem(item *items.Item) *EventContext {
	c.Item = item
	return c
}

// WithMovement adds movement context (from/to rooms)
func (c *EventContext) WithMovement(from, to *rooms.Room) *EventContext {
	c.FromRoom = from
	c.ToRoom = to
	return c
}

// WithCombat adds combat context
func (c *EventContext) WithCombat(target interface{}, damage int32, killer interface{}) *EventContext {
	c.Target = target
	c.Damage = damage
	c.Killer = killer
	return c
}

// WithDialog adds dialog context
func (c *EventContext) WithDialog(dialogID, nodeID, conversationID string, optionSelected int) *EventContext {
	c.DialogID = dialogID
	c.NodeID = nodeID
	c.ConversationID = conversationID
	c.OptionSelected = optionSelected
	return c
}

// WithQuest adds quest context
func (c *EventContext) WithQuest(questID, objectiveID string, progress int) *EventContext {
	c.QuestID = questID
	c.ObjectiveID = objectiveID
	c.Progress = progress
	return c
}

// WithAction adds action context
func (c *EventContext) WithAction(name string, params map[string]interface{}) *EventContext {
	c.ActionName = name
	c.ActionParams = params
	return c
}

// Set adds custom data to the context
func (c *EventContext) Set(key string, value interface{}) *EventContext {
	c.Data[key] = value
	return c
}

// Get retrieves custom data from the context
func (c *EventContext) Get(key string) (interface{}, bool) {
	val, ok := c.Data[key]
	return val, ok
}

// ToMap converts the context to a map for Lua
func (c *EventContext) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"eventType": string(c.EventType),
		"timestamp": c.Timestamp.Unix(),
	}

	if c.Room != nil {
		m["room"] = c.Room
	}
	if c.Character != nil {
		m["character"] = c.Character
	}
	if c.NPC != nil {
		m["npc"] = c.NPC
	}
	if c.Item != nil {
		m["item"] = c.Item
	}
	if c.FromRoom != nil {
		m["fromRoom"] = c.FromRoom
	}
	if c.ToRoom != nil {
		m["toRoom"] = c.ToRoom
	}
	if c.Target != nil {
		m["target"] = c.Target
	}
	if c.Damage != 0 {
		m["damage"] = c.Damage
	}
	if c.Killer != nil {
		m["killer"] = c.Killer
	}
	if c.DialogID != "" {
		m["dialogId"] = c.DialogID
	}
	if c.NodeID != "" {
		m["nodeId"] = c.NodeID
	}
	if c.ConversationID != "" {
		m["conversationId"] = c.ConversationID
	}
	if c.OptionSelected != 0 {
		m["optionSelected"] = c.OptionSelected
	}
	if c.QuestID != "" {
		m["questId"] = c.QuestID
	}
	if c.ObjectiveID != "" {
		m["objectiveId"] = c.ObjectiveID
	}
	if c.Progress != 0 {
		m["progress"] = c.Progress
	}
	if c.ActionName != "" {
		m["actionName"] = c.ActionName
	}
	if c.ActionParams != nil {
		m["actionParams"] = c.ActionParams
	}

	// Merge custom data
	for k, v := range c.Data {
		m[k] = v
	}

	return m
}
