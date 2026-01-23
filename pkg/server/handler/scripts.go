package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/scripts"
	s "github.com/talesmud/talesmud/pkg/scripts"
	"github.com/talesmud/talesmud/pkg/service"
)

//ScriptsHandler ...
type ScriptsHandler struct {
	Service service.ScriptsService
	Runner  s.ScriptRunner
}

//GetScripts returns the list of scripts
func (handler *ScriptsHandler) GetScripts(c *gin.Context) {
	if scripts, err := handler.Service.FindAll(); err == nil {
		c.JSON(http.StatusOK, scripts)
	} else {
		c.Error(err)
	}
}

//PostScript ... creates a new charactersheet
func (handler *ScriptsHandler) PostScript(c *gin.Context) {

	var script s.Script
	if err := c.ShouldBindJSON(&script); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if script.Language == "" {
		script.Language = scripts.ScriptLanguageLua
	}
	if script.Language != scripts.ScriptLanguageLua {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only Lua scripts are supported"})
		return
	}

	log.WithField("script", script.Name).Info("Creating new script")

	if script, err := handler.Service.Store(&script); err == nil {
		c.JSON(http.StatusOK, script)
	} else {
		c.Error(err)
	}
}

//ExecuteScript ... Updates a script
func (handler *ScriptsHandler) ExecuteScript(c *gin.Context) {

	id := c.Param("id")

	if script, err := handler.Service.FindByID(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
			return
		}

		scriptCtx := scripts.NewScriptContext()
		if len(body) > 0 {
			var payload interface{}
			if err := json.Unmarshal(body, &payload); err == nil {
				scriptCtx.Set("ctx", payload)
			} else {
				scriptCtx.Set("ctx", string(body))
			}
		} else {
			scriptCtx.Set("ctx", nil)
		}

		result := handler.Runner.RunWithResult(*script, scriptCtx)

		c.JSON(http.StatusOK, gin.H{
			"success":    result.Success,
			"result":     result.Result,
			"error":      result.Error,
			"duration":   result.Duration.String(),
			"durationMs": result.Duration.Milliseconds(),
		})

	}

}

//GetScriptTypes ...
func (handler *ScriptsHandler) GetScriptTypes(c *gin.Context) {
	c.JSON(http.StatusOK, handler.Service.ScriptTypes())
}

//PutScript ... Updates a script
func (handler *ScriptsHandler) PutScript(c *gin.Context) {

	id := c.Param("id")
	var script scripts.Script
	if err := c.ShouldBindJSON(&script); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := handler.Service.FindByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if existing.GetLanguage() != scripts.ScriptLanguageLua {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot update non-Lua scripts; delete and recreate as Lua"})
		return
	}

	script.Language = scripts.ScriptLanguageLua

	log.WithField("script", script.Name).Info("Updating script")

	if err := handler.Service.Update(id, &script); err == nil {
		c.JSON(http.StatusOK, gin.H{"status": "updated script"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

//DeleteScript ... Updates a script
func (handler *ScriptsHandler) DeleteScript(c *gin.Context) {

	id := c.Param("id")

	log.WithField("script", id).Info("Deleting script")

	if err := handler.Service.Delete(id); err == nil {
		c.JSON(http.StatusOK, gin.H{"status": "deleted script"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}
