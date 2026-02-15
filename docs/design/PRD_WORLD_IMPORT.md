# PRD: World Reset & Import System

## Overview

This document describes the implementation of a **World Reset & Import** feature that allows administrators to reset all game content and import fresh world data from YAML configuration files stored in the `import/` folder.

## Problem Statement

During MVP development, game content (rooms, NPCs, items, dialogs, scripts) is authored in a separate GitHub repository as YAML files. We need a mechanism to:

1. Reset the game world to a clean state
2. Import fresh content from YAML configuration files
3. Preserve player accounts while resetting gameplay state
4. Move existing characters to the starting room

## Goals

- **G1**: Provide an API endpoint to trigger world reset and import
- **G2**: Read and parse YAML files from the `import/` folder
- **G3**: Preserve user accounts (OAuth bindings) during reset
- **G4**: Reset character locations to the starting room
- **G5**: Enable triggering from Creator UI with confirmation dialog
- **G6**: Provide clear success/error feedback

## Non-Goals

- Automatic scheduled resets (manual trigger only)
- Partial imports (always full world replacement)
- Character data migration between world versions
- Rollback capability (out of scope for MVP)

---

## Current Architecture

### Database Tables

| Table | Type | Reset Behavior |
|-------|------|----------------|
| `users` | Player Data | **PRESERVE** |
| `characters` | Player Data | **RESET LOCATION** |
| `rooms` | Game Content | DROP & IMPORT |
| `items` | Game Content | DROP & IMPORT |
| `npcs` | Game Content | DROP & IMPORT |
| `npc_spawners` | Game Content | DROP & IMPORT |
| `dialogs` | Game Content | DROP & IMPORT |
| `scripts` | Game Content | DROP & IMPORT |
| `loot_tables` | Game Content | DROP & IMPORT |
| `charactertemplates` | Game Content | DROP & IMPORT |
| `conversations` | Session Data | DROP |
| `parties` | Session Data | DROP |

### Existing Infrastructure

- **Export Handler**: `pkg/server/handler/export.go` - existing `/admin/export` and `/admin/import` endpoints
- **Exporter Data Structure**: `pkg/exporter/data.go` - defines the data format
- **Repository Pattern**: All repositories implement `Drop()` and `Import()` methods
- **Admin Auth**: Basic HTTP Auth via `ADMIN_USER` and `ADMIN_PASSWORD` env vars

### Import Folder Structure

```
import/
└── mvp-rpg-1/
    ├── rooms/
    │   ├── R0001.yaml
    │   ├── R0002.yaml
    │   └── ...
    ├── items/
    │   ├── ITM0001.yaml
    │   └── ...
    ├── npcs/
    │   ├── NPC0001.yaml
    │   └── ...
    ├── dialogs/
    │   ├── DLG0001_mira_thornwood.yaml
    │   └── ...
    ├── scripts/
    │   ├── SCR0001_awakening_enter.yaml
    │   └── ...
    ├── loot_tables/
    │   ├── LT0001_catacomb_rat.yaml
    │   └── ...
    ├── npc_spawners/
    │   └── ...
    └── character_templates/
        └── ...
```

---

## Proposed Solution

### New API Endpoint

```
POST /admin/reset-world
Content-Type: application/json
Authorization: Basic <credentials>

Request Body:
{
  "folder": "mvp-rpg-1",
  "startRoomId": "R0001",
  "clearInventories": false
}
```

### Request Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `folder` | string | Yes | - | Subfolder name under `import/` |
| `startRoomId` | string | Yes | - | Room ID to move all characters to |
| `clearInventories` | bool | No | `false` | If true, clear all character inventories |

### Response

**Success (200 OK)**:
```json
{
  "success": true,
  "message": "World reset complete",
  "stats": {
    "roomsImported": 45,
    "itemsImported": 120,
    "npcsImported": 32,
    "dialogsImported": 28,
    "scriptsImported": 15,
    "lootTablesImported": 8,
    "spawnersImported": 12,
    "characterTemplatesImported": 4,
    "charactersReset": 5,
    "usersPreserved": 3
  }
}
```

**Error (400 Bad Request)**:
```json
{
  "success": false,
  "error": "Import folder not found: mvp-rpg-2"
}
```

**Error (500 Internal Server Error)**:
```json
{
  "success": false,
  "error": "Failed to parse rooms/R0001.yaml: yaml: line 5: mapping values are not allowed here",
  "phase": "import_rooms"
}
```

---

## Implementation Details

### Phase 1: Drop Game Content

```go
// Order matters - drop dependent entities first
handler.ConversationsService.Drop()  // Session data
handler.PartiesService.Drop()        // Session data
handler.NPCSpawnersService.Drop()    // References NPCs
handler.NPCsService.Drop()           // References dialogs, loot tables
handler.DialogsService.Drop()
handler.LootTablesService.Drop()
handler.ItemsService.Drop()
handler.ScriptsService.Drop()
handler.RoomsService.Drop()
handler.CharacterTemplatesService.Drop()
```

### Phase 2: Reset Characters

```go
// Get all characters
characters, _ := handler.CharactersService.FindAll()

for _, char := range characters {
    // Reset location to starting room
    char.CurrentRoomID = request.StartRoomId

    // Clear combat state
    char.CombatInstanceID = ""

    // Optionally clear inventory
    if request.ClearInventories {
        char.Inventory = []items.Item{}
        char.EquippedItems = make(map[string]string)
    }

    // Reset bound room if it no longer exists
    char.BoundRoomID = request.StartRoomId

    handler.CharactersService.Update(char.ID, char)
}
```

### Phase 3: Read YAML Files

Create a new package `pkg/importer/yaml_reader.go`:

```go
package importer

import (
    "os"
    "path/filepath"
    "gopkg.in/yaml.v3"
)

type WorldData struct {
    Rooms              []*rooms.Room
    Items              []*items.Item
    NPCs               []*npc.NPC
    Dialogs            []*dialogs.Dialog
    Scripts            []*scripts.Script
    LootTables         []*items.LootTable
    NPCSpawners        []*npc.Spawner
    CharacterTemplates []*characters.CharacterTemplate
}

func ReadWorldFromFolder(basePath string) (*WorldData, error) {
    data := &WorldData{}

    // Read rooms
    roomFiles, _ := filepath.Glob(filepath.Join(basePath, "rooms", "*.yaml"))
    for _, file := range roomFiles {
        room, err := readYAMLFile[rooms.Room](file)
        if err != nil {
            return nil, fmt.Errorf("failed to read %s: %w", file, err)
        }
        data.Rooms = append(data.Rooms, room)
    }

    // Repeat for items, npcs, dialogs, scripts, loot_tables, spawners, templates
    // ...

    return data, nil
}

func readYAMLFile[T any](path string) (*T, error) {
    content, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var entity T
    if err := yaml.Unmarshal(content, &entity); err != nil {
        return nil, err
    }

    return &entity, nil
}
```

### Phase 4: Import World Data

```go
// Import in dependency order (dependencies first)
for _, script := range data.Scripts {
    handler.ScriptsService.Import(script)
}

for _, room := range data.Rooms {
    handler.RoomsService.Import(room)
}

for _, item := range data.Items {
    handler.ItemsService.Import(item)
}

for _, lootTable := range data.LootTables {
    handler.LootTablesService.Import(lootTable)
}

for _, dialog := range data.Dialogs {
    handler.DialogsService.Import(dialog)
}

for _, npc := range data.NPCs {
    handler.NPCsService.Import(npc)
}

for _, spawner := range data.NPCSpawners {
    handler.NPCSpawnersService.Import(spawner)
}

for _, template := range data.CharacterTemplates {
    handler.CharacterTemplatesService.Import(template)
}
```

---

## Data Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                  Creator UI / Admin Client                       │
│                                                                  │
│  [Reset World Button] → Confirmation Dialog → POST Request      │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ▼
┌──────────────────────────────────────────────────────────────────┐
│              POST /admin/reset-world                             │
│              Authorization: Basic admin:password                 │
│              Body: { folder, startRoomId, clearInventories }    │
└──────────────────────────────┬──────────────────────────────────┘
                               │
                               ▼
┌──────────────────────────────────────────────────────────────────┐
│                    ResetWorldHandler                             │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ 1. Validate request parameters                             │ │
│  │ 2. Verify import folder exists                             │ │
│  │ 3. Parse all YAML files (fail fast on errors)              │ │
│  └────────────────────────────────────────────────────────────┘ │
│                               │                                  │
│                               ▼                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ 4. DROP game content tables                                │ │
│  │    - conversations, parties (session)                      │ │
│  │    - npc_spawners, npcs, dialogs, loot_tables             │ │
│  │    - items, scripts, rooms, character_templates           │ │
│  └────────────────────────────────────────────────────────────┘ │
│                               │                                  │
│                               ▼                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ 5. RESET characters                                        │ │
│  │    - Set CurrentRoomID = startRoomId                       │ │
│  │    - Clear CombatInstanceID                                │ │
│  │    - Optionally clear inventory                            │ │
│  └────────────────────────────────────────────────────────────┘ │
│                               │                                  │
│                               ▼                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ 6. IMPORT new world data                                   │ │
│  │    - scripts → rooms → items → loot_tables                │ │
│  │    - dialogs → npcs → spawners → templates                │ │
│  └────────────────────────────────────────────────────────────┘ │
│                               │                                  │
│                               ▼                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ 7. Return success response with stats                      │ │
│  └────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────┘
```

---

## File Changes Required

### New Files

| File | Description |
|------|-------------|
| `pkg/importer/yaml_reader.go` | YAML file reading and parsing |
| `pkg/importer/world_data.go` | WorldData struct definition |
| `pkg/server/handler/reset_world.go` | Reset world endpoint handler |

### Modified Files

| File | Changes |
|------|---------|
| `pkg/server/server.go` | Add route for `/admin/reset-world` |
| `pkg/service/facade.go` | Add any missing service accessors |
| `go.mod` | Add `gopkg.in/yaml.v3` dependency |

### Dependencies

```bash
go get gopkg.in/yaml.v3
```

---

## YAML File Format Examples

### Room (rooms/R0001.yaml)

```yaml
id: "R0001"
name: "The Awakening Chamber"
area: "Sunken Sanctum"
description: "You slowly regain consciousness in a dimly lit stone chamber..."
exits:
  - direction: "north"
    targetRoomId: "R0002"
    description: "A narrow passage leads north"
actions:
  - id: "examine_altar"
    name: "Examine Altar"
    scriptId: "SCR0002"
onEnterScriptId: "SCR0001"
canBind: false
metadata:
  theme: "underground"
  lighting: "dim"
```

### NPC (npcs/NPC0001.yaml)

```yaml
id: "NPC0001"
name: "Mira Thornwood"
description: "A weathered healer with kind eyes..."
roomId: "R0001"
dialogTrait:
  dialogId: "DLG0001"
  greeting: "Ah, you're finally awake..."
merchantTrait: null
enemyTrait: null
```

### Item (items/ITM0001.yaml)

```yaml
id: "ITM0001"
name: "Rusty Dagger"
description: "A worn but serviceable blade"
isTemplate: true
itemType: "weapon"
slot: "mainHand"
stats:
  damage: 3
  speed: 1.2
value: 5
```

---

## Error Handling

### Validation Errors (Pre-execution)

- Import folder does not exist
- startRoomId not found in import data
- YAML parsing errors (report file and line number)
- Missing required fields in YAML files

### Execution Errors

- Database connection failures
- Individual entity import failures (log and continue vs fail fast - TBD)

### Recommended Behavior

1. **Parse all YAML files first** before any database changes
2. **Validate startRoomId exists** in the room data being imported
3. **Fail fast** on any error - do not partially import
4. **Log detailed errors** including file paths and line numbers

---

## Security Considerations

- Endpoint protected by Basic Auth (existing admin credentials)
- Only reads from predefined `import/` folder (no arbitrary path traversal)
- Folder parameter validated against directory listing
- No user-uploaded content - files must exist on server

---

## Testing Checklist

- [ ] Reset with valid folder and startRoomId
- [ ] Reset with non-existent folder (should fail)
- [ ] Reset with non-existent startRoomId (should fail)
- [ ] Reset with malformed YAML file (should fail with clear error)
- [ ] Verify users are preserved after reset
- [ ] Verify characters are moved to startRoomId
- [ ] Verify character inventories cleared when flag is true
- [ ] Verify character inventories preserved when flag is false
- [ ] Verify all game content is replaced
- [ ] Verify NPCs spawn correctly after import
- [ ] Verify dialogs work after import
- [ ] Verify scripts execute after import

---

## Future Enhancements (Out of Scope)

- **Dry Run Mode**: Validate import without making changes
- **Rollback**: Automatic backup before reset
- **Partial Import**: Import only specific entity types
- **Version Tracking**: Track which world version is currently loaded
- **Hot Reload**: Import without kicking players (graceful migration)
- **Diff Preview**: Show what would change before confirming

---

## Implementation Order

1. Add YAML dependency to `go.mod`
2. Create `pkg/importer/yaml_reader.go` with generic YAML parsing
3. Create `pkg/importer/world_data.go` with WorldData struct
4. Create `pkg/server/handler/reset_world.go` with endpoint handler
5. Add route in `pkg/server/server.go`
6. Test with existing `import/mvp-rpg-1/` data
7. Add Creator UI button (optional, can use curl for MVP)
