package game

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/instances"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/resources"
	"github.com/talesmud/talesmud/pkg/ruleset"
	"github.com/talesmud/talesmud/pkg/service"
)

func TestReleaseToSafetyEndsCombatWithoutPenalty(t *testing.T) {
	ruleset.SetSafeRoom(ruleset.SafeBind)
	t.Cleanup(ruleset.Reset)

	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "wild", nil)
	storeTestRoom(t, facade, "haven", nil)
	char, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "char-safe"},
		Name:             "Hero",
		BelongsUser:      *traits.BelongsToUser("user-safe"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "wild"},
		BoundRoomID:      "haven",
		Gold:             40,
		MaxHitPoints:     20,
		CurrentHitPoints: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	g.NPCManager.RegisterExistingNPC(&npc.NPC{
		Entity:      &entities.Entity{ID: "npc-safe"},
		Name:        "Rat",
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "wild"},
		EnemyTrait:  &npc.EnemyTrait{AttackPower: 1},
	}, "wild")
	seedPacedCombat(t, g, char.ID, "npc-safe", true)

	g.ReleaseToSafety(char.ID)
	stored, err := facade.CharactersService().FindByID(char.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.InCombat || stored.AwaitingReset || stored.Gold != 40 || stored.CurrentHitPoints == 0 {
		t.Fatalf("after release combat=%v reset=%v gold=%d hp=%d", stored.InCombat, stored.AwaitingReset, stored.Gold, stored.CurrentHitPoints)
	}
	if stored.CurrentRoomID != "haven" {
		t.Fatalf("room=%s, want haven", stored.CurrentRoomID)
	}
}

func TestReleaseLeavesRealRoomWhenSafeRoomStays(t *testing.T) {
	ruleset.SetSafeRoom(ruleset.SafeStay)
	t.Cleanup(ruleset.Reset)
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "wild", nil)
	char, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "char-stay"},
		Name:             "Hero",
		BelongsUser:      *traits.BelongsToUser("user-stay"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "wild"},
		Gold:             9,
		MaxHitPoints:     20,
		CurrentHitPoints: 20,
		InCombat:         true,
	})
	if err != nil {
		t.Fatal(err)
	}
	g.NPCManager.RegisterExistingNPC(&npc.NPC{
		Entity:      &entities.Entity{ID: "npc-stay"},
		Name:        "Rat",
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "wild"},
		EnemyTrait:  &npc.EnemyTrait{},
	}, "wild")
	seedPacedCombat(t, g, char.ID, "npc-stay", true)
	g.ReleaseToSafety(char.ID)
	stored, err := facade.CharactersService().FindByID(char.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.CurrentRoomID != "wild" || stored.Gold != 9 || stored.InCombat || stored.AwaitingReset {
		t.Fatalf("stay release: room=%s gold=%d combat=%v reset=%v", stored.CurrentRoomID, stored.Gold, stored.InCombat, stored.AwaitingReset)
	}
}

func TestInstanceDisconnectReturnsWithoutDefeat(t *testing.T) {
	ruleset.SetSafeRoom(ruleset.SafeStay)
	t.Cleanup(ruleset.Reset)
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "R0108", nil)
	storeTestRoom(t, facade, "R0201", nil)
	char, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "char-wood"},
		Name:             "Hero",
		BelongsUser:      *traits.BelongsToUser("user-wood"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "R0108"},
		Gold:             15,
		MaxHitPoints:     25,
		CurrentHitPoints: 25,
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := g.RoomInstances.Generate(char.ID, 1, instances.ProcSpec{
		TemplateIDs:  []string{"R0201"},
		Count:        1,
		ReturnRoomID: "R0108",
		Seed:         1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := facade.CharactersService().Modify(char.ID, func(ch *characters.Character) error {
		ch.CurrentRoomID = res.EntryRoomID
		ch.InCombat = true
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	g.NPCManager.RegisterExistingNPC(&npc.NPC{
		Entity:      &entities.Entity{ID: "npc-wood"},
		Name:        "Wolf",
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: res.EntryRoomID},
		EnemyTrait:  &npc.EnemyTrait{},
	}, res.EntryRoomID)
	seedPacedCombat(t, g, char.ID, "npc-wood", true)

	g.ReleaseToSafety(char.ID)
	stored, err := facade.CharactersService().FindByID(char.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.CurrentRoomID != "R0108" || stored.Gold != 15 || stored.InCombat || stored.AwaitingReset || stored.CurrentHitPoints == 0 {
		t.Fatalf("instance release room=%s gold=%d hp=%d combat=%v reset=%v", stored.CurrentRoomID, stored.Gold, stored.CurrentHitPoints, stored.InCombat, stored.AwaitingReset)
	}
	if _, err := facade.RoomsService().FindByID(res.EntryRoomID); err == nil {
		t.Fatal("clone room still exists after leaving it")
	}
}

func TestSessionStartRefillsAndHeals(t *testing.T) {
	t.Cleanup(ruleset.Reset)
	if err := ruleset.LoadBytes([]byte("new_day:\n  full_heal: true\n  timezone: UTC\nresources:\n  walks:\n    allowance: 4\n    reset: calendar\n    timezone: UTC\n")); err != nil {
		t.Fatal(err)
	}
	client, err := sqlite.Open(filepath.Join(t.TempDir(), "day.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	g := New(facade)
	store, err := resources.New(client.DB())
	if err != nil {
		t.Fatal(err)
	}
	store.Configure(ruleset.ResourceAllowances())
	g.Resources = store
	storeTestRoom(t, facade, "square", nil)
	char, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "char-day"},
		Name:             "Hero",
		BelongsUser:      *traits.BelongsToUser("user-day"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "square"},
		MaxHitPoints:     30,
		CurrentHitPoints: 3,
		AwaitingReset:    true,
		LastResetDay:     time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02"),
	})
	if err != nil {
		t.Fatal(err)
	}
	yesterday := time.Now().UTC().AddDate(0, 0, -1)
	store.SetNow(func() time.Time { return yesterday })
	if _, err := store.Consume(char.ID, "walks", 4); err != nil {
		t.Fatal(err)
	}
	store.SetNow(nil)
	store.SetNow(func() time.Time { return time.Now().UTC() })
	fresh, err := facade.CharactersService().FindByID(char.ID)
	if err != nil {
		t.Fatal(err)
	}
	g.ApplySessionStart(fresh)
	stored, err := facade.CharactersService().FindByID(char.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.CurrentHitPoints != 30 || stored.AwaitingReset {
		t.Fatalf("heal hp=%d reset=%v", stored.CurrentHitPoints, stored.AwaitingReset)
	}
	res, ok, err := store.Get(char.ID, "walks")
	if err != nil || !ok || res.Remaining != 4 {
		t.Fatalf("walks %+v ok=%v err=%v", res, ok, err)
	}
}

func TestDisconnectContinueLeavesTheFight(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	if ruleset.Disconnect() != ruleset.DisconnectContinue {
		t.Fatalf("default disconnect = %s", ruleset.Disconnect())
	}
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "wild", nil)
	storeTestRoom(t, facade, "haven", nil)
	char, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "char-cont"},
		Name:             "Hero",
		BelongsUser:      *traits.BelongsToUser("user-cont"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "wild"},
		BoundRoomID:      "haven",
		Gold:             40,
		MaxHitPoints:     20,
		CurrentHitPoints: 20,
		InCombat:         true,
	})
	if err != nil {
		t.Fatal(err)
	}
	g.NPCManager.RegisterExistingNPC(&npc.NPC{
		Entity:      &entities.Entity{ID: "npc-cont"},
		Name:        "Rat",
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "wild"},
		EnemyTrait:  &npc.EnemyTrait{AttackPower: 1},
	}, "wild")
	seedPacedCombat(t, g, char.ID, "npc-cont", true)
	user := &entities.User{Entity: &entities.Entity{ID: "user-cont"}}
	g.SetUserSessionCharacter(user, char)
	g.DisconnectUserSession(user.ID)
	if g.CombatController.GetCombatInstance(char.ID) == nil {
		t.Fatal("continue ended the fight")
	}
	stored, err := facade.CharactersService().FindByID(char.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.CurrentRoomID != "wild" || stored.Gold != 40 || stored.AwaitingReset {
		t.Fatalf("room=%s gold=%d reset=%v", stored.CurrentRoomID, stored.Gold, stored.AwaitingReset)
	}
}

func TestBareAttackAskDoesNotStart(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	if ruleset.BareAttack() != ruleset.BareAttackAsk {
		t.Fatalf("default bare attack = %s", ruleset.BareAttack())
	}
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "R-ask", nil)
	if _, err := facade.NPCsService().Import(&npc.NPC{
		Entity:           &entities.Entity{ID: "ENM-ask"},
		Name:             "Bramble Wolf",
		IsTemplate:       true,
		MaxHitPoints:     12,
		CurrentHitPoints: 12,
		Level:            1,
		EnemyTrait:       &npc.EnemyTrait{AttackPower: 1},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := g.NPCManager.SpawnInstanceDirect("ENM-ask", "R-ask"); err != nil {
		t.Fatal(err)
	}
	user := &entities.User{Entity: &entities.Entity{ID: "user-ask"}}
	char, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "char-ask"},
		Name:             "Hero",
		BelongsUser:      *traits.BelongsToUser("user-ask"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "R-ask"},
		MaxHitPoints:     20,
		CurrentHitPoints: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !(&commands.AttackCommand{}).Execute(g, &messages.Message{FromUser: user, Character: char, Data: "attack"}) {
		t.Fatal("attack was not handled")
	}
	if g.CombatController.GetCombatInstance(char.ID) != nil {
		t.Fatal("ask started a fight")
	}
	saw := false
	for _, msg := range drainGameMessages(g.SendMessage()) {
		if text := messageText(msg); strings.Contains(text, "Attack whom?") {
			saw = true
		}
	}
	if !saw {
		t.Fatal("missing Attack whom?")
	}
}

func messageText(msg interface{}) string {
	switch m := msg.(type) {
	case messages.MessageResponse:
		return m.GetMessage()
	case *messages.MessageResponse:
		if m != nil {
			return m.GetMessage()
		}
	}
	return ""
}

func TestBareAttackStartsAndAdvances(t *testing.T) {
	ruleset.SetPacing(ruleset.PacingTurnBased)
	ruleset.SetBareAttack(ruleset.BareAttackFirst)
	t.Cleanup(ruleset.Reset)
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "R-bare", nil)
	if _, err := facade.NPCsService().Import(&npc.NPC{
		Entity:           &entities.Entity{ID: "ENM-bare"},
		Name:             "Bramble Wolf",
		IsTemplate:       true,
		MaxHitPoints:     12,
		CurrentHitPoints: 12,
		Level:            1,
		EnemyTrait:       &npc.EnemyTrait{AttackPower: 1, Defense: 0},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := g.NPCManager.SpawnInstanceDirect("ENM-bare", "R-bare"); err != nil {
		t.Fatal(err)
	}
	user := &entities.User{Entity: &entities.Entity{ID: "user-bare"}}
	char, err := facade.CharactersService().Store(&characters.Character{
		Entity:           &entities.Entity{ID: "char-bare"},
		Name:             "Hero",
		BelongsUser:      *traits.BelongsToUser("user-bare"),
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "R-bare"},
		Level:            1,
		MaxHitPoints:     40,
		CurrentHitPoints: 40,
		Attributes: []characters.Attribute{
			{Name: "Strength", Short: "STR", Value: 16},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	msg := &messages.Message{FromUser: user, Character: char, Data: "attack"}
	if !(&commands.AttackCommand{}).Execute(g, msg) {
		t.Fatal("bare attack failed")
	}
	inst := g.CombatController.GetCombatInstance(char.ID)
	if inst == nil {
		t.Fatal("bare attack did not start combat")
	}
	for i := 0; i < 8; i++ {
		current := inst.GetCurrentTurnCombatant()
		if inst.Phase == "waitingPlayer" && current != nil && current.ID == char.ID {
			break
		}
		inst.NextActionAt = time.Time{}
		g.CombatController.processAllTurns(inst)
	}
	current := inst.GetCurrentTurnCombatant()
	if inst.Phase != "waitingPlayer" || current == nil || current.ID != char.ID {
		t.Fatalf("player turn did not open: phase=%s", inst.Phase)
	}
	beforeHP := inst.Enemies[0].CurrentHP
	beforeTurn := inst.CurrentTurnIdx
	beforeRound := inst.Round
	msg.Data = "attack"
	if !(&commands.AttackCommand{}).Execute(g, msg) {
		t.Fatal("in-combat bare attack failed")
	}
	after := inst.GetEnemyByID(inst.Enemies[0].ID)
	if after == nil {
		t.Fatal("enemy missing")
	}
	if inst.Phase == "waitingPlayer" && inst.CurrentTurnIdx == beforeTurn && inst.Round == beforeRound && after.CurrentHP == beforeHP {
		t.Fatal("bare attack did not advance the turn")
	}
}

func TestEnsureLivingRoomUsesBind(t *testing.T) {
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "haven", nil)
	char, err := facade.CharactersService().Store(&characters.Character{
		Entity:      &entities.Entity{ID: "char-miss"},
		Name:        "Hero",
		BelongsUser: *traits.BelongsToUser("user-miss"),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "gone"},
		BoundRoomID: "haven",
	})
	if err != nil {
		t.Fatal(err)
	}
	fresh, err := facade.CharactersService().FindByID(char.ID)
	if err != nil {
		t.Fatal(err)
	}
	g.EnsureLivingRoom(fresh)
	stored, err := facade.CharactersService().FindByID(char.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.CurrentRoomID != "haven" {
		t.Fatalf("room=%s", stored.CurrentRoomID)
	}
}
