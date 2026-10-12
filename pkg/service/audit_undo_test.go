package service

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/repository"
)

func TestUndoCRUDRestoresUpdateAndDeleteAndRefusesConflict(t *testing.T) {
	client, err := sqlite.Open(filepath.Join(t.TempDir(), "audit.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	facade := NewFacade(repository.NewSQLiteFactory(client), nil)
	auditSvc := facade.AuditService()
	actor := AuditActor{UserID: "admin", Name: "Admin"}

	original := &rooms.Room{Entity: &entities.Entity{ID: "room-1"}, Name: "Old Name"}
	if _, err := facade.RoomsService().Store(original); err != nil {
		t.Fatalf("store: %v", err)
	}
	before, err := auditSvc.LoadRaw("rooms", "room-1")
	if err != nil || len(before) == 0 {
		t.Fatalf("before snapshot: %v %s", err, before)
	}
	updated := &rooms.Room{Entity: &entities.Entity{ID: "room-1"}, Name: "New Name"}
	if err := facade.RoomsService().Update("room-1", updated); err != nil {
		t.Fatalf("update: %v", err)
	}
	after, err := auditSvc.LoadRaw("rooms", "room-1")
	if err != nil {
		t.Fatalf("after: %v", err)
	}
	entry, err := auditSvc.Record(AuditRecord{
		Actor: actor, Action: "update", EntityType: "rooms", EntityID: "room-1",
		Before: before, After: after, Undoable: true, Summary: "rename room",
	})
	if err != nil {
		t.Fatalf("record: %v", err)
	}

	changed := &rooms.Room{Entity: &entities.Entity{ID: "room-1"}, Name: "Someone else"}
	if err := facade.RoomsService().Update("room-1", changed); err != nil {
		t.Fatalf("conflict update: %v", err)
	}
	if _, err := auditSvc.UndoCRUD(entry.ID, actor); err == nil {
		t.Fatal("expected conflict")
	} else if ae, ok := err.(*AuditError); !ok || ae.Status != 409 {
		t.Fatalf("conflict status: %v", err)
	}
	// Put the audited after-state back so undo can proceed.
	if err := facade.RoomsService().Update("room-1", updated); err != nil {
		t.Fatalf("restore after: %v", err)
	}
	undo, err := auditSvc.UndoCRUD(entry.ID, actor)
	if err != nil {
		t.Fatalf("undo update: %v", err)
	}
	got, err := facade.RoomsService().FindByID("room-1")
	if err != nil || got.Name != "Old Name" {
		t.Fatalf("restored name: %+v %v", got, err)
	}
	if _, err := auditSvc.UndoCRUD(entry.ID, actor); err == nil {
		t.Fatal("second undo should fail")
	}

	// Undo of the undo puts the new name back.
	redo, err := auditSvc.UndoCRUD(undo.ID, actor)
	if err != nil {
		t.Fatalf("undo of undo: %v", err)
	}
	got, _ = facade.RoomsService().FindByID("room-1")
	if got.Name != "New Name" {
		t.Fatalf("redo name %q", got.Name)
	}
	_ = redo

	deletedBefore, _ := auditSvc.LoadRaw("rooms", "room-1")
	if err := facade.RoomsService().Delete("room-1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	del, err := auditSvc.Record(AuditRecord{
		Actor: actor, Action: "delete", EntityType: "rooms", EntityID: "room-1",
		Before: deletedBefore, Undoable: true, Summary: "delete room",
	})
	if err != nil {
		t.Fatalf("record delete: %v", err)
	}
	if _, err := auditSvc.UndoCRUD(del.ID, actor); err != nil {
		t.Fatalf("undo delete: %v", err)
	}
	got, err = facade.RoomsService().FindByID("room-1")
	if err != nil || got.Name != "New Name" {
		t.Fatalf("recreated %+v %v", got, err)
	}

	rows, err := auditSvc.List("rooms", "room-1", 10)
	if err != nil || len(rows) < 3 {
		t.Fatalf("list %d %v", len(rows), err)
	}
	raw, _ := json.Marshal(rows[0])
	if !json.Valid(raw) {
		t.Fatal("row is not json")
	}
}
