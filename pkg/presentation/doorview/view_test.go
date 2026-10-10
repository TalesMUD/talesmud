package doorview

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/gamemode"
	"github.com/talesmud/talesmud/pkg/mudserver/game"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
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
	plain := stripANSI(frame.ANSI)
	if frame.Type != "door_frame" || !strings.Contains(plain, "Market Square") || !strings.Contains(plain, "HP 11/20") {
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
	if !strings.Contains(frame.ANSI, "Choose a path") || !strings.Contains(frame.ANSI, "1 Fenwatch") {
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
	entered := stripANSI(frame.ANSI)
	if !strings.Contains(entered, "Market Square") || !strings.Contains(entered, "N North") || !strings.Contains(entered, "Actions: news") {
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
	if frame.Keys["f"] != "news" || frame.Keys["d"] != "" || !strings.Contains(stripANSI(frame.ANSI), "F News") {
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
	view.OnNotice(user, "The healer asks 40 coin.", "message", messages.LastNoticeGen(user.ID), func(msg any) {
		if f, ok := msg.(ansi.Frame); ok {
			frame = f
		}
	})
	if !strings.Contains(frame.ANSI, "The healer asks 40 coin.") {
		t.Fatalf("notice was not painted:\n%s", frame.ANSI)
	}
}

func TestFitBodyKeepsTheLatestLine(t *testing.T) {
	body := []string{"Level 1   HP 25/25   Gold 50"}
	for i := 0; i < 24; i++ {
		body = append(body, "picture")
	}
	out := fitBody([]string{body[0]}, body[1:], []string{"Vault 0 coin. On hand 50."}, 19)
	if len(out) != 19 {
		t.Fatalf("len %d", len(out))
	}
	if out[0] != "Level 1   HP 25/25   Gold 50" {
		t.Fatalf("status dropped: %q", out[0])
	}
	if out[len(out)-1] != "Vault 0 coin. On hand 50." {
		t.Fatalf("service line dropped: %q", out[len(out)-1])
	}
	log := make([]string, 0, 31)
	for i := 0; i < 30; i++ {
		log = append(log, fmt.Sprintf("hit %d", i))
	}
	log = append(log, "VICTORY!")
	status := "You 20/25   Wolf L1 10/30"
	out = fitBody([]string{"Level 1   HP 25/25   Gold 50", status}, []string{"picture", "Exits: north"}, log, 19)
	if len(out) != 19 || out[0] != "Level 1   HP 25/25   Gold 50" || out[1] != status || out[len(out)-1] != "VICTORY!" {
		t.Fatalf("long fight log = %v", out)
	}
	if strings.Contains(strings.Join(out, "\n"), "picture") {
		t.Fatalf("picture stayed in front of the log: %v", out)
	}
}

func TestClearRecentDropsTheCombatLog(t *testing.T) {
	v := &View{}
	v.pushRecent("user", []string{"VICTORY!", "You attack Wolf."})
	v.clearRecent("user")
	if got := v.peekRecent("user"); len(got) != 0 {
		t.Fatalf("recent survived: %v", got)
	}
}

func TestNotableLinesKeepTheVictory(t *testing.T) {
	text := "\n═══════════════════════════════════════════════════\n              VICTORY!\n═══════════════════════════════════════════════════\n\nDefeated: Thorn Choir\n\nREWARDS:\n  + 12 XP\n  + 23 Gold\n"
	lines := notableLines(text)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "VICTORY!") || !strings.Contains(joined, "+ 12 XP") || !strings.Contains(joined, "+ 23 Gold") || !strings.Contains(joined, "Defeated: Thorn Choir") {
		t.Fatalf("lines = %v", lines)
	}
	if strings.Contains(joined, "════") {
		t.Fatalf("box rule kept: %v", lines)
	}
}

func TestFightLogStaysUntilTheLeaveKey(t *testing.T) {
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
		Level:            1,
		MaxHitPoints:     25,
		CurrentHitPoints: 20,
		InCombat:         true,
	})
	if err != nil {
		t.Fatal(err)
	}
	user := &entities.User{Entity: &entities.Entity{ID: "user-1"}, RefID: "user-1", LastCharacter: created.ID}
	view := &View{Game: game.New(facade), Title: "Sample"}
	var frame ansi.Frame
	sink := func(msg any) {
		if f, ok := msg.(ansi.Frame); ok {
			frame = f
		}
	}
	view.OnNotice(user, "Thorn Choir hits you for 4 damage.", "combatAction", 0, sink)
	view.OnNotice(user, "\nVICTORY!\nDefeated: Thorn Choir\n  + 12 XP\n  + 23 Gold\n", "combatEnd", 0, sink)
	if !strings.Contains(frame.ANSI, "VICTORY!") || !strings.Contains(frame.ANSI, "+ 12 XP") || !strings.Contains(frame.ANSI, "hits you") {
		t.Fatalf("fight log missing:\n%s", frame.ANSI)
	}
	view.OnInput(user, "a", sink)
	if !strings.Contains(frame.ANSI, "VICTORY!") || !strings.Contains(frame.ANSI, "hits you") {
		t.Fatalf("in-combat key cleared the log:\n%s", frame.ANSI)
	}
	drainDoor(view.Game)
	stored, err := facade.CharactersService().FindByID(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	stored.InCombat = false
	if err := facade.CharactersService().Update(stored.ID, stored); err != nil {
		t.Fatal(err)
	}
	view.OnInput(user, "a", sink)
	if strings.Contains(frame.ANSI, "VICTORY!") || strings.Contains(frame.ANSI, "hits you") || strings.Contains(frame.ANSI, "Nobody here to fight.") {
		t.Fatalf("extra attack after the win kept the log:\n%s", frame.ANSI)
	}
	for _, msg := range drainDoor(view.Game) {
		if rsp, ok := msg.(interface{ GetMessage() string }); ok && strings.Contains(rsp.GetMessage(), "Nobody here") {
			t.Fatalf("extra attack was sent: %s", rsp.GetMessage())
		}
	}
	view.OnNotice(user, "\nVICTORY!\nDefeated: Thorn Choir\n  + 12 XP\n", "combatEnd", 0, sink)
	view.OnInput(user, "", sink)
	if strings.Contains(frame.ANSI, "VICTORY!") {
		t.Fatalf("enter kept the fight log:\n%s", frame.ANSI)
	}
}

func TestStaleFlashIsDropped(t *testing.T) {
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
	created, err := facade.CharactersService().Store(&characters.Character{
		Entity:      &entities.Entity{ID: "hero"},
		Name:        "Hero",
		BelongsUser: *traits.BelongsToUser("user-1"),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "square"},
		Level:       1,
	})
	if err != nil {
		t.Fatal(err)
	}
	user := &entities.User{Entity: &entities.Entity{ID: "user-1"}, RefID: "user-1", LastCharacter: created.ID}
	view := &View{Game: game.New(facade), Title: "Sample"}
	view.setFloor(user.ID, 5)
	var frame ansi.Frame
	sink := func(msg any) {
		if f, ok := msg.(ansi.Frame); ok {
			frame = f
		}
	}
	view.OnConnect(user, sink)
	frame = ansi.Frame{}
	view.OnNotice(user, "old stats", "message", 5, sink)
	if strings.Contains(frame.ANSI, "old stats") {
		t.Fatalf("stale line painted:\n%s", frame.ANSI)
	}
	view.OnNotice(user, "new stats", "message", 6, sink)
	if !strings.Contains(frame.ANSI, "new stats") {
		t.Fatalf("fresh line missing:\n%s", frame.ANSI)
	}
}

func TestCompassExitBeatsTheMenuBind(t *testing.T) {
	root := t.TempDir()
	body := []byte("areas:\n  market:\n    n:\n      command: news\n      label: News\n")
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
		Entity:  &entities.Entity{ID: "square"},
		Name:    "Market Square",
		Area:    "market",
		Exits:   &rooms.Exits{{Name: "north", Target: "lane"}},
		Actions: &rooms.Actions{{Name: "news", Type: rooms.RoomActionTypeResponse, Response: "A notice on the board."}},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity: &entities.Entity{ID: "lane"},
		Name:   "North Lane",
		Area:   "market",
		Exits:  &rooms.Exits{{Name: "south", Target: "square"}},
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
	view := &View{Game: game.New(facade), Title: "Sample"}
	var frame ansi.Frame
	view.OnInput(user, "n", func(msg any) {
		if f, ok := msg.(ansi.Frame); ok {
			frame = f
		}
	})
	if !strings.Contains(frame.ANSI, "North Lane") {
		t.Fatalf("n did not take the north exit:\n%s", frame.ANSI)
	}
	for _, msg := range drainDoor(view.Game) {
		if rsp, ok := msg.(interface{ GetMessage() string }); ok && strings.Contains(rsp.GetMessage(), "notice") {
			t.Fatalf("n ran news: %s", rsp.GetMessage())
		}
	}
	view.OnInput(user, "e", func(msg any) {
		if f, ok := msg.(ansi.Frame); ok {
			frame = f
		}
	})
	if !strings.Contains(frame.ANSI, "No exit that way.") {
		t.Fatalf("missing east was silent:\n%s", frame.ANSI)
	}
}

func TestOpenWaysAreListedAndDownStaysOffTheFooter(t *testing.T) {
	room := &rooms.Room{Exits: &rooms.Exits{{Name: "east"}, {Name: "west"}}}
	binds := applyOpenExits(map[string]keyBind{}, room, nil)
	if binds["e"].Label != "East" || binds["w"].Label != "West" {
		t.Fatalf("ways = %#v", binds)
	}
	if commandFooter(room, nil) != ": command" {
		t.Fatalf("footer %q", commandFooter(room, nil))
	}
	down := &rooms.Room{Exits: &rooms.Exits{{Name: "down"}}}
	if commandFooter(down, nil) != "d down   : command" {
		t.Fatalf("down footer %q", commandFooter(down, nil))
	}
}

func TestStatsSheetShowsHitPointsGoldGearAndGems(t *testing.T) {
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
	created, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "hero"},
		Name:             "Hero",
		BelongsUser:      *traits.BelongsToUser("user-1"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "square"},
		Level:            4,
		MaxHitPoints:     41,
		CurrentHitPoints: 30,
		Gold:             1234,
		Flags:            map[string]interface{}{"gems": float64(3), "path": "Gravebound"},
		EquippedItems: map[items.ItemSlot]*items.Item{
			items.ItemSlotMainHand: {Name: "Hedge knife", Slot: items.ItemSlotMainHand, Attributes: map[string]interface{}{"damage": float64(2)}},
			items.ItemSlotChest:    {Name: "Travel cloak", Slot: items.ItemSlotChest, Attributes: map[string]interface{}{"defense": float64(1)}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	user := &entities.User{Entity: &entities.Entity{ID: "user-1"}, RefID: "user-1", LastCharacter: created.ID}
	view := &View{Game: game.New(facade), Title: "Sample"}
	var frame ansi.Frame
	view.OnInput(user, "stats", func(msg any) {
		if f, ok := msg.(ansi.Frame); ok {
			frame = f
		}
	})
	for _, want := range []string{"HP 30/41", "Gold 1234", "Gems 3", "Hedge knife", "Travel cloak", "Gravebound"} {
		if !strings.Contains(frame.ANSI, want) {
			t.Fatalf("stats missing %q:\n%s", want, frame.ANSI)
		}
	}
}

func TestStatusStripMatchesTheHeader(t *testing.T) {
	plain := stripANSI(statusStrip(3, 11, 20, 6, 40))
	if plain != headerLine(3, 11, 20, 6, 40) {
		t.Fatalf("strip %q header %q", plain, headerLine(3, 11, 20, 6, 40))
	}
	if !strings.Contains(statusStrip(1, 4, 20, 0, 0), "\x1b[1;31m") {
		t.Fatal("low hit points stayed green")
	}
	if !strings.Contains(statusStrip(1, 8, 20, 0, 0), "\x1b[1;33m") {
		t.Fatal("mid hit points were not yellow")
	}
	if !strings.Contains(statusStrip(1, 12, 20, 0, 0), "\x1b[1;32m") {
		t.Fatal("high hit points were not green")
	}
}

func TestInsetCatalogSkipsPictureRows(t *testing.T) {
	interior := "║" + strings.Repeat(" ", 20) + "║"
	picture := "║" + "████████" + strings.Repeat(" ", 12) + "║"
	art := strings.Join([]string{
		"████ ANVIL ████",
		picture,
		interior,
		interior,
		"╚════════════════════╝",
	}, "\n")
	out := insetCatalog(art, []string{"1 knife", "2 sword", "3 extra"})
	plain := stripANSI(strings.Join(out, "\n"))
	if !strings.Contains(plain, "1 knife") || !strings.Contains(plain, "2 sword") || !strings.Contains(plain, "3 extra") {
		t.Fatalf("catalog missing:\n%s", plain)
	}
	if !strings.Contains(plain, "████████") {
		t.Fatalf("picture row was overwritten:\n%s", plain)
	}
	if strings.Contains(out[1], "knife") {
		t.Fatal("picture row received an item")
	}
}

func TestRailsLayoutKeepsTheStatusWords(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "screens"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "screens", "layout.yaml"), []byte("style: rails\nart_rows: 12\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var art strings.Builder
	for i := 1; i <= 13; i++ {
		fmt.Fprintf(&art, "ROW%02d\n", i)
	}
	if err := os.WriteFile(filepath.Join(root, "screens", "square.ans"), []byte(art.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "keymap.yaml"), []byte("rooms:\n  square:\n    f:\n      command: news\n      label: News\n"), 0o644); err != nil {
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
	view := &View{Game: game.New(facade), Title: "Sample"}
	var frame ansi.Frame
	view.OnConnect(user, func(msg any) {
		if f, ok := msg.(ansi.Frame); ok {
			frame = f
		}
	})
	plain := stripANSI(frame.ANSI)
	if !strings.Contains(plain, "HP 11/20") || !strings.Contains(plain, "N North") || !strings.Contains(plain, "Stalls and steam.") {
		t.Fatalf("rails words:\n%s", plain)
	}
	if !strings.Contains(frame.ANSI, "▀") || strings.Contains(frame.ANSI, strings.Repeat("-", 40)) {
		t.Fatal("rails chrome missing")
	}
	if strings.Contains(plain, "ROW13") || !strings.Contains(plain, "ROW12") {
		t.Fatalf("art clip:\n%s", plain)
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
