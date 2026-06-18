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
