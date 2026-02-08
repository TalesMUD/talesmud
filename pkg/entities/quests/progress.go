package quests

import (
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
)

// QuestStatus represents the current state of a quest for a player
type QuestStatus string

const (
	QuestStatusAvailable QuestStatus = "available"
	QuestStatusActive    QuestStatus = "active"
	QuestStatusCompleted QuestStatus = "completed"
	QuestStatusFailed    QuestStatus = "failed"
	QuestStatusAbandoned QuestStatus = "abandoned"
)

// ObjectiveProgress tracks a player's progress on a single objective
type ObjectiveProgress struct {
	ObjectiveID string `json:"objectiveId"`
	Current     int32  `json:"current"`
	Required    int32  `json:"required"`
	Completed   bool   `json:"completed"`
}

// QuestProgress tracks a player's progress on a quest
type QuestProgress struct {
	*entities.Entity `bson:",inline"`

	CharacterID string      `json:"characterId"`
	QuestID     string      `json:"questId"`
	Status      QuestStatus `json:"status"`

	Objectives []ObjectiveProgress `json:"objectives"`

	AcceptedAt  time.Time `json:"acceptedAt,omitempty"`
	CompletedAt time.Time `json:"completedAt,omitempty"`
}
