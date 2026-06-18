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

type fakeQuestValidationFacade struct {
	npcs    *fakeNPCsRepo
	items   *fakeItemsRepo
	rooms   *fakeRoomsRepo
	scripts *fakeScriptsRepo
}

func (f *fakeQuestValidationFacade) CharactersService() CharactersService { return nil }
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

type fakeQuestProgressRepo struct{}

func (r *fakeQuestProgressRepo) FindAll() ([]*quests.QuestProgress, error) { return nil, nil }
func (r *fakeQuestProgressRepo) FindByID(id string) (*quests.QuestProgress, error) {
	return nil, errors.New("not found")
}
func (r *fakeQuestProgressRepo) FindByCharacterID(characterID string) ([]*quests.QuestProgress, error) {
	return nil, nil
}
func (r *fakeQuestProgressRepo) FindByCharacterAndQuest(characterID, questID string) (*quests.QuestProgress, error) {
	return nil, nil
}
func (r *fakeQuestProgressRepo) Store(progress *quests.QuestProgress) (*quests.QuestProgress, error) {
	return progress, nil
}
func (r *fakeQuestProgressRepo) Import(progress *quests.QuestProgress) (*quests.QuestProgress, error) {
	return progress, nil
}
func (r *fakeQuestProgressRepo) Update(id string, progress *quests.QuestProgress) error { return nil }
func (r *fakeQuestProgressRepo) Delete(id string) error                                 { return nil }

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

var _ CharactersService = (*unusedCharactersService)(nil)

type unusedCharactersService struct{}

func (s *unusedCharactersService) Drop() error { return nil }
func (s *unusedCharactersService) FindByID(id string) (*characters.Character, error) {
	return nil, errors.New("unused")
}
func (s *unusedCharactersService) FindAllForUser(userID string) ([]*characters.Character, error) {
	return nil, nil
}
func (s *unusedCharactersService) FindByName(name string) ([]*characters.Character, error) {
	return nil, nil
}
func (s *unusedCharactersService) FindAll() ([]*characters.Character, error) { return nil, nil }
func (s *unusedCharactersService) Update(id string, character *characters.Character) error {
	return nil
}
func (s *unusedCharactersService) Delete(id string) error { return nil }
func (s *unusedCharactersService) Store(character *characters.Character) (*characters.Character, error) {
	return character, nil
}
func (s *unusedCharactersService) Import(character *characters.Character) (*characters.Character, error) {
	return character, nil
}
func (s *unusedCharactersService) IsCharacterNameTaken(name string) bool { return false }
func (s *unusedCharactersService) GetCharacterTemplates() []*characters.CharacterTemplate {
	return nil
}
func (s *unusedCharactersService) CreateNewCharacter(dto *dto.CreateCharacterDTO) (*characters.Character, error) {
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
