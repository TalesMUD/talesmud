package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	npc "github.com/talesmud/talesmud/pkg/entities/npcs"
	"github.com/talesmud/talesmud/pkg/service"
)

// NPCsHandler handles NPC-related HTTP requests
type NPCsHandler struct {
	Service service.NPCsService
}

// GetNPCs returns all NPCs
func (h *NPCsHandler) GetNPCs(c *gin.Context) {
	// Check for optional roomID filter
	roomID := c.Query("roomID")

	var npcs []*npc.NPC
	var err error

	if roomID != "" {
		npcs, err = h.Service.FindByRoom(roomID)
	} else {
		npcs, err = h.Service.FindAll()
	}

	if err == nil {
		c.JSON(http.StatusOK, npcs)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// GetNPCByID returns a single NPC by ID
func (h *NPCsHandler) GetNPCByID(c *gin.Context) {
	id := c.Param("id")

	if npc, err := h.Service.FindByID(id); err == nil {
		c.JSON(http.StatusOK, npc)
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "NPC not found"})
	}
}

// PostNPC creates a new NPC
func (h *NPCsHandler) PostNPC(c *gin.Context) {
	var n npc.NPC
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("npc", n.Name).Info("Creating new NPC")

	if newNPC, err := h.Service.Store(&n); err == nil {
		c.JSON(http.StatusOK, newNPC)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// UpdateNPCByID updates an existing NPC
func (h *NPCsHandler) UpdateNPCByID(c *gin.Context) {
	id := c.Param("id")
	var n npc.NPC
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("npc", n.Name).Info("Updating NPC")

	if err := h.Service.Update(id, &n); err == nil {
		c.JSON(http.StatusOK, gin.H{"status": "updated NPC"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

// DeleteNPCByID deletes an NPC
func (h *NPCsHandler) DeleteNPCByID(c *gin.Context) {
	id := c.Param("id")

	if err := h.Service.Delete(id); err == nil {
		c.JSON(http.StatusOK, gin.H{"status": "deleted NPC"})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
