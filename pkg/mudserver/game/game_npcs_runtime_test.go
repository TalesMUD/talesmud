package game

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/mudserver/game/util"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

func newNPCTestGame(t *testing.T) (*Game, service.Facade) {
	t.Helper()

	client, err := sqlite.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	return New(facade), facade
}

func storeTestRoom(t *testing.T, facade service.Facade, id string, exits rooms.Exits) {
	t.Helper()
	if _, err := facade.RoomsService().Store(&rooms.Room{
		Entity:      &entities.Entity{ID: id},
		Name:        id,
		Description: id,
		Exits:       &exits,
	}); err != nil {
		t.Fatalf("store room %s: %v", id, err)
	}
}

func drainNPCMessages(ch <-chan interface{}) []interface{} {
	var result []interface{}
	for {
		select {
		case msg := <-ch:
			result = append(result, msg)
		default:
			return result
		}
	}
}

func TestHandleNPCUpdatesMovesPatrolNPCToNextPatrolRoom(t *testing.T) {
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "room-a", rooms.Exits{{Name: "east", Target: "room-b"}})
	storeTestRoom(t, facade, "room-b", rooms.Exits{{Name: "east", Target: "room-c"}})
	storeTestRoom(t, facade, "room-c", rooms.Exits{{Name: "west", Target: "room-a"}})

	guard := &npc.NPC{
		Entity:      &entities.Entity{ID: "guard-1"},
		Name:        "Patrol Guard",
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "room-a"},
		SpawnRoomID: "room-a",
		State:       "patrol",
		PatrolPath:  []string{"room-a", "room-b", "room-c"},
	}
	g.NPCManager.RegisterExistingNPC(guard, "room-a")

	g.handleNPCUpdates()

	if got := g.NPCManager.GetInstance("guard-1").CurrentRoomID; got != "room-b" {
		t.Fatalf("expected patrol guard to move to room-b, got %s", got)
	}
}

func TestHandleNPCUpdatesWandersOnlyWithinConfiguredRadius(t *testing.T) {
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "spawn", rooms.Exits{{Name: "north", Target: "near"}})
	storeTestRoom(t, facade, "near", rooms.Exits{{Name: "south", Target: "spawn"}, {Name: "north", Target: "far"}})
	storeTestRoom(t, facade, "far", rooms.Exits{{Name: "south", Target: "near"}})

	wanderer := &npc.NPC{
		Entity:       &entities.Entity{ID: "wanderer-1"},
		Name:         "Wanderer",
		CurrentRoom:  traits.CurrentRoom{CurrentRoomID: "spawn"},
		SpawnRoomID:  "spawn",
		State:        "idle",
		WanderRadius: 1,
	}
	g.NPCManager.RegisterExistingNPC(wanderer, "spawn")

	g.handleNPCUpdates()
	if got := g.NPCManager.GetInstance("wanderer-1").CurrentRoomID; got != "near" {
		t.Fatalf("expected wanderer to move to adjacent room within radius, got %s", got)
	}

	g.handleNPCUpdates()
	if got := g.NPCManager.GetInstance("wanderer-1").CurrentRoomID; got != "spawn" {
		t.Fatalf("expected wanderer to avoid room outside radius and return toward spawn, got %s", got)
	}
}

func TestHandleNPCUpdatesSendsIdleDialogOnCooldown(t *testing.T) {
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "tavern", rooms.Exits{})

	if _, err := facade.DialogsService().Import(&dialogs.Dialog{
		Entity: &entities.Entity{ID: "idle-dialog"},
		NodeID: "main",
		Text:   "Lovely weather for standing around.",
	}); err != nil {
		t.Fatalf("store idle dialog: %v", err)
	}

	barkeep := &npc.NPC{
		Entity:            &entities.Entity{ID: "barkeep-1"},
		Name:              "Barkeep",
		CurrentRoom:       traits.CurrentRoom{CurrentRoomID: "tavern"},
		State:             "idle",
		IdleDialogID:      "idle-dialog",
		IdleDialogTimeout: time.Minute,
		LastIdleDialog:    time.Now().Add(-2 * time.Minute),
	}
	g.NPCManager.RegisterExistingNPC(barkeep, "tavern")

	g.handleNPCUpdates()

	var sawChatter bool
	for _, out := range drainNPCMessages(g.SendMessage()) {
		if rsp, ok := out.(messages.MessageResponse); ok &&
			rsp.Audience == messages.MessageAudienceRoom &&
			rsp.AudienceID == "tavern" &&
			rsp.Username == "Barkeep" &&
			strings.Contains(rsp.Message, "Lovely weather") {
			sawChatter = true
		}
	}
	if !sawChatter {
		t.Fatal("expected idle dialog to be sent to the NPC room")
	}

	g.handleNPCUpdates()
	if got := len(drainNPCMessages(g.SendMessage())); got != 0 {
		t.Fatalf("expected idle dialog cooldown to suppress immediate repeat, got %d messages", got)
	}
}

func TestGetRoomNPCsIncludesInteractionState(t *testing.T) {
	g, facade := newNPCTestGame(t)
	storeTestRoom(t, facade, "square", rooms.Exits{})

	elder := &npc.NPC{
		Entity:       &entities.Entity{ID: "elder-1"},
		Name:         "Elder",
		CurrentRoom:  traits.CurrentRoom{CurrentRoomID: "square"},
		State:        "idle",
		DialogID:     "elder-dialog",
		IdleDialogID: "elder-bark",
	}
	g.NPCManager.RegisterExistingNPC(elder, "square")

	if _, err := facade.QuestsService().Store(&quests.Quest{
		Entity:      &entities.Entity{ID: "elder-quest"},
		Name:        "A Small Favor",
		Description: "Help the elder.",
		Source: quests.QuestSource{
			Type:  "npc",
			NPCID: elder.ID,
		},
	}); err != nil {
		t.Fatalf("store quest: %v", err)
	}

	room, err := facade.RoomsService().FindByID("square")
	if err != nil {
		t.Fatalf("load room: %v", err)
	}
	npcs := util.GetRoomNPCs(room, g)
	if len(npcs) != 1 {
		t.Fatalf("expected one room NPC, got %d", len(npcs))
	}
	if !npcs[0].HasDialog {
		t.Fatal("expected NPC payload to mark interactive dialog")
	}
	if !npcs[0].IsQuestGiver {
		t.Fatal("expected NPC payload to mark quest giver")
	}
	if !npcs[0].HasIdleDialog {
		t.Fatal("expected NPC payload to mark idle chatter")
	}
}
