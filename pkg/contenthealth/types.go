// Package contenthealth runs one rule set over a world snapshot.
// It wraps the existing validators, adds graph checks, and applies pack rules.
package contenthealth

import "time"

const (
	RuleUnreachable    = "unreachable-rooms"
	RuleRevealMissing  = "reveal-missing-exit"
	RuleHiddenNoReveal = "hidden-exit-no-revealer"
	RuleQuest          = "quest-impossible"
	RuleBoss           = "boss-without-spawner"
	RuleUnknownTier    = "unknown-difficulty"
	RuleUnreferenced   = "unreferenced-script"
	RuleDangling       = "dangling-exits"
	RuleMissingItem    = "missing-room-item"
)

// Related is another entity a hit points at.
type Related struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// Hit is one rule failure on one entity.
type Hit struct {
	RuleID     string    `json:"ruleId"`
	Severity   string    `json:"severity"`
	EntityType string    `json:"entityType"`
	EntityID   string    `json:"entityId"`
	EntityName string    `json:"entityName,omitempty"`
	Field      string    `json:"field,omitempty"`
	Message    string    `json:"message"`
	Related    []Related `json:"related,omitempty"`
	FixHint    string    `json:"fixHint,omitempty"`
	Group      string    `json:"group,omitempty"`
}

// Rule is one check and the hits it produced.
type Rule struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Severity string `json:"severity"`
	Muted    bool   `json:"muted"`
	Summary  string `json:"summary"`
	Hits     []Hit  `json:"hits"`
}

// Summary counts unmuted hits. Muted is a hit count, not a rule count.
type Summary struct {
	Errors        int `json:"errors"`
	Warnings      int `json:"warnings"`
	Info          int `json:"info"`
	Muted         int `json:"muted"`
	Drift         int `json:"drift"`
	LiveAnomalies int `json:"liveAnomalies"`
}

// Anomaly is a live-world problem outside the content snapshot.
type Anomaly struct {
	Kind       string `json:"kind"`
	EntityType string `json:"entityType"`
	EntityID   string `json:"entityId"`
	Name       string `json:"name,omitempty"`
	Message    string `json:"message"`
}

// Report is the content-health result.
type Report struct {
	GeneratedAt   time.Time  `json:"generatedAt"`
	ContentCommit string     `json:"contentCommit"`
	Summary       Summary    `json:"summary"`
	Rules         []Rule     `json:"rules"`
	LiveAnomalies []Anomaly  `json:"liveAnomalies"`
	Drift         []DriftRow `json:"-"`
}

// Balance is the tier sets the difficulty rules consult.
// Tiers are DifficultyMultipliers keys. BossTiers are phase_tiers and enrage_tiers.
// Neither set aliases one tier name onto another.
type Balance struct {
	Tiers     map[string]bool
	BossTiers map[string]bool
}

// LiveCharacter is the slice of a character the anomaly check needs.
type LiveCharacter struct {
	ID            string
	Name          string
	CurrentRoomID string
}

// LiveView is optional. A nil view skips live anomalies.
type LiveView struct {
	Characters      []LiveCharacter
	RoomIDs         map[string]bool
	InstanceRoomIDs []string
}

// Options controls one health run.
type Options struct {
	Rules    []PackRule
	Muted    []string
	Balance  *Balance
	Live     *LiveView
	Baseline *Baseline
}
