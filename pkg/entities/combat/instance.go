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
	CombatStateDefeat  CombatState = "defeat"  // No one left fighting, and at least one player died
	CombatStateFled    CombatState = "fled"    // No one left fighting, and every player fled
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
	CombatActionSkill   CombatAction = "skill"
	CombatActionFlee    CombatAction = "flee"
	CombatActionTimeout CombatAction = "timeout" // Forced defend due to timeout
	CombatActionBolt    CombatAction = "bolt"
	CombatActionRig     CombatAction = "rig"
)

// CombatPhase tracks authored turn pacing for a combat instance (C1).
type CombatPhase string

const (
	CombatPhaseIdle          CombatPhase = ""              // Ready to start/advance a turn
	CombatPhaseWaitingPlayer CombatPhase = "waitingPlayer" // Player decision window open
	CombatPhasePlayingBeat   CombatPhase = "playingBeat"   // Post-action beat budget running
	CombatPhaseResolving     CombatPhase = "resolving"     // Action being resolved this tick
)

// OnHitDot is a content-authored weapon proc applied after a successful basic attack hit.
// SkillID is used as the refresh key (reapply refreshes duration; no stack spam).
type OnHitDot struct {
	Active   bool   `json:"active,omitempty"`
	ID       string `json:"id,omitempty"`       // stable id for refresh (e.g. vigil_burn)
	Name     string `json:"name,omitempty"`     // display name (e.g. Vigil Burn)
	Damage   int32  `json:"damage,omitempty"`   // damage per tick
	Duration int    `json:"duration,omitempty"` // rounds
	Message  string `json:"message,omitempty"`  // optional combat-log flavor on apply
}

// StatusEffect represents an active buff, debuff, DoT, or HoT on a combatant
type StatusEffect struct {
	ID       string  `json:"id"`
	SkillID  string  `json:"skillId"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`     // "buff", "debuff", "dot", "hot", "stun"
	Stat     string  `json:"stat"`     // "attack", "defense", "dodge", "mana_shield"
	Value    int32   `json:"value"`    // Flat modifier or damage/heal per tick
	Percent  float64 `json:"percent"`  // Percentage modifier (e.g. 0.30 = +30%)
	Duration int     `json:"duration"` // Rounds remaining
	SourceID string  `json:"sourceId"` // Who applied this effect
}

// CombatantRef represents a participant in combat with their combat stats
type CombatantRef struct {
	ID         string        `json:"id"`
	Type       CombatantType `json:"type"`
	Name       string        `json:"name"`
	Portrait   string        `json:"portrait,omitempty"`
	TemplateID string        `json:"templateId,omitempty"`
	Initiative int           `json:"initiative"`
	IsAlive    bool          `json:"isAlive"`
	HasFled    bool          `json:"hasFled"`

	// Snapshot of combat stats at combat start
	Level      int32  `json:"level,omitempty"`
	ClassID    string `json:"classId,omitempty"`
	Difficulty string `json:"difficulty,omitempty"`
	// TelegraphAbility is the wind-up name while TelegraphTurns is still counting down.
	TelegraphAbility string `json:"telegraphAbility,omitempty"`
	TelegraphTurns   int    `json:"telegraphTurns,omitempty"`
	BossPhase        int    `json:"bossPhase,omitempty"` // one-based; never moves backward
	BossPhaseLabel   string `json:"bossPhaseLabel,omitempty"`
	BossPhaseCount   int    `json:"bossPhaseCount,omitempty"`
	Enraged          bool   `json:"enraged,omitempty"`
	MaxHP            int32  `json:"maxHp"`
	CurrentHP        int32  `json:"currentHp"`
	AttackPower      int32  `json:"attackPower"`
	Defense          int32  `json:"defense"`
	// AttackSpeed is attacks per round. 0 matches one swing and never holds.
	AttackSpeed        float64 `json:"attackSpeed,omitempty"`
	AttackActions      int     `json:"attackActions,omitempty"`
	OnAggroScript      string  `json:"onAggroScript,omitempty"`
	OnDeathScript      string  `json:"onDeathScript,omitempty"`
	OnFleeScript       string  `json:"onFleeScript,omitempty"`
	OnLowHealthScript  string  `json:"onLowHealthScript,omitempty"`
	LowHealthThreshold float64 `json:"lowHealthThreshold,omitempty"`
	// LowHealthSeen is set when this fight first observed the NPC's HP.
	LowHealthSeen bool `json:"-"`
	// LowHealthHP is the HP at the previous low-health check.
	LowHealthHP int32 `json:"-"`
	// Summoned adds are script spawns. They grant no loot, XP, or quest credit
	// and are removed when the fight ends.
	Summoned bool `json:"summoned,omitempty"`

	// Attribute modifiers (calculated from character attributes)
	STRMod int `json:"strMod"`
	DEXMod int `json:"dexMod"`
	CONMod int `json:"conMod"`
	INTMod int `json:"intMod"`
	WISMod int `json:"wisMod"`

	// Mana (for caster classes)
	MaxMana     int32 `json:"maxMana,omitempty"`
	CurrentMana int32 `json:"currentMana,omitempty"`
	ManaRegen   int32 `json:"manaRegen,omitempty"`

	// Status effects
	DefenseBonus  int32          `json:"defenseBonus"`            // From defend action
	StatusEffects []StatusEffect `json:"statusEffects,omitempty"` // Active buffs, debuffs, DoTs, HoTs

	// On-hit weapon proc (snapshotted from equipped main-hand at combat start).
	// Content-driven: engine applies generically; game-set names/numbers live in item attrs / Lua.
	OnHitScriptID string   `json:"onHitScriptId,omitempty"`
	OnHitDot      OnHitDot `json:"onHitDot,omitempty"`

	// Skill cooldowns (skillID → rounds remaining)
	SkillCooldowns map[string]int `json:"skillCooldowns,omitempty"`

	// Equipped skills (copied from character at combat start)
	EquippedSkills []string `json:"equippedSkills,omitempty"`

	// Class kit state for this fight. Charges start empty; the skill button arms them.
	// BraceLeft halves the next landed hit. PinLeft arms Pinned. SmokeMiss misses the next swing.
	BraceLeft    int             `json:"braceLeft,omitempty"`
	SlipLeft     int             `json:"slipLeft,omitempty"`
	PinLeft      int             `json:"pinLeft,omitempty"`
	Pinned       bool            `json:"pinned,omitempty"`
	SmokeMiss    bool            `json:"smokeMiss,omitempty"`
	StandRounds  int             `json:"standRounds,omitempty"`
	HobbleRounds int             `json:"hobbleRounds,omitempty"`
	GlyphCut     int32           `json:"glyphCut,omitempty"`
	KitSpent     map[string]bool `json:"kitSpent,omitempty"`
	SlipTo       string          `json:"slipTo,omitempty"`

	// Race and weapon kind, snapshotted so racial bonuses apply once per swing.
	RaceID        string `json:"raceId,omitempty"`
	WeaponSubType string `json:"weaponSubType,omitempty"`

	// BoltLeft arms ScrapArmed on a target. RigLeft drops one turret.
	BoltLeft   int  `json:"boltLeft,omitempty"`
	RigLeft    int  `json:"rigLeft,omitempty"`
	ScrapArmed bool `json:"scrapArmed,omitempty"`

	// Ward. Grit stacks when a damaging hit lands, cap 5, for this fight only.
	// GuardCharges is the next hit inside GuardRounds. GuardSelf makes that hit grant two Grit.
	Grit          int    `json:"grit,omitempty"`
	GuardTargetID string `json:"guardTargetId,omitempty"`
	GuardRounds   int    `json:"guardRounds,omitempty"`
	GuardCharges  int    `json:"guardCharges,omitempty"`
	GuardSelf     bool   `json:"guardSelf,omitempty"`

	// Auto-attack system
	AutoAttackTargetID string       `json:"autoAttackTargetId,omitempty"` // Persistent target for auto-attacks
	QueuedAction       CombatAction `json:"queuedAction,omitempty"`       // Next action override (flee, defend, attack, skill)
	QueuedTargetID     string       `json:"queuedTargetId,omitempty"`     // Target for queued attack
	QueuedSkillID      string       `json:"queuedSkillId,omitempty"`      // Skill ID for queued skill action
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

	// Authored pacing (C1) — gate processAllTurns so fights are readable
	Phase            CombatPhase `json:"phase"`
	NextActionAt     time.Time   `json:"nextActionAt"`               // Do not resolve next turn before this
	DecisionDeadline time.Time   `json:"decisionDeadline,omitempty"` // Player decision window end

	// State
	State        CombatState `json:"state"`
	CreatedAt    time.Time   `json:"createdAt"`
	LastActionAt time.Time   `json:"lastActionAt"`

	// Configuration
	TurnTimeoutSec int `json:"turnTimeoutSec"` // Legacy absolute turn timeout (default 60); DecisionWindowSeconds is the player action window

	// Combat Log
	Log []CombatLogEntry `json:"log"`

	// Rig is the room turret. It is not a pet, not a follower, and not in turn order.
	Rig *RigTurret `json:"rig,omitempty"`

	// HookOnce records npcID|hook so onAggro, onDeath, onFlee, and onLowHealth run once per fight.
	HookOnce map[string]bool `json:"-"`

	// SummonsUsed counts adds spawned by tales.combat.summon during this fight.
	SummonsUsed int `json:"-"`
}

// RigTurret is a construct dropped in the fight's room. It does not move or follow.
type RigTurret struct {
	ID         string  `json:"id"`
	OwnerID    string  `json:"ownerId"`
	Name       string  `json:"name"`
	RoomID     string  `json:"roomId"`
	RoundsLeft int     `json:"roundsLeft"`
	Follows    bool    `json:"follows"`
	Mult       float64 `json:"mult,omitempty"`
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

// AllPlayersDead reports a defeat: nobody is still fighting, and at least one
// player is actually dead. A fled or slipped player is not dead. A party that
// only escaped returns false so the fight can end as fled instead.
func (c *CombatInstance) AllPlayersDead() bool {
	anyDead := false
	for _, p := range c.Players {
		if p.IsAlive && !p.HasFled {
			return false
		}
		if !p.IsAlive {
			anyDead = true
		}
	}
	return anyDead
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

// AllPlayersFled returns true when every player has fled or slipped and none
// are dead. A mixed party (some dead, some fled) is not a flee.
func (c *CombatInstance) AllPlayersFled() bool {
	if len(c.Players) == 0 {
		return false
	}
	for _, p := range c.Players {
		if !p.IsAlive || !p.HasFled {
			return false
		}
	}
	return true
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
