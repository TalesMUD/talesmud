# Claude Code Instructions

## Documentation Update Requirements

When working on this project, keep documentation in sync with code changes.

### PROJECT.md Updates
Update `PROJECT.md` whenever you:
- Add a new feature or capability
- Modify existing feature behavior
- Remove or deprecate a feature
- Change user-facing functionality
- Add or modify configuration options

### ARCHITECTURE.md Updates
Update `ARCHITECTURE.md` whenever you:
- Add, remove, or rename major components or modules
- Change the system's high-level structure
- Modify data flow between components
- Add or change external dependencies
- Alter database schemas or data models
- Change API contracts or interfaces
- Modify deployment architecture
- Update infrastructure components

### FEATURES.md Updates
Update `FEATURES.md` whenever you:
- Add new entity fields or change data structures
- Add new Lua API functions or scripting capabilities
- Implement new game systems or mechanics
- Add new Creator UI features or components
- Change existing feature behavior
- Add new best practices or patterns
- Implement new per-character state tracking
- Add new script execution contexts

## Update Process
1. After completing a feature or architectural change, review the relevant documentation file
2. Update the documentation to reflect the current state of the project
3. Keep descriptions concise and accurate
4. Ensure any diagrams or examples remain valid

## Creator UI Conventions

### Entity ID Selection — MANDATORY
When building Creator UI forms that need to select another entity by ID (e.g., a room, NPC, item template, dialog, script, quest), **NEVER use a simple `<select>` dropdown**. Instead, always use the `EntitySelectButton` component with a centered `EntitySelectModal` containing a filterable DataTable.

**Why:** The game will eventually have hundreds of rooms, NPCs, items, etc. Simple dropdowns become unusable at scale. The modal-based selector provides per-column filtering, sorting, and search.

**Usage pattern:**
```svelte
import EntitySelectButton from "./EntitySelectButton.svelte";
import { roomColumns } from "./tableColumns.js"; // or npcColumns, scriptColumns, etc.

<EntitySelectButton
  value={someEntityId}
  elements={entityArray}
  columns={roomColumns}
  title="Select Room"
  placeholder="Select a room..."
  on:change={(e) => someEntityId = e.detail}
/>
```

**Note:** Simple enum dropdowns (race, class, difficulty, item quality, etc.) with small fixed option lists are fine to keep as `<select>`. The rule applies only to entity ID references that grow over time.
