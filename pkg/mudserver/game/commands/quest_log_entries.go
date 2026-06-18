package commands

import (
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func buildQuestLogEntries(game def.GameCtrl, progressList []*quests.QuestProgress) []messages.QuestLogEntry {
	entries := make([]messages.QuestLogEntry, 0, len(progressList))
	for _, progress := range progressList {
		entry := messages.QuestLogEntry{
			QuestID:    progress.QuestID,
			Status:     string(progress.Status),
			Objectives: buildObjectiveProgress(progress, nil),
		}

		quest, err := game.GetFacade().QuestsService().FindByID(progress.QuestID)
		if err == nil && quest != nil {
			entry.QuestName = quest.Name
			entry.Description = quest.Description
			entry.Category = quest.Category
			entry.Level = quest.Level
			entry.Objectives = buildObjectiveProgress(progress, quest)
			entry.Rewards = &messages.QuestReward{
				XP:              quest.Rewards.XP,
				Gold:            quest.Rewards.Gold,
				ItemTemplateIDs: quest.Rewards.ItemTemplateIDs,
			}
		}

		if !progress.AcceptedAt.IsZero() {
			entry.AcceptedAt = progress.AcceptedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		if !progress.CompletedAt.IsZero() {
			entry.CompletedAt = progress.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		}

		entries = append(entries, entry)
	}
	return entries
}

func buildObjectiveProgress(progress *quests.QuestProgress, quest *quests.Quest) []messages.QuestObjectiveProgress {
	objectives := make([]messages.QuestObjectiveProgress, len(progress.Objectives))
	for i, op := range progress.Objectives {
		objDesc := ""
		objRequired := op.Required
		if quest != nil {
			for _, questObj := range quest.Objectives {
				if questObj.ID == op.ObjectiveID {
					objDesc = questObj.Description
					if questObj.Amount > 0 {
						objRequired = questObj.Amount
					}
					break
				}
			}
		}

		objectives[i] = messages.QuestObjectiveProgress{
			ObjectiveID: op.ObjectiveID,
			Description: objDesc,
			Current:     op.Current,
			Required:    objRequired,
			Completed:   op.Completed,
		}
	}
	return objectives
}
