package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/service"
)

// GuestHandler handles guest session creation.
type GuestHandler struct {
	GuestService service.GuestService
}

// CreateGuestSession creates a temporary guest account and returns a signed token.
func (h *GuestHandler) CreateGuestSession(c *gin.Context) {
	var pick struct {
		TemplateID string `json:"templateId"`
		Race       string `json:"race"`
	}
	if c.Request.Body != nil {
		raw, readErr := io.ReadAll(c.Request.Body)
		if readErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid guest request"})
			return
		}
		if len(bytes.TrimSpace(raw)) > 0 {
			if err := json.Unmarshal(raw, &pick); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid guest request"})
				return
			}
		}
	}
	token, err := h.GuestService.CreateGuestSessionPick(c.ClientIP(), pick.TemplateID, pick.Race)
	if err != nil {
		if errors.Is(err, characters.ErrRaceNotAllowed) || errors.Is(err, characters.ErrRaceRequired) || errors.Is(err, characters.ErrGuestPickIncomplete) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
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
