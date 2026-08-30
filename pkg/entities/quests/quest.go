package quests

import (
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
)

// ObjectiveType defines the kind of objective a player must complete
type ObjectiveType string

const (
	ObjectiveKill    ObjectiveType = "kill"    // Kill N of NPC template
	ObjectiveCollect ObjectiveType = "collect" // Have N of item template in inventory
	ObjectiveDeliver ObjectiveType = "deliver" // Bring item to NPC
	ObjectiveVisit   ObjectiveType = "visit"   // Enter a specific room
	ObjectiveTalk    ObjectiveType = "talk"    // Talk to NPC (reach dialog node)
	ObjectiveCustom  ObjectiveType = "custom"  // Lua script check
)

// Objective represents a single quest objective
type Objective struct {
	ID          string        `json:"id"`
	Type        ObjectiveType `json:"type"`
	Description string        `json:"description"`

	// Data-driven fields (used by type)
	TargetID   string `json:"targetId,omitempty"`
	TargetName string `json:"targetName,omitempty"`
	Amount     int32  `json:"amount,omitempty"`

	// Deliver-specific
	DeliverToNPCID   string `json:"deliverToNpcId,omitempty"`
	DeliverToNPCName string `json:"deliverToNpcName,omitempty"`

	// Talk-specific
	DialogNodeID string `json:"dialogNodeId,omitempty"`

	// Custom script (ObjectiveCustom or override for any type)
	CheckScriptID string `json:"checkScriptId,omitempty"`

	// Ordering
	Order int32 `json:"order,omitempty"`
}

// Reward represents quest completion rewards
type Reward struct {
	XP              int32    `json:"xp,omitempty"`
	Gold            int64    `json:"gold,omitempty"`
	ItemTemplateIDs []string `json:"itemTemplateIds,omitempty"`
}

// QuestSource defines how a player can acquire a quest
type QuestSource struct {
	Type   string `json:"type"`            // "npc", "item", "auto", "script"
	NPCID  string `json:"npcId,omitempty"` // NPC that offers quest
	ItemID string `json:"itemId,omitempty"`
}

// Quest represents a quest definition created by builders
type Quest struct {
	*entities.Entity `bson:",inline"`

	Name        string `json:"name"`
	Description string `json:"description"`

	// Classification
	Category   string `json:"category,omitempty"` // "main", "side", "daily"
	Area       string `json:"area,omitempty"`
	Level      int32  `json:"level,omitempty"`
	Repeatable bool   `json:"repeatable,omitempty"`

	// Source - how player gets this quest
	Source QuestSource `json:"source"`

	// Objectives
	Objectives []Objective `json:"objectives"`

	// Completion
	OnAcceptScriptID   string `json:"onAcceptScriptId,omitempty"`
	OnCompleteScriptID string `json:"onCompleteScriptId,omitempty"`
	Rewards            Reward `json:"rewards"`

	// Prerequisites
	RequiredQuestIDs []string `json:"requiredQuestIds,omitempty"`
	RequiredLevel    int32    `json:"requiredLevel,omitempty"`

	// Dialog integration
	AcceptDialogText   string `json:"acceptDialogText,omitempty"`
	ProgressDialogText string `json:"progressDialogText,omitempty"`
	CompleteDialogText string `json:"completeDialogText,omitempty"`

	Created time.Time `json:"created,omitempty"`
	Updated time.Time `json:"updated,omitempty"`
}
