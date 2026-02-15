# PRD: Enhanced NPC System

**Version:** 1.0
**Date:** 2026-01-21
**Status:** Implemented
**Branch:** NPCs

---

## 1. Executive Summary

This document outlines the requirements for enhancing the NPC (Non-Player Character) system in TalesMUD. The goal is to transform the current basic NPC implementation into a robust system supporting three distinct NPC archetypes: **Standard NPCs**, **Enemies**, and **Merchants** - all built on a unified base with specialized traits.

Key additions include:
- NPC Templates for spawnable/instanced NPCs
- Enhanced Enemy system with combat integration
- Merchant system with inventory and trading
- Configurable behaviors (movement, aggro, respawn)

---

## 2. Current State Analysis

### What Exists
- Basic NPC entity with name, description, race, class, level, HP
- Minimal `EnemyTrait` and `MerchantTrait` (type string only)
- Dialog system integration (`DialogID`, `IdleDialogID`)
- SQLite repository with CRUD operations
- Lua scripting module (`tales.npcs.*`)
- Empty game loop hooks (`updateNPC()`, `handleNPCUpdates()`)

### Key Gaps
| Area | Gap |
|------|-----|
| Templates | No template system for spawnable NPCs |
| Instancing | No way to spawn multiple instances from a single definition |
| Enemies | No combat stats, loot tables, aggro behavior |
| Merchants | No inventory, pricing, buy/sell mechanics |
| AI/Behavior | Game loop hooks exist but are empty |
| Movement | No patrol or wander behavior |

---

## 3. Architecture Decision

### Recommendation: Unified Base + Trait Composition

After analyzing the codebase, the recommended approach is to **keep NPC as a unified base** with optional trait composition rather than splitting into separate entity types.

**Rationale:**
1. **Existing Pattern** - The codebase already uses this pattern (`EnemyTrait`, `MerchantTrait`)
2. **Flexibility** - NPCs can combine traits (e.g., an enemy that also sells items when defeated)
3. **Shared Behavior** - Movement, dialog, room presence are common to all types
4. **Simpler Storage** - Single table, single repository, consistent API
5. **Scripting Consistency** - One `tales.npcs.*` module handles all types

**Trade-off:** Slightly more complex NPC struct vs. multiple simpler entity types

### Template vs Instance Model

```
┌─────────────────────────────────────────────────────────────────┐
│                        NPC Template                              │
│  (Definition - stored in DB, used as blueprint)                 │
│  - IsTemplate: true                                              │
│  - Defines base stats, traits, behaviors                        │
│  - Not placed in rooms directly                                  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ Spawner creates instances
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     NPC Instance                                 │
│  (Runtime copy - exists in game world)                          │
│  - IsTemplate: false                                             │
│  - TemplateID: references source template                       │
│  - InstanceID: unique suffix (e.g., "Rat-abc123")               │
│  - Has current HP, position, state                              │
│  - Can be killed, respawned, etc.                               │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    Singleton NPC                                 │
│  (Unique NPC - only one in world)                               │
│  - IsTemplate: false                                             │
│  - TemplateID: empty                                             │
│  - Unique quest givers, merchants, bosses                       │
└─────────────────────────────────────────────────────────────────┘
```

---

## 4. Requirements

### 4.1 NPC Base Entity Enhancements

#### 4.1.1 Template Support
| Field | Type | Description |
|-------|------|-------------|
| `IsTemplate` | bool | If true, this NPC is a template (blueprint) |
| `TemplateID` | string | For instances: ID of source template |
| `InstanceSuffix` | string | Unique suffix for instanced NPCs (e.g., "abc123") |

**Display Name Logic:**
- Template: `"Rat"` (stored name)
- Instance: `"Rat"` (display) but internally `"Rat-abc123"` for targeting

#### 4.1.2 Behavior Configuration
| Field | Type | Description |
|-------|------|-------------|
| `SpawnRoomID` | string | Room where NPC spawns/respawns |
| `RespawnTime` | duration | Time to respawn after death (0 = no respawn) |
| `WanderRadius` | int | Rooms away from spawn it can wander (0 = stationary) |
| `PatrolPath` | []string | Ordered list of room IDs for patrol route |

#### 4.1.3 State Tracking
| Field | Type | Description |
|-------|------|-------------|
| `IsDead` | bool | Currently dead (awaiting respawn) |
| `DeathTime` | time.Time | When NPC died (for respawn timing) |
| `State` | string | FSM state: "idle", "combat", "patrol", "dead" |

---

### 4.2 Enhanced Enemy Trait

Replace the minimal `EnemyTrait` with a comprehensive combat configuration.

```go
type EnemyTrait struct {
    // Classification
    CreatureType  string   `json:"creatureType"`  // "beast", "humanoid", "undead", "elemental", "construct", "demon", "dragon", "aberration"
    CombatStyle   string   `json:"combatStyle"`   // "melee", "ranged", "magic", "swarm", "brute", "agile"
    Difficulty    string   `json:"difficulty"`    // "trivial", "easy", "normal", "hard", "boss"

    // Combat Stats
    AttackPower   int32    `json:"attackPower"`   // Base damage
    Defense       int32    `json:"defense"`       // Damage reduction
    AttackSpeed   float64  `json:"attackSpeed"`   // Attacks per second

    // Behavior
    AggroRadius   int      `json:"aggroRadius"`   // Rooms away to detect players (0 = passive)
    AggroOnSight  bool     `json:"aggroOnSight"`  // Auto-attack on sight
    CallForHelp   bool     `json:"callForHelp"`   // Alert nearby enemies when attacked
    FleeThreshold float64  `json:"fleeThreshold"` // HP % to flee (0 = never)

    // Rewards
    XPReward      int64    `json:"xpReward"`      // XP granted on kill
    GoldDrop      Range    `json:"goldDrop"`      // Min/max gold drop
    LootTableID   string   `json:"lootTableID"`   // Reference to loot table

    // Scripts
    OnAggroScript   string `json:"onAggroScript"`   // Script when entering combat
    OnDeathScript   string `json:"onDeathScript"`   // Script when killed
    OnFleeScript    string `json:"onFleeScript"`    // Script when fleeing
}

type Range struct {
    Min int32 `json:"min"`
    Max int32 `json:"max"`
}
```

---

### 4.3 Enhanced Merchant Trait

Replace the minimal `MerchantTrait` with trading functionality.

```go
type MerchantTrait struct {
    // Classification
    MerchantType    string   `json:"merchantType"`    // "general", "weapons", "armor", "magic"

    // Inventory
    Inventory       []MerchantItem `json:"inventory"` // Items for sale
    RestockTime     duration `json:"restockTime"`     // Time to restock (0 = infinite stock)

    // Pricing
    BuyPriceModifier  float64 `json:"buyPriceModifier"`  // Multiplier for buying from players (e.g., 0.5)
    SellPriceModifier float64 `json:"sellPriceModifier"` // Multiplier for selling to players (e.g., 1.0)

    // Restrictions
    RequiredReputation int32  `json:"requiredReputation"` // Min reputation to trade
    AcceptedItemTypes []string `json:"acceptedItemTypes"` // Item types merchant will buy

    // Currency
    Currency        string   `json:"currency"`        // "gold", "tokens", custom currency

    // Scripts
    OnTradeScript   string   `json:"onTradeScript"`   // Script after successful trade
}

type MerchantItem struct {
    ItemTemplateID string `json:"itemTemplateId"` // Reference to item template
    Stock          int32  `json:"stock"`          // Current stock (-1 = unlimited)
    MaxStock       int32  `json:"maxStock"`       // Max stock after restock
    PriceOverride  int32  `json:"priceOverride"`  // Custom price (0 = use item base price)
}
```

---

### 4.4 NPC Spawner System

A mechanism to spawn NPC instances from templates.

#### Spawner Entity
```go
type NPCSpawner struct {
    ID            string        `json:"id"`
    TemplateID    string        `json:"templateId"`    // NPC template to spawn
    RoomID        string        `json:"roomId"`        // Where to spawn
    MaxInstances  int           `json:"maxInstances"`  // Max alive at once
    SpawnInterval time.Duration `json:"spawnInterval"` // Time between spawns
    InitialCount  int           `json:"initialCount"`  // Spawn on world load

    // Runtime state
    ActiveInstances []string    `json:"activeInstances"` // IDs of current instances
    LastSpawnTime   time.Time   `json:"lastSpawnTime"`
}
```

#### Spawning Rules
1. Spawner checks `MaxInstances` before spawning
2. Instance gets unique suffix: `fmt.Sprintf("%s-%s", template.Name, shortUUID())`
3. Instance copies template data but has own HP, state
4. On instance death: remove from `ActiveInstances`, start respawn timer
5. Respawn creates new instance with fresh stats

---

### 4.5 Lua Scripting Extensions

#### New NPC Functions
```lua
-- Template/Instance management
tales.npcs.spawnFromTemplate(templateId, roomId)  -- Create instance
tales.npcs.isTemplate(id)                          -- Check if template
tales.npcs.getTemplateId(id)                       -- Get source template
tales.npcs.getInstancesOfTemplate(templateId)      -- Get all instances

-- Enemy functions
tales.npcs.getAggroTargets(id)       -- Get characters in aggro range
tales.npcs.attack(npcId, targetId)   -- Initiate attack
tales.npcs.flee(id)                  -- Make NPC flee
tales.npcs.dropLoot(id, roomId)      -- Spawn loot in room

-- Merchant functions
tales.npcs.getInventory(id)                    -- Get merchant inventory
tales.npcs.buyFromPlayer(npcId, charId, itemId) -- NPC buys item
tales.npcs.sellToPlayer(npcId, charId, itemId)  -- NPC sells item
tales.npcs.restockMerchant(id)                  -- Force restock

-- Behavior
tales.npcs.setState(id, state)       -- Set FSM state
tales.npcs.getState(id)              -- Get current state
tales.npcs.patrol(id)                -- Start patrol behavior
tales.npcs.wander(id)                -- Random movement within radius
```

---

### 4.6 Game Loop Integration

#### NPC Update Cycle (per tick)
```
For each NPC:
  1. Skip if IsTemplate == true
  2. Check respawn timer if dead
  3. Based on State:
     - "idle": Check aggro radius, run idle behaviors
     - "patrol": Move along patrol path
     - "combat": Execute combat logic
     - "fleeing": Move away from threat
  4. Check idle dialog timer
  5. Execute periodic scripts if defined
```

#### Spawner Update Cycle
```
For each Spawner:
  1. Remove dead instances from ActiveInstances
  2. If len(ActiveInstances) < MaxInstances:
     - If SpawnInterval elapsed: spawn new instance
```

---

### 4.7 Player Commands

| Command | Description |
|---------|-------------|
| `look <npc>` | View NPC description |
| `talk <npc>` | Start dialog (if has DialogID) |
| `attack <npc>` | Initiate combat (if enemy) |
| `trade <npc>` / `shop <npc>` | Open trade UI (if merchant) |
| `buy <item> from <npc>` | Purchase item |
| `sell <item> to <npc>` | Sell item |

---

### 4.8 Admin/Builder Commands

| Command | Description |
|---------|-------------|
| `/npc create <name>` | Create new NPC |
| `/npc spawn <templateId>` | Spawn instance from template |
| `/npc kill <npc>` | Kill NPC |
| `/npc respawn <npc>` | Force respawn |
| `/npc teleport <npc> <room>` | Move NPC to room |
| `/spawner create <templateId>` | Create spawner in current room |

---

## 5. Data Model Changes

### 5.1 Updated NPC Struct

```go
type NPC struct {
    *entities.Entity
    traits.BelongsUser
    traits.CurrentRoom

    // Identity
    Name        string `json:"name"`
    Description string `json:"description"`

    // Template System
    IsTemplate     bool   `json:"isTemplate"`
    TemplateID     string `json:"templateId,omitempty"`
    InstanceSuffix string `json:"instanceSuffix,omitempty"`

    // Stats
    Race             enums.Race  `json:"race"`
    Class            enums.Class `json:"class"`
    Level            int32       `json:"level"`
    CurrentHitPoints int32       `json:"currentHitPoints"`
    MaxHitPoints     int32       `json:"maxHitPoints"`

    // Behavior Configuration
    SpawnRoomID  string        `json:"spawnRoomId,omitempty"`
    RespawnTime  time.Duration `json:"respawnTime,omitempty"`
    WanderRadius int           `json:"wanderRadius,omitempty"`
    PatrolPath   []string      `json:"patrolPath,omitempty"`

    // State
    IsDead    bool      `json:"isDead"`
    DeathTime time.Time `json:"deathTime,omitempty"`
    State     string    `json:"state"` // "idle", "combat", "patrol", "dead"

    // Dialog
    DialogID          string        `json:"dialogId,omitempty"`
    IdleDialogID      string        `json:"idleDialogId,omitempty"`
    IdleDialogTimeout time.Duration `json:"idleDialogTimeout,omitempty"`

    // Behavior Traits (optional)
    EnemyTrait    *EnemyTrait    `json:"enemyTrait,omitempty"`
    MerchantTrait *MerchantTrait `json:"merchantTrait,omitempty"`

    // Metadata
    Created time.Time `json:"created"`
}
```

### 5.2 New Database Table

```sql
CREATE TABLE IF NOT EXISTS npc_spawners (
    id TEXT PRIMARY KEY,
    data TEXT NOT NULL
);
```

---

## 6. API Changes

### 6.1 New Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/npcs/:id/spawn` | Spawn instance from template |
| GET | `/api/npcs/templates` | List all NPC templates |
| GET | `/api/npcs/instances` | List all NPC instances |
| GET | `/api/npcs/instances/:templateId` | List instances of specific template |
| POST | `/api/spawners` | Create spawner |
| GET | `/api/spawners` | List spawners |
| PUT | `/api/spawners/:id` | Update spawner |
| DELETE | `/api/spawners/:id` | Delete spawner |

### 6.2 Updated Endpoints

| Method | Endpoint | Change |
|--------|----------|--------|
| GET | `/api/npcs` | Add `?isTemplate=true/false` filter |
| POST | `/api/npcs` | Accept enhanced trait structures |

---

## 7. Implementation Phases

### Phase 1: Foundation (Template System)
1. Update NPC struct with template fields
2. Add `IsTemplate`, `TemplateID`, `InstanceSuffix`
3. Implement `spawnFromTemplate` in service layer
4. Add Lua function `tales.npcs.spawnFromTemplate()`
5. Update API with template filtering

### Phase 2: Enhanced Enemy Trait
1. Expand `EnemyTrait` struct
2. Add combat stats, aggro settings, loot config
3. Implement basic combat loop in game tick
4. Add Lua combat functions
5. Implement loot drop system

### Phase 3: Enhanced Merchant Trait
1. Expand `MerchantTrait` struct
2. Implement merchant inventory system
3. Add buy/sell transaction logic
4. Add Lua merchant functions
5. Implement restock timer

### Phase 4: Spawner System
1. Create `NPCSpawner` entity
2. Add spawner repository and service
3. Implement spawn/despawn logic in game loop
4. Add spawner API endpoints
5. Add Lua spawner functions

### Phase 5: Behavior System
1. Implement NPC state machine
2. Add patrol behavior
3. Add wander behavior
4. Implement aggro detection
5. Add flee behavior

---

## 8. Success Criteria

- [ ] Can create NPC templates and spawn multiple instances
- [ ] Instances have unique identifiers and independent state
- [ ] Enemies can aggro, attack, die, drop loot, and respawn
- [ ] Merchants can buy/sell items with proper pricing
- [ ] Spawners maintain correct instance counts
- [ ] NPCs can patrol or wander between rooms
- [ ] All functionality accessible via Lua scripting
- [ ] Backward compatible with existing NPC data

---

## 9. Decisions

1. **Combat System Scope:** Separate PRD - this PRD covers NPC structure, templates, and spawning only
2. **Merchant System:** Separate PRD - excluded from initial implementation
3. **Loot Tables:** Embedded in EnemyTrait (not separate entity)
4. **Instance Persistence:** Memory only - instances are recreated on server restart
5. **Currency:** Gold only

---

## 10. Appendix: Example NPCs

### Rat Enemy Template
```json
{
  "name": "Rat",
  "description": "A mangy sewer rat with beady red eyes.",
  "isTemplate": true,
  "level": 1,
  "maxHitPoints": 15,
  "currentHitPoints": 15,
  "enemyTrait": {
    "creatureType": "beast",
    "combatStyle": "swarm",
    "difficulty": "trivial",
    "attackPower": 3,
    "defense": 0,
    "attackSpeed": 1.5,
    "aggroRadius": 0,
    "aggroOnSight": false,
    "xpReward": 10,
    "goldDrop": {"min": 0, "max": 2}
  },
  "respawnTime": "30s",
  "wanderRadius": 1
}
```

### Town Merchant (Singleton)
```json
{
  "name": "Gareth the Blacksmith",
  "description": "A burly man with soot-stained arms and a friendly smile.",
  "isTemplate": false,
  "level": 10,
  "dialogId": "blacksmith-dialog",
  "merchantTrait": {
    "merchantType": "weapons",
    "buyPriceModifier": 0.4,
    "sellPriceModifier": 1.0,
    "inventory": [
      {"itemTemplateId": "iron-sword", "stock": 5, "maxStock": 5},
      {"itemTemplateId": "iron-shield", "stock": 3, "maxStock": 3}
    ],
    "restockTime": "1h",
    "acceptedItemTypes": ["weapon", "armor"]
  }
}
```

### Goblin with Dialog
```json
{
  "name": "Goblin Scout",
  "description": "A sneaky goblin carrying a crude dagger.",
  "isTemplate": true,
  "level": 3,
  "maxHitPoints": 35,
  "race": {"id": "goblin", "name": "Goblin"},
  "class": {"id": "rogue", "name": "Rogue"},
  "dialogId": "goblin-taunt",
  "enemyTrait": {
    "creatureType": "humanoid",
    "combatStyle": "agile",
    "difficulty": "easy",
    "attackPower": 8,
    "defense": 2,
    "aggroRadius": 1,
    "aggroOnSight": true,
    "callForHelp": true,
    "xpReward": 25,
    "goldDrop": {"min": 5, "max": 15},
    "lootTableID": "goblin-loot"
  },
  "respawnTime": "2m"
}
```
