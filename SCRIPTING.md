# TalesMUD Scripting System

This document describes the scripting system in TalesMUD, including the Lua scripting engine, available API modules, and event system.

## Overview

TalesMUD uses **Lua** (via [gopher-lua](https://github.com/yuin/gopher-lua)) as the primary scripting language for dynamic game content. JavaScript support is deprecated but maintained for backward compatibility.

## Scripting Languages

### Lua (Recommended)
- Modern, fast, and widely used in game development
- Full access to the `tales.*` module API
- Event-driven scripting support
- 5-second execution timeout for safety

### JavaScript (Deprecated)
- Uses Otto engine (ES5 only)
- Legacy support for existing scripts
- Will show deprecation warnings in logs

## Script Entity

Scripts are stored in the database with the following structure:

```go
type Script struct {
    ID          string         // Unique identifier
    Name        string         // Human-readable name
    Description string         // What the script does
    Code        string         // Lua or JavaScript code
    Type        ScriptType     // Categorization (item, room, npc, quest, event)
    Language    ScriptLanguage // "lua" or "javascript"
}
```

### Script Types

| Type | Use Case |
|------|----------|
| `item` | Item creation and behavior |
| `room` | Room actions and triggers |
| `npc` | NPC behavior and AI |
| `quest` | Quest logic |
| `event` | Event handlers |
| `custom` | General purpose |

## Lua API Reference

All Lua API functions are available under the `tales` global module.

### tales.items

```lua
-- Get item by ID
local item = tales.items.get(itemID)

-- Find items by name
local items = tales.items.findByName("sword")

-- Get item template by ID
local template = tales.items.getTemplate(templateID)

-- Find item templates by name
local templates = tales.items.findTemplates("sword")

-- Create item from template
local newItem = tales.items.createFromTemplate(templateID)

-- Delete an item
local success = tales.items.delete(itemID)
```

### tales.rooms

```lua
-- Get room by ID
local room = tales.rooms.get(roomID)

-- Find rooms by name
local rooms = tales.rooms.findByName("tavern")

-- Find rooms by area
local rooms = tales.rooms.findByArea("dungeon")

-- Get all rooms
local allRooms = tales.rooms.getAll()

-- Get characters in a room
local characters = tales.rooms.getCharacters(roomID)

-- Get NPCs in a room
local npcs = tales.rooms.getNPCs(roomID)

-- Get items in a room
local items = tales.rooms.getItems(roomID)
```

### tales.characters

```lua
-- Get character by ID
local char = tales.characters.get(characterID)

-- Find characters by name
local chars = tales.characters.findByName("hero")

-- Get all characters
local allChars = tales.characters.getAll()

-- Get the room a character is in
local room = tales.characters.getRoom(characterID)

-- Damage a character
local success = tales.characters.damage(characterID, amount)

-- Heal a character
local success = tales.characters.heal(characterID, amount)

-- Teleport character to room
local success = tales.characters.teleport(characterID, roomID)

-- Give XP to character
local success = tales.characters.giveXP(characterID, amount)
```

### tales.npcs

```lua
-- Get NPC by ID
local npc = tales.npcs.get(npcID)

-- Find NPCs by name
local npcs = tales.npcs.findByName("guard")

-- Find NPCs in a room
local npcs = tales.npcs.findInRoom(roomID)

-- Get all NPCs
local allNPCs = tales.npcs.getAll()

-- Damage an NPC
local success = tales.npcs.damage(npcID, amount)

-- Heal an NPC
local success = tales.npcs.heal(npcID, amount)

-- Move NPC to room
local success = tales.npcs.moveTo(npcID, roomID)

-- Check if NPC is dead
local isDead = tales.npcs.isDead(npcID)

-- Check if NPC is an enemy
local isEnemy = tales.npcs.isEnemy(npcID)

-- Check if NPC is a merchant
local isMerchant = tales.npcs.isMerchant(npcID)

-- Delete an NPC
local success = tales.npcs.delete(npcID)
```

### tales.dialogs

```lua
-- Get dialog by ID
local dialog = tales.dialogs.get(dialogID)

-- Find dialogs by name
local dialogs = tales.dialogs.findByName("greeting")

-- Get all dialogs
local allDialogs = tales.dialogs.getAll()

-- Get conversation between character and NPC
local conv = tales.dialogs.getConversation(characterID, npcID)

-- Set conversation context variable
local success = tales.dialogs.setContext(conversationID, "key", "value")

-- Get conversation context variable
local value = tales.dialogs.getContext(conversationID, "key")

-- Check if dialog node was visited
local visited = tales.dialogs.hasVisited(conversationID, nodeID)

-- Get visit count for dialog node
local count = tales.dialogs.getVisitCount(conversationID, nodeID)
```

### tales.game

```lua
-- Send message to all players in room
tales.game.msgToRoom(roomID, "A rumbling sound echoes...")

-- Send message to specific character
tales.game.msgToCharacter(characterID, "You feel a chill...")

-- Send message to specific user
tales.game.msgToUser(userID, "System message")

-- Broadcast to all connected players
tales.game.broadcast("Server announcement!")

-- Send to room except one player
tales.game.msgToRoomExcept(roomID, "Others see this", excludeCharacterID)

-- Log a message
tales.game.log("info", "Something happened")
```

### tales.utils

```lua
-- Generate random number (inclusive)
local num = tales.utils.random(1, 100)

-- Generate random float 0-1
local f = tales.utils.randomFloat()

-- Generate UUID
local id = tales.utils.uuid()

-- Get current Unix timestamp
local now = tales.utils.now()

-- Get current timestamp in milliseconds
local nowMs = tales.utils.nowMs()

-- Format timestamp to string
local str = tales.utils.formatTime(timestamp)

-- Roll dice (e.g., "2d6", "1d20+5")
local result = tales.utils.roll("2d6+3")

-- Percentage chance (returns true/false)
local success = tales.utils.chance(25) -- 25% chance

-- Pick random element from array
local picked = tales.utils.pick({"sword", "axe", "spear"})

-- Shuffle array
local shuffled = tales.utils.shuffle(myArray)

-- Clamp value between min and max
local clamped = tales.utils.clamp(value, 0, 100)

-- Linear interpolation
local lerped = tales.utils.lerp(0, 100, 0.5) -- Returns 50
```

## Context Variable

Scripts receive a `ctx` global variable containing context data:

```lua
-- For item creation scripts
ctx.item     -- The created item
ctx.template -- The source template

-- For event scripts (varies by event type)
ctx.eventType  -- Event type string
ctx.timestamp  -- Unix timestamp
ctx.room       -- Room where event occurred
ctx.character  -- Character involved
ctx.npc        -- NPC involved
ctx.item       -- Item involved
```

## Event System

The event system allows scripts to respond to game events.

### Event Types

| Event | Description |
|-------|-------------|
| `player.enter_room` | Player enters a room |
| `player.leave_room` | Player leaves a room |
| `player.join` | Player connects to game |
| `player.quit` | Player disconnects |
| `player.death` | Player dies |
| `player.level_up` | Player levels up |
| `item.pickup` | Player picks up item |
| `item.drop` | Player drops item |
| `item.use` | Player uses item |
| `item.create` | Item is created |
| `npc.death` | NPC dies |
| `npc.spawn` | NPC spawns |
| `npc.idle` | NPC idle tick |
| `dialog.start` | Dialog begins |
| `dialog.end` | Dialog ends |
| `dialog.option` | Dialog option selected |
| `room.action` | Room action triggered |
| `room.update` | Room update tick |
| `quest.start` | Quest started |
| `quest.complete` | Quest completed |
| `quest.fail` | Quest failed |
| `quest.progress` | Quest progress updated |

### Event Context

Event scripts receive context data specific to the event type:

```lua
-- Player movement events
ctx.character  -- The moving character
ctx.room       -- Current room (after move for enter, before for leave)
ctx.fromRoom   -- Previous room (for enter_room)
ctx.toRoom     -- Destination room (for leave_room)

-- NPC events
ctx.npc        -- The NPC
ctx.killer     -- Who killed the NPC (for npc.death)

-- Dialog events
ctx.character     -- Player in dialog
ctx.npc           -- NPC in dialog
ctx.dialogId      -- Dialog ID
ctx.nodeId        -- Current node ID
ctx.conversationId -- Conversation ID
ctx.optionSelected -- Selected option index

-- Quest events
ctx.character  -- Player with quest
ctx.questId    -- Quest ID
ctx.objectiveId -- Objective ID (for progress)
ctx.progress   -- Progress amount
```

## Example Scripts

### Random Item Stats

```lua
-- Randomize weapon stats when created from template
local item = ctx.item
item.minDamage = tales.utils.random(1, 5)
item.maxDamage = item.minDamage + tales.utils.random(5, 15)

-- Set quality based on damage
if item.maxDamage > 15 then
    item.quality = "rare"
else
    item.quality = "common"
end

return item
```

### Dungeon Entrance Atmosphere

```lua
-- Room enter event handler for dungeon areas
if ctx.room.area == "dungeon" then
    tales.game.msgToCharacter(ctx.character.ID,
        "Shadows dance on the walls as your torch flickers...")

    -- 20% chance to spawn a monster
    if tales.utils.chance(20) then
        -- Spawn logic would go here once NPC spawning is implemented
        tales.game.msgToRoom(ctx.room.ID,
            "A skeleton emerges from the darkness!")
    end
end
```

### NPC Death Handler

```lua
-- Handle NPC death for quest progress
local npc = ctx.npc
local killer = ctx.killer

-- Update kill quest progress
if npc.Name == "Goblin" then
    -- Quest progress would be updated here
    tales.game.msgToCharacter(killer.ID,
        "Quest progress: Goblin slain!")
end

-- Boss kill announcement
if npc.Name == "Goblin King" then
    tales.game.broadcast(killer.Name .. " has slain the Goblin King!")
end
```

### Dynamic Merchant Greeting

```lua
-- Dialog start event for merchants
if ctx.npc.MerchantTrait then
    -- Set greeting based on time of day
    local hour = os.date("*t").hour
    local greeting

    if hour < 12 then
        greeting = "Good morning, traveler!"
    elseif hour < 18 then
        greeting = "Good afternoon! Looking to buy?"
    else
        greeting = "Evening! Last chance for today's deals!"
    end

    tales.dialogs.setContext(ctx.conversationId, "greeting", greeting)
end
```

## Safety and Sandboxing

Lua scripts run in a sandboxed environment with:

- **5-second timeout**: Scripts that run too long are terminated
- **Disabled modules**: `os`, `io`, `debug`, `loadfile`, `dofile` are removed
- **VM pooling**: Lua states are reused for performance
- **Isolated execution**: Each script runs in its own context

## Files

| File | Purpose |
|------|---------|
| `pkg/scripts/scripts.go` | Script entity definition |
| `pkg/scripts/scriptrunner.go` | Runner interface |
| `pkg/scripts/runner/factory.go` | Multi-runner factory |
| `pkg/scripts/runner/lua/luarunner.go` | Lua runner implementation |
| `pkg/scripts/runner/lua/sandbox.go` | Sandbox configuration |
| `pkg/scripts/runner/lua/pool.go` | VM pool for performance |
| `pkg/scripts/runner/lua/modules/*.go` | Lua API modules |
| `pkg/scripts/events/*.go` | Event system |
| `pkg/scripts/runner/defaultscriptrunner.go` | JavaScript runner (deprecated) |

## REST API

Scripts can be managed via REST API:

- `GET /api/scripts` - List all scripts
- `GET /api/scripts/:id` - Get script by ID
- `POST /api/scripts` - Create new script
- `PUT /api/scripts/:id` - Update script
- `DELETE /api/scripts/:id` - Delete script
- `POST /api/run-script/:id` - Execute script with context
- `GET /api/script-types` - Get available script types

## Migration from JavaScript

Existing JavaScript scripts continue to work but will show deprecation warnings. To migrate:

1. Set the script's `Language` field to `"lua"`
2. Convert JavaScript syntax to Lua:
   - `var` → `local`
   - `function() {}` → `function() end`
   - `T_functionName()` → `tales.module.functionName()`
   - Array indexing starts at 1 in Lua
   - Use `..` for string concatenation instead of `+`

### JavaScript to Lua Example

**JavaScript (old):**
```javascript
var item = ctx;
item.minDamage = Math.floor(Math.random() * 5) + 1;
item.maxDamage = item.minDamage + Math.floor(Math.random() * 10) + 5;
ctx = JSON.stringify(item);
```

**Lua (new):**
```lua
local item = ctx.item
item.minDamage = tales.utils.random(1, 5)
item.maxDamage = item.minDamage + tales.utils.random(5, 15)
return item
```
