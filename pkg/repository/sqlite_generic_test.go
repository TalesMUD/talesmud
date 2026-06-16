package repository

import (
	"path/filepath"
	"testing"

	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

func TestSQLiteGenericUpdateDeleteReportMissingEntity(t *testing.T) {
	client, err := sqlite.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer client.Close()

	repo := NewSQLiteRoomsRepository(client)
	room := &rooms.Room{Entity: &entities.Entity{ID: "missing"}, Name: "Missing"}

	if err := repo.Update(room.ID, room); err == nil {
		t.Fatal("expected update of missing room to return an error")
	}
	if err := repo.Delete(room.ID); err == nil {
		t.Fatal("expected delete of missing room to return an error")
	}
}
