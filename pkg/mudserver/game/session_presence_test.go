package game

import (
	"path/filepath"
	"testing"

	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

func newSessionTestGame(t *testing.T) (*Game, service.Facade) {
	t.Helper()

	client, err := sqlite.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	return New(facade), facade
}

func TestRoomPresenceUsesLiveSessionsInsteadOfStaleRoomCharacters(t *testing.T) {
	g, facade := newSessionTestGame(t)

	roomChars := rooms.Characters{"char-online", "char-stale"}
	room := &rooms.Room{
		Entity:     &entities.Entity{ID: "room-1"},
		Name:       "Town Square",
		Characters: &roomChars,
	}
	if _, err := facade.RoomsService().Import(room); err != nil {
		t.Fatalf("import room: %v", err)
	}

	onlineUser := &entities.User{Entity: &entities.Entity{ID: "user-online"}, RefID: "ref-online", LastCharacter: "char-online"}
	staleUser := &entities.User{Entity: &entities.Entity{ID: "user-stale"}, RefID: "ref-stale", LastCharacter: "char-stale", IsOnline: true}
	if _, err := facade.UsersService().Import(onlineUser); err != nil {
		t.Fatalf("import online user: %v", err)
	}
	if _, err := facade.UsersService().Import(staleUser); err != nil {
		t.Fatalf("import stale user: %v", err)
	}

	onlineChar := &characters.Character{
		Entity:      &entities.Entity{ID: "char-online"},
		Name:        "Aryn",
		BelongsUser: *traits.BelongsToUser("user-online"),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "room-1"},
	}
	staleChar := &characters.Character{
		Entity:      &entities.Entity{ID: "char-stale"},
		Name:        "Bran",
		BelongsUser: *traits.BelongsToUser("user-stale"),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "room-1"},
	}
	if _, err := facade.CharactersService().Import(onlineChar); err != nil {
		t.Fatalf("import online character: %v", err)
	}
	if _, err := facade.CharactersService().Import(staleChar); err != nil {
		t.Fatalf("import stale character: %v", err)
	}

	g.ConnectUserSession(onlineUser)
	g.SetUserSessionCharacter(onlineUser, onlineChar)

	players := g.GetRoomPlayers("room-1", "char-online")
	if len(players) != 1 {
		t.Fatalf("expected one live player, got %d: %#v", len(players), players)
	}
	if players[0].CharacterName != "Aryn" {
		t.Fatalf("expected Aryn in room presence, got %q", players[0].CharacterName)
	}
	if !players[0].IsYou {
		t.Fatal("expected viewer character to be marked as self")
	}

	g.DisconnectUserSession("user-online")
	if players := g.GetRoomPlayers("room-1", "char-online"); len(players) != 0 {
		t.Fatalf("expected disconnected user to disappear immediately, got %#v", players)
	}
}
