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
│              (MongoDB or SQLite backends)                        │
└─────────────────────────────────────────────────────────────────┘
                             │
              ┌──────────────┴──────────────┐
              ▼                             ▼
┌───────────────────────────┐  ┌───────────────────────────────┐
│          MongoDB           │  │            SQLite             │
│  (collections by entity)   │  │ (tables with JSON payloads)   │
└───────────────────────────┘  └───────────────────────────────┘
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

#### Database Driver Selection

`server.NewApp()` selects the storage backend at startup:

- `DB_DRIVER=mongo` → MongoDB repositories
- `DB_DRIVER=sqlite` → SQLite repositories (single-file DB)

When `DB_DRIVER=sqlite`, use `SQLITE_PATH` to point to the DB file.

#### Route Organization

```
/                          # Static files (frontend)
/health                    # Health check
/ws                        # WebSocket (game connection)
/api/
    ├── characters/        # Character CRUD
    ├── rooms/             # Room CRUD
    ├── items/             # Item CRUD
    ├── item-templates/    # Item template CRUD
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

#### Update Cycle (Every 10 seconds)

1. **Room Updates**
   - Remove offline characters from rooms
   - Sync room state to database
   - Clean up stale connections

2. **NPC Updates** (Planned)
   - Process NPC behaviors
   - Execute idle dialogs
   - Handle NPC movement

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

### Repository Layer (`pkg/repository/`)

Data access layer with backend-specific implementations (MongoDB or SQLite) selected at startup.

#### Generic Repository

```go
type GenericRepo struct {
    db         *db.Client
    collection string
    generator  func() interface{}  // Entity factory
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

- **MongoDB**: wrapper client and BSON query builder.
- **SQLite**: JSON document storage (one row per entity) with JSON1 queries, WAL mode, and busy timeout.

#### Query Parameter Builder

```go
NewQueryParams().
    With(QueryParam{Key: "area", Value: "oldtown"}).
    With(QueryParam{Key: "roomType", Value: "tavern"}).
    AsBSON()
```

#### Collections

| Collection | Entity |
|------------|--------|
| users | User accounts |
| characters | Player characters |
| rooms | Game locations |
| items | Item instances |
| itemtemplates | Item blueprints |
| npc | Non-player characters |
| scripts | Game scripts |
| parties | Player groups |

## Entity Model

### Base Entity

```go
type Entity struct {
    _ID primitive.ObjectID `bson:"_id,omitempty"`  // MongoDB ID
    ID  string             `bson:"id"`             // UUID
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

    Inventory     items.Inventory
    EquippedItems map[items.ItemSlot]*items.Item
    AllTimeStats
}
```

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

    EnemyTrait *EnemyTrait  // Combat behavior
}
```

### Item Entity

```go
type Item struct {
    *entities.Entity
    traits.LookAt

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
}
```

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
│       ├── ItemTemplatesEditor.svelte
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
- MongoDB driver handles internal synchronization
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
└── util/              # Utilities
```
