# TalesMUD Project Documentation

## Overview

TalesMUD is a browser-based Multi-User Dungeon (MUD) framework built with Go and Svelte. It provides a complete platform for creating and playing text-based multiplayer adventure games, featuring real-time WebSocket communication, a web-based content editor, and persistent game state via SQLite.

**Repository:** [github.com/TalesMUD/talesmud](https://github.com/TalesMUD/talesmud)

## Documentation Index

- **Architecture:** `ARCHITECTURE.md`
- **Core Systems & Features:** `FEATURES.md` (comprehensive reference for all systems, data structures, and APIs)
- **Game design + MVP backlog:** `docs/design/GAME_DESIGN.md`
- **Scripting system:** `docs/design/SCRIPTING.md`
- **World map implementation:** `docs/design/WORLD_MAP_IMPLEMENTATION.md`
- **Quest authoring guide:** `docs/design/QUEST_AUTHORING.md`
- **Player guide:** `docs/player-guide/`
- **Development docs:** `docs/development/`

## MVP Roadmap (next up)

Planned epics (see `game-design/GAME_DESIGN.md`):

- Enemy NPCs + combat
- Combat instances (ad-hoc rooms)
- Items/loot/containers
- Inventory + equip/unequip
- Merchants/trading

## Features

### Core Game Features

- **Room-Based World System**
  - Interconnected rooms with customizable exits (directional, named, teleport)
  - Hidden/secret exits (toggleable visibility in editor)
  - Room actions for custom player interactions (respond, broadcast, run script)
  - Room action names match case-insensitively and beat global `examine`/`take`/`use` when they collide
  - Movement exits match case-insensitively; room presence fan-out does not block the game loop on SQLite
  - Response actions send the narrative `response` text, not the help `description`
  - Action descriptions shown in room text ("You can:" section)
  - Visual backgrounds and mood settings
  - Coordinate-based world mapping (X, Y, Z grid)
  - Dynamic item and NPC spawning
  - Unique NPCs auto-spawn into their assigned room on server start via `CurrentRoomID`

- **Character System**
  - Full RPG character creation with races and classes
  - Six-attribute system (STR, DEX, CON, INT, WIS, CHA)
  - Equipment system with 10 equipment slots
  - Inventory management
  - Experience and leveling with flattened early-game XP curve (piecewise formula: gentle L2-5, transitional L6-15, steeper L16+)
  - Exploration XP: awards 5 XP per new room discovered, 15 XP for first room in a new area/zone
  - **Distributable Attribute Points**: 2 points per level-up for players to allocate freely into STR, DEX, INT, WIS, or STA
  - Class-based attribute caps prevent degenerate builds (e.g., warriors cap INT at 5, wizards cap STR at 5)
  - Terminal command `spend <attr> [amount]` to allocate points; `spend` with no args shows status table with current values, spent/cap per attribute
  - Character widget shows unspent points badge and interactive "+" buttons on each attribute when points are available
  - **Derived Combat Stats Display**: Character widget shows computed ATK (weapon damage + STR modifier), DEF (total armor from equipment), and MP/RND (mana regen per combat round, caster classes only). These update live when equipment or attributes change.
  - Existing characters receive retroactive points on login ((level - 1) * 2)
  - Server-side room/area discovery tracking per character
  - All-time statistics tracking (including rooms discovered)
  - Mana system for caster classes (Mage, Cleric, Druid) with level and INT scaling
  - Mana regeneration: out-of-combat (5%/tick passive, 15%/tick resting), in-combat (1+WISMod per round)
  - Mana potions (Small/Medium/Large) as consumable items

- **Skills & Spells System**
  - Database-stored skills, editable via Creator UI (Skills tab)
  - Multi-class support: skills can be assigned to multiple classes (e.g., Heal for Cleric and Druid)
  - 29 default abilities across 6 classes, seeded on first run
  - Two resource types: mana-based (casters) and cooldown-based (physical classes)
  - Equippable skill slots (1-4 per class, level-gated progression)
  - Skill management: equip/unequip outside combat, locked during combat
  - Skill effects: damage, heal, buff, debuff, DoT, HoT, stun, multi-hit, AoE
  - Status effect system: buffs, debuffs, DoTs, HoTs with duration tracking
  - Attribute-scaled damage: STR (warrior), DEX (rogue/ranger), INT (mage/druid), WIS (cleric)
  - Mana shield absorption mechanic
  - Skill cooldown tracking per combat instance
  - Combat commands clear stale character combat flags when no live combat instance exists
  - In-combat commands: `cast <skill> [target]`, numeric shortcuts `1`-`4`
  - Management commands: `skills`, `skills equip <name>`, `skills unequip <name>`
  - YAML import/export for skills data

- **Item System**
  - Multiple item types: Currency, Consumable, Armor, Weapon, Collectible, Quest, Crafting Material
  - Quality tiers: Normal, Magic, Rare, Legendary, Mythic
  - Item templates for reusable definitions
  - Stackable consumables and partial drops persist reduced quantities consistently between character inventory and item instances
  - Container support with nested items

- **Quest System**
  - Data-driven quest definitions with multiple objective types: Kill, Collect, Deliver, Visit, Talk, Custom (Lua)
  - Quest progress tracking per character with persistent state
  - NPC dialog integration: automatic quest offer/turn-in options injected into NPC conversations, including quest-only NPCs without full dialog trees
  - Real-time quest log WebSocket updates include quest definition details, objectives, and rewards after dialog quest actions
  - Quest rewards: XP, Gold, and item grants on completion
  - Quest prerequisites: required quest completions and level requirements
  - Repeatable quests support
  - Quest categories and area labels for filtering and organizing regional quest lines
  - QuestTracker: automatic progress updates from game events (NPC kills, item pickups with stack quantities, room entries, dialog nodes, NPC delivery checks)
  - Accepting a collect quest pre-fills progress from matching items already in the character inventory, including stack quantities
  - Delivery objectives require and consume matching inventory items before progress is granted
  - Player quest log shows enriched objective descriptions, ready-to-turn-in state, and quest notifications for accept/progress/ready/complete events
  - Lua scripting API (`tales.quests`) for custom quest logic
  - Creator UI: full quest editor with objectives, rewards, prerequisites, dialog text configuration, validation feedback, and player flow preview
  - Player commands: `quests`/`ql` (quest log), `quest <name>` (details), `abandon <name>` (abandon quest)

- **Guest Mode (Play as Guest)**
  - Anonymous 30-minute demo sessions without Auth0 registration
  - "Play as Guest" button on welcome screen
  - Random character with random class from system templates
  - Spawns in `ServerSettings.StartRoomID` (default `R0001` when that room exists)
  - Auto-grants `source.type: auto` quests for the start room's zone (Z00 catacombs: QST0001–QST0004)
  - Full starter items equipped automatically
  - Per-character level cap of 5 for guest characters
  - Full chat access during session
  - 5-minute warning before session expiry
  - Auto-deletion of guest user + character after session ends or disconnect (5-min grace period for reconnection)
  - Server-configurable: `GuestsAllowed` (default: true), `MaxGuestAccounts` (default: 20)
  - IP-based rate limiting (10 guest sessions per IP per hour)
  - HMAC-SHA256 guest tokens (separate from Auth0 JWTs), signed with `GUEST_SECRET` env var
  - Background cleanup goroutine removes expired guest accounts every 5 minutes

- **New Player Onboarding**
  - Phase-based flow: Welcome Screen, Nickname Setup, Character Creation Wizard, Game
  - Unauthenticated users see a cinematic welcome landing screen (not the game UI)
  - Signup and Login via Auth0 with dedicated CTA buttons
  - "Play as Guest" option for anonymous demo play
  - First-time users prompted to choose a display name/nickname
  - Three-step character creation wizard: Choose Template, Name Character, Confirm & Create
  - Automatic phase detection from user profile and character data
  - Guest users skip onboarding (character auto-created server-side)

- **Room Text Overlay**
  - Game text (combat, actions, player messages) displayed as translucent overlay on room image
  - Auto-dismiss with duration scaled by text length (2-4 seconds)
  - Smooth fade-in/fade-out animations
  - Stacks up to 5 messages during rapid sequences (e.g. combat)
  - Always enabled on mobile; optional toggle for desktop in Settings > Interface

- **Multiplayer**
  - Real-time player interactions via WebSocket
  - Players see each other in rooms
  - Global and room-based chat
  - Private tells/whispers and minimal party chat
  - Party flow: create parties, invite online players, accept/decline invites, list members, leave party, and send party chat
  - Emote system
  - Live session-based player presence tracking for room UI, chat routing, `who`, and silent room presence refreshes
  - Reconnect-aware client state with visible connecting/reconnecting status and automatic reconnect attempts
  - In-game character switcher for changing active characters without leaving the play UI

### Content Creation

- **Web-Based Editor**
  - Full-width filterable data tables for browsing all entity types (Rooms, Items, Item Templates, NPCs, Dialogs, Quests, Skills, Scripts, Character Templates)
  - Per-column filtering (text search, enum dropdowns) with instant client-side filtering and sorting
  - Side-by-side master-detail layout: data table + edit form shown together, closeable to full-width table view
  - **Entity Selection Modal**: All entity ID selectors (rooms, NPCs, items, scripts, dialogs, quests, character template starting items) use a centered modal dialog with a full filterable DataTable instead of simple dropdowns. This scales to hundreds of entries with per-column search, sort, and filter support. Components: `EntitySelectButton` (inline trigger) + `EntitySelectModal` (table dialog). **UI Guideline: Never use `<select>` dropdowns for entity ID references. Always use `EntitySelectButton` with the appropriate column definitions from `tableColumns.js`.**
  - Room editor with exit, action, spawner, items, and NPC resident configuration
  - Item and item template management with attributes and properties
  - NPC editor with behavior controls for state, spawn room, wander radius, patrol paths, idle chatter, enemy traits, and merchant traits
  - Lua script editor with syntax highlighting and integrated test runner
  - Dialog tree editor with options and alternate texts
  - Quest editor with validation and player flow preview for source, objectives, turn-in, and rewards
  - Character template editor with archetype selection and starting gear
  - Skills editor with multi-class assignment, resource types, effects, and secondary effects
  - World map visualization (GridWorldEditor)
  - World Health diagnostics for broken cross-system references across rooms, NPCs, items, loot tables, quests, dialogs, scripts, spawners, and character template starting gear
  - Creator quality validation: inline warnings/errors, broken-reference detection, save blocking for invalid references, and a world health diagnostics tab
  - Preview/test tools for dialogs, quests, rooms, merchants, and Lua scripts
  - CRUD operations with live preview

- **Scripting System**
  - Lua-based scripting via gopher-lua (primary)
  - Room action scripts (type "script" triggers Lua execution with room.action context)
  - Room on-enter scripts
  - Item behavior scripts
  - NPC behavior scripts
  - Quest scripting support
  - Game API: messaging, inventory checks (`hasItem`, `hasEquipped`), character flags (`getFlag`/`setFlag`), item rewards (`giveItem`), per-character exit reveals (`revealExit`)

### AI / LLM Integration

- **Groq API Integration**
  - Reusable Groq LLM client (`pkg/service/groq/`) for AI-powered text generation
  - Character creation: AI-generated names and descriptions based on selected template (archetype, race, class, backstory)
  - Protected API endpoint: `POST /api/generate/character`
  - Graceful degradation when `GROQ_API_KEY` is not configured
  - Designed for extension to other AI features (NPC dialogue, room descriptions, quest generation)

### Technical Features

- **Authentication & Authorization**
  - Auth0 OAuth2 integration
  - JWT-based API protection
  - Guest mode with HMAC-SHA256 tokens (no Auth0 required)
  - Dual auth middleware: tries guest token first, falls back to Auth0 JWT
  - Frontend session state avoids logging or retaining auth token excerpts
  - Basic auth for legacy admin endpoints (export/import), with explicit credentials required and insecure release defaults rejected
  - Session management
  - Three-tier role system: MUD Admin, MUD Creator, Player
  - MUD Admin configured via `MUD_ADMIN_OAUTHID` env var (has full access)
  - MUD Creators can modify game content (Creator area)
  - Players can play the game and access only their own characters and quest progress
  - User ban system (bans by Reference ID and email)

- **User Management (Admin Only)**
  - View all registered users with ID, Name, Nickname, Email, Access Level
  - Promote players to Creator role or demote Creators to Player
  - Ban/unban players with double-confirmation modal
  - Banned users are blocked from all authenticated endpoints

- **Optional Landing Page**
  - Serve a static landing page at `/` from the OS filesystem via `LANDING_PATH`
  - When disabled (default), the main app SPA is served at `/` as usual
  - Auth0 callbacks (`?code=` / `?error=`) pass through to the main SPA automatically
  - Static assets (images, CSS) in the landing directory are served alongside `index.html`

- **Data Persistence**
  - SQLite for all game data
  - World export/import functionality
  - YAML/JSON data file support

## Technology Stack

### Backend
| Component | Technology |
|-----------|------------|
| Language | Go 1.18 |
| HTTP Framework | Gin |
| Database | SQLite |
| WebSocket | Gorilla WebSocket |
| Authentication | Auth0 JWT |
| Scripting | Otto (JavaScript VM) |
| AI / LLM | Groq API (llama-3.3-70b) |
| Logging | Logrus |

### Frontend
| Component | Technology |
|-----------|------------|
| Framework | Svelte 3.59 |
| UI Library | Materialize CSS |
| Terminal | xterm.js |
| HTTP Client | Axios |
| Router | yrv |
| Build Tool | Rollup |

### Infrastructure
| Component | Technology |
|-----------|------------|
| Container | Docker |
| Orchestration | Docker Compose |
| CI/CD | GitHub Actions |
| Database | SQLite |

## Project Structure

```
talesmud/
├── cmd/                    # Application entry points
│   ├── tales/              # Main server
│   └── dialog_sandbox/     # Dialog testing tool
├── pkg/                    # Go packages
│   ├── entities/           # Data models (characters, rooms, items, NPCs, dialogs)
│   ├── mudserver/          # Game server (WebSocket, game loop, commands)
│   ├── server/             # HTTP API server
│   ├── service/            # Business logic layer
│   ├── repository/         # Data access layer
│   ├── db/                 # Database utilities (SQLite)
│   └── scripts/            # Script execution engine
├── public/                 # Frontend
│   └── app/
│       └── src/            # Svelte source
│           ├── game/       # Game client
│           ├── creator/    # Content editor
│           ├── characters/ # Character management
│           └── api/        # API clients
├── api/                    # API test files & sample data
├── data/                   # Sample game data
└── bin/                    # Compiled binaries
```

## Game Commands

| Command | Aliases | Description |
|---------|---------|-------------|
| `north`, `south`, `east`, `west` | `n`, `s`, `e`, `w` | Move between rooms |
| `look` | `l` | Examine current room |
| `inventory` | `i` | Display inventory |
| `selectcharacter` | `sc` | Select active character |
| `listcharacters` | `lc` | List your characters |
| `newcharacter` | `nc` | Create new character |
| `who` | - | List online players |
| `party create` | `p create` | Create a party with your current character |
| `party invite <player>` | `p invite <player>` | Invite an online player to a party |
| `party accept` / `party decline` | - | Respond to a pending party invite |
| `party list` | - | List party members |
| `party leave` | - | Leave the current party |
| `party say <message>` / `party <message>` | `p say <message>` / `p <message>` | Send party chat |
| `scream` | - | Broadcast to room |
| `shrug` | - | Emote action |
| `help` | `h` | Show help |
| `attack` | `a`, `hit` | Attack a target / switch combat target |
| `defend` | `d`, `guard` | Queue defensive stance for next combat turn |
| `flee` | `run`, `escape` | Queue flee attempt for next combat turn |
| `status` | `cs`, `combat` | Show combat status |
| `cast` | `spell` | Use a skill in combat: cast \<skill\> [target] |
| `skills` | `spells`, `abilities` | Manage skills: skills [equip\|unequip] [name] |
| `quests` | `ql`, `questlog` | Show quest log |
| `quest` | - | Show quest details: quest [name] |
| `abandon` | - | Abandon a quest: abandon [name] |
| `spend` | - | Spend attribute points: spend \<attr\> [amount] |
| `pickup` | `get`, `take` | Pick up an item from the room (room-placed catalog items copy per character and stay for the next guest; dropped loot is still taken) |
| `drop` | - | Drop an item to the room (blocked for bound items) |
| `destroy` | `discard` | Destroy an item from inventory |
| `examine` | `inspect` | Examine an item in detail |
| `use` | `eat`, `drink`, `consume` | Use a consumable item |
| `equip` | `wear` | Equip an item |
| `unequip` | `remove` | Unequip an item |
| `equipment` | `eq`, `gear` | Show equipped items |
| `list` | `shop`, `trade` | List merchant inventory |
| `buy` | - | Buy from merchant; stackable quantities fit in a single stack when possible |
| `sell` | - | Sell to merchant (blocked for bound items and unsupported item types) |
| `value` | `price` | Check item sell price |

## Current Development Status

### Branch: NPCs (Active Development)

The NPCs branch represents the latest development work, focusing on NPC systems and player-NPC interactions.

#### Completed Features

1. **NPC Entity System**
   - Core NPC data structure mirroring player characters
   - Trait-based composition (DialogTrait, MerchantTrait, EnemyTrait)
   - Room integration with NPC presence tracking
   - Health, level, and class systems
   - Runtime NPC behavior loop for idle, patrol, dead/respawn, and combat states
   - Deterministic patrol paths and bounded wandering from spawn rooms

2. **Dialog Engine**
   - Full dialog tree system with branching conversations
   - State management tracking visited dialogs
   - Template rendering with dynamic variables ({{PLAYER}}, {{NPC}}, {{TIME}})
   - Conditional option display based on conversation history
   - Alternate text variations for natural dialogue
   - Ordered responses (different text on repeated visits)
   - Dialog sandbox for testing conversations

3. **Dialog Features**
   - Interactive dialogs (triggered by player interaction)
   - Idle dialogs (ambient NPC chatter with timeout)
   - Show-once options
   - Dialog exit markers
   - YAML serialization for dialog definitions

4. **Auto-Attack Combat System**
   - Automatic combat rounds (players and NPCs auto-attack each turn)
   - Turn-order initiative system with auto-processing
   - Players can queue special actions between auto-attacks: target switch, defend, flee
   - Combat starts with `attack`/`kill` and proceeds automatically
   - No turn timeouts or AFK mechanics needed

5. **NPC Behavior and Quest Interaction**
   - NPC update loop handles idle wandering, ordered patrol paths, respawn cleanup, and idle chatter cooldowns
   - `talk` and `speak` open NPC dialogs and inject quest offer/progress/turn-in options
   - Quest-giving NPCs without a main dialog open quest-only conversations for numbered quest choices
   - MUD client NPC cards show enemy, merchant, quest giver, dialog, idle chatter, and current state badges
   - Creator NPC editor exposes behavior controls and uses modal entity selectors for patrol and room references

### Recent Commits (NPCs Branch)

| Commit | Description |
|--------|-------------|
| `b17856d` | Fixed Svelte issues |
| `6b621a2` | Huge improvements on player and NPC interaction |
| `29674d5` | New work on Dialogs |
| `b54c92e` | Further work on dialogs |
| `b5ae2c3` | More progress on dialog logic |

## Configuration

### Environment Variables (.env)

```bash
# Server Configuration
GIN_MODE=debug
PORT=8010

# Optional comma-separated CORS origins in addition to the built-in production
# domains and local dev origins for the Creator and MUD clients.
CORS_ALLOWED_ORIGINS=

# SQLite database path
SQLITE_PATH=./talesmud.db

# Auth0
AUTH0_AUDIENCE=http://talesofapirate.com/dnd/api
AUTH0_DOMAIN=https://owndnd.eu.auth0.com/
AUTH0_WK_JWKS=https://owndnd.eu.auth0.com/.well-known/jwks.json
AUTH_ENABLED=false

# Admin (basic auth for legacy export/import)
# Leave either value blank to disable these endpoints.
# admin/admin is rejected in release mode.
ADMIN_USER=admin
ADMIN_PASSWORD=changeme

# MUD Admin OAuth ID (Auth0 sub claim, e.g. "twitter|16651340")
# The user with this OAuth ID gets full admin access
MUD_ADMIN_OAUTHID=

# Guest mode secret key for signing guest JWTs (HMAC-SHA256)
# If not set, a random key is generated at startup (guest tokens won't survive server restart)
GUEST_SECRET=

# Optional landing page (path to directory with index.html + static assets)
# LANDING_PATH=./public/landing

# AI Generation (Groq API) — used for character name/description generation
# Get a key at https://console.groq.com
GROQ_API_KEY=
```

## Building & Running

### Prerequisites
- Go 1.24+
- Node.js (for frontend build)

### Build Commands

```bash
# Build everything
make build

# Build frontend only
make build-frontend

# Build backend only
make build-backend

# Backend-only builds and server runs prepare fallback embedded frontend assets
# if pkg/webui/dist or pkg/webuiplay/dist has not been generated yet. Run
# `make build` to embed freshly rebuilt frontend bundles.

# Run the server
make run-server

# Run the server with SQLite (single binary + embedded frontend)
DB_DRIVER=sqlite SQLITE_PATH=./talesmud.db ./bin/tales

# Run frontend dev server
make run-frontend

# Run dialog sandbox
make run-dialogs-sandbox
```

### Docker Deployment

```bash
# Start with Docker Compose
docker-compose up -d
```

### Data Import

Import world data into SQLite:

```bash
go run cmd/migrate/main.go -input export.json -sqlite talesmud.db
```

## API Endpoints

### Public Endpoints
- `GET /health` - Health check
- `GET /api/templates/characters` - Character creation templates
- `GET /api/room-of-the-day` - Featured room
- `POST /api/guest` - Create guest session (returns HMAC token)
- `GET /api/server-info` - Public server info (guest mode status)

### Protected Endpoints (Require Auth - Player Level)
- `GET /api/characters`, `POST /api/newcharacter` - Character management; direct character object access is owner/admin only
- `POST /api/generate/character` - AI-powered character name/description generation
- `GET /api/rooms`, `GET /api/items`, `GET /api/skills` - Read game data
- `GET /api/user`, `PUT /api/user` - User profile

### Protected Endpoints (Player Level - Quests)
- `GET /api/quests` - List all quest definitions
- `GET /api/quests/:id` - Get quest by ID
- `GET /api/quest-progress/:characterId` - Get character quest log for own/admin character
- `POST /api/quest-progress/:characterId/accept/:questId` - Accept quest for own/admin character
- `POST /api/quest-progress/:characterId/abandon/:questId` - Abandon quest for own/admin character
- `POST /api/quest-progress/:characterId/complete/:questId` - Complete a ready quest for own/admin character

### Creator Endpoints (Require Creator or Admin Role)
- `POST/PUT/DELETE /api/rooms` - Room management
- `POST/PUT/DELETE /api/items` - Item management
- `POST/PUT/DELETE /api/scripts` - Script management
- `POST/PUT/DELETE /api/npcs` - NPC management
- `POST/PUT/DELETE /api/dialogs` - Dialog management
- `POST/PUT/DELETE /api/quests` - Quest management
- `POST/PUT/DELETE /api/skills` - Skill management
- `GET /api/world/validation` - World Health diagnostics
- `GET /api/diagnostics/world` - World health diagnostics across rooms, NPCs, dialogs, quests, loot, items, and scripts
- `POST /api/validate/:entityType` - Validate a draft Creator entity before save
- `POST /api/preview/dialog`, `/api/preview/quest`, `/api/preview/room`, `/api/preview/merchant` - Preview/test draft content with validation issues
- `PUT /api/settings` - Server settings

### Admin API Endpoints (Require Admin Role)
- `GET /api/admin/users` - List all users
- `PUT /api/admin/users/:id/role` - Change user role
- `POST /api/admin/users/:id/ban` - Ban user
- `POST /api/admin/users/:id/unban` - Unban user

### Legacy Admin Endpoints (Basic Auth)
- `GET /admin/export` - Export world data; requires explicit `ADMIN_USER` and `ADMIN_PASSWORD`
- `POST /admin/import` - Import world data; validates JSON before replacing stored data
- `GET /admin/world` - World map rendering

### WebSocket
- `GET /ws` - Game connection (authenticated)

## File Statistics

| Category | Count |
|----------|-------|
| Go source files | 86 |
| Svelte components | ~319 |
| JavaScript files | 23 |
| Total backend code | ~484KB |
| Total frontend code | ~344KB |

## License

See LICENSE file for details.

## Contributing

This project is actively developed. The NPCs branch contains the latest work on NPC systems and dialog interactions.

### Development Workflow
1. Fork the repository
2. Create a feature branch from `NPCs` (current active branch)
3. Make changes following existing patterns
4. Test with dialog sandbox for NPC-related changes
5. Submit pull request

## Related Resources

- [MUD Wikipedia](https://en.wikipedia.org/wiki/MUD) - Background on Multi-User Dungeons
- [Go Documentation](https://golang.org/doc/) - Go language reference
- [Svelte Tutorial](https://svelte.dev/tutorial) - Svelte framework guide
- [SQLite Documentation](https://www.sqlite.org/docs.html) - Database documentation
