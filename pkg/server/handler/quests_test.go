package handler

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/quests"
)

func TestBuildQuestLogObjectivesAddsDefinitionText(t *testing.T) {
	quest := &quests.Quest{
		Objectives: []quests.Objective{
			{
				ID:          "visit-square",
				Description: "Visit the market square",
				Amount:      1,
			},
		},
	}
	progress := &quests.QuestProgress{
		Objectives: []quests.ObjectiveProgress{
			{
				ObjectiveID: "visit-square",
				Current:     1,
				Completed:   true,
			},
		},
	}

	objectives := buildQuestLogObjectives(quest, progress)
	if len(objectives) != 1 {
		t.Fatalf("expected 1 objective, got %d", len(objectives))
	}
	if objectives[0].Description != "Visit the market square" {
		t.Fatalf("expected objective description from quest definition, got %q", objectives[0].Description)
	}
	if objectives[0].Required != 1 {
		t.Fatalf("expected required count from quest definition, got %d", objectives[0].Required)
	}
}
