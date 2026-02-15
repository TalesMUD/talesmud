# TalesMUD Core Systems & Features Reference

**Purpose**: This document catalogs ALL core systems, data structures, features, and APIs available in TalesMUD. It is designed to provide complete context for AI agents reworking game content to leverage the latest capabilities.

**Last Updated**: 2026-02-15

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
    NoPickup     bool  // Cannot be picked up
    CopyOnPickup bool  // Create personal copy instead of removing from room

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
2. **Room remains unchanged** - Item stays in room for other players
3. **Character tracking**: Marks item as collected via `Character.Flags["collected_copy_items"]`
4. **Subsequent pickups**: Player sees "You have already collected X"
5. **Client behavior**: Item is hidden in room display for that character

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
   - Defeat (all players dead) → 10% gold loss, respawn at bind point
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
9. **Character Templates** - Archetype editor
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
**Feature**: Items that create personal copies instead of removing from room.

**Best Practice**:
- Use for **quest items** that all players should find
- Use for **story artifacts** that don't deplete
- Set `Item.CopyOnPickup = true` in Creator UI
- System tracks via `Character.Flags["collected_copy_items"]`

**Example**: Ancient scroll in library - all players can read it, but each gets a copy

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
