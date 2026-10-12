package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/contenthealth"
	"github.com/talesmud/talesmud/pkg/entities/characters"
)

// DebugQuest returns the static and live debugger view for one quest.
func (h *QuestsHandler) DebugQuest(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "quest id is required"})
		return
	}
	if h.Facade == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "quest debug is unavailable"})
		return
	}
	world, err := contenthealth.WorldFromFacade(h.Facade)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	characters, err := h.debugCharacters(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	view, ok := contenthealth.DebugQuest(world, id, characters)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "quest not found"})
		return
	}
	c.JSON(http.StatusOK, view)
}

func (h *QuestsHandler) debugCharacters(questID string) ([]contenthealth.DebugCharacterInput, error) {
	out := []contenthealth.DebugCharacterInput{}
	if h.Facade == nil || h.Facade.QuestsService() == nil {
		return out, nil
	}
	rows, err := h.Facade.QuestsService().ListProgress()
	if err != nil {
		return nil, err
	}
	byID := map[string]*characters.Character{}
	if h.Facade.CharactersService() != nil {
		chars, err := h.Facade.CharactersService().FindAll()
		if err != nil {
			return nil, err
		}
		for _, character := range chars {
			if character != nil && character.ID != "" {
				byID[character.ID] = character
			}
		}
	}
	guest := map[string]bool{}
	if h.Facade.UsersService() != nil {
		if users, err := h.Facade.UsersService().FindAll(); err == nil {
			for _, user := range users {
				if user != nil && user.IsGuest && user.ID != "" {
					guest[user.ID] = true
				}
			}
		}
	}
	for _, row := range rows {
		if row == nil || row.QuestID != questID {
			continue
		}
		in := contenthealth.DebugCharacterInput{ID: row.CharacterID, Progress: row}
		if character := byID[row.CharacterID]; character != nil {
			in.ID = character.ID
			in.Name = character.Name
			in.Level = character.Level
			in.CurrentRoomID = character.CurrentRoomID
			in.Guest = guest[character.BelongsUserID]
		}
		out = append(out, in)
	}
	return out, nil
}
