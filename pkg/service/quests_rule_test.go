package service

import (
	"errors"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/conversations"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/settings"
	"github.com/talesmud/talesmud/pkg/entities/skills"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/scripts"
	"github.com/talesmud/talesmud/pkg/server/dto"
)

func TestValidateQuestRejectsMissingSourceNPC(t *testing.T) {
	svc := newQuestValidationTestService()
	quest := validTestQuest()
	quest.Source.NPCID = "missing-npc"

	issues := svc.ValidateQuest(quest)

	assertHasIssue(t, issues, "missing_source_npc")
}

func TestValidateQuestRejectsDuplicateObjectiveIDs(t *testing.T) {
	svc := newQuestValidationTestService()
	quest := validTestQuest()
	quest.Objectives = append(quest.Objectives, quests.Objective{
		ID:          "obj1",
		Type:        quests.ObjectiveVisit,
		Description: "Return to the hall",
		TargetID:    "room1",
		Amount:      1,
	})

	issues := svc.ValidateQuest(quest)

	assertHasIssue(t, issues, "duplicate_objective_id")
}

func TestValidateQuestRejectsSelfPrerequisite(t *testing.T) {
	svc := newQuestValidationTestService()
	quest := validTestQuest()
	quest.RequiredQuestIDs = []string{quest.ID}

	issues := svc.ValidateQuest(quest)

	assertHasIssue(t, issues, "self_prerequisite")
}

func TestValidateQuestRejectsNonNPCSourceWithoutTurnInNPC(t *testing.T) {
	svc := newQuestValidationTestService()
	quest := validTestQuest()
	quest.Source = quests.QuestSource{Type: "auto"}
	quest.Objectives = []quests.Objective{{
		ID:          "obj1",
		Type:        quests.ObjectiveVisit,
		Description: "Find the ruin",
		TargetID:    "room1",
		Amount:      1,
	}}

	issues := svc.ValidateQuest(quest)

	assertHasIssue(t, issues, "missing_turn_in_npc")
}

func TestValidateQuestAcceptsValidNPCQuest(t *testing.T) {
	svc := newQuestValidationTestService()

	issues := svc.ValidateQuest(validTestQuest())

	for _, issue := range issues {
		if issue.Severity == "error" {
			t.Fatalf("expected no validation errors, got %#v", issues)
		}
	}
}

func TestBuildQuestLogIncludesObjectiveDescriptionsAndReadyFlag(t *testing.T) {
	quest := validTestQuest()
	quest.Objectives[0].Description = "Bring back the relic"
	quest.Objectives[0].Amount = 1
	questRepo := &fakeQuestsRepo{quests: map[string]*quests.Quest{
		quest.ID: quest,
	}}
	progressRepo := &fakeQuestProgressRepo{progress: []*quests.QuestProgress{{
		Entity:      &entities.Entity{ID: "progress1"},
		CharacterID: "char1",
		QuestID:     quest.ID,
		Status:      quests.QuestStatusActive,
		Objectives: []quests.ObjectiveProgress{{
			ObjectiveID: "obj1",
			Current:     1,
			Required:    1,
			Completed:   true,
		}},
	}}}
	svc := NewQuestsService(questRepo, progressRepo)

	entries, err := svc.BuildQuestLog("char1")

	if err != nil {
		t.Fatalf("BuildQuestLog returned error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 quest log entry, got %d", len(entries))
	}
	entry := entries[0]
	if entry.QuestName != "Find the Relic" {
		t.Fatalf("expected quest name to be enriched, got %q", entry.QuestName)
	}
	if len(entry.Objectives) != 1 || entry.Objectives[0].Description != "Bring back the relic" {
		t.Fatalf("expected objective description to be enriched, got %#v", entry.Objectives)
	}
	if !entry.ReadyToTurnIn {
		t.Fatalf("expected completed active quest to be ready to turn in")
	}
}

func TestApplyQuestEventKillUpdatesMatchingActiveObjective(t *testing.T) {
	fixture := newQuestRuleTestFixture(validTestQuest())

	results, err := fixture.svc.ApplyQuestEvent(QuestEvent{
		Type:        QuestEventNPCKilled,
		CharacterID: "char1",
		NPCID:       "npc1",
	})

	if err != nil {
		t.Fatalf("ApplyQuestEvent returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 event result, got %d", len(results))
	}
	if results[0].Kind != QuestEventResultReadyToTurnIn {
		t.Fatalf("expected ready-to-turn-in result, got %q", results[0].Kind)
	}
	progress := fixture.progressRepo.progress[0]
	if got := progress.Objectives[0].Current; got != 1 {
		t.Fatalf("expected kill objective current to be 1, got %d", got)
	}
	if !progress.Objectives[0].Completed {
		t.Fatalf("expected kill objective to be completed")
	}
}

func TestApplyQuestEventVisitDoesNotIncrementCompletedObjective(t *testing.T) {
	quest := validTestQuest()
	quest.Objectives = []quests.Objective{{
		ID:          "obj1",
		Type:        quests.ObjectiveVisit,
		Description: "Visit the hall",
		TargetID:    "room1",
		Amount:      1,
	}}
	fixture := newQuestRuleTestFixture(quest)
	fixture.progressRepo.progress[0].Objectives[0].Current = 1
	fixture.progressRepo.progress[0].Objectives[0].Completed = true

	results, err := fixture.svc.ApplyQuestEvent(QuestEvent{
		Type:        QuestEventRoomEnter,
		CharacterID: "char1",
		RoomID:      "room1",
	})

	if err != nil {
		t.Fatalf("ApplyQuestEvent returned error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected no result for already completed objective, got %#v", results)
	}
	if got := fixture.progressRepo.progress[0].Objectives[0].Current; got != 1 {
		t.Fatalf("expected completed visit current to stay 1, got %d", got)
	}
}

func TestApplyQuestEventDeliverRequiresMatchingNPCAndItem(t *testing.T) {
	quest := validTestQuest()
	quest.Objectives = []quests.Objective{{
		ID:             "obj1",
		Type:           quests.ObjectiveDeliver,
		Description:    "Deliver the relic",
		TargetID:       "item1",
		DeliverToNPCID: "npc1",
		Amount:         1,
	}}
	fixture := newQuestRuleTestFixture(quest)

	results, err := fixture.svc.ApplyQuestEvent(QuestEvent{
		Type:        QuestEventTalkToNPC,
		CharacterID: "char1",
		NPCID:       "npc1",
	})
	if err != nil {
		t.Fatalf("ApplyQuestEvent returned error without item: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected no deliver progress without item, got %#v", results)
	}

	fixture.character.Inventory.Items = append(fixture.character.Inventory.Items, &items.Item{
		Entity:     &entities.Entity{ID: "item-instance-1"},
		Name:       "Relic",
		TemplateID: "item1",
	})
	results, err = fixture.svc.ApplyQuestEvent(QuestEvent{
		Type:        QuestEventTalkToNPC,
		CharacterID: "char1",
		NPCID:       "npc1",
	})

	if err != nil {
		t.Fatalf("ApplyQuestEvent returned error with item: %v", err)
	}
	if len(results) != 1 || results[0].Kind != QuestEventResultReadyToTurnIn {
		t.Fatalf("expected ready deliver result, got %#v", results)
	}
	if !fixture.progressRepo.progress[0].Objectives[0].Completed {
		t.Fatalf("expected deliver objective to complete")
	}
}

func TestTurnInQuestGrantsRewardsAndCompletesOnce(t *testing.T) {
	quest := validTestQuest()
	quest.Rewards.XP = 25
	quest.Rewards.Gold = 7
	quest.Rewards.ItemTemplateIDs = []string{"item1"}
	fixture := newQuestRuleTestFixture(quest)
	fixture.progressRepo.progress[0].Objectives[0].Current = 1
	fixture.progressRepo.progress[0].Objectives[0].Completed = true

	result, err := fixture.svc.TurnInQuest("char1", quest.ID, "npc1")

	if err != nil {
		t.Fatalf("TurnInQuest returned error: %v", err)
	}
	if result.QuestName != quest.Name {
		t.Fatalf("expected turn-in result quest name %q, got %q", quest.Name, result.QuestName)
	}
	if fixture.progressRepo.progress[0].Status != quests.QuestStatusCompleted {
		t.Fatalf("expected quest progress to be completed, got %q", fixture.progressRepo.progress[0].Status)
	}
	if fixture.character.XP != 25 || fixture.character.Gold != 7 {
		t.Fatalf("expected rewards to update XP/gold, got xp=%d gold=%d", fixture.character.XP, fixture.character.Gold)
	}
	if len(fixture.character.Inventory.Items) != 1 {
		t.Fatalf("expected one reward item in inventory, got %d", len(fixture.character.Inventory.Items))
	}

	_, err = fixture.svc.TurnInQuest("char1", quest.ID, "npc1")
	if err == nil {
		t.Fatalf("expected second turn-in to fail")
	}
}

func TestTurnInQuestRejectsWrongNPC(t *testing.T) {
	fixture := newQuestRuleTestFixture(validTestQuest())
	fixture.progressRepo.progress[0].Objectives[0].Current = 1
	fixture.progressRepo.progress[0].Objectives[0].Completed = true

	_, err := fixture.svc.TurnInQuest("char1", "quest1", "npc2")

	if err == nil {
		t.Fatalf("expected wrong turn-in NPC to be rejected")
	}
	if fixture.progressRepo.progress[0].Status != quests.QuestStatusActive {
		t.Fatalf("expected quest to remain active, got %q", fixture.progressRepo.progress[0].Status)
	}
}

func assertHasIssue(t *testing.T, issues []QuestValidationIssue, code string) {
	t.Helper()
	for _, issue := range issues {
		if issue.Code == code && issue.Severity == "error" {
			return
		}
	}
	t.Fatalf("expected error issue %q, got %#v", code, issues)
}

func validTestQuest() *quests.Quest {
	return &quests.Quest{
		Entity:      &entities.Entity{ID: "quest1"},
		Name:        "Find the Relic",
		Description: "Recover the old relic.",
		Source:      quests.QuestSource{Type: "npc", NPCID: "npc1"},
		Objectives: []quests.Objective{{
			ID:          "obj1",
			Type:        quests.ObjectiveKill,
			Description: "Defeat the guard",
			TargetID:    "npc1",
			Amount:      1,
		}},
	}
}

func newQuestValidationTestService() QuestsService {
	questRepo := &fakeQuestsRepo{quests: map[string]*quests.Quest{
		"quest1": validTestQuest(),
	}}
	progressRepo := &fakeQuestProgressRepo{}
	svc := NewQuestsService(questRepo, progressRepo)
	svc.SetFacade(&fakeQuestValidationFacade{
		npcs: &fakeNPCsRepo{npcs: map[string]*npc.NPC{
			"npc1": {Entity: &entities.Entity{ID: "npc1"}, Name: "Guard", IsTemplate: true},
		}},
		items: &fakeItemsRepo{items: map[string]*items.Item{
			"item1": {Entity: &entities.Entity{ID: "item1"}, Name: "Relic", IsTemplate: true},
		}},
		rooms: &fakeRoomsRepo{rooms: map[string]*rooms.Room{
			"room1": {Entity: &entities.Entity{ID: "room1"}, Name: "Hall"},
		}},
		scripts: &fakeScriptsRepo{scripts: map[string]*scripts.Script{
			"script1": {Entity: &entities.Entity{ID: "script1"}, Name: "Quest Check"},
		}},
	})
	return svc
}

type questRuleTestFixture struct {
	svc          QuestsService
	progressRepo *fakeQuestProgressRepo
	character    *characters.Character
}

func newQuestRuleTestFixture(quest *quests.Quest) questRuleTestFixture {
	progressRepo := &fakeQuestProgressRepo{progress: []*quests.QuestProgress{{
		Entity:      &entities.Entity{ID: "progress1"},
		CharacterID: "char1",
		QuestID:     quest.ID,
		Status:      quests.QuestStatusActive,
		Objectives: []quests.ObjectiveProgress{{
			ObjectiveID: quest.Objectives[0].ID,
			Current:     0,
			Required:    1,
			Completed:   false,
		}},
	}}}
	questRepo := &fakeQuestsRepo{quests: map[string]*quests.Quest{
		quest.ID: quest,
	}}
	character := &characters.Character{
		Entity: &entities.Entity{ID: "char1"},
		Name:   "Hero",
		Inventory: items.Inventory{
			Size:  10,
			Items: []*items.Item{},
		},
	}
	svc := NewQuestsService(questRepo, progressRepo)
	svc.SetFacade(&fakeQuestValidationFacade{
		chars: &fakeCharactersService{characters: map[string]*characters.Character{
			"char1": character,
		}},
		npcs: &fakeNPCsRepo{npcs: map[string]*npc.NPC{
			"npc1": {Entity: &entities.Entity{ID: "npc1"}, Name: "Guard", IsTemplate: true},
			"npc2": {Entity: &entities.Entity{ID: "npc2"}, Name: "Other", IsTemplate: true},
		}},
		items: &fakeItemsRepo{items: map[string]*items.Item{
			"item1": {Entity: &entities.Entity{ID: "item1"}, Name: "Relic", IsTemplate: true},
		}},
		rooms: &fakeRoomsRepo{rooms: map[string]*rooms.Room{
			"room1": {Entity: &entities.Entity{ID: "room1"}, Name: "Hall"},
		}},
		scripts: &fakeScriptsRepo{scripts: map[string]*scripts.Script{
			"script1": {Entity: &entities.Entity{ID: "script1"}, Name: "Quest Check"},
		}},
	})
	return questRuleTestFixture{svc: svc, progressRepo: progressRepo, character: character}
}

type fakeQuestValidationFacade struct {
	chars   *fakeCharactersService
	npcs    *fakeNPCsRepo
	items   *fakeItemsRepo
	rooms   *fakeRoomsRepo
	scripts *fakeScriptsRepo
}

func (f *fakeQuestValidationFacade) CharactersService() CharactersService { return f.chars }
func (f *fakeQuestValidationFacade) PartiesService() PartiesService       { return nil }
func (f *fakeQuestValidationFacade) UsersService() UsersService           { return nil }
func (f *fakeQuestValidationFacade) RoomsService() RoomsService           { return NewRoomsService(f.rooms) }
func (f *fakeQuestValidationFacade) ScriptsService() ScriptsService {
	return NewScriptsService(f.scripts)
}
func (f *fakeQuestValidationFacade) ItemsService() ItemsService { return NewItemsService(f.items) }
func (f *fakeQuestValidationFacade) NPCsService() NPCsService   { return NewNPCsService(f.npcs) }
func (f *fakeQuestValidationFacade) NPCSpawnersService() NPCSpawnersService {
	return nil
}
func (f *fakeQuestValidationFacade) DialogsService() DialogsService               { return nil }
func (f *fakeQuestValidationFacade) ConversationsService() ConversationsService   { return nil }
func (f *fakeQuestValidationFacade) LootTablesService() LootTablesService         { return nil }
func (f *fakeQuestValidationFacade) ServerSettingsService() ServerSettingsService { return nil }
func (f *fakeQuestValidationFacade) QuestsService() QuestsService                 { return nil }
func (f *fakeQuestValidationFacade) SkillsService() SkillsService                 { return nil }
func (f *fakeQuestValidationFacade) CharacterTemplatesRepo() repository.CharacterTemplatesRepository {
	return nil
}
func (f *fakeQuestValidationFacade) GuestService() GuestService           { return nil }
func (f *fakeQuestValidationFacade) GuestStatsService() GuestStatsService { return nil }
func (f *fakeQuestValidationFacade) Runner() scripts.ScriptRunner         { return nil }

type fakeQuestsRepo struct {
	quests map[string]*quests.Quest
}

func (r *fakeQuestsRepo) FindAll() ([]*quests.Quest, error) {
	result := make([]*quests.Quest, 0, len(r.quests))
	for _, q := range r.quests {
		result = append(result, q)
	}
	return result, nil
}
func (r *fakeQuestsRepo) FindByID(id string) (*quests.Quest, error) {
	if q, ok := r.quests[id]; ok {
		return q, nil
	}
	return nil, errors.New("quest not found")
}
func (r *fakeQuestsRepo) FindByName(name string) ([]*quests.Quest, error) { return nil, nil }
func (r *fakeQuestsRepo) FindBySourceNPC(npcID string) ([]*quests.Quest, error) {
	var result []*quests.Quest
	for _, q := range r.quests {
		if q.Source.NPCID == npcID {
			result = append(result, q)
		}
	}
	return result, nil
}
func (r *fakeQuestsRepo) Store(quest *quests.Quest) (*quests.Quest, error) {
	r.quests[quest.ID] = quest
	return quest, nil
}
func (r *fakeQuestsRepo) Import(quest *quests.Quest) (*quests.Quest, error) {
	return r.Store(quest)
}
func (r *fakeQuestsRepo) Update(id string, quest *quests.Quest) error {
	r.quests[id] = quest
	return nil
}
func (r *fakeQuestsRepo) Delete(id string) error {
	delete(r.quests, id)
	return nil
}
func (r *fakeQuestsRepo) Drop() error { return nil }

type fakeQuestProgressRepo struct {
	progress []*quests.QuestProgress
}

func (r *fakeQuestProgressRepo) FindAll() ([]*quests.QuestProgress, error) { return r.progress, nil }
func (r *fakeQuestProgressRepo) FindByID(id string) (*quests.QuestProgress, error) {
	for _, progress := range r.progress {
		if progress.ID == id {
			return progress, nil
		}
	}
	return nil, errors.New("not found")
}
func (r *fakeQuestProgressRepo) FindByCharacterID(characterID string) ([]*quests.QuestProgress, error) {
	var result []*quests.QuestProgress
	for _, progress := range r.progress {
		if progress.CharacterID == characterID {
			result = append(result, progress)
		}
	}
	return result, nil
}
func (r *fakeQuestProgressRepo) FindByCharacterAndQuest(characterID, questID string) (*quests.QuestProgress, error) {
	for _, progress := range r.progress {
		if progress.CharacterID == characterID && progress.QuestID == questID {
			return progress, nil
		}
	}
	return nil, nil
}
func (r *fakeQuestProgressRepo) Store(progress *quests.QuestProgress) (*quests.QuestProgress, error) {
	r.progress = append(r.progress, progress)
	return progress, nil
}
func (r *fakeQuestProgressRepo) Import(progress *quests.QuestProgress) (*quests.QuestProgress, error) {
	return progress, nil
}
func (r *fakeQuestProgressRepo) Update(id string, progress *quests.QuestProgress) error {
	for i, existing := range r.progress {
		if existing.ID == id {
			r.progress[i] = progress
			return nil
		}
	}
	r.progress = append(r.progress, progress)
	return nil
}
func (r *fakeQuestProgressRepo) Delete(id string) error {
	for i, progress := range r.progress {
		if progress.ID == id {
			r.progress = append(r.progress[:i], r.progress[i+1:]...)
			return nil
		}
	}
	return nil
}

type fakeNPCsRepo struct {
	npcs map[string]*npc.NPC
}

func (r *fakeNPCsRepo) FindAll() ([]*npc.NPC, error) {
	result := make([]*npc.NPC, 0, len(r.npcs))
	for _, n := range r.npcs {
		result = append(result, n)
	}
	return result, nil
}
func (r *fakeNPCsRepo) FindByID(id string) (*npc.NPC, error) {
	if n, ok := r.npcs[id]; ok {
		return n, nil
	}
	return nil, errors.New("npc not found")
}
func (r *fakeNPCsRepo) FindByName(name string) ([]*npc.NPC, error)   { return nil, nil }
func (r *fakeNPCsRepo) FindByRoom(roomID string) ([]*npc.NPC, error) { return nil, nil }
func (r *fakeNPCsRepo) Store(n *npc.NPC) (*npc.NPC, error)           { r.npcs[n.ID] = n; return n, nil }
func (r *fakeNPCsRepo) Import(n *npc.NPC) (*npc.NPC, error)          { return r.Store(n) }
func (r *fakeNPCsRepo) Update(id string, n *npc.NPC) error           { r.npcs[id] = n; return nil }
func (r *fakeNPCsRepo) Delete(id string) error                       { delete(r.npcs, id); return nil }
func (r *fakeNPCsRepo) Drop() error                                  { return nil }

type fakeItemsRepo struct {
	items map[string]*items.Item
}

func (r *fakeItemsRepo) FindByID(id string) (*items.Item, error) {
	if item, ok := r.items[id]; ok {
		return item, nil
	}
	return nil, errors.New("item not found")
}
func (r *fakeItemsRepo) FindByName(name string) ([]*items.Item, error) { return nil, nil }
func (r *fakeItemsRepo) FindAll(query repository.ItemsQuery) ([]*items.Item, error) {
	result := make([]*items.Item, 0, len(r.items))
	for _, item := range r.items {
		result = append(result, item)
	}
	return result, nil
}
func (r *fakeItemsRepo) Update(id string, item *items.Item) error { r.items[id] = item; return nil }
func (r *fakeItemsRepo) Delete(id string) error                   { delete(r.items, id); return nil }
func (r *fakeItemsRepo) Store(item *items.Item) (*items.Item, error) {
	r.items[item.ID] = item
	return item, nil
}
func (r *fakeItemsRepo) Import(item *items.Item) (*items.Item, error) { return r.Store(item) }
func (r *fakeItemsRepo) Drop() error                                  { return nil }
func (r *fakeItemsRepo) FindAllTemplates(query repository.ItemsQuery) ([]*items.Item, error) {
	return r.FindAll(query)
}
func (r *fakeItemsRepo) FindAllInstances(query repository.ItemsQuery) ([]*items.Item, error) {
	return r.FindAll(query)
}
func (r *fakeItemsRepo) FindTemplateByName(name string) ([]*items.Item, error) { return nil, nil }
func (r *fakeItemsRepo) FindByTemplateID(templateID string) ([]*items.Item, error) {
	return nil, nil
}

type fakeRoomsRepo struct {
	rooms map[string]*rooms.Room
}

func (r *fakeRoomsRepo) Drop() error { return nil }
func (r *fakeRoomsRepo) FindByID(id string) (*rooms.Room, error) {
	if room, ok := r.rooms[id]; ok {
		return room, nil
	}
	return nil, errors.New("room not found")
}
func (r *fakeRoomsRepo) FindByName(name string) ([]*rooms.Room, error) { return nil, nil }
func (r *fakeRoomsRepo) FindAll() ([]*rooms.Room, error) {
	result := make([]*rooms.Room, 0, len(r.rooms))
	for _, room := range r.rooms {
		result = append(result, room)
	}
	return result, nil
}
func (r *fakeRoomsRepo) FindAllWithQuery(query repository.RoomsQuery) ([]*rooms.Room, error) {
	return r.FindAll()
}
func (r *fakeRoomsRepo) Update(id string, room *rooms.Room) error { r.rooms[id] = room; return nil }
func (r *fakeRoomsRepo) Delete(id string) error                   { delete(r.rooms, id); return nil }
func (r *fakeRoomsRepo) Store(room *rooms.Room) (*rooms.Room, error) {
	r.rooms[room.ID] = room
	return room, nil
}
func (r *fakeRoomsRepo) Import(room *rooms.Room) (*rooms.Room, error) { return r.Store(room) }

type fakeScriptsRepo struct {
	scripts map[string]*scripts.Script
}

func (r *fakeScriptsRepo) Drop() error { return nil }
func (r *fakeScriptsRepo) FindByID(id string) (*scripts.Script, error) {
	if script, ok := r.scripts[id]; ok {
		return script, nil
	}
	return nil, errors.New("script not found")
}
func (r *fakeScriptsRepo) FindByName(name string) ([]*scripts.Script, error) { return nil, nil }
func (r *fakeScriptsRepo) FindAll() ([]*scripts.Script, error) {
	result := make([]*scripts.Script, 0, len(r.scripts))
	for _, script := range r.scripts {
		result = append(result, script)
	}
	return result, nil
}
func (r *fakeScriptsRepo) Update(id string, script *scripts.Script) error {
	r.scripts[id] = script
	return nil
}
func (r *fakeScriptsRepo) Delete(id string) error { delete(r.scripts, id); return nil }
func (r *fakeScriptsRepo) Store(script *scripts.Script) (*scripts.Script, error) {
	r.scripts[script.ID] = script
	return script, nil
}
func (r *fakeScriptsRepo) Import(script *scripts.Script) (*scripts.Script, error) {
	return r.Store(script)
}

var _ CharactersService = (*fakeCharactersService)(nil)

type fakeCharactersService struct {
	characters map[string]*characters.Character
}

func (s *fakeCharactersService) Drop() error { return nil }
func (s *fakeCharactersService) FindByID(id string) (*characters.Character, error) {
	if character, ok := s.characters[id]; ok {
		return character, nil
	}
	return nil, errors.New("character not found")
}
func (s *fakeCharactersService) FindAllForUser(userID string) ([]*characters.Character, error) {
	return nil, nil
}
func (s *fakeCharactersService) FindByName(name string) ([]*characters.Character, error) {
	return nil, nil
}
func (s *fakeCharactersService) FindAll() ([]*characters.Character, error) { return nil, nil }
func (s *fakeCharactersService) Update(id string, character *characters.Character) error {
	s.characters[id] = character
	return nil
}
func (s *fakeCharactersService) Delete(id string) error { return nil }
func (s *fakeCharactersService) Store(character *characters.Character) (*characters.Character, error) {
	s.characters[character.ID] = character
	return character, nil
}
func (s *fakeCharactersService) Import(character *characters.Character) (*characters.Character, error) {
	s.characters[character.ID] = character
	return character, nil
}
func (s *fakeCharactersService) IsCharacterNameTaken(name string) bool { return false }
func (s *fakeCharactersService) GetCharacterTemplates() []*characters.CharacterTemplate {
	return nil
}
func (s *fakeCharactersService) Modify(id string, fn func(*characters.Character) error) error {
	ch, ok := s.characters[id]
	if !ok || ch == nil {
		return errors.New("character not found")
	}
	if err := fn(ch); err != nil {
		return err
	}
	s.characters[id] = ch
	return nil
}
func (s *fakeCharactersService) CreateNewCharacter(dto *dto.CreateCharacterDTO) (*characters.Character, error) {
	return nil, errors.New("unused")
}

var _ DialogsService = (*unusedDialogsService)(nil)

type unusedDialogsService struct{}

func (s *unusedDialogsService) FindAll() ([]*dialogs.Dialog, error) { return nil, nil }
func (s *unusedDialogsService) FindByID(id string) (*dialogs.Dialog, error) {
	return nil, errors.New("unused")
}
func (s *unusedDialogsService) FindByName(name string) (*dialogs.Dialog, error) {
	return nil, errors.New("unused")
}
func (s *unusedDialogsService) Store(dialog *dialogs.Dialog) (*dialogs.Dialog, error) {
	return dialog, nil
}
func (s *unusedDialogsService) Import(dialog *dialogs.Dialog) (*dialogs.Dialog, error) {
	return dialog, nil
}
func (s *unusedDialogsService) Update(id string, dialog *dialogs.Dialog) error { return nil }
func (s *unusedDialogsService) Delete(id string) error                         { return nil }
func (s *unusedDialogsService) Drop() error                                    { return nil }

var _ ConversationsService = (*unusedConversationsService)(nil)

type unusedConversationsService struct{}

func (s *unusedConversationsService) FindByID(id string) (*conversations.Conversation, error) {
	return nil, errors.New("unused")
}
func (s *unusedConversationsService) FindByCharacterAndTarget(characterID, targetID string) (*conversations.Conversation, error) {
	return nil, nil
}
func (s *unusedConversationsService) FindAllForCharacter(characterID string) ([]*conversations.Conversation, error) {
	return nil, nil
}
func (s *unusedConversationsService) Store(conv *conversations.Conversation) (*conversations.Conversation, error) {
	return conv, nil
}
func (s *unusedConversationsService) Update(id string, conv *conversations.Conversation) error {
	return nil
}
func (s *unusedConversationsService) Delete(id string) error { return nil }
func (s *unusedConversationsService) GetOrCreateConversation(characterID, targetID string, targetType conversations.TargetType, dialogID string) (*conversations.Conversation, error) {
	return nil, errors.New("unused")
}
func (s *unusedConversationsService) GetCurrentNode(conv *conversations.Conversation, dialog *dialogs.Dialog) *dialogs.Dialog {
	return nil
}
func (s *unusedConversationsService) GetFilteredOptions(conv *conversations.Conversation, node *dialogs.Dialog) []*dialogs.Dialog {
	return nil
}
func (s *unusedConversationsService) AdvanceConversation(conv *conversations.Conversation, nodeID string) error {
	return nil
}
func (s *unusedConversationsService) ResetConversation(conv *conversations.Conversation) error {
	return nil
}

var _ ServerSettingsService = (*unusedServerSettingsService)(nil)

type unusedServerSettingsService struct{}

func (s *unusedServerSettingsService) Get() (*settings.ServerSettings, error) { return nil, nil }
func (s *unusedServerSettingsService) Update(settings *settings.ServerSettings) error {
	return nil
}

var _ SkillsService = (*unusedSkillsService)(nil)

type unusedSkillsService struct{}

func (s *unusedSkillsService) Drop() error                       { return nil }
func (s *unusedSkillsService) FindAll() ([]*skills.Skill, error) { return nil, nil }
func (s *unusedSkillsService) FindByID(id string) (*skills.Skill, error) {
	return nil, errors.New("unused")
}
func (s *unusedSkillsService) Store(skill *skills.Skill) (*skills.Skill, error) { return skill, nil }
func (s *unusedSkillsService) Import(skill *skills.Skill) (*skills.Skill, error) {
	return skill, nil
}
func (s *unusedSkillsService) Update(id string, skill *skills.Skill) error { return nil }
func (s *unusedSkillsService) Delete(id string) error                      { return nil }
