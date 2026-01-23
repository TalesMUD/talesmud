package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/dialogs"
	"github.com/talesmud/talesmud/pkg/service"
)

// DialogsHandler handles dialog-related HTTP requests
type DialogsHandler struct {
	Service service.DialogsService
}

// GetDialogs returns all dialogs
func (h *DialogsHandler) GetDialogs(c *gin.Context) {
	if dialogs, err := h.Service.FindAll(); err == nil {
		c.JSON(http.StatusOK, dialogs)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// GetDialogByID returns a single dialog by ID
func (h *DialogsHandler) GetDialogByID(c *gin.Context) {
	id := c.Param("id")

	if dialog, err := h.Service.FindByID(id); err == nil {
		c.JSON(http.StatusOK, dialog)
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "dialog not found"})
	}
}

// PostDialog creates a new dialog
func (h *DialogsHandler) PostDialog(c *gin.Context) {
	var dialog dialogs.Dialog
	if err := c.ShouldBindJSON(&dialog); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("dialog", dialog.Name).Info("Creating new dialog")

	if newDialog, err := h.Service.Store(&dialog); err == nil {
		c.JSON(http.StatusOK, newDialog)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// UpdateDialogByID updates an existing dialog
func (h *DialogsHandler) UpdateDialogByID(c *gin.Context) {
	id := c.Param("id")
	var dialog dialogs.Dialog
	if err := c.ShouldBindJSON(&dialog); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("dialog", dialog.Name).Info("Updating dialog")

	if err := h.Service.Update(id, &dialog); err == nil {
		c.JSON(http.StatusOK, gin.H{"status": "updated dialog"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

// DeleteDialogByID deletes a dialog
func (h *DialogsHandler) DeleteDialogByID(c *gin.Context) {
	id := c.Param("id")

	if err := h.Service.Delete(id); err == nil {
		c.JSON(http.StatusOK, gin.H{"status": "deleted dialog"})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
