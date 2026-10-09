package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/contenthealth"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

// HealthHandler serves content-health reports for creators.
type HealthHandler struct {
	Facade  service.Facade
	Health  repository.ContentHealthRepository
	Loaders healthLoaders
}

type healthLoaders struct {
	room    func(string) (any, error)
	npc     func(string) (any, error)
	item    func(string) (any, error)
	loot    func(string) (any, error)
	spawner func(string) (any, error)
	dialog  func(string) (any, error)
	quest   func(string) (any, error)
	script  func(string) (any, error)
	skill   func(string) (any, error)
	tmpl    func(string) (any, error)
}

func (h *HealthHandler) loaders() healthLoaders {
	if h.Loaders.room != nil {
		return h.Loaders
	}
	return healthLoaders{
		room:    func(id string) (any, error) { return h.Facade.RoomsService().FindByID(id) },
		npc:     func(id string) (any, error) { return h.Facade.NPCsService().FindByID(id) },
		item:    func(id string) (any, error) { return h.Facade.ItemsService().FindByID(id) },
		loot:    func(id string) (any, error) { return h.Facade.LootTablesService().FindByID(id) },
		spawner: func(id string) (any, error) { return h.Facade.NPCSpawnersService().FindByID(id) },
		dialog:  func(id string) (any, error) { return h.Facade.DialogsService().FindByID(id) },
		quest:   func(id string) (any, error) { return h.Facade.QuestsService().FindByID(id) },
		script:  func(id string) (any, error) { return h.Facade.ScriptsService().FindByID(id) },
		skill:   func(id string) (any, error) { return h.Facade.SkillsService().FindByID(id) },
		tmpl:    func(id string) (any, error) { return h.Facade.CharacterTemplatesRepo().FindByID(id) },
	}
}

// Get returns the content-health report.
func (h *HealthHandler) Get(c *gin.Context) {
	report, err := contenthealth.RunFromFacade(h.Facade, h.Health)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}

// Drift returns entities that differ from the import baseline.
func (h *HealthHandler) Drift(c *gin.Context) {
	report, err := contenthealth.RunFromFacade(h.Facade, h.Health)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base, err := contenthealth.ReadBaseline(h.Health)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	body := gin.H{
		"contentCommit": report.ContentCommit,
		"changes":       report.Drift,
	}
	if base != nil {
		body["importedAt"] = base.ImportedAt
	}
	if body["changes"] == nil {
		body["changes"] = []contenthealth.DriftRow{}
	}
	c.JSON(http.StatusOK, body)
}

// Export returns one drifted entity as importer YAML.
func (h *HealthHandler) Export(c *gin.Context) {
	entityType := strings.TrimSpace(c.Query("type"))
	id := strings.TrimSpace(c.Query("id"))
	if entityType == "" || id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type and id are required"})
		return
	}
	entity, err := h.find(entityType, id)
	if err != nil || entity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "entity not found"})
		return
	}
	payload, err := contenthealth.MarshalYAML(entityType, entity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=\""+entityType+"-"+id+".yaml\"")
	c.Data(http.StatusOK, "application/yaml", payload)
}

func (h *HealthHandler) find(entityType, id string) (any, error) {
	loaders := h.loaders()
	switch strings.ToLower(strings.ReplaceAll(entityType, "_", "")) {
	case "room":
		return loaders.room(id)
	case "npc":
		return loaders.npc(id)
	case "item":
		return loaders.item(id)
	case "loottable":
		return loaders.loot(id)
	case "spawner":
		return loaders.spawner(id)
	case "dialog":
		return loaders.dialog(id)
	case "quest":
		return loaders.quest(id)
	case "script":
		return loaders.script(id)
	case "skill":
		return loaders.skill(id)
	case "charactertemplate":
		return loaders.tmpl(id)
	default:
		return nil, nil
	}
}

type muteRequest struct {
	RuleID string `json:"ruleId"`
	Muted  bool   `json:"muted"`
}

// Mute records whether a rule's hits are excluded from the failure counts.
func (h *HealthHandler) Mute(c *gin.Context) {
	var body muteRequest
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.RuleID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ruleId is required"})
		return
	}
	settings, err := h.Facade.ServerSettingsService().Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	next := make([]string, 0, len(settings.MutedHealthRuleIDs))
	seen := map[string]bool{}
	for _, id := range settings.MutedHealthRuleIDs {
		if id == "" || id == body.RuleID || seen[id] {
			continue
		}
		seen[id] = true
		next = append(next, id)
	}
	if body.Muted {
		next = append(next, body.RuleID)
	}
	settings.MutedHealthRuleIDs = next
	if err := h.Facade.ServerSettingsService().Update(settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ruleId": body.RuleID, "muted": body.Muted, "mutedHealthRuleIDs": next})
}
