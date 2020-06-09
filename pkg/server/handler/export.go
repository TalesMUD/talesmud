package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	e "github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/service"
)

//ExportHandler ...
type ExportHandler struct {
	RoomsService      service.RoomsService
	CharactersService service.CharactersService
	UserService       service.UsersService
}

type exportStructure struct {
	Characters []*characters.Character `json:"characters"`
	Rooms      []*rooms.Room           `json:"rooms"`
	Users      []*e.User               `json:"users"`
}

//Export Exports all data structures as JSON
func (handler *ExportHandler) Export(c *gin.Context) {

	d := exportStructure{}

	d.Rooms, _ = handler.RoomsService.FindAll()
	d.Characters, _ = handler.CharactersService.FindAll()
	d.Users, _ = handler.UserService.FindAll()

	//c.JSON(http.StatusOK, d)
	c.YAML(http.StatusOK, d)
}

//Import Imports all data structures
func (handler *ExportHandler) Import(c *gin.Context) {

	var data exportStructure
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

	c.JSON(http.StatusOK, gin.H{"status": "Import successful"})
}
