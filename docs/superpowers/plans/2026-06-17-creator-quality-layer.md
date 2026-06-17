# Creator Quality Layer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Creator quality layer that blocks broken references on save, exposes world health diagnostics, shows inline validation warnings, adds safe preview/test tools, and verifies scalable entity selectors.

**Architecture:** Add a focused backend validation package that produces structured issues from a live world snapshot, expose diagnostics/validation/preview endpoints through Creator-protected routes, and call the same validation before CRUD saves. The Svelte Creator consumes those APIs through shared validation and preview helpers, then renders a common validation panel, health tab, row indicators, and editor-specific preview modals.

**Tech Stack:** Go 1.24, Gin, existing service/repository facade, Svelte 4, Vite, Axios, existing Creator `CRUDEditor`, `DataTable`, `EntitySelectButton`, and modal patterns.

---

## File Structure

Create:

- `pkg/service/validation/types.go`: `Severity`, `Issue`, `Result`, helpers for adding/counting issues.
- `pkg/service/validation/snapshot.go`: `WorldSnapshot`, map builders, `BuildSnapshot(service.Facade)`.
- `pkg/service/validation/scripts.go`: reusable Lua static checks moved from importer.
- `pkg/service/validation/validator.go`: entity and full-world validation entry points.
- `pkg/service/validation/validator_test.go`: unit tests for broken references and warning/error counts.
- `pkg/server/handler/validation.go`: diagnostics, draft validation, and preview handlers.
- `pkg/server/handler/validation_test.go`: handler coverage for validation/save behavior.
- `public/app/src/api/validation.js`: diagnostics and draft validation API calls.
- `public/app/src/api/previews.js`: preview API calls.
- `public/app/src/creator/ValidationPanel.svelte`: shared inline issue renderer.
- `public/app/src/creator/WorldHealth.svelte`: Creator health page.
- `public/app/src/creator/DialogPreviewModal.svelte`: safe dialog preview.
- `public/app/src/creator/QuestPreviewModal.svelte`: quest preview.
- `public/app/src/creator/RoomPreviewModal.svelte`: room preview.
- `public/app/src/creator/MerchantPreviewModal.svelte`: merchant stock preview.

Modify:

- `pkg/importer/validator.go`: replace local Lua static helpers with calls to `pkg/service/validation`.
- `pkg/server/server.go`: instantiate validation handler and register Creator routes.
- `pkg/server/handler/rooms.go`: validate create/update before save.
- `pkg/server/handler/items.go`: validate create/update before save.
- `pkg/server/handler/npcs.go`: validate create/update before save.
- `pkg/server/handler/npcspawners.go`: validate create/update before save.
- `pkg/server/handler/dialogs.go`: validate create/update before save.
- `pkg/server/handler/loottables.go`: validate create/update before save.
- `pkg/server/handler/quests.go`: validate create/update before save.
- `pkg/server/handler/scripts.go`: validate create/update before save.
- `public/app/src/AppContent.svelte`: add `/creator/health` route.
- `public/app/src/creator/Creator.svelte`: add Health tab.
- `public/app/src/creator/CRUDEditor.svelte`: common inline validation, blocked save display, row issue indicators, preview action hook.
- `public/app/src/creator/RoomsEditor.svelte`: set `entityType: "room"` and wire room preview.
- `public/app/src/creator/ItemTemplatesEditor.svelte`: set `entityType: "item"`.
- `public/app/src/creator/NPCsEditor.svelte`: set `entityType: "npc"` and wire merchant preview.
- `public/app/src/creator/DialogsEditor.svelte`: set `entityType: "dialog"` and wire dialog preview.
- `public/app/src/creator/QuestsEditor.svelte`: set `entityType: "quest"` and wire quest preview.
- `public/app/src/creator/ScriptsEditor.svelte`: set `entityType: "script"` and show static validation near runtime test output.
- `PROJECT.md`, `ARCHITECTURE.md`, `FEATURES.md`: document behavior, architecture, and Creator authoring capabilities.

---

### Task 1: Shared Validation Types

**Files:**
- Create: `pkg/service/validation/types.go`
- Test: `pkg/service/validation/validator_test.go`

- [ ] **Step 1: Write the failing type/helper test**

Add this test to `pkg/service/validation/validator_test.go`:

```go
package validation

import "testing"

func TestResultAddCountsErrorsAndWarnings(t *testing.T) {
	var result Result

	result.Add(Issue{Severity: SeverityError, Code: "missing_room", EntityType: "room", EntityID: "room-a"})
	result.Add(Issue{Severity: SeverityWarning, Code: "empty_action", EntityType: "room", EntityID: "room-a"})

	if result.Valid {
		t.Fatalf("result.Valid = true, want false when errors exist")
	}
	if result.Errors != 1 {
		t.Fatalf("Errors = %d, want 1", result.Errors)
	}
	if result.Warnings != 1 {
		t.Fatalf("Warnings = %d, want 1", result.Warnings)
	}
	if len(result.Issues) != 2 {
		t.Fatalf("Issues length = %d, want 2", len(result.Issues))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./pkg/service/validation -run TestResultAddCountsErrorsAndWarnings -count=1
```

Expected: FAIL because package `pkg/service/validation` or type `Result` does not exist.

- [ ] **Step 3: Implement minimal validation types**

Create `pkg/service/validation/types.go`:

```go
package validation

type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

type Issue struct {
	Severity   Severity `json:"severity"`
	Code       string   `json:"code"`
	EntityType string   `json:"entityType"`
	EntityID   string   `json:"entityId,omitempty"`
	Field      string   `json:"field,omitempty"`
	RefType    string   `json:"refType,omitempty"`
	RefID      string   `json:"refId,omitempty"`
	Message    string   `json:"message"`
}

type Result struct {
	Valid    bool    `json:"valid"`
	Errors   int     `json:"errors"`
	Warnings int     `json:"warnings"`
	Issues   []Issue `json:"issues"`
}

func (r *Result) Add(issue Issue) {
	r.Issues = append(r.Issues, issue)
	switch issue.Severity {
	case SeverityError:
		r.Errors++
	case SeverityWarning:
		r.Warnings++
	}
	r.Valid = r.Errors == 0
}

func (r *Result) Merge(other Result) {
	for _, issue := range other.Issues {
		r.Add(issue)
	}
}

func NewResult() Result {
	return Result{Valid: true, Issues: []Issue{}}
}

func Error(code, entityType, entityID, field, refType, refID, message string) Issue {
	return Issue{Severity: SeverityError, Code: code, EntityType: entityType, EntityID: entityID, Field: field, RefType: refType, RefID: refID, Message: message}
}

func Warning(code, entityType, entityID, field, refType, refID, message string) Issue {
	return Issue{Severity: SeverityWarning, Code: code, EntityType: entityType, EntityID: entityID, Field: field, RefType: refType, RefID: refID, Message: message}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
go test ./pkg/service/validation -run TestResultAddCountsErrorsAndWarnings -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/service/validation/types.go pkg/service/validation/validator_test.go
git commit -m "feat: add structured creator validation issues"
```

---

### Task 2: World Snapshot Loader

**Files:**
- Create: `pkg/service/validation/snapshot.go`
- Modify: `pkg/service/facade.go`
- Test: `pkg/service/validation/validator_test.go`

- [ ] **Step 1: Write the failing snapshot test**

Append to `pkg/service/validation/validator_test.go`:

```go
func TestSnapshotHasRoomAndScriptLookupHelpers(t *testing.T) {
	snapshot := WorldSnapshot{
		RoomIDs:   map[string]bool{"room-a": true},
		ScriptIDs: map[string]bool{"script-a": true},
	}

	if !snapshot.HasRoom("room-a") {
		t.Fatalf("HasRoom(room-a) = false, want true")
	}
	if snapshot.HasRoom("missing-room") {
		t.Fatalf("HasRoom(missing-room) = true, want false")
	}
	if !snapshot.HasScript("script-a") {
		t.Fatalf("HasScript(script-a) = false, want true")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./pkg/service/validation -run TestSnapshotHasRoomAndScriptLookupHelpers -count=1
```

Expected: FAIL because `WorldSnapshot` is undefined.

- [ ] **Step 3: Implement snapshot maps and helpers**

Create `pkg/service/validation/snapshot.go`:

```go
package validation

import (
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/scripts"
	"github.com/talesmud/talesmud/pkg/service"
)

type WorldSnapshot struct {
	Rooms       map[string]*rooms.Room
	Items       map[string]*items.Item
	NPCs        map[string]*npc.NPC
	Dialogs     map[string]*dialogs.Dialog
	LootTables  map[string]*items.LootTable
	Spawners    map[string]*npc.NPCSpawner
	Quests      map[string]*quests.Quest
	Scripts     map[string]*scripts.Script
	RoomIDs     map[string]bool
	ItemIDs     map[string]bool
	NPCIDs      map[string]bool
	DialogIDs   map[string]bool
	LootTableIDs map[string]bool
	SpawnerIDs  map[string]bool
	QuestIDs    map[string]bool
	ScriptIDs   map[string]bool
}

func NewWorldSnapshot() WorldSnapshot {
	return WorldSnapshot{
		Rooms: map[string]*rooms.Room{}, Items: map[string]*items.Item{}, NPCs: map[string]*npc.NPC{},
		Dialogs: map[string]*dialogs.Dialog{}, LootTables: map[string]*items.LootTable{}, Spawners: map[string]*npc.NPCSpawner{},
		Quests: map[string]*quests.Quest{}, Scripts: map[string]*scripts.Script{},
		RoomIDs: map[string]bool{}, ItemIDs: map[string]bool{}, NPCIDs: map[string]bool{}, DialogIDs: map[string]bool{},
		LootTableIDs: map[string]bool{}, SpawnerIDs: map[string]bool{}, QuestIDs: map[string]bool{}, ScriptIDs: map[string]bool{},
	}
}

func (s WorldSnapshot) HasRoom(id string) bool { return id == "" || s.RoomIDs[id] }
func (s WorldSnapshot) HasItem(id string) bool { return id == "" || s.ItemIDs[id] }
func (s WorldSnapshot) HasNPC(id string) bool { return id == "" || s.NPCIDs[id] }
func (s WorldSnapshot) HasDialog(id string) bool { return id == "" || s.DialogIDs[id] }
func (s WorldSnapshot) HasLootTable(id string) bool { return id == "" || s.LootTableIDs[id] }
func (s WorldSnapshot) HasQuest(id string) bool { return id == "" || s.QuestIDs[id] }
func (s WorldSnapshot) HasScript(id string) bool { return id == "" || s.ScriptIDs[id] }

func BuildSnapshot(f service.Facade) (WorldSnapshot, error) {
	snapshot := NewWorldSnapshot()

	roomsList, err := f.RoomsService().FindAll()
	if err != nil { return snapshot, err }
	for _, room := range roomsList { snapshot.Rooms[room.ID] = room; snapshot.RoomIDs[room.ID] = true }

	itemsList, err := f.ItemsService().FindAll(repository.ItemsQuery{})
	if err != nil { return snapshot, err }
	for _, item := range itemsList { snapshot.Items[item.ID] = item; snapshot.ItemIDs[item.ID] = true }

	npcsList, err := f.NPCsService().FindAll()
	if err != nil { return snapshot, err }
	for _, n := range npcsList { snapshot.NPCs[n.ID] = n; snapshot.NPCIDs[n.ID] = true }

	dialogsList, err := f.DialogsService().FindAll()
	if err != nil { return snapshot, err }
	for _, dialog := range dialogsList { snapshot.Dialogs[dialog.ID] = dialog; snapshot.DialogIDs[dialog.ID] = true }

	lootTables, err := f.LootTablesService().FindAll()
	if err != nil { return snapshot, err }
	for _, table := range lootTables { snapshot.LootTables[table.ID] = table; snapshot.LootTableIDs[table.ID] = true }

	spawners, err := f.NPCSpawnersService().FindAll()
	if err != nil { return snapshot, err }
	for _, spawner := range spawners { snapshot.Spawners[spawner.ID] = spawner; snapshot.SpawnerIDs[spawner.ID] = true }

	questsList, err := f.QuestsService().FindAll()
	if err != nil { return snapshot, err }
	for _, quest := range questsList { snapshot.Quests[quest.ID] = quest; snapshot.QuestIDs[quest.ID] = true }

	scriptsList, err := f.ScriptsService().FindAll()
	if err != nil { return snapshot, err }
	for _, script := range scriptsList { snapshot.Scripts[script.ID] = script; snapshot.ScriptIDs[script.ID] = true }

	return snapshot, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
go test ./pkg/service/validation -run TestSnapshotHasRoomAndScriptLookupHelpers -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/service/validation/snapshot.go pkg/service/validation/validator_test.go
git commit -m "feat: load creator validation world snapshot"
```

---

### Task 3: Entity Validators

**Files:**
- Create: `pkg/service/validation/validator.go`
- Test: `pkg/service/validation/validator_test.go`

- [ ] **Step 1: Write failing validator tests**

Append tests:

```go
func TestValidateRoomReportsBrokenExitAndActionScript(t *testing.T) {
	exits := rooms.Exits{{Name: "north", Target: "missing-room"}}
	actions := rooms.Actions{{Name: "pull lever", Type: rooms.RoomActionTypeScript, ScriptId: "missing-script"}}
	room := &rooms.Room{Entity: &entities.Entity{ID: "room-a"}, Name: "Room A", Exits: &exits, Actions: &actions}

	result := ValidateRoom(room, NewWorldSnapshot())

	if result.Errors != 2 {
		t.Fatalf("Errors = %d, want 2: %#v", result.Errors, result.Issues)
	}
}

func TestValidateQuestReportsBrokenRewardAndSource(t *testing.T) {
	quest := &quests.Quest{
		Entity: &entities.Entity{ID: "quest-a"},
		Name: "Quest A",
		Source: quests.QuestSource{Type: "npc", NPCID: "missing-npc"},
		Rewards: quests.Reward{ItemTemplateIDs: []string{"missing-item"}},
	}

	result := ValidateQuest(quest, NewWorldSnapshot())

	if result.Errors != 2 {
		t.Fatalf("Errors = %d, want 2: %#v", result.Errors, result.Issues)
	}
}
```

Add imports:

```go
import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
)
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
go test ./pkg/service/validation -run 'TestValidate(Room|Quest)' -count=1
```

Expected: FAIL because `ValidateRoom` and `ValidateQuest` are undefined.

- [ ] **Step 3: Implement minimal validators and dispatcher**

Create `pkg/service/validation/validator.go` with functions for all entity types named in the design:

```go
package validation

import (
	"fmt"

	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/scripts"
)

func ValidateRoom(room *rooms.Room, snapshot WorldSnapshot) Result {
	result := NewResult()
	if room == nil { return result }
	if room.Exits != nil {
		for i, exit := range *room.Exits {
			if exit.Target != "" && !snapshot.HasRoom(exit.Target) {
				result.Add(Error("missing_room", "room", room.ID, fmt.Sprintf("exits[%d].target", i), "room", exit.Target, "Exit target references a missing room."))
			}
			if exit.Target != "" && exit.Name == "" {
				result.Add(Warning("exit_without_name", "room", room.ID, fmt.Sprintf("exits[%d].name", i), "", "", "Exit has a target but no name."))
			}
		}
	}
	if room.OnEnterScriptID != "" && !snapshot.HasScript(room.OnEnterScriptID) {
		result.Add(Error("missing_script", "room", room.ID, "onEnterScriptID", "script", room.OnEnterScriptID, "On-enter script references a missing script."))
	}
	if room.Actions != nil {
		for i, action := range *room.Actions {
			if action.Type == rooms.RoomActionTypeScript && action.ScriptId == "" {
				result.Add(Warning("script_action_without_script", "room", room.ID, fmt.Sprintf("actions[%d].scriptId", i), "script", "", "Script action has no script ID."))
			}
			if action.Type == rooms.RoomActionTypeScript && action.ScriptId != "" && !snapshot.HasScript(action.ScriptId) {
				result.Add(Error("missing_script", "room", room.ID, fmt.Sprintf("actions[%d].scriptId", i), "script", action.ScriptId, "Room action references a missing script."))
			}
		}
	}
	if room.Items != nil {
		for i, itemID := range *room.Items {
			if itemID != "" && !snapshot.HasItem(itemID) {
				result.Add(Error("missing_item", "room", room.ID, fmt.Sprintf("items[%d]", i), "item", itemID, "Room item references a missing item."))
			}
		}
	}
	if room.NPCs != nil {
		for i, npcID := range *room.NPCs {
			if npcID != "" && !snapshot.HasNPC(npcID) {
				result.Add(Error("missing_npc", "room", room.ID, fmt.Sprintf("npcs[%d]", i), "npc", npcID, "Room resident references a missing NPC."))
			}
		}
	}
	return result
}

func ValidateQuest(quest *quests.Quest, snapshot WorldSnapshot) Result {
	result := NewResult()
	if quest == nil { return result }
	if quest.Source.NPCID != "" && !snapshot.HasNPC(quest.Source.NPCID) {
		result.Add(Error("missing_npc", "quest", quest.ID, "source.npcId", "npc", quest.Source.NPCID, "Quest source NPC references a missing NPC."))
	}
	if quest.Source.ItemID != "" && !snapshot.HasItem(quest.Source.ItemID) {
		result.Add(Error("missing_item", "quest", quest.ID, "source.itemId", "item", quest.Source.ItemID, "Quest source item references a missing item."))
	}
	for i, rewardID := range quest.Rewards.ItemTemplateIDs {
		if rewardID != "" && !snapshot.HasItem(rewardID) {
			result.Add(Error("missing_item", "quest", quest.ID, fmt.Sprintf("rewards.itemTemplateIds[%d]", i), "item", rewardID, "Quest reward references a missing item template."))
		}
	}
	for i, reqID := range quest.RequiredQuestIDs {
		if reqID != "" && !snapshot.HasQuest(reqID) {
			result.Add(Error("missing_quest", "quest", quest.ID, fmt.Sprintf("requiredQuestIds[%d]", i), "quest", reqID, "Quest prerequisite references a missing quest."))
		}
	}
	if quest.OnCompleteScriptID != "" && !snapshot.HasScript(quest.OnCompleteScriptID) {
		result.Add(Error("missing_script", "quest", quest.ID, "onCompleteScriptId", "script", quest.OnCompleteScriptID, "Quest completion script references a missing script."))
	}
	for i, objective := range quest.Objectives {
		if objective.ID == "" {
			result.Add(Warning("objective_without_id", "quest", quest.ID, fmt.Sprintf("objectives[%d].id", i), "", "", "Quest objective has no stable ID."))
		}
		if objective.CheckScriptID != "" && !snapshot.HasScript(objective.CheckScriptID) {
			result.Add(Error("missing_script", "quest", quest.ID, fmt.Sprintf("objectives[%d].checkScriptId", i), "script", objective.CheckScriptID, "Quest objective check script references a missing script."))
		}
	}
	return result
}

func ValidateNPC(n *npc.NPC, snapshot WorldSnapshot) Result { result := NewResult(); return result }
func ValidateDialog(dialog *dialogs.Dialog, snapshot WorldSnapshot) Result { result := NewResult(); return result }
func ValidateLootTable(table *items.LootTable, snapshot WorldSnapshot) Result { result := NewResult(); return result }
func ValidateItem(item *items.Item, snapshot WorldSnapshot) Result { result := NewResult(); return result }
func ValidateSpawner(spawner *npc.NPCSpawner, snapshot WorldSnapshot) Result { result := NewResult(); return result }
func ValidateScript(script *scripts.Script, snapshot WorldSnapshot, roomOnEnter bool) Result { return ValidateLuaScript(script, roomOnEnter) }
```

- [ ] **Step 3a: Add tests for remaining validator groups**

Add these tests before implementing the remaining validator bodies:

```go
func TestValidateNPCReportsBrokenMerchantStockAndDialog(t *testing.T) {
	n := &npc.NPC{
		Entity: &entities.Entity{ID: "npc-a"},
		Name: "Merchant",
		DialogID: "missing-dialog",
		MerchantTrait: &npc.MerchantTrait{Inventory: []npc.MerchantItem{{ItemTemplateID: "missing-item", Quantity: 1, MaxQuantity: 1}}},
	}

	result := ValidateNPC(n, NewWorldSnapshot())

	if result.Errors != 2 {
		t.Fatalf("Errors = %d, want 2: %#v", result.Errors, result.Issues)
	}
}

func TestValidateLootTableReportsBrokenEntryAndInvalidRange(t *testing.T) {
	table := &items.LootTable{
		Entity: &entities.Entity{ID: "loot-a"},
		Name: "Loot",
		Entries: []items.LootEntry{{ItemTemplateID: "missing-item", DropChance: 2, MinQuantity: 5, MaxQuantity: 1}},
	}

	result := ValidateLootTable(table, NewWorldSnapshot())

	if result.Errors != 3 {
		t.Fatalf("Errors = %d, want 3: %#v", result.Errors, result.Issues)
	}
}

func TestValidateItemReportsBrokenUseScript(t *testing.T) {
	item := &items.Item{Entity: &entities.Entity{ID: "item-a"}, Name: "Potion", OnUseScriptID: "missing-script"}

	result := ValidateItem(item, NewWorldSnapshot())

	if result.Errors != 1 {
		t.Fatalf("Errors = %d, want 1: %#v", result.Errors, result.Issues)
	}
}

func TestValidateSpawnerReportsBrokenTemplateAndRoom(t *testing.T) {
	spawner := &npc.NPCSpawner{Entity: &entities.Entity{ID: "spawner-a"}, TemplateID: "missing-npc", RoomID: "missing-room"}

	result := ValidateSpawner(spawner, NewWorldSnapshot())

	if result.Errors != 2 {
		t.Fatalf("Errors = %d, want 2: %#v", result.Errors, result.Issues)
	}
}

func TestValidateDialogReportsBrokenQuestLink(t *testing.T) {
	dialog := &dialogs.Dialog{Entity: &entities.Entity{ID: "dialog-a"}, Name: "Dialog", NodeID: "main", QuestID: "missing-quest"}

	result := ValidateDialog(dialog, NewWorldSnapshot())

	if result.Errors != 1 {
		t.Fatalf("Errors = %d, want 1: %#v", result.Errors, result.Issues)
	}
}
```

Add imports for `dialogs`, `items`, and `npc` in the test file.

- [ ] **Step 3b: Implement remaining validator bodies**

Replace the stub validator bodies with explicit field checks for the test cases and the design rules:

```go
func ValidateNPC(n *npc.NPC, snapshot WorldSnapshot) Result {
	result := NewResult()
	if n == nil { return result }
	if n.SpawnRoomID != "" && !snapshot.HasRoom(n.SpawnRoomID) { result.Add(Error("missing_room", "npc", n.ID, "spawnRoomId", "room", n.SpawnRoomID, "NPC spawn room references a missing room.")) }
	if n.CurrentRoomID != "" && !snapshot.HasRoom(n.CurrentRoomID) { result.Add(Error("missing_room", "npc", n.ID, "currentRoomID", "room", n.CurrentRoomID, "NPC current room references a missing room.")) }
	if n.DialogID != "" && !snapshot.HasDialog(n.DialogID) { result.Add(Error("missing_dialog", "npc", n.ID, "dialogID", "dialog", n.DialogID, "NPC dialog references a missing dialog.")) }
	if n.IdleDialogID != "" && !snapshot.HasDialog(n.IdleDialogID) { result.Add(Error("missing_dialog", "npc", n.ID, "idleDialogID", "dialog", n.IdleDialogID, "NPC idle dialog references a missing dialog.")) }
	if n.EnemyTrait != nil {
		if n.EnemyTrait.LootTableID != "" && !snapshot.HasLootTable(n.EnemyTrait.LootTableID) { result.Add(Error("missing_loot_table", "npc", n.ID, "enemyTrait.lootTableId", "loottable", n.EnemyTrait.LootTableID, "Enemy loot table references a missing loot table.")) }
		for i, itemID := range n.EnemyTrait.GuaranteedLoot {
			if itemID != "" && !snapshot.HasItem(itemID) { result.Add(Error("missing_item", "npc", n.ID, fmt.Sprintf("enemyTrait.guaranteedLoot[%d]", i), "item", itemID, "Enemy guaranteed loot references a missing item template.")) }
		}
	}
	if n.MerchantTrait != nil {
		for i, stock := range n.MerchantTrait.Inventory {
			if stock.ItemTemplateID != "" && !snapshot.HasItem(stock.ItemTemplateID) { result.Add(Error("missing_item", "npc", n.ID, fmt.Sprintf("merchantTrait.inventory[%d].itemTemplateId", i), "item", stock.ItemTemplateID, "Merchant stock references a missing item template.")) }
			if stock.Quantity < -1 { result.Add(Warning("invalid_stock_quantity", "npc", n.ID, fmt.Sprintf("merchantTrait.inventory[%d].quantity", i), "", "", "Merchant stock quantity is invalid.")) }
			if stock.Quantity >= 0 && stock.MaxQuantity > 0 && stock.MaxQuantity < stock.Quantity { result.Add(Warning("stock_max_below_quantity", "npc", n.ID, fmt.Sprintf("merchantTrait.inventory[%d].maxQuantity", i), "", "", "Merchant max quantity is below current quantity.")) }
		}
	}
	return result
}
```

Replace the remaining stubs with these implementations:

```go
func ValidateDialog(dialog *dialogs.Dialog, snapshot WorldSnapshot) Result {
	result := NewResult()
	if dialog == nil { return result }
	seen := map[string]bool{}
	var walk func(node *dialogs.Dialog, path string)
	walk = func(node *dialogs.Dialog, path string) {
		if node == nil { return }
		if node.NodeID != "" {
			if seen[node.NodeID] {
				result.Add(Error("duplicate_dialog_node", "dialog", dialog.ID, path+".nodeId", "dialog_node", node.NodeID, "Dialog contains a duplicate node ID."))
			}
			seen[node.NodeID] = true
		}
		if node.QuestID != "" && !snapshot.HasQuest(node.QuestID) {
			result.Add(Error("missing_quest", "dialog", dialog.ID, path+".questId", "quest", node.QuestID, "Dialog quest link references a missing quest."))
		}
		for i, option := range node.Options {
			walk(option, fmt.Sprintf("%s.options[%d]", path, i))
		}
		if node.Answer != nil {
			walk(node.Answer, path+".answer")
		}
	}
	walk(dialog, "root")
	return result
}

func ValidateLootTable(table *items.LootTable, snapshot WorldSnapshot) Result {
	result := NewResult()
	if table == nil { return result }
	for i, entry := range table.Entries {
		prefix := fmt.Sprintf("entries[%d]", i)
		if entry.ItemTemplateID != "" && !snapshot.HasItem(entry.ItemTemplateID) {
			result.Add(Error("missing_item", "loottable", table.ID, prefix+".itemTemplateId", "item", entry.ItemTemplateID, "Loot entry references a missing item template."))
		}
		if entry.DropChance < 0 || entry.DropChance > 1 {
			result.Add(Error("invalid_drop_chance", "loottable", table.ID, prefix+".dropChance", "", "", "Loot entry drop chance must be between 0.0 and 1.0."))
		}
		if entry.MinQuantity > entry.MaxQuantity {
			result.Add(Error("invalid_quantity_range", "loottable", table.ID, prefix+".minQuantity", "", "", "Loot entry minimum quantity is greater than maximum quantity."))
		}
		if !entry.Guaranteed && entry.DropChance == 0 {
			result.Add(Warning("zero_drop_chance", "loottable", table.ID, prefix+".dropChance", "", "", "Non-guaranteed loot entry has a zero drop chance."))
		}
	}
	return result
}

func ValidateItem(item *items.Item, snapshot WorldSnapshot) Result {
	result := NewResult()
	if item == nil { return result }
	if item.OnUseScriptID != "" && !snapshot.HasScript(item.OnUseScriptID) {
		result.Add(Error("missing_script", "item", item.ID, "onUseScriptId", "script", item.OnUseScriptID, "Item use script references a missing script."))
	}
	if item.TemplateID != "" && !snapshot.HasItem(item.TemplateID) {
		result.Add(Error("missing_item", "item", item.ID, "templateId", "item", item.TemplateID, "Item instance references a missing template."))
	}
	if item.Stackable && item.MaxStack < 1 {
		result.Add(Warning("invalid_max_stack", "item", item.ID, "maxStack", "", "", "Stackable item has an invalid maximum stack size."))
	}
	return result
}

func ValidateSpawner(spawner *npc.NPCSpawner, snapshot WorldSnapshot) Result {
	result := NewResult()
	if spawner == nil { return result }
	if spawner.TemplateID != "" && !snapshot.HasNPC(spawner.TemplateID) {
		result.Add(Error("missing_npc", "spawner", spawner.ID, "templateId", "npc", spawner.TemplateID, "Spawner template references a missing NPC."))
	}
	if spawner.RoomID != "" && !snapshot.HasRoom(spawner.RoomID) {
		result.Add(Error("missing_room", "spawner", spawner.ID, "roomId", "room", spawner.RoomID, "Spawner room references a missing room."))
	}
	return result
}
```

- [ ] **Step 4: Run package tests**

Run:

```bash
go test ./pkg/service/validation -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/service/validation/validator.go pkg/service/validation/validator_test.go
git commit -m "feat: validate creator entity references"
```

---

### Task 4: Reusable Lua Static Validation

**Files:**
- Create: `pkg/service/validation/scripts.go`
- Modify: `pkg/importer/validator.go`
- Test: `pkg/service/validation/validator_test.go`

- [ ] **Step 1: Write failing Lua validation test**

Append:

```go
func TestValidateLuaScriptReportsUnknownGameFunction(t *testing.T) {
	script := &scripts.Script{
		Entity: &entities.Entity{ID: "script-a"},
		Name: "Bad Script",
		Language: scripts.ScriptLanguageLua,
		Type: scripts.ScriptTypeRoom,
		Code: `tales.game.nope("x")`,
	}

	result := ValidateLuaScript(script, true)

	if result.Warnings != 1 {
		t.Fatalf("Warnings = %d, want 1: %#v", result.Warnings, result.Issues)
	}
	if result.Issues[0].Code != "unknown_lua_game_function" {
		t.Fatalf("Code = %s, want unknown_lua_game_function", result.Issues[0].Code)
	}
}
```

Add `github.com/talesmud/talesmud/pkg/scripts` to imports.

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./pkg/service/validation -run TestValidateLuaScriptReportsUnknownGameFunction -count=1
```

Expected: FAIL because `ValidateLuaScript` is undefined.

- [ ] **Step 3: Move importer Lua helpers into validation package**

Create `pkg/service/validation/scripts.go` by moving these concepts out of `pkg/importer/validator.go`:

- `knownGameFunctions`
- `reGameCall`
- `reCtxRoomID`
- argument counting
- snippet extraction

Expose:

```go
func ValidateLuaScript(script *scripts.Script, isOnEnterScript bool) Result
```

Use warning codes:

- `unknown_lua_game_function`
- `wrong_lua_game_function_arg_count`
- `room_script_ctx_room_id`

- [ ] **Step 4: Update importer to use shared Lua validation**

In `pkg/importer/validator.go`, replace the body of `validateScriptCode` with an adapter that constructs a `scripts.Script`, calls `validation.ValidateLuaScript`, and records each returned issue through `w.addValidation("%s", issue.Message)`. Keep importer return value as the number of warnings.

- [ ] **Step 5: Run tests**

Run:

```bash
go test ./pkg/service/validation ./pkg/importer -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add pkg/service/validation/scripts.go pkg/service/validation/validator_test.go pkg/importer/validator.go
git commit -m "refactor: share lua script validation"
```

---

### Task 5: Validation and Preview Handler

**Files:**
- Create: `pkg/server/handler/validation.go`
- Modify: `pkg/server/server.go`
- Test: `pkg/server/handler/validation_test.go`

- [ ] **Step 1: Write failing handler test for unknown entity type**

Create `pkg/server/handler/validation_test.go`:

```go
package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestValidationHandlerRejectsUnknownEntityType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Params = gin.Params{{Key: "entityType", Value: "bad"}}

	h := &ValidationHandler{}
	h.ValidateEntity(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./pkg/server/handler -run TestValidationHandlerRejectsUnknownEntityType -count=1
```

Expected: FAIL because `ValidationHandler` does not exist.

- [ ] **Step 3: Implement handler skeleton**

Create `pkg/server/handler/validation.go`:

```go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/service"
	"github.com/talesmud/talesmud/pkg/service/validation"
)

type ValidationHandler struct {
	Facade service.Facade
}

func (h *ValidationHandler) WorldDiagnostics(c *gin.Context) {
	snapshot, err := validation.BuildSnapshot(h.Facade)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, validation.ValidateWorld(snapshot))
}

func (h *ValidationHandler) ValidateEntity(c *gin.Context) {
	entityType := c.Param("entityType")
	if !validation.IsSupportedEntityType(entityType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported entity type"})
		return
	}
	result, err := h.validateRequestBody(c, entityType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
```

Add `IsSupportedEntityType`, `ValidateWorld`, and a private `validateRequestBody` implementation. `validateRequestBody` binds the request JSON into the correct entity struct, builds a snapshot, and calls the matching validator.

- [ ] **Step 4: Register Creator routes**

In `pkg/server/server.go`, instantiate:

```go
	validationHandler := &handler.ValidationHandler{
		Facade: app.Facade,
	}
```

Inside the Creator route group add:

```go
			creator.GET("diagnostics/world", validationHandler.WorldDiagnostics)
			creator.POST("validate/:entityType", validationHandler.ValidateEntity)
			creator.POST("preview/dialog", validationHandler.PreviewDialog)
			creator.POST("preview/quest", validationHandler.PreviewQuest)
			creator.POST("preview/room", validationHandler.PreviewRoom)
			creator.POST("preview/merchant", validationHandler.PreviewMerchant)
```

- [ ] **Step 5: Run handler tests**

Run:

```bash
go test ./pkg/server/handler -run TestValidationHandler -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add pkg/server/handler/validation.go pkg/server/handler/validation_test.go pkg/server/server.go
git commit -m "feat: add creator validation endpoints"
```

---

### Task 6: Backend Save Gates

**Files:**
- Modify: `pkg/server/handler/rooms.go`
- Modify: `pkg/server/handler/items.go`
- Modify: `pkg/server/handler/npcs.go`
- Modify: `pkg/server/handler/npcspawners.go`
- Modify: `pkg/server/handler/dialogs.go`
- Modify: `pkg/server/handler/loottables.go`
- Modify: `pkg/server/handler/quests.go`
- Modify: `pkg/server/handler/scripts.go`
- Modify: `pkg/server/server.go`
- Test: `pkg/server/handler/validation_test.go`

- [ ] **Step 1: Write failing save rejection test**

Add a handler test that posts a room with an exit target of `missing-room` through a `RoomsHandler` configured with a validation service/facade. Expected status is `400` and response JSON includes `issues`.

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
go test ./pkg/server/handler -run TestRoomSaveRejectsBrokenExit -count=1
```

Expected: FAIL because room saves do not validate before storing.

- [ ] **Step 3: Add validation dependency to handlers**

For each validated handler struct, add:

```go
	Facade service.Facade
```

Then add a helper in `pkg/server/handler/validation.go`:

```go
func rejectIfInvalid(c *gin.Context, result validation.Result) bool {
	if result.Errors == 0 {
		return false
	}
	c.JSON(http.StatusBadRequest, result)
	return true
}
```

In each create/update method:

```go
snapshot, err := validation.BuildSnapshot(handler.Facade)
if err != nil {
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	return
}
if rejectIfInvalid(c, validation.ValidateRoom(&room, snapshot)) {
	return
}
```

Use the matching validator for each entity type.

- [ ] **Step 4: Pass facade into handlers**

In `pkg/server/server.go`, set `Facade: app.Facade` on rooms, items, scripts, NPCs, spawners, dialogs, loot tables, and quests handlers.

- [ ] **Step 5: Run handler tests**

Run:

```bash
go test ./pkg/server/handler -run 'Test.*Save.*Rejects|TestValidationHandler' -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add pkg/server/handler pkg/server/server.go
git commit -m "feat: block creator saves with broken references"
```

---

### Task 7: Frontend Validation API and Panel

**Files:**
- Create: `public/app/src/api/validation.js`
- Create: `public/app/src/api/previews.js`
- Create: `public/app/src/creator/ValidationPanel.svelte`
- Modify: `public/app/src/creator/CRUDEditor.svelte`

- [ ] **Step 1: Add API helpers**

Create `public/app/src/api/validation.js`:

```js
import axios from "axios";
import { backend } from "./base.js";

function authHeaders(token) {
  return { Authorization: `Bearer ${token}`, "Content-Type": "application/json" };
}

function validateEntity(token, entityType, entity, cb, errorCb) {
  axios
    .post(`${backend}/validate/${entityType}`, entity, { headers: authHeaders(token) })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

function getWorldDiagnostics(token, cb, errorCb) {
  axios
    .get(`${backend}/diagnostics/world`, { headers: { Authorization: `Bearer ${token}` } })
    .then((result) => cb(result.data))
    .catch((err) => errorCb(err));
}

export { validateEntity, getWorldDiagnostics };
```

Create `public/app/src/api/previews.js` with `previewDialog`, `previewQuest`, `previewRoom`, and `previewMerchant` functions that `POST` to the matching `/preview/*` endpoint.

- [ ] **Step 2: Add shared validation panel**

Create `public/app/src/creator/ValidationPanel.svelte`:

```svelte
<script>
  export let result = null;
  export let loading = false;
  export let unavailable = "";

  $: issues = result?.issues || [];
  $: errors = issues.filter((issue) => issue.severity === "error");
  $: warnings = issues.filter((issue) => issue.severity === "warning");
</script>

{#if loading}
  <div class="rounded border border-slate-700 bg-slate-900/50 px-3 py-2 text-xs text-slate-400">Checking content...</div>
{:else if unavailable}
  <div class="rounded border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-xs text-amber-200">{unavailable}</div>
{:else if issues.length}
  <div class="space-y-2">
    {#each errors as issue}
      <div class="rounded border border-red-500/40 bg-red-500/10 px-3 py-2 text-xs text-red-100">
        <div class="font-semibold">{issue.message}</div>
        <div class="font-mono text-[10px] opacity-80">{issue.field} {issue.refId}</div>
      </div>
    {/each}
    {#each warnings as issue}
      <div class="rounded border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-xs text-amber-100">
        <div class="font-semibold">{issue.message}</div>
        <div class="font-mono text-[10px] opacity-80">{issue.field} {issue.refId}</div>
      </div>
    {/each}
  </div>
{/if}
```

- [ ] **Step 3: Wire `CRUDEditor` validation**

Modify `public/app/src/creator/CRUDEditor.svelte`:

- import `ValidationPanel` and `validateEntity`
- add state: `validationResult`, `validationLoading`, `validationUnavailable`
- call validation when `config.entityType` and `$store.selectedElement` are present
- call validation before `create` and `update`
- block save when result has errors
- render `<ValidationPanel result={validationResult} loading={validationLoading} unavailable={validationUnavailable} />` above the form slots

- [ ] **Step 4: Build frontend**

Run:

```bash
cd public/app && npm run build
```

Expected: build succeeds.

- [ ] **Step 5: Commit**

```bash
git add public/app/src/api/validation.js public/app/src/api/previews.js public/app/src/creator/ValidationPanel.svelte public/app/src/creator/CRUDEditor.svelte
git commit -m "feat: show inline creator validation"
```

---

### Task 8: Health Tab and Row Indicators

**Files:**
- Create: `public/app/src/creator/WorldHealth.svelte`
- Modify: `public/app/src/AppContent.svelte`
- Modify: `public/app/src/creator/Creator.svelte`
- Modify: `public/app/src/creator/CRUDEditor.svelte`

- [ ] **Step 1: Add Health tab component**

Create `WorldHealth.svelte` that calls `getWorldDiagnostics`, groups `issues` by `entityType`, and renders filters for severity/entity type/text. Use existing `card`, `btn`, and `input-base` styles.

- [ ] **Step 2: Add route**

In `AppContent.svelte`, import:

```js
import WorldHealth from "./creator/WorldHealth.svelte";
```

Add:

```svelte
<Route exact path="/creator/health">
  <CreatorLayout>
    <WorldHealth />
  </CreatorLayout>
</Route>
```

- [ ] **Step 3: Add tab**

In `Creator.svelte`, add:

```js
{ name: "Health", nav: "/creator/health" },
```

- [ ] **Step 4: Add row indicators**

Use diagnostics result in `CRUDEditor` to build a row indicator:

```js
const issueIndicator = (element) => {
  const issues = issuesByEntity[element.id] || [];
  if (issues.some((issue) => issue.severity === "error")) {
    return { color: "#ef4444", title: "Has validation errors" };
  }
  if (issues.length) {
    return { color: "#f59e0b", title: "Has validation warnings" };
  }
  return null;
};
```

Pass `rowIndicator={config.rowIndicator || issueIndicator}` to `DataTable`.

- [ ] **Step 5: Build frontend**

Run:

```bash
cd public/app && npm run build
```

Expected: build succeeds.

- [ ] **Step 6: Commit**

```bash
git add public/app/src/creator/WorldHealth.svelte public/app/src/AppContent.svelte public/app/src/creator/Creator.svelte public/app/src/creator/CRUDEditor.svelte
git commit -m "feat: add creator world health diagnostics"
```

---

### Task 9: Editor Entity Types and Preview Modals

**Files:**
- Create: `public/app/src/creator/DialogPreviewModal.svelte`
- Create: `public/app/src/creator/QuestPreviewModal.svelte`
- Create: `public/app/src/creator/RoomPreviewModal.svelte`
- Create: `public/app/src/creator/MerchantPreviewModal.svelte`
- Modify: `public/app/src/creator/RoomsEditor.svelte`
- Modify: `public/app/src/creator/ItemTemplatesEditor.svelte`
- Modify: `public/app/src/creator/NPCsEditor.svelte`
- Modify: `public/app/src/creator/DialogsEditor.svelte`
- Modify: `public/app/src/creator/QuestsEditor.svelte`
- Modify: `public/app/src/creator/ScriptsEditor.svelte`

- [ ] **Step 1: Add config entity types**

For each editor config, add:

```js
entityType: "room"
```

Use the correct value per editor: `room`, `item`, `npc`, `dialog`, `quest`, `script`.

- [ ] **Step 2: Add focused preview modals**

Each modal accepts `open`, `preview`, and emits `close`. It renders server-provided summary rows and validation issues. Keep modals visually consistent with `ScriptsGuideModal.svelte` and `RoomItemsModal.svelte`.

- [ ] **Step 3: Wire preview buttons**

Use `config.extraActions` in each editor:

- Dialogs: `Preview Dialog`
- Quests: `Preview Quest`
- Rooms: `Preview Room`
- NPCs: conditionally show `Preview Merchant` when selected NPC has `merchantTrait`
- Scripts: keep `Run Test`, and render static validation result in `ValidationPanel`

- [ ] **Step 4: Build frontend**

Run:

```bash
cd public/app && npm run build
```

Expected: build succeeds.

- [ ] **Step 5: Commit**

```bash
git add public/app/src/creator/*PreviewModal.svelte public/app/src/creator/RoomsEditor.svelte public/app/src/creator/ItemTemplatesEditor.svelte public/app/src/creator/NPCsEditor.svelte public/app/src/creator/DialogsEditor.svelte public/app/src/creator/QuestsEditor.svelte public/app/src/creator/ScriptsEditor.svelte
git commit -m "feat: add creator preview tools"
```

---

### Task 10: Entity Selector Audit

**Files:**
- Modify any Creator component found by the audit that uses `<select>` for an entity ID reference.

- [ ] **Step 1: Run selector audit**

Run:

```bash
rg -n "<select|EntitySelectButton|roomId|RoomID|npcId|NPCID|itemTemplateId|dialogID|scriptId|questId" public/app/src/creator -g '!old/**'
```

Expected: every entity ID reference is either already `EntitySelectButton` or is flagged for replacement. Enum/filter selects remain allowed.

- [ ] **Step 2: Replace any entity ID `<select>`**

For every entity ID `<select>` found, replace with the variant that matches the referenced entity. Room ID example:

```svelte
<EntitySelectButton
  value={roomId}
  elements={rooms}
  columns={roomColumns}
  title="Select Room"
  placeholder="Select a room..."
  on:change={(e) => roomId = e.detail}
/>
```

Use this exact mapping:

- Room IDs: `roomColumns`
- NPC IDs: `npcColumns`
- Item or item template IDs: `itemTemplateColumns`
- Dialog IDs: `dialogColumns`
- Quest IDs: `questColumns`
- Script IDs: `scriptColumns`

- [ ] **Step 3: Build frontend**

Run:

```bash
cd public/app && npm run build
```

Expected: build succeeds.

- [ ] **Step 4: Commit**

```bash
git add public/app/src/creator
git commit -m "fix: enforce creator entity selector usage"
```

---

### Task 11: Documentation Updates

**Files:**
- Modify: `PROJECT.md`
- Modify: `ARCHITECTURE.md`
- Modify: `FEATURES.md`

- [ ] **Step 1: Update project docs**

In `PROJECT.md`, add Creator Quality Layer bullets under Content Creation:

```markdown
- Creator Quality Layer: shared backend validation, inline editor warnings, save blocking for broken references, world health diagnostics, and preview/test tools for dialogs, scripts, quests, rooms, and merchants.
```

- [ ] **Step 2: Update architecture docs**

In `ARCHITECTURE.md`, add the validation service to the service layer and route list:

```markdown
├── diagnostics/world      # Creator world health diagnostics
├── validate/:entityType   # Creator draft validation
├── preview/dialog         # Safe dialog preview
├── preview/quest          # Quest authoring preview
├── preview/room           # Room authoring preview
└── preview/merchant       # Merchant stock preview
```

Describe `pkg/service/validation` as loading a `WorldSnapshot` from the facade and returning structured issues.

- [ ] **Step 3: Update features docs**

In `FEATURES.md`, add a Creator Quality Layer subsection under Creator UI capabilities with:

- structured validation issue schema
- world health tab
- inline validation panels
- broken-reference warnings
- preview/test button coverage
- entity selector rule reaffirmation

- [ ] **Step 4: Commit**

```bash
git add PROJECT.md ARCHITECTURE.md FEATURES.md
git commit -m "docs: document creator quality layer"
```

---

### Task 12: Final Verification

**Files:**
- No source edits expected unless verification finds a defect.

- [ ] **Step 1: Run Go tests**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 2: Run frontend build**

Run:

```bash
cd public/app && npm run build
```

Expected: build succeeds.

- [ ] **Step 3: Run selector audit**

Run:

```bash
rg -n "<select" public/app/src/creator -g '!old/**'
```

Expected: remaining `<select>` usages are enum/filter controls only, not entity ID references.

- [ ] **Step 4: Run endpoint smoke tests manually with an authenticated Creator token**

Use browser dev tools or existing API client files to verify:

- `GET /api/diagnostics/world` returns JSON with `valid`, `errors`, `warnings`, and `issues`.
- `POST /api/validate/room` returns an error issue for a missing exit target.
- Saving a room with a missing exit target returns `400`.
- Creator Health tab renders the same issue.
- Room, quest, dialog, merchant, and script preview/test buttons return visible output.

- [ ] **Step 5: Commit verification fixes**

If verification requires fixes, stage the actual modified project areas:

```bash
git add pkg public PROJECT.md ARCHITECTURE.md FEATURES.md
git commit -m "fix: complete creator quality verification"
```

If no fixes are required, do not create an empty commit.
