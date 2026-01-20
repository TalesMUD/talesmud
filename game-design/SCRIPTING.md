# TalesMUD Scripting System

This document describes the Lua scripting system in TalesMUD, its API, and examples for content creators.

## Overview

TalesMUD uses an embedded Lua scripting engine ([gopher-lua](https://github.com/yuin/gopher-lua)) to provide dynamic game content. Scripts are stored as entities in the database and can be attached to rooms, items, NPCs, and triggered by events.

## Architecture

```
Script Entity (stored in DB)
       |
   ScriptsService (CRUD operations)
       |
   LuaRunner (gopher-lua VM pool)
       |
   tales.* API Modules
       |
   Game Services (Items, Rooms, Characters, NPCs, Dialogs)
```

### Key Features

- **VM Pool**: 10 reusable Lua states for performance
- **Sandbox**: Dangerous modules disabled (os, io, debug, loadfile, dofile)
- **Timeout**: 5-second execution limit per script
- **Go-Lua Bridge**: Automatic conversion between Go structs and Lua tables via luar

## Script Entity Structure

Scripts are defined in `pkg/scripts/scripts.go`:

```go
type Script struct {
    *entities.Entity          // ID, timestamps, ownership
    Name        string        // Human-readable name
    Description string        // What the script does
    Code        string        // Lua code
    Type        ScriptType    // Categorization
    Language    ScriptLanguage // "lua" (default)
}
```

### Script Types

| Type | Constant | Intended Use |
|------|----------|--------------|
| None | `none` | Uncategorized |
| Custom | `custom` | General purpose |
| Item | `item` | Item creation/behavior |
| Room | `room` | Room actions and entry triggers |
| Quest | `quest` | Quest logic |
| NPC | `npc` | NPC behavior |
| Event | `event` | Event handlers |

## Files

| File | Purpose |
|------|---------|
| `pkg/scripts/scripts.go` | Script entity definition |
| `pkg/scripts/scriptrunner.go` | ScriptRunner interface |
| `pkg/scripts/runner/lua/luarunner.go` | Lua VM implementation |
| `pkg/scripts/runner/lua/modules/*.go` | API module implementations |
| `pkg/scripts/events/events.go` | Event type definitions |
| `pkg/service/scripts.go` | Service layer |
| `pkg/repository/scripts.go` | MongoDB repository |
| `pkg/repository/scripts_sqlite.go` | SQLite repository |
| `pkg/server/handler/scripts.go` | REST API endpoints |

---

## Lua API Reference

All API functions are accessible via the `tales` global table.

### tales.items - Item Operations

```lua
-- Get item by ID
local item = tales.items.get(itemID)

-- Find items by name (partial match)
local items = tales.items.findByName("sword")

-- Get item template by ID
local template = tales.items.getTemplate(templateID)

-- Find templates by name
local templates = tales.items.findTemplates("iron")

-- Create item instance from template
local newItem = tales.items.createFromTemplate(templateID)

-- Delete an item
local success = tales.items.delete(itemID)
```

### tales.rooms - Room Operations

```lua
-- Get room by ID
local room = tales.rooms.get(roomID)

-- Find rooms by name
local rooms = tales.rooms.findByName("tavern")

-- Find rooms by area
local rooms = tales.rooms.findByArea("Darkwood Forest")

-- Get all rooms
local allRooms = tales.rooms.getAll()

-- Get character IDs in a room
local charIDs = tales.rooms.getCharacters(roomID)

-- Get NPC IDs in a room
local npcIDs = tales.rooms.getNPCs(roomID)

-- Get item IDs in a room
local itemIDs = tales.rooms.getItems(roomID)
```

### tales.characters - Character Operations

```lua
-- Get character by ID
local char = tales.characters.get(characterID)

-- Find characters by name
local chars = tales.characters.findByName("Gandalf")

-- Get all characters
local allChars = tales.characters.getAll()

-- Get the room a character is in
local room = tales.characters.getRoom(characterID)

-- Damage a character (reduces HP, clamped to 0)
local success = tales.characters.damage(characterID, 10)

-- Heal a character (increases HP, clamped to max)
local success = tales.characters.heal(characterID, 10)

-- Teleport character to a room (handles room membership)
local success = tales.characters.teleport(characterID, roomID)

-- Give XP to a character
local success = tales.characters.giveXP(characterID, 100)
```

### tales.npcs - NPC Operations

```lua
-- Get NPC by ID
local npc = tales.npcs.get(npcID)

-- Find NPCs by name
local npcs = tales.npcs.findByName("Guard")

-- Find NPCs in a specific room
local npcs = tales.npcs.findInRoom(roomID)

-- Get all NPCs
local allNPCs = tales.npcs.getAll()

-- Damage an NPC
local success = tales.npcs.damage(npcID, 15)

-- Heal an NPC
local success = tales.npcs.heal(npcID, 10)

-- Move NPC to a different room
local success = tales.npcs.moveTo(npcID, roomID)

-- Check if NPC is dead (HP <= 0)
local isDead = tales.npcs.isDead(npcID)

-- Check if NPC has EnemyTrait
local isEnemy = tales.npcs.isEnemy(npcID)

-- Check if NPC has MerchantTrait
local isMerchant = tales.npcs.isMerchant(npcID)

-- Delete an NPC
local success = tales.npcs.delete(npcID)
```

### tales.dialogs - Dialog & Conversation Operations

```lua
-- Get dialog by ID
local dialog = tales.dialogs.get(dialogID)

-- Find dialogs by name
local dialogs = tales.dialogs.findByName("merchant_greeting")

-- Get all dialogs
local allDialogs = tales.dialogs.getAll()

-- Get conversation between character and target (NPC/item)
local conv = tales.dialogs.getConversation(characterID, targetID)

-- Set conversation context variable
local success = tales.dialogs.setContext(conversationID, "questStatus", "accepted")

-- Get conversation context variable
local value = tales.dialogs.getContext(conversationID, "questStatus")

-- Check if dialog node was visited
local visited = tales.dialogs.hasVisited(conversationID, "intro_node")

-- Get visit count for a dialog node
local count = tales.dialogs.getVisitCount(conversationID, "intro_node")
```

### tales.game - Messaging

```lua
-- Send message to all players in a room
local success = tales.game.msgToRoom(roomID, "The ground shakes...")

-- Send message to a specific character
local success = tales.game.msgToCharacter(characterID, "You feel a chill...")

-- Send message to a specific user (by user ID)
local success = tales.game.msgToUser(userID, "System notification")

-- Broadcast message to all connected players
local success = tales.game.broadcast("Server will restart in 5 minutes")

-- Send to room except one character
local success = tales.game.msgToRoomExcept(roomID, "Player vanishes!", excludeCharID)

-- Log a message (levels: "debug", "info", "warn", "error")
tales.game.log("info", "Script executed successfully")
```

### tales.utils - Utility Functions

```lua
-- Random integer between min and max (inclusive)
local num = tales.utils.random(1, 100)

-- Random float between 0 and 1
local flt = tales.utils.randomFloat()

-- Generate UUID
local id = tales.utils.uuid()

-- Current Unix timestamp (seconds)
local ts = tales.utils.now()

-- Current Unix timestamp (milliseconds)
local tsMs = tales.utils.nowMs()

-- Format timestamp to readable string
local str = tales.utils.formatTime(timestamp) -- "2024-01-15 14:30:00"

-- Roll dice using standard notation
local result = tales.utils.roll("2d6")      -- 2-12
local result = tales.utils.roll("1d20+5")   -- 6-25
local result = tales.utils.roll("3d6-2")    -- 1-16

-- Percentage chance (returns true/false)
local success = tales.utils.chance(25)  -- 25% chance of true

-- Pick random element from array
local item = tales.utils.pick({"sword", "axe", "bow"})

-- Shuffle array in place
local shuffled = tales.utils.shuffle(myArray)

-- Clamp value between min and max
local clamped = tales.utils.clamp(value, 0, 100)

-- Linear interpolation
local lerped = tales.utils.lerp(0, 100, 0.5)  -- returns 50
```

---

## Script Context

Scripts receive context through the global `ctx` table. The context varies by trigger type.

### Item Creation Context
When a script runs during item creation:
```lua
ctx.item      -- The newly created item
ctx.template  -- The source item template
```

### Room Entry Context (OnEnterScriptID)
When a player enters a room:
```lua
ctx.eventType  -- "player.enter_room"
ctx.timestamp  -- Unix timestamp
ctx.room       -- The room being entered
ctx.toRoom     -- Same as ctx.room
ctx.character  -- The entering character
ctx.user       -- The user object
```

### General Event Context
```lua
ctx.eventType  -- Event type string
ctx.timestamp  -- Unix timestamp
ctx.room       -- Related room (if applicable)
ctx.character  -- Related character (if applicable)
ctx.npc        -- Related NPC (if applicable)
ctx.item       -- Related item (if applicable)
```

---

## Event System

TalesMUD defines 28 event types for script triggers. Currently only `player.enter_room` is wired into the game loop.

### Player Events
| Event | Description |
|-------|-------------|
| `player.enter_room` | Player enters a room (IMPLEMENTED) |
| `player.leave_room` | Player leaves a room |
| `player.join` | Player joins the game |
| `player.quit` | Player disconnects |
| `player.death` | Player character dies |
| `player.level_up` | Player gains a level |

### Item Events
| Event | Description |
|-------|-------------|
| `item.pickup` | Player picks up item |
| `item.drop` | Player drops item |
| `item.use` | Player uses item |
| `item.equip` | Player equips item |
| `item.unequip` | Player unequips item |
| `item.create` | Item is created |

### NPC Events
| Event | Description |
|-------|-------------|
| `npc.death` | NPC is killed |
| `npc.spawn` | NPC spawns |
| `npc.update` | NPC state changes |
| `npc.idle` | NPC idle tick |

### Dialog Events
| Event | Description |
|-------|-------------|
| `dialog.start` | Conversation begins |
| `dialog.end` | Conversation ends |
| `dialog.option` | Dialog option selected |

### Room Events
| Event | Description |
|-------|-------------|
| `room.update` | Room state changes |
| `room.action` | Room action triggered |

### Quest Events
| Event | Description |
|-------|-------------|
| `quest.start` | Quest begins |
| `quest.complete` | Quest completed |
| `quest.fail` | Quest failed |
| `quest.progress` | Quest progress updated |

### Combat Events
| Event | Description |
|-------|-------------|
| `combat.start` | Combat begins |
| `combat.end` | Combat ends |
| `combat.damage` | Damage dealt |

### Timer Events
| Event | Description |
|-------|-------------|
| `timer.tick` | Periodic timer |

---

## Example Scripts

### Random Item Stats on Creation
```lua
-- Script attached to item template
-- Randomizes weapon damage when item is created

local item = ctx.item
if item then
    local baseDamage = tales.utils.roll("2d6")
    item.Properties = item.Properties or {}
    item.Properties["damageMin"] = baseDamage
    item.Properties["damageMax"] = baseDamage + tales.utils.random(5, 10)

    -- 10% chance for magic quality
    if tales.utils.chance(10) then
        item.Quality = "magic"
        item.Properties["damageMax"] = item.Properties["damageMax"] + 3
    end
end

return ctx
```

### Room Entry Announcement
```lua
-- Script attached to room via OnEnterScriptID
-- Announces when players enter a special area

local char = ctx.character
local room = ctx.room

if char and room then
    -- Message to the entering player
    tales.game.msgToCharacter(char.ID,
        "You feel an ancient presence watching you...")

    -- Message to others in the room
    tales.game.msgToRoomExcept(room.ID,
        char.Name .. " enters cautiously, disturbing the dust.",
        char.ID)

    -- Log for debugging
    tales.game.log("info", char.Name .. " entered " .. room.Name)
end
```

### Trapped Room
```lua
-- Room entry script that damages careless players

local char = ctx.character
if not char then return end

-- 30% chance to trigger trap
if tales.utils.chance(30) then
    local damage = tales.utils.roll("1d6")
    tales.characters.damage(char.ID, damage)

    tales.game.msgToCharacter(char.ID,
        "A dart shoots from the wall! You take " .. damage .. " damage.")

    tales.game.msgToRoomExcept(ctx.room.ID,
        char.Name .. " triggers a trap and cries out in pain!",
        char.ID)
else
    tales.game.msgToCharacter(char.ID,
        "You carefully navigate the trapped hallway.")
end
```

### Dynamic NPC Greeting
```lua
-- Could be used in dialog system or NPC idle behavior

local hour = os.date("*t").hour  -- Note: os is sandboxed, use server time
local greeting

if hour < 12 then
    greeting = "Good morning, traveler!"
elseif hour < 18 then
    greeting = "Good afternoon, traveler!"
else
    greeting = "Good evening, traveler!"
end

-- Using game messaging
tales.game.msgToRoom(ctx.room.ID, "The innkeeper says: \"" .. greeting .. "\"")
```

### Quest Item Check
```lua
-- Check if player has required item (concept for future quests)

local function hasItem(charID, itemName)
    local char = tales.characters.get(charID)
    if not char or not char.Inventory or not char.Inventory.Items then
        return false
    end

    for _, item in ipairs(char.Inventory.Items) do
        if item.Name and item.Name:lower():find(itemName:lower()) then
            return true
        end
    end
    return false
end

local char = ctx.character
if char and hasItem(char.ID, "ancient key") then
    tales.game.msgToCharacter(char.ID,
        "The door recognizes your key and swings open!")
    -- In future: unlock exit, teleport player, etc.
else
    tales.game.msgToCharacter(char.ID,
        "The door is sealed with ancient magic.")
end
```

---

## REST API Endpoints

Scripts can be managed via the REST API:

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/scripts` | List all scripts |
| GET | `/api/scripts/:id` | Get script by ID |
| POST | `/api/scripts` | Create script |
| PUT | `/api/scripts/:id` | Update script |
| DELETE | `/api/scripts/:id` | Delete script |
| POST | `/api/run-script/:id` | Execute script with context |

### Testing Scripts via API

```bash
curl -X POST http://localhost:8080/api/run-script/SCRIPT_ID \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"testData": "value", "characterID": "char-123"}'
```

---

## Current Integration Points

### Implemented
- **Room Entry Scripts**: Set `OnEnterScriptID` on a room entity. Fires when player enters the room or selects a character in that room.
- **Item Template Scripts**: Reference a script ID on an item template. Fires when items are created from that template.

### Planned (Not Yet Wired)
- Room Action Scripts (`RoomActionTypeScript`)
- NPC Behavior Scripts
- Dialog Script Hooks
- Combat Event Scripts
- Quest Scripts

---

## Best Practices

### 1. Keep Scripts Focused
Each script should do one thing well. Complex behaviors should be split across multiple scripts.

### 2. Use Logging for Debugging
```lua
tales.game.log("debug", "Script checkpoint: processing " .. ctx.character.Name)
```

### 3. Validate Context
Always check that expected context values exist:
```lua
if not ctx.character then
    tales.game.log("warn", "No character in context!")
    return
end
```

### 4. Handle Errors Gracefully
API functions return `nil` on failure. Check return values:
```lua
local room = tales.rooms.get(roomID)
if not room then
    tales.game.log("error", "Room not found: " .. roomID)
    return
end
```

### 5. Avoid Infinite Loops
The 5-second timeout will kill runaway scripts, but design defensively:
```lua
local MAX_ITERATIONS = 100
local count = 0
while condition and count < MAX_ITERATIONS do
    -- work
    count = count + 1
end
```

### 6. Use Dice Notation for Randomness
`tales.utils.roll()` is more readable than math.random for game mechanics:
```lua
-- Good
local damage = tales.utils.roll("2d6+3")

-- Less clear
local damage = math.random(1,6) + math.random(1,6) + 3
```

---

## Sandbox Restrictions

The following Lua features are disabled for security:

- `os` module (system commands, file access)
- `io` module (file operations)
- `debug` module (introspection)
- `loadfile` / `dofile` (loading external code)
- Direct file system access

Available standard libraries:
- `base` (print, pairs, ipairs, type, etc.)
- `string` (manipulation functions)
- `table` (table operations)
- `math` (mathematical functions)

---

## Future Enhancements

### High Priority
1. Wire remaining event types to script execution
2. Add room action script support
3. Implement NPC behavior script hooks

### Medium Priority
4. Script versioning and rollback
5. Script debugging tools in admin UI
6. Persistent script variables (saved between executions)

### Low Priority
7. Visual script editor
8. Script performance profiling
9. Script dependency system
