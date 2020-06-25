package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/scripts"
	"github.com/talesmud/talesmud/pkg/service"
)

//ScriptsHandler ...
type ScriptsHandler struct {
	Service service.ScriptsService
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

	var script scripts.Script
	if err := c.ShouldBindJSON(&script); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	// content ...
	/*var script scripts.Script
	if err := c.ShouldBindJSON(&script); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}*/

	var data string
	c.Bind(&data)
	var js
	if c.BindJSON(&json) == nil {
	if script, err := handler.Service.FindByID(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {

		runner := scripts.ScriptRunner{}
		result := runner.Run(*script, data)

		c.JSON(http.StatusOK, gin.H{
			"status": "Executed script",	
			"result": result})

	}

}

//PutScript ... Updates a script
func (handler *ScriptsHandler) PutScript(c *gin.Context) {

	id := c.Param("id")
	var script scripts.Script
	if err := c.ShouldBindJSON(&script); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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
