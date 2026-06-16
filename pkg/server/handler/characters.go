package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/server/dto"
	"github.com/talesmud/talesmud/pkg/service"
)

// CharactersHandler ...
type CharactersHandler struct {
	Service service.CharactersService
}

func canAccessCharacter(c *gin.Context, character *characters.Character) bool {
	if character == nil {
		return false
	}
	usr, exists := c.Get("user")
	if !exists {
		return false
	}
	user, ok := usr.(*entities.User)
	if !ok {
		return false
	}
	return user.IsAdmin() || character.BelongsUserID == user.ID
}

func currentUserFromContext(c *gin.Context) (*entities.User, bool) {
	usr, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	user, ok := usr.(*entities.User)
	return user, ok
}

// GetCharacters returns the list of item templates
func (csh *CharactersHandler) GetCharacters(c *gin.Context) {
	user, ok := currentUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	var (
		result []*characters.Character
		err    error
	)
	if user.IsAdmin() {
		result, err = csh.Service.FindAll()
	} else {
		result, err = csh.Service.FindAllForUser(user.ID)
	}

	if err == nil {
		c.JSON(http.StatusOK, result)
	} else {
		c.Error(err)
	}
}

// GetMyCharacters returns only the authenticated user's characters
func (csh *CharactersHandler) GetMyCharacters(c *gin.Context) {
	user, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userObj, ok := user.(*entities.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user object"})
		return
	}

	if characters, err := csh.Service.FindAllForUser(userObj.ID); err == nil {
		c.JSON(http.StatusOK, characters)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// GetCharacterTemplates returns the list of item templates
func (csh *CharactersHandler) GetCharacterTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, csh.Service.GetCharacterTemplates())
}

// GetCharacterByID returns a single charactersheet
func (csh *CharactersHandler) GetCharacterByID(c *gin.Context) {

	id := c.Param("id")

	if character, err := csh.Service.FindByID(id); err == nil {
		if !canAccessCharacter(c, character) {
			c.JSON(http.StatusForbidden, gin.H{"error": "character access denied"})
			return
		}
		c.JSON(http.StatusOK, character)
	} else {
		c.Error(err)
	}
}

// DeleteCharacterByID returns a single charactersheet
func (csh *CharactersHandler) DeleteCharacterByID(c *gin.Context) {

	id := c.Param("id")
	character, err := csh.Service.FindByID(id)
	if err != nil {
		c.Error(err)
		return
	}
	if !canAccessCharacter(c, character) {
		c.JSON(http.StatusForbidden, gin.H{"error": "character access denied"})
		return
	}

	if err := csh.Service.Delete(id); err == nil {
		c.JSON(http.StatusOK, "deleted")
	} else {
		c.Error(err)
	}
}

// UpdateCharacterByID creates a new charactersheet
func (csh *CharactersHandler) UpdateCharacterByID(c *gin.Context) {

	id := c.Param("id")
	existing, err := csh.Service.FindByID(id)
	if err != nil {
		c.Error(err)
		return
	}
	if !canAccessCharacter(c, existing) {
		c.JSON(http.StatusForbidden, gin.H{"error": "character access denied"})
		return
	}

	var character characters.Character
	if err := c.ShouldBindJSON(&character); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	character.BelongsUserID = existing.BelongsUserID

	log.WithField("character", character.Name).Info("Updating character")

	if err := csh.Service.Update(id, &character); err == nil {
		c.JSON(http.StatusOK, gin.H{"status": "updated character"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

// PostCharacter ... creates a new charactersheet
func (csh *CharactersHandler) PostCharacter(c *gin.Context) {

	var character characters.Character
	if err := c.ShouldBindJSON(&character); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("character", character.Name).Info("Creating new character")
	user, ok := currentUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}
	if character.BelongsUserID == "" {
		character.BelongsUserID = user.ID
	}
	if !user.IsAdmin() && character.BelongsUserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "character access denied"})
		return
	}

	if newCharacter, err := csh.Service.Store(&character); err == nil {
		c.JSON(http.StatusOK, newCharacter)
	} else {
		c.Error(err)
	}
}

// CreateNewCharacter ... creates a new charactersheet
func (csh *CharactersHandler) CreateNewCharacter(c *gin.Context) {

	var dto dto.CreateCharacterDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if user, ok := c.Get("user"); ok {
		if userObj, ok := user.(*entities.User); ok {
			dto.UserID = userObj.ID
		}
	}

	if dto.UserID == "" {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Could not create character: No UserID found"))
		return
	}

	log.WithField("character", dto.Name).Info("Creating new character based on template")

	if newCharacter, err := csh.Service.CreateNewCharacter(&dto); err == nil {
		c.JSON(http.StatusOK, newCharacter)
	} else {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
}
