package game

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func TestTwoGuestsShareTownAndGetPrivateCellars(t *testing.T) {
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "R0201", rooms.Exits{
		{Name: "down", Target: "R0210", Type: "instance", Instance: true},
	})
	storeTestRoom(t, facade, "R0210", rooms.Exits{
		{Name: "up", Target: "R0201"},
	})

	mk := func(id, name string) *characters.Character {
		ch := &characters.Character{
			Entity:      &entities.Entity{ID: id},
			Name:        name,
			CurrentRoom: traits.CurrentRoom{CurrentRoomID: "R0201"},
		}
		if _, err := facade.CharactersService().Store(ch); err != nil {
			t.Fatal(err)
		}
		return ch
	}
	a := mk("guest-a", "GuestA")
	b := mk("guest-b", "GuestB")
	userA := &entities.User{Entity: &entities.Entity{ID: "user-a"}}
	userB := &entities.User{Entity: &entities.Entity{ID: "user-b"}}

	hub, err := facade.RoomsService().FindByID("R0201")
	if err != nil {
		t.Fatal(err)
	}
	hub.AddCharacter(a.ID)
	hub.AddCharacter(b.ID)
	_ = facade.RoomsService().Update("R0201", hub)

	take := commands.TakeExit("down")
	if !take(hub, g, &messages.Message{FromUser: userA, Character: a, Data: "down"}) {
		t.Fatal("guest A could not enter cellar")
	}
	if !take(hub, g, &messages.Message{FromUser: userB, Character: b, Data: "down"}) {
		t.Fatal("guest B could not enter cellar")
	}

	freshA, _ := facade.CharactersService().FindByID(a.ID)
	freshB, _ := facade.CharactersService().FindByID(b.ID)
	if freshA.CurrentRoomID == "R0201" || freshB.CurrentRoomID == "R0201" {
		t.Fatal("guests should be in cellars")
	}
	if freshA.CurrentRoomID == freshB.CurrentRoomID {
		t.Fatalf("cellars must be private, both in %s", freshA.CurrentRoomID)
	}
	if freshA.CurrentRoomID == "R0210" || freshB.CurrentRoomID == "R0210" {
		t.Fatal("must not share the template cellar")
	}

	hubNow, _ := facade.RoomsService().FindByID("R0201")
	if hubNow.ID != "R0201" {
		t.Fatal("town hub must remain the shared room")
	}
}

func TestTwoGuestsShareInnAndGetPrivateTaggedCellars(t *testing.T) {
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "R0203", rooms.Exits{
		{Name: "down", Target: "R0215", Type: "direction"},
	})
	storeTestRoom(t, facade, "R0215", rooms.Exits{
		{Name: "up", Target: "R0203"},
		{Name: "deeper", Target: "R0230", Hidden: true},
	})
	storeTestRoom(t, facade, "R0230", rooms.Exits{
		{Name: "back", Target: "R0215"},
	})
	tag := func(id string, tags ...string) {
		t.Helper()
		r, err := facade.RoomsService().FindByID(id)
		if err != nil {
			t.Fatal(err)
		}
		r.Tags = tags
		if err := facade.RoomsService().Update(id, r); err != nil {
			t.Fatal(err)
		}
	}
	tag("R0203", "shared", "inn")
	tag("R0215", "instance", "cellar")
	tag("R0230", "instance", "hidden")

	mk := func(id, name string) *characters.Character {
		ch := &characters.Character{
			Entity:      &entities.Entity{ID: id},
			Name:        name,
			CurrentRoom: traits.CurrentRoom{CurrentRoomID: "R0203"},
		}
		if _, err := facade.CharactersService().Store(ch); err != nil {
			t.Fatal(err)
		}
		return ch
	}
	a := mk("guest-a", "GuestA")
	b := mk("guest-b", "GuestB")
	userA := &entities.User{Entity: &entities.Entity{ID: "user-a"}}
	userB := &entities.User{Entity: &entities.Entity{ID: "user-b"}}

	hub, err := facade.RoomsService().FindByID("R0203")
	if err != nil {
		t.Fatal(err)
	}
	hub.AddCharacter(a.ID)
	hub.AddCharacter(b.ID)
	_ = facade.RoomsService().Update("R0203", hub)

	take := commands.TakeExit("down")
	if !take(hub, g, &messages.Message{FromUser: userA, Character: a, Data: "down"}) {
		t.Fatal("guest A could not enter tagged cellar")
	}
	if !take(hub, g, &messages.Message{FromUser: userB, Character: b, Data: "down"}) {
		t.Fatal("guest B could not enter tagged cellar")
	}

	freshA, _ := facade.CharactersService().FindByID(a.ID)
	freshB, _ := facade.CharactersService().FindByID(b.ID)
	if freshA.CurrentRoomID == freshB.CurrentRoomID {
		t.Fatalf("cellars must be private, both in %s", freshA.CurrentRoomID)
	}
	if freshA.CurrentRoomID == "R0215" || freshB.CurrentRoomID == "R0215" {
		t.Fatal("must not share the template cellar")
	}
	if freshA.CurrentRoomID[:6] != "R0215~" || freshB.CurrentRoomID[:6] != "R0215~" {
		t.Fatalf("expected R0215~ clones, got %s and %s", freshA.CurrentRoomID, freshB.CurrentRoomID)
	}

	hubNow, _ := facade.RoomsService().FindByID("R0203")
	if hubNow.ID != "R0203" {
		t.Fatal("inn hub must remain the shared room")
	}
	if clone, err := facade.RoomsService().FindByID(freshA.CurrentRoomID); err != nil || clone == nil {
		t.Fatal("guest A clone missing")
	} else {
		deeper, ok := clone.GetExit("deeper")
		if !ok || deeper.Target == "R0230" || deeper.Target[:6] != "R0230~" {
			t.Fatalf("hidden wing should be cloned, deeper=%v ok=%v", deeper, ok)
		}
	}
}
