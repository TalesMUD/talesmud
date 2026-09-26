package instances

import (
	"testing"
	"time"
)

func TestGenerateIsPerCharacterAndFiltersLevel(t *testing.T) {
	svc := testRooms(t)
	for _, id := range []string{"glade", "thicket"} {
		if _, err := svc.Import(room(id)); err != nil {
			t.Fatal(err)
		}
	}
	m := NewManager()
	spec := ProcSpec{
		TemplateIDs:  []string{"glade", "thicket"},
		Count:        3,
		ReturnRoomID: "R0201",
		Seed:         7,
		Timeout:      time.Minute,
		Encounters: []Encounter{
			{TemplateID: "wolf", MinLevel: 1, MaxLevel: 3, Weight: 1},
			{TemplateID: "bear", MinLevel: 8, MaxLevel: 12, Weight: 5},
		},
	}
	low, err := m.Generate(svc, "low", 2, spec)
	if err != nil {
		t.Fatal(err)
	}
	high, err := m.Generate(svc, "high", 10, spec)
	if err != nil {
		t.Fatal(err)
	}
	if low.EntryRoomID == high.EntryRoomID || low.EntryRoomID == "" {
		t.Fatalf("entries %q %q", low.EntryRoomID, high.EntryRoomID)
	}
	if _, err := m.Generate(svc, "low", 2, spec); err == nil {
		t.Fatal("second instance for the same character should fail")
	}
	for _, sp := range low.Spawns {
		if sp.TemplateID != "wolf" {
			t.Fatalf("low level spawned %s", sp.TemplateID)
		}
	}
	if len(low.Spawns) == 0 {
		t.Fatal("expected a wolf in the low band")
	}
	for _, sp := range high.Spawns {
		if sp.TemplateID != "bear" {
			t.Fatalf("high level spawned %s", sp.TemplateID)
		}
	}
	none, err := m.Generate(svc, "mid", 5, spec)
	if err != nil {
		t.Fatal(err)
	}
	if len(none.Spawns) != 0 {
		t.Fatalf("level 5 matched %+v", none.Spawns)
	}

	entry, err := svc.FindByID(low.EntryRoomID)
	if err != nil {
		t.Fatal(err)
	}
	north, ok := entry.GetExit("north")
	if !ok || north.Target == low.EntryRoomID || !north.Instance {
		t.Fatalf("north %+v", north)
	}
	out, ok := entry.GetExit("out")
	if !ok || out.Target != "R0201" || !out.Instance {
		t.Fatalf("out %+v", out)
	}

	deleted := m.NoteLeave(svc, "low", low.EntryRoomID, "R0201")
	if len(deleted) != 3 {
		t.Fatalf("leave deleted %v", deleted)
	}
	if _, err := svc.FindByID(low.EntryRoomID); err == nil {
		t.Fatal("clone survived leave")
	}
}

func TestExpireOnlyProcedural(t *testing.T) {
	svc := testRooms(t)
	if _, err := svc.Import(room("glade")); err != nil {
		t.Fatal(err)
	}
	m := NewManager()
	graphClone, err := m.Enter(svc, "cellar", "R0201", "R0210")
	if err != nil {
		t.Fatal(err)
	}
	gen, err := m.Generate(svc, "walker", 1, ProcSpec{
		TemplateIDs:  []string{"glade"},
		Count:        1,
		ReturnRoomID: "R0201",
		Seed:         1,
		Timeout:      time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	deleted := m.Expire(svc, time.Now().Add(time.Hour))
	if len(deleted) != 1 || deleted[0] != gen.EntryRoomID {
		t.Fatalf("expired %v", deleted)
	}
	if !m.IsClone(graphClone) {
		t.Fatal("authored instance was expired")
	}
	if _, err := svc.FindByID(graphClone); err != nil {
		t.Fatal(err)
	}
}
