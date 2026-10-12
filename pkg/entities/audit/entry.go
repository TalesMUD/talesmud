// Package audit is the creator and live-ops change log.
package audit

import (
	"encoding/json"
	"time"
)

// SourceCreator marks a row written by a Creator CRUD request.
const SourceCreator = "creator"

// SourceOps marks a row written by a live-ops action.
const SourceOps = "ops"

// Inverse is the live-ops call that puts the world back.
// CRUD rows leave it empty and restore Before instead.
type Inverse struct {
	Action string          `json:"action"`
	Body   json.RawMessage `json:"body,omitempty"`
}

// Entry is one audited write.
type Entry struct {
	ID          string          `json:"id"`
	Time        time.Time       `json:"time"`
	ActorUserID string          `json:"actorUserId"`
	ActorName   string          `json:"actorName"`
	Action      string          `json:"action"`
	EntityType  string          `json:"entityType"`
	EntityID    string          `json:"entityId"`
	Before      json.RawMessage `json:"before,omitempty"`
	After       json.RawMessage `json:"after,omitempty"`
	Source      string          `json:"source"`
	Undoable    bool            `json:"undoable"`
	UndoneBy    string          `json:"undoneBy,omitempty"`
	Summary     string          `json:"summary,omitempty"`
	Inverse     *Inverse        `json:"inverse,omitempty"`
}
