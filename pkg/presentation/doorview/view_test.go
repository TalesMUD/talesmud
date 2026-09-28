package doorview

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/gamemode"
	"github.com/talesmud/talesmud/pkg/mudserver/game"
	"github.com/talesmud/talesmud/pkg/presentation/ansi"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/ruleset"
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
		Name:        "Market Square",
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
	if !strings.Contains(frame.ANSI, "Choose a path") || !strings.Contains(frame.ANSI, "1 Warrior") {
		t.Fatalf("name did not offer a path:\n%s", frame.ANSI)
	}
	view.OnInput(user, "1", func(msg any) {
		if f, ok := msg.(ansi.Frame); ok {
			frame = f
		}
	})
	if !strings.Contains(frame.ANSI, "M Man") {
		t.Fatalf("path did not offer sex:\n%s", frame.ANSI)
	}
	view.OnInput(user, "m", func(msg any) {
		if f, ok := msg.(ansi.Frame); ok {
			frame = f
		}
	})
	if !strings.Contains(frame.ANSI, "Market Square") || !strings.Contains(frame.ANSI, "Actions: news") {
		t.Fatalf("created character did not enter the start room:\n%s", frame.ANSI)
	}
	if user.LastCharacter == "" {
		t.Fatal("character was not selected")
	}
}

func TestConnectAppliesPendingNewDay(t *testing.T) {
	ruleset.SetNewDay(true, "UTC")
	t.Cleanup(ruleset.Reset)

	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "view.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity: &entities.Entity{ID: "square"},
		Name:   "Market Square",
		Exits:  &rooms.Exits{},
	}); err != nil {
		t.Fatal(err)
	}
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
	created, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "hero"},
		Name:             "Hero",
		BelongsUser:      *traits.BelongsToUser("user-1"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "square"},
		Level:            2,
		MaxHitPoints:     20,
		CurrentHitPoints: 4,
		AwaitingReset:    true,
		LastResetDay:     yesterday,
	})
	if err != nil {
		t.Fatal(err)
	}
	user := &entities.User{Entity: &entities.Entity{ID: "user-1"}, RefID: "user-1", LastCharacter: created.ID}
	view := &View{Game: game.New(facade), Title: "Sample"}
	view.OnConnect(user, func(any) {})
	stored, err := facade.CharactersService().FindByID(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	today := time.Now().UTC().Format("2006-01-02")
	if stored.CurrentHitPoints != 20 || stored.AwaitingReset || stored.LastResetDay != today {
		t.Fatalf("session start hp=%d reset=%v day=%s", stored.CurrentHitPoints, stored.AwaitingReset, stored.LastResetDay)
	}
}

func TestKeyMapBindsRoomAndLeavesDownAlone(t *testing.T) {
	root := t.TempDir()
	body := []byte("rooms:\n  square:\n    f:\n      command: news\n      label: News\n    d:\n      command: deposit\n      label: Deposit\n")
	if err := os.WriteFile(filepath.Join(root, "keymap.yaml"), body, 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(t.TempDir(), "mode.yaml")
	if err := os.WriteFile(cfg, []byte("world_pack: "+root+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	prev := gamemode.Current()
	if err := gamemode.ApplyFile(cfg); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		restore := filepath.Join(t.TempDir(), "restore.yaml")
		_ = os.WriteFile(restore, []byte("presentation: "+prev.Presentation+"\nworld_pack: \""+prev.WorldPack+"\"\n"), 0o644)
		_ = gamemode.ApplyFile(restore)
	})

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
		Exits:       &rooms.Exits{{Name: "down", Target: "square"}},
		Actions:     &rooms.Actions{{Name: "news", Type: rooms.RoomActionTypeResponse, Response: "A notice on the board."}},
	}); err != nil {
		t.Fatal(err)
	}
	created, err := facade.CharactersService().Store(&characters.Character{
		Entity:      &entities.Entity{ID: "hero"},
		Name:        "Hero",
		BelongsUser: *traits.BelongsToUser("user-1"),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "square"},
	})
	if err != nil {
		t.Fatal(err)
	}
	user := &entities.User{Entity: &entities.Entity{ID: "user-1"}, RefID: "user-1", LastCharacter: created.ID}
	g := game.New(facade)
	view := &View{Game: g, Title: "Sample"}
	var frame ansi.Frame
	view.OnConnect(user, func(msg any) {
		if f, ok := msg.(ansi.Frame); ok {
			frame = f
		}
	})
	if frame.Keys["f"] != "news" || frame.Keys["d"] != "" || !strings.Contains(frame.ANSI, "F News") {
		t.Fatalf("keys=%v frame:\n%s", frame.Keys, frame.ANSI)
	}
	view.OnInput(user, "f", func(msg any) {
		if f, ok := msg.(ansi.Frame); ok {
			frame = f
		}
	})
	saw := false
	for _, msg := range drainDoor(g) {
		if rsp, ok := msg.(interface{ GetMessage() string }); ok && strings.Contains(rsp.GetMessage(), "notice") {
			saw = true
		}
	}
	if !saw {
		t.Fatal("f did not run the news action")
	}
	view.OnNotice(user, "The healer asks 40 coin.", func(msg any) {
		if f, ok := msg.(ansi.Frame); ok {
			frame = f
		}
	})
	if !strings.Contains(frame.ANSI, "The healer asks 40 coin.") {
		t.Fatalf("notice was not painted:\n%s", frame.ANSI)
	}
}

func drainDoor(g *game.Game) []any {
	var out []any
	for {
		select {
		case msg := <-g.SendMessage():
			out = append(out, msg)
		default:
			return out
		}
	}
}
