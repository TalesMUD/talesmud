package door

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/daily"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/scripts/runner"
	"github.com/talesmud/talesmud/pkg/service"
)

type scriptRoll struct {
	seq []int
	i   int
}

func (s *scriptRoll) Intn(n int) int {
	if n <= 1 {
		return 0
	}
	v := n - 1
	if s.i < len(s.seq) {
		v = s.seq[s.i]
		s.i++
		if v >= n {
			v = n - 1
		}
		if v < 0 {
			v = 0
		}
	}
	return v
}

func TestTownFightHealBankSurviveRelog(t *testing.T) {
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "door.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close() })

	facade := service.NewFacade(repository.NewSQLiteFactory(client), runner.NewMultiRunner())
	pack, err := Load(filepath.Join(moduleRoot(t), "..", "talesmud-door", "worlds", "aethermoor-door"))
	if err != nil {
		t.Fatal(err)
	}
	store := daily.New(client.DB(), time.UTC)
	hub := NewHub(facade, store, pack, pack.DailyFights, time.UTC)
	hub.SetRoller(&scriptRoll{seq: []int{0, 19, 3}})
	hub.SetNow(func() time.Time { return time.Date(2026, 9, 21, 15, 4, 0, 0, time.UTC) })

	created, err := facade.UsersService().Create(&entities.User{
		RefID:    "local:mara",
		Username: "mara",
		Email:    "mara@example.com",
		Nickname: "mara",
		Role:     entities.RolePlayer,
	})
	if err != nil {
		t.Fatal(err)
	}

	var frames []Frame
	send := func(v any) {
		frame, ok := v.(Frame)
		if !ok {
			t.Fatalf("sent %T", v)
		}
		assertGrid(t, frame)
		frames = append(frames, frame)
	}

	hub.OnConnect(created, send)
	hub.OnInput(created, "Mara Quinn", send)
	hub.OnInput(created, "f", send)
	hub.OnInput(created, "1", send)
	if created.LastCharacter == "" {
		t.Fatal("warrior was not stored on the user")
	}
	if plain := stripANSI(frames[len(frames)-1].ANSI); !strings.Contains(plain, "Town Square") && !strings.Contains(plain, "AETHERMOOR DOOR") {
		t.Fatalf("expected town after creation:\n%s", plain)
	}
	if frames[len(frames)-1].Type != "door_frame" || frames[len(frames)-1].Screen != "town_square" {
		t.Fatalf("frame = %+v", frames[len(frames)-1].Screen)
	}

	if err := facade.CharactersService().Modify(created.LastCharacter, func(ch *characters.Character) error {
		ch.CurrentHitPoints = 5
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	hub.OnInput(created, "h", send)
	hub.OnInput(created, "h", send)
	hub.OnInput(created, "t", send)
	healed, err := facade.CharactersService().FindByID(created.LastCharacter)
	if err != nil {
		t.Fatal(err)
	}
	if healed.CurrentHitPoints != healed.MaxHitPoints {
		t.Fatalf("HP = %d/%d", healed.CurrentHitPoints, healed.MaxHitPoints)
	}
	if healed.Gold != starterGold-int64(pack.HealCost) {
		t.Fatalf("gold after heal = %d", healed.Gold)
	}

	hub.OnInput(created, "b", send)
	hub.OnInput(created, "d", send)
	if frames[len(frames)-1].InputMode != "line" {
		t.Fatal("deposit did not switch to line input")
	}
	hub.OnInput(created, "15", send)
	banked, err := facade.CharactersService().FindByID(created.LastCharacter)
	if err != nil {
		t.Fatal(err)
	}
	if banked.Door == nil || banked.Door.BankGold != 15 {
		t.Fatalf("bank = %+v", banked.Door)
	}
	hub.OnInput(created, "t", send)

	before, err := store.Get(created.LastCharacter, daily.ForestFightsKey, pack.DailyFights)
	if err != nil {
		t.Fatal(err)
	}
	hub.OnInput(created, "f", send)
	if frames[len(frames)-1].Screen != "forest" {
		t.Fatalf("forest screen = %s", frames[len(frames)-1].Screen)
	}
	hub.OnInput(created, "a", send)
	after, err := store.Get(created.LastCharacter, daily.ForestFightsKey, pack.DailyFights)
	if err != nil {
		t.Fatal(err)
	}
	if after.Remaining != before.Remaining-1 {
		t.Fatalf("fights %d -> %d", before.Remaining, after.Remaining)
	}
	won, err := facade.CharactersService().FindByID(created.LastCharacter)
	if err != nil {
		t.Fatal(err)
	}
	if won.Door == nil || won.Door.Wins != 1 || won.Gold <= banked.Gold {
		t.Fatalf("win did not pay: gold %d wins %+v", won.Gold, won.Door)
	}
	if won.Door.BankGold != 15 {
		t.Fatalf("vault changed during the fight: %d", won.Door.BankGold)
	}
	hub.OnInput(created, "y", send)
	if stats := stripANSI(frames[len(frames)-1].ANSI); !strings.Contains(stats, "Mara Quinn") || !strings.Contains(stats, "Others on the door") {
		t.Fatalf("stats/leaderboard stub missing:\n%s", stats)
	}
	hub.OnDisconnect(created)
	reloaded, err := facade.UsersService().FindByID(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	frames = nil
	hub.OnConnect(reloaded, send)
	plain := stripANSI(frames[len(frames)-1].ANSI)
	if !strings.Contains(plain, "vault 15") || !strings.Contains(plain, "walks 24/25") {
		t.Fatalf("relog frame lost state:\n%s", plain)
	}
	if strings.Contains(plain, "Legend of the Red Dragon") {
		t.Fatal("town copy used a locked title")
	}
}

func TestPackCopyIsOriginal(t *testing.T) {
	root := filepath.Join(moduleRoot(t), "..", "talesmud-door", "worlds", "aethermoor-door")
	var texts []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		texts = append(texts, string(raw))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range texts {
		for _, banned := range []string{"Legend of the Red Dragon", "Seth Able", "Turgon", "Abdul"} {
			if strings.Contains(text, banned) {
				t.Fatalf("pack contains %q", banned)
			}
		}
	}
}

func assertGrid(t *testing.T, frame Frame) {
	t.Helper()
	if !strings.HasPrefix(frame.ANSI, "\x1b[2J\x1b[H") {
		t.Fatal("missing clear")
	}
	lines := strings.Split(strings.TrimPrefix(frame.ANSI, "\x1b[2J\x1b[H"), "\r\n")
	if len(lines) != Rows {
		t.Fatalf("rows = %d", len(lines))
	}
	for i, ln := range lines {
		if n := visibleLen(ln); n != Cols {
			t.Fatalf("line %d width %d: %q", i, n, stripANSI(ln))
		}
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
