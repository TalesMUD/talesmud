# Quest Delivery + Quest UX Design

Date: 2026-06-18

## Goal

Turn existing quest definitions, progress storage, and Creator support into a cohesive player-facing quest flow. The default completion model is NPC turn-in: objective completion makes a quest ready to turn in, while completion and rewards happen only through a valid turn-in action.

## Current State

The repo already has quest definitions, per-character progress, HTTP progress endpoints, MUD quest commands, NPC dialog injection, quest event hooks, and player-client quest notifications.

Important existing integration points:

- `pkg/service/quests.go` owns quest CRUD and basic progress operations.
- `pkg/mudserver/game/quest_tracker.go` handles kill, item pickup, room entry, dialog node, and talk-to-NPC events.
- `pkg/mudserver/game/commands/talk.go` discovers quest options from NPCs and injects them into dialogs.
- `pkg/mudserver/game/commands/dialog_select.go` handles quest accept, progress, and complete dialog choices.
- `pkg/server/handler/quests.go` exposes quest CRUD and progress APIs.
- `public/mud-client/src/game/Client.js`, `QuestLog.svelte`, `QuestLogWidget.svelte`, and `QuestNotifications.svelte` render player quest state.
- `public/app/src/creator/QuestsEditor.svelte` edits quest definitions.

The current flow is functional but split across services, commands, and client refreshes. Completion and rewards are duplicated in dialog handling and HTTP handling. Objective completion is sometimes sent as `questCompleted` even though rewards have not been granted. Deliver objectives currently advance when talking to an NPC without authoritative item validation.

## Design Principles

- The quest service is the authority for quest legality, availability, progress, turn-in, rewards, and validation.
- MUD commands render choices and dispatch player actions; they do not duplicate quest rules.
- Event hooks remain simple, but they pass enough event context for the quest service to decide whether progress is valid.
- Quest messages use precise states: accepted, progress, ready to turn in, completed.
- Creator validation matches server validation so builders see problems before save, but the server remains authoritative.
- Existing entity selector patterns in Creator must be preserved.

## Backend Design

### Quest Validation

Add server-side validation on quest create and update. Validation returns a structured list of issues with severity, field path, code, and message. Create/update rejects errors and may allow warnings.

Validation rules:

- Quest must have a name, description, source type, and at least one objective.
- Source type must be one of `npc`, `item`, `auto`, or `script`.
- NPC source requires an existing NPC ID.
- Item source requires an existing item template ID.
- Script-related fields require existing script IDs.
- Under the NPC-turn-in model, every quest must have a turn-in NPC path: either an NPC source or at least one deliver objective with a delivery NPC. Non-NPC-source quests without a delivery NPC are invalid because the player would have no place to turn them in.
- Objective IDs must be present and unique within the quest.
- Objective amount defaults to one at runtime, but Creator/server validation flags values below one for objective types that count progress.
- Kill objectives require an existing NPC target.
- Collect objectives require an existing item template target.
- Deliver objectives require an existing item template target and existing delivery NPC.
- Visit objectives require an existing room target.
- Talk objectives require an existing NPC target. Dialog node IDs are optional until dialog-node-level selection exists, but if provided they must be validated where the dialog model can support it.
- Custom objectives require an existing check script.
- Reward item template IDs must exist.
- Required quest IDs must exist, must not include the quest itself, and must not create prerequisite cycles.

### Quest Service Contract

Extend `QuestsService` with rule-oriented methods:

- `ValidateQuest(quest *quests.Quest) []QuestValidationIssue`
- `GetQuestDialogOptions(characterID, npcTemplateID, npcInstanceID string) ([]QuestDialogOption, error)`
- `ApplyQuestEvent(event QuestEvent) ([]QuestEventResult, error)`
- `CanTurnInQuest(characterID, questID, npcID string) (bool, string, error)`
- `TurnInQuest(characterID, questID, npcID string) (*QuestTurnInResult, error)`
- `BuildQuestLog(characterID string) ([]QuestLogEntry, error)`

`BuildQuestLog` becomes the single enrichment path for HTTP and WebSocket quest-log updates.

### Event Hooks

Keep existing hook points, but route matching through `ApplyQuestEvent`:

- NPC kill: matches kill objectives by NPC template ID or unique NPC ID.
- Item pickup: matches collect objectives by item template ID or unique item ID.
- Room entry: matches visit objectives by room ID and does not repeatedly increment an already completed visit objective.
- Dialog node: matches talk objectives by NPC and optional node ID.
- Talk to NPC: validates deliver objectives, including required item ownership and delivery NPC match.

Event results include quest ID, quest name, objective progress, and a result kind: `progress`, `objectiveCompleted`, or `readyToTurnIn`.

### Turn-In Rules

By default, quests complete only through turn-in.

A quest is ready to turn in when:

- Progress exists for the character.
- Progress status is active.
- All objectives are completed.
- The player is interacting with the source NPC or a delivery NPC designated by one of the quest's deliver objectives.

Turn-in performs these operations atomically at service level as far as current repositories allow:

- Re-check active status and objective completion.
- Re-check NPC is valid for turn-in.
- For deliver objectives, verify required items are still present and consume the delivered quantity from inventory by matching template ID first, then item ID.
- Mark progress completed.
- Grant XP, gold, and reward item instances.
- Update character stats and inventory.
- Return a structured result for MUD/API responses.

`questCompleted` messages are emitted only after turn-in and reward grant. When objectives are complete before turn-in, emit `questReadyToTurnIn`.

### Dialog Delivery

Move quest option construction out of `talk.go` into `QuestsService.GetQuestDialogOptions`.

Dialog options are:

- `[Quest] <name>` for available NPC-source quests.
- `[In Progress] <name>` for active quests with incomplete objectives.
- `[Turn In] <name>` for active quests ready to turn in at this NPC.

NPC response text uses quest dialog fields when present. If blank, the service returns default text:

- Offer: "I have work for you: <quest name>."
- Progress: "You still have work to do on <quest name>."
- Turn-in: "You have done well. Here is your reward."

### API

Add validation and preview endpoints for Creator:

- `POST /api/quests/validate` validates an unsaved quest payload.
- `POST /api/quests/preview` returns dialog option preview, objective summary, prerequisite summary, and reward summary for an unsaved quest payload.

Existing create/update handlers call `ValidateQuest` before storing. Existing quest progress endpoints keep their current paths but use the shared service turn-in and quest-log methods.

## Frontend Design

### Player Quest Log

Quest log entries include:

- quest ID and name
- status
- ready-to-turn-in boolean
- category and level
- description
- objective descriptions and counts
- rewards
- accepted and completed timestamps

The log visually separates:

- Active
- Ready to Turn In
- Completed
- Abandoned or Failed

Pinned quest display should prioritize ready-to-turn-in quests and show objective counts without requiring expansion.

### Quest Notifications

Notifications use stable types:

- `accepted`: quest accepted.
- `progress`: objective progress changed.
- `ready`: all objectives complete; return to the turn-in NPC.
- `completed`: quest turned in and rewards granted.

Client deduplication should avoid stacking repeated progress notifications for the same already-completed objective.

### Creator Validation and Preview

`QuestsEditor.svelte` adds a validation and preview panel without replacing the existing CRUD editor.

Validation panel:

- Runs local validation reactively for obvious missing fields.
- Calls server validation before save and when the builder clicks "Validate".
- Shows errors and warnings grouped by quest section and field path.

Preview panel:

- Shows how the NPC option labels will appear.
- Shows offer, progress, and turn-in NPC text with defaults applied.
- Shows objective progress labels as players will see them.
- Shows rewards and prerequisites.
- Shows whether the quest can be offered by the chosen NPC.

The editor continues using `EntitySelectButton` for entity references.

## Data Flow

1. Creator saves quest.
2. Server validates quest definition and stores only valid definitions.
3. Player talks to NPC.
4. Talk command asks quest service for dialog options.
5. Player accepts quest through dialog.
6. Game events call quest tracker.
7. Quest tracker sends event context to quest service.
8. Quest service updates progress and returns event results.
9. Server sends precise quest notification and refreshed quest log.
10. Player returns to valid turn-in NPC.
11. Dialog shows `[Turn In]`.
12. Turn-in uses quest service to complete quest and grant rewards.
13. Server sends reward, character, inventory, and quest-log updates.

## Error Handling

- Invalid quest definitions are rejected on create/update with validation errors.
- Runtime references that disappear after a quest is accepted are handled gracefully: event progress skips missing quest definitions, quest logs omit missing definitions, and turn-in fails with a clear message.
- Reward grant failures leave the quest active unless the service can prove rewards were granted.
- Commands show player-safe errors; logs include quest ID, character ID, and failing operation.

## Testing Strategy

Backend tests:

- Validation rejects missing/invalid NPC, item, room, script, duplicate objective IDs, self-prerequisites, and cycles.
- NPC-source quest options reflect unavailable, available, active, ready, completed, abandoned, and repeatable states.
- Kill, collect, visit, talk, and deliver events update only matching active objectives.
- Deliver events require item ownership and matching NPC.
- Ready-to-turn-in emits ready state without completing or granting rewards.
- Turn-in grants rewards once and marks progress completed.
- Quest-log enrichment includes descriptions, rewards, status, and ready flag.

Frontend checks:

- Player client handles accepted/progress/ready/completed notification types.
- Quest log renders ready-to-turn-in entries distinctly.
- Creator validation panel renders server validation issues.
- Creator preview panel renders default and custom dialog text.

Verification commands:

- `go test ./...`
- `npm run build` in `public/mud-client`
- `npm run build` in `public/app`

## Non-Goals

- A new general-purpose event bus.
- Auto-complete quest definitions.
- Full dialog-node picker redesign.
- Full inventory item consumption overhaul beyond what deliver objectives need.
- New database tables unless validation or turn-in cannot be implemented with current repositories.
