# Quest Delivery + Quest UX Design

## Goal

Turn the existing quest data and progress tracking into a cohesive player-facing flow. Players should discover quests through NPCs, accept them with clear context, see objective progress as it happens, know when a quest is ready to turn in, and receive rewards through a server-validated completion path. Creators should get immediate feedback when a quest definition cannot work.

## Current State

The project already has quest definitions, per-character quest progress, quest progress REST endpoints, a `QuestTracker`, NPC dialog quest options, text commands, a player quest log, quest notifications, and a Creator quest editor. The main issue is cohesion:

- Quest definition create/update accepts many impossible or incomplete definitions.
- Deliver objectives can progress when a player talks to the target NPC without proving the character has the required item.
- Quest dialog actions are handled in duplicated paths for embedded dialog options and injected quest options.
- Progress messages and quest log entries are not always enriched with objective descriptions, quest names, or useful next steps.
- The Creator editor lets authors assemble quests but does not summarize validation problems or preview the player delivery flow.

## Scope

This work is a focused cohesion pass over the existing system. It keeps the current quest data model shape, commands, REST routes, dialog system, player quest log surfaces, and Creator editor patterns. It does not introduce a new quest graph editor, a new quest scripting language, or a full visual redesign of the player UI.

## Backend Design

### Quest Validation

Add a server-side validation layer used by `QuestsService.Store` and `QuestsService.Update`. Validation should return actionable errors and reject invalid definitions before they are persisted.

Definition-level rules:

- `name` and `description` are required.
- `source.type` must be one of `npc`, `item`, `auto`, or `script`.
- `source.npcId` is required for NPC-sourced quests.
- `source.itemId` is required for item-sourced quests.
- A quest must have at least one objective.
- `requiredQuestIds` cannot contain the quest's own ID.
- Reward XP and gold cannot be negative.
- Reward item template IDs cannot be empty.

Objective-level rules:

- Objective IDs are required and unique within a quest.
- Objective descriptions are required.
- Objective types must be one of `kill`, `collect`, `deliver`, `visit`, `talk`, or `custom`.
- `amount` defaults to 1 for progress, but persisted definitions cannot use negative amounts.
- Kill objectives require `targetId` for an NPC template or unique NPC.
- Collect objectives require `targetId` for an item template or item.
- Deliver objectives require `targetId` for the item and `deliverToNpcId` for the delivery NPC.
- Visit objectives require `targetId` for the room.
- Talk objectives require either `targetId` or `dialogNodeId`.
- Custom objectives require `checkScriptId`.

When the facade is available, validation should also check that referenced NPCs, items, rooms, scripts, and prerequisite quests exist. Missing references should block saves because they create definitions that can never progress.

### Quest Delivery And Turn-In

NPC-sourced quests remain available through the existing `talk` flow. The NPC quest option list should continue to check both template ID and instance ID so unique NPCs and spawned instances work.

Turn-in rules:

- A quest is turn-in ready only when all objectives are complete.
- Completion must happen server-side through a shared quest action path, not through client trust.
- Repeated completion is blocked unless the quest is repeatable and has been accepted again.
- Rewards are granted after the quest status is marked complete.
- Character, inventory, and quest log update messages are sent after completion.

Deliver objective rules:

- Talking to a delivery NPC should only progress a deliver objective when the character has enough matching items.
- Matching uses item `TemplateID` when present, otherwise item `ID`, matching the existing collect-objective convention.
- The deliver objective amount defaults to 1.
- Delivered items are consumed when the objective completes. Stackable quantities should decrement first; non-stackable instances should be removed until the required amount is satisfied.
- If the character lacks the item, the quest should not progress and should not emit a false completion notification.

### Event Hooks

Keep the existing event hooks for kill, pickup, dialog node, talk-to-NPC, and room entry, but make their behavior more precise:

- Kill: increment matching active kill objectives by the defeated NPC template ID or unique ID.
- Pickup: increment collect objectives by picked-up quantity for stackable items instead of always adding 1.
- Room entry: mark visit objectives once and avoid noisy repeat notifications after completion.
- Dialog node: match talk objectives by NPC and optional dialog node.
- Talk-to-NPC: handle deliver objectives with inventory validation and item consumption.

The `QuestTracker` should only send progress notifications for objectives whose progress actually changed. When all objectives become complete, the event should say the quest is ready to turn in, not that it is completed, unless the player has actually completed the quest through a turn-in action.

### Shared Quest Dialog Actions

Refactor the duplicated quest action handling in `dialog_select.go` into shared helpers:

- Accept quest.
- Show progress.
- Complete quest and grant rewards.
- Send refreshed quest log.
- Build quest accepted/completed/progress messages.

Injected quest options and embedded dialog-tree quest actions should both call the same helper so fixes apply consistently.

### Enriched Quest Log Messages

Quest log message building should consistently include:

- Quest ID and name.
- Status.
- Description.
- Category and level.
- Objective descriptions, current count, required count, and completion state.
- Rewards.
- Accepted and completed timestamps.

The existing API and WebSocket quest log representations should stay compatible with current clients. New fields can be additive if useful, but the implementation should not require a client-side migration.

## Player UI Design

Keep the existing quest log panel/widget and quest notification component. Improve clarity rather than layout:

- Objectives should display readable descriptions from the quest definition.
- Active objectives should show `current/required`.
- Completed objectives should be visually distinct and keep their true counts.
- Active quests with all objectives complete should show a ready-to-turn-in state.
- Rewards should remain visible in active and completed quests.
- Pinned quests should continue to work.

Notifications should become specific:

- Quest accepted: quest name and first objective summary.
- Objective progressed: objective description and `current/required`.
- Objective complete: objective description.
- Ready to turn in: quest name and source NPC name when the quest source is an NPC and the source can be resolved.
- Quest completed: rewards summary.
- Quest abandoned: quest name.

Notification IDs should avoid suppressing legitimate repeated progress on the same quest objective.

## Creator UI Design

Enhance `QuestsEditor.svelte` without replacing the editor:

- Add a validation summary panel near the top of the editor.
- Show blocking errors and softer warnings separately.
- Reuse the same rule names and wording as the backend where practical.
- Validate as the author edits, before save.
- Preserve `EntitySelectButton` for all entity ID references.

Add a preview panel that shows:

- Source summary.
- Prerequisites.
- Objective checklist in player-facing order.
- Rewards.
- Dialog flow states: offer, in progress, ready to turn in.

The preview is not an execution sandbox. It is a static authoring preview that helps creators see how a quest will read to a player and whether the NPC source/turn-in route is coherent.

## API And Data Compatibility

The quest data model remains compatible. Validation makes previously accepted invalid definitions fail on create/update. Existing stored invalid quests may continue to load, but editing and saving them will require fixing validation errors.

No database schema migration is required.

## Error Handling

Validation errors returned by the backend should be deterministic and human-readable. Where possible, include the field path, for example `objectives[1].targetId`.

Runtime quest events should fail quietly for non-matching objectives and log unexpected service errors. Player-facing messages should only be sent for meaningful quest state changes.

## Testing

Backend tests:

- Quest validation rejects missing source IDs, missing objective targets, duplicate objective IDs, invalid objective types, negative rewards, and self-prerequisites.
- Quest validation accepts a complete NPC-sourced kill quest.
- Accepting a quest initializes objective progress with required counts.
- Collect progress uses stack quantity on pickup.
- Deliver progress requires matching inventory items.
- Deliver progress consumes delivered items.
- Talking to the wrong NPC does not progress a deliver objective.
- Completing a quest grants XP, gold, and reward items once.
- Quest log entries include objective descriptions and reward data.

Frontend verification:

- Run the Creator app build in `public/app`.
- Run the MUD client build in `public/mud-client`.
- Verify the quest validation panel renders without console errors.
- Verify the quest preview handles empty, invalid, and valid quest definitions.
- Verify quest notifications render accepted, progress, ready-to-turn-in, and completed states without overflowing the notification card.

Manual smoke test:

1. Create or use an NPC-sourced quest with collect and deliver objectives.
2. Talk to the source NPC and accept the quest.
3. Pick up the required item and observe progress notification.
4. Talk to the wrong NPC and confirm no delivery progress.
5. Talk to the delivery NPC and confirm delivery progress and item consumption.
6. Turn in the quest and confirm rewards, inventory, character, and quest log update.

## Documentation Updates

Update:

- `PROJECT.md` for user-facing quest delivery, validation, and Creator improvements.
- `ARCHITECTURE.md` for validation flow, event hooks, and quest turn-in data flow.
- `FEATURES.md` for deliver objective behavior, notifications, Creator validation/preview, and quest log semantics.
- `docs/design/QUEST_AUTHORING.md` for validation rules and deliver item consumption behavior.

## Risks

- Existing invalid quest data may become impossible to resave until corrected. This is acceptable because the validation prevents definitions that cannot work.
- Consuming delivered items changes documented behavior. The current authoring guide says deliver does not remove items; this design intentionally changes that to match player expectations.
- QuestTracker currently emits `questCompleted` when objectives are complete. The implementation must avoid treating objective completion as quest completion in the UI.

## Acceptance Criteria

- NPC-sourced quests can be offered, accepted, progressed, marked ready for turn-in, and completed through NPC dialog.
- Deliver objectives require and consume the requested item.
- Kill, pickup, dialog, and room-entry hooks update only matching active objectives and produce clear progress messages.
- Invalid quest definitions are rejected server-side with actionable validation errors.
- Creator shows validation and preview information before save.
- Player quest log and notifications make objective progress and turn-in readiness clear.
- Project documentation reflects the new quest delivery behavior.
- Backend tests and frontend build verification pass.
