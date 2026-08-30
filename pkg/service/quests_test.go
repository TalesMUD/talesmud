package service

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/repository"
)

func newTestFacade(t *testing.T) Facade {
	t.Helper()

	client, err := sqlite.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	return NewFacade(repository.NewSQLiteFactory(client), nil)
}

func TestGrantAutoQuestsAcceptsZ00TutorialQuests(t *testing.T) {
	facade := newTestFacade(t)
	exits := rooms.Exits{}
	chars := rooms.Characters{}
	roomsByID := map[string]string{"R0006": "Z00_catacombs_intro", "R0102": "Z01_meadows_forest_path"}
	for id, area := range roomsByID {
		if _, err := facade.RoomsService().Import(&rooms.Room{
			Entity: &entities.Entity{ID: id}, Name: id, Area: area, Exits: &exits, Characters: &chars,
		}); err != nil {
			t.Fatalf("import %s: %v", id, err)
		}
	}
	character, err := facade.CharactersService().Store(&characters.Character{
		Name:        "Wanderer",
		BelongsUser: *traits.BelongsToUser("user-1"),
	})
	if err != nil {
		t.Fatalf("store character: %v", err)
	}
	storeAuto := func(name, target string) *quests.Quest {
		t.Helper()
		q, err := facade.QuestsService().Store(&quests.Quest{
			Name:        name,
			Description: name,
			Source:      quests.QuestSource{Type: "auto"},
			Objectives: []quests.Objective{{
				ID: "obj", Type: quests.ObjectiveVisit, Description: "go", TargetID: target, Amount: 1,
			}},
		})
		if err != nil {
			t.Fatalf("store %s: %v", name, err)
		}
		return q
	}
	z00 := storeAuto("Find the Exit", "R0006")
	z01 := storeAuto("A Breath of Fresh Air", "R0102")

	granted := facade.QuestsService().GrantAutoQuests(character.ID, "Z00_catacombs_intro")
	if granted != 1 {
		t.Fatalf("expected 1 Z00 auto quest, granted %d", granted)
	}
	if p, _ := facade.QuestsService().GetProgress(character.ID, z00.ID); p == nil {
		t.Fatal("Z00 auto quest not granted")
	}
	if p, _ := facade.QuestsService().GetProgress(character.ID, z01.ID); p != nil {
		t.Fatal("Z01 auto quest should not grant in Z00")
	}

	grantedZ01 := facade.QuestsService().GrantAutoQuests(character.ID, "Z01_meadows_forest_path")
	if grantedZ01 != 1 {
		t.Fatalf("expected 1 Z01 auto quest, granted %d", grantedZ01)
	}
	if p, _ := facade.QuestsService().GetProgress(character.ID, z01.ID); p == nil {
		t.Fatal("Z01 auto quest not granted on meadows enter")
	}
}

func TestAcceptQuestPrefillsCollectObjectiveWithStackQuantity(t *testing.T) {
	facade := newTestFacade(t)

	templateID := "herb-template"
	if _, err := facade.ItemsService().Import(&items.Item{
		Entity:     &entities.Entity{ID: templateID},
		Name:       "Moon Herb",
		IsTemplate: true,
		Stackable:  true,
	}); err != nil {
		t.Fatalf("store item template: %v", err)
	}

	character := &characters.Character{
		Entity:      &entities.Entity{ID: "char-collect"},
		Name:        "Collector",
		BelongsUser: *traits.BelongsToUser("user-1"),
		Inventory: items.Inventory{
			Size: 10,
			Items: []*items.Item{
				{
					Entity:     &entities.Entity{ID: "herb-stack"},
					TemplateID: templateID,
					Name:       "Moon Herb",
					Stackable:  true,
					Quantity:   3,
				},
			},
		},
	}
	if _, err := facade.CharactersService().Store(character); err != nil {
		t.Fatalf("store character: %v", err)
	}

	quest := &quests.Quest{
		Entity:      &entities.Entity{ID: "quest-collect"},
		Name:        "Gather Herbs",
		Description: "Gather herbs already in your bag.",
		Source:      quests.QuestSource{Type: "auto"},
		Objectives: []quests.Objective{
			{
				ID:          "collect-herbs",
				Type:        quests.ObjectiveCollect,
				Description: "Collect 5 Moon Herbs",
				TargetID:    templateID,
				Amount:      5,
			},
		},
	}
	if _, err := facade.QuestsService().Store(quest); err != nil {
		t.Fatalf("store quest: %v", err)
	}

	progress, err := facade.QuestsService().AcceptQuest(character.ID, quest.ID)
	if err != nil {
		t.Fatalf("accept quest: %v", err)
	}
	if len(progress.Objectives) != 1 {
		t.Fatalf("expected one objective, got %d", len(progress.Objectives))
	}
	if got := progress.Objectives[0].Current; got != 3 {
		t.Fatalf("expected stack quantity progress 3, got %d", got)
	}
	if progress.Objectives[0].Completed {
		t.Fatal("objective should not be complete before reaching required quantity")
	}
}

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
				Source:      quests.QuestSource{Type: "npc", NPCID: "npc-1"},
				Objectives:  []quests.Objective{{ID: "kill", Type: quests.ObjectiveKill, Description: "Kill one", TargetID: "npc-1", Amount: 1}},
			},
			wantErr: "name is required",
		},
		{
			name: "npc source without npc id",
			quest: &quests.Quest{
				Name:        "Broken Source",
				Description: "Missing NPC source",
				Source:      quests.QuestSource{Type: "npc"},
				Objectives:  []quests.Objective{{ID: "kill", Type: quests.ObjectiveKill, Description: "Kill one", TargetID: "npc-1", Amount: 1}},
			},
			wantErr: "source.npcId is required",
		},
		{
			name: "duplicate objective id",
			quest: &quests.Quest{
				Name:        "Duplicate",
				Description: "Duplicate objective IDs",
				Source:      quests.QuestSource{Type: "auto"},
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
				Name:        "Delivery",
				Description: "Missing delivery item",
				Source:      quests.QuestSource{Type: "auto"},
				Objectives:  []quests.Objective{{ID: "deliver", Type: quests.ObjectiveDeliver, Description: "Deliver item", DeliverToNPCID: "npc-1"}},
			},
			wantErr: "objectives[0].targetId is required",
		},
		{
			name: "self prerequisite",
			quest: &quests.Quest{
				Entity:           &entities.Entity{ID: "quest-self"},
				Name:             "Loop",
				Description:      "Self prereq",
				Source:           quests.QuestSource{Type: "auto"},
				RequiredQuestIDs: []string{"quest-self"},
				Objectives:       []quests.Objective{{ID: "visit", Type: quests.ObjectiveVisit, Description: "Visit room", TargetID: "room-1"}},
			},
			wantErr: "requiredQuestIds[0] cannot reference this quest",
		},
		{
			name: "negative reward",
			quest: &quests.Quest{
				Name:        "Bad Reward",
				Description: "Negative reward",
				Source:      quests.QuestSource{Type: "auto"},
				Rewards:     quests.Reward{XP: -1},
				Objectives:  []quests.Objective{{ID: "visit", Type: quests.ObjectiveVisit, Description: "Visit room", TargetID: "room-1"}},
			},
			wantErr: "rewards.xp cannot be negative",
		},
		{
			name: "missing referenced npc",
			quest: &quests.Quest{
				Name:        "Missing NPC",
				Description: "References an NPC that is not stored",
				Source:      quests.QuestSource{Type: "npc", NPCID: "missing-npc"},
				Objectives:  []quests.Objective{{ID: "talk", Type: quests.ObjectiveTalk, Description: "Talk to missing NPC", TargetID: "missing-npc"}},
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
		Name:        "Rat Problem",
		Description: "Clear the cellar.",
		Source:      quests.QuestSource{Type: "npc", NPCID: "npc-1"},
		Objectives:  []quests.Objective{{ID: "kill-rat", Type: quests.ObjectiveKill, Description: "Kill 3 rats", TargetID: "rat-template", Amount: 3}},
		Rewards:     quests.Reward{XP: 25, Gold: 4},
	}
	if _, err := facade.QuestsService().Store(quest); err != nil {
		t.Fatalf("store valid quest: %v", err)
	}
}

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
		Entity:      &entities.Entity{ID: "char-deliver"},
		Name:        "Courier",
		BelongsUser: *traits.BelongsToUser("user-1"),
		Inventory: items.Inventory{
			Size: 10,
			Items: []*items.Item{
				{Entity: &entities.Entity{ID: "letter-stack"}, TemplateID: templateID, Name: "Sealed Letter", Stackable: true, Quantity: 2},
			},
		},
	}
	if _, err := facade.CharactersService().Store(character); err != nil {
		t.Fatalf("store character: %v", err)
	}

	quest := &quests.Quest{
		Entity:      &entities.Entity{ID: "quest-deliver"},
		Name:        "Courier Run",
		Description: "Deliver the letters.",
		Source:      quests.QuestSource{Type: "auto"},
		Objectives: []quests.Objective{
			{ID: "deliver-letters", Type: quests.ObjectiveDeliver, Description: "Deliver 2 letters", TargetID: templateID, DeliverToNPCID: "captain", Amount: 2},
		},
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
		Entity:      &entities.Entity{ID: "char-empty"},
		Name:        "Empty Courier",
		BelongsUser: *traits.BelongsToUser("user-1"),
		Inventory:   items.Inventory{Size: 10},
	}
	if _, err := facade.CharactersService().Store(character); err != nil {
		t.Fatalf("store character: %v", err)
	}

	quest := &quests.Quest{
		Entity:      &entities.Entity{ID: "quest-missing-delivery"},
		Name:        "Missing Item",
		Description: "Deliver an item.",
		Source:      quests.QuestSource{Type: "auto"},
		Objectives: []quests.Objective{
			{ID: "deliver", Type: quests.ObjectiveDeliver, Description: "Deliver item", TargetID: "item-template", DeliverToNPCID: "captain", Amount: 1},
		},
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
