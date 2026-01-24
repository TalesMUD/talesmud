# TalesMUD Architecture Documentation

## System Overview

TalesMUD follows a layered architecture with clear separation between HTTP API, real-time game server, business logic, and data access layers.

```
┌─────────────────────────────────────────────────────────────────┐
│                         Clients                                  │
│              (Browser / WebSocket / REST API)                    │
└─────────────────────────────────────────────────────────────────┘
                              │
         ┌────────────────────┼────────────────────┐
         ▼                    ▼                    ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│   HTTP Server   │  │   MUD Server    │  │  Embedded SPA   │
│   (Gin Router)  │  │  (WebSocket)    │  │  (go:embed FS)  │
└────────┬────────┘  └────────┬────────┘  └─────────────────┘
         │                    │
         │                    ▼
         │           ┌─────────────────┐
         │           │   Game Engine   │
         │           │ (Command Loop)  │
         │           └────────┬────────┘
         │                    │
         ▼                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Service Layer                               │
│    (Characters, Rooms, Users, Items, Scripts, Parties)          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Repository Layer                             │
│                     (SQLite backend)                             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
              ┌───────────────────────────────┐
              │            SQLite             │
              │ (tables with JSON payloads)   │
              └───────────────────────────────┘
```

## Backend Architecture

### Entry Point

**File:** `cmd/tales/main.go`

```go
func main() {
    godotenv.Load()      // Load .env configuration
    server.NewApp().Run() // Initialize and start server
}
```

### Server Layer (`pkg/server/`)

The HTTP server handles REST API requests and serves as the entry point for WebSocket connections.

#### Server Structure

```go
type app struct {
    Router    *gin.Engine      // HTTP router
    Facade    service.Facade   // Service layer access
    mud       mud.MUDServer    // Game server instance
}
```

#### Database Configuration

`server.NewApp()` initializes the SQLite storage backend at startup.

Use `SQLITE_PATH` to specify the database file path (defaults to `talesmud.db`).

#### Route Organization

```
/                          # Static files (frontend)
/health                    # Health check
/ws                        # WebSocket (game connection)
/api/
    ├── characters/        # Character CRUD
    ├── rooms/             # Room CRUD
    ├── items/             # Item CRUD (templates: ?isTemplate=true)
    ├── scripts/           # Script CRUD
    ├── user               # User profile
    └── templates/         # Public templates
/admin/
    ├── export             # World export
    ├── import             # World import
    └── world              # World map
```

#### Authentication Middleware

**File:** `pkg/server/auth.go`

```
Request → Extract JWT → Validate with Auth0 JWKS → Find/Create User → Set Context
```

- Supports both query parameter (`?access_token=`) and Authorization header
- Validates against Auth0 JWKS endpoint
- Creates new user on first login
- Sets `userid` and `user` in Gin context

### MUD Server (`pkg/mudserver/`)

The MUD server manages WebSocket connections and routes messages between clients and the game engine.

#### Connection Model

```go
type Connection struct {
    User   *entities.User   // Authenticated user
    ws     *websocket.Conn  // WebSocket connection
    mu     sync.Mutex       // Thread-safe writes
    active bool             // Connection state
}
```

#### Server Components

```go
type server struct {
    Facade    service.Facade           // Service access
    Game      *game.Game               // Game instance
    Clients   map[string]*Connection   // Active connections
    Broadcast chan interface{}         // Global messages
    Upgrader  websocket.Upgrader       // HTTP→WS upgrade
}
```

#### Concurrent Goroutines

The MUD server runs 4 concurrent goroutines:

1. **Message Receiver** - Routes game output to appropriate clients
2. **Game Loop** - Processes commands and updates game state
3. **Broadcast Handler** - Sends global messages to all clients
4. **Timeout Handler** - Sends ping every 60 seconds

#### Message Flow

```
Client WebSocket
       │
       ▼ (JSON IncomingMessage)
HandleConnections()
       │
       ▼ (creates Message struct)
OnMessageReceived channel
       │
       ▼
Game.Run() processes
       │
       ▼ (creates MessageResponse)
SendMessage channel
       │
       ▼
receiveMessages() routes by Audience
       │
       ├──► sendMessage(userID)     [Single user]
       ├──► sendToRoom(room)        [All in room]
       ├──► sendToRoomWithout()     [Room except origin]
       └──► Broadcast channel       [Global]
```

### Game Engine (`pkg/mudserver/game/`)

The game engine contains the core game loop, command processing, and state management.

#### Game Structure

```go
type Game struct {
    Facade           service.Facade        // Service access
    SystemUser       *entities.User        // System actor

    // Event Channels
    onMessageReceived chan interface{}     // Player input
    sendMessage       chan interface{}     // Game output
    OnUserJoined      chan *m.UserJoined   // Login events
    OnUserQuit        chan *m.UserQuit     // Logout events

    // Processors
    CommandProcessor  *CommandProcessor    // Global commands
    RoomProcessor     *RoomProcessor       // Room commands

    Avatars          map[string]*Avatar    // Active players
}
```

#### Game Loop

```go
func (g *Game) Run() {
    for {
        select {
        case joined := <-g.OnUserJoined:
            g.handleUserJoined(joined.User)

        case quit := <-g.OnUserQuit:
            g.handleUserQuit(quit.User)

        case msg := <-g.onMessageReceived:
            message := msg.(*m.Message)
            // Try global commands first
            if !g.CommandProcessor.Process(g, message) {
                // Then room commands
                g.RoomProcessor.Process(g, message)
            }
        }
    }
}
```

#### Update Cycle

The game runs multiple update tickers:

1. **Room Updates** (10 seconds)
   - Remove offline characters from rooms
   - Sync room state to database
   - Clean up stale connections

2. **NPC Updates** (10 seconds)
   - Process NPC state machine (idle, patrol, combat, fleeing)
   - Handle respawning for dead NPCs
   - Execute behavior updates (wander, patrol paths)

3. **Spawner Updates** (5 seconds)
   - Check each spawner's instance count
   - Spawn new instances when below max capacity
   - Clean up dead instances from spawner tracking

4. **Combat Updates** (2 seconds)
   - Check turn timeouts (60 seconds per player turn)
   - Process NPC turns when it's their turn
   - Handle AFK auto-flee after 3 consecutive timeouts
   - Check global combat timeout (30 minutes)

#### NPCInstanceManager

Manages in-memory NPC instances that are spawned from templates:

```go
type NPCInstanceManager struct {
    instances    map[string]*npc.NPC      // Active NPC instances
    spawnerState map[string]*SpawnerState // Runtime spawner state
    facade       service.Facade
}
```

Key methods:
- `Initialize()` - Load spawners and create initial instances
- `SpawnInstance(spawner)` - Create instance from spawner template
- `SpawnInstanceDirect(templateID, roomID)` - Spawn without spawner
- `GetInstancesInRoom(roomID)` - Query instances by room
- `KillInstance(id)` - Mark instance as dead
- `RespawnInstance(id)` - Restore instance to alive state

#### CombatController

Manages turn-based combat instances:

```go
type CombatController struct {
    manager *combat.Manager    // In-memory combat instances
    engine  *combat.Engine     // Combat logic and calculations
    game    *Game              // Game reference for notifications
}
```

Key methods:
- `IsPlayerInCombat(characterID)` - Check if player is in combat
- `IsNPCInCombat(npcID)` - Check if NPC is in combat
- `InitiateCombat(roomID, players, enemies)` - Start new combat
- `ProcessPlayerAttack(characterID, targetID)` - Handle attack action
- `ProcessPlayerDefend(characterID)` - Handle defend action
- `ProcessPlayerFlee(characterID)` - Handle flee attempt
- `GetCombatStatus(characterID)` - Get formatted combat status
- `Update()` - Process combat tick (turn timeouts, NPC turns)

### Combat System (`pkg/mudserver/game/combat/`)

Turn-based combat occurs in isolated **Combat Instances** that manage fights between players and NPCs.

#### Combat Instance Model

```go
type CombatInstance struct {
    ID              string
    OriginRoomID    string          // Room where combat started

    Players         []CombatantRef  // Player combatants
    Enemies         []CombatantRef  // NPC combatants

    TurnOrder       []CombatantRef  // Initiative-sorted order
    CurrentTurnIdx  int             // Current turn
    TurnStartTime   time.Time       // For timeout tracking
    Round           int             // Current combat round

    State           CombatState     // pending, active, victory, defeat, fled, timeout
    Log             []CombatLogEntry
}

type CombatantRef struct {
    ID, Name        string
    Type            CombatantType   // player or npc
    Initiative      int             // Turn order priority
    IsAlive, HasFled bool

    MaxHP, CurrentHP int32
    AttackPower, Defense int32
    STRMod, DEXMod, CONMod int      // Attribute modifiers
    DefenseBonus    int32           // From defend action
}
```

#### Combat Flow

```
1. INITIATION
   └── Player attacks NPC OR aggressive NPC detects player
   └── Create CombatInstance with participants
   └── Roll initiative (1d20 + DEX mod), sort turn order

2. COMBAT ROUND
   └── For each combatant in turn order:
       ├── Player turn: 60-second timer, choose action
       │   ├── attack <target> - Roll to hit, deal damage
       │   ├── defend - +50% defense until next turn
       │   ├── flee - Chance-based escape (50% + DEX bonus)
       │   └── timeout - Auto-defend after 60s
       └── NPC turn: AI decides action
           ├── If HP < flee threshold → attempt flee
           └── Otherwise → attack weakest player

3. RESOLUTION
   ├── Victory (all enemies dead) → XP/gold rewards, loot drops
   ├── Defeat (all players dead) → 10% gold loss, respawn at bind point
   └── Fled (all players escaped) → NPCs reset to idle
```

#### Combat Commands

| Command | Aliases | Description |
|---------|---------|-------------|
| `attack <target>` | `a`, `hit` | Attack enemy (or initiate combat) |
| `defend` | `d`, `guard` | Take defensive stance (+50% defense) |
| `flee` | `run`, `escape` | Attempt to escape (50% + DEX bonus) |
| `status` | `cs`, `combat` | Show combat status |
| `bind` | - | Bind respawn point at current room |

#### Combat Update Cycle (2 seconds)

```go
func (c *CombatController) Update() {
    for _, instance := range c.manager.GetActiveInstances() {
        // Check turn timeout
        if instance.IsTurnTimedOut() {
            // Player: auto-defend, check AFK auto-flee
            // Advance to next turn
        }

        // Check global timeout (30 minutes)
        if time.Since(instance.CreatedAt) >= 30*time.Minute {
            // End combat as timeout
        }
    }
}
```

### Command System (`pkg/mudserver/game/commands/`)

#### Command Interface

```go
type Command interface {
    Execute(game def.GameCtrl, message *messages.Message) bool
    Key() CommandKey
}

type CommandKey interface {
    Matches(key string, cmd string) bool
}
```

#### Command Types

- **ExactCommandKey** - Exact match (e.g., "help")
- **StartsWithCommandKey** - Prefix match (e.g., "sc" matches "sc warrior")

#### Command Processor Flow

```
Input: "selectcharacter warrior"
       │
       ▼
Parse first word: "selectcharacter"
       │
       ▼
Look up in commands map
       │
       ▼
CommandKey.Matches("sc", "selectcharacter") → true
       │
       ▼
Command.Execute(game, message)
       │
       ▼
Return: handled (true/false)
```

#### Room Processor Flow

```
Input: "north"
       │
       ▼
Check character has room
       │
       ▼
Try static commands (n/s/e/w/look/room)
       │
       ▼
Try dynamic exits (custom exit names)
       │
       ▼
Try room actions (custom interactions)
       │
       ▼
Execute matched command
```

### Item Commands

| Command | Aliases | Description |
|---------|---------|-------------|
| `pickup <item>` | `get`, `take` | Pick up item from room |
| `drop <item> [qty]` | - | Drop item to room |
| `examine <item>` | `inspect` | Detailed item inspection |
| `inventory` | `i` | Show inventory (categorized) |
| `equipment` | `eq`, `gear` | Show equipped items |
| `equip <item>` | `wear` | Equip item from inventory |
| `unequip <slot\|item>` | `remove` | Unequip to inventory |

### Trade Commands

| Command | Aliases | Description |
|---------|---------|-------------|
| `list` | `shop` | Show merchant inventory |
| `buy <item> [qty]` | - | Purchase from merchant |
| `sell <item> [qty]` | - | Sell to merchant |
| `value <item>` | `price` | Check sell price |

### Service Layer (`pkg/service/`)

Business logic layer using the Facade pattern.

#### Facade Interface

```go
type Facade interface {
    CharactersService() CharactersService
    PartiesService() PartiesService
    UsersService() UsersService
    RoomsService() RoomsService
    ScriptsService() ScriptsService
    ItemsService() ItemsService
    NPCsService() NPCsService
    NPCSpawnersService() NPCSpawnersService
    DialogsService() DialogsService
    ConversationsService() ConversationsService
    LootTablesService() LootTablesService
    Runner() scripts.ScriptRunner
}
```

#### Service Responsibilities

| Service | Responsibilities |
|---------|------------------|
| UsersService | User CRUD, online status, find/create on login |
| CharactersService | Character CRUD, templates, name validation |
| RoomsService | Room CRUD, room queries |
| ItemsService | Item CRUD, create from template |
| ScriptsService | Script CRUD, execution |
| PartiesService | Party/group management |
| LootTablesService | Loot table CRUD, loot rolling |

### Repository Layer (`pkg/repository/`)

Data access layer with SQLite backend using JSON document storage.

#### Generic Repository

```go
type sqliteGenericRepo struct {
    db        *sql.DB
    table     string
    generator func() interface{}  // Entity factory
}
```

#### Operations

- `FindByID(id)` - Single entity by ID
- `FindByField(key, value)` - Query by field
- `FindAll(collector)` - Stream all with callback
- `FindAllWithParam(params, collector)` - Parameterized query
- `Store(entity)` - Insert new
- `Update(entity)` - Update existing
- `Delete(id)` - Remove entity

### Database Layer (`pkg/db/` + `pkg/db/sqlite/`)

SQLite JSON document storage with one row per entity, using JSON1 extension for queries, WAL mode, and busy timeout.

#### Query Parameter Builder

```go
NewQueryParams().
    With(QueryParam{Key: "area", Value: "oldtown"}).
    With(QueryParam{Key: "roomType", Value: "tavern"})
```

#### Collections

| Collection | Entity |
|------------|--------|
| users | User accounts |
| characters | Player characters |
| charactertemplates | Character blueprints |
| rooms | Game locations |
| items | Items (templates and instances, distinguished by `isTemplate` field) |
| npcs | NPC templates and singletons |
| npc_spawners | NPC spawn point definitions |
| dialogs | Dialog trees |
| conversations | Active dialog sessions |
| scripts | Game scripts |
| parties | Player groups |
| loot_tables | Loot drop configurations |

## Entity Model

### Base Entity

```go
type Entity struct {
    ID string `json:"id"`  // UUID
}
```

### Trait Composition

Entities use embedded traits for shared behavior:

```go
// Location tracking
type CurrentRoom struct {
    CurrentRoomID string `bson:"currentRoomID"`
}

// User ownership
type BelongsUser struct {
    BelongsUserID string `bson:"belongsUserID"`
}

// Inspection text
type LookAt struct {
    Detail string `bson:"detail"`
}
```

### Character Entity

```go
type Character struct {
    *entities.Entity
    traits.BelongsUser
    traits.CurrentRoom

    Name, Description string
    Race, Class

    CurrentHitPoints, MaxHitPoints int32
    XP, Level int32
    Gold int64

    Inventory     items.Inventory
    EquippedItems map[items.ItemSlot]*items.Item

    // Combat state
    InCombat         bool    // Currently in combat
    CombatInstanceID string  // Active combat instance
    BoundRoomID      string  // Respawn location (set via /bind)

    AllTimeStats
}
```

Helper methods for combat:
- `GetAttribute(short)` - Get attribute value by short name (STR, DEX, etc.)
- `GetAttributeModifier(short)` - Get modifier ((value - 10) / 2)
- `GetSTRMod()`, `GetDEXMod()`, `GetCONMod()` - Specific modifiers
- `GetWeaponDamage()` - Main hand weapon damage (1 if unarmed)
- `GetArmorDefense()` - Total defense from equipped armor

### Room Entity

```go
type Room struct {
    *entities.Entity
    traits.LookAt

    Name, Description string
    RoomType, Area    string
    Tags              []string

    Actions    *Actions     // Custom interactions
    Exits      *Exits       // Room connections

    Items      *Items       // Items in room
    Characters *Characters  // Players in room
    NPCs       *NPCs        // NPCs in room

    Coords *struct{X, Y, Z int32}  // Grid position
    Meta   *struct{Mood, Background string}

    CanBind bool  // If true, players can /bind here for respawn
}
```

### NPC Entity

```go
type NPC struct {
    *entities.Entity
    traits.BelongsUser
    traits.CurrentRoom

    Name, Description string
    Race, Class

    CurrentHitPoints, MaxHitPoints int32
    Level int32

    // Template System
    IsTemplate     bool    // True if this is a blueprint
    TemplateID     string  // For instances: source template
    InstanceSuffix string  // Unique suffix (e.g., "abc123")

    // Behavior Configuration
    SpawnRoomID  string         // Respawn location
    RespawnTime  time.Duration  // Time to respawn (0 = no respawn)
    WanderRadius int            // Rooms to wander from spawn
    PatrolPath   []string       // Room IDs for patrol route

    // State Tracking
    IsDead    bool       // Currently dead
    DeathTime time.Time  // When died
    State     string     // FSM: idle, combat, patrol, dead, fleeing

    // Combat state
    InCombat         bool    // Currently in combat instance
    CombatInstanceID string  // Active combat instance

    // Behavior Traits
    EnemyTrait    *EnemyTrait    // Combat behavior
    MerchantTrait *MerchantTrait // Trading behavior (see below)

    // Dialog references
    DialogID          string
    IdleDialogID      string
    IdleDialogTimeout time.Duration
}
```

#### EnemyTrait

```go
type EnemyTrait struct {
    // Classification
    CreatureType string  // beast, humanoid, undead, elemental, construct, demon, dragon, aberration
    CombatStyle  string  // melee, ranged, magic, swarm, brute, agile
    Difficulty   string  // trivial, easy, normal, hard, boss

    // Combat Stats
    AttackPower  int32
    Defense      int32
    AttackSpeed  float64

    // Behavior
    AggroRadius   int     // Detection range (0 = passive)
    AggroOnSight  bool    // Auto-attack on detection
    CallForHelp   bool    // Alert nearby enemies
    FleeThreshold float64 // HP % to flee

    // Rewards
    XPReward       int64
    GoldDrop       Range     // {Min, Max}
    LootTableID    string    // Reference to loot table
    GuaranteedLoot []string  // Item template IDs that always drop
    MaxDrops       int32     // Max items from loot table (0 = unlimited)

    // Event Scripts
    OnAggroScript string
    OnDeathScript string
    OnFleeScript  string
}
```

#### MerchantTrait

```go
type MerchantTrait struct {
    MerchantType   string         // general, blacksmith, alchemist, etc.
    Inventory      []MerchantItem // Items for sale
    RestockMinutes int32          // Restock interval (0 = never)
    LastRestock    time.Time      // Last restock timestamp
    BuyMultiplier  float64        // Price multiplier when buying (1.0 = normal)
    SellMultiplier float64        // Price multiplier when selling (0.5 = half price)
    AcceptedTypes  []string       // Item types merchant will buy (empty = all)
    RejectedTags   []string       // Tags that prevent buying (soulbound, quest)
}

type MerchantItem struct {
    ItemTemplateID string  // Item template to sell
    BasePrice      int64   // Override price (0 = use item's BasePrice)
    PriceOverride  int64   // Force specific price (ignores multipliers)
    Quantity       int32   // Current stock (-1 = unlimited)
    MaxQuantity    int32   // Max after restock
    RequiredLevel  int32   // Player level requirement
}
```

### NPCSpawner Entity

Defines automatic NPC spawning points:

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

### LootTable Entity

Defines item drop tables for NPCs:

```go
type LootTable struct {
    *entities.Entity

    Name           string       // Display name
    Description    string       // Editor description
    Entries        []LootEntry  // Potential drops
    GoldMultiplier float64      // Multiplier for gold drops (1.0 = normal)
    DropBonus      float64      // Flat bonus to all drop chances
}

type LootEntry struct {
    ItemTemplateID string   // Item to drop
    DropChance     float64  // 0.0-1.0 probability
    MinQuantity    int32    // Min items (stackable)
    MaxQuantity    int32    // Max items (stackable)
    Guaranteed     bool     // Always drops
    MinPlayerLevel int32    // Level requirement
    RequiredTags   []string // Player must have tags
}
```

### Item Entity

Items use a unified template/instance pattern similar to NPCs:

```go
type Item struct {
    *entities.Entity
    traits.LookAt

    // Template System (matching NPC pattern)
    IsTemplate     bool    // True if this is a blueprint
    TemplateID     string  // For instances: source template ID
    InstanceSuffix string  // Unique suffix for instance identification

    Name, Description string
    Type    ItemType     // weapon, armor, etc.
    SubType ItemSubType  // sword, shield, etc.
    Slot    ItemSlot     // equipment slot
    Quality ItemQuality  // rarity tier
    Level   int32

    Properties map[string]interface{}
    Attributes map[string]interface{}

    // Container support
    Closed, Locked bool
    Items          Items
    MaxItems       int32

    // Interaction flags
    NoPickup   bool   // Cannot be picked up

    // Stacking and economy
    Stackable bool   // Can stack in inventory
    Quantity  int32  // Current stack count
    MaxStack  int32  // Max stack size (0 = unlimited)
    BasePrice int64  // Default price in gold
}
```

**Template/Instance Lifecycle:**
- Templates (`IsTemplate=true`) are blueprints stored in the database
- Instances (`IsTemplate=false`, `TemplateID` set) are created from templates
- Use `CreateInstanceFromTemplate(templateID)` to spawn new instances
- Instances track their source template via `TemplateID`

## Dialog System

### Dialog Structure

```go
type Dialog struct {
    ID                     string
    Text                   string      // Primary speech
    AlternateTexts         []string    // Random variations
    OrderedTexts           *bool       // Sequential on repeat

    Options                []*Dialog   // Player choices
    Answer                 *Dialog     // Auto-response

    RequiresVisitedDialogs []string    // Prerequisites
    ShowOnlyOnce           *bool       // One-time option
    IsDialogExit           *bool       // End conversation
}
```

### Dialog State

```go
type DialogState struct {
    CurrentDialogID  string              // Active node
    DialogVisited    map[string]int      // Visit counts
    Context          map[string]string   // Static variables
    DynamicContext   map[string]func() string  // Runtime variables
}
```

### Dialog Flow

```
┌─────────────────────────────────────────┐
│              Start: "main"               │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│         Display NPC Text                 │
│   (with variable substitution)           │
└─────────────────┬───────────────────────┘
                  │
        ┌─────────┴─────────┐
        ▼                   ▼
┌───────────────┐   ┌───────────────┐
│  Has Options  │   │  Has Answer   │
│    (choice)   │   │   (auto)      │
└───────┬───────┘   └───────┬───────┘
        │                   │
        ▼                   ▼
┌───────────────┐   ┌───────────────┐
│ Player Selects│   │ Jump to Answer│
│    Option     │   │     Node      │
└───────┬───────┘   └───────────────┘
        │
        ▼
┌───────────────────────────────────────────┐
│ Navigate to selected dialog node          │
│ (check IsDialogExit → return to "main")   │
└───────────────────────────────────────────┘
```

## Message System

### Message Types

```go
const (
    MessageTypeDefault          // Generic message
    MessageTypeEnterRoom        // Room entry
    MessageTypeCreateCharacter  // Character creation prompt
    MessageTypeSelectCharacter  // Character selection
    MessageTypeCharacterSelected // Selection confirmed
    MessageTypePing             // Keep-alive
)
```

### Message Audience

```go
const (
    MessageAudienceOrigin           // To sender only
    MessageAudienceUser             // To specific user
    MessageAudienceRoom             // To all in room
    MessageAudienceRoomWithoutOrigin // To room except sender
    MessageAudienceGlobal           // To all connected
    MessageAudienceSystem           // System announcement
)
```

### Message Structure

```go
// Incoming (Client → Server)
type IncomingMessage struct {
    Message string
}

// Internal Processing
type Message struct {
    FromUser  *entities.User
    Character *characters.Character
    Data      string
}

// Outgoing (Server → Client)
type MessageResponse struct {
    Audience   AudienceType
    AudienceID string
    OriginID   string
    Type       MessageType
    Username   string
    Message    string
}
```

## Frontend Architecture

### Component Hierarchy

```
App.svelte
├── AppContent.svelte (router)
│   ├── Welcome.svelte (landing)
│   ├── Game.svelte (gameplay)
│   │   ├── MUDXPlus.svelte (UI overlay)
│   │   │   └── Inventory.svelte
│   │   └── xterm (terminal)
│   ├── Characters.svelte
│   │   ├── CharacterCard.svelte
│   │   └── CharacterCreator.svelte
│   └── Creator.svelte (editor)
│       ├── RoomsEditor.svelte
│       ├── ItemsEditor.svelte
│       ├── ItemTemplatesEditor.svelte (uses unified items API with isTemplate filter)
│       ├── ScriptsEditor.svelte
│       └── WorldEditor.svelte
├── UserForm.svelte
└── UserMenu.svelte
```

### State Management

| Store | Purpose |
|-------|---------|
| `stores.js` | Global user state, menu state |
| `auth.js` | Auth0 authentication state |
| `MUDXPlusStore.js` | Game UI state (exits, actions, background) |
| `CRUDEditorStore.js` | Editor state (elements, selection, filters) |
| `Client.js` | WebSocket client state (room, character) |

### WebSocket Client

**File:** `public/app/src/game/Client.js`

```javascript
class GameClient {
    ws          // WebSocket connection
    activeRoom  // Current room data
    currentCharacter

    sendMessage(msg)      // Send command to server
    setWSClient(ws)       // Update WebSocket
    onMessage(type, handler)  // Register handler
}
```

### Terminal Integration

Uses xterm.js with custom input handling:

- **LocalEchoController** - Input line editing, history
- **HistoryController** - Command history with up/down navigation
- **FitAddon** - Responsive terminal sizing

## Concurrency Model

### Goroutines

```
main()
  │
  ├── HTTP Server (Gin)
  │
  └── MUD Server
        │
        ├── receiveMessages()      // Game → Clients
        ├── Game.Run()             // Command processing
        ├── handleBroadcast()      // Global messages
        ├── handleTimeouts()       // Ping/keep-alive
        │
        └── Per-client handlers    // WebSocket read loops
```

### Channels

| Channel | Buffer | Purpose |
|---------|--------|---------|
| `OnMessageReceived` | 20 | Player input → Game |
| `SendMessage` | 20 | Game output → MUD Server |
| `OnUserJoined` | 20 | Login events |
| `OnUserQuit` | 20 | Logout events |
| `Broadcast` | unbuffered | Global messages |

### Thread Safety

- `Connection.mu sync.Mutex` - Protects WebSocket writes
- SQLite uses WAL mode and busy timeout for concurrent access
- Channels prevent race conditions on shared state

## Design Patterns

### Patterns Used

| Pattern | Usage |
|---------|-------|
| **Facade** | Service layer (`service.Facade`) |
| **Repository** | Data access (`GenericRepo`) |
| **Factory** | Item templates, character creation |
| **Observer** | Channel-based event system |
| **Command** | Game command handlers |
| **Singleton** | Game state cache |
| **Composition** | Entity traits |
| **State Machine** | Dialog execution |

### Dependency Injection

```
main()
  └── NewApp()
        ├── db.Connect()
        ├── NewFacade(repos...)
        │     ├── NewCharactersRepo(db)
        │     ├── NewRoomsRepo(db)
        │     ├── NewUsersRepo(db)
        │     └── ...
        └── mud.NewServer(facade)
              └── game.NewGame(facade)
```

## Security

### Authentication Flow

```
Client → Auth0 Login → JWT Token
                          │
                          ▼
HTTP Request with JWT → AuthMiddleware
                          │
                          ▼
                    Validate JWT
                    (JWKS lookup)
                          │
                          ▼
                    Find/Create User
                          │
                          ▼
                    Set Context (userid, user)
                          │
                          ▼
                    Handler executes
```

### Authorization

- Protected endpoints require valid JWT
- Admin endpoints require basic auth
- User can only access own characters
- Room/item CRUD available to authenticated users

## Scalability Considerations

### Current Limitations

1. **Single Game Instance** - Game loop runs on single server
2. **In-Memory State** - Game state cache not distributed
3. **Channel Buffers** - Fixed size (20) may need tuning

### Potential Improvements

1. **Clustering** - Distribute game across multiple nodes
2. **Redis State** - Shared state for multi-server
3. **Message Queues** - Replace channels for distribution
4. **Horizontal Scaling** - Stateless API servers

## File Organization

```
pkg/
├── entities/          # Data models
│   ├── characters/    # Player characters
│   ├── rooms/         # Locations
│   ├── items/         # Items and templates
│   ├── npcs/          # Non-player characters
│   ├── dialogs/       # Conversation system
│   ├── skills/        # Character abilities
│   └── traits/        # Shared behaviors
├── mudserver/         # Game server
│   ├── game/          # Game engine
│   │   ├── commands/  # Command handlers
│   │   ├── messages/  # Message types
│   │   └── def/       # Interfaces
│   └── mudserver.go   # WebSocket handling
├── server/            # HTTP API
│   ├── handler/       # Route handlers
│   ├── dto/           # Data transfer objects
│   └── auth.go        # Authentication
├── service/           # Business logic
├── repository/        # Data access
├── db/                # Database client
├── scripts/           # Script execution
│   ├── scripts.go     # Script entity and types
│   ├── scriptrunner.go # Runner interface
│   ├── events/        # Event system
│   │   ├── events.go  # Event types
│   │   ├── context.go # Event context
│   │   └── registry.go # Event handlers
│   └── runner/
│       ├── factory.go # Multi-runner factory
│       ├── defaultscriptrunner.go # JavaScript (deprecated)
│       └── lua/       # Lua runner
│           ├── luarunner.go
│           ├── sandbox.go
│           ├── pool.go
│           └── modules/ # tales.* API
└── util/              # Utilities
```

## Scripting System Architecture

The scripting system uses Lua (via gopher-lua) for dynamic game content. JavaScript support is deprecated but maintained for backward compatibility.

### Script Runner Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      MultiRunner                                 │
│           (Routes to appropriate runner by language)            │
└─────────────────────────────────────────────────────────────────┘
                    │                    │
         ┌──────────┴──────────┐ ┌──────┴─────────┐
         ▼                     ▼ ▼                ▼
┌─────────────────────┐  ┌─────────────────────────────┐
│  DefaultScriptRunner │  │       LuaRunner             │
│     (JavaScript)     │  │  (Primary, recommended)     │
│     [DEPRECATED]     │  └──────────────┬──────────────┘
└─────────────────────┘                   │
                                         ▼
                           ┌──────────────────────────┐
                           │        VM Pool           │
                           │  (Reusable Lua states)   │
                           └──────────────┬───────────┘
                                         │
                           ┌─────────────┼─────────────┐
                           ▼             ▼             ▼
                    ┌──────────┐  ┌──────────┐  ┌──────────┐
                    │ Sandbox  │  │ Modules  │  │ Context  │
                    │ Security │  │tales.*API│  │  Data    │
                    └──────────┘  └──────────┘  └──────────┘
```

### Lua API Modules

| Module | Purpose |
|--------|---------|
| `tales.items` | Item and template operations |
| `tales.rooms` | Room queries and management |
| `tales.characters` | Character operations (damage, heal, teleport) |
| `tales.npcs` | NPC operations (templates, instances, spawning) |
| `tales.dialogs` | Dialog and conversation management |
| `tales.game` | Messaging (room, character, broadcast) |
| `tales.utils` | Utilities (random, UUID, dice rolling) |

#### tales.npcs Functions

| Function | Description |
|----------|-------------|
| `get(id)` | Get NPC by ID (template or persisted) |
| `getTemplates()` | Get all NPC templates |
| `isTemplate(id)` | Check if NPC is a template |
| `spawnFromTemplate(templateId, roomId)` | Create instance in memory |
| `getInstance(id)` | Get instance from NPCInstanceManager |
| `getInstancesInRoom(roomId)` | Get all instances in a room |
| `kill(id)` | Mark instance as dead |
| `setState(id, state)` | Set instance FSM state |
| `getState(id)` | Get instance state |
| `damageInstance(id, amount)` | Apply damage, returns true if died |
| `healInstance(id, amount)` | Restore health |
| `moveInstance(id, roomId)` | Move instance to room |

### Event System

Events allow scripts to respond to game actions:

```
Game Events                 Event Registry              Script Handlers
┌─────────┐                ┌─────────────┐            ┌─────────────────┐
│ player  │ ──dispatch──▶  │  handlers   │ ──execute──▶│ Script 1 (Lua) │
│ entered │                │  by event   │            │ Script 2 (Lua) │
│  room   │                │   type      │            │ ...            │
└─────────┘                └─────────────┘            └─────────────────┘
```

Event types include:
- Player events: `player.enter_room`, `player.leave_room`, `player.join`, `player.quit`
- NPC events: `npc.death`, `npc.spawn`, `npc.idle`
- Item events: `item.pickup`, `item.drop`, `item.use`, `item.create`
- Dialog events: `dialog.start`, `dialog.end`, `dialog.option`
- Room events: `room.action`, `room.update`
- Quest events: `quest.start`, `quest.complete`, `quest.progress`

See [SCRIPTING.md](SCRIPTING.md) for full documentation.
