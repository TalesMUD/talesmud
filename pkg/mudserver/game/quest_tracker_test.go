package game

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func TestQuestProgressMessageTypeIsReadyToTurnInWhenObjectivesComplete(t *testing.T) {
	progress := &quests.QuestProgress{
		Objectives: []quests.ObjectiveProgress{
			{ObjectiveID: "obj1", Current: 1, Required: 1, Completed: true},
		},
	}
	if got := questProgressMessageType(progress); got != messages.MessageTypeQuestReady {
		t.Fatalf("expected questReady, got %s", got)
	}
}

func TestQuestProgressMessageTypeIsProgressWhenAnyObjectiveIncomplete(t *testing.T) {
	progress := &quests.QuestProgress{
		Objectives: []quests.ObjectiveProgress{
			{ObjectiveID: "obj1", Current: 1, Required: 1, Completed: true},
			{ObjectiveID: "obj2", Current: 0, Required: 1, Completed: false},
		},
	}
	if got := questProgressMessageType(progress); got != messages.MessageTypeQuestProgress {
		t.Fatalf("expected questProgress, got %s", got)
	}
}
