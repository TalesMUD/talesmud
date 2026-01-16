package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/exporter"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

// ExportHandler ...
type ExportHandler struct {
	RoomsService      service.RoomsService
	CharactersService service.CharactersService
	UserService       service.UsersService
	ItemsService      service.ItemsService
	ScriptService     service.ScriptsService
	NPCsService       service.NPCsService
	DialogsService    service.DialogsService
	PartiesService    service.PartiesService
}

// Export Exports all data structures as JSON
func (handler *ExportHandler) Export(c *gin.Context) {

	d := exporter.Data{}

	d.Rooms, _ = handler.RoomsService.FindAll()
	d.Characters, _ = handler.CharactersService.FindAll()
	d.Users, _ = handler.UserService.FindAll()
	d.ItemTemplates, _ = handler.ItemsService.ItemTemplates().FindAll(repository.ItemsQuery{})
	d.Items, _ = handler.ItemsService.Items().FindAll(repository.ItemsQuery{})
	d.Scripts, _ = handler.ScriptService.FindAll()
	d.NPCs, _ = handler.NPCsService.FindAll()
	d.Dialogs, _ = handler.DialogsService.FindAll()
	d.Parties, _ = handler.PartiesService.FindAll()

	//c.JSON(http.StatusOK, d)
	c.IndentedJSON(http.StatusOK, d)
}

// Import Imports all data structures
func (handler *ExportHandler) Import(c *gin.Context) {

	// drop all collections before importing
	handler.RoomsService.Drop()
	handler.CharactersService.Drop()
	handler.UserService.Drop()
	handler.ItemsService.Items().Drop()
	handler.ItemsService.ItemTemplates().Drop()
	handler.ScriptService.Drop()
	handler.NPCsService.Drop()
	handler.DialogsService.Drop()

	var data exporter.Data
	//if err := c.ShouldBindYAML(&data); err != nil {
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, room := range data.Rooms {
		handler.RoomsService.Import(room)
	}

	for _, character := range data.Characters {
		handler.CharactersService.Import(character)
	}

	for _, user := range data.Users {
		handler.UserService.Import(user)
	}

	for _, item := range data.Items {
		handler.ItemsService.Items().Import(item)
	}
	for _, itemTemplate := range data.ItemTemplates {
		handler.ItemsService.ItemTemplates().Import(itemTemplate)
	}

	for _, script := range data.Scripts {
		handler.ScriptService.Import(script)
	}
	for _, npc := range data.NPCs {
		handler.NPCsService.Import(npc)
	}
	for _, dialog := range data.Dialogs {
		handler.DialogsService.Import(dialog)
	}
	for _, party := range data.Parties {
		handler.PartiesService.Store(party)
	}

	c.JSON(http.StatusOK, gin.H{"status": "Import successful"})
}
