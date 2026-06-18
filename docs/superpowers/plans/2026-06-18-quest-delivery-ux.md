# Quest Delivery + Quest UX Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make quests flow cleanly from NPC offer through objective progress, ready-to-turn-in state, reward turn-in, player notifications, and Creator validation/preview.

**Architecture:** `pkg/service/quests.go` becomes the quest-rule authority for validation, dialog options, event progress, turn-in, reward grant, and quest-log enrichment. MUD commands, quest tracker, HTTP handlers, and clients call that service instead of duplicating rules. Frontend changes consume the richer quest-log/message contract and add Creator validation/preview around the existing editor.

**Tech Stack:** Go 1.24 services/handlers, existing SQLite repositories, Svelte 3 MUD client, Svelte 4 Creator app, Rollup/Vite builds.

---

## File Structure

- Modify `pkg/service/quests.go`: add validation issue/result types, quest dialog option types, quest event types, enriched quest-log type aliases, validation, event matching, turn-in, and log enrichment.
- Create `pkg/service/quests_test.go`: service-level tests for validation, event progress, deliver validation, ready state, and turn-in.
- Modify `pkg/mudserver/game/messages/messagetypes.go`: add `questReadyToTurnIn`.
- Modify `pkg/mudserver/game/messages/responses.go`: add ready flag and optional rewards/items to quest update/log messages.
- Modify `pkg/mudserver/game/quest_tracker.go`: delegate objective matching to `QuestsService.ApplyQuestEvent`.
- Modify `pkg/mudserver/game/commands/talk.go`: get quest dialog options from service.
- Modify `pkg/mudserver/game/commands/dialog_select.go`: call service accept/turn-in/log helpers and remove duplicated reward rules.
- Modify `pkg/mudserver/game/commands/select_character.go`: use service quest-log enrichment.
- Modify `pkg/server/handler/quests.go`: use validation on create/update, add validate/preview endpoints, use shared quest-log and turn-in methods.
- Modify `pkg/server/server.go`: register Creator validation/preview routes.
- Modify `public/mud-client/src/game/Client.js`: handle ready-to-turn-in notification and richer quest-log entries.
- Modify `public/mud-client/src/game/MUDXPlusStore.js`: deduplicate quest notifications by ID/type.
- Modify `public/mud-client/src/game/ui/QuestLog.svelte` and `public/mud-client/src/game/widgets/QuestLogWidget.svelte`: show ready-to-turn-in section/state.
- Modify `public/mud-client/src/game/ui/QuestNotifications.svelte`: style `ready` notifications.
- Modify `public/app/src/api/quests.js`: add `validateQuest` and `previewQuest`.
- Modify `public/app/src/creator/QuestsEditor.svelte`: add validation/preview panel using existing `EntitySelectButton` patterns.
- Update `PROJECT.md`, `ARCHITECTURE.md`, and `FEATURES.md` after behavior changes.

---

### Task 1: Add Quest Validation Contract

**Files:**
- Modify: `pkg/service/quests.go`
- Test: `pkg/service/quests_test.go`

- [ ] **Step 1: Write failing validation tests**

Add tests that build in-memory repositories or thin fakes for quests, progress, NPCs, items, rooms, and scripts. Test names:

```go
func TestValidateQuestRejectsMissingSourceNPC(t *testing.T) {}
func TestValidateQuestRejectsDuplicateObjectiveIDs(t *testing.T) {}
func TestValidateQuestRejectsSelfPrerequisite(t *testing.T) {}
func TestValidateQuestRejectsNonNPCSourceWithoutTurnInNPC(t *testing.T) {}
func TestValidateQuestAcceptsValidNPCQuest(t *testing.T) {}
```

The invalid tests should assert at least one issue with `Severity == "error"` and a stable code such as `missing_source_npc`, `duplicate_objective_id`, `self_prerequisite`, and `missing_turn_in_npc`.

- [ ] **Step 2: Run validation tests and verify RED**

Run: `go test ./pkg/service -run 'TestValidateQuest'`

Expected: FAIL because `ValidateQuest` and `QuestValidationIssue` are not defined.

- [ ] **Step 3: Add validation types and interface method**

Add to `pkg/service/quests.go`:

```go
type QuestValidationIssue struct {
	Severity string `json:"severity"`
	Path     string `json:"path"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}
```

Extend `QuestsService`:

```go
ValidateQuest(quest *quests.Quest) []QuestValidationIssue
```

- [ ] **Step 4: Implement minimal validation**

Implement helper methods in `questsService`:

```go
func (s *questsService) ValidateQuest(quest *quests.Quest) []QuestValidationIssue {
	var issues []QuestValidationIssue
	addError := func(path, code, msg string) {
		issues = append(issues, QuestValidationIssue{Severity: "error", Path: path, Code: code, Message: msg})
	}
	// validate name, description, source, objective IDs, prerequisite self-reference,
	// and NPC turn-in path using s.facade when entity lookups are needed.
	return issues
}
```

Validation must use `s.facade.NPCsService().FindByID`, `ItemsService().FindByID`, `RoomsService().FindByID`, and `ScriptsService().FindByID` when `s.facade` is available.

- [ ] **Step 5: Run validation tests and verify GREEN**

Run: `go test ./pkg/service -run 'TestValidateQuest'`

Expected: PASS.

---

### Task 2: Centralize Quest Log Enrichment and Ready State

**Files:**
- Modify: `pkg/service/quests.go`
- Modify: `pkg/mudserver/game/messages/responses.go`
- Test: `pkg/service/quests_test.go`

- [ ] **Step 1: Write failing quest-log tests**

Add:

```go
func TestBuildQuestLogIncludesObjectiveDescriptionsAndReadyFlag(t *testing.T) {}
```

Arrange an active quest with one completed objective and assert returned entry has:

```go
entry.QuestName == "Find the Relic"
entry.Objectives[0].Description == "Bring back the relic"
entry.ReadyToTurnIn == true
```

- [ ] **Step 2: Run quest-log test and verify RED**

Run: `go test ./pkg/service -run TestBuildQuestLogIncludesObjectiveDescriptionsAndReadyFlag`

Expected: FAIL because `BuildQuestLog` and service log-entry types are not defined.

- [ ] **Step 3: Add service quest-log types**

Add to `pkg/service/quests.go`:

```go
type QuestObjectiveProgressEntry struct {
	ObjectiveID string `json:"objectiveId"`
	Description string `json:"description"`
	Current     int32  `json:"current"`
	Required    int32  `json:"required"`
	Completed   bool   `json:"completed"`
}

type QuestRewardEntry struct {
	XP              int32    `json:"xp,omitempty"`
	Gold            int64    `json:"gold,omitempty"`
	ItemTemplateIDs []string `json:"itemTemplateIds,omitempty"`
}

type QuestLogEntry struct {
	QuestID       string                       `json:"questId"`
	QuestName     string                       `json:"questName"`
	Description   string                       `json:"description,omitempty"`
	Category      string                       `json:"category,omitempty"`
	Level         int32                        `json:"level,omitempty"`
	Status        string                       `json:"status"`
	ReadyToTurnIn bool                         `json:"readyToTurnIn"`
	Objectives    []QuestObjectiveProgressEntry `json:"objectives"`
	Rewards       *QuestRewardEntry            `json:"rewards,omitempty"`
	AcceptedAt    string                       `json:"acceptedAt,omitempty"`
	CompletedAt   string                       `json:"completedAt,omitempty"`
}
```

Extend `QuestsService`:

```go
BuildQuestLog(characterID string) ([]QuestLogEntry, error)
```

- [ ] **Step 4: Implement `BuildQuestLog`**

Use existing `GetQuestLog`, `FindByID`, and objective definitions. `ReadyToTurnIn` is true when status is active and all objective progress rows are completed.

- [ ] **Step 5: Add ready field to WebSocket message entries**

Modify `pkg/mudserver/game/messages/responses.go`:

```go
ReadyToTurnIn bool `json:"readyToTurnIn,omitempty"`
```

on `QuestLogEntry`.

- [ ] **Step 6: Run quest-log tests and verify GREEN**

Run: `go test ./pkg/service -run TestBuildQuestLogIncludesObjectiveDescriptionsAndReadyFlag`

Expected: PASS.

---

### Task 3: Add Event Progress and Turn-In Service Rules

**Files:**
- Modify: `pkg/service/quests.go`
- Modify: `pkg/mudserver/game/messages/messagetypes.go`
- Test: `pkg/service/quests_test.go`

- [ ] **Step 1: Write failing event and turn-in tests**

Add tests:

```go
func TestApplyQuestEventKillUpdatesMatchingActiveObjective(t *testing.T) {}
func TestApplyQuestEventVisitDoesNotIncrementCompletedObjective(t *testing.T) {}
func TestApplyQuestEventDeliverRequiresMatchingNPCAndItem(t *testing.T) {}
func TestTurnInQuestGrantsRewardsAndCompletesOnce(t *testing.T) {}
func TestTurnInQuestRejectsWrongNPC(t *testing.T) {}
```

- [ ] **Step 2: Run tests and verify RED**

Run: `go test ./pkg/service -run 'TestApplyQuestEvent|TestTurnInQuest'`

Expected: FAIL because event and turn-in types/methods do not exist.

- [ ] **Step 3: Add event and result types**

Add to `pkg/service/quests.go`:

```go
type QuestEventType string

const (
	QuestEventNPCKilled  QuestEventType = "npc_killed"
	QuestEventItemPickup QuestEventType = "item_pickup"
	QuestEventRoomEnter  QuestEventType = "room_enter"
	QuestEventDialogNode QuestEventType = "dialog_node"
	QuestEventTalkToNPC  QuestEventType = "talk_to_npc"
)

type QuestEvent struct {
	Type        QuestEventType
	CharacterID string
	UserID      string
	NPCID       string
	NPCTemplateID string
	ItemID      string
	ItemTemplateID string
	RoomID      string
	DialogID    string
	DialogNodeID string
}

type QuestEventResultKind string

const (
	QuestEventResultProgress           QuestEventResultKind = "progress"
	QuestEventResultObjectiveCompleted QuestEventResultKind = "objectiveCompleted"
	QuestEventResultReadyToTurnIn      QuestEventResultKind = "readyToTurnIn"
)

type QuestEventResult struct {
	Kind       QuestEventResultKind
	QuestID    string
	QuestName  string
	Objectives []QuestObjectiveProgressEntry
}

type QuestTurnInResult struct {
	QuestID      string
	QuestName    string
	Progress     *quests.QuestProgress
	Rewards      QuestRewardEntry
	GrantedItems []string
}
```

Extend `QuestsService`:

```go
ApplyQuestEvent(event QuestEvent) ([]QuestEventResult, error)
CanTurnInQuest(characterID, questID, npcID string) (bool, string, error)
TurnInQuest(characterID, questID, npcID string) (*QuestTurnInResult, error)
```

- [ ] **Step 4: Implement event matching**

Move matching logic from `quest_tracker.go` into `ApplyQuestEvent`. For deliver objectives, find character inventory, match item by `TemplateID == obj.TargetID || ID == obj.TargetID`, and match NPC by `obj.DeliverToNPCID == event.NPCTemplateID || obj.DeliverToNPCID == event.NPCID`.

- [ ] **Step 5: Implement turn-in**

`TurnInQuest` must:

1. Call `CanTurnInQuest`.
2. Remove deliver items from inventory by exact item ID after locating by template or ID.
3. Call `CompleteQuest`.
4. Call `GrantQuestRewards`.
5. Reload progress and return reward data.

- [ ] **Step 6: Add ready message type**

In `pkg/mudserver/game/messages/messagetypes.go` add:

```go
MessageTypeQuestReadyToTurnIn = "questReadyToTurnIn"
```

- [ ] **Step 7: Run event and turn-in tests and verify GREEN**

Run: `go test ./pkg/service -run 'TestApplyQuestEvent|TestTurnInQuest'`

Expected: PASS.

---

### Task 4: Wire Backend Commands, Tracker, and API to the Service

**Files:**
- Modify: `pkg/mudserver/game/quest_tracker.go`
- Modify: `pkg/mudserver/game/commands/talk.go`
- Modify: `pkg/mudserver/game/commands/dialog_select.go`
- Modify: `pkg/mudserver/game/commands/select_character.go`
- Modify: `pkg/server/handler/quests.go`
- Modify: `pkg/server/server.go`

- [ ] **Step 1: Wire `quest_tracker.go` to `ApplyQuestEvent`**

Each existing hook constructs a `service.QuestEvent` and calls `ApplyQuestEvent`. Convert results to `messages.NewQuestUpdateMessage`. Use `questReadyToTurnIn` when result kind is ready.

- [ ] **Step 2: Run Go tests**

Run: `go test ./pkg/mudserver/game ./pkg/service`

Expected: PASS.

- [ ] **Step 3: Move quest dialog options to service**

Add `QuestDialogOption` to `pkg/service/quests.go`:

```go
type QuestDialogOption struct {
	Text      string
	NPCText   string
	QuestID   string
	QuestName string
	Action    string
}
```

Replace local option lookup in `talk.go` with `game.GetFacade().QuestsService().GetQuestDialogOptions(...)`.

- [ ] **Step 4: Replace dialog turn-in logic**

In `dialog_select.go`, replace direct `CompleteQuest` plus `GrantQuestRewards` calls with:

```go
result, err := game.GetFacade().QuestsService().TurnInQuest(char.ID, questID, npcTemplateOrInstanceID)
```

Then send reward, character update, inventory update, and quest log messages from `result`.

- [ ] **Step 5: Use shared quest-log enrichment**

Replace `convertQuestLogToEntries` and `sendQuestLogToPlayer` custom enrichment with a converter from `service.QuestLogEntry` to `messages.QuestLogEntry`, or make message structs align and map fields directly.

- [ ] **Step 6: Add HTTP validation and preview endpoints**

In `pkg/server/handler/quests.go`, add:

```go
func (h *QuestsHandler) ValidateQuest(c *gin.Context) {}
func (h *QuestsHandler) PreviewQuest(c *gin.Context) {}
```

`ValidateQuest` binds a quest payload and returns `{ "issues": issues, "valid": len(errors) == 0 }`.

`PreviewQuest` binds a quest payload and returns labels/text/objectives/rewards using the same defaults as the service.

- [ ] **Step 7: Register routes**

In `pkg/server/server.go`, under creator quest routes, add:

```go
creator.POST("quests/validate", questsHandler.ValidateQuest)
creator.POST("quests/preview", questsHandler.PreviewQuest)
```

- [ ] **Step 8: Run backend tests**

Run: `go test ./...`

Expected: PASS.

---

### Task 5: Improve Player Quest UX

**Files:**
- Modify: `public/mud-client/src/game/Client.js`
- Modify: `public/mud-client/src/game/MUDXPlusStore.js`
- Modify: `public/mud-client/src/game/ui/QuestNotifications.svelte`
- Modify: `public/mud-client/src/game/ui/QuestLog.svelte`
- Modify: `public/mud-client/src/game/widgets/QuestLogWidget.svelte`

- [ ] **Step 1: Add ready notification handling**

In `Client.js`, add handler:

```js
messageHandlers["questReadyToTurnIn"] = (msg) => {
  renderer(msg.message);
  requestQuestLog();
  if (mux) {
    mux.addQuestNotification({
      id: `quest-ready-${msg.questId || Date.now()}`,
      questId: msg.questId,
      type: 'ready',
      questName: msg.questName || 'Quest',
      message: msg.message || 'Ready to turn in',
    });
  }
};
```

- [ ] **Step 2: Deduplicate quest notifications**

In `MUDXPlusStore.js`, update `addQuestNotification` to replace existing notifications with the same `id` before appending.

- [ ] **Step 3: Style ready notification**

In `QuestNotifications.svelte`, add `.notification.ready` border/title color styles.

- [ ] **Step 4: Add ready quest grouping**

In `QuestLog.svelte` and `QuestLogWidget.svelte`, derive:

```js
$: readyQuests = sortedQuests.filter((q) => q.status === 'active' && q.readyToTurnIn);
$: unpinnedActiveQuests = sortedQuests.filter((q) => !q.readyToTurnIn && !pinnedQuests.includes(q.questId) && q.status === 'active');
```

Render a `Ready to Turn In` section before regular active quests.

- [ ] **Step 5: Build MUD client**

Run from `public/mud-client`: `npm run build`

Expected: build exits 0.

---

### Task 6: Add Creator Validation and Preview

**Files:**
- Modify: `public/app/src/api/quests.js`
- Modify: `public/app/src/creator/QuestsEditor.svelte`

- [ ] **Step 1: Add API helpers**

In `quests.js`, add:

```js
function validateQuest(token, quest, cb, errorCb) {
  axios.post(`${backend}/quests/validate`, quest, { headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` } })
    .then((r) => cb(r.data))
    .catch((err) => errorCb(err));
}

function previewQuest(token, quest, cb, errorCb) {
  axios.post(`${backend}/quests/preview`, quest, { headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` } })
    .then((r) => cb(r.data))
    .catch((err) => errorCb(err));
}
```

Export both helpers.

- [ ] **Step 2: Add editor state**

In `QuestsEditor.svelte`, import the helpers and add:

```js
let validationIssues = [];
let preview = null;
let validationMessage = "";
```

- [ ] **Step 3: Add validate and preview actions**

Add functions:

```js
function runValidation() {
  if (!$authToken || !$store.selectedElement) return;
  validateQuest($authToken, $store.selectedElement, (data) => {
    validationIssues = data.issues || [];
    validationMessage = data.valid ? "Quest definition is valid." : "Quest definition has errors.";
  }, () => {
    validationMessage = "Validation failed.";
  });
}

function runPreview() {
  if (!$authToken || !$store.selectedElement) return;
  previewQuest($authToken, $store.selectedElement, (data) => {
    preview = data;
  }, () => {
    preview = { error: "Preview failed." };
  });
}
```

- [ ] **Step 4: Add validation/preview panel**

Add a Creator panel near the top of the content slot with `Validate` and `Preview` buttons. Render issues as field path, severity, and message. Render preview labels/text/objective/reward summaries.

- [ ] **Step 5: Build Creator app**

Run from `public/app`: `npm run build`

Expected: build exits 0.

---

### Task 7: Documentation and Final Verification

**Files:**
- Modify: `PROJECT.md`
- Modify: `ARCHITECTURE.md`
- Modify: `FEATURES.md`

- [ ] **Step 1: Update documentation**

Document:

- NPC turn-in as default completion model.
- Quest service owns validation, event progress, turn-in, rewards, and log enrichment.
- `questReadyToTurnIn` message type.
- Creator validation and preview endpoints.
- Player ready-to-turn-in UI behavior.

- [ ] **Step 2: Run full backend verification**

Run: `go test ./...`

Expected: PASS.

- [ ] **Step 3: Run MUD client build**

Run from `public/mud-client`: `npm run build`

Expected: build exits 0.

- [ ] **Step 4: Run Creator build**

Run from `public/app`: `npm run build`

Expected: build exits 0.

- [ ] **Step 5: Inspect final diff**

Run: `git diff --stat` and `git diff --check`.

Expected: no whitespace errors and changes only in planned files.
