package conversations

import (
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
)

// TargetType represents what kind of entity the conversation is with
type TargetType string

const (
	// TargetTypeNPC indicates the conversation is with an NPC
	TargetTypeNPC TargetType = "npc"
	// TargetTypeItem indicates the conversation is with an item (e.g., artifact)
	TargetTypeItem TargetType = "item"
)

// Conversation tracks the state of a dialog between a player character and a target (NPC or item).
// This allows players to resume conversations where they left off.
type Conversation struct {
	*entities.Entity `bson:",inline"`

	// CharacterID is the player's character ID
	CharacterID string `bson:"characterID" json:"characterID"`

	// TargetID is the NPC or Item ID that the conversation is with
	TargetID string `bson:"targetID" json:"targetID"`

	// TargetType indicates whether this is a conversation with an NPC or Item
	TargetType TargetType `bson:"targetType" json:"targetType"`

	// DialogID references the Dialog entity being used for this conversation
	DialogID string `bson:"dialogID" json:"dialogID"`

	// CurrentNodeID is the current position in the dialog tree (e.g., "main", "greeting")
	CurrentNodeID string `bson:"currentNodeID" json:"currentNodeID"`

	// VisitedNodes tracks how many times each dialog node has been visited
	// Key is the node ID, value is the visit count
	VisitedNodes map[string]int `bson:"visitedNodes" json:"visitedNodes"`

	// Context stores template variables for dialog rendering (e.g., player name, NPC name)
	Context map[string]string `bson:"context,omitempty" json:"context,omitempty"`

	// LastInteracted is when the player last interacted with this conversation
	LastInteracted time.Time `bson:"lastInteracted" json:"lastInteracted"`

	// Created is when this conversation was first started
	Created time.Time `bson:"created" json:"created"`
}

// NewConversation creates a new conversation state
func NewConversation(characterID, targetID string, targetType TargetType, dialogID string) *Conversation {
	now := time.Now()
	return &Conversation{
		Entity:         entities.NewEntity(),
		CharacterID:    characterID,
		TargetID:       targetID,
		TargetType:     targetType,
		DialogID:       dialogID,
		CurrentNodeID:  "main", // Start at the main/root node
		VisitedNodes:   make(map[string]int),
		Context:        make(map[string]string),
		LastInteracted: now,
		Created:        now,
	}
}

// MarkVisited increments the visit count for a dialog node
func (c *Conversation) MarkVisited(nodeID string) {
	if c.VisitedNodes == nil {
		c.VisitedNodes = make(map[string]int)
	}
	c.VisitedNodes[nodeID]++
}

// GetVisitCount returns how many times a dialog node has been visited
func (c *Conversation) GetVisitCount(nodeID string) int {
	if c.VisitedNodes == nil {
		return 0
	}
	return c.VisitedNodes[nodeID]
}

// HasVisited returns true if the player has visited a specific dialog node
func (c *Conversation) HasVisited(nodeID string) bool {
	return c.GetVisitCount(nodeID) > 0
}

// HasVisitedAll returns true if all the specified nodes have been visited
func (c *Conversation) HasVisitedAll(nodeIDs []string) bool {
	for _, nodeID := range nodeIDs {
		if !c.HasVisited(nodeID) {
			return false
		}
	}
	return true
}

// UpdateInteraction updates the last interaction time
func (c *Conversation) UpdateInteraction() {
	c.LastInteracted = time.Now()
}

// SetContext sets a context variable for dialog rendering
func (c *Conversation) SetContext(key, value string) {
	if c.Context == nil {
		c.Context = make(map[string]string)
	}
	c.Context[key] = value
}
