package commands_test

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/mudserver/game/util"
)

func TestLookHidesAlreadyTakenCopyOnPickup(t *testing.T) {
	g, facade := newSelectionTestGame(t)
	torch := &items.Item{
		Entity:     &entities.Entity{ID: "ITM0001"},
		Name:       "Dusty Torch",
		IsTemplate: true,
	}
	if _, err := facade.ItemsService().Import(torch); err != nil {
		t.Fatal(err)
	}
	exits := rooms.Exits{}
	chars := rooms.Characters{}
	roomItems := rooms.Items{"ITM0001"}
	room := &rooms.Room{
		Entity:      &entities.Entity{ID: "R0003"},
		Name:        "Torch Alcove",
		Description: "A rack of torches.",
		Exits:       &exits,
		Characters:  &chars,
		Items:       &roomItems,
	}
	if _, err := facade.RoomsService().Import(room); err != nil {
		t.Fatal(err)
	}
	character, err := facade.CharactersService().Store(&characters.Character{
		Name:        "Wanderer",
		BelongsUser: *traits.BelongsToUser("user-1"),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "R0003"},
		Inventory:   items.Inventory{Size: 10},
	})
	if err != nil {
		t.Fatal(err)
	}
	character.MarkCollectedCopyItem("ITM0001")
	_ = facade.CharactersService().Update(character.ID, character)

	user := &entities.User{Entity: &entities.Entity{ID: "user-1"}}
	msg := &messages.Message{FromUser: user, Character: character, Data: "look"}
	if !commands.Look(room, g, msg) {
		t.Fatal("look failed")
	}
	for _, out := range drainSelectionMessages(g.SendMessage()) {
		if rsp, ok := out.(messages.MessageResponse); ok {
			if strings.Contains(rsp.Message, "Dusty Torch") {
				t.Fatal("look must hide already-taken copyOnPickup torch")
			}
		}
	}
}

func TestStripHiddenExitsFromEnterRoomJSON(t *testing.T) {
	g, facade := newSelectionTestGame(t)
	exits := rooms.Exits{
		{Name: "north", Target: "R2", Hidden: false},
		{Name: "down", Target: "R3", Hidden: true},
	}
	chars := rooms.Characters{}
	roomItems := rooms.Items{}
	room := &rooms.Room{
		Entity:     &entities.Entity{ID: "R1"},
		Name:       "Hall",
		Exits:      &exits,
		Characters: &chars,
		Items:      &roomItems,
	}
	if _, err := facade.RoomsService().Import(room); err != nil {
		t.Fatal(err)
	}
	user := &entities.User{Entity: &entities.Entity{ID: "user-1"}, LastCharacter: "c1"}
	char := &characters.Character{Entity: &entities.Entity{ID: "c1"}, Name: "W"}
	enter := messages.NewEnterRoomMessage(room, user, g, char)
	if enter.Room.Exits == nil {
		t.Fatal("expected exits")
	}
	for _, e := range *enter.Room.Exits {
		if e.Hidden || strings.EqualFold(e.Name, "down") {
			t.Fatalf("hidden exit leaked into EnterRoomMessage: %#v", e)
		}
	}
	stripped := util.StripHiddenExits(room)
	if len(*stripped.Exits) != 1 {
		t.Fatalf("expected 1 visible exit, got %d", len(*stripped.Exits))
	}
}

func TestPickupWeaponAppendsEquipHint(t *testing.T) {
	g, facade := newSelectionTestGame(t)
	dagger := &items.Item{
		Entity: &entities.Entity{ID: "dagger-1"},
		Name:   "Rusty Dagger",
		Type:   items.ItemTypeWeapon,
		Slot:   items.ItemSlotMainHand,
	}
	if _, err := facade.ItemsService().Import(dagger); err != nil {
		t.Fatal(err)
	}
	exits := rooms.Exits{}
	chars := rooms.Characters{}
	roomItems := rooms.Items{"dagger-1"}
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity:     &entities.Entity{ID: "R-pick"},
		Name:       "Cache",
		Exits:      &exits,
		Characters: &chars,
		Items:      &roomItems,
	}); err != nil {
		t.Fatal(err)
	}
	user := &entities.User{Entity: &entities.Entity{ID: "user-1"}}
	character, err := facade.CharactersService().Store(&characters.Character{
		Name:        "Wanderer",
		BelongsUser: *traits.BelongsToUser("user-1"),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "R-pick"},
		Inventory:   items.Inventory{Size: 10},
	})
	if err != nil {
		t.Fatal(err)
	}
	msg := &messages.Message{FromUser: user, Character: character, Data: "take dagger"}
	if !(&commands.PickupCommand{}).Execute(g, msg) {
		t.Fatal("pickup failed")
	}
	var sawHint bool
	for _, out := range drainSelectionMessages(g.SendMessage()) {
		if rsp, ok := out.(messages.MessageResponse); ok {
			if strings.Contains(rsp.Message, "equip it with: equip Rusty Dagger") {
				sawHint = true
			}
		}
	}
	if !sawHint {
		t.Fatal("expected equip hint when main_hand empty")
	}
}
