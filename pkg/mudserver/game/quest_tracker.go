package game

import (
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	def "github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/service"
)

// QuestTracker listens to game events and updates quest progress
type QuestTracker struct {
	facade service.Facade
	game   def.GameCtrl
}

// NewQuestTracker creates a new quest tracker
func NewQuestTracker(facade service.Facade, game def.GameCtrl) *QuestTracker {
	return &QuestTracker{
		facade: facade,
		game:   game,
	}
}

// OnNPCKilled is called when an NPC dies in combat
func (qt *QuestTracker) OnNPCKilled(characterID, userID string, deadNPC *npc.NPC) {
	if deadNPC == nil {
		return
	}
	progressList, err := qt.facade.QuestsService().GetQuestLog(characterID)
	if err != nil {
		return
	}

	for _, progress := range progressList {
		if progress.Status != quests.QuestStatusActive {
			continue
		}

		quest, err := qt.facade.QuestsService().FindByID(progress.QuestID)
		if err != nil || quest == nil {
			continue
		}

		for _, obj := range quest.Objectives {
			if obj.Type != quests.ObjectiveKill {
				continue
			}
			if !quests.KillTargetMatches(obj, deadNPC.TemplateID, deadNPC.ID, deadNPC.Name) {
				continue
			}

			updated, err := qt.facade.QuestsService().IncrementObjective(characterID, progress.QuestID, obj.ID, 1)
			if err != nil {
				log.WithError(err).Error("Failed to increment kill objective")
				continue
			}

			qt.sendProgressUpdate(userID, quest, updated, obj.ID)
		}
	}
}

// OnItemPickup is called when a player picks up an item
func (qt *QuestTracker) OnItemPickup(characterID, userID string, item *items.Item) {
	progressList, err := qt.facade.QuestsService().GetQuestLog(characterID)
	if err != nil {
		return
	}

	templateID := item.TemplateID
	if templateID == "" {
		templateID = item.ID
	}

	for _, progress := range progressList {
		if progress.Status != quests.QuestStatusActive {
			continue
		}

		quest, err := qt.facade.QuestsService().FindByID(progress.QuestID)
		if err != nil || quest == nil {
			continue
		}

		for _, obj := range quest.Objectives {
			if obj.Type != quests.ObjectiveCollect {
				continue
			}
			if obj.TargetID != templateID {
				continue
			}

			amount := item.Quantity
			if amount < 1 {
				amount = 1
			}
			updated, err := qt.facade.QuestsService().IncrementObjective(characterID, progress.QuestID, obj.ID, amount)
			if err != nil {
				log.WithError(err).Error("Failed to increment collect objective")
				continue
			}

			qt.sendProgressUpdate(userID, quest, updated, obj.ID)
		}
	}
}

// OnRoomEnter is called when a player enters a room
func (qt *QuestTracker) OnRoomEnter(characterID, userID string, room *rooms.Room) {
	progressList, err := qt.facade.QuestsService().GetQuestLog(characterID)
	if err != nil {
		return
	}

	for _, progress := range progressList {
		if progress.Status != quests.QuestStatusActive {
			continue
		}

		quest, err := qt.facade.QuestsService().FindByID(progress.QuestID)
		if err != nil || quest == nil {
			continue
		}

		for _, obj := range quest.Objectives {
			if obj.Type != quests.ObjectiveVisit {
				continue
			}
			if !quests.RoomIDMatches(obj.TargetID, room.ID) {
				continue
			}

			updated, err := qt.facade.QuestsService().IncrementObjective(characterID, progress.QuestID, obj.ID, 1)
			if err != nil {
				log.WithError(err).Error("Failed to increment visit objective")
				continue
			}

			qt.sendProgressUpdate(userID, quest, updated, obj.ID)
		}
	}
}

// OnDialogNode is called when a player reaches a dialog node
func (qt *QuestTracker) OnDialogNode(characterID, userID, npcID, dialogID, nodeID string) {
	progressList, err := qt.facade.QuestsService().GetQuestLog(characterID)
	if err != nil {
		return
	}

	for _, progress := range progressList {
		if progress.Status != quests.QuestStatusActive {
			continue
		}

		quest, err := qt.facade.QuestsService().FindByID(progress.QuestID)
		if err != nil || quest == nil {
			continue
		}

		for _, obj := range quest.Objectives {
			if obj.Type != quests.ObjectiveTalk {
				continue
			}

			// Match by NPC ID or dialog node ID
			npcMatch := obj.TargetID == npcID || obj.TargetID == ""
			nodeMatch := obj.DialogNodeID == "" || obj.DialogNodeID == nodeID
			if !npcMatch || !nodeMatch {
				continue
			}

			updated, err := qt.facade.QuestsService().IncrementObjective(characterID, progress.QuestID, obj.ID, 1)
			if err != nil {
				log.WithError(err).Error("Failed to increment talk objective")
				continue
			}

			qt.sendProgressUpdate(userID, quest, updated, obj.ID)
		}
	}
}

// OnTalkToNPC is called when a player talks to an NPC (for deliver objectives)
func (qt *QuestTracker) OnTalkToNPC(characterID, userID string, talkNPC *npc.NPC) {
	progressList, err := qt.facade.QuestsService().GetQuestLog(characterID)
	if err != nil {
		return
	}

	for _, progress := range progressList {
		if progress.Status != quests.QuestStatusActive {
			continue
		}

		quest, err := qt.facade.QuestsService().FindByID(progress.QuestID)
		if err != nil || quest == nil {
			continue
		}

		for _, obj := range quest.Objectives {
			if obj.Type != quests.ObjectiveDeliver {
				continue
			}

			npcTemplateID := talkNPC.TemplateID
			if npcTemplateID == "" {
				npcTemplateID = talkNPC.ID
			}

			if obj.DeliverToNPCID != npcTemplateID && obj.DeliverToNPCID != talkNPC.ID {
				continue
			}

			objectiveCompleted := false
			for _, op := range progress.Objectives {
				if op.ObjectiveID == obj.ID && op.Completed {
					objectiveCompleted = true
					break
				}
			}
			if objectiveCompleted {
				continue
			}

			updated, err := qt.facade.QuestsService().CompleteDeliveryObjective(characterID, progress.QuestID, obj.ID)
			if err != nil {
				continue
			}

			qt.sendProgressUpdate(userID, quest, updated, obj.ID)
		}
	}
}

// sendProgressUpdate sends a quest progress update to the player
func (qt *QuestTracker) sendProgressUpdate(userID string, quest *quests.Quest, progress *quests.QuestProgress, changedObjectiveID string) {
	if progress == nil || quest == nil {
		return
	}

	objectives := qt.buildObjectiveProgress(quest, progress)
	msgType := questProgressMessageType(progress)
	msg := messages.NewQuestUpdateMessage(userID, quest.ID, quest.Name, string(msgType), objectives)

	for i := range objectives {
		if objectives[i].ObjectiveID == changedObjectiveID {
			msg.ChangedObjective = &objectives[i]
			break
		}
	}

	qt.game.SendMessage() <- msg
}

func questProgressMessageType(progress *quests.QuestProgress) messages.MessageType {
	for _, obj := range progress.Objectives {
		if !obj.Completed {
			return messages.MessageTypeQuestProgress
		}
	}
	return messages.MessageTypeQuestReady
}

// buildObjectiveProgress builds the objective progress list for messages
func (qt *QuestTracker) buildObjectiveProgress(quest *quests.Quest, progress *quests.QuestProgress) []messages.QuestObjectiveProgress {
	result := make([]messages.QuestObjectiveProgress, 0)

	for _, obj := range quest.Objectives {
		qop := messages.QuestObjectiveProgress{
			ObjectiveID: obj.ID,
			Description: obj.Description,
			Required:    obj.Amount,
		}
		if qop.Required < 1 {
			qop.Required = 1
		}

		// Find matching progress
		for _, op := range progress.Objectives {
			if op.ObjectiveID == obj.ID {
				qop.Current = op.Current
				qop.Completed = op.Completed
				break
			}
		}

		result = append(result, qop)
	}

	return result
}
