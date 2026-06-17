package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/entities/items"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/entities/quests"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/scripts"
	"github.com/talesmud/talesmud/pkg/service"
	"github.com/talesmud/talesmud/pkg/service/validation"
)

type ValidationHandler struct {
	Facade service.Facade
}

func (h *ValidationHandler) WorldDiagnostics(c *gin.Context) {
	snapshot, err := validation.BuildSnapshot(h.Facade)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, validation.ValidateWorld(snapshot))
}

func (h *ValidationHandler) ValidateEntity(c *gin.Context) {
	entityType := c.Param("entityType")
	if !validation.IsSupportedEntityType(entityType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported entity type"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	snapshot, err := validation.BuildSnapshot(h.Facade)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result, err := validation.ValidateJSONEntity(entityType, body, snapshot)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *ValidationHandler) PreviewDialog(c *gin.Context) {
	var dialog dialogs.Dialog
	if err := c.ShouldBindJSON(&dialog); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	snapshot, ok := validationSnapshot(c, h.Facade)
	if !ok {
		return
	}
	result := validation.ValidateDialog(&dialog, snapshot)
	c.JSON(http.StatusOK, gin.H{
		"name":         dialog.Name,
		"nodeId":       dialog.NodeID,
		"text":         dialog.Text,
		"optionsCount": len(dialog.Options),
		"issues":       result.Issues,
	})
}

func (h *ValidationHandler) PreviewQuest(c *gin.Context) {
	var quest quests.Quest
	if err := c.ShouldBindJSON(&quest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	snapshot, ok := validationSnapshot(c, h.Facade)
	if !ok {
		return
	}
	result := validation.ValidateQuest(&quest, snapshot)
	c.JSON(http.StatusOK, gin.H{
		"name":            quest.Name,
		"source":          quest.Source,
		"objectivesCount": len(quest.Objectives),
		"rewards":         quest.Rewards,
		"issues":          result.Issues,
	})
}

func (h *ValidationHandler) PreviewRoom(c *gin.Context) {
	var room rooms.Room
	if err := c.ShouldBindJSON(&room); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	snapshot, ok := validationSnapshot(c, h.Facade)
	if !ok {
		return
	}
	result := validation.ValidateRoom(&room, snapshot)
	exitsCount, actionsCount, itemsCount, npcsCount := 0, 0, 0, 0
	if room.Exits != nil {
		exitsCount = len(*room.Exits)
	}
	if room.Actions != nil {
		actionsCount = len(*room.Actions)
	}
	if room.Items != nil {
		itemsCount = len(*room.Items)
	}
	if room.NPCs != nil {
		npcsCount = len(*room.NPCs)
	}
	c.JSON(http.StatusOK, gin.H{
		"name":         room.Name,
		"area":         room.Area,
		"description":  room.Description,
		"exitsCount":   exitsCount,
		"actionsCount": actionsCount,
		"itemsCount":   itemsCount,
		"npcsCount":    npcsCount,
		"issues":       result.Issues,
	})
}

func (h *ValidationHandler) PreviewMerchant(c *gin.Context) {
	var n npc.NPC
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	snapshot, ok := validationSnapshot(c, h.Facade)
	if !ok {
		return
	}
	result := validation.ValidateNPC(&n, snapshot)
	rows := []gin.H{}
	if n.MerchantTrait != nil {
		for _, stock := range n.MerchantTrait.Inventory {
			itemName := stock.ItemTemplateID
			basePrice := stock.BasePrice
			if item := snapshot.Items[stock.ItemTemplateID]; item != nil {
				itemName = item.Name
				if basePrice == 0 {
					basePrice = item.BasePrice
				}
			}
			rows = append(rows, gin.H{
				"itemTemplateId": stock.ItemTemplateID,
				"itemName":       itemName,
				"quantity":       stock.Quantity,
				"maxQuantity":    stock.MaxQuantity,
				"buyPrice":       n.MerchantTrait.GetBuyPrice(&stock, basePrice),
				"requiredLevel":  stock.RequiredLevel,
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"name":   n.Name,
		"stock":  rows,
		"issues": result.Issues,
	})
}

func rejectIfInvalid(c *gin.Context, result validation.Result) bool {
	if result.Errors == 0 {
		return false
	}
	c.JSON(http.StatusBadRequest, result)
	return true
}

func validationSnapshot(c *gin.Context, facade service.Facade) (validation.WorldSnapshot, bool) {
	snapshot, err := validation.BuildSnapshot(facade)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return validation.WorldSnapshot{}, false
	}
	return snapshot, true
}

func rejectInvalidRoom(c *gin.Context, facade service.Facade, room *rooms.Room) bool {
	snapshot, ok := validationSnapshot(c, facade)
	return !ok || rejectIfInvalid(c, validation.ValidateRoom(room, snapshot))
}

func rejectInvalidItem(c *gin.Context, facade service.Facade, item *items.Item) bool {
	snapshot, ok := validationSnapshot(c, facade)
	return !ok || rejectIfInvalid(c, validation.ValidateItem(item, snapshot))
}

func rejectInvalidNPC(c *gin.Context, facade service.Facade, n *npc.NPC) bool {
	snapshot, ok := validationSnapshot(c, facade)
	return !ok || rejectIfInvalid(c, validation.ValidateNPC(n, snapshot))
}

func rejectInvalidSpawner(c *gin.Context, facade service.Facade, spawner *npc.NPCSpawner) bool {
	snapshot, ok := validationSnapshot(c, facade)
	return !ok || rejectIfInvalid(c, validation.ValidateSpawner(spawner, snapshot))
}

func rejectInvalidDialog(c *gin.Context, facade service.Facade, dialog *dialogs.Dialog) bool {
	snapshot, ok := validationSnapshot(c, facade)
	return !ok || rejectIfInvalid(c, validation.ValidateDialog(dialog, snapshot))
}

func rejectInvalidLootTable(c *gin.Context, facade service.Facade, table *items.LootTable) bool {
	snapshot, ok := validationSnapshot(c, facade)
	return !ok || rejectIfInvalid(c, validation.ValidateLootTable(table, snapshot))
}

func rejectInvalidQuest(c *gin.Context, facade service.Facade, quest *quests.Quest) bool {
	snapshot, ok := validationSnapshot(c, facade)
	return !ok || rejectIfInvalid(c, validation.ValidateQuest(quest, snapshot))
}

func rejectInvalidScript(c *gin.Context, facade service.Facade, script *scripts.Script) bool {
	snapshot, ok := validationSnapshot(c, facade)
	return !ok || rejectIfInvalid(c, validation.ValidateScript(script, snapshot, false))
}
