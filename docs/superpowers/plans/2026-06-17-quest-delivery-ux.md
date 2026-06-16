# Quest Delivery + UX Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make quests flow cleanly from NPC offer through objective progress, ready-to-turn-in state, validated turn-in, rewards, player notifications, and Creator authoring feedback.

**Architecture:** Add focused quest validation and quest runtime helpers around the existing `QuestsService`, `QuestTracker`, dialog commands, player store/UI, and Creator editor. Keep the current data model and route shape, using additive message fields and local UI validation rather than new schemas or migrations.

**Tech Stack:** Go service/repository tests with SQLite temp DBs, Svelte 3/4 UIs, Rollup MUD client build, Vite Creator app build, Markdown project docs.

---

## File Structure

- Create `pkg/service/quest_validation.go`: validation rules for quest definitions.
- Modify `pkg/service/quests.go`: call validation from `Store` and `Update`; add delivery item helpers.
- Modify `pkg/service/quests_test.go`: service-level TDD coverage for validation and delivery item behavior.
- Modify `pkg/mudserver/game/quest_tracker.go`: changed-objective events, stack pickup quantities, ready-to-turn-in state, delivery inventory checks.
- Modify `pkg/mudserver/game/messages/messagetypes.go`: add a ready-to-turn-in message type.
- Modify `pkg/mudserver/game/messages/responses.go`: add optional fields for changed objective and ready-to-turn-in notifications.
- Modify `pkg/mudserver/game/commands/dialog_select.go`: centralize quest dialog action handling and enriched quest log building.
- Modify `pkg/mudserver/game/commands/talk.go`: use shared quest dialog option helpers and preserve NPC source behavior.
- Modify `pkg/mudserver/game/commands/select_character.go`: reuse enriched quest log builder.
- Modify `pkg/server/handler/quests.go`: reuse enriched quest log builder or service helper where practical.
- Modify `public/mud-client/src/game/Client.js`: create clearer quest notifications from additive message fields.
- Modify `public/mud-client/src/game/MUDXPlusStore.js`: cap notifications with stable unique IDs.
- Modify `public/mud-client/src/game/ui/QuestLog.svelte`: show ready-to-turn-in and clearer objective progress.
- Modify `public/mud-client/src/game/ui/QuestNotifications.svelte`: add ready-to-turn-in styling and overflow-safe text.
- Modify `public/app/src/creator/QuestsEditor.svelte`: add validation summary and static preview.
- Modify `PROJECT.md`, `ARCHITECTURE.md`, `FEATURES.md`, `docs/design/QUEST_AUTHORING.md`: document behavior changes.

## Task 1: Server-Side Quest Definition Validation

**Files:**
- Create: `pkg/service/quest_validation.go`
- Modify: `pkg/service/quests.go`
- Test: `pkg/service/quests_test.go`

- [ ] **Step 1: Write failing validation tests**

Append tests covering invalid and valid quest definitions:

```go
func TestStoreQuestRejectsInvalidDefinitions(t *testing.T) {
	facade := newTestFacade(t)

	tests := []struct {
		name    string
		quest   *quests.Quest
		wantErr string
	}{
		{
			name: "missing name",
			quest: &quests.Quest{
				Description: "Missing name",
				Source: quests.QuestSource{Type: "npc", NPCID: "npc-1"},
				Objectives: []quests.Objective{{ID: "kill", Type: quests.ObjectiveKill, Description: "Kill one", TargetID: "npc-1", Amount: 1}},
			},
			wantErr: "name is required",
		},
		{
			name: "npc source without npc id",
			quest: &quests.Quest{
				Name: "Broken Source", Description: "Missing NPC source", Source: quests.QuestSource{Type: "npc"},
				Objectives: []quests.Objective{{ID: "kill", Type: quests.ObjectiveKill, Description: "Kill one", TargetID: "npc-1", Amount: 1}},
			},
			wantErr: "source.npcId is required",
		},
		{
			name: "duplicate objective id",
			quest: &quests.Quest{
				Name: "Duplicate", Description: "Duplicate objective IDs", Source: quests.QuestSource{Type: "auto"},
				Objectives: []quests.Objective{
					{ID: "same", Type: quests.ObjectiveVisit, Description: "Visit A", TargetID: "room-1"},
					{ID: "same", Type: quests.ObjectiveVisit, Description: "Visit B", TargetID: "room-2"},
				},
			},
			wantErr: "objectives[1].id duplicates objectives[0].id",
		},
		{
			name: "deliver without item",
			quest: &quests.Quest{
				Name: "Delivery", Description: "Missing delivery item", Source: quests.QuestSource{Type: "auto"},
				Objectives: []quests.Objective{{ID: "deliver", Type: quests.ObjectiveDeliver, Description: "Deliver item", DeliverToNPCID: "npc-1"}},
			},
			wantErr: "objectives[0].targetId is required",
		},
		{
			name: "self prerequisite",
			quest: &quests.Quest{
				Entity: &entities.Entity{ID: "quest-self"}, Name: "Loop", Description: "Self prereq", Source: quests.QuestSource{Type: "auto"},
				RequiredQuestIDs: []string{"quest-self"},
				Objectives: []quests.Objective{{ID: "visit", Type: quests.ObjectiveVisit, Description: "Visit room", TargetID: "room-1"}},
			},
			wantErr: "requiredQuestIds[0] cannot reference this quest",
		},
		{
			name: "negative reward",
			quest: &quests.Quest{
				Name: "Bad Reward", Description: "Negative reward", Source: quests.QuestSource{Type: "auto"}, Rewards: quests.Reward{XP: -1},
				Objectives: []quests.Objective{{ID: "visit", Type: quests.ObjectiveVisit, Description: "Visit room", TargetID: "room-1"}},
			},
			wantErr: "rewards.xp cannot be negative",
		},
		{
			name: "missing referenced npc",
			quest: &quests.Quest{
				Name: "Missing NPC", Description: "References an NPC that is not stored", Source: quests.QuestSource{Type: "npc", NPCID: "missing-npc"},
				Objectives: []quests.Objective{{ID: "talk", Type: quests.ObjectiveTalk, Description: "Talk to missing NPC", TargetID: "missing-npc"}},
			},
			wantErr: "source.npcId references missing NPC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := facade.QuestsService().Store(tt.quest)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestStoreQuestAcceptsCompleteNPCKillQuest(t *testing.T) {
	facade := newTestFacade(t)
	if _, err := facade.NPCsService().Import(&npc.NPC{Entity: &entities.Entity{ID: "npc-1"}, Name: "Quest Giver"}); err != nil {
		t.Fatalf("store source npc: %v", err)
	}
	if _, err := facade.NPCsService().Import(&npc.NPC{Entity: &entities.Entity{ID: "rat-template"}, Name: "Rat", IsTemplate: true}); err != nil {
		t.Fatalf("store target npc: %v", err)
	}
	quest := &quests.Quest{
		Name: "Rat Problem", Description: "Clear the cellar.", Source: quests.QuestSource{Type: "npc", NPCID: "npc-1"},
		Objectives: []quests.Objective{{ID: "kill-rat", Type: quests.ObjectiveKill, Description: "Kill 3 rats", TargetID: "rat-template", Amount: 3}},
		Rewards: quests.Reward{XP: 25, Gold: 4},
	}
	if _, err := facade.QuestsService().Store(quest); err != nil {
		t.Fatalf("store valid quest: %v", err)
	}
}
```

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./pkg/service -run 'TestStoreQuest(RejectsInvalidDefinitions|AcceptsCompleteNPCKillQuest)' -v`

Expected: invalid-definition subtests fail because `Store` currently persists invalid quests without validation.

- [ ] **Step 3: Implement validation**

Create `pkg/service/quest_validation.go` with:

```go
package service

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/quests"
)

type questValidationError struct {
	errs []string
}

func (e *questValidationError) Error() string {
	return "invalid quest definition: " + strings.Join(e.errs, "; ")
}

func validateQuestDefinition(quest *quests.Quest) error {
	if quest == nil {
		return &questValidationError{errs: []string{"quest is required"}}
	}
	var errs []string
	add := func(format string, args ...interface{}) {
		errs = append(errs, fmt.Sprintf(format, args...))
	}

	if strings.TrimSpace(quest.Name) == "" {
		add("name is required")
	}
	if strings.TrimSpace(quest.Description) == "" {
		add("description is required")
	}
	switch quest.Source.Type {
	case "npc":
		if strings.TrimSpace(quest.Source.NPCID) == "" {
			add("source.npcId is required for npc-sourced quests")
		}
	case "item":
		if strings.TrimSpace(quest.Source.ItemID) == "" {
			add("source.itemId is required for item-sourced quests")
		}
	case "auto", "script":
	case "":
		add("source.type is required")
	default:
		add("source.type %q is invalid", quest.Source.Type)
	}
	if len(quest.Objectives) == 0 {
		add("objectives must include at least one objective")
	}
	if quest.Rewards.XP < 0 {
		add("rewards.xp cannot be negative")
	}
	if quest.Rewards.Gold < 0 {
		add("rewards.gold cannot be negative")
	}
	for i, id := range quest.Rewards.ItemTemplateIDs {
		if strings.TrimSpace(id) == "" {
			add("rewards.itemTemplateIds[%d] cannot be empty", i)
		}
	}
	for i, id := range quest.RequiredQuestIDs {
		if strings.TrimSpace(id) == "" {
			add("requiredQuestIds[%d] cannot be empty", i)
			continue
		}
		if quest.Entity != nil && quest.ID != "" && id == quest.ID {
			add("requiredQuestIds[%d] cannot reference this quest", i)
		}
	}

	seenObjectives := map[string]int{}
	for i, obj := range quest.Objectives {
		path := fmt.Sprintf("objectives[%d]", i)
		if strings.TrimSpace(obj.ID) == "" {
			add("%s.id is required", path)
		} else if first, exists := seenObjectives[obj.ID]; exists {
			add("%s.id duplicates objectives[%d].id", path, first)
		} else {
			seenObjectives[obj.ID] = i
		}
		if strings.TrimSpace(obj.Description) == "" {
			add("%s.description is required", path)
		}
		if obj.Amount < 0 {
			add("%s.amount cannot be negative", path)
		}
		switch obj.Type {
		case quests.ObjectiveKill, quests.ObjectiveCollect, quests.ObjectiveVisit:
			if strings.TrimSpace(obj.TargetID) == "" {
				add("%s.targetId is required for %s objectives", path, obj.Type)
			}
		case quests.ObjectiveDeliver:
			if strings.TrimSpace(obj.TargetID) == "" {
				add("%s.targetId is required for deliver objectives", path)
			}
			if strings.TrimSpace(obj.DeliverToNPCID) == "" {
				add("%s.deliverToNpcId is required for deliver objectives", path)
			}
		case quests.ObjectiveTalk:
			if strings.TrimSpace(obj.TargetID) == "" && strings.TrimSpace(obj.DialogNodeID) == "" {
				add("%s.targetId or %s.dialogNodeId is required for talk objectives", path, path)
			}
		case quests.ObjectiveCustom:
			if strings.TrimSpace(obj.CheckScriptID) == "" {
				add("%s.checkScriptId is required for custom objectives", path)
			}
		case "":
			add("%s.type is required", path)
		default:
			add("%s.type %q is invalid", path, obj.Type)
		}
	}

	if len(errs) > 0 {
		return &questValidationError{errs: errs}
	}
	return nil
}
```

Add facade-backed reference checks to `pkg/service/quest_validation.go`:

```go
func validateQuestReferences(quest *quests.Quest, facade Facade) error {
	if facade == nil || quest == nil {
		return nil
	}
	var errs []string
	addMissing := func(field, kind, id string) {
		errs = append(errs, fmt.Sprintf("%s references missing %s %q", field, kind, id))
	}
	if quest.Source.Type == "npc" && quest.Source.NPCID != "" {
		if found, err := facade.NPCsService().FindByID(quest.Source.NPCID); err != nil || found == nil {
			addMissing("source.npcId", "NPC", quest.Source.NPCID)
		}
	}
	if quest.Source.Type == "item" && quest.Source.ItemID != "" {
		if found, err := facade.ItemsService().FindByID(quest.Source.ItemID); err != nil || found == nil {
			addMissing("source.itemId", "item", quest.Source.ItemID)
		}
	}
	for i, id := range quest.RequiredQuestIDs {
		if id == "" {
			continue
		}
		if found, err := facade.QuestsService().FindByID(id); err != nil || found == nil {
			addMissing(fmt.Sprintf("requiredQuestIds[%d]", i), "quest", id)
		}
	}
	for i, obj := range quest.Objectives {
		path := fmt.Sprintf("objectives[%d]", i)
		switch obj.Type {
		case quests.ObjectiveKill, quests.ObjectiveTalk:
			if obj.TargetID != "" {
				if found, err := facade.NPCsService().FindByID(obj.TargetID); err != nil || found == nil {
					addMissing(path+".targetId", "NPC", obj.TargetID)
				}
			}
		case quests.ObjectiveCollect:
			if obj.TargetID != "" {
				if found, err := facade.ItemsService().FindByID(obj.TargetID); err != nil || found == nil {
					addMissing(path+".targetId", "item", obj.TargetID)
				}
			}
		case quests.ObjectiveDeliver:
			if obj.TargetID != "" {
				if found, err := facade.ItemsService().FindByID(obj.TargetID); err != nil || found == nil {
					addMissing(path+".targetId", "item", obj.TargetID)
				}
			}
			if obj.DeliverToNPCID != "" {
				if found, err := facade.NPCsService().FindByID(obj.DeliverToNPCID); err != nil || found == nil {
					addMissing(path+".deliverToNpcId", "NPC", obj.DeliverToNPCID)
				}
			}
		case quests.ObjectiveVisit:
			if obj.TargetID != "" {
				if found, err := facade.RoomsService().FindByID(obj.TargetID); err != nil || found == nil {
					addMissing(path+".targetId", "room", obj.TargetID)
				}
			}
		case quests.ObjectiveCustom:
			if obj.CheckScriptID != "" {
				if found, err := facade.ScriptsService().FindByID(obj.CheckScriptID); err != nil || found == nil {
					addMissing(path+".checkScriptId", "script", obj.CheckScriptID)
				}
			}
		}
	}
	if len(errs) > 0 {
		return &questValidationError{errs: errs}
	}
	return nil
}
```

Modify `Store` and `Update` in `pkg/service/quests.go`:

```go
func (s *questsService) Store(quest *quests.Quest) (*quests.Quest, error) {
	if err := validateQuestDefinition(quest); err != nil {
		return nil, err
	}
	if err := validateQuestReferences(quest, s.facade); err != nil {
		return nil, err
	}
	quest.Created = time.Now()
	quest.Updated = time.Now()
	return s.questsRepo.Store(quest)
}

func (s *questsService) Update(id string, quest *quests.Quest) error {
	if quest.Entity == nil {
		quest.Entity = &entities.Entity{ID: id}
	} else if quest.ID == "" {
		quest.ID = id
	}
	if err := validateQuestDefinition(quest); err != nil {
		return err
	}
	if err := validateQuestReferences(quest, s.facade); err != nil {
		return err
	}
	quest.Updated = time.Now()
	return s.questsRepo.Update(id, quest)
}
```

Add the `entities` import to `pkg/service/quests.go`. Add `strings`, `npc "github.com/talesmud/talesmud/pkg/entities/npcs"`, and any needed entity imports to `pkg/service/quests_test.go`. Update `TestAcceptQuestPrefillsCollectObjectiveWithStackQuantity` to import the referenced item template before storing the quest:

```go
if _, err := facade.ItemsService().Import(&items.Item{
	Entity: &entities.Entity{ID: templateID},
	Name: "Moon Herb",
	IsTemplate: true,
	Stackable: true,
}); err != nil {
	t.Fatalf("store item template: %v", err)
}
```

- [ ] **Step 4: Run service tests**

Run: `go test ./pkg/service -run 'TestStoreQuest|TestAcceptQuestPrefillsCollectObjectiveWithStackQuantity' -v`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/service/quest_validation.go pkg/service/quests.go pkg/service/quests_test.go
git commit -m "feat: validate quest definitions"
```

## Task 2: Delivery Objective Inventory Rules

**Files:**
- Modify: `pkg/entities/items/inventory.go`
- Modify: `pkg/service/quests.go`
- Test: `pkg/service/quests_test.go`

- [ ] **Step 1: Write failing delivery tests**

Append:

```go
func TestCompleteDeliveryObjectiveRequiresAndConsumesItems(t *testing.T) {
	facade := newTestFacade(t)
	templateID := "sealed-letter-template"
	if _, err := facade.ItemsService().Import(&items.Item{Entity: &entities.Entity{ID: templateID}, Name: "Sealed Letter", IsTemplate: true, Stackable: true}); err != nil {
		t.Fatalf("store delivery item template: %v", err)
	}
	if _, err := facade.NPCsService().Import(&npc.NPC{Entity: &entities.Entity{ID: "captain"}, Name: "Captain"}); err != nil {
		t.Fatalf("store delivery npc: %v", err)
	}
	character := &characters.Character{
		Entity: &entities.Entity{ID: "char-deliver"}, Name: "Courier", BelongsUser: *traits.BelongsToUser("user-1"),
		Inventory: items.Inventory{Size: 10, Items: []*items.Item{
			{Entity: &entities.Entity{ID: "letter-stack"}, TemplateID: templateID, Name: "Sealed Letter", Stackable: true, Quantity: 2},
		}},
	}
	if _, err := facade.CharactersService().Store(character); err != nil {
		t.Fatalf("store character: %v", err)
	}

	quest := &quests.Quest{
		Entity: &entities.Entity{ID: "quest-deliver"}, Name: "Courier Run", Description: "Deliver the letters.", Source: quests.QuestSource{Type: "auto"},
		Objectives: []quests.Objective{{ID: "deliver-letters", Type: quests.ObjectiveDeliver, Description: "Deliver 2 letters", TargetID: templateID, DeliverToNPCID: "captain", Amount: 2}},
	}
	if _, err := facade.QuestsService().Store(quest); err != nil {
		t.Fatalf("store quest: %v", err)
	}
	if _, err := facade.QuestsService().AcceptQuest(character.ID, quest.ID); err != nil {
		t.Fatalf("accept quest: %v", err)
	}

	updated, err := facade.QuestsService().CompleteDeliveryObjective(character.ID, quest.ID, "deliver-letters")
	if err != nil {
		t.Fatalf("complete delivery: %v", err)
	}
	if !updated.Objectives[0].Completed || updated.Objectives[0].Current != 2 {
		t.Fatalf("expected completed delivery objective, got %#v", updated.Objectives[0])
	}
	stored, _ := facade.CharactersService().FindByID(character.ID)
	if got := stored.Inventory.CountMatchingTemplate(templateID); got != 0 {
		t.Fatalf("expected delivered items consumed, got %d", got)
	}
}

func TestCompleteDeliveryObjectiveRejectsMissingItems(t *testing.T) {
	facade := newTestFacade(t)
	if _, err := facade.ItemsService().Import(&items.Item{Entity: &entities.Entity{ID: "item-template"}, Name: "Quest Item", IsTemplate: true}); err != nil {
		t.Fatalf("store delivery item template: %v", err)
	}
	if _, err := facade.NPCsService().Import(&npc.NPC{Entity: &entities.Entity{ID: "captain"}, Name: "Captain"}); err != nil {
		t.Fatalf("store delivery npc: %v", err)
	}
	character := &characters.Character{
		Entity: &entities.Entity{ID: "char-empty"}, Name: "Empty Courier", BelongsUser: *traits.BelongsToUser("user-1"),
		Inventory: items.Inventory{Size: 10},
	}
	if _, err := facade.CharactersService().Store(character); err != nil {
		t.Fatalf("store character: %v", err)
	}
	quest := &quests.Quest{
		Entity: &entities.Entity{ID: "quest-missing-delivery"}, Name: "Missing Item", Description: "Deliver an item.", Source: quests.QuestSource{Type: "auto"},
		Objectives: []quests.Objective{{ID: "deliver", Type: quests.ObjectiveDeliver, Description: "Deliver item", TargetID: "item-template", DeliverToNPCID: "captain", Amount: 1}},
	}
	if _, err := facade.QuestsService().Store(quest); err != nil {
		t.Fatalf("store quest: %v", err)
	}
	if _, err := facade.QuestsService().AcceptQuest(character.ID, quest.ID); err != nil {
		t.Fatalf("accept quest: %v", err)
	}
	if _, err := facade.QuestsService().CompleteDeliveryObjective(character.ID, quest.ID, "deliver"); err == nil {
		t.Fatal("expected missing delivery item error")
	}
}
```

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./pkg/service -run 'TestCompleteDeliveryObjective' -v`

Expected: FAIL because `CompleteDeliveryObjective`, `CountMatchingTemplate`, and consume helpers do not exist.

- [ ] **Step 3: Add inventory matching and consumption helpers**

Add to `pkg/entities/items/inventory.go`:

```go
func itemMatchesTemplateOrID(item *Item, targetID string) bool {
	if item == nil || targetID == "" {
		return false
	}
	if item.TemplateID != "" {
		return item.TemplateID == targetID
	}
	return item.ID == targetID
}

func (inv *Inventory) CountMatchingTemplate(targetID string) int32 {
	var count int32
	for _, item := range inv.Items {
		if !itemMatchesTemplateOrID(item, targetID) {
			continue
		}
		if item.Quantity > 0 {
			count += item.Quantity
		} else {
			count++
		}
	}
	return count
}

func (inv *Inventory) ConsumeMatchingTemplate(targetID string, amount int32) error {
	if amount < 1 {
		amount = 1
	}
	if inv.CountMatchingTemplate(targetID) < amount {
		return errors.New("not enough matching items in inventory")
	}
	remaining := amount
	filtered := inv.Items[:0]
	for _, item := range inv.Items {
		if remaining > 0 && itemMatchesTemplateOrID(item, targetID) {
			qty := item.Quantity
			if qty < 1 {
				qty = 1
			}
			if item.Stackable && qty > remaining {
				item.Quantity = qty - remaining
				remaining = 0
				filtered = append(filtered, item)
				continue
			}
			remaining -= qty
			continue
		}
		filtered = append(filtered, item)
	}
	inv.Items = filtered
	return nil
}
```

- [ ] **Step 4: Add service delivery method**

Extend `QuestsService` in `pkg/service/quests.go`:

```go
CompleteDeliveryObjective(characterID, questID, objectiveID string) (*quests.QuestProgress, error)
```

Implement:

```go
func (s *questsService) CompleteDeliveryObjective(characterID, questID, objectiveID string) (*quests.QuestProgress, error) {
	if s.facade == nil {
		return nil, errors.New("facade not initialized")
	}
	quest, err := s.questsRepo.FindByID(questID)
	if err != nil || quest == nil {
		return nil, errors.New("quest not found")
	}
	var objective *quests.Objective
	for i := range quest.Objectives {
		if quest.Objectives[i].ID == objectiveID && quest.Objectives[i].Type == quests.ObjectiveDeliver {
			objective = &quest.Objectives[i]
			break
		}
	}
	if objective == nil {
		return nil, errors.New("delivery objective not found")
	}
	character, err := s.facade.CharactersService().FindByID(characterID)
	if err != nil || character == nil {
		return nil, errors.New("character not found")
	}
	required := objective.Amount
	if required < 1 {
		required = 1
	}
	if character.Inventory.CountMatchingTemplate(objective.TargetID) < required {
		return nil, errors.New("required delivery item is not in inventory")
	}
	if err := character.Inventory.ConsumeMatchingTemplate(objective.TargetID, required); err != nil {
		return nil, err
	}
	if err := s.facade.CharactersService().Update(character.ID, character); err != nil {
		return nil, err
	}
	return s.IncrementObjective(characterID, questID, objectiveID, required)
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./pkg/service -run 'TestCompleteDeliveryObjective|TestAcceptQuestPrefillsCollectObjectiveWithStackQuantity' -v`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add pkg/entities/items/inventory.go pkg/service/quests.go pkg/service/quests_test.go
git commit -m "feat: require delivery quest items"
```

## Task 3: Quest Tracker Progress Semantics

**Files:**
- Modify: `pkg/mudserver/game/quest_tracker.go`
- Modify: `pkg/mudserver/game/messages/messagetypes.go`
- Modify: `pkg/mudserver/game/messages/responses.go`
- Test: `pkg/mudserver/game/quest_tracker_test.go`

- [ ] **Step 1: Write failing tests for stack pickup and ready state**

Create `pkg/mudserver/game/quest_tracker_test.go` with a focused unit test around pure helpers to add in this task:

```go
package game

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func TestQuestProgressMessageTypeIsReadyToTurnInWhenObjectivesComplete(t *testing.T) {
	progress := &quests.QuestProgress{
		Objectives: []quests.ObjectiveProgress{{ObjectiveID: "obj1", Current: 1, Required: 1, Completed: true}},
	}
	if got := questProgressMessageType(progress); got != messages.MessageTypeQuestReady {
		t.Fatalf("expected questReady, got %s", got)
	}
}

func TestQuestProgressMessageTypeIsProgressWhenAnyObjectiveIncomplete(t *testing.T) {
	progress := &quests.QuestProgress{
		Objectives: []quests.ObjectiveProgress{
			{ObjectiveID: "obj1", Current: 1, Required: 1, Completed: true},
			{ObjectiveID: "obj2", Current: 0, Required: 1, Completed: false},
		},
	}
	if got := questProgressMessageType(progress); got != messages.MessageTypeQuestProgress {
		t.Fatalf("expected questProgress, got %s", got)
	}
}
```

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./pkg/mudserver/game -run 'TestQuestProgressMessageType' -v`

Expected: FAIL because `MessageTypeQuestReady` and `questProgressMessageType` do not exist.

- [ ] **Step 3: Add ready message type and helper**

In `pkg/mudserver/game/messages/messagetypes.go` add:

```go
MessageTypeQuestReady = "questReady" // Quest objectives complete and ready for turn-in
```

In `pkg/mudserver/game/messages/responses.go`, add an optional changed-objective field while preserving the existing `objectives` payload:

```go
type QuestUpdateMessage struct {
	MessageResponse
	QuestName        string                  `json:"questName"`
	QuestID          string                  `json:"questId"`
	Status           string                  `json:"status"`
	Objectives       []QuestObjectiveProgress `json:"objectives,omitempty"`
	ChangedObjective *QuestObjectiveProgress `json:"changedObjective,omitempty"`
}
```

In `pkg/mudserver/game/quest_tracker.go` add:

```go
func questProgressMessageType(progress *quests.QuestProgress) messages.MessageType {
	for _, obj := range progress.Objectives {
		if !obj.Completed {
			return messages.MessageTypeQuestProgress
		}
	}
	return messages.MessageTypeQuestReady
}
```

Update `sendProgressUpdate` to use `questProgressMessageType(progress)` instead of `MessageTypeQuestCompleted`.

Add a `changedObjectiveID` argument to `sendProgressUpdate`, find the matching objective in `buildObjectiveProgress`, and set `msg.ChangedObjective` before sending:

```go
func (qt *QuestTracker) sendProgressUpdate(userID string, quest *quests.Quest, progress *quests.QuestProgress, changedObjectiveID string) {
	objectives := qt.buildObjectiveProgress(quest, progress)
	msgType := questProgressMessageType(progress)
	msg := messages.NewQuestUpdateMessage(userID, quest.ID, quest.Name, string(msgType), objectives)
	for i := range objectives {
		if objectives[i].ObjectiveID == changedObjectiveID {
			msg.ChangedObjective = &objectives[i]
			break
		}
	}
	qt.game.SendMessage() <- msg
}
```

Update all `qt.sendProgressUpdate(...)` calls to pass `obj.ID`.

- [ ] **Step 4: Make pickup quantity and delivery behavior precise**

Change pickup progress in `OnItemPickup`:

```go
amount := item.Quantity
if amount < 1 {
	amount = 1
}
updated, err := qt.facade.QuestsService().IncrementObjective(characterID, progress.QuestID, obj.ID, amount)
```

Change deliver progress in `OnTalkToNPC` to call:

```go
updated, err := qt.facade.QuestsService().CompleteDeliveryObjective(characterID, progress.QuestID, obj.ID)
if err != nil {
	continue
}
qt.sendProgressUpdate(userID, quest, updated)
```

- [ ] **Step 5: Run tracker tests**

Run: `go test ./pkg/mudserver/game -run 'TestQuestProgressMessageType' -v`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add pkg/mudserver/game/quest_tracker.go pkg/mudserver/game/messages/messagetypes.go pkg/mudserver/game/messages/responses.go pkg/mudserver/game/quest_tracker_test.go
git commit -m "feat: report quest ready state"
```

## Task 4: Shared Quest Dialog Actions And Enriched Quest Logs

**Files:**
- Modify: `pkg/mudserver/game/commands/dialog_select.go`
- Modify: `pkg/mudserver/game/commands/talk.go`
- Modify: `pkg/mudserver/game/commands/select_character.go`
- Modify: `pkg/server/handler/quests.go`
- Test: `pkg/mudserver/game/commands/quest_messages_test.go`

- [ ] **Step 1: Write failing tests for enriched quest log conversion**

Create `pkg/mudserver/game/commands/quest_messages_test.go`:

```go
package commands

import (
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/quests"
)

func TestBuildQuestLogEntryIncludesDefinitionDetails(t *testing.T) {
	acceptedAt := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	quest := &quests.Quest{
		Entity: &entities.Entity{ID: "quest-1"}, Name: "Rat Problem", Description: "Clear the cellar.", Category: "side", Level: 2,
		Objectives: []quests.Objective{{ID: "kill-rats", Type: quests.ObjectiveKill, Description: "Kill 3 rats", Amount: 3}},
		Rewards: quests.Reward{XP: 25, Gold: 4},
	}
	progress := &quests.QuestProgress{
		QuestID: quest.ID, Status: quests.QuestStatusActive, AcceptedAt: acceptedAt,
		Objectives: []quests.ObjectiveProgress{{ObjectiveID: "kill-rats", Current: 1, Required: 3}},
	}
	entry := buildQuestLogEntry(quest, progress)
	if entry.QuestName != quest.Name || entry.Description != quest.Description || entry.Objectives[0].Description != "Kill 3 rats" {
		t.Fatalf("entry missing quest details: %#v", entry)
	}
	if entry.Rewards == nil || entry.Rewards.XP != 25 || entry.Rewards.Gold != 4 {
		t.Fatalf("entry missing rewards: %#v", entry.Rewards)
	}
	if entry.AcceptedAt == "" {
		t.Fatal("entry missing accepted timestamp")
	}
}
```

- [ ] **Step 2: Run tests and verify they fail**

Run: `go test ./pkg/mudserver/game/commands -run TestBuildQuestLogEntryIncludesDefinitionDetails -v`

Expected: FAIL because `buildQuestLogEntry` does not exist.

- [ ] **Step 3: Add shared quest log builder**

Add to `dialog_select.go` or a new `quest_messages.go` in the same package:

```go
func buildQuestLogEntry(quest *quests.Quest, progress *quests.QuestProgress) messages.QuestLogEntry {
	objectives := make([]messages.QuestObjectiveProgress, 0, len(progress.Objectives))
	for _, op := range progress.Objectives {
		objDesc := ""
		required := op.Required
		for _, obj := range quest.Objectives {
			if obj.ID == op.ObjectiveID {
				objDesc = obj.Description
				if obj.Amount > 0 {
					required = obj.Amount
				}
				break
			}
		}
		objectives = append(objectives, messages.QuestObjectiveProgress{
			ObjectiveID: op.ObjectiveID,
			Description: objDesc,
			Current: op.Current,
			Required: required,
			Completed: op.Completed,
		})
	}
	entry := messages.QuestLogEntry{
		QuestID: progress.QuestID,
		QuestName: quest.Name,
		Status: string(progress.Status),
		Description: quest.Description,
		Category: quest.Category,
		Level: quest.Level,
		Objectives: objectives,
		Rewards: &messages.QuestReward{XP: quest.Rewards.XP, Gold: quest.Rewards.Gold, ItemTemplateIDs: quest.Rewards.ItemTemplateIDs},
	}
	if !progress.AcceptedAt.IsZero() {
		entry.AcceptedAt = progress.AcceptedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if !progress.CompletedAt.IsZero() {
		entry.CompletedAt = progress.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return entry
}
```

Refactor existing `convertQuestLogToEntries` and `sendQuestLogToPlayer` to load quest definitions and call `buildQuestLogEntry`.

- [ ] **Step 4: Refactor dialog quest action duplication**

Add a shared function:

```go
func handleQuestDialogAction(game def.GameCtrl, message *messages.Message, npcName, questID, action, npcText string, activeConv *conversations.Conversation) {
	if npcText != "" {
		game.SendMessage() <- message.Reply("[" + npcName + "] " + npcText)
	}
	switch action {
	case "accept":
		sendQuestAccepted(game, message, npcName, questID)
	case "complete":
		sendQuestCompleted(game, message, npcName, questID)
	case "progress":
		sendQuestProgressSummary(game, message, npcName, questID)
	}
	game.GetFacade().ConversationsService().ResetConversation(activeConv)
}
```

Update `handleQuestAction` and `handleQuestDialogOption` to call this function instead of maintaining separate accept/complete/progress branches.

- [ ] **Step 5: Run command tests**

Run: `go test ./pkg/mudserver/game/commands -run TestBuildQuestLogEntryIncludesDefinitionDetails -v`

Expected: PASS.

- [ ] **Step 6: Run broader backend tests**

Run: `go test ./pkg/service ./pkg/mudserver/game ./pkg/mudserver/game/commands ./pkg/server/handler -v`

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add pkg/mudserver/game/commands pkg/server/handler/quests.go
git commit -m "refactor: unify quest dialog actions"
```

## Task 5: Player Quest Notifications And Log Clarity

**Files:**
- Modify: `public/mud-client/src/game/Client.js`
- Modify: `public/mud-client/src/game/MUDXPlusStore.js`
- Modify: `public/mud-client/src/game/ui/QuestLog.svelte`
- Modify: `public/mud-client/src/game/ui/QuestNotifications.svelte`

- [ ] **Step 1: Build once before changes**

Run: `npm --prefix public/mud-client run build`

Expected: PASS before edits.

- [ ] **Step 2: Update notification handling**

In `Client.js`, add a handler for `questReady`:

```js
messageHandlers["questReady"] = (msg) => {
  renderer(msg.message);
  requestQuestLog();
  if (mux) {
    mux.addQuestNotification({
      id: `quest-ready-${msg.questId || Date.now()}-${Date.now()}`,
      questId: msg.questId,
      type: "ready",
      questName: msg.questName || "Quest",
      message: "Ready to turn in",
    });
  }
};
```

Revise `questProgress` to prefer changed objective counts:

```js
const changed = msg.changedObjective || (msg.objectives || []).find(o => !o.completed);
if (changed) {
  mux.addQuestNotification({
    id: `quest-progress-${msg.questId || Date.now()}-${changed.objectiveId || Date.now()}-${changed.current || 0}`,
    questId: msg.questId,
    type: changed.completed ? "objective-complete" : "progress",
    questName: msg.questName || "Quest",
    message: changed.completed
      ? `Objective complete: ${changed.description || "Objective"}`
      : `${changed.description || "Objective"} (${changed.current}/${changed.required})`,
  });
}
```

- [ ] **Step 3: Cap notifications**

In `MUDXPlusStore.js`, update `addQuestNotification`:

```js
addQuestNotification: (notification) => {
  const enriched = { ...notification, id: notification.id || `quest-note-${Date.now()}-${Math.random()}` };
  update((state) => {
    state.questNotifications = [...state.questNotifications, enriched].slice(-5);
    return state;
  });
  setTimeout(() => {
    update((state) => {
      state.questNotifications = state.questNotifications.filter(n => n.id !== enriched.id);
      return state;
    });
  }, 5000);
},
```

- [ ] **Step 4: Add ready state to QuestLog**

Add helper in `QuestLog.svelte`:

```js
function isReadyToTurnIn(quest) {
  return quest?.status === "active" && (quest.objectives || []).length > 0 && (quest.objectives || []).every(obj => obj.completed);
}
```

Use it in quest rows:

```svelte
{#if isReadyToTurnIn(quest)}
  <span class="quest-badge ready-badge">Ready</span>
{/if}
```

Add CSS:

```css
.ready-badge {
  background: #16a34a;
  color: white;
}
.objective.completed .quest-progress {
  color: #22c55e;
}
```

- [ ] **Step 5: Add ready notification styling**

In `QuestNotifications.svelte`:

```css
.notification.ready {
  border-left: 4px solid #22c55e;
}
.notification.ready .notification-title {
  color: #22c55e;
}
.notification.objective-complete {
  border-left: 4px solid #84cc16;
}
.notification.objective-complete .notification-title {
  color: #84cc16;
}
.notification-message {
  overflow-wrap: anywhere;
}
```

- [ ] **Step 6: Build MUD client**

Run: `npm --prefix public/mud-client run build`

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add public/mud-client/src/game
git commit -m "feat: clarify quest progress notifications"
```

## Task 6: Creator Quest Validation And Preview

**Files:**
- Modify: `public/app/src/creator/QuestsEditor.svelte`

- [ ] **Step 1: Build once before changes**

Run: `npm --prefix public/app run build`

Expected: PASS before edits.

- [ ] **Step 2: Add local validation helpers**

In `QuestsEditor.svelte`, add:

```js
function validateQuestDraft(quest) {
  const errors = [];
  const warnings = [];
  const addError = (field, message) => errors.push({ field, message });
  const addWarning = (field, message) => warnings.push({ field, message });

  if (!quest?.name?.trim()) addError("name", "Name is required.");
  if (!quest?.description?.trim()) addError("description", "Description is required.");
  if (!quest?.source?.type) addError("source.type", "Source type is required.");
  if (quest?.source?.type === "npc" && !quest.source.npcId) addError("source.npcId", "Source NPC is required.");
  if (quest?.source?.type === "item" && !quest.source.itemId) addError("source.itemId", "Source item is required.");
  if (!quest?.objectives?.length) addError("objectives", "At least one objective is required.");
  if ((quest?.rewards?.xp || 0) < 0) addError("rewards.xp", "XP cannot be negative.");
  if ((quest?.rewards?.gold || 0) < 0) addError("rewards.gold", "Gold cannot be negative.");

  const seen = new Map();
  (quest?.objectives || []).forEach((obj, i) => {
    const path = `objectives[${i}]`;
    if (!obj.id?.trim()) addError(`${path}.id`, "Objective ID is required.");
    if (seen.has(obj.id)) addError(`${path}.id`, `Objective ID duplicates objectives[${seen.get(obj.id)}].id.`);
    if (obj.id) seen.set(obj.id, i);
    if (!obj.description?.trim()) addError(`${path}.description`, "Objective description is required.");
    if ((obj.amount || 0) < 0) addError(`${path}.amount`, "Amount cannot be negative.");
    if (["kill", "collect", "visit"].includes(obj.type) && !obj.targetId) addError(`${path}.targetId`, "Target is required.");
    if (obj.type === "deliver" && !obj.targetId) addError(`${path}.targetId`, "Delivery item is required.");
    if (obj.type === "deliver" && !obj.deliverToNpcId) addError(`${path}.deliverToNpcId`, "Delivery NPC is required.");
    if (obj.type === "talk" && !obj.targetId && !obj.dialogNodeId) addError(`${path}.targetId`, "Talk objectives need an NPC or dialog node.");
    if (obj.type === "custom" && !obj.checkScriptId) addError(`${path}.checkScriptId`, "Custom objectives need a check script.");
    if (obj.type === "deliver" && quest?.source?.type !== "npc") addWarning(`${path}.deliverToNpcId`, "Delivery quests usually read best when sourced from an NPC.");
  });

  return { errors, warnings };
}

$: questValidation = validateQuestDraft($store.selectedElement || {});
```

- [ ] **Step 3: Render validation summary**

Near the top of the `<div slot="content">` block:

```svelte
{#if questValidation.errors.length || questValidation.warnings.length}
  <section class="rounded border border-slate-700/60 bg-slate-900/40 p-4 space-y-3">
    <div class="text-xs font-bold uppercase tracking-wider text-slate-300">Quest Validation</div>
    {#if questValidation.errors.length}
      <div class="space-y-1">
        {#each questValidation.errors as item}
          <div class="text-xs text-red-300"><span class="font-mono">{item.field}</span>: {item.message}</div>
        {/each}
      </div>
    {/if}
    {#if questValidation.warnings.length}
      <div class="space-y-1">
        {#each questValidation.warnings as item}
          <div class="text-xs text-amber-300"><span class="font-mono">{item.field}</span>: {item.message}</div>
        {/each}
      </div>
    {/if}
  </section>
{/if}
```

- [ ] **Step 4: Render static preview**

Below Dialog Text:

```svelte
<section class="rounded border border-slate-700/60 bg-slate-950/30 p-4 space-y-3">
  <div class="text-xs font-bold uppercase tracking-wider text-slate-300">Player Preview</div>
  <div class="space-y-2 text-sm">
    <div class="font-semibold text-slate-100">{$store.selectedElement.name || "Unnamed Quest"}</div>
    <p class="text-slate-400 text-xs">{$store.selectedElement.description || "No description yet."}</p>
    <div class="text-xs text-slate-400">
      Source: {$store.selectedElement.source?.type || "none"}
    </div>
    <div>
      <div class="label-caps">Objectives</div>
      {#each $store.selectedElement.objectives || [] as obj}
        <div class="text-xs text-slate-300">[ ] {obj.description || obj.id || "Unnamed objective"} ({obj.amount || 1})</div>
      {/each}
    </div>
    <div>
      <div class="label-caps">Dialog States</div>
      <div class="text-xs text-slate-400">Offer: {$store.selectedElement.acceptDialogText || "Default quest offer text"}</div>
      <div class="text-xs text-slate-400">In Progress: {$store.selectedElement.progressDialogText || "Default progress text"}</div>
      <div class="text-xs text-slate-400">Ready: {$store.selectedElement.completeDialogText || "Default turn-in text"}</div>
    </div>
  </div>
</section>
```

- [ ] **Step 5: Build Creator app**

Run: `npm --prefix public/app run build`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add public/app/src/creator/QuestsEditor.svelte
git commit -m "feat: preview and validate quest drafts"
```

## Task 7: Documentation Sync

**Files:**
- Modify: `PROJECT.md`
- Modify: `ARCHITECTURE.md`
- Modify: `FEATURES.md`
- Modify: `docs/design/QUEST_AUTHORING.md`

- [ ] **Step 1: Update docs**

Document these facts:

```markdown
- NPC quest options show offer, progress, ready-to-turn-in, and completion actions.
- Deliver objectives require matching inventory items and consume them when delivered.
- QuestTracker emits progress for changed objectives and ready-to-turn-in when all objectives are complete.
- Quest create/update rejects impossible definitions with field-specific errors.
- Creator quest editing includes validation and a static player preview.
```

- [ ] **Step 2: Check docs for stale deliver behavior**

Run: `rg -n "deliver.*does not|does \\*\\*not\\*\\* currently remove|onCompleteScriptId.*removal|questCompleted.*objectives" PROJECT.md ARCHITECTURE.md FEATURES.md docs/design/QUEST_AUTHORING.md`

Expected: no stale claims that delivery does not remove items.

- [ ] **Step 3: Commit**

```bash
git add PROJECT.md ARCHITECTURE.md FEATURES.md docs/design/QUEST_AUTHORING.md
git commit -m "docs: update quest delivery behavior"
```

## Task 8: Final Verification

**Files:**
- No planned edits.

- [ ] **Step 1: Run backend tests**

Run: `go test ./...`

Expected: PASS.

- [ ] **Step 2: Run Creator build**

Run: `npm --prefix public/app run build`

Expected: PASS.

- [ ] **Step 3: Run MUD client build**

Run: `npm --prefix public/mud-client run build`

Expected: PASS.

- [ ] **Step 4: Inspect git state**

Run: `git status --short`

Expected: clean working tree.

- [ ] **Step 5: Completion audit**

Verify each acceptance criterion from `docs/superpowers/specs/2026-06-17-quest-delivery-ux-design.md` against current code, tests, builds, and docs before reporting the goal complete.
