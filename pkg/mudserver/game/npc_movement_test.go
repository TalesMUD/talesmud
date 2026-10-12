package game

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func movementTestNPC(t *testing.T, loop string) (*Game, *npc.NPC) {
	t.Helper()
	g, f := newNPCTestGame(t)
	storeTestRoom(t, f, "a", rooms.Exits{{Name: "east", Target: "b"}, {Name: "west", Target: "c"}})
	storeTestRoom(t, f, "b", rooms.Exits{{Name: "west", Target: "a"}, {Name: "north", Target: "c"}})
	storeTestRoom(t, f, "c", rooms.Exits{{Name: "south", Target: "b"}, {Name: "east", Target: "a"}})
	n := &npc.NPC{Entity: &entities.Entity{ID: "walker"}, Name: "Pack Seller", MaxHitPoints: 100, CurrentHitPoints: 100, State: "idle", SpawnRoomID: "a", Movement: &npc.Movement{Mode: "route", Route: []string{"a", "b", "c"}, Loop: loop, IntervalSeconds: 90, JitterSeconds: 30, PauseOnEngageSeconds: 120}}
	g.NPCManager.RegisterExistingNPC(n, "a")
	return g, n
}

func TestMovementTimingAndIndependentJitter(t *testing.T) {
	g, n := movementTestNPC(t, "loop")
	now := time.Now()
	if n.NextMoveAt.Before(now.Add(89*time.Second)) || n.NextMoveAt.After(now.Add(121*time.Second)) {
		t.Fatal(n.NextMoveAt)
	}
	if _, moved := g.updateTimedNPC(n, n.NextMoveAt.Add(-time.Nanosecond)); moved {
		t.Fatal("moved before timer")
	}
	due := n.NextMoveAt
	if _, moved := g.updateTimedNPC(n, due); !moved || n.CurrentRoomID != "b" {
		t.Fatal("did not move when due")
	}
	if delta := n.NextMoveAt.Sub(due); delta < 90*time.Second || delta > 120*time.Second {
		t.Fatal(delta)
	}
	values := map[time.Duration]bool{}
	for i := 0; i < 100; i++ {
		values[nextMoveAt(now, n.Movement).Sub(now)] = true
	}
	if len(values) < 2 {
		t.Fatal("jitter is not sampled independently")
	}
	if _, moved := g.updateTimedNPC(n, due); moved {
		t.Fatal("two moves on one due tick")
	}
}

func TestMovementRequiresExplicitOptInAcrossManyTicks(t *testing.T) {
	for _, kind := range []string{"ordinary", "shop", "dialog", "patrol-state-only"} {
		t.Run(kind, func(t *testing.T) {
			g, n := movementTestNPC(t, "loop")
			n.Movement = nil
			n.WanderRadius = 0
			n.PatrolPath = nil
			n.NextMoveAt = time.Time{}
			if kind == "shop" {
				n.MerchantTrait = npc.NewMerchantTrait()
			}
			if kind == "dialog" {
				n.DialogID = "greeting"
			}
			if kind == "patrol-state-only" {
				n.State = "patrol"
			}
			for i := 0; i < 1000; i++ {
				g.handleNPCUpdates()
				if _, moved := g.updateTimedNPC(n, time.Now().Add(time.Duration(i+1)*time.Hour)); moved {
					t.Fatal("unconfigured NPC moved")
				}
				if n.CurrentRoomID != "a" || n.EffectiveMovement() != nil || !n.NextMoveAt.IsZero() {
					t.Fatal("movement was implicitly enabled")
				}
			}
		})
	}
}

func TestMovementRouteLoopAndPingpong(t *testing.T) {
	for _, tc := range []struct {
		loop string
		want []string
	}{{"loop", []string{"b", "c", "a", "b"}}, {"pingpong", []string{"b", "c", "b", "a", "b"}}} {
		t.Run(tc.loop, func(t *testing.T) {
			g, n := movementTestNPC(t, tc.loop)
			for _, want := range tc.want {
				if _, moved := g.updateTimedNPC(n, n.NextMoveAt); !moved || n.CurrentRoomID != want {
					t.Fatalf("room %s want %s", n.CurrentRoomID, want)
				}
			}
		})
	}
}

func TestMovementHoldsEveryOpenSessionAndRefreshesPause(t *testing.T) {
	for _, kind := range []string{"dialog", "shop"} {
		t.Run(kind, func(t *testing.T) {
			g, n := movementTestNPC(t, "loop")
			g.NPCManager.EngageNPC("p1", n.ID, kind)
			g.NPCManager.EngageNPC("p2", n.ID, kind)
			future := time.Now().Add(time.Hour)
			if _, moved := g.updateTimedNPC(n, future); moved {
				t.Fatal("moved mid-session")
			}
			if n.PausedUntil != future.Add(120*time.Second) {
				t.Fatal("did not refresh open hold")
			}
			g.NPCManager.CloseNPCSession("p1", kind)
			if _, moved := g.updateTimedNPC(n, future.Add(time.Minute)); moved {
				t.Fatal("ignored second player's hold")
			}
			g.NPCManager.CloseNPCSession("p2", kind)
			if _, moved := g.updateTimedNPC(n, n.PausedUntil.Add(-time.Nanosecond)); moved {
				t.Fatal("moved during close pause")
			}
			if _, moved := g.updateTimedNPC(n, future.Add(5*time.Minute)); !moved {
				t.Fatal("did not resume after close")
			}
		})
	}
}

func TestMovementEngagementCommands(t *testing.T) {
	g, n := movementTestNPC(t, "loop")
	char := &characters.Character{Entity: &entities.Entity{ID: "player"}, Name: "Traveler"}
	char.CurrentRoomID = "a"
	user := &entities.User{Entity: &entities.Entity{ID: "user"}}
	n.DialogID = "greeting"
	_, err := g.Facade.DialogsService().Import(&dialogs.Dialog{Entity: &entities.Entity{ID: n.DialogID}, NodeID: "main", Text: "Hello", Options: dialogs.Options(dialogs.NewDialog("bye", "Goodbye", nil, nil))})
	if err != nil {
		t.Fatal(err)
	}
	n.MerchantTrait = npc.NewMerchantTrait()
	// A real template keeps SHOP populated.
	n.MerchantTrait.Inventory = []npc.MerchantItem{{ItemTemplateID: "supplies", Quantity: -1, MaxQuantity: -1}}
	for _, cmd := range []string{"examine Pack Seller", "look Pack Seller", "target Pack Seller", "talk Pack Seller", "shop", "buy absent", "sell absent", "trade"} {
		t.Run(cmd, func(t *testing.T) {
			n.PausedUntil = time.Now().Add(-time.Hour)
			m := &messages.Message{Data: cmd, Character: char, FromUser: user}
			if !g.CommandProcessor.Process(g, m) {
				g.RoomProcessor.Process(g, m)
			}
			if n.PausedUntil.Before(time.Now().Add(119 * time.Second)) {
				t.Fatal("pause did not refresh")
			}
			drainNPCMessages(g.SendMessage())
		})
	}
	if !g.NPCManager.NPCSessionOpen(char.ID, n.ID, "shop") {
		t.Fatal("shop hold missing")
	}
	g.CommandProcessor.Process(g, &messages.Message{Data: "shopclose", Character: char, FromUser: user})
	if g.NPCManager.movementHeld(n.ID) {
		t.Fatal("shop dismissal did not release")
	}
}

func TestMovementReleaseOnLeaveAndCharacterDisconnect(t *testing.T) {
	g, n := movementTestNPC(t, "loop")
	g.NPCManager.EngageNPC("p", n.ID, "dialog")
	g.NotePlayerEntered("p", "b")
	if g.NPCManager.movementHeld(n.ID) {
		t.Fatal("leaving did not release")
	}
	user := &entities.User{Entity: &entities.Entity{ID: "user"}}
	char := &characters.Character{Entity: &entities.Entity{ID: "p"}, Name: "Traveler"}
	char.CurrentRoomID = "a"
	g.ConnectUserSession(user)
	g.SetUserSessionCharacter(user, char)
	g.NPCManager.EngageNPC("p", n.ID, "shop")
	g.DisconnectUserSession(user.ID)
	if g.NPCManager.movementHeld(n.ID) {
		t.Fatal("disconnect did not release")
	}
}

func TestMovementReleaseOnCharacterSwitch(t *testing.T) {
	g, n := movementTestNPC(t, "loop")
	user := &entities.User{Entity: &entities.Entity{ID: "user"}}
	first := &characters.Character{Entity: &entities.Entity{ID: "p"}, Name: "Traveler"}
	first.CurrentRoomID = "a"
	second := &characters.Character{Entity: &entities.Entity{ID: "q"}, Name: "Other"}
	second.CurrentRoomID = "a"
	g.ConnectUserSession(user)
	g.SetUserSessionCharacter(user, first)
	g.NPCManager.EngageNPC(first.ID, n.ID, "dialog")
	if !g.NPCManager.movementHeld(n.ID) {
		t.Fatal("expected an open dialog hold")
	}
	g.SetUserSessionCharacter(user, second)
	if g.NPCManager.movementHeld(n.ID) {
		t.Fatal("character switch did not release")
	}
}

func TestDialogGoodbyeReleasesHold(t *testing.T) {
	g, n := movementTestNPC(t, "loop")
	char := &characters.Character{Entity: &entities.Entity{ID: "player"}, Name: "Traveler"}
	char.CurrentRoomID = "a"
	user := &entities.User{Entity: &entities.Entity{ID: "user"}}
	n.DialogID = "greeting"
	exit := true
	later := dialogs.NewDialog("later", "Later.", nil, nil)
	bye := dialogs.NewDialog("end", "Just browsing. Goodbye.", nil, nil)
	bye.Answer = &dialogs.Dialog{NodeID: "end", Text: "Come back anytime.", IsDialogExit: &exit}
	ask := dialogs.NewDialog("who", "Who are you?", nil, nil)
	travels := dialogs.NewDialog("end", "Safe travels.", nil, nil)
	travels.Answer = &dialogs.Dialog{NodeID: "end", Text: "And you.", IsDialogExit: &exit}
	ask.Answer = &dialogs.Dialog{NodeID: "who", Text: "Pack Seller.", Options: []*dialogs.Dialog{travels}}
	root := &dialogs.Dialog{
		Entity:  &entities.Entity{ID: n.DialogID},
		NodeID:  "main",
		Text:    "What do you need?",
		Options: []*dialogs.Dialog{ask, bye, later},
	}
	if _, err := g.Facade.DialogsService().Import(root); err != nil {
		t.Fatal(err)
	}
	play := func(data string) []interface{} {
		m := &messages.Message{Data: data, Character: char, FromUser: user}
		if !g.CommandProcessor.Process(g, m) {
			g.RoomProcessor.Process(g, m)
		}
		return drainNPCMessages(g.SendMessage())
	}
	sawEnd := func(outs []interface{}, text string) bool {
		for _, out := range outs {
			msg, ok := out.(messages.MessageResponse)
			if ok && msg.Type == messages.MessageTypeDialogEnd && strings.Contains(msg.Message, text) {
				return true
			}
		}
		return false
	}

	play("talk Pack Seller")
	if !g.NPCManager.movementHeld(n.ID) {
		t.Fatal("talk did not hold")
	}
	if play("1") == nil || !g.NPCManager.movementHeld(n.ID) {
		t.Fatal("continuing the dialog released the hold")
	}
	if !sawEnd(play("1"), "And you.") || g.NPCManager.movementHeld(n.ID) {
		t.Fatal("leaf goodbye did not end the dialog")
	}

	play("talk Pack Seller")
	if !g.NPCManager.movementHeld(n.ID) {
		t.Fatal("second talk did not hold")
	}
	if !sawEnd(play("2"), "Come back anytime.") || g.NPCManager.movementHeld(n.ID) {
		t.Fatal("root goodbye did not end the dialog")
	}

	play("talk Pack Seller")
	if !sawEnd(play("3"), "Later.") || g.NPCManager.movementHeld(n.ID) {
		t.Fatal("option with no answer did not end the dialog")
	}

	play("talk Pack Seller")
	if !g.NPCManager.NPCSessionOpen(char.ID, n.ID, "dialog") {
		t.Fatal("dialog session missing")
	}
	outs := play("dialogclose")
	if g.NPCManager.movementHeld(n.ID) || !sawEnd(outs, "") {
		t.Fatal("dialogclose did not release")
	}
}

func TestMovementCombatDeadAndUnsafeAggression(t *testing.T) {
	for _, state := range []string{"combat", "dead", "zero-hp", "boss", "low-level"} {
		t.Run(state, func(t *testing.T) {
			g, n := movementTestNPC(t, "loop")
			switch state {
			case "combat":
				n.InCombat = true
			case "dead":
				n.IsDead = true
			case "zero-hp":
				n.CurrentHitPoints = 0
			case "boss":
				n.Level = 12
				n.EnemyTrait = &npc.EnemyTrait{AggroOnSight: true, Difficulty: "boss"}
			case "low-level":
				n.Level = 3
				n.EnemyTrait = &npc.EnemyTrait{AggroOnSight: true}
			}
			if _, moved := g.updateTimedNPC(n, n.NextMoveAt.Add(time.Hour)); moved {
				t.Fatal("unsafe NPC moved")
			}
		})
	}
}

func TestMovementRoamConstraintsAndExitRestrictions(t *testing.T) {
	g, f := newNPCTestGame(t)
	exits := rooms.Exits{{Name: "east", Target: "ok"}, {Name: "hidden", Target: "hidden", Hidden: true}, {Name: "locked", Target: "locked", Locked: true}, {Name: "params", Target: "params", Params: map[string]interface{}{"locked": true}}, {Name: "instance", Target: "instance", Instance: true}, {Name: "typed-instance", Target: "typed", Type: "instance"}, {Name: "teleport", Target: "teleport", Type: rooms.RoomExitTypeTeleport}, {Name: "clone", Target: "clone~x"}, {Name: "private", Target: "private"}, {Name: "other-zone", Target: "other-zone"}, {Name: "off-whitelist", Target: "off-whitelist"}, {Name: "no-wander", Target: "no-wander"}, {Name: "no-npc", Target: "no-npc"}}
	for _, e := range append(exits, rooms.Exit{Target: "spawn"}) {
		r := testRoomWithExits(e.Target, nil)
		r.Area = "zone"
		if e.Target == "spawn" {
			r.Exits = &exits
		}
		if e.Target == "private" {
			r.Tags = []string{"instance"}
		}
		if e.Target == "other-zone" {
			r.Area = "other"
		}
		if e.Target == "no-wander" {
			r.Tags = []string{"noWander"}
		}
		if e.Target == "no-npc" {
			r.Tags = []string{"noNPC"}
		}
		if _, err := f.RoomsService().Store(r); err != nil {
			t.Fatal(err)
		}
	}
	n := &npc.NPC{Entity: &entities.Entity{ID: "roamer"}, Name: "Traveler", State: "idle", SpawnRoomID: "spawn", Movement: &npc.Movement{Mode: "roam", Zones: []string{"zone"}}}
	g.NPCManager.RegisterExistingNPC(n, "spawn")
	if _, moved := g.updateTimedNPC(n, n.NextMoveAt); !moved || (n.CurrentRoomID != "ok" && n.CurrentRoomID != "off-whitelist") {
		t.Fatalf("exit constraints ignored: %s", n.CurrentRoomID)
	}
	// Optional whitelist further narrows the same zone.
	g.NPCManager.MoveInstance(n.ID, "spawn")
	n.Movement.Rooms = []string{"spawn", "ok"}
	for i := 0; i < 20; i++ {
		if _, moved := g.updateTimedNPC(n, n.NextMoveAt); !moved || n.CurrentRoomID != "ok" {
			t.Fatalf("whitelist ignored: %s", n.CurrentRoomID)
		}
		g.NPCManager.MoveInstance(n.ID, "spawn")
	}
}

func TestMovementInvalidStoredRoomResetsAndInvalidRouteSkips(t *testing.T) {
	g, n := movementTestNPC(t, "loop")
	n.CurrentRoomID = "missing"
	g.NPCManager.initializeMovement(n)
	if n.CurrentRoomID != "a" {
		t.Fatal("missing current room did not reset")
	}
	n.SpawnRoomID = "missing"
	n.CurrentRoomID = "missing"
	g.NPCManager.initializeMovement(n)
	if n.CurrentRoomID != "a" {
		t.Fatal("missing spawn did not use first route room")
	}
	n.Movement.Route = []string{"a", "missing"}
	if _, moved := g.updateTimedNPC(n, n.NextMoveAt.Add(time.Hour)); moved {
		t.Fatal("invalid route moved")
	}
}

func TestMovementBroadcastProseAndRoomNPCPayload(t *testing.T) {
	g, n := movementTestNPC(t, "loop")
	// Existing test fixtures also prove the manager positions used by roomUpdate.
	event, moved := g.updateTimedNPC(n, n.NextMoveAt)
	if !moved {
		t.Fatal("no movement")
	}
	g.broadcastNPCMovements([]npcMovementEvent{event})
	outputs := drainNPCMessages(g.SendMessage())
	found := map[string]string{}
	for _, out := range outputs {
		if msg, ok := out.(messages.MessageResponse); ok {
			found[msg.AudienceID] = msg.Message
		}
	}
	if !strings.Contains(found["a"], "Pack Seller heads east.") || !strings.Contains(found["b"], "from the west") {
		t.Fatal(found)
	}
	if len(g.NPCManager.GetInstancesInRoom("a")) != 0 || len(g.NPCManager.GetInstancesInRoom("b")) != 1 {
		t.Fatal("NPC list not moved")
	}
}

func TestMovementSpawnerInheritsAndOverrideDoesNotAlias(t *testing.T) {
	g, n := movementTestNPC(t, "pingpong")
	n.IsTemplate = true
	n.SpawnRoomID = ""
	if _, err := g.Facade.NPCsService().Import(n); err != nil {
		t.Fatal(err)
	}
	inherited, err := g.NPCManager.SpawnInstance(&npc.NPCSpawner{Entity: &entities.Entity{ID: "s1"}, TemplateID: n.ID, RoomID: "a"})
	if err != nil {
		t.Fatal(err)
	}
	if inherited.Movement.Mode != "route" || inherited.NextMoveAt.IsZero() {
		t.Fatal("movement not inherited/scheduled")
	}
	override := &npc.Movement{Mode: "roam", Rooms: []string{"a", "b"}, IntervalSeconds: 150}
	overridden, err := g.NPCManager.SpawnInstance(&npc.NPCSpawner{Entity: &entities.Entity{ID: "s2"}, TemplateID: n.ID, RoomID: "b", Movement: override})
	if err != nil {
		t.Fatal(err)
	}
	if overridden.Movement.Mode != "roam" || overridden.Movement.IntervalSeconds != 150 {
		t.Fatal("override lost")
	}
	overridden.Movement.Rooms[0] = "c"
	if override.Rooms[0] != "a" || n.Movement.Mode != "route" {
		t.Fatal("override aliases content")
	}
}

func TestStationaryNPCOverlayReleasesPreviousWanderer(t *testing.T) {
	g, n := movementTestNPC(t, "loop")
	stationary := &npc.NPC{Entity: &entities.Entity{ID: "stationary"}, Name: "Innkeeper", State: "idle"}
	g.NPCManager.RegisterExistingNPC(stationary, "a")
	g.NPCManager.EngageNPC("p", n.ID, "dialog")
	g.NPCManager.EngageNPC("p", stationary.ID, "shop")
	if g.NPCManager.movementHeld(n.ID) {
		t.Fatal("replaced overlay holds old wanderer")
	}
}

func TestMovementRoomUpdateRemovesAndAddsNPCForOnlineOccupants(t *testing.T) {
	g, n := movementTestNPC(t, "loop")
	for _, roomID := range []string{"a", "b"} {
		char := &characters.Character{Entity: &entities.Entity{ID: "char-" + roomID}, Name: "Watcher-" + roomID}
		char.BelongsUserID = "user-" + roomID
		char.CurrentRoomID = roomID
		if _, err := g.Facade.CharactersService().Import(char); err != nil {
			t.Fatal(err)
		}
		user := &entities.User{Entity: &entities.Entity{ID: char.BelongsUserID}, RefID: char.BelongsUserID, IsOnline: true, LastCharacter: char.ID}
		if _, err := g.Facade.UsersService().Import(user); err != nil {
			t.Fatal(err)
		}
		room, _ := g.Facade.RoomsService().FindByID(roomID)
		room.Characters = &rooms.Characters{char.ID}
		if err := g.Facade.RoomsService().Update(room.ID, room); err != nil {
			t.Fatal(err)
		}
	}
	event, moved := g.updateTimedNPC(n, n.NextMoveAt)
	if !moved {
		t.Fatal("no move")
	}
	g.broadcastNPCMovements([]npcMovementEvent{event})
	payloads := map[string]int{}
	for _, out := range drainNPCMessages(g.SendMessage()) {
		body, err := json.Marshal(out)
		if err != nil {
			t.Fatal(err)
		}
		var msg struct {
			Type       string `json:"type"`
			AudienceID string `json:"audienceId"`
			NPCs       []struct {
				ID string `json:"id"`
			} `json:"npcs"`
		}
		if err := json.Unmarshal(body, &msg); err != nil {
			t.Fatal(err)
		}
		if msg.Type == "roomUpdate" {
			update, ok := out.(*messages.EnterRoomMessage)
			if !ok {
				t.Fatalf("roomUpdate has unexpected type %T", out)
			}
			payloads[update.AudienceID] = len(msg.NPCs)
		}
	}
	if payloads["user-a"] != 0 || payloads["user-b"] != 1 || len(payloads) != 2 {
		t.Fatalf("room update payloads: %#v", payloads)
	}
}
