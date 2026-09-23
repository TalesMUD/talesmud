package door

import (
	"path/filepath"
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

func TestEnsureNewDayHealsOncePerCalendarDay(t *testing.T) {
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
	if !pack.HealOnNewDay {
		t.Fatal("pack HealOnNewDay should default true")
	}

	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}
	dayA := time.Date(2026, 9, 22, 22, 0, 0, 0, berlin)
	dayB := time.Date(2026, 9, 23, 8, 0, 0, 0, berlin)

	store := daily.New(client.DB(), berlin)
	store.SetNow(func() time.Time { return dayA })
	hub := NewHub(facade, store, pack, pack.DailyFights, berlin)
	hub.SetNow(func() time.Time { return dayA })

	user, err := facade.UsersService().Create(&entities.User{
		RefID:    "local:atla-test",
		Username: "atla-test",
		Email:    "atla-test@example.com",
		Nickname: "atla-test",
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
		frames = append(frames, frame)
	}

	hub.OnConnect(user, send)
	hub.OnInput(user, "Atla", send)
	hub.OnInput(user, "m", send)
	hub.OnInput(user, "1", send)
	if user.LastCharacter == "" {
		t.Fatal("no character")
	}

	// Simulate yesterday's damage with no LastDay (migrated play state).
	if err := facade.CharactersService().Modify(user.LastCharacter, func(ch *characters.Character) error {
		ch.CurrentHitPoints = 4
		if ch.Door == nil {
			ch.Door = &characters.DoorProfile{}
		}
		ch.Door.LastDay = ""
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	// Exhaust forest fights on day A so rollover is visible.
	if _, err := store.Consume(user.LastCharacter, daily.ForestFightsKey, pack.DailyFights, pack.DailyFights); err != nil {
		t.Fatal(err)
	}

	// Cross midnight: new calendar day must full-heal and refill fights.
	store.SetNow(func() time.Time { return dayB })
	hub.SetNow(func() time.Time { return dayB })
	hub.OnDisconnect(user)
	frames = nil
	reloaded, err := facade.UsersService().FindByID(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	hub.OnConnect(reloaded, send)

	healed, err := facade.CharactersService().FindByID(reloaded.LastCharacter)
	if err != nil {
		t.Fatal(err)
	}
	if healed.CurrentHitPoints != healed.MaxHitPoints {
		t.Fatalf("new-day HP = %d/%d", healed.CurrentHitPoints, healed.MaxHitPoints)
	}
	if healed.Door == nil || healed.Door.LastDay != "2026-09-23" {
		t.Fatalf("LastDay = %+v", healed.Door)
	}
	plain := stripANSI(frames[len(frames)-1].ANSI)
	if !strings.Contains(plain, "new day dawns") {
		t.Fatalf("missing new-day notice:\n%s", plain)
	}
	fights, err := store.Get(reloaded.LastCharacter, daily.ForestFightsKey, pack.DailyFights)
	if err != nil {
		t.Fatal(err)
	}
	if fights.Day != "2026-09-23" || fights.Remaining != pack.DailyFights {
		t.Fatalf("forest fights after rollover = %+v", fights)
	}

	// Same-day damage must not re-heal on reconnect.
	if err := facade.CharactersService().Modify(reloaded.LastCharacter, func(ch *characters.Character) error {
		ch.CurrentHitPoints = 4
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	hub.OnDisconnect(reloaded)
	frames = nil
	again, err := facade.UsersService().FindByID(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	hub.OnConnect(again, send)
	same, err := facade.CharactersService().FindByID(again.LastCharacter)
	if err != nil {
		t.Fatal(err)
	}
	if same.CurrentHitPoints != 4 {
		t.Fatalf("same-day reconnect re-healed to %d", same.CurrentHitPoints)
	}
	plain = stripANSI(frames[len(frames)-1].ANSI)
	if strings.Contains(plain, "new day dawns") {
		t.Fatalf("unexpected new-day notice on same-day relog:\n%s", plain)
	}
}
