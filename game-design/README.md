# Game Design Documentation

This folder contains all game design, feature planning, and implementation documentation for TalesMUD.

## Contents

### [GAME_DESIGN.md](GAME_DESIGN.md)
The comprehensive game design document and MVP backlog. Contains:
- MVP definition and scope
- Current implemented features (characters, items, NPCs, dialogs, scripting)
- Design principles
- Epic breakdowns for upcoming features:
  - Enemy NPCs
  - Combat system
  - Combat instances
  - Items, containers, and loot
  - Inventory & equipment
  - Merchants & trading
- Command reference
- Data model specifications
- Implementation stories (JIRA-style)

### [SCRIPTING.md](SCRIPTING.md)
Documentation for the Lua scripting system:
- Lua engine architecture (gopher-lua VM pool)
- Complete API reference for 7 modules:
  - `tales.items` - Item and template operations
  - `tales.rooms` - Room queries
  - `tales.characters` - Character management
  - `tales.npcs` - NPC operations
  - `tales.dialogs` - Dialog/conversation management
  - `tales.game` - Messaging system
  - `tales.utils` - Utilities (dice rolls, random, UUID, etc.)
- Event system (28 event types)
- Script context and triggers
- Example scripts and best practices
- Sandbox restrictions and security

### [CHARACTER_TEMPLATES.md](CHARACTER_TEMPLATES.md)
Documentation for the character template system:
- Template entity structure
- Available races and classes
- Attributes system (STR, DEX, CON, INT, WIS, CHA)
- Equipment slots
- Starting items configuration
- REST API reference
- Design guidelines for balanced templates

### [DIALOG_SYSTEM.md](DIALOG_SYSTEM.md)
Documentation for the dialog and conversation system:
- Dialog tree structure
- Conversation state management
- Features: random text, conditional options, one-time choices
- Mustache template variables
- NPC dialog integration
- Idle/ambient dialogs
- Commands: `talk`, numeric selection
- Scripting integration
- Design guidelines for quest dialogs

### [WORLD_MAP_IMPLEMENTATION.md](WORLD_MAP_IMPLEMENTATION.md)
Technical documentation for the interactive world map feature:
- Implementation overview
- Backend graph data endpoint
- Position calculation algorithm
- Frontend Svelte Flow integration
- Custom room node components
- Visual design specifications
- Usage guide

## Quick Reference

| System | Status | Documentation |
|--------|--------|---------------|
| Characters | Implemented | GAME_DESIGN.md, CHARACTER_TEMPLATES.md |
| Rooms | Implemented | GAME_DESIGN.md, WORLD_MAP_IMPLEMENTATION.md |
| NPCs | Implemented | GAME_DESIGN.md, DIALOG_SYSTEM.md |
| Dialogs | Implemented | DIALOG_SYSTEM.md |
| Items | Implemented | GAME_DESIGN.md |
| Scripting (Lua) | Implemented | SCRIPTING.md |
| Combat | Planned | GAME_DESIGN.md (Epics A-C) |
| Merchants | Planned | GAME_DESIGN.md (Epic F) |

## Related Documentation

For technical architecture and API documentation, see:
- [`../ARCHITECTURE.md`](../ARCHITECTURE.md) - System architecture and component design
- [`../PROJECT.md`](../PROJECT.md) - Project overview and getting started guide
- [`../AGENTS.md`](../AGENTS.md) - Claude Code agent instructions
