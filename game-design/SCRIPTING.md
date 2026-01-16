# TalesMUD Scripting System

This document describes the scripting system in TalesMUD, its current capabilities, and recommendations for future development.

## Overview

TalesMUD uses an embedded JavaScript engine ([Otto](https://github.com/robertkrimen/otto)) to provide scripting capabilities for dynamic game content. Scripts are stored as entities in the database and can be attached to various game objects.

## Current Implementation

### Architecture

```
Script Entity (stored in DB)
       ↓
   ScriptsService (CRUD operations)
       ↓
   ScriptRunner Interface
       ↓
   DefaultScriptRunner (Otto JavaScript VM)
       ↓
   T_* Built-in Functions (API)
       ↓
   Game Services (Items, Rooms, Characters)
```

### Script Entity Structure

Scripts are defined in `pkg/scripts/scripts.go`:

```go
type Script struct {
    *entities.Entity          // ID, timestamps, ownership
    Name        string        // Human-readable name
    Description string        // What the script does
    Code        string        // JavaScript code
    Type        ScriptType    // Categorization
}
```

### Script Types

| Type | Constant | Intended Use |
|------|----------|--------------|
| None | `none` | Uncategorized |
| Custom | `custom` | General purpose |
| Item | `item` | Item creation/behavior |
| Room | `room` | Room actions and triggers |
| Quest | `quest` | Quest logic |
| NPC | `npc` | NPC behavior |

### Files

| File | Purpose |
|------|---------|
| `pkg/scripts/scripts.go` | Script entity definition |
| `pkg/scripts/scriptrunner.go` | ScriptRunner interface |
| `pkg/scripts/runner/defaultscriptrunner.go` | JavaScript implementation |
| `pkg/service/scripts.go` | Service layer |
| `pkg/repository/scripts.go` | MongoDB repository |
| `pkg/repository/scripts_sqlite.go` | SQLite repository |
| `pkg/server/handler/scripts.go` | REST API endpoints |

## Current Capabilities

### What Works Today

1. **Item Template Scripts** - Scripts execute when items are created from templates
   - Location: `pkg/service/items.go:124-129`
   - The script receives the newly created item as `ctx`

2. **Script CRUD via REST API**
   - `GET /api/scripts` - List all scripts
   - `GET /api/scripts/:id` - Get script by ID
   - `POST /api/scripts` - Create script
   - `PUT /api/scripts/:id` - Update script
   - `DELETE /api/scripts/:id` - Delete script
   - `POST /api/run-script/:id` - Execute script with context

3. **Export/Import** - Scripts are included in world data exports

### Available Script API Functions

Scripts have access to these built-in functions (prefixed with `T_`):

#### Item Functions
```javascript
// Find item templates by name (returns JSON array)
T_findItemTemplate(name)

// Get a specific item template by ID (returns JSON)
T_getItemTemplate(templateID)

// Create a new item instance from a template (returns JSON)
T_createItemFromTemplate(templateID)
```

#### Room Functions
```javascript
// Find rooms by name (returns JSON array)
T_findRoom(roomName)

// Get a specific room by ID (returns JSON)
T_getRoom(roomID)

// Update a room (pass JSON string)
T_updateRoom(roomJsonString)
```

#### Game Functions
```javascript
// Send a system message to all players in a room
T_msgToRoom(roomID, message)
```

#### Context Variable
```javascript
// The context passed to the script (e.g., an item being created)
ctx
```

### Example Scripts

#### Random Item Stats on Creation
```javascript
// Randomize damage for a weapon when created
var item = ctx;
item.minDamage = Math.floor(Math.random() * 5) + 1;
item.maxDamage = item.minDamage + Math.floor(Math.random() * 10) + 5;
ctx = JSON.stringify(item);
```

#### Room Announcement
```javascript
// Announce something to a room
T_msgToRoom("room-id-here", "The ground trembles beneath your feet...");
```

### What's Partially Implemented

1. **Room Action Scripts** - The `RoomActionTypeScript` is defined but not executed
   - Location: `pkg/mudserver/game/commands/roomprocessor.go:128-131`
   - Currently falls through to error logging

2. **NPC Scripts** - Script type exists but no integration with NPC entities

3. **Quest Scripts** - Script type defined but no quest system infrastructure

## Known Limitations

1. **No Timeout Protection** - Scripts can theoretically run indefinitely (TODO in code)
2. **Limited API** - Only basic item, room, and messaging functions
3. **No Event System** - Scripts can't respond to game events
4. **No Character Functions** - Can't interact with players/NPCs from scripts
5. **Synchronous Execution** - All scripts run synchronously

## Recommendations for Improvement

### High Priority

#### 1. Complete Room Action Script Execution

Add script execution in `roomprocessor.go`:

```go
case rooms.RoomActionTypeScript:
    if scriptID, ok := action.Params["scriptId"].(string); ok {
        if script, err := scriptsService.FindByID(scriptID); err == nil {
            ctx := map[string]interface{}{
                "room": room,
                "character": character,
                "action": action,
            }
            scriptRunner.Run(*script, ctx)
        }
    }
```

#### 2. Add Script Timeout Protection

```go
func (runner *DefaultScriptRunner) Run(script scripts.Script, ctx interface{}) interface{} {
    vm := runner.newScriptRuntime()
    vm.Set("ctx", ctx)

    // Interrupt after 5 seconds
    vm.Interrupt = make(chan func(), 1)
    go func() {
        time.Sleep(5 * time.Second)
        vm.Interrupt <- func() {
            panic("script execution timeout")
        }
    }()

    defer func() {
        if r := recover(); r != nil {
            logrus.Error("Script timed out or panicked")
        }
    }()

    _, err := vm.Run(script.Code)
    // ...
}
```

#### 3. Expand Script API

Add character/NPC functions:
```javascript
T_getCharacter(characterID)
T_getNPC(npcID)
T_moveCharacterToRoom(characterID, roomID)
T_giveItemToCharacter(characterID, itemID)
T_msgToCharacter(characterID, message)
T_damageCharacter(characterID, amount)
T_healCharacter(characterID, amount)
```

Add inventory functions:
```javascript
T_addItemToInventory(characterID, itemID)
T_removeItemFromInventory(characterID, itemID)
T_hasItem(characterID, itemName)
```

#### 4. Add NPC Script Integration

Add script reference to NPC entity:
```go
type NPC struct {
    // ... existing fields
    BehaviorScriptID string `bson:"behaviorScriptID,omitempty" json:"behaviorScriptID,omitempty"`
    OnTalkScriptID   string `bson:"onTalkScriptID,omitempty" json:"onTalkScriptID,omitempty"`
    OnDeathScriptID  string `bson:"onDeathScriptID,omitempty" json:"onDeathScriptID,omitempty"`
}
```

### Medium Priority

#### 5. Event-Driven Script System

Create an event system where scripts can register for events:

| Event | Trigger | Context |
|-------|---------|---------|
| `onPlayerEnterRoom` | Player enters a room | player, room, fromRoom |
| `onPlayerLeaveRoom` | Player leaves a room | player, room, toRoom |
| `onItemPickup` | Player picks up item | player, item, room |
| `onItemDrop` | Player drops item | player, item, room |
| `onNPCDeath` | NPC is killed | npc, killer, room |
| `onDialogOption` | Dialog option selected | player, npc, dialog, option |
| `onTimerTick` | Periodic timer | world state |

#### 6. Quest System Foundation

Create quest entities with script hooks:
```go
type Quest struct {
    *entities.Entity
    Name              string
    Description       string
    StartScriptID     string   // Runs when quest starts
    CompleteScriptID  string   // Runs when quest completes
    FailScriptID      string   // Runs when quest fails
    Objectives        []QuestObjective
}

type QuestObjective struct {
    ID               string
    Description      string
    Type             string   // "kill", "collect", "visit", "talk"
    Target           string   // NPC ID, Item ID, Room ID, etc.
    Required         int
    OnProgressScript string   // Runs when progress is made
}
```

#### 7. Dialog Script Integration

Add script hooks to dialogs:
```go
type Dialog struct {
    // ... existing fields
    OnEnterScriptID string `bson:"onEnterScriptID,omitempty"`
    OnExitScriptID  string `bson:"onExitScriptID,omitempty"`
    Condition       string `bson:"condition,omitempty"` // JS expression
}
```

### Low Priority

#### 8. Script Editor Improvements
- Syntax highlighting in web UI
- Script testing/debugging tools
- Script versioning

#### 9. Script Variables/State
- Persistent script variables (saved between executions)
- Global game state accessible to scripts

## Language Comparison: Should We Switch?

### Current: JavaScript (Otto)

**Pros:**
- Already implemented and working
- Familiar syntax for most developers
- Includes underscore.js library
- Pure Go implementation (no CGO)
- Good for simple scripts

**Cons:**
- Otto is ES5 only (no modern JavaScript features)
- Not as performant as native solutions
- Project has limited maintenance
- No async/await support

### Alternative: Lua (via gopher-lua or golua)

**Pros:**
- Industry standard for game scripting (Unity, Roblox, WoW, etc.)
- Extremely fast execution
- Very small memory footprint
- Simple, clean syntax
- Excellent sandboxing
- Well-documented for game development
- [gopher-lua](https://github.com/yuin/gopher-lua) is pure Go, actively maintained

**Cons:**
- Different syntax (may require learning)
- Fewer web developers familiar with it
- Would require rewriting existing scripts

**Example Lua script:**
```lua
-- Random item stats
local item = ctx
item.minDamage = math.random(1, 5)
item.maxDamage = item.minDamage + math.random(5, 15)
return item
```

### Alternative: Python (embedded)

**Pros:**
- Very familiar to most developers
- Rich standard library
- Great for complex logic

**Cons:**
- Heavy runtime overhead
- Complex embedding (requires CGO with CPython)
- Security sandboxing is difficult
- Not designed for embedding

### Alternative: Starlark (Google's Python dialect)

**Pros:**
- Python-like syntax
- Pure Go implementation
- Designed for configuration/scripting
- Good sandboxing
- Used by Bazel, Buck, and other tools

**Cons:**
- Limited standard library
- Less feature-rich than full Python
- Smaller community

### Alternative: JavaScript (goja)

**Pros:**
- ES6+ support (modern JavaScript)
- Better performance than Otto
- Pure Go implementation
- Drop-in replacement potential

**Cons:**
- Still not as fast as Lua
- More memory usage than Lua

## Recommendation

### For Immediate Use: Stay with JavaScript (Otto)

The current implementation works and switching languages now would delay more important features. Focus on:
1. Completing room action script execution
2. Adding timeout protection
3. Expanding the API

### For Future Consideration: Evaluate goja or Lua

**If staying with JavaScript:** Consider migrating from Otto to [goja](https://github.com/dop251/goja) for ES6+ support and better performance.

**If switching languages:** Lua with [gopher-lua](https://github.com/yuin/gopher-lua) is the best choice for game scripting:
- Battle-tested in game industry
- Best performance
- Smallest memory footprint
- Easy to learn for content creators
- Excellent documentation for game use cases

The switch to Lua would be a larger undertaking but would provide a more robust, industry-standard scripting foundation.

## Quick Start for Content Creators

### Creating a Script

1. Go to Admin UI > Scripts
2. Click "New Script"
3. Fill in:
   - **Name**: Descriptive name (e.g., "Random Sword Stats")
   - **Type**: Select appropriate type (item, room, npc, etc.)
   - **Description**: What the script does
   - **Code**: JavaScript code

### Attaching to Item Templates

1. Create or edit an Item Template
2. Select your script from the "Script" dropdown
3. Save the template
4. When items are created from this template, the script runs

### Testing Scripts

Use the REST API endpoint:
```bash
curl -X POST http://localhost:8080/api/run-script/SCRIPT_ID \
  -H "Content-Type: application/json" \
  -d '{"testData": "value"}'
```

## Future Vision

The goal is to enable content creators to build rich, dynamic game experiences:

- **Dynamic Quests**: NPCs that give quests based on player actions
- **Living World**: Weather systems, day/night cycles, random events
- **Smart NPCs**: Enemies with patrol routes, merchants with dynamic prices
- **Interactive Environments**: Traps, puzzles, secret doors
- **Player-triggered Events**: Completing quests unlocks new areas
- **Scheduled Events**: World bosses, invasions, festivals

All of this should be achievable through scripts without modifying the core game code.
