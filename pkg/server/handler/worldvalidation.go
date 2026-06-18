package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/service"
)

// WorldValidationHandler exposes world content diagnostics to creators.
type WorldValidationHandler struct {
	Service service.WorldValidationService
}

func (h *WorldValidationHandler) GetWorldValidation(c *gin.Context) {
	report, err := h.Service.Validate()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}
