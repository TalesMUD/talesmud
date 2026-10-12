package contenthealth

import (
	"strings"
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
)

func TestDebugQuestRewardChainAndGraph(t *testing.T) {
	start := testRoom("R0", "Start")
	clerk := &npc.NPC{Entity: &entities.Entity{ID: "N1"}, Name: "Clerk", SpawnRoomID: "R0", DialogID: "DLG1"}
	item := &items.Item{Entity: &entities.Entity{ID: "ITM0059"}, Name: "Dispatch", IsTemplate: true}
	dialog := &dialogs.Dialog{
		Entity: &entities.Entity{ID: "DLG1"},
		Name:   "Clerk talk",
		NodeID: "root",
		Options: []*dialogs.Dialog{{
			NodeID:  "offer",
			Text:    "I will look.",
			QuestID: "QST0302",
			Action:  "accept",
		}, {
			NodeID:  "done",
			Text:    "Here it is.",
			QuestID: "QST0303",
			Action:  "complete",
		}},
	}
	grant := &quests.Quest{
		Entity:     &entities.Entity{ID: "QST0302"},
		Name:       "Grant",
		Category:   "side",
		Level:      2,
		Source:     quests.QuestSource{Type: "npc", NPCID: "N1"},
		Objectives: []quests.Objective{{ID: "see", Type: quests.ObjectiveVisit, Description: "look", TargetID: "R0", Amount: 1}},
		Rewards:    quests.Reward{XP: 40, Gold: 12, ItemTemplateIDs: []string{"ITM0059"}},
	}
	deliver := &quests.Quest{
		Entity:           &entities.Entity{ID: "QST0303"},
		Name:             "Deliver",
		Source:           quests.QuestSource{Type: "auto"},
		RequiredQuestIDs: []string{"QST0302"},
		TurnIn:           "npc:N1",
		Objectives: []quests.Objective{{
			ID: "drop", Type: quests.ObjectiveDeliver, Description: "hand it over", TargetID: "ITM0059", DeliverToNPCID: "N1", Amount: 1,
		}},
	}
	world := World{
		StartRoomID: "R0",
		Rooms:       []*rooms.Room{start},
		NPCs:        []*npc.NPC{clerk},
		Items:       []*items.Item{item},
		Dialogs:     []*dialogs.Dialog{dialog},
		Quests:      []*quests.Quest{grant, deliver},
	}
	accepted := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	view, ok := DebugQuest(world, "QST0303", []DebugCharacterInput{{
		ID: "C1", Name: "Warden", Level: 4, CurrentRoomID: "R0",
		Progress: &quests.QuestProgress{
			CharacterID: "C1", QuestID: "QST0303", Status: quests.QuestStatusActive,
			Objectives: []quests.ObjectiveProgress{{ObjectiveID: "drop", Current: 0, Required: 1}},
			AcceptedAt: accepted,
		},
	}})
	if !ok {
		t.Fatal("expected quest")
	}
	if view.Verdict != verdictGreen {
		t.Fatalf("verdict %s reason %s", view.Verdict, view.Reason)
	}
	var objective DebugStep
	found := false
	for _, step := range view.Steps {
		if step.Kind == "objective" {
			objective = step
			found = true
		}
	}
	if !found || objective.Verdict != verdictGreen {
		t.Fatalf("objective step %+v", objective)
	}
	reward := placeBy(objective.Places, "quest", "QST0302")
	if reward == nil || !reward.Reachable || reward.How != "quest reward" {
		t.Fatalf("reward place %+v", reward)
	}
	if len(view.Prerequisites) != 1 || view.Prerequisites[0].ID != "QST0302" {
		t.Fatalf("prereqs %+v", view.Prerequisites)
	}
	grantView, ok := DebugQuest(world, "QST0302", nil)
	if !ok || len(grantView.Next) != 1 || grantView.Next[0].ID != "QST0303" {
		t.Fatalf("next %+v ok %v", grantView.Next, ok)
	}
	offer := placeBy(grantView.Offer.Places, "dialog", "DLG1")
	if offer == nil || offer.NodeID != "offer" || offer.How != "accepts quest" {
		t.Fatalf("offer dialog %+v", grantView.Offer.Places)
	}
	turn := placeBy(view.TurnIn.Places, "dialog", "DLG1")
	if turn == nil || turn.NodeID != "done" {
		t.Fatalf("turn-in dialog %+v", view.TurnIn.Places)
	}
	if view.Rewards.XP != 0 {
		t.Fatalf("deliver quest has no rewards, got %+v", view.Rewards)
	}
	if grantView.Rewards.XP != 40 || grantView.Rewards.Gold != 12 || len(grantView.Rewards.Items) != 1 {
		t.Fatalf("grant rewards %+v", grantView.Rewards)
	}
	if len(view.Characters) != 1 || view.Characters[0].Step != "deliver 0/1" || view.Characters[0].Status != "active" || view.Characters[0].ObjectiveID != "drop" {
		t.Fatalf("characters %+v", view.Characters)
	}
	if view.Ops.QuestStep.Path != QuestStepOpPath || !view.Ops.QuestStep.Implemented || view.Ops.QuestStep.Method != "POST" {
		t.Fatalf("ops %+v", view.Ops.QuestStep)
	}
}

func TestDebugQuestUnreachableVisitIsRed(t *testing.T) {
	world := World{
		StartRoomID: "R0",
		Rooms:       []*rooms.Room{testRoom("R0", "Start"), testRoom("R1", "Island")},
		Quests: []*quests.Quest{{
			Entity:     &entities.Entity{ID: "Q1"},
			Name:       "Visit",
			Source:     quests.QuestSource{Type: "auto"},
			Objectives: []quests.Objective{{ID: "go", Type: quests.ObjectiveVisit, TargetID: "R1", Amount: 1}},
		}},
	}
	view, ok := DebugQuest(world, "Q1", nil)
	if !ok || view.Verdict != verdictRed || !strings.Contains(view.Reason, "unreachable") {
		t.Fatalf("view %+v", view)
	}
	room := placeBy(view.Steps[2].Places, "room", "R1")
	if room == nil || room.Reachable {
		t.Fatalf("room place %+v steps %+v", room, view.Steps)
	}
}

func TestDebugQuestCycleAndCustom(t *testing.T) {
	clerk := &npc.NPC{Entity: &entities.Entity{ID: "N1"}, Name: "Clerk", SpawnRoomID: "R0"}
	item := &items.Item{Entity: &entities.Entity{ID: "ITM0059"}, Name: "Dispatch", IsTemplate: true}
	grant := &quests.Quest{
		Entity:           &entities.Entity{ID: "QST0302"},
		Name:             "Grant",
		Source:           quests.QuestSource{Type: "auto"},
		RequiredQuestIDs: []string{"QST0303"},
		Objectives:       []quests.Objective{{ID: "see", Type: quests.ObjectiveVisit, TargetID: "R0", Amount: 1}},
		Rewards:          quests.Reward{ItemTemplateIDs: []string{"ITM0059"}},
	}
	deliver := &quests.Quest{
		Entity:           &entities.Entity{ID: "QST0303"},
		Name:             "Deliver",
		Source:           quests.QuestSource{Type: "auto"},
		RequiredQuestIDs: []string{"QST0302"},
		Objectives: []quests.Objective{
			{ID: "drop", Type: quests.ObjectiveDeliver, TargetID: "ITM0059", DeliverToNPCID: "N1", Amount: 1},
			{ID: "magic", Type: quests.ObjectiveCustom, Description: "scripted"},
		},
	}
	world := World{
		StartRoomID: "R0",
		Rooms:       []*rooms.Room{testRoom("R0", "Start")},
		NPCs:        []*npc.NPC{clerk},
		Items:       []*items.Item{item},
		Quests:      []*quests.Quest{grant, deliver},
	}
	view, ok := DebugQuest(world, "QST0303", nil)
	if !ok || view.Verdict != verdictRed {
		t.Fatalf("cycle verdict %s %s", view.Verdict, view.Reason)
	}
	var deliverStep, customStep DebugStep
	for _, step := range view.Steps {
		if step.ObjectiveID == "drop" {
			deliverStep = step
		}
		if step.ObjectiveID == "magic" {
			customStep = step
		}
	}
	reward := placeBy(deliverStep.Places, "quest", "QST0302")
	if reward == nil || reward.Reachable || !strings.Contains(reward.Detail, "precede") {
		t.Fatalf("cycle reward place %+v", reward)
	}
	if customStep.Verdict != verdictAmber || !strings.Contains(customStep.Reason, "not checked statically") {
		t.Fatalf("custom step %+v", customStep)
	}
	if _, ok := DebugQuest(world, "missing", nil); ok {
		t.Fatal("missing quest should not build")
	}
}

func placeBy(places []Place, kind, id string) *Place {
	for i := range places {
		if places[i].Type == kind && places[i].ID == id {
			return &places[i]
		}
	}
	return nil
}
