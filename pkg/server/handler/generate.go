package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/service/groq"
)

// GenerateHandler handles AI text generation requests.
type GenerateHandler struct {
	GroqClient *groq.Client
}

// GenerateCharacterRequest is the JSON body for the generate endpoint.
type GenerateCharacterRequest struct {
	Generate     string `json:"generate" binding:"required"` // "name", "description", or "both"
	TemplateName string `json:"templateName"`
	Archetype    string `json:"archetype"`
	Race         string `json:"race"`
	Class        string `json:"class"`
	Description  string `json:"description"`
	Backstory    string `json:"backstory"`
	OriginArea   string `json:"originArea"`
	CurrentName  string `json:"currentName,omitempty"`
}

// GenerateCharacterResponse is the JSON response.
type GenerateCharacterResponse struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// GenerateCharacter generates a character name and/or description using the Groq LLM.
func (h *GenerateHandler) GenerateCharacter(c *gin.Context) {
	if h.GroqClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "AI generation is not configured",
		})
		return
	}

	var req GenerateCharacterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Generate = strings.ToLower(req.Generate)
	if req.Generate != "name" && req.Generate != "description" && req.Generate != "both" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "generate must be 'name', 'description', or 'both'",
		})
		return
	}

	resp := GenerateCharacterResponse{}

	if req.Generate == "name" || req.Generate == "both" {
		name, err := h.generateName(c, &req)
		if err != nil {
			log.WithError(err).Error("Failed to generate character name")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate name. Please try again.",
			})
			return
		}
		resp.Name = name
	}

	if req.Generate == "description" || req.Generate == "both" {
		nameForDesc := req.CurrentName
		if resp.Name != "" {
			nameForDesc = resp.Name
		}
		desc, err := h.generateDescription(c, &req, nameForDesc)
		if err != nil {
			log.WithError(err).Error("Failed to generate character description")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate description. Please try again.",
			})
			return
		}
		resp.Description = desc
	}

	c.JSON(http.StatusOK, resp)
}

func (h *GenerateHandler) generateName(c *gin.Context, req *GenerateCharacterRequest) (string, error) {
	systemPrompt := `You are a fantasy name generator for a MUD (multi-user dungeon) game.
Generate a single character name that fits the given archetype, race, and class.
Rules:
- Return ONLY the name, nothing else (no quotes, no explanation, no punctuation except hyphens or apostrophes within the name)
- The name should be 1-3 words maximum
- It should sound appropriate for the fantasy race and class
- It should be memorable and easy to type
- Do not use famous existing character names from popular media`

	userPrompt := buildNameUserPrompt(req)

	result, err := h.GroqClient.Complete(c.Request.Context(), systemPrompt, userPrompt, 0.9)
	if err != nil {
		return "", err
	}

	name := strings.TrimSpace(result)
	name = strings.Trim(name, "\"'`")
	if len(name) > 40 {
		name = name[:40]
	}
	return name, nil
}

func (h *GenerateHandler) generateDescription(c *gin.Context, req *GenerateCharacterRequest, characterName string) (string, error) {
	systemPrompt := `You are a creative writer for a MUD (multi-user dungeon) game.
Generate a brief character description for a player character.
Rules:
- Return ONLY the description, nothing else (no quotes, no labels, no explanation)
- Keep it to 1-2 sentences, maximum 180 characters
- Write in third person
- Focus on appearance, demeanor, or a notable trait
- Make it evocative and fitting for the archetype
- Do not include game mechanics or stats`

	userPrompt := buildDescriptionUserPrompt(req, characterName)

	result, err := h.GroqClient.Complete(c.Request.Context(), systemPrompt, userPrompt, 0.8)
	if err != nil {
		return "", err
	}

	desc := strings.TrimSpace(result)
	desc = strings.Trim(desc, "\"'`")
	if len(desc) > 200 {
		desc = desc[:200]
	}
	return desc, nil
}

func buildNameUserPrompt(req *GenerateCharacterRequest) string {
	var parts []string
	if req.TemplateName != "" {
		parts = append(parts, "Template: "+req.TemplateName)
	}
	if req.Archetype != "" {
		parts = append(parts, "Archetype: "+req.Archetype)
	}
	if req.Race != "" {
		parts = append(parts, "Race: "+req.Race)
	}
	if req.Class != "" {
		parts = append(parts, "Class: "+req.Class)
	}
	if req.OriginArea != "" {
		parts = append(parts, "Origin: "+req.OriginArea)
	}
	if req.Backstory != "" {
		parts = append(parts, "Backstory hint: "+req.Backstory)
	}
	if len(parts) == 0 {
		return "Generate a fantasy character name."
	}
	return "Generate a fantasy character name for:\n" + strings.Join(parts, "\n")
}

func buildDescriptionUserPrompt(req *GenerateCharacterRequest, characterName string) string {
	var parts []string
	if characterName != "" {
		parts = append(parts, "Character name: "+characterName)
	}
	if req.TemplateName != "" {
		parts = append(parts, "Template: "+req.TemplateName)
	}
	if req.Archetype != "" {
		parts = append(parts, "Archetype: "+req.Archetype)
	}
	if req.Race != "" {
		parts = append(parts, "Race: "+req.Race)
	}
	if req.Class != "" {
		parts = append(parts, "Class: "+req.Class)
	}
	if req.OriginArea != "" {
		parts = append(parts, "Origin: "+req.OriginArea)
	}
	if req.Backstory != "" {
		parts = append(parts, "Backstory: "+req.Backstory)
	}
	if len(parts) == 0 {
		return "Generate a brief fantasy character description."
	}
	return "Generate a brief fantasy character description for:\n" + strings.Join(parts, "\n")
}
