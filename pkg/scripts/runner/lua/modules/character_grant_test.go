package modules

import (
	"fmt"
	"path/filepath"
	"testing"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/scripts"
	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
	"github.com/talesmud/talesmud/pkg/service"
)

func TestLuaGrantClearGearAndEquip(t *testing.T) {
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "grant.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	created, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "c-grant"},
		Name:             "Grant",
		BelongsUser:      *traits.BelongsToUser("user-grant"),
		Level:            2,
		MaxHitPoints:     20,
		CurrentHitPoints: 18,
		Inventory: items.Inventory{
			Size: 10,
			Items: []*items.Item{
				{Entity: &entities.Entity{ID: "potion"}, Name: "Vial", Type: items.ItemTypeConsumable},
				{Entity: &entities.Entity{ID: "spare"}, Name: "Spare", Type: items.ItemTypeWeapon, Slot: items.ItemSlotMainHand},
			},
		},
		EquippedItems: map[items.ItemSlot]*items.Item{
			items.ItemSlotMainHand: {Entity: &entities.Entity{ID: "worn-w"}, Name: "Stick", Type: items.ItemTypeWeapon, Slot: items.ItemSlotMainHand},
			items.ItemSlotChest:    {Entity: &entities.Entity{ID: "worn-a"}, Name: "Coat", Type: items.ItemTypeArmor, Slot: items.ItemSlotChest},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := facade.ItemsService().Import(&items.Item{
		Entity:     &entities.Entity{ID: "ITM_STICK"},
		Name:       "Ash Walking Stick",
		Type:       items.ItemTypeWeapon,
		Slot:       items.ItemSlotMainHand,
		IsTemplate: true,
		Attributes: map[string]interface{}{"damage": 2},
	}); err != nil {
		t.Fatal(err)
	}
	id := created.ID
	runner := luarunner.NewLuaRunner()
	RegisterAllModules(runner)
	runner.SetServices(facade, nil)
	defer runner.Shutdown()

	result := runner.RunWithResult(scripts.Script{
		Name:     "grant",
		Language: scripts.ScriptLanguageLua,
		Code: fmt.Sprintf(`
local a = tales.characters.grant(%q, "attack", 3)
local b = tales.characters.grant(%q, "attack", 50)
local c = tales.characters.grant(%q, "defense", 2)
local d = tales.characters.grant(%q, "maxHP", 4)
local e = tales.characters.grant(%q, "nope", 1)
local f = tales.characters.grant(%q, "attack", 0)
local removed = tales.game.clearGear(%q)
local wore = tales.game.equipFromTemplate(%q, "ITM_STICK")
local missing = tales.game.equipFromTemplate(%q, "no-such")
return {a, b, c, d, e, f, removed, wore, missing}
`, id, id, id, id, id, id, id, id, id),
	}, scripts.NewScriptContext())
	if result == nil || !result.Success {
		t.Fatalf("script: %+v", result)
	}
	row, ok := result.Result.([]interface{})
	if !ok || len(row) != 9 {
		t.Fatalf("result %#v", result.Result)
	}
	want := []interface{}{true, true, true, true, false, false, float64(3), true, false}
	for i := range want {
		if row[i] != want[i] {
			t.Fatalf("index %d = %#v want %#v full %#v", i, row[i], want[i], row)
		}
	}
	stored, err := facade.CharactersService().FindByID(id)
	if err != nil {
		t.Fatal(err)
	}
	if stored.BonusAttack != 13 || stored.BonusDefense != 2 {
		t.Fatalf("bonus atk %d def %d", stored.BonusAttack, stored.BonusDefense)
	}
	if stored.MaxHitPoints != 24 || stored.CurrentHitPoints != 22 {
		t.Fatalf("hp %d/%d", stored.CurrentHitPoints, stored.MaxHitPoints)
	}
	if len(stored.Inventory.Items) != 1 || stored.Inventory.Items[0].Name != "Vial" {
		t.Fatalf("bag %#v", stored.Inventory.Items)
	}
	worn := stored.EquippedItems[items.ItemSlotMainHand]
	if worn == nil || worn.TemplateID != "ITM_STICK" || worn.IsTemplate {
		t.Fatalf("worn %+v", worn)
	}
	if stored.EquippedItems[items.ItemSlotChest] != nil {
		t.Fatal("chest piece should have been cleared")
	}
}
