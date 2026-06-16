package commands_test

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

func newTradeTestGame(t *testing.T) (*game.Game, service.Facade) {
	t.Helper()

	client, err := sqlite.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	return game.New(facade), facade
}

func drainTradeMessages(ch <-chan interface{}) []interface{} {
	var result []interface{}
	for {
		select {
		case msg := <-ch:
			result = append(result, msg)
		default:
			return result
		}
	}
}

func TestBuyStackableQuantityUsesOneInventorySlot(t *testing.T) {
	g, facade := newTradeTestGame(t)

	template := &items.Item{
		Entity:      &entities.Entity{ID: "potion-template"},
		IsTemplate:  true,
		Name:        "Healing Potion",
		Type:        items.ItemTypeConsumable,
		Slot:        items.ItemSlotInventory,
		Stackable:   true,
		Quantity:    1,
		MaxStack:    20,
		BasePrice:   2,
		Description: "Restores health.",
	}
	if _, err := facade.ItemsService().Import(template); err != nil {
		t.Fatalf("import item template: %v", err)
	}

	character, err := facade.CharactersService().Store(&characters.Character{
		Name:        "Buyer",
		BelongsUser: *traits.BelongsToUser("user-1"),
		CurrentRoom: traits.CurrentRoom{
			CurrentRoomID: "room-shop",
		},
		Gold: 100,
		Inventory: items.Inventory{
			Size:  1,
			Items: []*items.Item{},
		},
	})
	if err != nil {
		t.Fatalf("store character: %v", err)
	}

	merchant := &npc.NPC{
		Entity:        &entities.Entity{ID: "merchant-1"},
		Name:          "Shopkeep",
		CurrentRoom:   traits.CurrentRoom{CurrentRoomID: "room-shop"},
		MerchantTrait: npc.NewMerchantTrait(),
	}
	merchant.MerchantTrait.Inventory = []npc.MerchantItem{
		{
			ItemTemplateID: template.ID,
			Quantity:       10,
			MaxQuantity:    10,
		},
	}
	g.NPCManager.RegisterExistingNPC(merchant, "room-shop")

	msg := &messages.Message{
		FromUser:  &entities.User{Entity: &entities.Entity{ID: "user-1"}},
		Character: character,
		Data:      "buy Healing Potion 5",
	}
	if !(&commands.BuyCommand{}).Execute(g, msg) {
		t.Fatal("buy command did not handle message")
	}

	if len(character.Inventory.Items) != 1 {
		t.Fatalf("expected one inventory stack, got %d items", len(character.Inventory.Items))
	}
	if got := character.Inventory.Items[0].Quantity; got != 5 {
		t.Fatalf("expected purchased stack quantity 5, got %d", got)
	}
	if got := character.Gold; got != 90 {
		t.Fatalf("expected gold 90 after buying five at 2g, got %d", got)
	}
	if got := merchant.MerchantTrait.Inventory[0].Quantity; got != 5 {
		t.Fatalf("expected merchant stock 5 after selling five, got %d", got)
	}

	var sawConfirmation bool
	for _, out := range drainTradeMessages(g.SendMessage()) {
		if rsp, ok := out.(messages.MessageResponse); ok && strings.Contains(rsp.Message, "5x Healing Potion") {
			sawConfirmation = true
		}
	}
	if !sawConfirmation {
		t.Fatal("expected confirmation message for five purchased potions")
	}
}

func TestListRestocksMerchantInventory(t *testing.T) {
	g, facade := newTradeTestGame(t)

	template := &items.Item{
		Entity:     &entities.Entity{ID: "torch-template"},
		IsTemplate: true,
		Name:       "Torch",
		Type:       items.ItemTypeConsumable,
		Slot:       items.ItemSlotInventory,
		BasePrice:  1,
	}
	if _, err := facade.ItemsService().Import(template); err != nil {
		t.Fatalf("import item template: %v", err)
	}

	character := &characters.Character{
		Entity:      &entities.Entity{ID: "char-list"},
		Name:        "Browser",
		BelongsUser: *traits.BelongsToUser("user-1"),
		CurrentRoom: traits.CurrentRoom{
			CurrentRoomID: "room-shop",
		},
		Gold: 10,
	}
	merchant := &npc.NPC{
		Entity:      &entities.Entity{ID: "merchant-2"},
		Name:        "Provisioner",
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "room-shop"},
		MerchantTrait: &npc.MerchantTrait{
			MerchantType:   "general",
			BuyMultiplier:  1,
			SellMultiplier: 0.5,
			RestockMinutes: 1,
			LastRestock:    time.Now().Add(-2 * time.Minute),
			Inventory: []npc.MerchantItem{
				{
					ItemTemplateID: template.ID,
					Quantity:       0,
					MaxQuantity:    4,
				},
			},
		},
	}
	g.NPCManager.RegisterExistingNPC(merchant, "room-shop")

	msg := &messages.Message{
		FromUser:  &entities.User{Entity: &entities.Entity{ID: "user-1"}},
		Character: character,
		Data:      "list",
	}
	if !(&commands.ListCommand{}).Execute(g, msg) {
		t.Fatal("list command did not handle message")
	}
	if got := merchant.MerchantTrait.Inventory[0].Quantity; got != 4 {
		t.Fatalf("expected lazy restock quantity 4, got %d", got)
	}

	var sawStock bool
	for _, out := range drainTradeMessages(g.SendMessage()) {
		if rsp, ok := out.(messages.MessageResponse); ok && strings.Contains(rsp.Message, "4 in stock") {
			sawStock = true
		}
	}
	if !sawStock {
		t.Fatal("expected shop listing to show restocked quantity")
	}
}
