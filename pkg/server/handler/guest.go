package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/service"
)

// GuestHandler handles guest session creation.
type GuestHandler struct {
	GuestService service.GuestService
}

// CreateGuestSession creates a temporary guest account and returns a signed token.
func (h *GuestHandler) CreateGuestSession(c *gin.Context) {
	token, err := h.GuestService.CreateGuestSession(c.ClientIP())
	if err != nil {
		switch err.Error() {
		case "rate limit exceeded":
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		case "guest mode is disabled":
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case "server is full, try again later":
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create guest session"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":     token,
		"expiresIn": 1800, // 30 minutes in seconds
	})
}
