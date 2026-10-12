package game

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/talesmud/talesmud/pkg/entities/quests"
)

// OpQuestStep changes one character's quest progress.
// op=complete with an objective id only marks that step. It does not grant rewards.
// op=complete with no objective id marks every step, completes the quest, and runs
// the normal reward grant. Undo restores the progress row and does not claw rewards back.
func (g *Game) OpQuestStep(characterID, questID, objectiveID, op string) (*OpResult, error) {
	if characterID == "" || questID == "" {
		return nil, opErr(400, "characterId and questId are required")
	}
	op = strings.TrimSpace(strings.ToLower(op))
	if op == "" {
		return nil, opErr(400, "op is required")
	}
	if _, err := g.Facade.CharactersService().FindByID(characterID); err != nil {
		return nil, opErr(404, "character not found")
	}
	if _, err := g.Facade.QuestsService().FindByID(questID); err != nil {
		return nil, opErr(404, "quest not found")
	}
	beforeProg, _ := g.Facade.QuestsService().GetProgress(characterID, questID)
	beforeRaw := jsonRaw(beforeProg)
	flow := ""
	ready := false
	var err error
	switch op {
	case "complete":
		flow, ready, err = g.completeQuestStep(characterID, questID, objectiveID, beforeProg)
	case "reset":
		if objectiveID == "" {
			return nil, opErr(400, "objectiveId is required to reset a step")
		}
		err = g.resetQuestStep(characterID, questID, objectiveID, beforeProg)
	case "abandon":
		if err = g.Facade.QuestsService().AbandonQuest(characterID, questID); err != nil {
			return nil, opErr(400, err.Error())
		}
	case "reset-quest":
		if beforeProg == nil {
			return &OpResult{
				Summary:    "No quest progress to reset.",
				Undoable:   false,
				EntityType: "quest-progress",
				EntityID:   questID,
				Detail: map[string]interface{}{
					"characterId": characterID,
					"questId":     questID,
					"note":        "The character had no progress for this quest.",
				},
			}, nil
		}
		if err = g.Facade.QuestsService().ReplaceProgress(characterID, questID, nil); err != nil {
			return nil, opErr(500, err.Error())
		}
	default:
		return nil, opErr(400, "op must be complete, reset, abandon, or reset-quest")
	}
	if err != nil {
		return nil, err
	}
	afterProg, _ := g.Facade.QuestsService().GetProgress(characterID, questID)
	afterRaw := jsonRaw(afterProg)
	if flow == "" && afterProg != nil {
		ready = questReady(afterProg)
	}
	summary := opSummary("Quest %s: %s.", questID, op)
	if flow == "rewards" {
		summary = opSummary("Completed %s and granted its rewards. Undo restores progress only.", questID)
	} else if flow == "progress-only" {
		summary = opSummary("Marked %s on %s done. No rewards.", objectiveID, questID)
	}
	return &OpResult{
		Summary:    summary,
		Undoable:   true,
		EntityType: "quest-progress",
		EntityID:   characterID + ":" + questID,
		Before:     beforeRaw,
		After:      afterRaw,
		Inverse: inverseOf("restore-progress", map[string]interface{}{
			"characterId": characterID,
			"questId":     questID,
			"before":      json.RawMessage(beforeRaw),
			"expect":      json.RawMessage(afterRaw),
		}),
		Detail: map[string]interface{}{
			"characterId":    characterID,
			"questId":        questID,
			"objectiveId":    objectiveID,
			"op":             op,
			"completionFlow": flow,
			"readyToTurnIn":  ready,
		},
	}, nil
}

func (g *Game) completeQuestStep(characterID, questID, objectiveID string, progress *quests.QuestProgress) (string, bool, error) {
	progress, err := g.ensureQuestProgress(characterID, questID, progress)
	if err != nil {
		return "", false, err
	}
	if progress.Status == quests.QuestStatusCompleted {
		return "", false, opErr(409, "quest is already completed")
	}
	if objectiveID != "" {
		if !markObjective(progress, objectiveID, true) {
			return "", false, opErr(404, "objective not found")
		}
		if err := g.Facade.QuestsService().UpdateProgress(progress); err != nil {
			return "", false, opErr(500, err.Error())
		}
		return "progress-only", questReady(progress), nil
	}
	for i := range progress.Objectives {
		req := progress.Objectives[i].Required
		if req < 1 {
			req = 1
		}
		progress.Objectives[i].Required = req
		progress.Objectives[i].Current = req
		progress.Objectives[i].Completed = true
	}
	if err := g.Facade.QuestsService().UpdateProgress(progress); err != nil {
		return "", false, opErr(500, err.Error())
	}
	if _, err := g.Facade.QuestsService().CompleteQuest(characterID, questID); err != nil {
		return "", false, opErr(400, err.Error())
	}
	if _, _, err := g.Facade.QuestsService().GrantQuestRewards(characterID, questID); err != nil {
		return "", false, opErr(500, err.Error())
	}
	return "rewards", false, nil
}

func (g *Game) resetQuestStep(characterID, questID, objectiveID string, progress *quests.QuestProgress) error {
	if progress == nil {
		return opErr(404, "quest not found in quest log")
	}
	if !markObjective(progress, objectiveID, false) {
		return opErr(404, "objective not found")
	}
	if progress.Status == quests.QuestStatusCompleted {
		progress.Status = quests.QuestStatusActive
		progress.CompletedAt = time.Time{}
	}
	if err := g.Facade.QuestsService().UpdateProgress(progress); err != nil {
		return opErr(500, err.Error())
	}
	return nil
}

func (g *Game) ensureQuestProgress(characterID, questID string, progress *quests.QuestProgress) (*quests.QuestProgress, error) {
	if progress != nil && progress.Status == quests.QuestStatusActive {
		return progress, nil
	}
	if progress != nil && progress.Status == quests.QuestStatusCompleted {
		return progress, nil
	}
	accepted, err := g.Facade.QuestsService().AcceptQuest(characterID, questID)
	if err != nil {
		return nil, opErr(400, err.Error())
	}
	return accepted, nil
}

func markObjective(progress *quests.QuestProgress, objectiveID string, done bool) bool {
	if progress == nil {
		return false
	}
	for i := range progress.Objectives {
		if progress.Objectives[i].ObjectiveID != objectiveID {
			continue
		}
		req := progress.Objectives[i].Required
		if req < 1 {
			req = 1
		}
		progress.Objectives[i].Required = req
		if done {
			progress.Objectives[i].Current = req
			progress.Objectives[i].Completed = true
		} else {
			progress.Objectives[i].Current = 0
			progress.Objectives[i].Completed = false
		}
		return true
	}
	return false
}

func questReady(progress *quests.QuestProgress) bool {
	if progress == nil || progress.Status != quests.QuestStatusActive {
		return false
	}
	if len(progress.Objectives) == 0 {
		return true
	}
	for _, obj := range progress.Objectives {
		if !obj.Completed {
			return false
		}
	}
	return true
}
