package validation

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

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

func TestValidateRoomReportsBrokenExitAndActionScript(t *testing.T) {
	exits := rooms.Exits{{Name: "north", Target: "missing-room"}}
	actions := rooms.Actions{{Name: "pull lever", Type: rooms.RoomActionTypeScript, ScriptId: "missing-script"}}
	room := &rooms.Room{Entity: &entities.Entity{ID: "room-a"}, Name: "Room A", Exits: &exits, Actions: &actions}

	result := ValidateRoom(room, NewWorldSnapshot())

	if result.Errors != 2 {
		t.Fatalf("Errors = %d, want 2: %#v", result.Errors, result.Issues)
	}
}

func TestValidateNPCReportsBrokenMerchantStockAndDialog(t *testing.T) {
	n := &npc.NPC{
		Entity:  &entities.Entity{ID: "npc-a"},
		Name:    "Merchant",
		DialogID: "missing-dialog",
		MerchantTrait: &npc.MerchantTrait{Inventory: []npc.MerchantItem{
			{ItemTemplateID: "missing-item", Quantity: 1, MaxQuantity: 1},
		}},
	}

	result := ValidateNPC(n, NewWorldSnapshot())

	if result.Errors != 2 {
		t.Fatalf("Errors = %d, want 2: %#v", result.Errors, result.Issues)
	}
}

func TestValidateLootTableReportsBrokenEntryAndInvalidRange(t *testing.T) {
	table := &items.LootTable{
		Entity: &entities.Entity{ID: "loot-a"},
		Name:   "Loot",
		Entries: []items.LootEntry{
			{ItemTemplateID: "missing-item", DropChance: 2, MinQuantity: 5, MaxQuantity: 1},
		},
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

func TestValidateQuestReportsBrokenRewardAndSource(t *testing.T) {
	quest := &quests.Quest{
		Entity:  &entities.Entity{ID: "quest-a"},
		Name:    "Quest A",
		Source:  quests.QuestSource{Type: "npc", NPCID: "missing-npc"},
		Rewards: quests.Reward{ItemTemplateIDs: []string{"missing-item"}},
	}

	result := ValidateQuest(quest, NewWorldSnapshot())

	if result.Errors != 2 {
		t.Fatalf("Errors = %d, want 2: %#v", result.Errors, result.Issues)
	}
}
