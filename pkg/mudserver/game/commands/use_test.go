package commands_test

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func TestUseConsumableStackPersistsReducedItemQuantity(t *testing.T) {
	g, facade := newTradeTestGame(t)

	potion := &items.Item{
		Entity:     &entities.Entity{ID: "potion-stack"},
		Name:       "Healing Potion",
		Type:       items.ItemTypeConsumable,
		Stackable:  true,
		Quantity:   3,
		MaxStack:   10,
		Consumable: true,
		Attributes: map[string]interface{}{
			"healthRestore": 5,
		},
	}
	if _, err := facade.ItemsService().Import(potion); err != nil {
		t.Fatalf("import potion: %v", err)
	}

	character, err := facade.CharactersService().Store(&characters.Character{
		Name:             "Potion Tester",
		BelongsUser:      *traits.BelongsToUser("user-potion"),
		CurrentHitPoints: 5,
		MaxHitPoints:     10,
		Inventory: items.Inventory{
			Size:  5,
			Items: []*items.Item{potion},
		},
	})
	if err != nil {
		t.Fatalf("store character: %v", err)
	}

	msg := &messages.Message{
		FromUser:  &entities.User{Entity: &entities.Entity{ID: "user-potion"}},
		Character: character,
		Data:      "use Healing Potion",
	}
	if !(&commands.UseCommand{}).Execute(g, msg) {
		t.Fatal("use command did not handle consumable")
	}

	if got := character.Inventory.Items[0].Quantity; got != 2 {
		t.Fatalf("expected character inventory quantity 2, got %d", got)
	}

	stored, err := facade.ItemsService().FindByID(potion.ID)
	if err != nil {
		t.Fatalf("find stored potion: %v", err)
	}
	if got := stored.Quantity; got != 2 {
		t.Fatalf("expected item repository quantity 2, got %d", got)
	}

	var sawUseMessage bool
	for _, out := range drainTradeMessages(g.SendMessage()) {
		if rsp, ok := out.(messages.MessageResponse); ok && strings.Contains(rsp.Message, "restore 5 health") {
			sawUseMessage = true
		}
	}
	if !sawUseMessage {
		t.Fatal("expected health restore message")
	}
}

func TestUseRefusesConsumableWithNoEffect(t *testing.T) {
	g, facade := newTradeTestGame(t)

	torch := &items.Item{
		Entity:     &entities.Entity{ID: "torch-noop"},
		Name:       "Dusty Torch",
		Type:       items.ItemTypeCollectible,
		SubType:    "light_source",
		Consumable: true,
		Quantity:   1,
	}
	if _, err := facade.ItemsService().Import(torch); err != nil {
		t.Fatalf("import torch: %v", err)
	}
	character, err := facade.CharactersService().Store(&characters.Character{
		Name:        "Torch Tester",
		BelongsUser: *traits.BelongsToUser("user-torch"),
		Inventory: items.Inventory{
			Size:  5,
			Items: []*items.Item{torch},
		},
	})
	if err != nil {
		t.Fatalf("store character: %v", err)
	}

	msg := &messages.Message{
		FromUser:  &entities.User{Entity: &entities.Entity{ID: "user-torch"}},
		Character: character,
		Data:      "use Dusty Torch",
	}
	if !(&commands.UseCommand{}).Execute(g, msg) {
		t.Fatal("use should handle command")
	}

	updated, err := facade.CharactersService().FindByID(character.ID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if updated.Inventory.FindItemByName("Dusty Torch") == nil {
		t.Fatal("no-op use must not delete the torch")
	}

	var sawGeneric, sawRefuse bool
	for _, out := range drainTradeMessages(g.SendMessage()) {
		if rsp, ok := out.(messages.MessageResponse); ok {
			if rsp.Message == "You use Dusty Torch." {
				sawGeneric = true
			}
			if strings.Contains(rsp.Message, "can't use") || strings.Contains(rsp.Message, "Nothing happens") {
				sawRefuse = true
			}
		}
	}
	if sawGeneric {
		t.Fatal("must not print generic You use X on no-op")
	}
	if !sawRefuse {
		t.Fatal("expected refuse message for consumable with no effect")
	}
}

func TestUseFlintOnTorchSetsTorchLit(t *testing.T) {
	g, facade := newTradeTestGame(t)

	flint := &items.Item{
		Entity:  &entities.Entity{ID: "flint-1"},
		Name:    "Flint and Steel",
		Type:    items.ItemTypeCollectible,
		SubType: "tool",
		Tags:    []string{"tool", "utility"},
	}
	torch := &items.Item{
		Entity:  &entities.Entity{ID: "torch-1"},
		Name:    "Dusty Torch",
		Type:    items.ItemTypeCollectible,
		SubType: "light_source",
		Tags:    []string{"light"},
	}
	if _, err := facade.ItemsService().Import(flint); err != nil {
		t.Fatal(err)
	}
	if _, err := facade.ItemsService().Import(torch); err != nil {
		t.Fatal(err)
	}
	character, err := facade.CharactersService().Store(&characters.Character{
		Name:        "Firestarter",
		BelongsUser: *traits.BelongsToUser("user-fire"),
		Inventory: items.Inventory{
			Size:  5,
			Items: []*items.Item{flint, torch},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	msg := &messages.Message{
		FromUser:  &entities.User{Entity: &entities.Entity{ID: "user-fire"}},
		Character: character,
		Data:      "use flint on torch",
	}
	if !(&commands.UseCommand{}).Execute(g, msg) {
		t.Fatal("use on target should handle")
	}

	updated, _ := facade.CharactersService().FindByID(character.ID)
	if updated.Flags == nil || updated.Flags["torch_lit"] != true {
		t.Fatalf("expected torch_lit flag, got %#v", updated.Flags)
	}
	storedTorch, _ := facade.ItemsService().FindByID(torch.ID)
	if storedTorch.Attributes == nil || storedTorch.Attributes["lit"] != true {
		t.Fatalf("expected lit attribute on torch, got %#v", storedTorch.Attributes)
	}

	var sawLight bool
	for _, out := range drainTradeMessages(g.SendMessage()) {
		if rsp, ok := out.(messages.MessageResponse); ok && strings.Contains(rsp.Message, "light") {
			sawLight = true
		}
	}
	if !sawLight {
		t.Fatal("expected lighting message")
	}
}
