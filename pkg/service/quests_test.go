package service

import (
	"path/filepath"
	"testing"

	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/quests"
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

func TestAcceptQuestPrefillsCollectObjectiveWithStackQuantity(t *testing.T) {
	facade := newTestFacade(t)

	templateID := "herb-template"
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
		Objectives: []quests.Objective{
			{
				ID:       "collect-herbs",
				Type:     quests.ObjectiveCollect,
				TargetID: templateID,
				Amount:   5,
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
