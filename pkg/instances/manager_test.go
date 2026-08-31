package instances

import (
	"path/filepath"
	"testing"

	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

func testRooms(t *testing.T) service.RoomsService {
	t.Helper()
	client, err := sqlite.Open(filepath.Join(t.TempDir(), "inst.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	seed := []*rooms.Room{
		room("R0201", rooms.Exit{Name: "down", Target: "R0210", Type: "instance", Instance: true}),
		room("R0210",
			rooms.Exit{Name: "up", Target: "R0201"},
			rooms.Exit{Name: "east", Target: "R0211"}),
		room("R0211", rooms.Exit{Name: "west", Target: "R0210"}),
	}
	for _, r := range seed {
		if _, err := facade.RoomsService().Import(r); err != nil {
			t.Fatalf("import %s: %v", r.ID, err)
		}
	}
	return facade.RoomsService()
}

func TestTwoCharactersGetPrivateCellarsAndSharedHub(t *testing.T) {
	svc := testRooms(t)
	m := NewManager()

	a, err := m.Enter(svc, "char-a", "R0201", "R0210")
	if err != nil {
		t.Fatal(err)
	}
	b, err := m.Enter(svc, "char-b", "R0201", "R0210")
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("two guests must not share a cellar room id")
	}
	if a == "R0210" || b == "R0210" {
		t.Fatal("clone must not be the template id")
	}

	ra, err := svc.FindByID(a)
	if err != nil {
		t.Fatal(err)
	}
	up, ok := ra.GetExit("up")
	if !ok || up.Target != "R0201" {
		t.Fatalf("clone must exit back to shared hub, got %+v", up)
	}
	east, ok := ra.GetExit("east")
	if !ok {
		t.Fatal("missing east")
	}
	if east.Target == "R0211" {
		t.Fatal("internal cellar exit must remap to the private copy")
	}

	hub, err := svc.FindByID("R0201")
	if err != nil {
		t.Fatal(err)
	}
	if hub.ID != "R0201" {
		t.Fatal("hub must stay the original room")
	}
}

func TestEmptyInstanceIsDestroyed(t *testing.T) {
	svc := testRooms(t)
	m := NewManager()
	dest, err := m.Enter(svc, "char-a", "R0201", "R0210")
	if err != nil {
		t.Fatal(err)
	}
	m.NoteLeave(svc, "char-a", dest, "R0201")
	if _, err := svc.FindByID(dest); err == nil {
		t.Fatal("empty cellar should be deleted")
	}
	if m.IsClone(dest) {
		t.Fatal("clone index should be cleared")
	}
}

func TestOccupantKeepsInstanceAlive(t *testing.T) {
	svc := testRooms(t)
	m := NewManager()
	a, _ := m.Enter(svc, "char-a", "R0201", "R0210")
	b, _ := m.Enter(svc, "char-b", "R0201", "R0210")
	m.NoteLeave(svc, "char-a", a, "R0201")
	if _, err := svc.FindByID(b); err != nil {
		t.Fatal("second guest's cellar must remain")
	}
}

func TestReenterExistingInstance(t *testing.T) {
	svc := testRooms(t)
	m := NewManager()
	first, err := m.Enter(svc, "char-a", "R0201", "R0210")
	if err != nil {
		t.Fatal(err)
	}
	again, err := m.Enter(svc, "char-a", "R0201", "R0210")
	if err != nil {
		t.Fatal(err)
	}
	if first != again {
		t.Fatalf("same character should reuse their cellar: %s vs %s", first, again)
	}
}
