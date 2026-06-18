package commands_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

func newTalkTestGame(t *testing.T) (*game.Game, service.Facade) {
	t.Helper()

	client, err := sqlite.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	return game.New(facade), facade
}

func drainTalkMessages(ch <-chan interface{}) []interface{} {
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

func storeTalkRoom(t *testing.T, facade service.Facade, id string) {
	t.Helper()
	if _, err := facade.RoomsService().Store(&rooms.Room{
		Entity:      &entities.Entity{ID: id},
		Name:        id,
		Description: id,
		Exits:       &rooms.Exits{},
	}); err != nil {
		t.Fatalf("store room: %v", err)
	}
}

func TestSpeakAliasStartsNPCDialog(t *testing.T) {
	g, facade := newTalkTestGame(t)
	storeTalkRoom(t, facade, "town")

	character := &characters.Character{
		Entity:      &entities.Entity{ID: "char-speak"},
		Name:        "Speaker",
		BelongsUser: *traits.BelongsToUser("user-1"),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "town"},
	}
	guide := &npc.NPC{
		Entity:      &entities.Entity{ID: "guide-1"},
		Name:        "Guide",
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "town"},
		State:       "idle",
	}
	g.NPCManager.RegisterExistingNPC(guide, "town")

	msg := &messages.Message{
		FromUser:  &entities.User{Entity: &entities.Entity{ID: "user-1"}},
		Character: character,
		Data:      "speak to guide",
	}

	if !g.CommandProcessor.Process(g, msg) {
		t.Fatal("expected speak alias to be handled by talk command")
	}
}

func TestQuestOnlyNPCSelectionAcceptsQuest(t *testing.T) {
	g, facade := newTalkTestGame(t)
	storeTalkRoom(t, facade, "quest-room")

	character, err := facade.CharactersService().Store(&characters.Character{
		Entity:      &entities.Entity{ID: "char-quest"},
		Name:        "Questor",
		BelongsUser: *traits.BelongsToUser("user-1"),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "quest-room"},
	})
	if err != nil {
		t.Fatalf("store character: %v", err)
	}

	questGiver := &npc.NPC{
		Entity:      &entities.Entity{ID: "quest-giver-1"},
		Name:        "Elder",
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "quest-room"},
		State:       "idle",
	}
	g.NPCManager.RegisterExistingNPC(questGiver, "quest-room")

	quest := &quests.Quest{
		Entity:           &entities.Entity{ID: "quest-1"},
		Name:             "First Errand",
		Description:      "Do a useful thing.",
		AcceptDialogText: "I knew you would help.",
		Source: quests.QuestSource{
			Type:  "npc",
			NPCID: questGiver.ID,
		},
		Objectives: []quests.Objective{
			{
				ID:          "talk-objective",
				Type:        quests.ObjectiveTalk,
				Description: "Talk to Elder.",
				TargetID:    questGiver.ID,
				Amount:      1,
			},
		},
	}
	if _, err := facade.QuestsService().Store(quest); err != nil {
		t.Fatalf("store quest: %v", err)
	}

	talk := &messages.Message{
		FromUser:  &entities.User{Entity: &entities.Entity{ID: "user-1"}},
		Character: character,
		Data:      "talk elder",
	}
	if !g.CommandProcessor.Process(g, talk) {
		t.Fatal("talk command did not handle quest NPC")
	}

	var sawQuestOption bool
	for _, out := range drainTalkMessages(g.SendMessage()) {
		if dialogMsg, ok := out.(*messages.DialogMessage); ok &&
			len(dialogMsg.Options) == 1 &&
			strings.Contains(dialogMsg.Options[0].Text, "First Errand") &&
			dialogMsg.ConversationID != "" {
			sawQuestOption = true
		}
	}
	if !sawQuestOption {
		t.Fatal("expected quest-only NPC to send selectable dialog with a conversation ID")
	}

	selectMsg := &messages.Message{
		FromUser:  &entities.User{Entity: &entities.Entity{ID: "user-1"}},
		Character: character,
		Data:      "1",
	}
	if !g.RoomProcessor.Process(g, selectMsg) {
		t.Fatal("expected numeric quest selection to be handled")
	}

	progress, err := facade.QuestsService().GetProgress(character.ID, quest.ID)
	if err != nil {
		t.Fatalf("get quest progress: %v", err)
	}
	if progress == nil || progress.Status != quests.QuestStatusActive {
		t.Fatalf("expected quest to be active after numeric selection, got %#v", progress)
	}
}
