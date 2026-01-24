# Product Requirements Document: Combat System

**Version:** 1.0
**Date:** 2026-01-23
**Status:** Draft
**Branch:** NPCs

---

## Executive Summary

This document outlines the turn-based combat system for TalesMUD. Combat occurs in isolated **Combat Instances** - temporary rooms where players and enemies fight in a Pokemon-style turn-based format. The system supports multi-enemy encounters, multi-player combat, and persistence across disconnections.

### Key Features

- **Turn-based combat** with initiative-based turn order
- **Combat instances** - isolated temporary rooms for fights
- **Multi-enemy encounters** - fight multiple NPCs simultaneously
- **Multi-player combat** - multiple players can fight together (formal party system later)
- **Flee mechanics** - chance-based escape system
- **Persistence** - players return to combat on reconnect
- **Loot & XP distribution** - rewards on victory

---

## Table of Contents

1. [Current State Analysis](#current-state-analysis)
2. [Architecture Decision](#architecture-decision)
3. [Core Concepts](#core-concepts)
4. [Combat Flow](#combat-flow)
5. [Data Models](#data-models)
6. [Combat Mechanics](#combat-mechanics)
7. [Commands](#commands)
8. [Integration Points](#integration-points)
9. [Implementation Phases](#implementation-phases)
10. [Edge Cases & Error Handling](#edge-cases--error-handling)

---

## Current State Analysis

### What Exists

| Component | Status | Notes |
|-----------|--------|-------|
| `EnemyTrait` | Implemented | Combat stats, aggro, loot config |
| `EnemyTrait.AggroOnSight` | Implemented | Trigger for aggressive enemies |
| `EnemyTrait.AttackPower/Defense` | Implemented | Base damage values |
| `EnemyTrait.XPReward/GoldDrop` | Implemented | Reward configuration |
| `NPC.State` | Implemented | FSM with "combat" state |
| `Character.EquippedItems` | Implemented | Weapon/armor slots |
| `Character.CurrentHitPoints` | Implemented | Health tracking |
| Loot Tables | Implemented | Item drop configuration |
| Item Instance Creation | Implemented | For loot drops |

### What's Missing

| Feature | Priority | Complexity |
|---------|----------|------------|
| Combat Instance (Room) | Critical | High |
| Turn-based combat loop | Critical | High |
| Initiative system | Critical | Medium |
| Damage calculation | Critical | Medium |
| Combat commands | Critical | Medium |
| Combat state management | Critical | High |
| Player death handling | High | Medium |
| Flee mechanics | High | Low |
| Multi-enemy support | High | Medium |
| Party combat support | High | High |
| Combat persistence | High | Medium |
| Combat timeout | Medium | Low |
| Combat log | Medium | Low |

---

## Architecture Decision

### Combat Instance Model

Rather than physically moving players between rooms, we use a **Combat Instance** overlay system:

```
┌─────────────────────────────────────────────────────────────────┐
│                        WORLD ROOM                               │
│  "Dark Forest Clearing"                                         │
│                                                                 │
│  Players: Alice (in combat), Bob                                │
│  NPCs: Goblin Scout (in combat), Merchant Tom                   │
│                                                                 │
│  [Alice and Goblin Scout are in Combat Instance #abc123]        │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ Combat Instance is a parallel state
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    COMBAT INSTANCE #abc123                      │
│  (Isolated combat arena - not a physical room)                  │
│                                                                 │
│  Players: [Alice]                                               │
│  Enemies: [Goblin Scout]                                        │
│                                                                 │
│  Turn Order: Alice (12) → Goblin Scout (8)                      │
│  Current Turn: Alice                                            │
│                                                                 │
│  State: active                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Why Combat Instances?

1. **Isolation** - Other players cannot interfere with ongoing combat
2. **Persistence** - Combat state survives disconnections
3. **Multi-combatant** - Clean handling of multiple players/enemies
4. **Map consistency** - Players remain "in" their room for world purposes
5. **No room teleportation** - Avoids complex room management

### Key Principles

1. **Combat Lock** - NPCs in combat cannot be targeted by non-participants
2. **Player Lock** - Players in combat cannot interact with the world
3. **Turn Exclusivity** - Only current turn holder can act
4. **Instance Cleanup** - Instances are destroyed when combat ends

---

## Core Concepts

### Combat Instance

A combat instance is a temporary, in-memory state machine that manages a fight.

```go
type CombatInstance struct {
    ID              string                 // Unique identifier
    OriginRoomID    string                 // Room where combat started

    // Participants
    Players         []CombatantRef         // Player combatants
    Enemies         []CombatantRef         // NPC combatants

    // Turn Management
    TurnOrder       []CombatantRef         // Initiative-sorted order
    CurrentTurnIdx  int                    // Index into TurnOrder
    TurnStartTime   time.Time              // For timeout tracking
    Round           int                    // Current combat round (starts at 1)

    // State
    State           CombatState            // pending, active, victory, defeat, fled
    CreatedAt       time.Time
    LastActionAt    time.Time

    // Combat Log
    Log             []CombatLogEntry       // History of actions
}

type CombatantRef struct {
    ID              string                 // Character or NPC ID
    Type            CombatantType          // "player" or "npc"
    Initiative      int                    // Rolled initiative value
    IsAlive         bool                   // Still in combat
    HasFled         bool                   // Successfully fled
}

type CombatState string
const (
    CombatStatePending  CombatState = "pending"   // Waiting for combat to begin
    CombatStateActive   CombatState = "active"    // Combat in progress
    CombatStateVictory  CombatState = "victory"   // All enemies defeated
    CombatStateDefeat   CombatState = "defeat"    // All players dead
    CombatStateFled     CombatState = "fled"      // All players fled
    CombatStateTimeout  CombatState = "timeout"   // Combat timed out
)
```

### Combatant States

Players and NPCs track their combat status:

```go
// Added to Character
type Character struct {
    // ... existing fields ...
    InCombat        bool   `json:"inCombat"`
    CombatInstanceID string `json:"combatInstanceId,omitempty"`
}

// Added to NPC
type NPC struct {
    // ... existing fields ...
    InCombat        bool   `json:"inCombat"`
    CombatInstanceID string `json:"combatInstanceId,omitempty"`
}
```

---

## Combat Flow

### 1. Combat Initiation

```
┌─────────────────────────────────────────────────────────────────┐
│                    COMBAT INITIATION                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  TRIGGER A: Player attacks NPC                                  │
│    > attack goblin                                              │
│    → Check: NPC has EnemyTrait? → Create combat instance        │
│                                                                 │
│  TRIGGER B: Player enters room with aggressive NPC              │
│    → Check: NPC.EnemyTrait.AggroOnSight == true?                │
│    → Check: NPC.InCombat == false?                              │
│    → Auto-create combat instance                                │
│                                                                 │
│  TRIGGER C: NPC detects player in aggro radius                  │
│    → Check: Player in NPC.EnemyTrait.AggroRadius rooms?         │
│    → NPC moves toward player and initiates                      │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 2. Instance Creation

When combat is triggered:

1. **Identify all participants**
   - Player(s) who triggered or are in the room
   - NPC(s) with `AggroOnSight` or that were attacked
   - If NPC has `CallForHelp`, add nearby enemies

2. **Validate participants**
   - Skip NPCs already in combat (different instance)
   - Skip players already in combat

3. **Create combat instance**
   - Generate unique ID
   - Record origin room

4. **Roll initiative for all combatants**
   ```
   Initiative = 1d20 + DEX modifier
   ```

5. **Sort turn order** by initiative (highest first)

6. **Mark all participants as in combat**
   - Set `InCombat = true`
   - Set `CombatInstanceID`

7. **Notify participants**
   - Display combat start message
   - Show turn order
   - Show HP status of all combatants

### 3. Combat Round

```
┌─────────────────────────────────────────────────────────────────┐
│                      COMBAT ROUND                               │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Round 1                                                        │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Turn 1: Alice (Player)                                   │   │
│  │   Timer: 60 seconds                                      │   │
│  │   Actions: attack, spell, item, defend, flee             │   │
│  │   > attack goblin                                        │   │
│  │   → Roll hit: 1d20 + STR mod vs enemy Defense            │   │
│  │   → Calculate damage: Weapon damage + STR mod            │   │
│  │   → Apply damage to target                               │   │
│  │   → Check if target dead                                 │   │
│  │   → End turn                                             │   │
│  └─────────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Turn 2: Goblin Scout (NPC)                              │   │
│  │   AI Decision:                                           │   │
│  │   1. If HP < FleeThreshold → attempt flee               │   │
│  │   2. Else → attack weakest player                        │   │
│  │   → Execute action                                       │   │
│  │   → End turn                                             │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  Check: All enemies dead? → VICTORY                             │
│  Check: All players dead? → DEFEAT                              │
│  Check: All players fled? → FLED                                │
│  Else: Start Round 2                                            │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 4. Turn Timer

- Each turn has a **60 second** time limit
- When timer expires:
  - Player: Auto-defend (skip action, gain defense bonus)
  - NPC: N/A (AI acts instantly)
- Warning at 15 seconds remaining
- If player is AFK for 3 consecutive turns: auto-flee attempt

### 5. Combat Resolution

#### Victory (All Enemies Dead)

1. Mark combat state as `victory`
2. Calculate total rewards:
   - Sum XP from all defeated enemies
   - Sum gold drops
   - Roll loot tables for each enemy
3. Distribute rewards (see Loot Distribution)
4. Drop loot items in origin room
5. Clear combat status from all participants
6. Destroy combat instance
7. Notify players of victory and rewards

#### Defeat (All Players Dead)

1. Mark combat state as `defeat`
2. For each dead player:
   - Calculate gold loss (configurable %, default 10%)
   - Set respawn location (player's bound room or world starting room)
   - Restore HP to 50%
   - Teleport to respawn location
3. Clear combat status from NPCs
4. Reset NPC HP (they won)
5. Destroy combat instance
6. Notify players of defeat

#### Bind Points & Respawn

Players can bind to specific rooms (inns, temples, safe houses) using the `/bind` command. Only rooms with `CanBind: true` support this.

**Room Property:**
```go
// Added to Room entity
CanBind bool `json:"canBind"` // If true, players can /bind here
```

**Character Property:**
```go
// Added to Character entity
BoundRoomID string `json:"boundRoomId,omitempty"` // Respawn location
```

**Respawn Priority:**
1. `Character.BoundRoomID` if set and room still exists
2. World starting room (configured in game settings)

**Bind Command:**
```
> bind

You bind your soul to the Sleeping Dragon Inn.
You will respawn here upon death.
```

Error cases:
- "You cannot bind here." (room doesn't support binding)
- "You are already bound to this location."

#### Fled (All Players Escaped)

1. Mark combat state as `fled`
2. For each fled player:
   - Already removed from instance
   - Back in origin room with reduced HP
3. Clear combat status from NPCs
4. Reset NPC HP to pre-combat values
5. Destroy combat instance

---

## Data Models

### Combat Instance Entity

```go
// pkg/entities/combat/instance.go

type CombatInstance struct {
    ID              string            `json:"id"`
    OriginRoomID    string            `json:"originRoomId"`

    // Participants
    Players         []CombatantRef    `json:"players"`
    Enemies         []CombatantRef    `json:"enemies"`

    // Turn Management
    TurnOrder       []CombatantRef    `json:"turnOrder"`
    CurrentTurnIdx  int               `json:"currentTurnIdx"`
    TurnStartTime   time.Time         `json:"turnStartTime"`
    Round           int               `json:"round"`

    // State
    State           CombatState       `json:"state"`
    CreatedAt       time.Time         `json:"createdAt"`
    LastActionAt    time.Time         `json:"lastActionAt"`

    // Configuration
    TurnTimeoutSec  int               `json:"turnTimeoutSec"`  // Default: 60

    // Combat Log
    Log             []CombatLogEntry  `json:"log"`
}

type CombatantRef struct {
    ID              string         `json:"id"`
    Type            CombatantType  `json:"type"`       // "player" or "npc"
    Name            string         `json:"name"`       // Display name
    Initiative      int            `json:"initiative"`
    IsAlive         bool           `json:"isAlive"`
    HasFled         bool           `json:"hasFled"`

    // Snapshot of combat stats at combat start
    MaxHP           int32          `json:"maxHp"`
    CurrentHP       int32          `json:"currentHp"`
    AttackPower     int32          `json:"attackPower"`
    Defense         int32          `json:"defense"`

    // Status effects
    DefenseBonus    int32          `json:"defenseBonus"`    // From defend action
}

type CombatantType string
const (
    CombatantTypePlayer CombatantType = "player"
    CombatantTypeNPC    CombatantType = "npc"
)

type CombatLogEntry struct {
    Timestamp   time.Time      `json:"timestamp"`
    Round       int            `json:"round"`
    ActorID     string         `json:"actorId"`
    ActorName   string         `json:"actorName"`
    Action      CombatAction   `json:"action"`
    TargetID    string         `json:"targetId,omitempty"`
    TargetName  string         `json:"targetName,omitempty"`
    Result      string         `json:"result"`              // "hit", "miss", "critical", "fled", "blocked"
    Damage      int32          `json:"damage,omitempty"`
    Message     string         `json:"message"`             // Human-readable description
}
```

### Combat Action Types

```go
type CombatAction string
const (
    CombatActionAttack   CombatAction = "attack"
    CombatActionSpell    CombatAction = "spell"
    CombatActionItem     CombatAction = "item"
    CombatActionDefend   CombatAction = "defend"
    CombatActionFlee     CombatAction = "flee"
    CombatActionTimeout  CombatAction = "timeout"  // Forced defend
)
```

### Combat Configuration

```go
// Global combat settings
type CombatConfig struct {
    TurnTimeoutSeconds      int     `json:"turnTimeoutSeconds"`      // Default: 60
    AFKAutoFleeAfterTurns   int     `json:"afkAutoFleeAfterTurns"`   // Default: 3
    DeathGoldLossPercent    float64 `json:"deathGoldLossPercent"`    // Default: 0.10 (10%)
    DeathRespawnHPPercent   float64 `json:"deathRespawnHpPercent"`   // Default: 0.50 (50%)
    FleeBaseChance          float64 `json:"fleeBaseChance"`          // Default: 0.50 (50%)
    FleeDexBonus            float64 `json:"fleeDexBonus"`            // Per DEX point bonus
    DefendBonusPercent      float64 `json:"defendBonusPercent"`      // Default: 0.50 (50% defense boost)
    CriticalHitChance       float64 `json:"criticalHitChance"`       // Default: 0.05 (5%)
    CriticalHitMultiplier   float64 `json:"criticalHitMultiplier"`   // Default: 2.0
    CombatTimeoutMinutes    int     `json:"combatTimeoutMinutes"`    // Default: 30
}
```

---

## Combat Mechanics

### Initiative

Roll at combat start, determines turn order for entire combat:

```
Initiative = 1d20 + (DEX / 4)

For NPCs: Use NPC.Level as DEX approximation, or add DEX to NPC model
```

Turn order is fixed for the duration of combat (re-roll only if new combatants join, which doesn't happen in this design).

### Attack Resolution

**Phase 1: Melee Only**

```
To Hit Roll:
  Roll = 1d20
  If Roll == 20: Critical Hit (auto-hit, double damage)
  If Roll == 1: Critical Miss (auto-miss)
  Hit = Roll + STR modifier >= Target.Defense + 10

Damage Calculation:
  Base Damage = Weapon.Damage + STR modifier

  For Players:
    Weapon.Damage = EquippedWeapon.Attributes["damage"] OR 1 (unarmed)

  For NPCs:
    Base Damage = EnemyTrait.AttackPower

  Defense Reduction:
    Reduction = Target.Defense / 2  (armor absorbs some damage)

  Final Damage = max(1, Base Damage - Reduction)

  If Critical: Final Damage *= 2
```

**Future Phases:**
- **Ranged Weapons (Bows):** DEX-based to-hit and damage
- **Spells:** INT/WIS-based, mana cost, spell effects

### Attribute Modifiers

```
Modifier = (Attribute - 10) / 2

Example:
  STR 14 → +2 modifier
  STR 8  → -1 modifier
  STR 10 → 0 modifier
```

### Defend Action

When a combatant defends:
- Skip offensive action
- Gain +50% defense bonus until next turn
- Defense bonus is applied to `CombatantRef.DefenseBonus`
- Cleared at start of combatant's next turn

### Flee Mechanics

```
Flee Chance = BaseChance + (DEX modifier * FleeDexBonus)

Default:
  BaseChance = 50%
  FleeDexBonus = 2% per DEX point above 10

Example:
  DEX 14 (+2 modifier): 50% + (2 * 2%) = 54% chance to flee

On Success:
  - Remove player from combat instance
  - Clear player's combat status
  - Player remains in origin room
  - Notify remaining participants
  - If last player fled → combat ends with "fled" state

On Failure:
  - Player loses their turn
  - Message: "You fail to escape! The enemies block your path."
  - Combat continues
```

### NPC AI Decision Making

Simple priority-based AI:

```
1. If HP <= FleeThreshold AND FleeThreshold > 0:
   → Attempt flee

2. If any ally NPC has "healer" CombatStyle AND self HP < 50%:
   → Request heal (future feature)

3. Attack player with lowest HP percentage
   → Basic attack using AttackPower
```

### Damage to Players

When a player's HP reaches 0:
- Mark as dead (`IsAlive = false`)
- Remove from turn order
- Check if all players dead → Defeat

### Damage to NPCs

When an NPC's HP reaches 0:
- Mark as dead (`IsAlive = false`)
- Remove from turn order
- Execute `OnDeathScript` if defined
- Check if all enemies dead → Victory

---

## Loot Distribution

When combat ends in victory:

### XP Distribution

```
Total XP = Sum of all enemy XPReward values
Per Player XP = Total XP / Number of living players at victory

XP is NOT split if player died during combat but others won.
Dead players get 0 XP.
```

### Gold Distribution

```
Total Gold = Sum of all enemy GoldDrop rolls
Per Player Gold = Total Gold / Number of living players at victory

Same rules as XP - dead players get nothing.
```

### Item Loot

1. Roll loot for each defeated enemy (using existing loot table system)
2. All items drop to the **origin room floor**
3. Players can pick up items after combat ends
4. First-come-first-serve (free for all)

Future enhancement: Need/Greed/Pass rolling system for group loot.

---

## Commands

### Combat Initiation

| Command | Syntax | Description |
|---------|--------|-------------|
| `attack` | `attack <target>` | Initiate combat with target NPC |

### Respawn System

| Command | Syntax | Description |
|---------|--------|-------------|
| `bind` | `bind` | Bind to current room (if room allows) |

### In-Combat Commands

| Command | Aliases | Syntax | Description | Phase |
|---------|---------|--------|-------------|-------|
| `attack` | `a`, `hit` | `attack <target>` | Melee attack target enemy | 1 |
| `defend` | `d`, `guard` | `defend` | Defend for damage reduction | 3 |
| `flee` | `run`, `escape` | `flee` | Attempt to escape combat | 3 |
| `item` | `use` | `item <item>` | Use an item (potion, etc.) | 3 |
| `status` | `combat`, `cs` | `status` | View combat status | 3 |
| `log` | `combatlog` | `log [lines]` | View combat log | 7 |
| `shoot` | `fire` | `shoot <target>` | Ranged attack (bow) | Future |
| `spell` | `cast` | `spell <spell> [target]` | Cast a spell | Future |

### Command Details

#### `attack <target>`

```
> attack goblin

You swing your Iron Sword at Goblin Scout!
Roll: 15 + 2 (STR) = 17 vs Defense 12 - HIT!
Damage: 8 + 2 = 10, reduced by 2 armor = 8 damage

Goblin Scout takes 8 damage! (12/20 HP remaining)
```

#### `defend`

```
> defend

You raise your guard, preparing for incoming attacks.
Defense increased by 50% until your next turn.

[Your turn ends]
```

#### `flee`

```
> flee

You attempt to escape from combat...
Flee roll: 47 vs 54% chance - SUCCESS!

You escape from combat!

[You are removed from the fight]
```

Or on failure:

```
> flee

You attempt to escape from combat...
Flee roll: 62 vs 54% chance - FAILED!

The enemies block your escape! You lose your turn.
```

#### `status`

```
> status

═══════════════════════════════════════════════════
                 COMBAT STATUS - Round 3
═══════════════════════════════════════════════════

YOUR PARTY:
  ► Alice (YOU)        ████████░░ 42/50 HP
    Bob                ███░░░░░░░ 15/50 HP

ENEMIES:
    Goblin Scout       ██████░░░░ 12/20 HP
    Goblin Warrior     ████████░░ 35/45 HP

TURN ORDER:
  1. Alice (12) ← CURRENT TURN [45s remaining]
  2. Goblin Warrior (10)
  3. Bob (9)
  4. Goblin Scout (6)

ACTIONS: attack | defend | item | flee | status
═══════════════════════════════════════════════════
```

---

## Integration Points

### With NPC System

- **Combat trigger**: Check `EnemyTrait.AggroOnSight` on room entry
- **Aggro radius**: NPC update loop checks `AggroRadius` for nearby players
- **Combat stats**: Use `EnemyTrait.AttackPower`, `Defense`
- **Flee behavior**: Check `EnemyTrait.FleeThreshold`
- **Call for help**: If `EnemyTrait.CallForHelp`, pull nearby enemies
- **Scripts**: Execute `OnAggroScript`, `OnDeathScript`

### With Item System

- **Weapon damage**: Read from `EquippedItems[MainHand].Attributes["damage"]`
- **Armor defense**: Sum defense from all equipped armor pieces
- **Consumables**: `item` command uses items during combat
- **Loot drops**: Use existing loot table system on victory

### With Character System

- **HP tracking**: Use `Character.CurrentHitPoints`, `MaxHitPoints`
- **Attributes**: Use STR, DEX, CON for modifiers
- **Death handling**: Set HP, handle respawn location
- **Gold management**: Deduct gold on death

### With Room System

- **Combat in room**: Players and NPCs remain "in" the origin room
- **Loot placement**: Items dropped to room's `Items` array
- **Room visibility**: Other players see "X is in combat" status
- **Bind points**: Rooms with `CanBind: true` allow `/bind` for respawn

### Multi-Player Combat (Without Formal Party System)

When combat is initiated, **all players currently in the room** may be pulled into combat:

1. **Aggressive NPC attacks**: All players in room are targeted
2. **Player attacks NPC**: Only the attacking player enters combat initially
3. **Other players in room**: Can choose to `attack` the same NPC to join the fight
4. **Join window**: Players can only join within Round 1 of combat

This allows cooperative play without requiring a formal party invite system. Party system (invite, leave, party chat, shared XP settings) is a future enhancement.

### With Scripting System

New Lua functions:

```lua
-- Combat queries
tales.combat.isInCombat(characterId)
tales.combat.getCombatInstance(characterId)
tales.combat.getCombatants(instanceId)

-- Combat manipulation (for scripts)
tales.combat.dealDamage(instanceId, targetId, amount)
tales.combat.healTarget(instanceId, targetId, amount)
tales.combat.addCombatant(instanceId, npcId)
tales.combat.endCombat(instanceId, result)  -- "victory", "defeat", "draw"

-- Events
-- combat.start: {instance, players, enemies, room}
-- combat.turn.start: {instance, combatant, round}
-- combat.turn.end: {instance, combatant, action, result}
-- combat.damage: {instance, attacker, target, damage}
-- combat.death: {instance, victim, killer}
-- combat.end: {instance, result, survivors, loot}
```

---

## Implementation Phases

### Phase 1: Core Combat Engine (Critical)

**Scope:** ~15 files

**Note:** Phase 1 supports **melee weapons only**. Ranged weapons (bows) and spells are added in later phases.

1. **Combat Instance Model**
   - Create `pkg/entities/combat/instance.go`
   - Define all combat types and structs
   - Create in-memory combat instance manager

2. **Combat Service**
   - Create `pkg/service/combat.go`
   - Instance lifecycle (create, update, destroy)
   - Turn management logic
   - Combat state machine

3. **Initiative System**
   - Roll initiative on combat start
   - Sort turn order
   - Track current turn

4. **Basic Melee Attack Command**
   - Implement `attack` command for combat initiation
   - Implement `attack` for in-combat targeting
   - Melee damage calculation (STR-based)
   - Unarmed combat fallback (1 base damage)

### Phase 2: Combat Resolution (Critical)

**Scope:** ~8 files

1. **Damage Calculation**
   - Weapon damage extraction
   - Armor reduction
   - Critical hits
   - Attribute modifiers

2. **HP Management**
   - Track damage in combat instance
   - Sync to Character/NPC on combat end
   - Death detection

3. **Victory/Defeat Handling**
   - Victory: Loot distribution, XP rewards
   - Defeat: Gold loss, respawn
   - Combat instance cleanup

4. **Turn Timer**
   - 60-second turn timeout
   - Auto-defend on timeout
   - AFK detection

### Phase 3: Combat Actions (High Priority)

**Scope:** ~6 files

1. **Defend Command**
   - Defense bonus application
   - Bonus clearing on next turn

2. **Flee Command**
   - Flee chance calculation
   - Success/failure handling
   - Combat exit logic

3. **Item Command (Basic)**
   - Use consumable items
   - Health potion support
   - Inventory integration

4. **Status Command**
   - Combat status display
   - Turn order visualization
   - HP bars

### Phase 4: Multi-Combatant Support (High Priority)

**Scope:** ~5 files

1. **Multi-Enemy Encounters**
   - Multiple NPCs in combat
   - Target selection (`attack goblin 2` or `attack [2]`)
   - Aggro grouping (CallForHelp)

2. **Multi-Player Combat**
   - Multiple players in same combat instance
   - Players in room when combat starts are included
   - Turn order with multiple players
   - Note: Formal party system (invite, party chat) is a separate feature

3. **Loot/XP Distribution**
   - Split rewards among living players at victory
   - Dead player exclusion (0 XP/gold if dead at end)
   - Room loot drops (free-for-all pickup)

### Phase 5: NPC AI (Medium Priority)

**Scope:** ~3 files

1. **AI Decision Engine**
   - Priority-based action selection
   - Target selection logic
   - Flee threshold behavior

2. **Combat Styles**
   - Different behaviors for melee/ranged/magic
   - Aggro management (future)

### Phase 6: Combat Triggers (Medium Priority)

**Scope:** ~4 files

1. **AggroOnSight**
   - Room entry trigger
   - Auto-combat initiation

2. **Aggro Radius**
   - NPC update loop integration
   - Proximity detection
   - Movement toward player

3. **Attack Command (Non-Combat)**
   - Initiate combat from world
   - Validate target is enemy

### Phase 7: Persistence & Polish (Medium Priority)

**Scope:** ~4 files

1. **Combat Persistence**
   - Store active combat instances
   - Restore on player reconnect
   - Handle server restart (optional)

2. **Combat Log**
   - Detailed action logging
   - Log command implementation

3. **Lua Integration**
   - Combat events
   - Scripting functions

---

## Edge Cases & Error Handling

### Player Disconnection

- Combat instance continues (NPCs still act)
- Player's turn is auto-defended
- After 3 auto-defends, player auto-flees
- On reconnect, player rejoins instance if still active

### Server Restart

- Combat instances are in-memory only (MVP)
- On restart, all combat ends:
  - Players return to origin room
  - NPCs reset to pre-combat state
  - No loot distributed
- Future: Persist combat state to database

### Invalid Targets

- Cannot attack NPC without EnemyTrait
- Cannot attack NPC already in different combat
- Cannot attack player (PvP not in scope)
- Clear error messages for each case

### Empty Combat

- If all players flee, combat ends with "fled" state
- If all enemies somehow die outside combat, check and end
- Validate at least 1 player and 1 enemy before starting

### Timeout

- Combat has 30-minute global timeout
- If reached, combat ends in "timeout" state
- All participants returned to normal state
- NPCs reset, no loot/XP distributed

### Command Validation

- Non-combat commands rejected during combat (most)
- Exception: `say`, `party`, `who`, `quit`
- Clear message: "You're in combat! Available actions: attack, defend, item, flee"

---

## UI/UX Considerations

### Combat Start Message

```
═══════════════════════════════════════════════════
              COMBAT INITIATED!
═══════════════════════════════════════════════════

A Goblin Scout attacks you!

YOUR PARTY:
  Alice (50/50 HP)

ENEMIES:
  Goblin Scout (20/20 HP)

Rolling initiative...
  Alice: 14
  Goblin Scout: 8

Turn order: Alice → Goblin Scout

═══════════════════════════════════════════════════
It's your turn! What do you do?
Actions: attack | defend | item | flee | status
═══════════════════════════════════════════════════
```

### Turn Prompt

```
───────────────────────────────────────────────────
Your turn! (Round 2) [55s remaining]
HP: 35/50  |  Enemies: Goblin Scout (8/20 HP)
───────────────────────────────────────────────────
>
```

### Victory Message

```
═══════════════════════════════════════════════════
              VICTORY!
═══════════════════════════════════════════════════

You defeated: Goblin Scout

REWARDS:
  Experience: 25 XP
  Gold: 12 gold pieces

LOOT DROPPED:
  - Rusty Dagger
  - Minor Health Potion

The items have been dropped on the ground.
Use 'pickup <item>' to collect them.
═══════════════════════════════════════════════════
```

### Defeat Message

```
═══════════════════════════════════════════════════
              DEFEAT
═══════════════════════════════════════════════════

You have been slain by Goblin Scout.

PENALTY:
  Gold lost: 15 gold pieces (10%)

You awaken at the Town Square with 25/50 HP.
═══════════════════════════════════════════════════
```

---

## Success Criteria

### MVP (Melee Combat)
- [ ] Players can initiate combat with `attack <npc>` command
- [ ] Aggressive NPCs auto-initiate combat on player room entry
- [ ] Combat follows turn-based order based on initiative
- [ ] Players can attack (melee), defend, use items, and flee
- [ ] NPCs execute basic AI combat actions
- [ ] Combat ends when all enemies or all players are defeated
- [ ] Victory grants XP, gold, and loot drops
- [ ] Defeat causes gold loss and respawn at bound location
- [ ] Players can `/bind` at designated rooms (inns, temples)
- [ ] Multiple enemies can be fought simultaneously
- [ ] Multiple players can join the same combat
- [ ] Players return to combat on reconnect
- [ ] 60-second turn timer enforced
- [ ] Combat status visible to participants

### Post-MVP
- [ ] Ranged weapons (bows) with DEX-based attacks
- [ ] Spell system with mana
- [ ] Formal party system
- [ ] Lua scripting for combat events

---

## Future Enhancements (Out of Scope for MVP)

### Near-Term (Post-MVP)
1. **Ranged Weapons (Bows)** - DEX-based attacks, ammo system
2. **Spell System** - Magic attacks with mana costs, spell slots
3. **Party System** - Formal party invite/join, party chat, shared XP options

### Long-Term
4. **Status Effects** - Poison, stun, bleed, buffs
5. **PvP Combat** - Player vs player duels
6. **Boss Mechanics** - Special abilities, phases
7. **Combat Formations** - Front/back row positioning
8. **Aggro/Threat System** - Tank mechanics
9. **Need/Greed Loot** - Loot rolling system
10. **Combat Achievements** - Kill tracking, titles
11. **Arena System** - Structured PvP/PvE challenges
12. **Combo System** - Chained attacks for bonus damage

---

## Appendix: Example Combat Flow

### Scenario: Alice attacks a Goblin

```
1. Alice types: attack goblin
2. System finds Goblin Scout in room with EnemyTrait
3. Combat instance created:
   - ID: "combat-abc123"
   - OriginRoomID: "forest-clearing-01"
   - Players: [Alice]
   - Enemies: [Goblin Scout]

4. Initiative rolled:
   - Alice: 1d20(12) + DEX mod(+2) = 14
   - Goblin: 1d20(6) + 2 = 8

5. Turn order: Alice (14) → Goblin (8)

6. Round 1, Turn 1 (Alice):
   - Alice types: attack goblin
   - Hit roll: 1d20(15) + STR(+2) = 17 vs Defense(12) = HIT
   - Damage: Sword(6) + STR(+2) = 8 - Armor(2) = 6 damage
   - Goblin HP: 20 → 14

7. Round 1, Turn 2 (Goblin):
   - AI checks: HP(14) > FleeThreshold(20%) = true, won't flee
   - AI attacks Alice
   - Hit roll: 1d20(11) + 3 = 14 vs Defense(14) = HIT
   - Damage: 5 - Armor(3) = 2 damage
   - Alice HP: 50 → 48

8. Round 2, Turn 1 (Alice):
   - Alice types: attack goblin
   - Critical hit! Roll: 20
   - Damage: (6 + 2) * 2 = 16 - 2 = 14 damage
   - Goblin HP: 14 → 0

9. Goblin dies:
   - Execute OnDeathScript (if any)
   - Remove from turn order
   - All enemies dead → VICTORY

10. Combat resolution:
    - XP awarded: 25 to Alice
    - Gold rolled: 8 gold to Alice
    - Loot rolled: Rusty Dagger dropped to room
    - Combat instance destroyed
    - Alice.InCombat = false
```

---

## Appendix: Data Model Summary

### New Files to Create

| File | Purpose |
|------|---------|
| `pkg/entities/combat/instance.go` | Combat instance entity |
| `pkg/entities/combat/types.go` | Combat enums and types |
| `pkg/mudserver/game/combat/manager.go` | In-memory combat instance manager |
| `pkg/mudserver/game/combat/engine.go` | Combat logic engine |
| `pkg/mudserver/game/combat/ai.go` | NPC combat AI |
| `pkg/mudserver/game/commands/combat.go` | Combat commands |
| `pkg/service/combat.go` | Combat service layer |

### Modified Files

| File | Changes |
|------|---------|
| `pkg/entities/characters/character.go` | Add `InCombat`, `CombatInstanceID`, `BoundRoomID` |
| `pkg/entities/npcs/npc.go` | Add `InCombat`, `CombatInstanceID` |
| `pkg/entities/rooms/rooms.go` | Add `CanBind` property |
| `pkg/mudserver/game/game.go` | Add combat manager, combat tick |
| `pkg/mudserver/game/commands/commandprocessor.go` | Combat command routing |
| `pkg/mudserver/game/commands/bind.go` | New bind command |
| `pkg/scripts/runner/lua/modules/combat.go` | New Lua module |

---

*Document version 1.0 - Initial draft based on design session*
