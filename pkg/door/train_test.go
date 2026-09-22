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
	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/scripts/runner"
	"github.com/talesmud/talesmud/pkg/service"
)

type trainSink struct {
	frames []Frame
}

func (s *trainSink) send(v any) {
	if f, ok := v.(Frame); ok {
		s.frames = append(s.frames, f)
	}
}

func (s *trainSink) last() Frame {
	if len(s.frames) == 0 {
		return Frame{}
	}
	return s.frames[len(s.frames)-1]
}

func (s *trainSink) plain() string {
	return stripANSI(s.last().ANSI)
}

func TestTrainStubGoneAndEligibleWinLevels(t *testing.T) {
	hub, user, sink, facade := trainFixture(t, &scriptRoll{seq: critWinSeq(80)})
	createDoorWarrior(t, hub, user, sink.send, "Train Win")

	need := leveling.GetXPRequired(2)
	if err := facade.CharactersService().Modify(user.LastCharacter, func(ch *characters.Character) error {
		ch.XP = need
		ch.Level = 1
		ch.CurrentHitPoints = ch.MaxHitPoints
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	hub.OnInput(user, "m", sink.send)
	if sink.last().Screen != "trainer" {
		t.Fatalf("screen after M = %s", sink.last().Screen)
	}
	hub.OnInput(user, "t", sink.send)
	if sink.last().Screen != "fight" {
		t.Fatalf("expected fight, got %s\n%s", sink.last().Screen, sink.plain())
	}
	if strings.Contains(sink.plain(), "Duels for level come next") {
		t.Fatal("train stub text still shown")
	}
	if !strings.Contains(sink.plain(), "Ashmarket Master") {
		t.Fatalf("master missing:\n%s", sink.plain())
	}

	winMasterDuel(t, hub, user, sink)
	ch, err := facade.CharactersService().FindByID(user.LastCharacter)
	if err != nil {
		t.Fatal(err)
	}
	if ch.Level != 2 {
		t.Fatalf("level = %d want 2 (xp=%d)", ch.Level, ch.XP)
	}
	if ch.CurrentHitPoints != ch.MaxHitPoints {
		t.Fatalf("HP not restored: %d/%d", ch.CurrentHitPoints, ch.MaxHitPoints)
	}

	hub.OnInput(user, "c", sink.send) // continue → trainer
	hub.OnInput(user, "r", sink.send) // town
	hub.OnInput(user, "y", sink.send) // stats
	stats := sink.plain()
	if !strings.Contains(stats, "Level 2") {
		t.Fatalf("stats missing Level 2:\n%s", stats)
	}
}

func TestTrainIneligibleWinNoLevel(t *testing.T) {
	hub, user, sink, facade := trainFixture(t, &scriptRoll{seq: critWinSeq(80)})
	createDoorWarrior(t, hub, user, sink.send, "Train Early")

	if err := facade.CharactersService().Modify(user.LastCharacter, func(ch *characters.Character) error {
		ch.XP = 10
		ch.Level = 1
		ch.CurrentHitPoints = ch.MaxHitPoints
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	hub.OnInput(user, "m", sink.send)
	hub.OnInput(user, "t", sink.send)
	winMasterDuel(t, hub, user, sink)

	ch, err := facade.CharactersService().FindByID(user.LastCharacter)
	if err != nil {
		t.Fatal(err)
	}
	if ch.Level != 1 {
		t.Fatalf("level = %d want 1", ch.Level)
	}
	plain := sink.plain()
	if !strings.Contains(plain, "not ready") && !strings.Contains(plain, "Ashwood") {
		t.Fatalf("expected not-ready notice:\n%s", plain)
	}
}

func TestTrainLossNoLevel(t *testing.T) {
	hub, user, sink, facade := trainFixture(t, &scriptRoll{seq: loseSeq(40)})
	createDoorWarrior(t, hub, user, sink.send, "Train Loss")

	need := leveling.GetXPRequired(2)
	if err := facade.CharactersService().Modify(user.LastCharacter, func(ch *characters.Character) error {
		ch.XP = need
		ch.Level = 1
		ch.CurrentHitPoints = 4
		ch.MaxHitPoints = 16
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	hub.OnInput(user, "m", sink.send)
	hub.OnInput(user, "t", sink.send)
	for i := 0; i < 30; i++ {
		hub.OnInput(user, "a", sink.send)
		ch, err := facade.CharactersService().FindByID(user.LastCharacter)
		if err != nil {
			t.Fatal(err)
		}
		sess := hub.session(user.ID)
		done := sess != nil && sess.fight != nil && sess.fight.Done
		if done || sink.last().Screen == "result" {
			if ch.Level != 1 {
				t.Fatalf("loss leveled to %d", ch.Level)
			}
			if ch.Door == nil || ch.Door.Losses < 1 {
				t.Fatalf("expected a loss recorded: %+v", ch.Door)
			}
			return
		}
		if ch.Level != 1 {
			t.Fatalf("level changed mid-fight to %d", ch.Level)
		}
	}
	t.Fatal("duel did not end in a loss")
}

func TestTrainCapBlocks(t *testing.T) {
	hub, user, sink, facade := trainFixture(t, &scriptRoll{seq: []int{0}})
	createDoorWarrior(t, hub, user, sink.send, "Train Cap")

	if err := facade.CharactersService().Modify(user.LastCharacter, func(ch *characters.Character) error {
		ch.Level = doorLevelCap
		ch.XP = leveling.GetXPRequired(doorLevelCap)
		ch.MaxLevelCap = doorLevelCap
		ch.CurrentHitPoints = ch.MaxHitPoints
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	hub.OnInput(user, "m", sink.send)
	hub.OnInput(user, "t", sink.send)
	if sink.last().Screen == "fight" {
		t.Fatal("cap should not start a duel")
	}
	if !strings.Contains(sink.plain(), "nothing left to teach") {
		t.Fatalf("cap notice missing:\n%s", sink.plain())
	}
	ch, err := facade.CharactersService().FindByID(user.LastCharacter)
	if err != nil {
		t.Fatal(err)
	}
	if ch.Level != doorLevelCap {
		t.Fatalf("level = %d", ch.Level)
	}
}

func TestForestWinDoesNotAutoLevel(t *testing.T) {
	hub, user, sink, facade := trainFixture(t, &scriptRoll{seq: critWinSeq(80)})
	createDoorWarrior(t, hub, user, sink.send, "Forest XP")

	need := leveling.GetXPRequired(2)
	before := need - 5
	if err := facade.CharactersService().Modify(user.LastCharacter, func(ch *characters.Character) error {
		ch.XP = before
		ch.Level = 1
		ch.CurrentHitPoints = ch.MaxHitPoints
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	hub.OnInput(user, "f", sink.send)
	won := false
	for i := 0; i < 40; i++ {
		hub.OnInput(user, "a", sink.send)
		ch, err := facade.CharactersService().FindByID(user.LastCharacter)
		if err != nil {
			t.Fatal(err)
		}
		if ch.Door != nil && ch.Door.Wins >= 1 {
			won = true
			if ch.Level != 1 {
				t.Fatalf("forest auto-leveled to %d (xp=%d)", ch.Level, ch.XP)
			}
			if ch.XP <= before {
				t.Fatalf("forest did not award XP: %d", ch.XP)
			}
			break
		}
		if ch.Level != 1 {
			t.Fatalf("forest auto-leveled mid-fight to %d", ch.Level)
		}
	}
	if !won {
		t.Fatalf("forest fight did not complete; last=\n%s", sink.plain())
	}
}

func trainFixture(t *testing.T, roll Roller) (*Hub, *entities.User, *trainSink, service.Facade) {
	t.Helper()
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "door-train.db"))
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
	hub.SetRoller(roll)
	hub.SetNow(func() time.Time { return time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC) })

	user, err := facade.UsersService().Create(&entities.User{
		RefID:    "local:train-" + strings.ReplaceAll(t.Name(), "/", "-"),
		Username: "u-" + strings.ReplaceAll(t.Name(), "/", "-"),
		Email:    strings.ReplaceAll(t.Name(), "/", "-") + "@example.com",
		Nickname: "train",
		Role:     entities.RolePlayer,
	})
	if err != nil {
		t.Fatal(err)
	}
	return hub, user, &trainSink{}, facade
}

func createDoorWarrior(t *testing.T, hub *Hub, user *entities.User, send func(any), name string) {
	t.Helper()
	hub.OnConnect(user, send)
	hub.OnInput(user, name, send)
	hub.OnInput(user, "f", send)
	hub.OnInput(user, "1", send)
	if user.LastCharacter == "" {
		t.Fatal("no warrior")
	}
}

func winMasterDuel(t *testing.T, hub *Hub, user *entities.User, sink *trainSink) {
	t.Helper()
	for i := 0; i < 60; i++ {
		hub.OnInput(user, "a", sink.send)
		sess := hub.session(user.ID)
		if sess != nil && sess.fight != nil && sess.fight.Done && sess.fight.Outcome == "win" {
			return
		}
		if sink.last().Screen == "result" {
			if sess != nil && sess.fight != nil && sess.fight.Outcome != "win" {
				t.Fatalf("ended with outcome %s\n%s", sess.fight.Outcome, sink.plain())
			}
			return
		}
	}
	t.Fatalf("master duel did not end in a win; last=\n%s", sink.plain())
}

func critWinSeq(n int) []int {
	seq := make([]int, 0, n*4)
	for i := 0; i < n; i++ {
		seq = append(seq, 19, 3) // player crit + dmg
		seq = append(seq, 0, 0)  // enemy miss path if still alive
	}
	return seq
}

func loseSeq(n int) []int {
	seq := make([]int, 0, n*4)
	for i := 0; i < n; i++ {
		seq = append(seq, 0)     // player fumble (roll 1)
		seq = append(seq, 19, 3) // enemy crit + dmg
	}
	return seq
}
