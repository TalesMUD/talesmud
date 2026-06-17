package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
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
	c.JSON(http.StatusOK, gin.H{"summary": "Dialog preview is available.", "issues": []validation.Issue{}})
}

func (h *ValidationHandler) PreviewQuest(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"summary": "Quest preview is available.", "issues": []validation.Issue{}})
}

func (h *ValidationHandler) PreviewRoom(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"summary": "Room preview is available.", "issues": []validation.Issue{}})
}

func (h *ValidationHandler) PreviewMerchant(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"summary": "Merchant preview is available.", "issues": []validation.Issue{}})
}

func rejectIfInvalid(c *gin.Context, result validation.Result) bool {
	if result.Errors == 0 {
		return false
	}
	c.JSON(http.StatusBadRequest, result)
	return true
}
