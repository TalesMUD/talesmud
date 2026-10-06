package commands

import (
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"strings"
)

func buildQuestLogEntries(game def.GameCtrl, progressList []*quests.QuestProgress) []messages.QuestLogEntry {
	entries := make([]messages.QuestLogEntry, 0, len(progressList))
	for _, progress := range progressList {
		quest, err := game.GetFacade().QuestsService().FindByID(progress.QuestID)
		if err == nil && quest != nil {
			entries = append(entries, buildQuestLogEntry(game, quest, progress))
			continue
		}

		entry := buildQuestLogEntry(game, nil, progress)
		entries = append(entries, entry)
	}
	return entries
}

func buildQuestLogEntry(game def.GameCtrl, quest *quests.Quest, progress *quests.QuestProgress) messages.QuestLogEntry {
	entry := messages.QuestLogEntry{
		QuestID:       progress.QuestID,
		Status:        string(progress.Status),
		ReadyToTurnIn: progress.Status == quests.QuestStatusActive && questLogObjectivesComplete(progress.Objectives),
		Objectives:    buildObjectiveProgress(progress, quest),
	}

	if quest != nil {
		entry.QuestName = quest.Name
		entry.Description = quest.Description
		entry.Category = quest.Category
		entry.Level = quest.Level
		entry.Rewards = &messages.QuestReward{
			XP:              quest.Rewards.XP,
			Gold:            quest.Rewards.Gold,
			ItemTemplateIDs: quest.Rewards.ItemTemplateIDs,
		}
		anywhere, npcID := quest.ResolveTurnIn()
		entry.TurnInAnywhere = anywhere
		entry.TurnInNpcID = npcID
		if npcID != "" {
			entry.TurnInNpcName, entry.TurnInRoomID = resolveTurnInNPC(game, npcID)
		}
	}

	if !progress.AcceptedAt.IsZero() {
		entry.AcceptedAt = progress.AcceptedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if !progress.CompletedAt.IsZero() {
		entry.CompletedAt = progress.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return entry
}

func questLogObjectivesComplete(objectives []quests.ObjectiveProgress) bool {
	if len(objectives) == 0 {
		return false
	}
	for _, objective := range objectives {
		if !objective.Completed {
			return false
		}
	}
	return true
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

func resolveTurnInNPC(game def.GameCtrl, npcID string) (name, roomID string) {
	npcID = strings.TrimSpace(npcID)
	if npcID == "" {
		return "", ""
	}
	// Prefer live instance location (unique NPC or spawned template instance).
	if game != nil {
		if mgr := game.GetNPCInstanceManager(); mgr != nil {
			if inst := mgr.GetInstance(npcID); inst != nil && !inst.IsDead {
				name = inst.Name
				if room := strings.TrimSpace(inst.CurrentRoomID); room != "" {
					return name, room
				}
				if room := strings.TrimSpace(inst.SpawnRoomID); room != "" {
					return name, room
				}
			}
			for _, inst := range mgr.GetAllInstances() {
				if inst == nil || inst.IsDead || inst.Entity == nil {
					continue
				}
				if inst.ID == npcID || inst.TemplateID == npcID {
					name = inst.Name
					if room := strings.TrimSpace(inst.CurrentRoomID); room != "" {
						return name, room
					}
					if room := strings.TrimSpace(inst.SpawnRoomID); room != "" {
						return name, room
					}
				}
			}
		}
		if facade := game.GetFacade(); facade != nil && facade.NPCsService() != nil {
			if n, err := facade.NPCsService().FindByID(npcID); err == nil && n != nil {
				name = n.Name
				if room := strings.TrimSpace(n.CurrentRoomID); room != "" {
					return name, room
				}
				return name, strings.TrimSpace(n.SpawnRoomID)
			}
		}
	}
	return name, roomID
}
