package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/repository"
)

// CharacterTemplatesHandler handles character template CRUD operations
type CharacterTemplatesHandler struct {
	Repo      repository.CharacterTemplatesRepository
	ItemsRepo repository.ItemsRepository
}

// GetCharacterTemplates returns all character templates
func (h *CharacterTemplatesHandler) GetCharacterTemplates(c *gin.Context) {
	if templates, err := h.Repo.FindAll(); err == nil {
		c.JSON(http.StatusOK, templates)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// GetCharacterTemplateByID returns a single character template
func (h *CharacterTemplatesHandler) GetCharacterTemplateByID(c *gin.Context) {
	id := c.Param("id")
	if template, err := h.Repo.FindByID(id); err == nil {
		c.JSON(http.StatusOK, template)
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
	}
}

// PostCharacterTemplate creates a new character template
func (h *CharacterTemplatesHandler) PostCharacterTemplate(c *gin.Context) {
	var template characters.CharacterTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("template", template.Name).Info("Creating new character template")

	if newTemplate, err := h.Repo.Store(&template); err == nil {
		c.JSON(http.StatusOK, newTemplate)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// UpdateCharacterTemplateByID updates an existing character template
func (h *CharacterTemplatesHandler) UpdateCharacterTemplateByID(c *gin.Context) {
	id := c.Param("id")
	var template characters.CharacterTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("template", template.Name).Info("Updating character template")

	if err := h.Repo.Update(id, &template); err == nil {
		c.JSON(http.StatusOK, gin.H{"status": "updated character template"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

// DeleteCharacterTemplateByID deletes a character template
func (h *CharacterTemplatesHandler) DeleteCharacterTemplateByID(c *gin.Context) {
	id := c.Param("id")
	if err := h.Repo.Delete(id); err == nil {
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// SeedCharacterTemplates seeds the database with default system presets
// Also seeds starter item templates if they don't exist
func (h *CharacterTemplatesHandler) SeedCharacterTemplates(c *gin.Context) {
	count, err := h.Repo.Count()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if count > 0 {
		c.JSON(http.StatusOK, gin.H{"status": "skipped", "message": "templates already exist", "count": count})
		return
	}

	// First, seed item templates and build a name->ID map
	itemNameToID := make(map[string]string)
	itemTemplatesCreated := 0

	starterItems := items.StarterItemTemplatePresets()
	for _, itemTemplate := range starterItems {
		// Check if item template with this name already exists
		existing, _ := h.ItemsRepo.FindTemplateByName(itemTemplate.Name)
		if len(existing) > 0 {
			itemNameToID[itemTemplate.Name] = existing[0].ID
			log.WithField("name", itemTemplate.Name).Info("Item template already exists, using existing ID")
		} else {
			// Create the item template
			itemTemplate.IsTemplate = true // Ensure flag is set
			created, err := h.ItemsRepo.Store(itemTemplate)
			if err != nil {
				log.WithError(err).WithField("name", itemTemplate.Name).Error("Failed to seed item template")
				continue
			}
			itemNameToID[itemTemplate.Name] = created.ID
			itemTemplatesCreated++
			log.WithField("name", created.Name).WithField("id", created.ID).Info("Seeded item template")
		}
	}

	// Now seed character templates with resolved item template IDs
	presets := characters.SystemCharacterTemplatePresets()
	charTemplatesCreated := 0

	for _, preset := range presets {
		// Resolve starting item template names to IDs
		for i, startingItem := range preset.StartingItems {
			if startingItem.ItemTemplateName != "" && startingItem.ItemTemplateID == "" {
				if id, ok := itemNameToID[startingItem.ItemTemplateName]; ok {
					preset.StartingItems[i].ItemTemplateID = id
				} else {
					log.WithField("name", startingItem.ItemTemplateName).Warn("Could not resolve item template name to ID")
				}
			}
		}

		if _, err := h.Repo.Store(preset); err != nil {
			log.WithError(err).WithField("preset", preset.Name).Error("Failed to seed character template")
		} else {
			charTemplatesCreated++
		}
	}

	log.WithField("charTemplates", charTemplatesCreated).WithField("itemTemplates", itemTemplatesCreated).Info("Seeded templates from system presets")
	c.JSON(http.StatusOK, gin.H{
		"status":                 "seeded",
		"characterTemplates":     charTemplatesCreated,
		"itemTemplatesCreated":   itemTemplatesCreated,
	})
}

// GetCharacterTemplatePresets returns the hardcoded system presets (for reference/re-seeding)
func (h *CharacterTemplatesHandler) GetCharacterTemplatePresets(c *gin.Context) {
	c.JSON(http.StatusOK, characters.SystemCharacterTemplatePresets())
}
