# Quest Authoring Guide

This guide explains how to create quests for TalesMUD, both via the REST API (for programmatic generation) and the Creator UI. It covers the complete data model, all objective types, reward configuration, NPC dialog integration, and prerequisite chains.

## Table of Contents

- [Quick Start: Minimal Quest via API](#quick-start-minimal-quest-via-api)
- [Quest Data Model](#quest-data-model)
- [Objective Types](#objective-types)
- [Quest Source Configuration](#quest-source-configuration)
- [Rewards](#rewards)
- [Prerequisites](#prerequisites)
- [NPC Dialog Integration](#npc-dialog-integration)
- [API Reference](#api-reference)
- [YAML Format for Import](#yaml-format-for-import)
- [Complete Examples](#complete-examples)
- [Entity ID Reference](#entity-id-reference)
- [Workflow: Creating a Quest End-to-End](#workflow-creating-a-quest-end-to-end)
- [Troubleshooting](#troubleshooting)

---

## Quick Start: Minimal Quest via API

Create a simple "kill 5 rats" quest with a single API call:

```bash
curl -X POST http://localhost:8010/api/quests \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Rat Problem",
    "description": "The cellars are overrun with rats. Clear them out.",
    "category": "side",
    "level": 1,
    "source": {
      "type": "npc",
      "npcId": "NPC0001"
    },
    "objectives": [
      {
        "id": "obj1",
        "type": "kill",
        "description": "Kill 5 Catacomb Rats",
        "targetId": "ENM0001",
        "targetName": "Catacomb Rat",
        "amount": 5,
        "order": 1
      }
    ],
    "rewards": {
      "xp": 50,
      "gold": 10
    },
    "acceptDialogText": "Those rats in the cellar are getting bold. Kill five of them and I will make it worth your while.",
    "completeDialogText": "Good work. Here is your payment."
  }'
```

The server returns the created quest with a generated `id` field (UUID).

---

## Quest Data Model

### Top-Level Quest Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Auto-generated | UUID, assigned by server on creation. Do not set when creating. |
| `name` | string | **Yes** | Display name shown in quest log and UI. |
| `description` | string | **Yes** | Journal text describing the quest. Shown when player inspects quest details. |
| `category` | string | No | Classification: `"main"`, `"side"`, or `"daily"`. Defaults to empty. |
| `level` | int32 | No | Recommended player level. Used for UI display, not enforced. |
| `repeatable` | bool | No | If `true`, quest can be accepted again after completion. Default `false`. |
| `source` | QuestSource | **Yes** | How the player acquires this quest. See [Quest Source](#quest-source-configuration). |
| `objectives` | Objective[] | **Yes** | Array of objectives. At least one required. See [Objective Types](#objective-types). |
| `rewards` | Reward | No | XP, gold, and items granted on completion. See [Rewards](#rewards). |
| `requiredQuestIds` | string[] | No | Quest IDs that must be completed before this quest is available. |
| `requiredLevel` | int32 | No | Minimum character level to accept. |
| `onCompleteScriptId` | string | No | Lua script ID to execute on quest completion (e.g., `"SCR0010"`). |
| `acceptDialogText` | string | No | NPC speech when offering the quest. Auto-injected into dialog. |
| `progressDialogText` | string | No | NPC speech while quest is in progress. |
| `completeDialogText` | string | No | NPC speech when player turns in the completed quest. |

### Full JSON Schema

```json
{
  "name": "string (required)",
  "description": "string (required)",
  "category": "main | side | daily",
  "level": 0,
  "repeatable": false,
  "source": {
    "type": "npc | item | auto | script",
    "npcId": "string (NPC template ID)",
    "itemId": "string (Item template ID)"
  },
  "objectives": [
    {
      "id": "string (unique within quest, e.g. 'obj1')",
      "type": "kill | collect | deliver | visit | talk | custom",
      "description": "string (shown to player)",
      "targetId": "string (entity ID, depends on type)",
      "targetName": "string (display name)",
      "amount": 0,
      "deliverToNpcId": "string (deliver type only)",
      "deliverToNpcName": "string (deliver type only)",
      "dialogNodeId": "string (talk type only)",
      "checkScriptId": "string (custom type, or override for any type)",
      "order": 0
    }
  ],
  "rewards": {
    "xp": 0,
    "gold": 0,
    "itemTemplateIds": ["string"]
  },
  "requiredQuestIds": ["string"],
  "requiredLevel": 0,
  "onCompleteScriptId": "string",
  "acceptDialogText": "string",
  "progressDialogText": "string",
  "completeDialogText": "string"
}
```

---

## Objective Types

Each quest has one or more objectives. The `type` field determines which other fields are used by the QuestTracker for automatic progress updates.

### Kill (`"kill"`)

Player must kill a certain number of NPCs matching a template.

| Field | Required | Description |
|-------|----------|-------------|
| `targetId` | **Yes** | NPC **template** ID (e.g., `"ENM0001"`). Matches spawned instances of this template. |
| `targetName` | Recommended | Display name (e.g., `"Catacomb Rat"`). |
| `amount` | **Yes** | Number to kill. |

**How tracking works:** When an NPC dies in combat, the QuestTracker checks the dead NPC's `TemplateID` (or `ID` if no template) against `targetId`. Each kill increments the counter by 1.

```json
{
  "id": "kill_rats",
  "type": "kill",
  "description": "Kill 10 Catacomb Rats",
  "targetId": "ENM0001",
  "targetName": "Catacomb Rat",
  "amount": 10,
  "order": 1
}
```

### Collect (`"collect"`)

Player must pick up items matching a template.

| Field | Required | Description |
|-------|----------|-------------|
| `targetId` | **Yes** | Item **template** ID (e.g., `"ITM0018"`). Matches instances created from this template. |
| `targetName` | Recommended | Display name (e.g., `"Rat Tail"`). |
| `amount` | **Yes** | Number to collect. |

**How tracking works:** When a player picks up an item (via `pickup`/`get`/`take` command), the QuestTracker checks the item's `TemplateID` (or `ID` if no template) against `targetId`. Each pickup increments by 1.

**Note:** The tracker counts pickup events, not current inventory count. If a player picks up 3 items, drops one, and picks up another, the count is 4.

```json
{
  "id": "collect_pelts",
  "type": "collect",
  "description": "Collect 5 Wolf Pelts",
  "targetId": "ITM0013",
  "targetName": "Wolf Pelt",
  "amount": 5,
  "order": 1
}
```

### Deliver (`"deliver"`)

Player must bring an item to a specific NPC. Tracked when the player talks to the delivery NPC.

| Field | Required | Description |
|-------|----------|-------------|
| `targetId` | **Yes** | Item **template** ID to deliver. |
| `targetName` | Recommended | Display name of item. |
| `amount` | No | Defaults to 1. |
| `deliverToNpcId` | **Yes** | NPC **template** ID to deliver to. |
| `deliverToNpcName` | Recommended | Display name of target NPC. |

**How tracking works:** When a player talks to an NPC, the QuestTracker checks if any active deliver objectives target this NPC (matching `TemplateID` or instance `ID`). If so, the objective is incremented. The tracker does **not** currently remove the item from inventory — use an `onCompleteScriptId` to handle item removal if needed.

```json
{
  "id": "deliver_letter",
  "type": "deliver",
  "description": "Deliver the Sealed Letter to Captain Aldric",
  "targetId": "ITM0025",
  "targetName": "Sealed Letter",
  "deliverToNpcId": "NPC0004",
  "deliverToNpcName": "Captain Aldric",
  "amount": 1,
  "order": 1
}
```

### Visit (`"visit"`)

Player must enter a specific room.

| Field | Required | Description |
|-------|----------|-------------|
| `targetId` | **Yes** | Room ID (e.g., `"R0301"`). |
| `targetName` | Recommended | Display name (e.g., `"Marsh Gate Trail"`). |
| `amount` | No | Defaults to 1. Typically 1 — just visit once. |

**How tracking works:** When a player moves to a new room, the QuestTracker checks the room's `ID` against `targetId`. Entry increments by 1.

```json
{
  "id": "visit_marsh",
  "type": "visit",
  "description": "Travel to Gloomfen Marsh",
  "targetId": "R0301",
  "targetName": "Marsh Gate Trail",
  "amount": 1,
  "order": 1
}
```

### Talk (`"talk"`)

Player must talk to an NPC, optionally reaching a specific dialog node.

| Field | Required | Description |
|-------|----------|-------------|
| `targetId` | No | NPC template ID. If empty, any NPC match on dialog node. |
| `targetName` | Recommended | Display name. |
| `dialogNodeId` | No | Specific dialog node ID to reach. If empty, any conversation counts. |
| `amount` | No | Defaults to 1. |

**How tracking works:** When a player selects a dialog option, the QuestTracker checks: (1) NPC match if `targetId` is set, and (2) dialog node match if `dialogNodeId` is set. Both conditions must pass (empty fields are treated as wildcard match).

```json
{
  "id": "talk_archivist",
  "type": "talk",
  "description": "Speak with Archivist Maren about the ruins",
  "targetId": "NPC0007",
  "targetName": "Archivist Maren",
  "dialogNodeId": "ruins_info",
  "amount": 1,
  "order": 1
}
```

### Custom (`"custom"`)

Progress is managed entirely by a Lua script. The QuestTracker does not auto-increment custom objectives.

| Field | Required | Description |
|-------|----------|-------------|
| `checkScriptId` | **Yes** | Lua script ID that manages this objective. |
| `description` | **Yes** | Shown to player. |
| `amount` | **Yes** | Required count (script must call `setProgress` to increment). |

**How to update from Lua:**

```lua
-- In any script context (room enter, item use, NPC death, etc.)
local characterID = ctx.character.id
local questID = "QST0005"
local objectiveID = "custom_puzzle"

-- Set progress directly
tales.quests.setProgress(characterID, questID, objectiveID, 1)

-- Or check and set
local progress = tales.quests.getProgress(characterID, questID)
if progress then
    tales.quests.setProgress(characterID, questID, objectiveID, 1)
end
```

```json
{
  "id": "custom_puzzle",
  "type": "custom",
  "description": "Solve the Ancient Rune Puzzle",
  "checkScriptId": "SCR0015",
  "amount": 1,
  "order": 1
}
```

### Script Override for Standard Types

Any standard objective type (`kill`, `collect`, etc.) can have a `checkScriptId` set. When present, the Lua script acts as an additional check or side effect alongside the built-in tracker. The built-in tracking still fires normally.

---

## Quest Source Configuration

The `source` field determines how a player can acquire the quest.

### NPC Source (`"npc"`)

The quest is offered by an NPC. When the player talks to this NPC, quest dialog options are **automatically injected** into the conversation.

```json
"source": {
  "type": "npc",
  "npcId": "NPC0004"
}
```

**Auto-injected dialog behavior:**
- If quest is available and prerequisites are met: an "offer" option appears using `acceptDialogText`
- If quest is active and in progress: `progressDialogText` is shown
- If quest is active and all objectives complete: a "turn-in" option appears using `completeDialogText`
- If `acceptDialogText` is empty, a default message is used: `"I have a task for you: {quest name}"`
- NPCs without their own dialog tree will get a synthetic dialog with just the quest options

### Item Source (`"item"`)

The quest is granted when the player picks up or uses a specific item. (Currently tracked via Lua scripts — use a room action script or item use script to call `tales.quests.grantQuest(characterID, questID)`.)

```json
"source": {
  "type": "item",
  "itemId": "ITM0020"
}
```

### Auto Source (`"auto"`)

The quest is automatically available in the player's available quest list. Useful for tutorial or zone-entry quests.

```json
"source": {
  "type": "auto"
}
```

### Script Source (`"script"`)

The quest is granted by a Lua script. Use `tales.quests.grantQuest(characterID, questID)` from any script to make the quest active.

```json
"source": {
  "type": "script"
}
```

---

## Rewards

Rewards are granted when the quest is completed (all objectives met and quest turned in or auto-completed).

```json
"rewards": {
  "xp": 100,
  "gold": 25,
  "itemTemplateIds": ["ITM0012", "ITM0015"]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `xp` | int32 | Experience points granted. |
| `gold` | int64 | Gold currency granted. |
| `itemTemplateIds` | string[] | Item template IDs — instances are created from these templates and added to the player's inventory. |

All reward fields are optional. Omit or set to zero/empty to skip that reward type.

---

## Prerequisites

### Required Quests

Set `requiredQuestIds` to enforce quest ordering. The quest will only be available if **all** listed quests have `completed` status for the character.

```json
"requiredQuestIds": ["QST0001", "QST0002"]
```

### Required Level

Set `requiredLevel` to restrict by character level. This is checked by `GetAvailableQuests` but not currently enforced at the `AcceptQuest` API level (the service only checks prerequisites and existing progress).

```json
"requiredLevel": 5
```

### Chain Example

```
QST0001 (no prereqs) → QST0002 (requires QST0001) → QST0003 (requires QST0002)
```

```json
// QST0001 - first in chain
{ "name": "Prove Yourself", "requiredQuestIds": [] }

// QST0002 - requires QST0001
{ "name": "The Real Job", "requiredQuestIds": ["<QST0001-id>"] }

// QST0003 - requires QST0002
{ "name": "Into the Depths", "requiredQuestIds": ["<QST0002-id>"] }
```

**Note:** `requiredQuestIds` uses the UUID generated by the server, not human-readable IDs. When creating quest chains programmatically, create them in order and capture the returned `id` to use in the next quest's `requiredQuestIds`.

---

## NPC Dialog Integration

When a player talks to an NPC (`talk <npc>` command), the system checks:

1. Does this NPC have any quests where `source.npcId` matches the NPC's template ID or instance ID?
2. For each matching quest, what is the player's current status?

Based on this, dialog options are automatically injected:

| Player State | Injected Option | Text Source |
|-------------|-----------------|-------------|
| Quest available, not yet accepted | `"[Quest] {quest name}"` | `acceptDialogText` or default |
| Quest active, objectives incomplete | `"[Quest] {quest name} (in progress)"` | `progressDialogText` or default |
| Quest active, all objectives complete | `"[Quest] Turn in: {quest name}"` | `completeDialogText` or default |
| Quest completed (non-repeatable) | Nothing injected | — |

### Tips for Dialog Text

- Write `acceptDialogText` as the NPC explaining the problem and asking for help
- Write `progressDialogText` as the NPC checking in on progress
- Write `completeDialogText` as the NPC acknowledging success and giving thanks
- These texts are shown as NPC speech in the dialog system
- If you also have a manually-authored dialog tree (via `dialogID` on the NPC), the quest options appear alongside the regular dialog options

### Manual Dialog Integration

For complex quest dialogs that need branching conversation trees, you can author a full dialog tree and use conditions like `quest:QST0003_active` in dialog options. See `DLG0003_captain_aldric.yaml` for an example of manually wired quest dialog with conditions.

---

## API Reference

All endpoints require authentication via `Authorization: Bearer <token>` header.

### Quest Definition CRUD (Creator/Admin Role Required for Writes)

#### List All Quests

```
GET /api/quests
```

Returns: `Quest[]`

#### Get Quest by ID

```
GET /api/quests/:id
```

Returns: `Quest` or `404`

#### Create Quest

```
POST /api/quests
Content-Type: application/json
Body: Quest (without id)
```

Returns: `Quest` (with generated `id`, `created`, `updated` timestamps)

#### Update Quest

```
PUT /api/quests/:id
Content-Type: application/json
Body: Quest (full object)
```

Returns: `{"status": "updated quest"}`

#### Delete Quest

```
DELETE /api/quests/:id
```

Returns: `{"status": "deleted"}`

### Quest Progress (Player Level)

#### Get Quest Log

```
GET /api/quest-progress/:characterId
```

Returns: `QuestProgress[]` — all quest progress entries for the character.

#### Accept Quest

```
POST /api/quest-progress/:characterId/accept/:questId
```

Returns: `QuestProgress` — newly created progress entry with `status: "active"`.

**Error cases:**
- Quest not found → 400
- Quest already active → 400
- Quest already completed (non-repeatable) → 400

#### Abandon Quest

```
POST /api/quest-progress/:characterId/abandon/:questId
```

Returns: `{"status": "quest abandoned"}`

**Error cases:**
- Quest not in quest log → 400
- Quest not active → 400

---

## YAML Format for Import

For bulk content creation, quests can be authored as YAML files following the project's data conventions. Place files in the data directory with the naming pattern `QST####.yaml`.

### ID Convention

Quest IDs follow the pattern: `QST####` (4-digit zero-padded number).

| Range | Purpose |
|-------|---------|
| QST0001–QST0099 | Tutorial / Zone 0 (Catacombs) |
| QST0100–QST0199 | Zone 1 (Meadows / Forest Path) |
| QST0200–QST0299 | Zone 2 (Oldtown) |
| QST0300–QST0399 | Zone 3 (Gloomfen Marsh) |
| QST0900–QST0999 | Daily / repeatable quests |

### YAML Quest File

```yaml
id: QST0201
name: Sewer Sweep
description: >
  Captain Aldric needs someone to clear the creatures infesting
  Oldtown's sewer system. The maintenance workers refuse to go
  back down until it is safe.
category: side
level: 2
repeatable: false

source:
  type: npc
  npcId: NPC0004

objectives:
  - id: kill_sewer_rats
    type: kill
    description: Kill 8 Sewer Rats
    targetId: ENM0001
    targetName: Catacomb Rat
    amount: 8
    order: 1

  - id: visit_sewer_end
    type: visit
    description: Reach the sewer terminus
    targetId: R0230
    targetName: Sewer Terminus
    amount: 1
    order: 2

rewards:
  xp: 75
  gold: 25
  itemTemplateIds:
    - ITM0012

requiredQuestIds: []
requiredLevel: 2

acceptDialogText: >
  Clear out whatever is attacking my workers. I will pay twenty-five
  marks for results. The sewer grate is near Bramwick's shop.
progressDialogText: >
  The sewers still are not clear. Get back down there.
completeDialogText: >
  The workers report it is quiet down there now. Good work.
  Here is your payment as promised.
```

### Importing YAML into the Database

YAML files are converted to JSON and posted to the API. For bulk import:

```bash
# Convert YAML to JSON and POST
for file in data/quests/QST*.yaml; do
  json=$(yq -o json "$file")
  curl -X POST http://localhost:8010/api/quests \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "$json"
done
```

Or use the admin import endpoint if a full content bundle is being loaded.

---

## Complete Examples

### Example 1: Kill Quest (Simple)

A basic combat quest with a single kill objective.

```json
{
  "name": "Pest Control",
  "description": "Mira's cellar is infested with rats. Time to earn your keep.",
  "category": "side",
  "level": 1,
  "source": {
    "type": "npc",
    "npcId": "NPC0001"
  },
  "objectives": [
    {
      "id": "kill_rats",
      "type": "kill",
      "description": "Kill 5 Catacomb Rats",
      "targetId": "ENM0001",
      "targetName": "Catacomb Rat",
      "amount": 5,
      "order": 1
    }
  ],
  "rewards": {
    "xp": 40,
    "gold": 8
  },
  "acceptDialogText": "Those rats in my cellar are getting bolder every night. Kill five of them and drinks are on the house.",
  "completeDialogText": "Finally, some peace. Here, you have earned this."
}
```

### Example 2: Multi-Objective Quest

A quest with kill + collect + visit objectives.

```json
{
  "name": "Marsh Expedition",
  "description": "The Archivist wants samples from Gloomfen Marsh for her research. Dangerous work, but the pay is good.",
  "category": "side",
  "level": 3,
  "source": {
    "type": "npc",
    "npcId": "NPC0007"
  },
  "objectives": [
    {
      "id": "kill_wolves",
      "type": "kill",
      "description": "Kill 3 Meadow Wolves",
      "targetId": "ENM0002",
      "targetName": "Meadow Wolf",
      "amount": 3,
      "order": 1
    },
    {
      "id": "collect_pelts",
      "type": "collect",
      "description": "Collect 3 Wolf Pelts",
      "targetId": "ITM0013",
      "targetName": "Wolf Pelt",
      "amount": 3,
      "order": 2
    },
    {
      "id": "reach_deep_marsh",
      "type": "visit",
      "description": "Reach the Deep Marsh",
      "targetId": "R0310",
      "targetName": "Deep Marsh",
      "amount": 1,
      "order": 3
    }
  ],
  "rewards": {
    "xp": 120,
    "gold": 40,
    "itemTemplateIds": ["ITM0012"]
  },
  "acceptDialogText": "I need wolf pelts from the marsh and a survey of the deeper areas. It is dangerous, but I will make it worth your while.",
  "progressDialogText": "Have you gathered the samples? The deeper marsh is where the real data lies.",
  "completeDialogText": "Excellent specimens. This will advance my research considerably. Your payment, as promised."
}
```

### Example 3: Delivery Quest

A quest requiring the player to bring an item to another NPC.

```json
{
  "name": "Urgent Message",
  "description": "Mira has a sealed letter that must reach Captain Aldric immediately.",
  "category": "main",
  "level": 1,
  "source": {
    "type": "npc",
    "npcId": "NPC0001"
  },
  "objectives": [
    {
      "id": "deliver_letter",
      "type": "deliver",
      "description": "Deliver the Sealed Letter to Captain Aldric",
      "targetId": "ITM0025",
      "targetName": "Sealed Letter",
      "deliverToNpcId": "NPC0004",
      "deliverToNpcName": "Captain Aldric",
      "amount": 1,
      "order": 1
    }
  ],
  "rewards": {
    "xp": 30,
    "gold": 5
  },
  "acceptDialogText": "Take this letter to Captain Aldric at the guard post. It is urgent — do not read it.",
  "completeDialogText": "Good, you delivered it. Here, for your trouble."
}
```

### Example 4: Talk / Dialog Quest

A quest requiring the player to speak with multiple NPCs.

```json
{
  "name": "Gather Intelligence",
  "description": "Captain Aldric wants you to question the townsfolk about the recent disappearances.",
  "category": "main",
  "level": 2,
  "source": {
    "type": "npc",
    "npcId": "NPC0004"
  },
  "objectives": [
    {
      "id": "talk_mira",
      "type": "talk",
      "description": "Question Mira Thornwood",
      "targetId": "NPC0001",
      "targetName": "Mira Thornwood",
      "amount": 1,
      "order": 1
    },
    {
      "id": "talk_bramwick",
      "type": "talk",
      "description": "Question Bramwick",
      "targetId": "NPC0002",
      "targetName": "Bramwick",
      "amount": 1,
      "order": 2
    },
    {
      "id": "talk_archivist",
      "type": "talk",
      "description": "Report findings to Archivist Maren",
      "targetId": "NPC0007",
      "targetName": "Archivist Maren",
      "dialogNodeId": "disappearance_report",
      "amount": 1,
      "order": 3
    }
  ],
  "rewards": {
    "xp": 60,
    "gold": 15
  },
  "acceptDialogText": "I need eyes and ears in town. Talk to the innkeeper, the general store merchant, and report what you learn to the Archivist.",
  "progressDialogText": "Have you spoken to everyone yet?",
  "completeDialogText": "Good intelligence. This confirms my suspicions."
}
```

### Example 5: Quest Chain (3 quests)

Three quests that must be completed in sequence.

```bash
# Step 1: Create the first quest (no prerequisites)
QUEST1=$(curl -s -X POST http://localhost:8010/api/quests \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Strange Noises",
    "description": "Investigate the strange sounds coming from the sewers.",
    "category": "main",
    "level": 2,
    "source": {"type": "npc", "npcId": "NPC0004"},
    "objectives": [
      {"id": "visit_sewer", "type": "visit", "description": "Enter the Sewers", "targetId": "R0220", "amount": 1, "order": 1}
    ],
    "rewards": {"xp": 30},
    "acceptDialogText": "We have been hearing noises from the sewers. Go investigate."
  }' | jq -r '.id')

echo "Quest 1 ID: $QUEST1"

# Step 2: Create the second quest (requires quest 1)
QUEST2=$(curl -s -X POST http://localhost:8010/api/quests \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Clear the Path\",
    \"description\": \"Clear the sewer creatures blocking the deeper tunnels.\",
    \"category\": \"main\",
    \"level\": 3,
    \"source\": {\"type\": \"npc\", \"npcId\": \"NPC0004\"},
    \"objectives\": [
      {\"id\": \"kill_rats\", \"type\": \"kill\", \"description\": \"Kill 10 Sewer Rats\", \"targetId\": \"ENM0001\", \"amount\": 10, \"order\": 1}
    ],
    \"rewards\": {\"xp\": 80, \"gold\": 20},
    \"requiredQuestIds\": [\"$QUEST1\"],
    \"acceptDialogText\": \"Now we know what is down there. Clear them out.\"
  }" | jq -r '.id')

echo "Quest 2 ID: $QUEST2"

# Step 3: Create the third quest (requires quest 2)
curl -s -X POST http://localhost:8010/api/quests \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"The Source\",
    \"description\": \"Find and destroy whatever is spawning the creatures.\",
    \"category\": \"main\",
    \"level\": 4,
    \"source\": {\"type\": \"npc\", \"npcId\": \"NPC0004\"},
    \"objectives\": [
      {\"id\": \"kill_boss\", \"type\": \"kill\", \"description\": \"Defeat the Hollow Knight\", \"targetId\": \"ENM0009\", \"amount\": 1, \"order\": 1},
      {\"id\": \"collect_shard\", \"type\": \"collect\", \"description\": \"Collect the Hollow Knight Shard\", \"targetId\": \"ITM0020\", \"amount\": 1, \"order\": 2}
    ],
    \"rewards\": {\"xp\": 200, \"gold\": 50, \"itemTemplateIds\": [\"ITM0015\"]},
    \"requiredQuestIds\": [\"$QUEST2\"],
    \"acceptDialogText\": \"The path is clear. Now find the source and end this.\",
    \"completeDialogText\": \"It is done. Oldtown owes you a debt.\"
  }"
```

### Example 6: Daily Repeatable Quest

```json
{
  "name": "Daily Hunt",
  "description": "The blacksmith needs a steady supply of boar tusks for her work.",
  "category": "daily",
  "level": 2,
  "repeatable": true,
  "source": {
    "type": "npc",
    "npcId": "NPC0003"
  },
  "objectives": [
    {
      "id": "collect_tusks",
      "type": "collect",
      "description": "Collect 3 Boar Tusks",
      "targetId": "ITM0014",
      "targetName": "Boar Tusk",
      "amount": 3,
      "order": 1
    }
  ],
  "rewards": {
    "xp": 25,
    "gold": 15
  },
  "acceptDialogText": "I always need more tusks. Bring me three and I will pay the usual rate.",
  "completeDialogText": "Good quality. Same time tomorrow?"
}
```

### Example 7: Custom Script Objective

A quest with a puzzle solved via Lua script.

```json
{
  "name": "The Rune Puzzle",
  "description": "Ancient runes guard a secret passage. Decipher them to proceed.",
  "category": "main",
  "level": 1,
  "source": {
    "type": "auto"
  },
  "objectives": [
    {
      "id": "solve_rune",
      "type": "custom",
      "description": "Solve the Ancient Rune Puzzle",
      "checkScriptId": "SCR0007",
      "amount": 1,
      "order": 1
    }
  ],
  "rewards": {
    "xp": 50
  }
}
```

The Lua script (`SCR0007`) would update progress:

```lua
-- Inside the rune puzzle script, after player solves it:
local characterID = ctx.character.id

-- Check if quest is active
if tales.quests.isActive(characterID, "QST0005") then
    tales.quests.setProgress(characterID, "QST0005", "solve_rune", 1)
end
```

---

## Entity ID Reference

### Existing NPCs (Quest Givers)

| ID | Name | Location | Role |
|----|------|----------|------|
| NPC0001 | Mira Thornwood | R0203 (Weary Wanderer) | Innkeeper, merchant |
| NPC0002 | Bramwick | R0202 (General Store) | Merchant |
| NPC0003 | Kara Ironhand | R0204 (Smithy) | Blacksmith |
| NPC0004 | Captain Aldric | R0205 (Guard Post) | Guard captain |
| NPC0005 | Guardsman Thom | R0201 (Town Gate) | Gate guard |
| NPC0007 | Archivist Maren | R0208 (Archive) | Scholar |
| NPC0008 | The Stranger | R0210 | Mysterious figure |
| NPC0013 | Darius Coinsworth | R0213 | Wealthy merchant |

### Existing Enemies (Kill Targets)

| ID | Name | Level | Type | Location |
|----|------|-------|------|----------|
| ENM0001 | Catacomb Rat | 1 | Beast | Z00 (Catacombs) |
| ENM0002 | Meadow Wolf | 2 | Beast | Z01 (Meadows) |
| ENM0003 | Wild Boar | 2 | Beast | Z01 (Meadows) |
| ENM0005 | Thornback Bear | 5 | Beast (miniboss) | Z01 (Forest) |
| ENM0006 | Night Whisper | 4 | Spirit | Z02 (Oldtown) |
| ENM0007 | Alley Thug | 2 | Humanoid | Z02 (Oldtown) |
| ENM0009 | The Hollow Knight | 6 | Construct (boss) | Z00 (Catacombs) |

### Existing Items (Collect / Reward Targets)

| ID | Name | Type | Notes |
|----|------|------|-------|
| ITM0001 | Dusty Torch | Collectible | Stackable, max 5 |
| ITM0006 | Copper Bits | Currency | Stackable, max 999 |
| ITM0010 | Leather Cap | Armor (head) | 1 defense |
| ITM0011 | Padded Vest | Armor (chest) | 2 defense |
| ITM0012 | Weak Health Potion | Consumable | Stackable, max 10 |
| ITM0013 | Wolf Pelt | Crafting | Stackable, max 10 |
| ITM0014 | Boar Tusk | Crafting | Stackable, max 20 |
| ITM0015 | Silver Mark | Currency | Stackable, max 999 |
| ITM0018 | Rat Tail | Junk/Trophy | Stackable, max 50 |
| ITM0020 | Hollow Knight Shard | Quest | Boss loot |
| ITM0021 | Sturdy Branch | Weapon | 2 damage, starter |

### Zones and Key Rooms

| Zone | ID Pattern | Key Rooms |
|------|-----------|-----------|
| Catacombs (Tutorial) | Z00, R0001–R0006 | R0001 (Start), R0005 (Rat Nest), R0006 (Stairwell) |
| Meadows / Forest | Z01, R0101–R0112 | R0101 (Emergence), R0106 (Deep Forest) |
| Oldtown (Hub) | Z02, R0201–R0222 | R0201 (Gate), R0203 (Tavern), R0205 (Guard Post) |
| Gloomfen Marsh | Z03, R0301–R0321 | R0301 (Marsh Gate), R0310 (Deep Marsh) |
| Gloomfen Depths | Z19, R1901+ | Underground areas |

---

## Lua API for Quests

The `tales.quests` module is available in all Lua script contexts.

| Function | Returns | Description |
|----------|---------|-------------|
| `tales.quests.get(questID)` | Quest table or nil | Get quest definition by ID |
| `tales.quests.getAll()` | Quest[] table | Get all quest definitions |
| `tales.quests.accept(characterID, questID)` | QuestProgress or nil | Accept a quest |
| `tales.quests.complete(characterID, questID)` | QuestProgress or nil | Force-complete a quest |
| `tales.quests.getProgress(characterID, questID)` | QuestProgress or nil | Get quest progress |
| `tales.quests.setProgress(characterID, questID, objectiveID, amount)` | boolean | Set objective current value |
| `tales.quests.isActive(characterID, questID)` | boolean | Check if quest is active |
| `tales.quests.isCompleted(characterID, questID)` | boolean | Check if quest is completed |
| `tales.quests.grantQuest(characterID, questID)` | QuestProgress or nil | Accept quest bypassing source |
| `tales.quests.abandon(characterID, questID)` | boolean | Abandon active quest |
| `tales.quests.getQuestLog(characterID)` | QuestProgress[] | Get all progress entries |

### Common Lua Patterns

**Grant quest from room entry script:**
```lua
local characterID = ctx.character.id
if not tales.quests.isActive(characterID, "QST0101") and
   not tales.quests.isCompleted(characterID, "QST0101") then
    tales.quests.grantQuest(characterID, "QST0101")
    tales.game.msgToUser(ctx.user.id, "New quest: Explore the Meadows")
end
```

**Complete custom objective from room action:**
```lua
local characterID = ctx.character.id
if tales.quests.isActive(characterID, "QST0005") then
    tales.quests.setProgress(characterID, "QST0005", "solve_rune", 1)
    tales.game.msgToUser(ctx.user.id, "Quest Updated: The runes glow with recognition.")
end
```

**On-complete script (referenced via `onCompleteScriptId`):**
```lua
-- Remove quest items from inventory, spawn new NPC, reveal exit, etc.
local characterID = ctx.character.id
tales.game.msgToUser(ctx.user.id, "The way forward is now clear.")
tales.game.revealExit(ctx.room.id, "hidden_passage")
```

---

## Workflow: Creating a Quest End-to-End

### Programmatic Workflow (for AI agents)

1. **Identify dependencies** — Determine which NPCs, enemies, items, and rooms the quest references. Verify they exist via `GET /api/npcs`, `GET /api/items`, `GET /api/rooms`.

2. **Create any missing entities first** — If the quest needs a new item type (e.g., a quest-specific drop), create the item template via `POST /api/items` before creating the quest.

3. **Create the quest** — `POST /api/quests` with the full quest JSON. Capture the returned `id`.

4. **Create follow-up quests** — If building a chain, use the captured `id` in `requiredQuestIds` of subsequent quests.

5. **Verify** — `GET /api/quests` to confirm all quests are stored. Check that `source.npcId` references valid NPCs.

### Creator UI Workflow

1. Navigate to **Creator > Quests** tab
2. Click **New** to create a new quest
3. Fill in the General section (name, description, category, level)
4. Set the Source (select NPC from dropdown, or choose item/auto/script)
5. Add Objectives — click "Add Objective" for each, select type, fill in type-specific fields
6. Configure Rewards — set XP, gold, add item template IDs
7. Optionally set Prerequisites — add required quest IDs and level
8. Write Dialog Text — accept, progress, and complete messages
9. Click **Save**

---

## Troubleshooting

### Quest dialog options not appearing

- Verify `source.type` is `"npc"` and `source.npcId` matches the NPC's template ID (e.g., `"NPC0004"`, not an instance UUID)
- The NPC must exist as a spawned instance in a room for players to talk to
- Check that the player has met prerequisites (`requiredQuestIds` all completed)

### Kill objective not tracking

- `targetId` must match the NPC's **template** ID (e.g., `"ENM0001"`), not a spawned instance ID
- The NPC must die in combat (killed via the combat system, not despawned)
- The player must be a participant in the combat that killed the NPC

### Collect objective not tracking

- `targetId` must match the item's **template** ID (e.g., `"ITM0013"`)
- The player must use the `pickup`/`get`/`take` command — items granted via scripts or loot do not trigger collect tracking
- Items granted via `tales.game.giveItem()` do **not** trigger the collect tracker

### Visit objective not tracking

- `targetId` must be the exact room ID (e.g., `"R0301"`)
- The player must physically move to the room via an exit (teleport via scripts may not trigger if the script doesn't call the move command)

### Quest not available to player

- Check `requiredQuestIds` — all listed quests must have `completed` status
- Check if the quest was previously completed and `repeatable` is `false`
- Check if the quest is already `active` for the character

### Rewards not granted

- Quest rewards (XP, gold, items) are granted by the game engine when the quest status changes to `completed`. This happens when:
  - Player turns in the quest at the source NPC (dialog auto-injection)
  - Quest is completed via Lua (`tales.quests.complete()`)
  - Quest is completed via the API (`POST .../accept` creates progress, completion logic runs server-side)
- Check server logs for errors during reward granting
