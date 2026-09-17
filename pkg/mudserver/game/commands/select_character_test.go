package commands_test

import (
	"path/filepath"
	"testing"

	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

func newSelectionTestGame(t *testing.T) (*game.Game, service.Facade) {
	t.Helper()

	client, err := sqlite.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	return game.New(facade), facade
}

func drainSelectionMessages(ch <-chan interface{}) []interface{} {
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

func TestSelectCharacterSendsRoomPresenceRefresh(t *testing.T) {
	g, facade := newSelectionTestGame(t)
	roomExits := rooms.Exits{}
	roomCharacters := rooms.Characters{}
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity:      &entities.Entity{ID: "room-1"},
		Name:        "Commons",
		Description: "A shared room.",
		Exits:       &roomExits,
		Characters:  &roomCharacters,
	}); err != nil {
		t.Fatalf("import room: %v", err)
	}
	user := &entities.User{Entity: &entities.Entity{ID: "user-1"}, RefID: "auth|1", IsOnline: true}
	if _, err := facade.UsersService().Import(user); err != nil {
		t.Fatalf("import user: %v", err)
	}
	if _, err := facade.CharactersService().Import(&characters.Character{
		Entity:      &entities.Entity{ID: "char-1"},
		Name:        "Aster",
		BelongsUser: *traits.BelongsToUser("user-1"),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "room-1"},
	}); err != nil {
		t.Fatalf("import character: %v", err)
	}

	msg := &messages.Message{FromUser: user, Data: "sc Aster"}
	if !(&commands.SelectCharacterCommand{}).Execute(g, msg) {
		t.Fatal("select character did not handle command")
	}

	var sawPresence bool
	for _, out := range drainSelectionMessages(g.SendMessage()) {
		if presence, ok := out.(*messages.RoomPresenceMessage); ok {
			sawPresence = true
			if presence.AudienceID != "room-1" {
				t.Fatalf("expected room presence for room-1, got %s", presence.AudienceID)
			}
			if len(presence.Players) != 1 || presence.Players[0].Name != "Aster" {
				t.Fatalf("expected presence for Aster, got %#v", presence.Players)
			}
		}
	}
	if !sawPresence {
		t.Fatal("expected select character to send room presence refresh")
	}
}

func TestSelectCharacterEmptyRoomSpawnsInR0001NotFirstRoom(t *testing.T) {
	g, facade := newSelectionTestGame(t)
	roomExits := rooms.Exits{}
	decoyChars := rooms.Characters{}
	startChars := rooms.Characters{}
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity:      &entities.Entity{ID: "R1901"},
		Name:        "Decoy",
		Description: "Should not spawn here.",
		Exits:       &roomExits,
		Characters:  &decoyChars,
	}); err != nil {
		t.Fatalf("import decoy: %v", err)
	}
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity:      &entities.Entity{ID: "R0001"},
		Name:        "Awakening Chamber",
		Description: "Start here.",
		Exits:       &roomExits,
		Characters:  &startChars,
	}); err != nil {
		t.Fatalf("import start room: %v", err)
	}
	user := &entities.User{Entity: &entities.Entity{ID: "user-1"}, RefID: "auth|1", IsOnline: true}
	if _, err := facade.UsersService().Import(user); err != nil {
		t.Fatalf("import user: %v", err)
	}
	if _, err := facade.CharactersService().Import(&characters.Character{
		Entity:      &entities.Entity{ID: "char-1"},
		Name:        "Wanderer",
		BelongsUser: *traits.BelongsToUser("user-1"),
	}); err != nil {
		t.Fatalf("import character: %v", err)
	}

	msg := &messages.Message{FromUser: user, Data: "sc Wanderer"}
	if !(&commands.SelectCharacterCommand{}).Execute(g, msg) {
		t.Fatal("select character did not handle command")
	}

	char, err := facade.CharactersService().FindByID("char-1")
	if err != nil {
		t.Fatalf("load character: %v", err)
	}
	if char.CurrentRoomID != "R0001" {
		t.Fatalf("expected spawn in R0001, got %q", char.CurrentRoomID)
	}
	startRoom, err := facade.RoomsService().FindByID("R0001")
	if err != nil {
		t.Fatalf("load start room: %v", err)
	}
	if !startRoom.IsCharacterInRoom("char-1") {
		t.Fatal("expected character added to R0001")
	}
	decoy, err := facade.RoomsService().FindByID("R1901")
	if err != nil {
		t.Fatalf("load decoy: %v", err)
	}
	if decoy.IsCharacterInRoom("char-1") {
		t.Fatal("character spawned in rooms[0] decoy instead of R0001")
	}
}

func TestSelectCharacterSwitchRemovesPreviousCharacterFromRoom(t *testing.T) {
	g, facade := newSelectionTestGame(t)
	roomExits := rooms.Exits{}
	oldRoomCharacters := rooms.Characters{"char-old"}
	newRoomCharacters := rooms.Characters{}
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity:      &entities.Entity{ID: "room-old"},
		Name:        "Old Room",
		Description: "Previous place.",
		Exits:       &roomExits,
		Characters:  &oldRoomCharacters,
	}); err != nil {
		t.Fatalf("import old room: %v", err)
	}
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity:      &entities.Entity{ID: "room-new"},
		Name:        "New Room",
		Description: "Next place.",
		Exits:       &roomExits,
		Characters:  &newRoomCharacters,
	}); err != nil {
		t.Fatalf("import new room: %v", err)
	}
	user := &entities.User{Entity: &entities.Entity{ID: "user-1"}, RefID: "auth|1", LastCharacter: "char-old", IsOnline: true}
	if _, err := facade.UsersService().Import(user); err != nil {
		t.Fatalf("import user: %v", err)
	}
	if _, err := facade.CharactersService().Import(&characters.Character{
		Entity:      &entities.Entity{ID: "char-old"},
		Name:        "Oldster",
		BelongsUser: *traits.BelongsToUser("user-1"),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "room-old"},
	}); err != nil {
		t.Fatalf("import old character: %v", err)
	}
	if _, err := facade.CharactersService().Import(&characters.Character{
		Entity:      &entities.Entity{ID: "char-new"},
		Name:        "Aster",
		BelongsUser: *traits.BelongsToUser("user-1"),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "room-new"},
	}); err != nil {
		t.Fatalf("import new character: %v", err)
	}

	msg := &messages.Message{FromUser: user, Data: "sc Aster"}
	if !(&commands.SelectCharacterCommand{}).Execute(g, msg) {
		t.Fatal("select character did not handle command")
	}

	oldRoom, err := facade.RoomsService().FindByID("room-old")
	if err != nil {
		t.Fatalf("load old room: %v", err)
	}
	if oldRoom.IsCharacterInRoom("char-old") {
		t.Fatal("expected previous character to be removed from old room")
	}
	newRoom, err := facade.RoomsService().FindByID("room-new")
	if err != nil {
		t.Fatalf("load new room: %v", err)
	}
	if !newRoom.IsCharacterInRoom("char-new") {
		t.Fatal("expected selected character to be added to new room")
	}
}

func TestSelectCharacterAppliesBankedXPLevelCatchUp(t *testing.T) {
	g, facade := newSelectionTestGame(t)
	roomExits := rooms.Exits{}
	roomCharacters := rooms.Characters{}
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity:      &entities.Entity{ID: "room-1"},
		Name:        "Commons",
		Description: "A shared room.",
		Exits:       &roomExits,
		Characters:  &roomCharacters,
	}); err != nil {
		t.Fatalf("import room: %v", err)
	}
	user := &entities.User{Entity: &entities.Entity{ID: "user-1"}, RefID: "auth|1", IsOnline: true}
	if _, err := facade.UsersService().Import(user); err != nil {
		t.Fatalf("import user: %v", err)
	}

	// Marcus-style banked XP: L15 with XP past L16 threshold (GetXPRequired(16)=2922)
	xpFor16 := leveling.GetXPRequired(16)
	if _, err := facade.CharactersService().Import(&characters.Character{
		Entity:           &entities.Entity{ID: "char-gimli"},
		Name:             "Gimli",
		BelongsUser:      *traits.BelongsToUser("user-1"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "room-1"},
		Level:            15,
		XP:               xpFor16 + 108, // 3030-style overshoot
		MaxHitPoints:     100,
		CurrentHitPoints: 100,
		Class:            characters.ClassWarrior,
		Attributes: []characters.Attribute{
			{Short: "STR", Value: 14},
			{Short: "DEX", Value: 10},
			{Short: "INT", Value: 8},
			{Short: "WIS", Value: 8},
			{Short: "STA", Value: 12},
		},
	}); err != nil {
		t.Fatalf("import character: %v", err)
	}

	msg := &messages.Message{FromUser: user, Data: "sc Gimli"}
	if !(&commands.SelectCharacterCommand{}).Execute(g, msg) {
		t.Fatal("select character did not handle command")
	}

	outs := drainSelectionMessages(g.SendMessage())
	var sawLevelUp bool
	var selected *messages.CharacterSelected
	for _, out := range outs {
		if resp, ok := out.(messages.MessageResponse); ok && resp.Type == messages.MessageTypeLevelUp {
			sawLevelUp = true
			if resp.Message == "" {
				t.Fatal("level-up message empty")
			}
		}
		if cs, ok := out.(*messages.CharacterSelected); ok {
			selected = cs
		}
	}
	if !sawLevelUp {
		t.Fatal("expected levelUp message on select with banked XP")
	}

	stored, err := facade.CharactersService().FindByID("char-gimli")
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if stored.Level != 16 {
		t.Fatalf("expected catch-up to level 16, got %d", stored.Level)
	}
	if selected == nil {
		t.Fatal("expected CharacterSelected message")
	}
	if selected.Character.Level != 16 {
		t.Fatalf("CharacterSelected still shows level %d", selected.Character.Level)
	}
	if selected.XPForNextLevel != leveling.GetXPRequired(17) {
		t.Fatalf("XPForNextLevel=%d want %d", selected.XPForNextLevel, leveling.GetXPRequired(17))
	}
}
