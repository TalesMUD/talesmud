package commands_test

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func TestTwoMerchantsUndercutSameItemTemplate(t *testing.T) {
	g, facade := newTradeTestGame(t)

	template := &items.Item{
		Entity:     &entities.Entity{ID: "ITM0012"},
		IsTemplate: true,
		Name:       "Weak Health Potion",
		Type:       items.ItemTypeConsumable,
		BasePrice:  8,
	}
	if _, err := facade.ItemsService().Import(template); err != nil {
		t.Fatalf("import item template: %v", err)
	}

	roomCheap := &rooms.Room{
		Entity:      &entities.Entity{ID: "R0218"},
		Name:        "Healer's Cottage",
		Description: "Herbs and remedies.",
	}
	roomPricey := &rooms.Room{
		Entity:      &entities.Entity{ID: "R0217"},
		Name:        "Coinsworth's Curios",
		Description: "Oddities and trinkets.",
	}
	if _, err := facade.RoomsService().Store(roomCheap); err != nil {
		t.Fatal(err)
	}
	if _, err := facade.RoomsService().Store(roomPricey); err != nil {
		t.Fatal(err)
	}

	elda := &npc.NPC{
		Entity:      &entities.Entity{ID: "NPC0014"},
		Name:        "Elda Rootwise",
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "R0218"},
		MerchantTrait: &npc.MerchantTrait{
			BuyMultiplier:  0.9,
			SellMultiplier: 0.7,
			Inventory: []npc.MerchantItem{{
				ItemTemplateID: template.ID,
				Quantity:       5,
				MaxQuantity:    5,
			}},
		},
	}
	darius := &npc.NPC{
		Entity:      &entities.Entity{ID: "NPC0013"},
		Name:        "Darius Coinsworth",
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "R0217"},
		MerchantTrait: &npc.MerchantTrait{
			BuyMultiplier:  1.6,
			SellMultiplier: 0.4,
			Inventory: []npc.MerchantItem{{
				ItemTemplateID: template.ID,
				Quantity:       5,
				MaxQuantity:    5,
			}},
		},
	}
	g.NPCManager.RegisterExistingNPC(elda, "R0218")
	g.NPCManager.RegisterExistingNPC(darius, "R0217")

	character, err := facade.CharactersService().Store(&characters.Character{
		Name:        "Shopper",
		BelongsUser: *traits.BelongsToUser("user-shop"),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "R0218"},
		Gold:        100,
	})
	if err != nil {
		t.Fatal(err)
	}

	listCheap := &messages.Message{
		FromUser:  &entities.User{Entity: &entities.Entity{ID: "user-shop"}},
		Character: character,
		Data:      "list",
	}
	if !(&commands.ListCommand{}).Execute(g, listCheap) {
		t.Fatal("list at cheap merchant failed")
	}
	var cheapPrice string
	for _, out := range drainTradeMessages(g.SendMessage()) {
		if rsp, ok := out.(messages.MessageResponse); ok && strings.Contains(rsp.Message, "Weak Health Potion") {
			cheapPrice = rsp.Message
		}
	}
	if !strings.Contains(cheapPrice, "7 gold") {
		t.Fatalf("expected Elda buy price 7 gold (0.9x8), got:\n%s", cheapPrice)
	}

	character.CurrentRoomID = "R0217"
	listPricey := &messages.Message{
		FromUser:  &entities.User{Entity: &entities.Entity{ID: "user-shop"}},
		Character: character,
		Data:      "list",
	}
	if !(&commands.ListCommand{}).Execute(g, listPricey) {
		t.Fatal("list at pricey merchant failed")
	}
	var priceyPrice string
	for _, out := range drainTradeMessages(g.SendMessage()) {
		if rsp, ok := out.(messages.MessageResponse); ok && strings.Contains(rsp.Message, "Weak Health Potion") {
			priceyPrice = rsp.Message
		}
	}
	if !strings.Contains(priceyPrice, "12 gold") {
		t.Fatalf("expected Darius buy price 12 gold (1.6x8), got:\n%s", priceyPrice)
	}
}
