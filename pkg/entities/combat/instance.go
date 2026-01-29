package combat

import (
	"time"

	"github.com/google/uuid"
)

// CombatState represents the current state of a combat instance
type CombatState string

const (
	CombatStatePending CombatState = "pending" // Waiting for combat to begin
	CombatStateActive  CombatState = "active"  // Combat in progress
	CombatStateVictory CombatState = "victory" // All enemies defeated
	CombatStateDefeat  CombatState = "defeat"  // All players dead
	CombatStateFled    CombatState = "fled"    // All players fled
	CombatStateTimeout CombatState = "timeout" // Combat timed out
)

// CombatantType identifies whether a combatant is a player or NPC
type CombatantType string

const (
	CombatantTypePlayer CombatantType = "player"
	CombatantTypeNPC    CombatantType = "npc"
)

// CombatAction represents the type of action taken in combat
type CombatAction string

const (
	CombatActionAttack  CombatAction = "attack"
	CombatActionDefend  CombatAction = "defend"
	CombatActionItem    CombatAction = "item"
	CombatActionFlee    CombatAction = "flee"
	CombatActionTimeout CombatAction = "timeout" // Forced defend due to timeout
)

// CombatantRef represents a participant in combat with their combat stats
type CombatantRef struct {
	ID         string        `json:"id"`
	Type       CombatantType `json:"type"`
	Name       string        `json:"name"`
	Initiative int           `json:"initiative"`
	IsAlive    bool          `json:"isAlive"`
	HasFled    bool          `json:"hasFled"`

	// Snapshot of combat stats at combat start
	MaxHP       int32 `json:"maxHp"`
	CurrentHP   int32 `json:"currentHp"`
	AttackPower int32 `json:"attackPower"`
	Defense     int32 `json:"defense"`

	// Attribute modifiers (calculated from character attributes)
	STRMod int `json:"strMod"`
	DEXMod int `json:"dexMod"`
	CONMod int `json:"conMod"`

	// Status effects
	DefenseBonus int32 `json:"defenseBonus"` // From defend action

	// Auto-attack system
	AutoAttackTargetID string       `json:"autoAttackTargetId,omitempty"` // Persistent target for auto-attacks
	QueuedAction       CombatAction `json:"queuedAction,omitempty"`      // Next action override (flee, defend, attack)
	QueuedTargetID     string       `json:"queuedTargetId,omitempty"`    // Target for queued attack
}

// CombatLogEntry represents a single action in the combat log
type CombatLogEntry struct {
	Timestamp  time.Time    `json:"timestamp"`
	Round      int          `json:"round"`
	ActorID    string       `json:"actorId"`
	ActorName  string       `json:"actorName"`
	Action     CombatAction `json:"action"`
	TargetID   string       `json:"targetId,omitempty"`
	TargetName string       `json:"targetName,omitempty"`
	Result     string       `json:"result"` // "hit", "miss", "critical", "fled", "blocked"
	Damage     int32        `json:"damage,omitempty"`
	Message    string       `json:"message"` // Human-readable description
}

// CombatInstance represents an isolated combat encounter
type CombatInstance struct {
	ID           string `json:"id"`
	OriginRoomID string `json:"originRoomId"`

	// Participants
	Players []CombatantRef `json:"players"`
	Enemies []CombatantRef `json:"enemies"`

	// Turn Management
	TurnOrder      []CombatantRef `json:"turnOrder"`
	CurrentTurnIdx int            `json:"currentTurnIdx"`
	TurnStartTime  time.Time      `json:"turnStartTime"`
	Round          int            `json:"round"`

	// State
	State        CombatState `json:"state"`
	CreatedAt    time.Time   `json:"createdAt"`
	LastActionAt time.Time   `json:"lastActionAt"`

	// Configuration
	TurnTimeoutSec int `json:"turnTimeoutSec"` // Default: 60

	// Combat Log
	Log []CombatLogEntry `json:"log"`
}

// NewCombatInstance creates a new combat instance with a generated UUID
func NewCombatInstance(originRoomID string) *CombatInstance {
	now := time.Now()
	return &CombatInstance{
		ID:             uuid.New().String(),
		OriginRoomID:   originRoomID,
		Players:        make([]CombatantRef, 0),
		Enemies:        make([]CombatantRef, 0),
		TurnOrder:      make([]CombatantRef, 0),
		CurrentTurnIdx: 0,
		Round:          1,
		State:          CombatStatePending,
		CreatedAt:      now,
		LastActionAt:   now,
		TurnTimeoutSec: 60,
		Log:            make([]CombatLogEntry, 0),
	}
}

// GetCurrentTurnCombatant returns the combatant whose turn it is
func (c *CombatInstance) GetCurrentTurnCombatant() *CombatantRef {
	if len(c.TurnOrder) == 0 || c.CurrentTurnIdx >= len(c.TurnOrder) {
		return nil
	}
	return &c.TurnOrder[c.CurrentTurnIdx]
}

// IsPlayerTurn returns true if it's currently a player's turn
func (c *CombatInstance) IsPlayerTurn() bool {
	current := c.GetCurrentTurnCombatant()
	if current == nil {
		return false
	}
	return current.Type == CombatantTypePlayer
}

// GetCombatantByID finds a combatant by their ID
func (c *CombatInstance) GetCombatantByID(id string) *CombatantRef {
	for i := range c.Players {
		if c.Players[i].ID == id {
			return &c.Players[i]
		}
	}
	for i := range c.Enemies {
		if c.Enemies[i].ID == id {
			return &c.Enemies[i]
		}
	}
	return nil
}

// GetPlayerByID finds a player combatant by their character ID
func (c *CombatInstance) GetPlayerByID(id string) *CombatantRef {
	for i := range c.Players {
		if c.Players[i].ID == id {
			return &c.Players[i]
		}
	}
	return nil
}

// GetEnemyByID finds an enemy combatant by their NPC ID
func (c *CombatInstance) GetEnemyByID(id string) *CombatantRef {
	for i := range c.Enemies {
		if c.Enemies[i].ID == id {
			return &c.Enemies[i]
		}
	}
	return nil
}

// GetLivingPlayers returns all players still alive in combat
func (c *CombatInstance) GetLivingPlayers() []*CombatantRef {
	result := make([]*CombatantRef, 0)
	for i := range c.Players {
		if c.Players[i].IsAlive && !c.Players[i].HasFled {
			result = append(result, &c.Players[i])
		}
	}
	return result
}

// GetLivingEnemies returns all enemies still alive in combat
func (c *CombatInstance) GetLivingEnemies() []*CombatantRef {
	result := make([]*CombatantRef, 0)
	for i := range c.Enemies {
		if c.Enemies[i].IsAlive {
			result = append(result, &c.Enemies[i])
		}
	}
	return result
}

// AllPlayersDead returns true if all players are dead
func (c *CombatInstance) AllPlayersDead() bool {
	for _, p := range c.Players {
		if p.IsAlive && !p.HasFled {
			return false
		}
	}
	return true
}

// AllEnemiesDead returns true if all enemies are dead
func (c *CombatInstance) AllEnemiesDead() bool {
	for _, e := range c.Enemies {
		if e.IsAlive {
			return false
		}
	}
	return true
}

// AllPlayersFled returns true if all players have fled (none dead, all fled)
func (c *CombatInstance) AllPlayersFled() bool {
	for _, p := range c.Players {
		if p.IsAlive && !p.HasFled {
			return false
		}
	}
	// Make sure at least one fled (not all dead)
	for _, p := range c.Players {
		if p.HasFled {
			return true
		}
	}
	return false
}

// AddLogEntry adds a new entry to the combat log
func (c *CombatInstance) AddLogEntry(entry CombatLogEntry) {
	entry.Timestamp = time.Now()
	entry.Round = c.Round
	c.Log = append(c.Log, entry)
}

// GetTurnTimeRemaining returns seconds remaining in current turn
func (c *CombatInstance) GetTurnTimeRemaining() int {
	elapsed := int(time.Since(c.TurnStartTime).Seconds())
	remaining := c.TurnTimeoutSec - elapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// IsTurnTimedOut returns true if the current turn has exceeded the timeout
func (c *CombatInstance) IsTurnTimedOut() bool {
	return time.Since(c.TurnStartTime).Seconds() >= float64(c.TurnTimeoutSec)
}

// UpdateCombatantInTurnOrder updates a combatant's data in the turn order
// This is needed because TurnOrder contains copies, not references
func (c *CombatInstance) UpdateCombatantInTurnOrder(id string) {
	// Find the source combatant (from Players or Enemies)
	var source *CombatantRef
	for i := range c.Players {
		if c.Players[i].ID == id {
			source = &c.Players[i]
			break
		}
	}
	if source == nil {
		for i := range c.Enemies {
			if c.Enemies[i].ID == id {
				source = &c.Enemies[i]
				break
			}
		}
	}
	if source == nil {
		return
	}

	// Update the turn order entry
	for i := range c.TurnOrder {
		if c.TurnOrder[i].ID == id {
			c.TurnOrder[i] = *source
			break
		}
	}
}
