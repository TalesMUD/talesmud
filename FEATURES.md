# TalesMUD Core Systems & Features Reference

**Purpose**: This document catalogs ALL core systems, data structures, features, and APIs available in TalesMUD. It is designed to provide complete context for AI agents reworking game content to leverage the latest capabilities.

**Last Updated**: 2026-06-18

---

## Table of Contents

1. [Data Structures (Entities)](#data-structures-entities)
2. [Room System](#room-system)
3. [Item System](#item-system)
4. [Character System](#character-system)
5. [NPC System](#npc-system)
6. [Combat System](#combat-system)
7. [Skills & Spells System](#skills--spells-system)
8. [Quest System](#quest-system)
9. [Dialog System](#dialog-system)
10. [Scripting System (Lua API)](#scripting-system-lua-api)
11. [Creator UI Capabilities](#creator-ui-capabilities)
12. [Recent Features & Best Practices](#recent-features--best-practices)
13. [Game Client Minimap](#game-client-minimap)
14. [Game Client Tab Container Widget](#game-client-tab-container-widget)
15. [Multiplayer Session and Presence](#multiplayer-session-and-presence)

---

## Data Structures (Entities)

### Base Entity
All entities inherit from:
```go
type Entity struct {
    ID string `json:"id"`  // UUID
}
```

### Common Traits
Reusable trait composition pattern:
- `BelongsUser` - User ownership tracking (`BelongsUserID`)
- `CurrentRoom` - Location tracking (`CurrentRoomID`)
- `LookAt` - Inspection text (`Detail` field)

---

## Room System

### Room Entity Structure
```go
type Room struct {
    *entities.Entity
    traits.LookAt

    Name, Description string
    RoomType, Area, AreaType string
    Tags []string

    // Scripting Hooks
    OnEnterScriptID string  // Lua script executed when player enters room

    Actions *Actions  // Custom player interactions
    Exits   *Exits    // Room connections

    // Live Data (runtime state)
    Items      *Items       // Item IDs in room
    Characters *Characters  // Character IDs in room
    NPCs       *NPCs        // NPC resident IDs (not instances!)

    // Grid Positioning (optional)
    Coords *struct{X, Y, Z int32}

    // Client Meta
    Meta *struct{
        Mood       string  // Ambient mood setting
        Background string  // Background image ID
    }

    CanBind bool  // Allow /bind command for respawn point
}
```

### Room Actions
Players can interact with rooms via custom actions:
```go
type Action struct {
    Name        string  // Command trigger (e.g., "examine statue")
    Description string  // Shown in "You can:" list
    Response    string  // Reply to player (type: response)
    Type        RoomActionType  // "response", "response_room", "script"
    ScriptId    string  // Lua script to execute (type: script)
    Params      map[string]interface{}
}
```

**Action Types**:
- `response` - Send message to player only
- `response_room` - Broadcast message to all in room
- `script` - Execute Lua script with context: `ctx.room`, `ctx.character`, `ctx.action`

### Room Exits
```go
type Exit struct {
    Name        string  // Exit command (e.g., "north", "trapdoor")
    Description string  // Shown in exits list
    Type        RoomExitType  // "normal", "direction", "teleport"
    Hidden      bool    // Not shown until revealed (per-character)
    Target      string  // Target room ID
    Params      map[string]interface{}
}
```

**Exit Types**:
- `direction` - Standard cardinal directions (n/s/e/w/u/d)
- `normal` - Named exits (e.g., "door", "gate")
- `teleport` - Instant transport to distant location

### Per-Character Hidden Exit Reveals
Hidden exits can be revealed on a **per-character basis** using scripting:
```lua
-- In a room action script:
tales.game.revealExit(roomID, "secret-passage", characterID)
```

- Reveal is tracked on `Character.RevealedExits` map (persisted)
- Other players do NOT see the exit unless they also reveal it
- Client receives automatic room update showing new exit

### Room Discovery & Exploration XP
```go
// On Character entity:
DiscoveredRooms map[string]bool  // Room IDs
DiscoveredAreas map[string]bool  // Area names
```

**Exploration XP Rewards**:
- **5 XP** per new room discovered
- **15 XP** for first room in a new area/zone

---

## Item System

### Item Entity Structure
```go
type Item struct {
    *entities.Entity
    traits.LookAt

    // Template/Instance Pattern (same as NPCs)
    IsTemplate     bool    // True if blueprint
    TemplateID     string  // Source template ID (for instances)
    InstanceSuffix string  // Unique suffix (e.g., "abc123")

    Name, Description string
    Type    ItemType     // "currency", "consumable", "armor", "weapon", "collectible", "quest", "crafting_material"
    SubType ItemSubType  // "sword", "twohandsword", "axe", "spear", "shield"
    Slot    ItemSlot     // Equipment slot or "inventory"
    Quality ItemQuality  // "normal", "magic", "rare", "legendary", "mythic"
    Level   int32

    // Custom Properties
    Properties map[string]interface{}  // Arbitrary key-value data
    Attributes map[string]interface{}  // Stats (e.g., "healthRestore": 50)

    // Container System
    Closed   bool
    Locked   bool
    LockedBy string  // Key item ID
    Items    Items   // Nested items
    MaxItems int32

    // Interaction Flags
    NoPickup           bool    // Cannot be picked up
    CopyOnPickup       bool    // Create personal copy instead of removing from room
    BoundToCharacterID string  // Item is bound to this character (cannot drop/sell/trade)

    // Stacking & Economy
    Stackable bool
    Quantity  int32
    MaxStack  int32
    BasePrice int64  // Gold value

    // Scripting & Consumption
    OnUseScriptID string  // Lua script executed when item is used
    Consumable    bool    // Remove/decrement on use

    // Metadata
    Tags      []string
    Created   time.Time
    CreatedBy string
    Meta      *struct{Img string}
}
```

### Item Types & Slots
**Item Types**:
- `currency` - Gold, tokens
- `consumable` - Potions, food, scrolls
- `armor` - Wearable protection
- `weapon` - Swords, axes, staves
- `collectible` - Quest items, trophies
- `quest` - Quest-specific items
- `crafting_material` - Future crafting system

**Equipment Slots**:
- `head`, `chest`, `legs`, `boots`, `hands`, `neck`, `ring1`, `ring2`
- `main_hand`, `off_hand`
- Special: Two-handed weapons occupy both hand slots

### CopyOnPickup Feature
When `Item.CopyOnPickup = true`:
1. **First pickup**: Creates a personal instance from template, adds to inventory
2. **Instance is bound**: `BoundToCharacterID` is set on the instance, making it character-bound
3. **Room remains unchanged** - Item stays in room for other players
4. **Character tracking**: Marks item as collected via `Character.Flags["collected_item:<templateID>"]`
5. **Subsequent pickups**: Player sees "You have already collected X"
6. **Client behavior**: Item is hidden in room display for that character

**Bound item restrictions** (enforced on items with `BoundToCharacterID`):
- Cannot be dropped (`drop` command rejects with message to use `destroy`)
- Cannot be sold to merchants
- Cannot be traded
- `destroy` command removes from inventory, clears collected flag, and refreshes room

**Helper methods**: `Item.IsBound()`, `Item.IsOwnedBy(characterID)`

**Use case**: Quest items, story artifacts that all players should find

### Item Usage & Consumables
```go
// Built-in effects via Attributes:
Attributes: {
    "healthRestore": 50,      // Restore 50 HP
    "manaRestore": 30,        // Restore 30 mana
    "useMessage": "You feel refreshed"  // Custom message
}
```

**Usage flow**:
1. Player uses item: `use health potion`
2. System checks `Attributes` for built-in effects
3. If `OnUseScriptID` is set, executes Lua script
4. If `Consumable = true`, decrements quantity or removes item

### Item Template/Instance Pattern
```go
// Create instance from template:
instance, err := itemsService.CreateInstanceFromTemplate(templateID)

// Check if item is an instance:
if item.IsInstance() {  // Has TemplateID and InstanceSuffix
    // ...
}
```

---

## Character System

### Character Entity Structure
```go
type Character struct {
    *entities.Entity
    traits.BelongsUser
    traits.CurrentRoom

    Name, Description string
    Race  Race   // Human, Elf, Dwarf, Halfling, Orc
    Class Class  // Warrior, Mage, Rogue, Cleric, Ranger, Druid

    // Core Stats
    CurrentHitPoints, MaxHitPoints int32
    CurrentMana, MaxMana int32  // Casters only
    XP, Level int32
    Gold int64
    MaxLevelCap int32  // Per-character level cap (0 = use global MaxLevel)

    // Attribute System
    Attributes Attributes  // []{Name, Short, Value} - STR, DEX, CON, INT, WIS, CHA

    // Distributable Attribute Points (2 per level)
    UnspentAttributePoints int32
    SpentAttributePoints   map[string]int32  // Tracks total spent per attribute

    // Inventory & Equipment
    Inventory     items.Inventory
    EquippedItems map[items.ItemSlot]*items.Item
    EquippedSkills []string  // Skill IDs (max 4, level-gated)

    // Combat State
    InCombat         bool
    CombatInstanceID string
    BoundRoomID      string  // Respawn location (set via /bind)

    // Per-Character Game State
    Flags map[string]interface{}  // Script-settable flags
    RevealedExits map[string][]string  // roomID → exit names
    DiscoveredRooms map[string]bool
    DiscoveredAreas map[string]bool

    // All-Time Statistics
    AllTimeStats struct {
        PlayersKilled   int32
        GoldCollected   int32
        QuestsCompleted int32
        RoomsDiscovered int32
    }
}
```

### Attribute System
Six core attributes (D&D-style):
- **STR** (Strength) - Melee damage, warrior scaling
- **DEX** (Dexterity) - Dodge, rogue/ranger scaling, initiative
- **CON** (Constitution) - HP bonus
- **INT** (Intelligence) - Mage/Druid spell power, mana capacity
- **WIS** (Wisdom) - Cleric spell power, mana regeneration
- **CHA** (Charisma) - Social interactions (future)

**Attribute Modifiers**:
```go
modifier = (value - 10) / 2
// Example: STR 14 → +2 modifier, STR 8 → -1 modifier
```

### Distributable Attribute Points
Players earn **2 points per level-up** to spend freely:
```bash
# Terminal commands:
spend          # Show status table with current values, spent/cap
spend str 3    # Allocate 3 points to STR
spend dex      # Allocate 1 point to DEX
```

**Class-based Caps** prevent degenerate builds:
- Warriors: INT capped at 5
- Mages: STR capped at 5
- Each class has primary/secondary attribute caps defined

**Retroactive Points**: Existing characters receive `(level - 1) * 2` points on login

### Character Methods (Combat-Related)
```go
// Attribute access
GetAttribute(short string) int32          // Get raw value
GetAttributeModifier(short string) int    // Get (value-10)/2
GetSTRMod(), GetDEXMod(), GetCONMod() int // Specific modifiers
GetINTMod(), GetWISMod() int

// Combat calculations
GetWeaponDamage() int32        // Main hand damage (1 if unarmed)
GetArmorDefense() int32        // Total from equipped armor
CalculateMaxMana() int32       // Caster: 20 + Level*5 + INTMod*4
CalculateManaRegen() int32     // In-combat: 1 + WISMod (min 1)
```

### Character Flags System
Arbitrary key-value storage for script state:
```lua
-- Set flag:
tales.game.setFlag(characterID, "puzzle_solved_statue", true)

-- Get flag:
local solved = tales.game.getFlag(characterID, "puzzle_solved_statue")
```

**Common uses**: Quest state, puzzle progress, secret discoveries

### CopyOnPickup Tracking
```go
// Character methods:
HasCollectedCopyItem(templateID string) bool
MarkCollectedCopyItem(templateID string)

// Stored in Flags["collected_copy_items"] as array of template IDs
```

---

## NPC System

### NPC Entity Structure
```go
type NPC struct {
    *entities.Entity
    traits.BelongsUser
    traits.CurrentRoom

    Name, Description string
    Race  Race
    Class Class

    CurrentHitPoints, MaxHitPoints int32
    Level int32

    // Template/Instance Pattern
    IsTemplate     bool    // Blueprint vs spawned instance
    TemplateID     string  // Source template for instances
    InstanceSuffix string  // Unique ID suffix

    // Behavior Configuration
    SpawnRoomID  string         // Where NPC respawns
    RespawnTime  time.Duration  // Time to respawn (0 = no respawn)
    WanderRadius int            // Rooms to wander from spawn
    PatrolPath   []string       // Room IDs for patrol route

    // State Machine
    IsDead    bool
    DeathTime time.Time
    State     string  // "idle", "combat", "patrol", "dead", "fleeing"

    // Combat State
    InCombat         bool
    CombatInstanceID string

    // Behavior Traits
    EnemyTrait    *EnemyTrait
    MerchantTrait *MerchantTrait

    // Dialog References
    DialogID          string          // Main dialog tree
    IdleDialogID      string          // Ambient chatter
    IdleDialogTimeout time.Duration   // Idle dialog frequency
}
```

### Enemy Trait
```go
type EnemyTrait struct {
    // Classification
    CreatureType string  // "beast", "humanoid", "undead", "elemental", "construct", "demon", "dragon", "aberration"
    CombatStyle  string  // "melee", "ranged", "magic", "swarm", "brute", "agile"
    Difficulty   string  // "trivial", "easy", "normal", "hard", "boss"

    // Combat Stats (base values, modified by difficulty multipliers)
    AttackPower  int32
    Defense      int32
    AttackSpeed  float64

    // AI Behavior
    AggroRadius   int     // Detection range in rooms (0 = passive)
    AggroOnSight  bool    // Auto-attack on detection
    CallForHelp   bool    // Alert nearby enemies
    FleeThreshold float64 // HP % to flee (e.g., 0.2 = flee at 20% HP)

    // Rewards
    XPReward       int64
    GoldDrop       Range{Min, Max int64}
    LootTableID    string    // Reference to loot table
    GuaranteedLoot []string  // Item template IDs that always drop
    MaxDrops       int32     // Max items from loot table (0 = unlimited)

    // Event Scripts
    OnAggroScript string  // Lua script on aggro
    OnDeathScript string  // Lua script on death
    OnFleeScript  string  // Lua script on flee
}
```

### Merchant Trait
```go
type MerchantTrait struct {
    MerchantType   string  // "general", "blacksmith", "alchemist"
    Inventory      []MerchantItem
    RestockMinutes int32
    LastRestock    time.Time
    BuyMultiplier  float64  // Price when buying (1.0 = normal)
    SellMultiplier float64  // Price when selling (0.5 = half)
    AcceptedTypes  []string // Item types merchant will buy (empty = all)
    RejectedTags   []string // Tags preventing buy (e.g., "soulbound", "quest")
}

type MerchantItem struct {
    ItemTemplateID string
    BasePrice      int64   // Override item base price
    PriceOverride  int64   // Force specific price (ignores multipliers)
    Quantity       int32   // Current stock (-1 = unlimited)
    MaxQuantity    int32   // Max after restock
    RequiredLevel  int32   // Player level requirement
}
```

Merchant commands are available in rooms with merchant NPCs:
- `list`, `shop`, or `trade` shows current stock and prices
- `buy <item> [quantity]` purchases stock; stackable quantities can fit in one inventory stack
- `sell <item> [quantity]` sells accepted, unbound inventory items
- `value <item>` / `price <item>` checks the merchant's sell price

Trading is blocked while the character is in combat. Merchant stock can restock lazily when a player interacts after the configured interval.

### NPC Spawner System
```go
type NPCSpawner struct {
    *entities.Entity

    TemplateID    string         // NPC template to spawn
    RoomID        string         // Where to spawn
    MaxInstances  int            // Max alive at once
    SpawnInterval time.Duration  // Time between spawns
    InitialCount  int            // Spawn on world load

    RespawnTimeOverride *time.Duration  // Override template respawn
}
```

**Spawner Behavior** (automated by server):
- Checks every 5 seconds
- Spawns instances up to `MaxInstances`
- Tracks live instances via `NPCInstanceManager`

### NPC Resident System
Two methods for placing NPCs:
1. **Via spawners** - Dynamic, respawning instances
2. **Via Room.NPCs list** - Static residents
3. **Via NPC.CurrentRoomID** - Unique NPCs auto-spawn into their assigned room on server start

**Important**: `Room.NPCs` contains template/unique NPC IDs, NOT instance IDs

---

## Combat System

### Combat Instance Model
```go
type CombatInstance struct {
    ID              string
    OriginRoomID    string  // Room where combat started

    Players         []CombatantRef
    Enemies         []CombatantRef

    TurnOrder       []CombatantRef  // Initiative-sorted
    CurrentTurnIdx  int
    TurnStartTime   time.Time
    Round           int

    State           CombatState  // "pending", "active", "victory", "defeat", "fled", "timeout"
    Log             []CombatLogEntry
}
```

### Combatant Reference
```go
type CombatantRef struct {
    ID, Name        string
    Type            CombatantType  // "player" or "npc"
    Initiative      int
    IsAlive, HasFled bool

    // Core Stats
    MaxHP, CurrentHP int32
    AttackPower, Defense int32
    STRMod, DEXMod, CONMod, INTMod, WISMod int
    DefenseBonus    int32  // From defend action

    // Mana & Skills
    MaxMana, CurrentMana, ManaRegen int32
    EquippedSkills []string
    SkillCooldowns map[string]int  // SkillID → rounds remaining
    StatusEffects  []StatusEffect
    QueuedSkillID  string
}
```

### Status Effects
```go
type StatusEffect struct {
    ID, SkillID, Name string
    Type    string    // "buff", "debuff", "dot", "hot", "stun"
    Stat    string    // "attack", "defense", "dodge"
    Value   int32     // Flat modifier
    Percent float64   // Percentage modifier (0.30 = +30%)
    Duration int      // Rounds remaining
    SourceID string   // Caster ID
}
```

### Combat Flow
```
1. INITIATION
   - Player: attack <npc>
   - NPC: aggro detection (AggroRadius)
   - Create CombatInstance, roll initiative (1d20 + DEX mod)

2. TURN ORDER
   - Sorted by initiative (highest first)
   - Each combatant takes action in sequence
   - Round increments after all combatants act

3. PLAYER ACTIONS (60-second timer per turn)
   - attack <target> - Melee/ranged attack
   - cast <skill> [target] - Use skill (mana/cooldown)
   - defend - +50% defense until next turn
   - flee - Attempt escape (50% + DEX bonus)
   - timeout - Auto-defend after 60 seconds

4. NPC ACTIONS (automated AI)
   - If HP < FleeThreshold → attempt flee
   - Otherwise → attack weakest player

5. ROUND START EFFECTS
   - Process DoT/HoT ticks
   - Regenerate mana (1 + WISMod per round)
   - Decrement cooldowns
   - Check stun effects

6. RESOLUTION
   - Victory (all enemies dead) → XP, gold, loot
   - Defeat (all players dead) → 10% XP loss, 1 gold loss, respawn at bind point
   - Fled (all players escaped) → NPCs reset to idle
   - Timeout (30 minutes) → Combat ends, no rewards
```

### Combat Commands
```bash
attack <target>    # Attack enemy or switch target
a <target>         # Alias

cast <skill> [target]  # Use skill
spell <skill>          # Alias
1, 2, 3, 4             # Quick-cast slot shortcuts

defend             # Defensive stance (+50% defense)
d, guard           # Aliases

flee               # Attempt to escape
run, escape        # Aliases

status             # Show combat status
cs, combat         # Aliases
```

### Difficulty Balance System
**Config File**: `config/combat_balance.yaml`

```yaml
difficulty_multipliers:
  trivial:
    hp: 3.75
    attack: 4.0
    defense: 1.0
  easy:
    hp: 2.9
    attack: 1.67
    defense: 1.0
  normal:
    hp: 2.67
    attack: 1.5
    defense: 2.0
  hard:
    hp: 1.25
    attack: 1.18
    defense: 1.25
  boss:
    hp: 2.0
    attack: 1.5
    defense: 1.67
```

**How it works**:
- Enemy base stats (HP, Attack, Defense) are defined on `EnemyTrait`
- On combat initiation, stats are multiplied by difficulty tier
- Allows fine-tuning balance without editing all NPCs

---

## Skills & Spells System

### Skill Entity Structure
```go
type Skill struct {
    *entities.Entity

    Name, Description string
    ClassIDs       []string  // Multi-class: ["warrior"], ["cleric", "druid"]
    LevelRequired  int32

    // Resource System
    ResourceType   ResourceType  // "mana" (casters) or "cooldown" (physical)
    ManaCost       int32         // Mana cost per use
    CooldownRounds int           // Rounds before reuse

    // Targeting & Effects
    Target         TargetType    // "enemy", "self", "all_enemies"
    Effect         EffectType    // "damage", "heal", "buff", "debuff", "dot", "hot"
    ScalingAttr    string        // "STR", "DEX", "INT", "WIS"
    BasePower      int32
    ScalingFactor  float64
    Duration       int           // Rounds (0 = instant)

    // Buff/Debuff Details
    BuffStat       string        // "ATK", "DEF", "STR", "DEX"
    BuffPercent    float64       // 0.30 = +30%

    // Special Mechanics
    IgnoresDefense bool
    HitCount       int           // Multi-hit attacks

    // Secondary Effect (optional, for hybrid skills)
    SecondaryEffect, SecondaryTarget, etc.
}
```

### Skill Storage & Cache
- **Database-backed** with in-memory cache
- `LoadFromDB()` at server startup
- `RefreshCache()` after CRUD operations
- Combat engine reads from cache (zero service threading)

### Skill Registry Functions
```go
SkillByID(id) *Skill                      // Look up by ID
SkillsForClass(classID) []*Skill          // All skills for class
AvailableSkills(classID, level) []*Skill  // Skills at level
MaxSkillSlots(classID, level) int         // Equippable slots
IsCasterClass(classID) bool               // Uses mana
```

### Skill Slot Progression
| Class | L1 | L10 | L15 | L20 | L30 |
|-------|:--:|:---:|:---:|:---:|:---:|
| Mage/Cleric/Druid | 2 | 2 | 3 | 3 | 4 |
| Warrior/Rogue/Ranger | 1 | 2 | 2 | 3 | 4 |

### Skill Management Commands
```bash
skills                  # List available and equipped skills
skills equip <name>     # Equip skill to next slot
skills unequip <name>   # Remove skill from slots
```

**Combat Restrictions**: Cannot equip/unequip during combat

### Default Skills Seeding
29 default skills seeded on first run when DB is empty:
- Warrior: Strike, Cleave, Berserker Rage, Shield Bash, Whirlwind
- Rogue: Backstab, Poison Strike, Shadow Step, Eviscerate
- Ranger: Aimed Shot, Multi-Shot, Hunter's Mark, Piercing Arrow
- Mage: Fireball, Ice Lance, Lightning Bolt, Arcane Missiles, Meteor
- Cleric: Heal, Holy Smite, Divine Shield, Prayer of Healing, Resurrection
- Druid: Heal, Moonfire, Thorns, Regrowth, Starfall

---

## Quest System

### Quest Entity Structure
```go
type Quest struct {
    *entities.Entity

    Name, Description string
    Category    string  // "main", "side", "daily"
    Area        string  // Optional region/zone label
    Level       int32
    Repeatable  bool

    Source      QuestSource  // "npc", "item", "auto", "script"
    Objectives  []Objective
    Rewards     Reward

    RequiredQuestIDs []string  // Prerequisite quests
    RequiredLevel    int32

    // NPC Dialog Integration
    AcceptDialogText, ProgressDialogText, CompleteDialogText string
    OnCompleteScriptID string  // Lua script on completion
}
```

### Quest Source Types
```go
type QuestSource struct {
    Type   string  // "npc", "item", "auto", "script"
    NPCID  string  // Quest-giving NPC
    ItemID string  // Quest-triggering item
}
```

### Objective Types
```go
type Objective struct {
    Type        ObjectiveType  // "kill", "collect", "deliver", "visit", "talk", "custom"
    Description string

    // Type-specific fields
    TargetNPCTemplateID string  // For kill
    TargetItemTemplateID string // For collect
    TargetRoomID        string  // For visit/deliver
    TargetDialogNodeID  string  // For talk
    ScriptID            string  // For custom (Lua)

    CurrentCount, RequiredCount int32
    Complete     bool
}
```

### Quest Progress (Per-Character)
```go
type QuestProgress struct {
    *entities.Entity

    CharacterID string
    QuestID     string
    Status      QuestStatus  // "available", "active", "completed", "failed", "abandoned"
    Objectives  []ObjectiveProgress
    AcceptedAt, CompletedAt time.Time
}
```

### Quest Rewards
```go
type Reward struct {
    XP    int64
    Gold  int64
    Items []RewardItem  // Item template IDs with quantities
}
```

### Quest Tracker (Automatic Progress)
The `QuestTracker` listens to game events and updates objectives:

| Event | Source | Objectives Updated |
|-------|--------|-------------------|
| NPC killed | Combat victory | Kill objectives (matches template ID) |
| Item pickup | Pickup command | Collect objectives (matches template ID) |
| Room enter | Room navigation | Visit objectives (matches room ID) |
| Dialog node | Talk command | Talk objectives (matches NPC + node) |
| Talk to NPC | Talk command | Deliver objectives (if has required item) |

### Quest Dialog Integration
When talking to a quest-source NPC:
1. **Offer option** - If quest is available and prerequisites met
2. **Progress option** - If quest is active but not complete
3. **Turn-in option** - If all objectives complete

Dialog options are **automatically injected** into NPC conversations.

### Quest Commands
```bash
quests          # Show quest log (active quests + progress)
ql, questlog    # Aliases

quest <name>    # Show quest details

abandon <name>  # Abandon active quest
```

### Quest UI Features

#### Quest Log Widget
The quest log widget provides a comprehensive quest management interface:

**Core Features:**
- **Real-time Quest Search**
  - Search across quest names, descriptions, and objectives
  - Case-insensitive, instant filtering
  - Clear button (×) to reset search

- **Category Filtering**
  - Filter by: All Types, Main, Side, Daily
  - Toggle visibility: Completed quests, Abandoned quests
  - Helps focus on relevant quests

- **Sort Options**
  - Sort by Status (Active → Completed → Abandoned)
  - Sort by Name (alphabetical)
  - Sort by Level (ascending)
  - Sort by Category (Main/Side/Daily)

- **Quest Pinning System**
  - Pin up to 5 priority quests
  - Pinned section always at top
  - Golden pulsing indicator for pinned quests
  - Persisted in localStorage
  - Pin/Unpin buttons in quest details

**Quest Display:**
- Quest name with level badge (e.g., "L5")
- Category badge with color coding:
  - 🟠 Main quests (amber/orange #f59e0b)
  - 🔵 Side quests (blue #3b82f6)
  - 🟣 Daily quests (purple #8b5cf6)
- Expandable quest details
- Objective progress (X/Y format) with checkmarks
- Rewards preview (XP, Gold, Items)
- Abandon button for active quests
- Pin/Unpin button for tracking

**Quest Sections:**
1. **📌 Pinned** - Priority quests (if any)
2. **Active** - In-progress quests
3. **Completed** - Finished quests with completion dates
4. **Abandoned** - Dropped quests (when visible)
5. **Failed** - Failed quests (when visible)

#### Quest History & Statistics Panel

Click the 📊 button in quest log header to access:

**Quest Statistics:**
- Total quests encountered
- Active quest count
- Completed quest count
- Completion rate percentage
- Total XP earned from quests
- Total gold earned from quests

**Category Breakdown:**
- Main quests completed
- Side quests completed
- Daily quests completed

**Quest Achievements:**
8 built-in achievements that unlock automatically:

| Achievement | Requirement |
|-------------|-------------|
| First Steps | Complete 1 quest |
| Quest Novice | Complete 5 quests |
| Quest Veteran | Complete 10 quests |
| Quest Master | Complete 25 quests |
| Story Seeker | Complete 5 main quests |
| Side Quest Hero | Complete 10 side quests |
| Daily Devotee | Complete 5 daily quests |
| Completionist | 100% completion rate (min 5 quests) |

**Achievement Display:**
- 🏆 Unlocked achievements (gold border, highlighted)
- 🔒 Locked achievements (grayed out, collapsible)
- Shows name and description
- Real-time progress tracking

#### Quest Notifications

Enhanced notification system with interactions:

**Notification Types:**
- **Quest Accepted** (amber border)
- **Quest Progress** (blue border) - shows completed objective
- **Quest Completed** (green border)

**Interactions:**
- **Click to View** - Opens quest in quest log
- **Dismiss Button (×)** - Manual dismiss with slide-out animation
- **Hover Effects** - Highlights notification
- **Auto-dismiss** - Removes after 5 seconds

**Features:**
- Unique notification IDs
- Slide-in and slide-out animations
- Smooth transitions
- Top-right positioning

### Quest API Endpoints

#### Get Quest Log with Full Details
```
GET /api/quest-progress/:characterId
```

Requires the authenticated user to own `characterId`, unless the user is an admin.

**Response includes:**
- Quest progress (status, objectives)
- Quest definition (name, description, category, level)
- Rewards (XP, gold, item template IDs)
- Timestamps (acceptedAt, completedAt)

**Response Format:**
```json
[
  {
    "questId": "quest-uuid",
    "questName": "The Wolf Problem",
    "status": "active",
    "description": "Wolves have been attacking travelers...",
    "category": "main",
    "level": 3,
    "objectives": [
      {
        "objectiveId": "obj-1",
        "description": "Defeat wolves",
        "current": 3,
        "required": 5,
        "completed": false
      }
    ],
    "rewards": {
      "xp": 500,
      "gold": 50,
      "itemTemplateIds": ["item-template-1"]
    },
    "acceptedAt": "2026-02-16T10:30:00Z",
    "completedAt": null
  }
]
```

### Quest WebSocket Messages

#### Quest Log Message
Sent on character selection and quest updates:
```json
{
  "type": "questLog",
  "quests": [ /* array of QuestLogEntry */ ]
}
```

#### Quest Update Messages
```json
{
  "type": "questAccepted",
  "questId": "quest-uuid",
  "questName": "The Wolf Problem",
  "message": "Quest accepted!"
}

{
  "type": "questProgress",
  "questId": "quest-uuid",
  "questName": "The Wolf Problem",
  "objectives": [ /* updated objectives */ ]
}

{
  "type": "questCompleted",
  "questId": "quest-uuid",
  "questName": "The Wolf Problem",
  "message": "Quest completed!"
}
```

### Quest Client Store

The MUD client stores quest data in `MUDXPlusStore`:

**State:**
```javascript
{
  quests: [],              // Full quest log with details
  questNotifications: [],  // Active notifications
  pinnedQuests: []        // Stored in localStorage
}
```

**Methods:**
- `updateQuests(questLog)` - Update full quest list
- `addQuestNotification(notification)` - Add notification with auto-dismiss

**LocalStorage:**
- `pinnedQuests` - JSON array of quest IDs (max 5)
- Persists across sessions
- Shared between widget and overlay

### Quest Tracker Implementation

Automatic progress tracking on game events:

**OnNPCKilled(characterID, userID, deadNPC):**
- Checks all active quests for kill objectives
- Matches NPC template ID
- Increments objective counter
- Sends progress update message

**OnItemPickup(characterID, userID, item):**
- Checks collect objectives
- Matches item template ID
- Increments counter
- Sends progress update

When a collect quest is accepted, matching items already in inventory initialize objective progress. Stackable quantities count toward the initial objective amount.

**OnRoomEnter(characterID, userID, room):**
- Checks visit objectives
- Matches room ID
- Marks objective complete
- Sends progress update

**OnDialogNode(characterID, userID, npcID, dialogID, nodeID):**
- Checks talk objectives
- Matches NPC and optional dialog node
- Marks complete
- Sends progress update

**OnTalkToNPC(characterID, userID, npc):**
- Checks deliver objectives
- Verifies player has required item
- Marks complete
- Sends progress update

### Quest Completion Flow

1. **Check Objectives** - All must be complete (checked server-side)
2. **Mark Complete** - Update quest progress status
3. **Grant Rewards**
   - Add XP to character
   - Add gold to character
   - Create items from templates
   - Add to character inventory
4. **Send Notification** - WebSocket message with rewards
5. **Update Quest Log** - Client refreshes via API call
6. **Update Achievements** - Calculated client-side on quest log update
```

### Quest Scripting API
```lua
-- Check quest status
local status = tales.quests.getStatus(characterID, questID)

-- Accept quest
tales.quests.accept(characterID, questID)

-- Complete quest
tales.quests.complete(characterID, questID)

-- Update objective progress
tales.quests.updateProgress(characterID, questID, objectiveIndex, count)

-- Grant quest items
tales.quests.grantItems(characterID, questID)

-- Abandon quest
tales.quests.abandon(characterID, questID)

-- Check if quest can be accepted
local canAccept = tales.quests.canAccept(characterID, questID)
```

---

## Dialog System

### Dialog Entity Structure
```go
type Dialog struct {
    ID   string
    Text string  // Primary speech

    // Text Variations
    AlternateTexts []string  // Random variations
    OrderedTexts   *bool     // Sequential on repeat

    // Dialog Flow
    Options []Dialog  // Player choices (branching)
    Answer  *Dialog   // Auto-response (linear)

    // Conditional Display
    RequiresVisitedDialogs []string  // Prerequisites
    ShowOnlyOnce           *bool     // One-time option
    IsDialogExit           *bool     // End conversation
}
```

### Dialog State (Per-Conversation)
```go
type DialogState struct {
    CurrentDialogID  string
    DialogVisited    map[string]int  // Node ID → visit count
    Context          map[string]string  // Static variables
    DynamicContext   map[string]func() string  // Runtime variables
}
```

### Dialog Variable Substitution
```
{{PLAYER}}   → Character name
{{NPC}}      → NPC name
{{TIME}}     → Current time
{{CUSTOM}}   → Script-set context variables
```

### Dialog Features
- **Branching** - Multiple player choices via `Options`
- **Linear** - Automatic progression via `Answer`
- **Conditional** - Options shown only if prerequisites visited
- **One-time** - `ShowOnlyOnce` options disappear after first selection
- **Variations** - Random or sequential alternate texts
- **Exit markers** - Return to "main" node on completion

### Dialog API
```lua
-- Get dialog
local dialog = tales.dialogs.get(dialogID)

-- Get conversation state
local conv = tales.dialogs.getConversation(characterID, npcID)

-- Set context variable
tales.dialogs.setContext(conversationID, "playerClass", "warrior")

-- Get context variable
local class = tales.dialogs.getContext(conversationID, "playerClass")

-- Check if visited
local visited = tales.dialogs.hasVisited(conversationID, nodeID)

-- Get visit count
local count = tales.dialogs.getVisitCount(conversationID, nodeID)
```

---

## Scripting System (Lua API)

### Lua Runner Architecture
- **Language**: Lua 5.1 (via gopher-lua)
- **VM Pool**: Reusable Lua states for performance
- **Timeout**: 5-second execution limit
- **Sandbox**: Restricted environment (no file I/O, no network)

### Script Types
- `item` - Item creation, behavior, OnUse handlers
- `room` - Room actions, OnEnter handlers
- `npc` - NPC behavior, aggro/death/flee events
- `quest` - Quest logic, custom objectives
- `event` - Event handlers (player.enter_room, npc.death, etc.)
- `custom` - General purpose

### Context Variable (`ctx`)
Scripts receive context data via global `ctx`:
```lua
-- Room enter script:
ctx.eventType   -- "player.enter_room"
ctx.room        -- Room entity
ctx.character   -- Character entity
ctx.user        -- User entity

-- Item use script:
ctx.eventType   -- "item.use"
ctx.item        -- Item entity
ctx.character   -- Character entity
ctx.room        -- Room entity (if in room)

-- Room action script:
ctx.eventType   -- "room.action"
ctx.room        -- Room entity
ctx.character   -- Character entity
ctx.action      -- Action entity
```

### tales.game Module

All functions below are called as `tales.game.functionName(...)`. Every parameter is **required** unless noted otherwise. Functions that perform writes return `bool` (true on success, false on failure). Missing or wrong-type arguments cause a **Lua error that aborts the script**.

#### Messaging
```lua
tales.game.msgToRoom(roomID, message)
-- Send message to ALL players in a room. Returns bool.
-- roomID: string (room entity ID, e.g., "R0004")
-- message: string

tales.game.msgToCharacter(characterID, message)
-- Send message to a specific character (via their user's socket). Returns bool.
-- characterID: string (character entity ID)
-- message: string

tales.game.msgToUser(userID, message)
-- Send message to a specific user by user ID. Returns bool.
-- userID: string (user entity ID — NOT character ID)
-- message: string

tales.game.broadcast(message)
-- Send message to ALL connected players globally. Returns bool.
-- message: string

tales.game.msgToRoomExcept(roomID, message, excludeCharacterID)
-- Send message to all players in room EXCEPT the specified character. Returns bool.
-- roomID: string
-- message: string
-- excludeCharacterID: string (character entity ID to exclude)
```

#### Logging
```lua
tales.game.log(level, message)
-- Write to server log. Returns nothing.
-- level: string — "debug", "info", "warn", "error"
-- message: string
```

#### Inventory & Equipment Checks
```lua
tales.game.hasItem(characterID, itemID)
-- Check if character has an item in inventory. Returns bool.
-- Checks by exact item ID first, then by TemplateID (so you can pass "ITM0001").
-- characterID: string
-- itemID: string (item instance ID or template ID)

tales.game.hasEquipped(characterID, slotName)
-- Check if character has an item equipped in a specific slot. Returns bool.
-- characterID: string
-- slotName: string — "head", "chest", "legs", "boots", "hands", "neck",
--   "ring1", "ring2", "main_hand", "off_hand"
```

#### Character Flags (Per-Character State)
```lua
tales.game.getFlag(characterID, flagName)
-- Get a flag value from a character. Returns the value or nil if not set.
-- characterID: string
-- flagName: string
-- Return types: bool, number, string, or nil

tales.game.setFlag(characterID, flagName, value)
-- Set a flag on a character. Persisted to database immediately. Returns bool.
-- characterID: string
-- flagName: string
-- value: bool, number, string, or nil (nil deletes the flag)
```

#### Hidden Exit Reveals (Per-Character)

**CRITICAL**: Note the different parameter orders between `revealExit` and `hasRevealedExit`.

```lua
tales.game.revealExit(roomID, exitName, characterID)
-- Reveal a hidden exit for a specific character. Returns bool.
-- The exit must already exist on the room with hidden=true.
-- The reveal is stored on the character (Character.RevealedExits), not the room.
-- Sends a silent room update to the client (no full re-render).
-- Idempotent: safe to call multiple times (returns true if already revealed).
-- ALL THREE PARAMETERS ARE REQUIRED.
-- roomID: string (room entity ID where the exit is defined)
-- exitName: string (exit name, case-insensitive match, e.g., "north")
-- characterID: string (character entity ID to reveal the exit for)

tales.game.hasRevealedExit(characterID, roomID, exitName)
-- Check if a character has revealed a specific hidden exit. Returns bool.
-- NOTE: Parameter order is (characterID, roomID, exitName) — different from revealExit!
-- characterID: string
-- roomID: string
-- exitName: string (case-insensitive match)
```

#### Item Rewards
```lua
tales.game.giveItem(characterID, templateID)
-- Create an item instance from a template and add it to the character's inventory.
-- Returns bool. Persisted to database immediately.
-- characterID: string
-- templateID: string (item template ID, e.g., "ITM0022")
```

#### CopyOnPickup Item Tracking
```lua
tales.game.hasCollectedItem(characterID, templateID)
-- Check if a character has already collected a CopyOnPickup item. Returns bool.
-- characterID: string
-- templateID: string (item template ID)

tales.game.resetCollectedItem(characterID, templateID)
-- Reset the collected flag for a CopyOnPickup item, allowing re-collection. Returns bool.
-- characterID: string
-- templateID: string (item template ID)
```

### tales.items Module
```lua
-- Get item
local item = tales.items.get(itemID)

-- Find by name
local items = tales.items.findByName("sword")

-- Templates
local template = tales.items.getTemplate(templateID)
local templates = tales.items.findTemplates("sword")

-- Create instance
local newItem = tales.items.createFromTemplate(templateID)

-- Delete
local success = tales.items.delete(itemID)
```

### tales.rooms Module
```lua
-- Get room
local room = tales.rooms.get(roomID)

-- Find rooms
local rooms = tales.rooms.findByName("tavern")
local rooms = tales.rooms.findByArea("dungeon")
local allRooms = tales.rooms.getAll()

-- Room contents
local characters = tales.rooms.getCharacters(roomID)
local npcs = tales.rooms.getNPCs(roomID)
local items = tales.rooms.getItems(roomID)
```

### tales.characters Module
```lua
-- Get character
local char = tales.characters.get(characterID)

-- Find characters
local chars = tales.characters.findByName("hero")
local allChars = tales.characters.getAll()

-- Get character's room
local room = tales.characters.getRoom(characterID)

-- Character operations
tales.characters.damage(characterID, amount)
tales.characters.heal(characterID, amount)
tales.characters.teleport(characterID, roomID)
tales.characters.giveXP(characterID, amount)
```

### tales.npcs Module
```lua
-- Get NPC
local npc = tales.npcs.get(npcID)

-- Find NPCs
local npcs = tales.npcs.findByName("guard")
local npcs = tales.npcs.findInRoom(roomID)
local allNPCs = tales.npcs.getAll()

-- NPC operations
tales.npcs.damage(npcID, amount)
tales.npcs.heal(npcID, amount)
tales.npcs.moveTo(npcID, roomID)

-- NPC checks
local isDead = tales.npcs.isDead(npcID)
local isEnemy = tales.npcs.isEnemy(npcID)
local isMerchant = tales.npcs.isMerchant(npcID)

-- Template & instance operations
local templates = tales.npcs.getTemplates()
local isTemplate = tales.npcs.isTemplate(npcID)
local instance = tales.npcs.spawnFromTemplate(templateID, roomID)
local inst = tales.npcs.getInstance(instanceID)
local instances = tales.npcs.getInstancesInRoom(roomID)

-- Instance operations
tales.npcs.kill(instanceID)
tales.npcs.setState(instanceID, state)  -- "idle", "combat", "patrol", "fleeing"
local state = tales.npcs.getState(instanceID)
local died = tales.npcs.damageInstance(instanceID, amount)
tales.npcs.healInstance(instanceID, amount)
tales.npcs.moveInstance(instanceID, roomID)

-- Delete
tales.npcs.delete(npcID)
```

### tales.dialogs Module
```lua
-- Get dialog
local dialog = tales.dialogs.get(dialogID)

-- Find dialogs
local dialogs = tales.dialogs.findByName("greeting")
local allDialogs = tales.dialogs.getAll()

-- Conversation management
local conv = tales.dialogs.getConversation(characterID, npcID)
tales.dialogs.setContext(conversationID, "key", "value")
local value = tales.dialogs.getContext(conversationID, "key")

-- Visit tracking
local visited = tales.dialogs.hasVisited(conversationID, nodeID)
local count = tales.dialogs.getVisitCount(conversationID, nodeID)
```

### tales.quests Module
```lua
-- Quest status
local status = tales.quests.getStatus(characterID, questID)

-- Quest operations
tales.quests.accept(characterID, questID)
tales.quests.complete(characterID, questID)
tales.quests.abandon(characterID, questID)

-- Progress tracking
tales.quests.updateProgress(characterID, questID, objectiveIndex, count)

-- Rewards
tales.quests.grantItems(characterID, questID)

-- Checks
local canAccept = tales.quests.canAccept(characterID, questID)
```

### tales.utils Module
```lua
-- Random numbers
local num = tales.utils.random(1, 100)  -- Inclusive
local f = tales.utils.randomFloat()     -- 0.0-1.0

-- UUID
local id = tales.utils.uuid()

-- Time
local now = tales.utils.now()           -- Unix timestamp
local nowMs = tales.utils.nowMs()       -- Milliseconds
local str = tales.utils.formatTime(timestamp)

-- Dice rolling
local result = tales.utils.roll("2d6+3")

-- Probability
local success = tales.utils.chance(25)  -- 25% chance → true/false

-- Array utilities
local picked = tales.utils.pick({"sword", "axe", "spear"})
local shuffled = tales.utils.shuffle(myArray)

-- Math utilities
local clamped = tales.utils.clamp(value, 0, 100)
local lerped = tales.utils.lerp(0, 100, 0.5)  -- Returns 50
```

### Script Execution Contexts

Each script type receives a `ctx` global table with different fields. Access fields as `ctx.fieldName`. Object fields expose entity properties (e.g., `ctx.character.ID`, `ctx.room.Name`).

**Room OnEnter Script** (`type: room`, assigned via room's `onEnterScript`):
- Triggered when player enters room (walk, teleport, or character select)
- Context variables:
  - `ctx.eventType` — string: `"player.enter_room"`
  - `ctx.room` — Room entity object (access `.ID`, `.Name`, etc.)
  - `ctx.toRoom` — Room entity object (same as `ctx.room`)
  - `ctx.character` — Character entity object (access `.ID`, `.Name`, etc.)
  - `ctx.user` — User entity object
- **Getting room ID**: Use `ctx.room.ID`
- **Getting character ID**: Use `ctx.character.ID` or store as `local charID = character.ID`

**Room Action Script** (`type: custom`, assigned via room action's `scriptId`):
- Triggered by player room action (e.g., `EXAMINE RUNES`, `PRESS CIRCLE`)
- Context variables:
  - `ctx.eventType` — string: `"room.action"`
  - `ctx.room` — Room entity object
  - `ctx.roomID` — string: room ID (convenience shortcut for `ctx.room.ID`)
  - `ctx.character` — Character entity object
  - `ctx.characterID` — string: character ID (convenience shortcut for `ctx.character.ID`)
  - `ctx.action` — string: action name (e.g., `"PRESS CIRCLE"`)
  - `ctx.params` — table: action params from YAML (e.g., `{action="press", symbol="circle"}`)
- **Note**: Room action scripts get BOTH object and string ID shortcuts (`ctx.roomID` and `ctx.room.ID` both work)

**Item OnUse Script** (`type: item`, assigned via item's `onUseScriptID`):
- Triggered by `use <item>` command
- Context variables:
  - `ctx.eventType` — string: `"item.use"`
  - `ctx.item` — Item entity object
  - `ctx.character` — Character entity object
  - `ctx.room` — Room entity object (if character is in a room)

**NPC Event Scripts** (`type: npc`):
- `OnAggroScript`, `OnDeathScript`, `OnFleeScript`
- Context varies by event type

### Context Variable Quick Reference

| Variable | OnEnter (room) | Action (custom) | Item Use |
|----------|:-:|:-:|:-:|
| `ctx.eventType` | `"player.enter_room"` | `"room.action"` | `"item.use"` |
| `ctx.room` | Room object | Room object | Room object |
| `ctx.roomID` | — | string shortcut | — |
| `ctx.character` | Character object | Character object | Character object |
| `ctx.characterID` | — | string shortcut | — |
| `ctx.user` | User object | — | — |
| `ctx.action` | — | action name string | — |
| `ctx.params` | — | action params table | — |
| `ctx.item` | — | — | Item object |

**Important**: In OnEnter scripts, you must use `ctx.room.ID` and `ctx.character.ID` (no string shortcuts). In Action scripts, you can use either `ctx.roomID` or `ctx.room.ID`.

---

## Creator UI Capabilities

### Data Table System
All entity editors use a unified **filterable, sortable data table**:
- **Per-column filtering** - Text search, enum dropdowns
- **Client-side filtering** - Instant, no API calls
- **Sorting** - Click column headers (asc → desc → none)
- **Master-detail layout** - Table + edit form side-by-side
- **Full-width mode** - Close detail panel to maximize table view

### Entity Select Modal
**MANDATORY UI GUIDELINE**:
- **NEVER use `<select>` dropdowns for entity ID references**
- **ALWAYS use `EntitySelectButton` + `EntitySelectModal`**
- Provides filterable DataTable for selecting rooms, NPCs, items, scripts, dialogs, quests
- Scales to hundreds of entries with search and filter

```svelte
<EntitySelectButton
  value={roomID}
  elements={rooms}
  columns={roomColumns}
  title="Select Room"
  placeholder="Select a room..."
  on:change={(e) => roomID = e.detail}
/>
```

### Creator Tabs
1. **Rooms** - Full room editor (exits, actions, spawners, NPCs, items, scripts)
2. **Items** - Item instances and templates
3. **Item Templates** - Reusable item blueprints
4. **NPCs** - NPC templates and unique NPCs
5. **Dialogs** - Dialog tree editor
6. **Quests** - Quest editor (objectives, rewards, prerequisites)
7. **Skills** - Skill/spell editor (multi-class, effects)
8. **Scripts** - Lua script editor with syntax highlighting
9. **Character Templates** - Archetype editor with modal item-template selection for starting gear
10. **World Map** - Grid-based world visualization

### Room Editor Features
- **Exit management** - Add/edit/delete exits, toggle hidden
- **Action management** - Custom interactions (response, broadcast, script)
- **NPC residents** - Assign unique NPCs to room
- **Item placement** - Add items to room
- **Spawner configuration** - NPC spawner setup
- **OnEnter script** - Lua script triggered on room entry
- **Background & mood** - Visual settings
- **Coordinates** - Grid positioning (X, Y, Z)
- **Bind point** - Allow `/bind` for respawn

### NPC Editor Features
- **Template vs. Unique** - Toggle `IsTemplate` flag
- **Basic info** - Name, description, race, class, level
- **Enemy trait** - Combat stats, difficulty, loot, AI behavior
- **Merchant trait** - Inventory, pricing, restock, accepted items
- **Dialog assignment** - Main dialog, idle dialog
- **Behavior** - Wander radius, patrol path, respawn time
- **Resident placement** - Assign `CurrentRoomID` for auto-spawn

### Quest Editor Features
- **Objectives** - Add kill/collect/deliver/visit/talk/custom objectives
- **Area** - Optional regional label with dynamic suggestions and table filtering
- **Rewards** - XP, gold, item grants
- **Prerequisites** - Required quests, level requirement
- **Dialog integration** - Accept/progress/complete text
- **Repeatable** - Daily/recurring quests
- **OnComplete script** - Lua script on quest completion

### Skill Editor Features
- **Multi-class assignment** - Skills can belong to multiple classes
- **Resource type** - Mana-based or cooldown-based
- **Effect system** - Damage, heal, buff, debuff, DoT, HoT, stun
- **Scaling** - Attribute-based damage/healing (STR/DEX/INT/WIS)
- **Secondary effects** - Hybrid skills (e.g., damage + debuff)

### Script Editor Features
- **Syntax highlighting** - Lua code highlighting
- **Test runner** - Execute scripts with test context
- **Type categorization** - item, room, npc, quest, event, custom

---

## Recent Features & Best Practices

### Per-Character Hidden Exit Reveals
**Feature**: Hidden exits can be revealed individually per character via scripting.

**Best Practice — Revealing an exit in an action script** (`type: custom`):
```lua
-- In a puzzle room action script (has ctx.roomID and ctx.characterID shortcuts):
local charID = ctx.character.ID
if tales.game.hasItem(charID, "bronze-key-template-id") then
    tales.game.msgToCharacter(charID, "The key fits! A secret passage opens.")
    -- ALL THREE PARAMS REQUIRED: roomID, exitName, characterID
    tales.game.revealExit(ctx.roomID, "secret-passage", charID)
else
    tales.game.msgToCharacter(charID, "You need a bronze key to unlock this.")
end
```

**Best Practice — State reconciliation in OnEnter script** (`type: room`):

When a puzzle sets a flag AND reveals an exit, the exit reveal could fail independently. Always reconcile state in the room's OnEnter script to ensure consistency:
```lua
-- In the room's onEnterScript (type: room — uses ctx.room.ID, NOT ctx.roomID):
local charID = ctx.character.ID
local puzzleSolved = tales.game.getFlag(charID, "puzzle_solved")
if puzzleSolved then
    -- Reconcile: ensure the exit is actually revealed for this character
    if not tales.game.hasRevealedExit(charID, ctx.room.ID, "north") then
        tales.game.revealExit(ctx.room.ID, "north", charID)
    end
    tales.game.msgToCharacter(charID, "The passage stands open.")
end
```

**Storage**: `Character.RevealedExits` map (roomID → exit names), persisted to DB
**Client**: Automatic silent room update shows new exit without re-rendering description
**Idempotent**: Calling `revealExit` on an already-revealed exit is a no-op (returns true)

**IMPORTANT — Parameter order difference**:
- `revealExit(roomID, exitName, characterID)` — room first
- `hasRevealedExit(characterID, roomID, exitName)` — character first

### CopyOnPickup Items
**Feature**: Items that create personal copies instead of removing from room. Instances are bound to the collecting character.

**Best Practice**:
- Use for **quest items** that all players should find
- Use for **story artifacts** that don't deplete
- Set `Item.CopyOnPickup = true` in Creator UI
- System tracks via `Character.Flags["collected_item:<templateID>"]`
- Instances are automatically bound (`BoundToCharacterID`) — cannot be dropped, sold, or traded
- Players use `destroy` command to discard bound items (clears collected flag, allowing re-pickup)

**Example**: Ancient scroll in library - all players can read it, but each gets a bound copy

### Item Consumables with Lua Scripts
**Feature**: Items can have both data-driven effects AND custom Lua logic.

**Best Practice**:
```go
// In Creator UI - Item Attributes:
{
  "healthRestore": 50,
  "useMessage": "You feel refreshed"
}

// Set OnUseScriptID for additional effects:
OnUseScriptID: "potion-buff-script"
```

**Script Example**:
```lua
-- potion-buff-script (type: item — ctx.character is a Character object)
local charID = ctx.character.ID
tales.game.msgToCharacter(charID, "A warm glow surrounds you.")
-- Could add temporary buff, quest progress, etc.
```

### Distributable Attribute Points
**Feature**: Players earn 2 points per level to spend on attributes.

**Best Practice**:
- Design quests/challenges that reward specific builds
- Use class caps to prevent degenerate builds (warriors can't max INT)
- Show unspent points badge in character widget
- Remind players to spend points at level-up

### Combat Balance via YAML Config
**Feature**: Difficulty multipliers loaded from `config/combat_balance.yaml`.

**Best Practice**:
- Set base stats conservatively on NPCs
- Use difficulty tier to scale for different zones
- Trivial: Tutorial enemies
- Easy: Early-game
- Normal: Mid-game
- Hard: Late-game
- Boss: Unique encounters

**Tuning**: Adjust YAML file without editing individual NPCs

### Quest Dialog Auto-Injection
**Feature**: Quest offer/progress/complete options auto-injected into NPC dialogs.

**Best Practice**:
- Design NPC dialog trees WITHOUT quest options
- System automatically adds quest options at runtime
- Set `AcceptDialogText`, `ProgressDialogText`, `CompleteDialogText` on Quest
- Quest-giver NPC should have general greeting dialog

### Exploration XP System
**Feature**: Automatic XP rewards for discovering rooms/areas.

**Best Practice**:
- Design areas with hidden rooms for bonus XP
- Use revealed exits to reward thorough exploration
- Track via `Character.DiscoveredRooms` and `DiscoveredAreas`
- **5 XP** per room, **15 XP** for first room in new area

### Room OnEnter Scripts
**Feature**: Lua scripts executed when player enters room. Script `type` must be `room`.

**Best Practice**:
```lua
-- OnEnter scripts use ctx.character.ID (uppercase .ID, not .id)
local character = ctx.character
if not character then return end
local charID = character.ID

-- First-visit message:
if not tales.game.getFlag(charID, "visited_dungeon") then
    tales.game.msgToCharacter(charID, "You enter a dark, foreboding dungeon.")
    tales.game.setFlag(charID, "visited_dungeon", true)
end

-- State reconciliation for hidden exits (if this room has puzzle-revealed exits):
local puzzleSolved = tales.game.getFlag(charID, "dungeon_puzzle_solved")
if puzzleSolved and not tales.game.hasRevealedExit(charID, ctx.room.ID, "north") then
    tales.game.revealExit(ctx.room.ID, "north", charID)
end
```

**Context**: `ctx.room` (Room object), `ctx.character` (Character object), `ctx.user` (User object)
**Note**: OnEnter scripts do NOT have `ctx.roomID` or `ctx.characterID` string shortcuts — use `ctx.room.ID` and `ctx.character.ID` instead.

### Multi-Class Skills
**Feature**: Skills can be assigned to multiple classes.

**Best Practice**:
- Healing skills for both Cleric and Druid
- Defensive buffs for Warrior, Ranger, Rogue
- Set `ClassIDs: ["cleric", "druid"]` in Creator UI
- Registry handles lookups automatically

### NPC Auto-Spawn via CurrentRoomID
**Feature**: Unique NPCs auto-spawn into their assigned room on server start.

**Best Practice**:
- Use for quest-givers, merchants, named NPCs
- Set `NPC.CurrentRoomID` in Creator UI
- Do NOT add to spawners
- Do NOT add to `Room.NPCs` list (system handles it)

### Item Template/Instance Pattern
**Feature**: Items follow same template/instance pattern as NPCs.

**Best Practice**:
- Create item **templates** with `IsTemplate = true`
- Use `CreateInstanceFromTemplate()` for loot, rewards
- Instances track `TemplateID` for quest matching
- Use `CopyOnPickup` for shared quest items

---

## Summary of Key Patterns

### Entity ID References
- **Rooms** - Reference other rooms via `Exit.Target`
- **Items** - Reference templates via `TemplateID`, scripts via `OnUseScriptID`
- **NPCs** - Reference templates via `TemplateID`, dialogs via `DialogID`, scripts via `OnAggroScript`/`OnDeathScript`/`OnFleeScript`
- **Quests** - Reference NPCs via `Source.NPCID`, scripts via `OnCompleteScriptID`
- **Skills** - Reference classes via `ClassIDs` array

### Per-Character State
- **Flags** - Arbitrary script data (`Character.Flags`)
- **Revealed Exits** - Hidden exit discoveries (`Character.RevealedExits`)
- **Discovered Rooms** - Exploration tracking (`Character.DiscoveredRooms`)
- **Collected Copy Items** - CopyOnPickup tracking (`Character.Flags["collected_copy_items"]`)
- **Quest Progress** - Active quests (`QuestProgress` entity)
- **Equipped Skills** - Active skill slots (`Character.EquippedSkills`)

### Template/Instance Pattern
**Used by**: Items, NPCs

**Fields**:
- `IsTemplate` - True if blueprint
- `TemplateID` - Source template ID (for instances)
- `InstanceSuffix` - Unique suffix (e.g., "abc123")

**Creation**:
```go
instance, err := service.CreateInstanceFromTemplate(templateID)
```

### Scripting Hooks
**Room**:
- `OnEnterScriptID` - Player enters room

**Item**:
- `OnUseScriptID` - Player uses item

**NPC**:
- `OnAggroScript` - Enemy aggros
- `OnDeathScript` - Enemy dies
- `OnFleeScript` - Enemy flees

**Quest**:
- `OnCompleteScriptID` - Quest completed

**Room Action**:
- `Action.ScriptId` - Player triggers action

---

## Game Client Minimap

### Overview
The minimap widget renders an auto-discovered map from visited rooms. It is canvas-based and tracks room positions via coordinate inference from directional exits.

### Room Rendering
- **Current room**: Amber glow + amber border (highlighted)
- **Nearby rooms** (BFS distance <= 2): Slightly brighter fill/border
- **All other rooms**: Uniform color, full opacity (no distance-based fading)
- **Travel path**: Rooms on the click-to-travel path highlighted in blue
- **Vertical exits**: Purple triangle indicators (up/down arrows) on rooms with vertical exits
- **Z-level label**: Shows current floor in the top-left corner

### Coordinate Inference
Rooms are positioned on a grid using coordinate inference:
- Server-provided `coords` are used when available
- Otherwise, coords are inferred from the previous room's exit direction
- Supports cardinal directions (north/south/east/west) and vertical directions (up/down)
- Direction aliases handled: "upward" -> "up", "downward" -> "down", etc.
- Hidden exits are included in spatial tracking (they define room adjacency)
- Fallback: non-standard exit names (e.g., "residence", "portal") place rooms at the nearest unoccupied adjacent position to prevent coordinate chain breaks

### Interaction
- **Click-to-travel**: Click a room to auto-navigate via BFS pathfinding
- **Panning**: Click and drag to pan the map view; auto-recenters on room change
- **Recenter button**: Appears when panned; click the crosshair icon to snap back to current room
- **Maximize/minimize**: Button in the top-left expands the minimap to a fullscreen overlay
- **Tooltips**: Hover over rooms to see their names

### State Persistence
- Visited rooms stored in browser `localStorage` under key `talesmud_visitedRooms`
- Rooms accumulate across zone transitions (not cleared on zone change)
- "Clear map data" button resets all visited room data

---

## Game Client Tab Container Widget

### Overview
The Tab Container widget allows players to group multiple widgets into a single resizable container with switchable tabs. This reduces layout clutter, especially on smaller screens (e.g., Character + Inventory + Equipment in one tab group).

### Features
- **Multiple tab containers**: Unlimited tab containers can exist on screen simultaneously
- **Tab switching**: Click tabs to switch between contained widgets; active tab is highlighted amber
- **State preservation**: All tab content is kept mounted (via CSS visibility) — scroll positions, terminal buffers, and widget state are preserved when switching tabs
- **Tab labels**: Automatically derived from widget registry (icon + name)
- **Add/remove tabs**: Available in edit mode only — "+" button opens a dropdown of available widget types, "x" button removes individual tabs
- **No nesting**: Tab containers cannot be placed inside other tab containers (enforced at registry, store, and UI levels)
- **maxInstances enforcement**: Widgets inside tab containers count toward their `maxInstances` limit (e.g., you can't add Character both as a standalone widget and inside a tab)
- **Persistence**: Tab configuration (which widgets, active tab) persists to localStorage alongside layout data

### Data Model
Tab containers store extra fields on the widget layout item:
```javascript
{
  id: 'tabcontainer-1707000000000',
  widgetType: 'tabcontainer',
  tabs: [
    { widgetType: 'character', id: 'tab-character-001' },
    { widgetType: 'inventory', id: 'tab-inventory-002' }
  ],
  activeTabIndex: 0
}
```

### Key Files
- `public/mud-client/src/game/widgets/TabContainerWidget.svelte` — The tab container component
- `public/mud-client/src/game/layout/WidgetComponents.js` — Shared component map used by both WidgetGrid and TabContainerWidget
- `public/mud-client/src/game/layout/WidgetRegistry.js` — Registry entry (`tabcontainer` type, `layout` category)
- `public/mud-client/src/game/layout/LayoutStore.js` — Tab management methods (`addTabToContainer`, `removeTabFromContainer`, `setActiveTab`)

### LayoutStore Tab Methods
```javascript
layoutStore.addTabToContainer(containerId, widgetType)  // Add a widget as a new tab
layoutStore.removeTabFromContainer(containerId, tabIndex) // Remove a tab by index
layoutStore.setActiveTab(containerId, tabIndex)           // Switch active tab
```

---

## Guest Mode System

### Overview
Guest mode allows anonymous visitors to play TalesMUD for 30 minutes without Auth0 registration. Designed for embedding a live demo on the website.

### Guest Service (`pkg/service/guest.go`)
```go
type GuestService interface {
    CreateGuestSession(remoteIP string) (token string, err error)
    ValidateGuestToken(tokenStr string) (userID string, err error)
    CleanupExpiredGuests()
    StartCleanupLoop()
}
```

### Guest Session Lifecycle
1. Client calls `POST /api/guest` (public endpoint)
2. Server checks `ServerSettings.GuestsAllowed` and `MaxGuestAccounts`
3. IP rate limit checked (10 per hour per IP)
4. Random character created from system template presets with full starter items
5. Character `MaxLevelCap` set to 5
6. User created with `IsGuest=true`, `GuestExpiresAt=now+30min`
7. HMAC-SHA256 token signed with `GUEST_SECRET` env var
8. Token returned to client, stored in `sessionStorage`
9. On WebSocket connect: timeout goroutines start (5-min warning + expiry)
10. On disconnect: 5-min grace period before deleting guest data
11. Background cleanup loop removes expired guests every 5 minutes

### Server Configuration
```go
// In ServerSettings:
GuestsAllowed    bool  // Enable/disable guest mode (default: true)
MaxGuestAccounts int   // Max concurrent guests (default: 20, 0 = unlimited)
```

### User Entity Guest Fields
```go
// In User:
IsGuest        bool       // Marks temporary guest account
GuestExpiresAt time.Time  // When guest session expires
```

### Per-Character Level Cap
```go
// In Character:
MaxLevelCap int32  // 0 = use global MaxLevel, otherwise per-character cap

// Helper method:
func (c *Character) GetEffectiveMaxLevel(globalMax int32) int32
```

The leveling system (`CheckLevelUp`, `ApplyLevelUp`) respects `MaxLevelCap` automatically.

### Frontend Guest Flow
- `WelcomeScreen.svelte` — "Play as Guest" button (amber/gold styling)
- `App.svelte` — `handleGuestPlay()` stores token in sessionStorage, skips onboarding
- `UserMenu.svelte` — Guest-aware: shows "Create Account" and "End Session" instead of Auth0 controls
- `api/guest.js` — `createGuestSession()` API client

### Authentication
- Guest HMAC tokens are validated before Auth0 JWTs in `AuthMiddleware`
- Token claims: `sub` (RefID), `uid` (user entity ID), `exp` (30min), `guest: true`
- If `GUEST_SECRET` is not set, a random key is generated at startup

---

## Multiplayer Session and Presence

### Character Selection Lifecycle
- WebSocket connect marks the user online immediately and reselects the user's `LastCharacter`.
- Selecting a character updates `User.LastCharacter`, removes any previous active character from its old room, adds the selected character to its current room, and sends the initial room, character, inventory, and quest state.
- Character switch and movement paths emit silent room presence refreshes so other clients update player lists without waiting for a full room render.

### Room Presence WebSocket Message
```json
{
  "type": "roomPresence",
  "roomId": "room-uuid",
  "players": [
    { "id": "character-uuid", "name": "Aster", "isYou": false }
  ]
}
```

The server broadcasts this to the room with online active characters only. The client marks `isYou` locally based on the currently selected character because a single broadcast is shared by multiple users.

### Reconnect Behavior
- The MUD server replaces an existing WebSocket connection when the same user reconnects.
- Stale socket close events cannot remove the replacement connection or mark the user offline.
- The client tracks `connecting`, `connected`, `reconnecting`, and `disconnected` states and retries with the same authenticated WebSocket URL.

### Party Commands
| Command | Description |
|---------|-------------|
| `party create` | Create a party with the current character as the first member |
| `party invite <player>` | Invite an online player by active character name |
| `party accept` | Accept the pending in-memory party invite |
| `party leave` | Leave the current party |
| `party say <message>` | Send party chat to online active party members |

Party membership is persisted by character ID in the existing `Party` entity. Party invites are an in-memory foundation layer and are only valid while the server process is running.

### Game Client UX
- `CharacterSwitcher.svelte` shows the selected character, connection state, and available characters from `/api/my-characters`.
- Switching characters sends `sc <character name>` over the existing command channel.
- `Client.js` handles `roomPresence` messages and updates `MUDXPlusStore.players` without changing the room description.

---

## Document Maintenance

**Update this file when**:
- Adding new entity fields
- Adding new Lua API functions
- Adding new game systems
- Adding new Creator UI features
- Changing data structures

**Keep in sync with**:
- `PROJECT.md` - High-level features
- `ARCHITECTURE.md` - System architecture
- `docs/design/SCRIPTING.md` - Scripting documentation
- Creator UI column definitions (`tableColumns.js`)

---

**End of Document**
