package doorview

import (
	"path/filepath"
	"strings"
	"testing"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/gamemode"
	"github.com/talesmud/talesmud/pkg/mudserver/game"
	"github.com/talesmud/talesmud/pkg/presentation/ansi"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

func TestViewPaintsTheLiveRoom(t *testing.T) {
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "view.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity:      &entities.Entity{ID: "square"},
		Name:        "Market Square",
		Description: "Stalls and steam.",
		Exits:       &rooms.Exits{{Name: "north", Target: "square"}},
	}); err != nil {
		t.Fatal(err)
	}
	created, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "hero"},
		Name:             "Hero",
		BelongsUser:      *traits.BelongsToUser("user-1"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "square"},
		Level:            3,
		MaxHitPoints:     20,
		CurrentHitPoints: 11,
		Gold:             6,
	})
	if err != nil {
		t.Fatal(err)
	}
	user := &entities.User{Entity: &entities.Entity{ID: "user-1"}, RefID: "user-1", LastCharacter: created.ID}
	g := game.New(facade)
	view := &View{Game: g, Title: "Sample"}
	gamemode.ApplyEnv()
	var frame ansi.Frame
	view.OnConnect(user, func(msg any) {
		if f, ok := msg.(ansi.Frame); ok {
			frame = f
		}
	})
	if frame.Type != "door_frame" || !strings.Contains(frame.ANSI, "Market Square") || !strings.Contains(frame.ANSI, "HP 11/20") {
		t.Fatalf("frame = %q", frame.ANSI)
	}
	view.OnInput(user, "l", func(msg any) {
		if f, ok := msg.(ansi.Frame); ok {
			frame = f
		}
	})
	if !strings.Contains(frame.ANSI, "Stalls and steam.") {
		t.Fatalf("look did not repaint the room:\n%s", frame.ANSI)
	}
}

func TestViewNamePromptCreatesAndSelects(t *testing.T) {
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "view.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity:      &entities.Entity{ID: "R0001"},
		Name:        "Ashmarket Square",
		Description: "Brass skyline.",
		Exits:       &rooms.Exits{{Name: "north", Target: "R0001"}},
		Actions:     &rooms.Actions{{Name: "news", Type: rooms.RoomActionTypeResponse, Description: "Read the notices.", Response: "A notice."}},
	}); err != nil {
		t.Fatal(err)
	}
	user := &entities.User{Entity: &entities.Entity{ID: "user-2"}, RefID: "user-2"}
	view := &View{Game: game.New(facade), Title: "Sample"}
	var frame ansi.Frame
	view.OnConnect(user, func(msg any) {
		if f, ok := msg.(ansi.Frame); ok {
			frame = f
		}
	})
	if frame.InputMode != "line" || !strings.Contains(frame.ANSI, "No characters yet.") {
		t.Fatalf("prompt = %q mode %s", frame.ANSI, frame.InputMode)
	}
	view.OnInput(user, "Bram", func(msg any) {
		if f, ok := msg.(ansi.Frame); ok {
			frame = f
		}
	})
	if !strings.Contains(frame.ANSI, "Ashmarket Square") || !strings.Contains(frame.ANSI, "Actions: news") {
		t.Fatalf("created character did not enter the start room:\n%s", frame.ANSI)
	}
	if user.LastCharacter == "" {
		t.Fatal("character was not selected")
	}
}
