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

// GetNPCs returns all NPCs with optional filtering
func (h *NPCsHandler) GetNPCs(c *gin.Context) {
	// Check for optional filters
	roomID := c.Query("roomID")
	isTemplateFilter := c.Query("isTemplate")

	var npcs []*npc.NPC
	var err error

	if roomID != "" {
		npcs, err = h.Service.FindByRoom(roomID)
	} else {
		npcs, err = h.Service.FindAll()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Filter by isTemplate if specified
	if isTemplateFilter != "" {
		wantTemplate := isTemplateFilter == "true"
		filtered := make([]*npc.NPC, 0)
		for _, n := range npcs {
			if n.IsTemplate == wantTemplate {
				filtered = append(filtered, n)
			}
		}
		npcs = filtered
	}

	c.JSON(http.StatusOK, npcs)
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

// GetNPCTemplates returns all NPC templates
func (h *NPCsHandler) GetNPCTemplates(c *gin.Context) {
	templates, err := h.Service.FindAllTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, templates)
}

// SpawnNPC creates an instance from a template
// Note: This only creates the NPC data, the caller must register it with the game's NPCManager
func (h *NPCsHandler) SpawnNPC(c *gin.Context) {
	templateID := c.Param("id")

	var req struct {
		RoomID string `json:"roomId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithFields(log.Fields{
		"template": templateID,
		"room":     req.RoomID,
	}).Info("Spawning NPC from template")

	instance, err := h.Service.SpawnFromTemplate(templateID, req.RoomID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, instance)
}
