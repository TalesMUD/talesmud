package service

import (
	"path/filepath"
	"testing"

	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/settings"
	"github.com/talesmud/talesmud/pkg/repository"
)

func TestResolveStartRoomPrefersR0001OverEarlierRooms(t *testing.T) {
	client, err := sqlite.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	facade := NewFacade(repository.NewSQLiteFactory(client), nil)

	exits := rooms.Exits{}
	chars := rooms.Characters{}
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity: &entities.Entity{ID: "R1901"}, Name: "Decoy", Exits: &exits, Characters: &chars,
	}); err != nil {
		t.Fatalf("import decoy: %v", err)
	}
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity: &entities.Entity{ID: "R0001"}, Name: "Awakening", Exits: &exits, Characters: &chars,
	}); err != nil {
		t.Fatalf("import start: %v", err)
	}

	id := ResolveStartRoomID(facade.ServerSettingsService(), facade.RoomsService())
	if id != "R0001" {
		t.Fatalf("expected R0001, got %q", id)
	}
}

func TestResolveStartRoomUsesConfiguredID(t *testing.T) {
	client, err := sqlite.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	facade := NewFacade(repository.NewSQLiteFactory(client), nil)

	exits := rooms.Exits{}
	chars := rooms.Characters{}
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity: &entities.Entity{ID: "R0101"}, Name: "Meadow", Exits: &exits, Characters: &chars,
	}); err != nil {
		t.Fatalf("import meadow: %v", err)
	}
	s := settings.NewDefaultServerSettings()
	s.StartRoomID = "R0101"
	if err := facade.ServerSettingsService().Update(s); err != nil {
		t.Fatalf("update settings: %v", err)
	}

	id := ResolveStartRoomID(facade.ServerSettingsService(), facade.RoomsService())
	if id != "R0101" {
		t.Fatalf("expected R0101, got %q", id)
	}
}
