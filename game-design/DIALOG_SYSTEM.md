# Dialog System

This document describes the dialog and conversation system in TalesMUD, enabling interactive NPC conversations with branching narratives.

## Overview

The dialog system provides:
- Branching conversation trees with multiple paths
- State tracking (visited nodes, context variables)
- Conditional options based on prior choices
- One-time dialog options
- Template rendering with Mustache syntax
- Random text variations
- Idle/ambient NPC dialog

## Architecture

```
Dialog Entity (stored in DB)
       |
   DialogsService (CRUD)
       |
Conversation Entity (per character + NPC pair)
       |
ConversationsService (state management)
       |
   TalkCommand → DialogSelectCommand
       |
   Game Messages (to player)
```

## Core Entities

### Dialog (Tree Structure)

Dialogs are defined in `pkg/entities/dialogs/dialogs.go`:

```go
type Dialog struct {
    *entities.Entity  // Database ID

    // Identification
    Name   string     // Human-readable name ("Guard Greeting")
    NodeID string     // Tree navigation ID ("main", "quest_offer")

    // Content
    Text           string   // Primary dialog text
    AlternateTexts []string // Random variations
    OrderedTexts   *bool    // Show texts in order on revisits

    // Structure
    Options []*Dialog   // Player choice options (branches)
    Answer  *Dialog     // Auto-response node (no player choice)

    // Conditions
    RequiresVisitedDialogs []string // Nodes that must be visited first
    ShowOnlyOnce           *bool    // Hide after first selection
    IsDialogExit           *bool    // Ends the conversation

    // Metadata
    Created   time.Time
    CreatedBy string
}
```

### Conversation (State)

Conversation state is defined in `pkg/entities/conversations/conversation.go`:

```go
type Conversation struct {
    *entities.Entity

    CharacterID string     // Player's character
    TargetID    string     // NPC or item ID
    TargetType  TargetType // "npc" or "item"
    DialogID    string     // The dialog tree being used

    CurrentNodeID string         // Current position in tree
    VisitedNodes  map[string]int // Visit counts per node
    Context       map[string]string // Template variables

    LastInteracted time.Time
    Created        time.Time
}
```

## Dialog Structure

### Basic Dialog Tree

```yaml
name: "Innkeeper Greeting"
nodeId: "main"
text: "Welcome to the Rusty Dragon, traveler. What can I get you?"
options:
  - nodeId: "drink"
    text: "I'll have an ale."
    answer:
      nodeId: "drink_response"
      text: "Coming right up! That'll be 2 gold."
      is_dialog_exit: true

  - nodeId: "room"
    text: "I need a room for the night."
    answer:
      nodeId: "room_response"
      text: "That'll be 5 gold per night. Shall I prepare one?"
      options:
        - nodeId: "room_yes"
          text: "Yes, please."
          is_dialog_exit: true
        - nodeId: "room_no"
          text: "Maybe later."
          is_dialog_exit: true

  - nodeId: "leave"
    text: "Goodbye."
    is_dialog_exit: true
```

### Node Types

| Type | Description |
|------|-------------|
| **Main Node** | Entry point, `nodeId: "main"` |
| **Option Node** | Player-selectable choice |
| **Answer Node** | NPC's automatic response |
| **Exit Node** | Ends the conversation |

## Key Features

### 1. Random Text Variations

Provide multiple text options for variety:

```yaml
nodeId: "greeting"
text: "Hello there!"
alternateTexts:
  - "Well met, traveler!"
  - "Greetings, friend."
  - "Ah, a visitor!"
```

One text is randomly selected each time the node is displayed.

### 2. Ordered Text Progression

Show different text based on visit count:

```yaml
nodeId: "guard_greeting"
text: "Halt! State your business."
alternateTexts:
  - "You again? What do you want?"
  - "Oh, it's you. Go on through."
orderedTexts: true
```

First visit shows the main text, second visit shows first alternate, etc.

### 3. Conditional Options

Show options only after visiting certain nodes:

```yaml
nodeId: "secret_option"
text: "I heard you know about the hidden passage..."
requires_visited_dialogs:
  - "overheard_conversation"
  - "found_map"
```

### 4. One-Time Options

Hide options after they've been selected:

```yaml
nodeId: "quest_accept"
text: "I'll help you find the artifact."
show_only_once: true
```

### 5. Template Variables

Use Mustache syntax for dynamic content:

```yaml
text: "Greetings, {{PLAYER}}! I am {{NPC}}."
```

Built-in variables:
- `{{PLAYER}}` - Character name
- `{{NPC}}` - NPC name

Custom variables can be set via conversation context.

### 6. Idle Dialog

NPCs can have ambient dialog that plays periodically:

- `DialogID` - Main interactive dialog (triggered by `talk`)
- `IdleDialogID` - Ambient dialog (triggered by timeout)
- `IdleDialogTimeout` - Seconds between idle messages

## Commands

### talk \<npc\>

Start a conversation with an NPC:

```
> talk guard
Guard: "Halt! State your business."
1) I'm just passing through.
2) I have a message for the captain.
3) [Leave]
```

### Numeric Selection

During a conversation, enter a number to select an option:

```
> 1
Guard: "Very well. Move along, citizen."
```

## REST API

### Dialogs

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/dialogs` | List all dialogs |
| GET | `/api/dialogs/:id` | Get dialog by ID |
| POST | `/api/dialogs` | Create dialog |
| PUT | `/api/dialogs/:id` | Update dialog |
| DELETE | `/api/dialogs/:id` | Delete dialog |

### Example Dialog Response

```json
{
  "id": "dialog-123",
  "name": "Guard Greeting",
  "nodeId": "main",
  "text": "Halt! State your business.",
  "options": [
    {
      "nodeId": "passing",
      "text": "I'm just passing through.",
      "answer": {
        "nodeId": "passing_response",
        "text": "Very well. Move along.",
        "is_dialog_exit": true
      }
    },
    {
      "nodeId": "leave",
      "text": "[Leave]",
      "is_dialog_exit": true
    }
  ]
}
```

## Conversation Flow

1. Player types `talk <npc>`
2. System finds NPC by name in current room
3. System loads NPC's `DialogID`
4. System retrieves or creates Conversation state
5. System navigates to `CurrentNodeID` (or "main")
6. Dialog text is rendered with context variables
7. Options are filtered by:
   - `RequiresVisitedDialogs` - must have visited required nodes
   - `ShowOnlyOnce` - not yet selected
8. Filtered options are displayed to player
9. Player selects number
10. System advances `CurrentNodeID`
11. If Answer node: display and continue to its options
12. If `IsDialogExit`: end conversation
13. Repeat from step 6

## Conversation Timeout

Conversations automatically reset after 5 minutes of inactivity. The `LastInteracted` timestamp is updated on each dialog interaction.

## Service Methods

### ConversationsService

```go
// Get or create conversation between character and target
GetOrCreateConversation(characterID, targetID string, targetType TargetType, dialogID string) (*Conversation, error)

// Get the current dialog node
GetCurrentNode(conversation *Conversation) (*Dialog, error)

// Filter options based on visit state and one-time rules
GetFilteredOptions(conversation *Conversation, node *Dialog) []*Dialog

// Move to a new node and mark visited
AdvanceConversation(conversation *Conversation, nodeID string) error

// Reset to main node
ResetConversation(conversation *Conversation) error
```

## NPC Integration

NPCs have two dialog-related fields:

```go
type NPC struct {
    // ... other fields

    // Interactive dialog (triggered by "talk" command)
    DialogID string

    // Ambient/idle dialog (triggered by timeout)
    IdleDialogID      string
    IdleDialogTimeout int32  // seconds
}
```

### Checking NPC Dialog Availability

```go
npc.HasDialog()      // Returns true if DialogID is set
npc.HasIdleDialog()  // Returns true if IdleDialogID is set
```

## Scripting Integration

Dialogs can be accessed via the Lua scripting API:

```lua
-- Get dialog by ID
local dialog = tales.dialogs.get(dialogID)

-- Find dialogs by name
local dialogs = tales.dialogs.findByName("merchant_greeting")

-- Get conversation state
local conv = tales.dialogs.getConversation(characterID, npcID)

-- Set/get context variables
tales.dialogs.setContext(convID, "questAccepted", "true")
local value = tales.dialogs.getContext(convID, "questAccepted")

-- Check visit state
local visited = tales.dialogs.hasVisited(convID, "intro_node")
local count = tales.dialogs.getVisitCount(convID, "intro_node")
```

## Files

| File | Purpose |
|------|---------|
| `pkg/entities/dialogs/dialogs.go` | Dialog entity and tree navigation |
| `pkg/entities/conversations/conversation.go` | Conversation state entity |
| `pkg/service/dialogs.go` | Dialog service |
| `pkg/service/conversations.go` | Conversation service with filtering |
| `pkg/repository/dialogs.go` | Dialog repository |
| `pkg/repository/conversations.go` | Conversation repository |
| `pkg/mudserver/game/commands/talk.go` | Talk command |
| `pkg/mudserver/game/commands/dialog_select.go` | Dialog selection handler |
| `pkg/server/handler/dialogs.go` | REST API handler |

## Design Guidelines

### Keep Trees Shallow
- Avoid deeply nested dialogs (3-4 levels max)
- Use Answer nodes to provide NPC responses
- Return to main topics frequently

### Meaningful Choices
- Make options feel distinct
- Avoid redundant options
- Consider consequences (quest acceptance, reputation, etc.)

### Use Conditional Options Sparingly
- Reserve for meaningful story progression
- Document requirements clearly
- Provide hints about missing requirements

### Test All Paths
- Verify all branches reach exit nodes
- Check that conditional options unlock correctly
- Test edge cases (revisits, timeouts)

## Example: Quest Dialog

```yaml
name: "Blacksmith Quest"
nodeId: "main"
text: "Ah, an adventurer! My apprentice went missing in the caves. Can you help?"
options:
  - nodeId: "accept"
    text: "I'll find your apprentice."
    show_only_once: true
    answer:
      nodeId: "accept_response"
      text: "Thank the gods! He went into the Darkhollow Caves. Be careful!"
      is_dialog_exit: true

  - nodeId: "reward"
    text: "What's in it for me?"
    answer:
      nodeId: "reward_response"
      text: "I'll forge you a fine blade if you bring him back safely."
      options:
        - nodeId: "reward_accept"
          text: "Deal. I'll find him."
          show_only_once: true
          is_dialog_exit: true
        - nodeId: "reward_decline"
          text: "I'll think about it."
          is_dialog_exit: true

  - nodeId: "decline"
    text: "Sorry, I can't help."
    answer:
      nodeId: "decline_response"
      text: "I understand. If you change your mind, please come back."
      is_dialog_exit: true

  # This option only appears after accepting the quest and returning
  - nodeId: "return_success"
    text: "I found your apprentice!"
    requires_visited_dialogs: ["accept", "rescued_apprentice"]
    answer:
      nodeId: "return_response"
      text: "You found him! Thank you! Here's your reward."
      is_dialog_exit: true
```

## Future Enhancements

- Script hooks on dialog transitions
- Dialog-triggered actions (give item, start quest, open shop)
- Voice/sound effects per node
- Dialog editor in admin UI
- Import/export dialog trees as YAML
- Localization support
