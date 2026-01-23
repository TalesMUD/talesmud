package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/service"
)

// NPCSpawnersHandler handles NPC spawner-related HTTP requests
type NPCSpawnersHandler struct {
	Service service.NPCSpawnersService
}

// GetSpawners returns all NPC spawners
func (h *NPCSpawnersHandler) GetSpawners(c *gin.Context) {
	// Check for optional filters
	roomID := c.Query("roomId")
	templateID := c.Query("templateId")

	var spawners []*npc.NPCSpawner
	var err error

	if roomID != "" {
		spawners, err = h.Service.FindByRoom(roomID)
	} else if templateID != "" {
		spawners, err = h.Service.FindByTemplate(templateID)
	} else {
		spawners, err = h.Service.FindAll()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, spawners)
}

// GetSpawnerByID returns a single spawner by ID
func (h *NPCSpawnersHandler) GetSpawnerByID(c *gin.Context) {
	id := c.Param("id")

	spawner, err := h.Service.FindByID(id)
	if err != nil || spawner == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Spawner not found"})
		return
	}
	c.JSON(http.StatusOK, spawner)
}

// PostSpawner creates a new NPC spawner
func (h *NPCSpawnersHandler) PostSpawner(c *gin.Context) {
	var s npc.NPCSpawner
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithFields(log.Fields{
		"template": s.TemplateID,
		"room":     s.RoomID,
		"max":      s.MaxInstances,
	}).Info("Creating new NPC spawner")

	newSpawner, err := h.Service.Store(&s)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, newSpawner)
}

// UpdateSpawnerByID updates an existing spawner
func (h *NPCSpawnersHandler) UpdateSpawnerByID(c *gin.Context) {
	id := c.Param("id")
	var s npc.NPCSpawner
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("spawner", id).Info("Updating NPC spawner")

	if err := h.Service.Update(id, &s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated spawner"})
}

// DeleteSpawnerByID deletes a spawner
func (h *NPCSpawnersHandler) DeleteSpawnerByID(c *gin.Context) {
	id := c.Param("id")

	log.WithField("spawner", id).Info("Deleting NPC spawner")

	if err := h.Service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted spawner"})
}
