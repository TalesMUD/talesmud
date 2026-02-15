package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/skills"
	"github.com/talesmud/talesmud/pkg/service"
)

// SkillsHandler handles HTTP requests for skills
type SkillsHandler struct {
	Service service.SkillsService
}

// GetSkills returns all skill definitions
func (h *SkillsHandler) GetSkills(c *gin.Context) {
	result, err := h.Service.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetSkillByID returns a skill definition by ID
func (h *SkillsHandler) GetSkillByID(c *gin.Context) {
	id := c.Param("id")

	skill, err := h.Service.FindByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if skill == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "skill not found"})
		return
	}
	c.JSON(http.StatusOK, skill)
}

// PostSkill creates a new skill definition
func (h *SkillsHandler) PostSkill(c *gin.Context) {
	var skill skills.Skill
	if err := c.ShouldBindJSON(&skill); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("name", skill.Name).Info("Creating new skill")

	newSkill, err := h.Service.Store(&skill)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, newSkill)
}

// UpdateSkillByID updates a skill definition
func (h *SkillsHandler) UpdateSkillByID(c *gin.Context) {
	id := c.Param("id")
	var skill skills.Skill
	if err := c.ShouldBindJSON(&skill); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("name", skill.Name).Info("Updating skill")

	if err := h.Service.Update(id, &skill); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updated skill"})
}

// DeleteSkillByID deletes a skill definition
func (h *SkillsHandler) DeleteSkillByID(c *gin.Context) {
	id := c.Param("id")

	if err := h.Service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
