package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/service"
	"github.com/talesmud/talesmud/pkg/worldmap"
)

// CharacterMapHandler serves the per-character discovered-world atlas.
type CharacterMapHandler struct {
	Characters service.CharactersService
	Rooms      service.RoomsService
}

// GetCharacterMap returns GET /api/characters/:id/map
func (h *CharacterMapHandler) GetCharacterMap(c *gin.Context) {
	id := c.Param("id")
	character, err := h.Characters.FindByID(id)
	if err != nil || character == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
		return
	}
	if !canAccessCharacter(c, character) {
		c.JSON(http.StatusForbidden, gin.H{"error": "character access denied"})
		return
	}

	rs, err := h.Rooms.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	atlas := worldmap.Reveal(worldmap.Compile(rs), character)
	c.JSON(http.StatusOK, atlas)
}
