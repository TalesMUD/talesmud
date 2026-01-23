package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/service"
)

// LootTablesHandler handles HTTP requests for loot tables
type LootTablesHandler struct {
	Service service.LootTablesService
}

// GetLootTables returns all loot tables
func (h *LootTablesHandler) GetLootTables(c *gin.Context) {
	result, err := h.Service.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetLootTableByID returns a loot table by ID
func (h *LootTablesHandler) GetLootTableByID(c *gin.Context) {
	id := c.Param("id")

	lootTable, err := h.Service.FindByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if lootTable == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "loot table not found"})
		return
	}
	c.JSON(http.StatusOK, lootTable)
}

// PostLootTable creates a new loot table
func (h *LootTablesHandler) PostLootTable(c *gin.Context) {
	var lootTable items.LootTable
	if err := c.ShouldBindJSON(&lootTable); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("name", lootTable.Name).Info("Creating new loot table")

	newLootTable, err := h.Service.Store(&lootTable)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, newLootTable)
}

// UpdateLootTableByID updates a loot table
func (h *LootTablesHandler) UpdateLootTableByID(c *gin.Context) {
	id := c.Param("id")
	var lootTable items.LootTable
	if err := c.ShouldBindJSON(&lootTable); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("name", lootTable.Name).Info("Updating loot table")

	if err := h.Service.Update(id, &lootTable); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated loot table"})
}

// DeleteLootTableByID deletes a loot table
func (h *LootTablesHandler) DeleteLootTableByID(c *gin.Context) {
	id := c.Param("id")

	if err := h.Service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// RollLootTable performs a test roll against a loot table
// Query params: ?playerLevel=1&baseGold=100
func (h *LootTablesHandler) RollLootTable(c *gin.Context) {
	id := c.Param("id")

	// Parse optional parameters
	playerLevel := int32(1)
	if lvl := c.Query("playerLevel"); lvl != "" {
		if parsed := parsePositiveInt32(lvl); parsed > 0 {
			playerLevel = parsed
		}
	}

	baseGold := int64(0)
	if gold := c.Query("baseGold"); gold != "" {
		if parsed := parsePositiveInt64(gold); parsed > 0 {
			baseGold = parsed
		}
	}

	result, err := h.Service.RollLoot(id, playerLevel, baseGold)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": result.Items,
		"gold":  result.Gold,
	})
}

// parsePositiveInt32 parses a positive int32 from string
func parsePositiveInt32(s string) int32 {
	var n int32
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int32(c-'0')
	}
	return n
}

// parsePositiveInt64 parses a positive int64 from string
func parsePositiveInt64(s string) int64 {
	var n int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int64(c-'0')
	}
	return n
}
