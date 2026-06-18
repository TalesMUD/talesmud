package commands_test

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func TestQuestOnlyNPCDialogCanAcceptQuestAndSendsEnrichedQuestLog(t *testing.T) {
	g, facade := newTradeTestGame(t)

	character, err := facade.CharactersService().Store(&characters.Character{
		Name:        "Questor",
		BelongsUser: *traits.BelongsToUser("user-quest"),
		CurrentRoom: traits.CurrentRoom{
			CurrentRoomID: "room-quest",
		},
	})
	if err != nil {
		t.Fatalf("store character: %v", err)
	}

	questGiver := &npc.NPC{
		Entity: &entities.Entity{ID: "npc-quest-giver"},
		Name:   "Archivist",
		CurrentRoom: traits.CurrentRoom{
			CurrentRoomID: "room-quest",
		},
		CurrentHitPoints: 10,
		MaxHitPoints:     10,
	}
	g.NPCManager.RegisterExistingNPC(questGiver, "room-quest")
	if _, err := facade.NPCsService().Import(questGiver); err != nil {
		t.Fatalf("store quest giver: %v", err)
	}
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity:      &entities.Entity{ID: "room-archive"},
		Name:        "Archive",
		Description: "Archive",
		Exits:       &rooms.Exits{},
	}); err != nil {
		t.Fatalf("store archive room: %v", err)
	}

	quest := &quests.Quest{
		Entity:      &entities.Entity{ID: "quest-missing-satchel"},
		Name:        "Missing Satchel",
		Description: "Find the archivist's missing satchel.",
		Category:    "side",
		Level:       2,
		Source: quests.QuestSource{
			Type:  "npc",
			NPCID: questGiver.ID,
		},
		Objectives: []quests.Objective{
			{
				ID:          "visit-archive",
				Type:        quests.ObjectiveVisit,
				Description: "Search the old archive.",
				TargetID:    "room-archive",
				Amount:      1,
			},
		},
		Rewards: quests.Reward{
			XP:   25,
			Gold: 5,
		},
		AcceptDialogText: "Please find my missing satchel.",
	}
	quest, err = facade.QuestsService().Store(quest)
	if err != nil {
		t.Fatalf("store quest: %v", err)
	}

	user := &entities.User{Entity: &entities.Entity{ID: "user-quest"}}
	talkMessage := &messages.Message{
		FromUser:  user,
		Character: character,
		Data:      "talk Archivist",
	}
	if !(&commands.TalkCommand{}).Execute(g, talkMessage) {
		t.Fatal("talk command did not handle quest-only NPC")
	}

	var dialog *messages.DialogMessage
	for _, out := range drainTradeMessages(g.SendMessage()) {
		if msg, ok := out.(*messages.DialogMessage); ok {
			dialog = msg
		}
	}
	if dialog == nil {
		t.Fatal("expected quest-only dialog message")
	}
	if dialog.ConversationID == "" {
		t.Fatal("expected quest-only dialog to include a conversation ID")
	}
	if len(dialog.Options) != 1 || !strings.Contains(dialog.Options[0].Text, "Missing Satchel") {
		t.Fatalf("expected quest option for Missing Satchel, got %#v", dialog.Options)
	}

	selectMessage := &messages.Message{
		FromUser:  user,
		Character: character,
		Data:      "1",
	}
	if !commands.DialogSelectCommand(nil, g, selectMessage) {
		t.Fatal("dialog select command did not handle quest-only selection")
	}

	progress, err := facade.QuestsService().GetProgress(character.ID, quest.ID)
	if err != nil {
		t.Fatalf("get quest progress: %v", err)
	}
	if progress == nil || progress.Status != quests.QuestStatusActive {
		t.Fatalf("expected quest to be active, got %#v", progress)
	}

	var questLog *messages.QuestLogMessage
	for _, out := range drainTradeMessages(g.SendMessage()) {
		if msg, ok := out.(messages.QuestLogMessage); ok {
			questLog = &msg
		}
	}
	if questLog == nil {
		t.Fatal("expected quest log update after accepting quest")
	}
	if len(questLog.Quests) != 1 {
		t.Fatalf("expected one quest log entry, got %d", len(questLog.Quests))
	}

	entry := questLog.Quests[0]
	if entry.QuestName != quest.Name {
		t.Fatalf("expected enriched quest name %q, got %q", quest.Name, entry.QuestName)
	}
	if entry.Description != quest.Description {
		t.Fatalf("expected enriched quest description %q, got %q", quest.Description, entry.Description)
	}
	if entry.Category != quest.Category {
		t.Fatalf("expected enriched category %q, got %q", quest.Category, entry.Category)
	}
	if entry.Rewards == nil || entry.Rewards.XP != 25 || entry.Rewards.Gold != 5 {
		t.Fatalf("expected enriched rewards, got %#v", entry.Rewards)
	}
	if len(entry.Objectives) != 1 || entry.Objectives[0].Description != "Search the old archive." {
		t.Fatalf("expected enriched objective description, got %#v", entry.Objectives)
	}
}
