package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/settings"
	"github.com/talesmud/talesmud/pkg/service"
)

// ServerSettingsHandler handles server settings HTTP requests.
type ServerSettingsHandler struct {
	Service service.ServerSettingsService
}

// GetServerSettings returns the full server settings (protected).
func (h *ServerSettingsHandler) GetServerSettings(c *gin.Context) {
	result, err := h.Service.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// UpdateServerSettings updates the server settings (protected).
func (h *ServerSettingsHandler) UpdateServerSettings(c *gin.Context) {
	var s settings.ServerSettings
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Info("Updating server settings")

	if err := h.Service.Update(&s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated settings"})
}

// GetServerInfo returns minimal public server info (no auth required).
func (h *ServerSettingsHandler) GetServerInfo(c *gin.Context) {
	result, err := h.Service.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"serverName": result.ServerName,
	})
}
