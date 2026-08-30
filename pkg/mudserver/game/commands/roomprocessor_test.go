package commands_test

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func processPlayerInput(g *game.Game, msg *messages.Message) {
	if !g.CommandProcessor.Process(g, msg) {
		g.RoomProcessor.Process(g, msg)
	}
}

func TestRoomActionBeatsExamineAndSendsResponse(t *testing.T) {
	g, facade := newSelectionTestGame(t)
	actions := rooms.Actions{
		{
			Name:        "EXAMINE PEDESTALS",
			Type:        rooms.RoomActionTypeResponse,
			Description: "Inspect the five stone pedestals.",
			Response:    "Five stone pedestals stand in a precise circle.",
		},
	}
	exits := rooms.Exits{}
	chars := rooms.Characters{}
	roomItems := rooms.Items{}
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity:      &entities.Entity{ID: "R0004"},
		Name:        "Rune Chamber",
		Description: "Runes.",
		Exits:       &exits,
		Characters:  &chars,
		Items:       &roomItems,
		Actions:     &actions,
	}); err != nil {
		t.Fatalf("import room: %v", err)
	}
	user := &entities.User{Entity: &entities.Entity{ID: "user-1"}, RefID: "auth|1"}
	character := &characters.Character{
		Entity:      &entities.Entity{ID: "char-1"},
		Name:        "Wanderer",
		BelongsUser: *traits.BelongsToUser("user-1"),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "R0004"},
	}
	if _, err := facade.CharactersService().Import(character); err != nil {
		t.Fatalf("import character: %v", err)
	}

	msg := &messages.Message{FromUser: user, Character: character, Data: "examine pedestals"}
	processPlayerInput(g, msg)

	var replies []string
	for _, out := range drainSelectionMessages(g.SendMessage()) {
		if resp, ok := out.(messages.MessageResponse); ok {
			replies = append(replies, resp.Message)
		}
	}
	found := false
	for _, r := range replies {
		if r == "Five stone pedestals stand in a precise circle." {
			found = true
		}
		if r == "Inspect the five stone pedestals." {
			t.Fatal("sent action Description instead of Response")
		}
	}
	if !found {
		t.Fatalf("expected narrative response, got %#v", replies)
	}
}

func TestGatherHerbsRoomActionMatches(t *testing.T) {
	g, facade := newSelectionTestGame(t)
	actions := rooms.Actions{
		{Name: "GATHER HERBS", Type: rooms.RoomActionTypeResponse, Response: "You pinch herbs from the meadow."},
	}
	exits := rooms.Exits{}
	chars := rooms.Characters{}
	roomItems := rooms.Items{}
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity:     &entities.Entity{ID: "R0102"},
		Name:       "Wildflower Field",
		Exits:      &exits,
		Characters: &chars,
		Items:      &roomItems,
		Actions:    &actions,
	}); err != nil {
		t.Fatalf("import room: %v", err)
	}
	user := &entities.User{Entity: &entities.Entity{ID: "user-1"}, RefID: "auth|1"}
	character := &characters.Character{
		Entity:      &entities.Entity{ID: "char-1"},
		Name:        "Wanderer",
		BelongsUser: *traits.BelongsToUser("user-1"),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "R0102"},
	}
	if _, err := facade.CharactersService().Import(character); err != nil {
		t.Fatalf("import character: %v", err)
	}
	msg := &messages.Message{FromUser: user, Character: character, Data: "gather herbs"}
	processPlayerInput(g, msg)
	var found bool
	for _, out := range drainSelectionMessages(g.SendMessage()) {
		if resp, ok := out.(messages.MessageResponse); ok && resp.Message == "You pinch herbs from the meadow." {
			found = true
		}
	}
	if !found {
		t.Fatal("gather herbs did not match the room action")
	}
}

func TestPickupRoomTemplateLeavesTemplateInRoom(t *testing.T) {
	g, facade := newSelectionTestGame(t)
	torch := &items.Item{
		Entity:     &entities.Entity{ID: "ITM0001"},
		Name:       "Dusty Torch",
		IsTemplate: true,
	}
	if _, err := facade.ItemsService().Import(torch); err != nil {
		t.Fatalf("import template: %v", err)
	}
	exits := rooms.Exits{}
	chars := rooms.Characters{}
	roomItems := rooms.Items{"ITM0001"}
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity:      &entities.Entity{ID: "R0003"},
		Name:        "Torch Alcove",
		Description: "Cache.",
		Exits:       &exits,
		Characters:  &chars,
		Items:       &roomItems,
	}); err != nil {
		t.Fatalf("import room: %v", err)
	}
	user := &entities.User{Entity: &entities.Entity{ID: "user-1"}, RefID: "auth|1"}
	character, err := facade.CharactersService().Store(&characters.Character{
		Name:        "Wanderer",
		BelongsUser: *traits.BelongsToUser("user-1"),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "R0003"},
		Inventory:   items.Inventory{Size: 10},
	})
	if err != nil {
		t.Fatalf("store character: %v", err)
	}

	msg := &messages.Message{FromUser: user, Character: character, Data: "take torch"}
	if !(&commands.PickupCommand{}).Execute(g, msg) {
		t.Fatal("pickup did not handle command")
	}

	room, err := facade.RoomsService().FindByID("R0003")
	if err != nil {
		t.Fatalf("load room: %v", err)
	}
	found := false
	for _, id := range room.GetItemIDs() {
		if id == "ITM0001" {
			found = true
		}
	}
	if !found {
		t.Fatal("template was stolen from the room")
	}
	updated, err := facade.CharactersService().FindByID(character.ID)
	if err != nil {
		t.Fatalf("load character: %v", err)
	}
	if updated.Inventory.FindItemByName("Dusty Torch") == nil {
		t.Fatal("character did not receive a torch copy")
	}
}

func TestTwoCharactersCopyRoomTorchWithoutStealing(t *testing.T) {
	g, facade := newSelectionTestGame(t)
	torch := &items.Item{
		Entity:     &entities.Entity{ID: "ITM0001"},
		Name:       "Dusty Torch",
		IsTemplate: true,
	}
	if _, err := facade.ItemsService().Import(torch); err != nil {
		t.Fatalf("import template: %v", err)
	}
	exits := rooms.Exits{}
	chars := rooms.Characters{}
	roomItems := rooms.Items{"ITM0001"}
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity:      &entities.Entity{ID: "R0003"},
		Name:        "Torch Alcove",
		Description: "Cache.",
		Exits:       &exits,
		Characters:  &chars,
		Items:       &roomItems,
	}); err != nil {
		t.Fatalf("import room: %v", err)
	}

	pickupAs := func(userID, charName string) {
		t.Helper()
		user := &entities.User{Entity: &entities.Entity{ID: userID}, RefID: userID}
		character, err := facade.CharactersService().Store(&characters.Character{
			Name:        charName,
			BelongsUser: *traits.BelongsToUser(userID),
			CurrentRoom: traits.CurrentRoom{CurrentRoomID: "R0003"},
			Inventory:   items.Inventory{Size: 10},
		})
		if err != nil {
			t.Fatalf("store %s: %v", charName, err)
		}
		msg := &messages.Message{FromUser: user, Character: character, Data: "take torch"}
		if !(&commands.PickupCommand{}).Execute(g, msg) {
			t.Fatalf("%s pickup did not handle command", charName)
		}
		updated, err := facade.CharactersService().FindByID(character.ID)
		if err != nil {
			t.Fatalf("load %s: %v", charName, err)
		}
		if updated.Inventory.FindItemByName("Dusty Torch") == nil {
			t.Fatalf("%s did not receive a torch copy", charName)
		}
	}

	pickupAs("user-1", "FirstGuest")
	pickupAs("user-2", "SecondGuest")

	room, err := facade.RoomsService().FindByID("R0003")
	if err != nil {
		t.Fatalf("load room: %v", err)
	}
	found := false
	for _, id := range room.GetItemIDs() {
		if id == "ITM0001" {
			found = true
		}
	}
	if !found {
		t.Fatal("second guest stole the world torch template")
	}
}

func TestPickupCopiesCatalogItemEvenWithoutIsTemplateFlag(t *testing.T) {
	g, facade := newSelectionTestGame(t)
	flint := &items.Item{
		Entity:     &entities.Entity{ID: "ITM0002"},
		Name:       "Flint and Steel",
		IsTemplate: false,
	}
	if _, err := facade.ItemsService().Import(flint); err != nil {
		t.Fatalf("import catalog item: %v", err)
	}
	exits := rooms.Exits{}
	chars := rooms.Characters{}
	roomItems := rooms.Items{"ITM0002"}
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity:      &entities.Entity{ID: "R0003"},
		Name:        "Torch Alcove",
		Exits:       &exits,
		Characters:  &chars,
		Items:       &roomItems,
	}); err != nil {
		t.Fatalf("import room: %v", err)
	}
	user := &entities.User{Entity: &entities.Entity{ID: "user-1"}, RefID: "auth|1"}
	character, err := facade.CharactersService().Store(&characters.Character{
		Name:        "Wanderer",
		BelongsUser: *traits.BelongsToUser("user-1"),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "R0003"},
		Inventory:   items.Inventory{Size: 10},
	})
	if err != nil {
		t.Fatalf("store character: %v", err)
	}

	msg := &messages.Message{FromUser: user, Character: character, Data: "take flint"}
	if !(&commands.PickupCommand{}).Execute(g, msg) {
		t.Fatal("pickup did not handle command")
	}

	room, err := facade.RoomsService().FindByID("R0003")
	if err != nil {
		t.Fatalf("load room: %v", err)
	}
	found := false
	for _, id := range room.GetItemIDs() {
		if id == "ITM0002" {
			found = true
		}
	}
	if !found {
		t.Fatal("catalog flint was stolen from the room")
	}
	updated, err := facade.CharactersService().FindByID(character.ID)
	if err != nil {
		t.Fatalf("load character: %v", err)
	}
	if updated.Inventory.FindItemByName("Flint and Steel") == nil {
		t.Fatal("character did not receive a flint copy")
	}
}
