# TalesMUD Six-System Build Roadmap

## Objective

Build forward on all six selected systems without replacing the current architecture:
NPC behavior and dialog integration, itemization/world interaction, quest delivery,
combat reliability and UX, Creator quality, and player session/multiplayer UX.

## Working Rules

- Preserve the existing dirty worktree and build on current patterns.
- Prefer additive foundations and validations before risky rewrites.
- Keep successful API response shapes stable unless the existing shape is a bug.
- Update `PROJECT.md`, `ARCHITECTURE.md`, and `FEATURES.md` when behavior or contracts change.
- Verify each slice with focused backend tests and frontend builds where touched.

## Goal 1: NPC Behavior + Dialog Integration

### Backend
- Implement deterministic patrol and bounded wander behavior in `game_npcs.go`.
- Trigger idle dialog messages from NPC update ticks with rate limiting.
- Ensure `talk` works consistently with resident and spawned NPC instances.
- Expose NPC interaction flags in room messages: enemy, merchant, quest giver, dialog available.

### Frontend
- Surface interaction affordances in room/entity panels.
- Keep dialog state clear when rooms change, NPCs move, or combat starts.
- Add Creator validation for NPC dialog IDs, idle dialog IDs, patrol paths, and merchant stock references.

## Goal 2: Itemization + World Interaction

### Backend
- Finish room item command coverage: pickup/drop/look/use/equip/unequip/container interactions.
- Ensure loot table rolls integrate with enemy death and stackable item quantities.
- Validate merchant stock references, prices, accepted item types, restock timing, and loot table IDs.

### Frontend
- Improve ground-item and inventory displays with clear quantity, equipped, bound, and usable states.
- Add Creator validation around item templates, loot tables, room items, merchant stock, and container items.

## Goal 3: Quest Delivery + Quest UX

### Backend
- Connect quest sources to NPCs and dialogs.
- Validate quest prerequisites, objective target IDs, reward item IDs, and completion paths.
- Keep quest progress authorization owner/admin-only.

### Frontend
- Improve quest log status grouping, objective progress, and turn-in clarity.
- Add Creator diagnostics for impossible or broken quest definitions.

## Goal 4: Combat Reliability + Player Combat UX

### Backend
- Stabilize combat outcome tests without arbitrary tuning hidden in unrelated changes.
- Validate enemy combat data, XP/gold rewards, loot table references, and respawn settings.
- Harden death, flee, reward, and room update boundaries.

### Frontend
- Clarify combat state, targets, queued actions, rewards, and death/respawn feedback.
- Keep mobile combat actions reachable without covering core room context.

## Goal 5: Creator Quality Layer

### Backend
- Add a world validation service and API that reports errors/warnings across rooms, NPCs,
  items, loot tables, quests, dialogs, scripts, spawners, and settings.
- Reuse the validator from import, admin diagnostics, and Creator health checks where possible.

### Frontend
- Add a Creator health/check panel showing validation issues grouped by entity type and severity.
- Link diagnostics to the relevant editor where feasible.

## Goal 6: Player Session, Character, and Multiplayer UX

### Backend
- Keep user profile updates token-derived and role-safe.
- Harden reconnect, online/offline, last-character, and WebSocket cleanup behavior.
- Add observable session state for online players, active characters, and room presence.

### Frontend
- Improve onboarding, reconnect, character selection, and online player presence.
- Keep guest and authenticated player flows consistent.

## First Build Slice: World Validation Foundation

The first implementation slice is a backend validation service plus Creator surface.
It advances all six goals by making broken references and incomplete content visible before
they break runtime play.

### Backend Deliverables
- New validation package/service with `Issue` records:
  - `severity`: `error` or `warning`
  - `system`: `npc`, `item`, `quest`, `combat`, `creator`, `session`, etc.
  - `entityType`, `entityId`, `field`, `message`
- Validate:
  - NPC patrol rooms, idle dialogs, dialog IDs, merchant item templates, enemy loot tables
  - Room exits/items/actions/spawners references
  - Quest source/objective/reward references and prerequisite quest IDs
  - Loot table item template references
  - Server settings start room
- API: `GET /api/world/validation` for creator/admin users.
- Unit tests using an in-memory SQLite test facade.

### Frontend Deliverables
- Creator navigation entry for "World Health".
- `WorldHealthPanel.svelte` that calls the validation API and groups issues by severity/system.
- Keep styling aligned with existing Creator utility classes and DataTable conventions.

### Verification
- `/usr/local/go/bin/go test ./pkg/service ./pkg/server/handler`
- Creator build with `/home/atla/.nvm/versions/node/v20.20.0/bin/node node_modules/vite/bin/vite.js build`
- Manual smoke check of validation API with an authenticated creator/admin token when available.

## Later Slices

1. NPC movement and idle dialog tick behavior.
2. Quest giver and quest turn-in flow.
3. Item containers and loot/ground-item UI completion.
4. Combat reliability pass and UX polish.
5. Session/reconnect and multiplayer presence polish.
