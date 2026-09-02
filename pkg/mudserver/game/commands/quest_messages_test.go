package commands

import (
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/quests"
)

func TestBuildQuestLogEntryIncludesDefinitionDetails(t *testing.T) {
	acceptedAt := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	quest := &quests.Quest{
		Entity:      &entities.Entity{ID: "quest-1"},
		Name:        "Rat Problem",
		Description: "Clear the cellar.",
		Category:    "side",
		Level:       2,
		Objectives: []quests.Objective{
			{ID: "kill-rats", Type: quests.ObjectiveKill, Description: "Kill 3 rats", Amount: 3},
		},
		Rewards: quests.Reward{XP: 25, Gold: 4},
	}
	progress := &quests.QuestProgress{
		QuestID:    quest.ID,
		Status:     quests.QuestStatusActive,
		AcceptedAt: acceptedAt,
		Objectives: []quests.ObjectiveProgress{
			{ObjectiveID: "kill-rats", Current: 1, Required: 3},
		},
	}

	entry := buildQuestLogEntry(nil, quest, progress)
	if entry.QuestName != quest.Name || entry.Description != quest.Description || entry.Objectives[0].Description != "Kill 3 rats" {
		t.Fatalf("entry missing quest details: %#v", entry)
	}
	if entry.Rewards == nil || entry.Rewards.XP != 25 || entry.Rewards.Gold != 4 {
		t.Fatalf("entry missing rewards: %#v", entry.Rewards)
	}
	if entry.AcceptedAt == "" {
		t.Fatal("entry missing accepted timestamp")
	}
}
