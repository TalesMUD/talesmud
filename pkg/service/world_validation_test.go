package service

import (
	"path/filepath"
	"testing"

	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/repository"
)

func worldValidationTestFacade(t *testing.T) Facade {
	t.Helper()

	client, err := sqlite.Open(filepath.Join(t.TempDir(), "world-validation.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	return NewFacade(repository.NewSQLiteFactory(client), nil)
}

func TestWorldValidationReportsBrokenReferences(t *testing.T) {
	facade := worldValidationTestFacade(t)

	validTemplate := &items.Item{
		Entity:     &entities.Entity{ID: "item-template-valid"},
		IsTemplate: true,
		Name:       "Valid Template",
	}
	if _, err := facade.ItemsService().Import(validTemplate); err != nil {
		t.Fatalf("store valid template: %v", err)
	}

	roomExits := rooms.Exits{{Name: "north", Target: "missing-room"}}
	roomItems := rooms.Items{"missing-item"}
	room := &rooms.Room{
		Entity: &entities.Entity{ID: "room-1"},
		Name:   "Room",
		Exits:  &roomExits,
		Items:  &roomItems,
	}
	if _, err := facade.RoomsService().Import(room); err != nil {
		t.Fatalf("store room: %v", err)
	}

	brokenNPC := &npc.NPC{
		Entity:       &entities.Entity{ID: "npc-1"},
		Name:         "Broken NPC",
		DialogID:     "missing-dialog",
		PatrolPath:   []string{"missing-patrol-room"},
		MaxHitPoints: 0,
		EnemyTrait: &npc.EnemyTrait{
			LootTableID:    "missing-loot-table",
			GuaranteedLoot: []string{"missing-guaranteed-template"},
			GoldDrop:       npc.Range{Min: 5, Max: 1},
		},
		MerchantTrait: &npc.MerchantTrait{
			Inventory: []npc.MerchantItem{{ItemTemplateID: "missing-merchant-template", Quantity: -2}},
		},
	}
	if _, err := facade.NPCsService().Import(brokenNPC); err != nil {
		t.Fatalf("store npc: %v", err)
	}

	if _, err := facade.NPCSpawnersService().Import(&npc.NPCSpawner{
		Entity:       &entities.Entity{ID: "spawner-1"},
		TemplateID:   "missing-npc-template",
		RoomID:       "missing-spawner-room",
		MaxInstances: -1,
	}); err != nil {
		t.Fatalf("store spawner: %v", err)
	}

	if _, err := facade.LootTablesService().Import(&items.LootTable{
		Entity: &entities.Entity{ID: "loot-1"},
		Name:   "Broken Loot",
		Entries: []items.LootEntry{{
			ItemTemplateID: "missing-loot-template",
			DropChance:     1.5,
			MinQuantity:    3,
			MaxQuantity:    1,
		}},
	}); err != nil {
		t.Fatalf("store loot table: %v", err)
	}

	if _, err := facade.CharacterTemplatesRepo().Import(&characters.CharacterTemplate{
		Entity: &entities.Entity{ID: "character-template-1"},
		Name:   "Broken Starter",
		StartingItems: []characters.StartingItem{{
			Slot:           items.ItemSlotInventory,
			ItemTemplateID: "missing-starter-template",
		}},
	}); err != nil {
		t.Fatalf("store character template: %v", err)
	}

	storedQuest, err := facade.QuestsService().Store(&quests.Quest{
		Entity: &entities.Entity{ID: "quest-1"},
		Name:   "Broken Quest",
		Source: quests.QuestSource{Type: "npc", NPCID: "missing-source-npc"},
		Objectives: []quests.Objective{{
			ID:       "visit",
			Type:     quests.ObjectiveVisit,
			TargetID: "missing-visit-room",
		}},
		Rewards:          quests.Reward{ItemTemplateIDs: []string{"missing-reward-template"}},
		RequiredQuestIDs: []string{"missing-prereq-quest"},
	})
	if err != nil {
		t.Fatalf("store quest: %v", err)
	}

	report, err := NewWorldValidationService(facade).Validate()
	if err != nil {
		t.Fatalf("validate world: %v", err)
	}

	for _, want := range []struct {
		system     string
		entityType string
		entityID   string
		field      string
	}{
		{"player-session", "room", "room-1", "exits[0].target"},
		{"item", "room", "room-1", "items[0]"},
		{"npc", "npc", "npc-1", "dialogID"},
		{"npc", "npc", "npc-1", "patrolPath[0]"},
		{"item", "npc", "npc-1", "merchantTrait.inventory[0].itemTemplateId"},
		{"item", "npc", "npc-1", "enemyTrait.lootTableId"},
		{"npc", "spawner", "spawner-1", "roomId"},
		{"item", "lootTable", "loot-1", "entries[0].itemTemplateId"},
		{"quest", "quest", storedQuest.ID, "source.npcId"},
		{"quest", "quest", storedQuest.ID, "objectives[0].targetId"},
		{"quest", "quest", storedQuest.ID, "rewards.itemTemplateIds[0]"},
		{"quest", "quest", storedQuest.ID, "requiredQuestIds[0]"},
		{"creator", "characterTemplate", "character-template-1", "startingItems[0].itemTemplateId"},
	} {
		if !hasValidationIssue(report, want.system, want.entityType, want.entityID, want.field) {
			t.Fatalf("expected validation issue for %#v in report %#v", want, report.Issues)
		}
	}
}

func TestWorldValidationAcceptsMinimalValidWorld(t *testing.T) {
	facade := worldValidationTestFacade(t)

	template := &items.Item{
		Entity:     &entities.Entity{ID: "item-template-valid"},
		IsTemplate: true,
		Name:       "Valid Template",
	}
	if _, err := facade.ItemsService().Import(template); err != nil {
		t.Fatalf("store valid template: %v", err)
	}
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity: &entities.Entity{ID: "room-1"},
		Name:   "Room",
	}); err != nil {
		t.Fatalf("store room: %v", err)
	}
	if _, err := facade.NPCsService().Import(&npc.NPC{
		Entity: &entities.Entity{ID: "npc-1"},
		Name:   "NPC",
	}); err != nil {
		t.Fatalf("store npc: %v", err)
	}
	if _, err := facade.QuestsService().Store(&quests.Quest{
		Entity: &entities.Entity{ID: "quest-1"},
		Name:   "Quest",
		Source: quests.QuestSource{Type: "npc", NPCID: "npc-1"},
		Objectives: []quests.Objective{{
			ID:       "visit",
			Type:     quests.ObjectiveVisit,
			TargetID: "room-1",
			Amount:   1,
		}},
		Rewards: quests.Reward{ItemTemplateIDs: []string{"item-template-valid"}},
	}); err != nil {
		t.Fatalf("store quest: %v", err)
	}

	report, err := NewWorldValidationService(facade).Validate()
	if err != nil {
		t.Fatalf("validate world: %v", err)
	}
	if report.ErrorCount != 0 {
		t.Fatalf("expected no validation errors, got %#v", report.Issues)
	}
}

func hasValidationIssue(report *ValidationReport, system, entityType, entityID, field string) bool {
	for _, issue := range report.Issues {
		if issue.System == system &&
			issue.EntityType == entityType &&
			issue.EntityID == entityID &&
			issue.Field == field {
			return true
		}
	}
	return false
}
