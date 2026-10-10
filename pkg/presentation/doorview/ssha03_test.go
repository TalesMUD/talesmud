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

func TestDoorFrameScrubsPlayerText(t *testing.T) {
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "osc.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	desc := "Stall \x1b]52;c;SGVsbG8=\x07 \x1bPdcspayload\x1b\\ \u202eBIDI"
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity:      &entities.Entity{ID: "square"},
		Name:        "Hall\r\nFAKE \u202eRoom",
		Description: desc,
		Exits:       &rooms.Exits{{Name: "north\r\nHIDDEN \u202e", Target: "square"}},
	}); err != nil {
		t.Fatal(err)
	}
	created, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "hero"},
		Name:             "Hero\u202e",
		BelongsUser:      *traits.BelongsToUser("user-1"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "square"},
		Level:            1,
		MaxHitPoints:     10,
		CurrentHitPoints: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	user := &entities.User{Entity: &entities.Entity{ID: "user-1"}, RefID: "user-1", LastCharacter: created.ID}
	view := &View{Game: game.New(facade), Title: "Sample \x1b]0;x\x07"}
	gamemode.ApplyEnv()
	var frame ansi.Frame
	view.OnConnect(user, func(msg any) {
		if f, ok := msg.(ansi.Frame); ok {
			frame = f
		}
	})
	raw := frame.ANSI
	for _, bad := range []string{"\x1b]52", "\x1b]0;", "\x1bP", "\nFAKE", "\rFAKE"} {
		if strings.Contains(raw, bad) {
			t.Fatalf("frame kept %q\n%s", bad, raw)
		}
	}
	if strings.ContainsRune(raw, '\u202e') {
		t.Fatal("frame kept bidi")
	}
	if !strings.Contains(raw, "Stall") || !strings.Contains(raw, "north") {
		t.Fatalf("room text missing\n%s", raw)
	}
	view.OnNotice(user, "Ada says: \x1b]52;c;SGVsbG8=\x07 hello \u202e", "say", 1, func(msg any) {
		if f, ok := msg.(ansi.Frame); ok {
			frame = f
		}
	})
	if strings.Contains(frame.ANSI, "\x1b]52") || strings.ContainsRune(frame.ANSI, '\u202e') {
		t.Fatal("say text kept controls")
	}
	if !strings.Contains(frame.ANSI, "hello") {
		t.Fatal("say text missing")
	}
}
