package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/service"
)

// QuestsHandler handles HTTP requests for quests
type QuestsHandler struct {
	Service service.QuestsService
}

// GetQuests returns all quest definitions
func (h *QuestsHandler) GetQuests(c *gin.Context) {
	result, err := h.Service.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetQuestByID returns a quest definition by ID
func (h *QuestsHandler) GetQuestByID(c *gin.Context) {
	id := c.Param("id")

	quest, err := h.Service.FindByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if quest == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "quest not found"})
		return
	}
	c.JSON(http.StatusOK, quest)
}

// PostQuest creates a new quest definition
func (h *QuestsHandler) PostQuest(c *gin.Context) {
	var quest quests.Quest
	if err := c.ShouldBindJSON(&quest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("name", quest.Name).Info("Creating new quest")

	newQuest, err := h.Service.Store(&quest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, newQuest)
}

// UpdateQuestByID updates a quest definition
func (h *QuestsHandler) UpdateQuestByID(c *gin.Context) {
	id := c.Param("id")
	var quest quests.Quest
	if err := c.ShouldBindJSON(&quest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("name", quest.Name).Info("Updating quest")

	if err := h.Service.Update(id, &quest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated quest"})
}

// DeleteQuestByID deletes a quest definition
func (h *QuestsHandler) DeleteQuestByID(c *gin.Context) {
	id := c.Param("id")

	if err := h.Service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// QuestLogEntry combines quest progress with quest definition for frontend
type QuestLogEntry struct {
	QuestID     string                      `json:"questId"`
	QuestName   string                      `json:"questName"`
	Status      string                      `json:"status"`
	Description string                      `json:"description"`
	Category    string                      `json:"category,omitempty"`
	Level       int32                       `json:"level,omitempty"`
	Objectives  []quests.ObjectiveProgress  `json:"objectives"`
	Rewards     *quests.Reward              `json:"rewards,omitempty"`
	AcceptedAt  string                      `json:"acceptedAt,omitempty"`
	CompletedAt string                      `json:"completedAt,omitempty"`
}

// GetQuestLog returns quest progress for a character with full quest details
func (h *QuestsHandler) GetQuestLog(c *gin.Context) {
	characterID := c.Param("characterId")

	progressList, err := h.Service.GetQuestLog(characterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build enriched quest log with quest definitions
	entries := make([]QuestLogEntry, 0, len(progressList))
	for _, progress := range progressList {
		quest, err := h.Service.FindByID(progress.QuestID)
		if err != nil || quest == nil {
			// Skip if quest definition not found
			continue
		}

		entry := QuestLogEntry{
			QuestID:     progress.QuestID,
			QuestName:   quest.Name,
			Status:      string(progress.Status),
			Description: quest.Description,
			Category:    quest.Category,
			Level:       quest.Level,
			Objectives:  progress.Objectives,
			Rewards:     &quest.Rewards,
		}

		if !progress.AcceptedAt.IsZero() {
			entry.AcceptedAt = progress.AcceptedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		if !progress.CompletedAt.IsZero() {
			entry.CompletedAt = progress.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		}

		entries = append(entries, entry)
	}

	c.JSON(http.StatusOK, entries)
}

// AcceptQuest accepts a quest for a character
func (h *QuestsHandler) AcceptQuest(c *gin.Context) {
	characterID := c.Param("characterId")
	questID := c.Param("questId")

	progress, err := h.Service.AcceptQuest(characterID, questID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, progress)
}

// AbandonQuest abandons a quest for a character
func (h *QuestsHandler) AbandonQuest(c *gin.Context) {
	characterID := c.Param("characterId")
	questID := c.Param("questId")

	if err := h.Service.AbandonQuest(characterID, questID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "quest abandoned"})
}

// CompleteQuest completes a quest and grants rewards
func (h *QuestsHandler) CompleteQuest(c *gin.Context) {
	characterID := c.Param("characterId")
	questID := c.Param("questId")

	// 1. Verify all objectives complete
	progress, err := h.Service.GetProgress(characterID, questID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if progress == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "quest not found in quest log"})
		return
	}

	allComplete := true
	for _, obj := range progress.Objectives {
		if !obj.Completed {
			allComplete = false
			break
		}
	}

	if !allComplete {
		c.JSON(http.StatusBadRequest, gin.H{"error": "quest objectives not complete"})
		return
	}

	// 2. Mark quest as completed
	progress, err = h.Service.CompleteQuest(characterID, questID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. Grant rewards
	grantedItems, err := h.Service.GrantQuestRewards(characterID, questID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. Get quest definition for reward info
	quest, _ := h.Service.FindByID(questID)

	// 5. Return completion result
	c.JSON(http.StatusOK, gin.H{
		"progress": progress,
		"rewards": gin.H{
			"xp":    quest.Rewards.XP,
			"gold":  quest.Rewards.Gold,
			"items": grantedItems,
		},
	})
}
