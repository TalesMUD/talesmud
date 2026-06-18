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
