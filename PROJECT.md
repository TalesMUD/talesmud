# TalesMUD Project Documentation

## Overview

TalesMUD is a browser-based Multi-User Dungeon (MUD) framework built with Go and Svelte. It provides a complete platform for creating and playing text-based multiplayer adventure games, featuring real-time WebSocket communication, a web-based content editor, and persistent game state via MongoDB.

**Repository:** [github.com/TalesMUD/talesmud](https://github.com/TalesMUD/talesmud)

## Features

### Core Game Features

- **Room-Based World System**
  - Interconnected rooms with customizable exits (directional, named, teleport)
  - Room actions for custom player interactions
  - Visual backgrounds and mood settings
  - Coordinate-based world mapping (X, Y, Z grid)
  - Dynamic item and NPC spawning

- **Character System**
  - Full RPG character creation with races and classes
  - Six-attribute system (STR, DEX, CON, INT, WIS, CHA)
  - Equipment system with 10 equipment slots
  - Inventory management
  - Experience and leveling
  - All-time statistics tracking

- **Item System**
  - Multiple item types: Currency, Consumable, Armor, Weapon, Collectible, Quest, Crafting Material
  - Quality tiers: Normal, Magic, Rare, Legendary, Mythic
  - Item templates for reusable definitions
  - Container support with nested items

- **Multiplayer**
  - Real-time player interactions via WebSocket
  - Players see each other in rooms
  - Global and room-based chat
  - Emote system
  - Player presence tracking

### Content Creation

- **Web-Based Editor**
  - Room editor with exit and action configuration
  - Item and item template management
  - Script editor for custom logic
  - World map visualization
  - CRUD operations with live preview

- **Scripting System**
  - JavaScript-based scripting via Otto VM
  - Room action scripts
  - Item behavior scripts
  - NPC behavior scripts
  - Quest scripting support

### Technical Features

- **Authentication**
  - Auth0 OAuth2 integration
  - JWT-based API protection
  - Basic auth for admin endpoints
  - Session management

- **Data Persistence**
  - MongoDB for all game data
  - World export/import functionality
  - YAML/JSON data file support

## Technology Stack

### Backend
| Component | Technology |
|-----------|------------|
| Language | Go 1.18 |
| HTTP Framework | Gin |
| Database | MongoDB |
| WebSocket | Gorilla WebSocket |
| Authentication | Auth0 JWT |
| Scripting | Otto (JavaScript VM) |
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
| Database | MongoDB 4.2.6 |

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
│   ├── db/                 # MongoDB client
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
| `scream` | - | Broadcast to room |
| `shrug` | - | Emote action |
| `help` | `h` | Show help |

## Current Development Status

### Branch: NPCs (Active Development)

The NPCs branch represents the latest development work, focusing on NPC systems and player-NPC interactions.

#### Completed Features

1. **NPC Entity System**
   - Core NPC data structure mirroring player characters
   - Trait-based composition (DialogTrait, MerchantTrait, EnemyTrait)
   - Room integration with NPC presence tracking
   - Health, level, and class systems

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

#### In Progress

- Integration of dialog system into game commands
- NPC behavior loop in game update cycle
- Talk/speak command implementation
- Frontend dialog UI

#### Planned Features

- Merchant trait trading system
- Enemy trait combat system
- NPC movement and patrol paths
- Quest-giving NPCs

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

# MongoDB
MONGODB_CONNECTION_STRING=mongodb://localhost:27017
MONGODB_DATABASE=talesmud
MONGODB_USER=talesmud
MONGODB_PASSWORD=talesmud

# Auth0
AUTH0_AUDIENCE=http://talesofapirate.com/dnd/api
AUTH0_DOMAIN=https://owndnd.eu.auth0.com/
AUTH0_WK_JWKS=https://owndnd.eu.auth0.com/.well-known/jwks.json
AUTH_ENABLED=false

# Admin
ADMIN_USER=admin
ADMIN_PASSWORD=admin
```

## Building & Running

### Prerequisites
- Go 1.18+
- Node.js (for frontend build)
- MongoDB 4.2+

### Build Commands

```bash
# Build everything
make build

# Build frontend only
make build-frontend

# Build backend only
make build-backend

# Run the server
make run-server

# Run frontend dev server
make run-frontend

# Run dialog sandbox
make run-dialogs-sandbox
```

### Docker Deployment

```bash
# Start with Docker Compose (includes MongoDB)
docker-compose up -d
```

## API Endpoints

### Public Endpoints
- `GET /health` - Health check
- `GET /api/templates/characters` - Character creation templates
- `GET /api/room-of-the-day` - Featured room

### Protected Endpoints (Require Auth)
- `GET/POST/PUT/DELETE /api/characters` - Character management
- `GET/POST/PUT/DELETE /api/rooms` - Room management
- `GET/POST/PUT/DELETE /api/items` - Item management
- `GET/POST/PUT/DELETE /api/item-templates` - Item template management
- `GET/POST/PUT/DELETE /api/scripts` - Script management
- `GET /api/user` - User profile

### Admin Endpoints (Basic Auth)
- `GET /admin/export` - Export world data
- `POST /admin/import` - Import world data
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
- [MongoDB Manual](https://docs.mongodb.com/manual/) - Database documentation
