package commands_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/service"
)

// Dialog shaped like the content importer output: the start node is "root",
// topics are choice+answer pairs sharing a node id, and "Back" points at "root".
func rootDialog(id string, topicOptions ...*dialogs.Dialog) *dialogs.Dialog {
	back := &dialogs.Dialog{NodeID: "root", Text: "Back.", Answer: &dialogs.Dialog{NodeID: "root", Text: "Root text."}}
	opts := append([]*dialogs.Dialog{back}, topicOptions...)
	return &dialogs.Dialog{
		Entity: &entities.Entity{ID: id},
		NodeID: "root",
		Text:   "Root text.",
		Options: []*dialogs.Dialog{
			{NodeID: "topic", Text: "Tell me more.", Answer: &dialogs.Dialog{NodeID: "topic", Text: "Topic text.", Options: opts}},
			{NodeID: "other", Text: "Something else.", Answer: &dialogs.Dialog{NodeID: "other", Text: "Other text.", Options: []*dialogs.Dialog{back}}},
		},
	}
}

type dialogFixture struct {
	g         *game.Game
	facade    service.Facade
	character *characters.Character
	user      *entities.User
	quest     *quests.Quest
}

func newDialogFixture(t *testing.T, dialogID string, build func(questID string) *dialogs.Dialog) *dialogFixture {
	t.Helper()
	g, facade := newTradeTestGame(t)
	character, err := facade.CharactersService().Store(&characters.Character{
		Name:        "Talker",
		BelongsUser: *traits.BelongsToUser("user-dlg"),
		CurrentRoom: traits.CurrentRoom{CurrentRoomID: "room-dlg"},
	})
	if err != nil {
		t.Fatalf("store character: %v", err)
	}
	giver := &npc.NPC{
		Entity:           &entities.Entity{ID: "npc-giver"},
		Name:             "Giver",
		CurrentRoom:      traits.CurrentRoom{CurrentRoomID: "room-dlg"},
		DialogID:         dialogID,
		CurrentHitPoints: 10,
		MaxHitPoints:     10,
	}
	g.NPCManager.RegisterExistingNPC(giver, "room-dlg")
	if _, err := facade.NPCsService().Import(giver); err != nil {
		t.Fatalf("store npc: %v", err)
	}
	if _, err := facade.RoomsService().Import(&rooms.Room{
		Entity: &entities.Entity{ID: "room-far"}, Name: "Far", Description: "Far", Exits: &rooms.Exits{},
	}); err != nil {
		t.Fatalf("store room: %v", err)
	}
	quest, err := facade.QuestsService().Store(&quests.Quest{
		Entity:      &entities.Entity{ID: "quest-errand"},
		Name:        "Errand",
		Description: "Run an errand.",
		Source:      quests.QuestSource{Type: "npc", NPCID: giver.ID},
		Objectives:  []quests.Objective{{ID: "obj", Description: "Go far.", Type: quests.ObjectiveVisit, TargetID: "room-far", Amount: 1}},
		Rewards:     quests.Reward{XP: 10},
	})
	if err != nil {
		t.Fatalf("store quest: %v", err)
	}
	dlg := build(quest.ID)
	dlg.ID = dialogID
	if _, err := facade.DialogsService().Import(dlg); err != nil {
		t.Fatalf("store dialog: %v", err)
	}
	return &dialogFixture{g: g, facade: facade, character: character, quest: quest,
		user: &entities.User{Entity: &entities.Entity{ID: "user-dlg"}}}
}

func (f *dialogFixture) lastDialog(t *testing.T) *messages.DialogMessage {
	t.Helper()
	var dialog *messages.DialogMessage
	for _, out := range drainTradeMessages(f.g.SendMessage()) {
		if msg, ok := out.(*messages.DialogMessage); ok {
			dialog = msg
		}
	}
	return dialog
}

func (f *dialogFixture) talk(t *testing.T) *messages.DialogMessage {
	t.Helper()
	(&commands.TalkCommand{}).Execute(f.g, &messages.Message{FromUser: f.user, Character: f.character, Data: "talk Giver"})
	d := f.lastDialog(t)
	if d == nil {
		t.Fatal("expected a dialog after talk")
	}
	return d
}

func (f *dialogFixture) choose(t *testing.T, label string, d *messages.DialogMessage) *messages.DialogMessage {
	t.Helper()
	for _, o := range d.Options {
		if strings.Contains(o.Text, label) {
			commands.DialogSelectCommand(nil, f.g, &messages.Message{FromUser: f.user, Character: f.character, Data: itoa(o.Index)})
			return f.lastDialog(t)
		}
	}
	t.Fatalf("option %q not offered in %#v", label, d.Options)
	return nil
}

func itoa(i int) string { return strconv.Itoa(i) }

func (f *dialogFixture) status(t *testing.T) quests.QuestStatus {
	t.Helper()
	p, _ := f.facade.QuestsService().GetProgress(f.character.ID, f.quest.ID)
	if p == nil {
		return ""
	}
	return p.Status
}

func hasOption(d *messages.DialogMessage, label string) bool {
	for _, o := range d.Options {
		if strings.Contains(o.Text, label) {
			return true
		}
	}
	return false
}

func TestGeneratedQuestOptionSelectableAtNestedNode(t *testing.T) {
	f := newDialogFixture(t, "dlg-nested", func(string) *dialogs.Dialog { return rootDialog("dlg-nested") })
	topic := f.choose(t, "Tell me more.", f.talk(t))
	if topic == nil || topic.Message == "" {
		t.Fatal("expected topic node")
	}
	f.choose(t, "[Quest] Errand", topic)
	if f.status(t) != quests.QuestStatusActive {
		t.Fatalf("quest not accepted from nested node, status %q", f.status(t))
	}
}

func TestGeneratedQuestOptionSelectableAfterReturningToRootNode(t *testing.T) {
	f := newDialogFixture(t, "dlg-root", func(string) *dialogs.Dialog { return rootDialog("dlg-root") })
	topic := f.choose(t, "Tell me more.", f.talk(t))
	root := f.choose(t, "Back.", topic)
	if root == nil || !hasOption(root, "[Quest] Errand") {
		t.Fatalf("expected [Quest] at the root node, got %#v", root)
	}
	f.choose(t, "[Quest] Errand", root)
	if f.status(t) != quests.QuestStatusActive {
		t.Fatalf("quest not accepted at node \"root\", status %q", f.status(t))
	}
}

func TestTurnInFromNestedNode(t *testing.T) {
	f := newDialogFixture(t, "dlg-turnin", func(string) *dialogs.Dialog { return rootDialog("dlg-turnin") })
	f.choose(t, "[Quest] Errand", f.talk(t))
	if _, err := f.facade.QuestsService().IncrementObjective(f.character.ID, f.quest.ID, "obj", 1); err != nil {
		t.Fatalf("progress: %v", err)
	}
	topic := f.choose(t, "Tell me more.", f.talk(t))
	f.choose(t, "[Turn In] Errand", topic)
	if f.status(t) != quests.QuestStatusCompleted {
		t.Fatalf("turn-in from nested node failed, status %q", f.status(t))
	}
}

func TestTalkAgainRestartsAtStartNode(t *testing.T) {
	f := newDialogFixture(t, "dlg-restart", func(string) *dialogs.Dialog { return rootDialog("dlg-restart") })
	topic := f.choose(t, "Tell me more.", f.talk(t))
	if !strings.Contains(topic.Message, "Topic text.") {
		t.Fatalf("expected topic text, got %q", topic.Message)
	}
	again := f.talk(t)
	if !strings.Contains(again.Message, "Root text.") || !hasOption(again, "Something else.") {
		t.Fatalf("talking again should restart at the start node, got %q %#v", again.Message, again.Options)
	}
	// Numbers now map to the start node's options.
	other := f.choose(t, "Something else.", again)
	if !strings.Contains(other.Message, "Other text.") {
		t.Fatalf("selection after restart went to the wrong node: %q", other.Message)
	}
}

func TestAuthoredQuestOptionSuppressesDuplicateGeneratedOne(t *testing.T) {
	f := newDialogFixture(t, "dlg-authored", func(questID string) *dialogs.Dialog {
		accept := &dialogs.Dialog{NodeID: "take", Text: "I will do it.", QuestID: questID, Action: "accept",
			Answer: &dialogs.Dialog{NodeID: "take", Text: "Good."}}
		return rootDialog("dlg-authored", accept)
	})
	root := f.talk(t)
	if !hasOption(root, "[Quest] Errand") {
		t.Fatalf("nodes without an authored accept still offer [Quest]: %#v", root.Options)
	}
	topic := f.choose(t, "Tell me more.", root)
	if hasOption(topic, "[Quest] Errand") {
		t.Fatalf("generated [Quest] duplicates the authored accept: %#v", topic.Options)
	}
	f.choose(t, "I will do it.", topic)
	if f.status(t) != quests.QuestStatusActive {
		t.Fatalf("authored accept failed, status %q", f.status(t))
	}
}
