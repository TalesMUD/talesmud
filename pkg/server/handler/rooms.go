package handler

import (
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/contenthealth"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/inspect"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

// RoomsHandler ...
type RoomsHandler struct {
	Service service.RoomsService
	Facade  service.Facade
}

// GetRooms returns the list of item templates
func (handler *RoomsHandler) GetRooms(c *gin.Context) {

	var query repository.RoomsQuery

	if c.ShouldBindQuery(&query) == nil {
		// WITH QUERY
	}

	if rooms, err := handler.Service.FindAllWithQuery(query); err == nil {
		c.JSON(http.StatusOK, rooms)
	} else {
		c.Error(err)
	}
}

// InspectRoom returns the static inspector for one content room.
// An instance-copy id resolves to its template. Live rows stay on the live endpoints.
func (handler *RoomsHandler) InspectRoom(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room id is required"})
		return
	}
	if handler.Facade == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "room inspect is unavailable"})
		return
	}
	world, err := contenthealth.WorldFromFacade(handler.Facade)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	target := roomInspectTarget(world, id)
	view, ok := inspect.Room(world.Graph(), target)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}
	if target != id {
		view.RequestedID = id
	}
	nameRoomItems(&view, world.Items)
	c.JSON(http.StatusOK, view)
}

func roomInspectTarget(world contenthealth.World, id string) string {
	for _, room := range world.Rooms {
		if room != nil && room.Entity != nil && room.ID == id {
			if i := strings.LastIndex(id, "~"); i > 0 {
				return id[:i]
			}
			return id
		}
	}
	if i := strings.LastIndex(id, "~"); i > 0 {
		return id[:i]
	}
	return id
}

func nameRoomItems(view *inspect.RoomView, list []*items.Item) {
	names := map[string]string{}
	for _, item := range list {
		if item == nil || item.Entity == nil || item.ID == "" {
			continue
		}
		names[item.ID] = item.Name
	}
	for i := range view.Items {
		name, ok := names[view.Items[i].ID]
		if !ok {
			continue
		}
		if name != "" && name != view.Items[i].ID {
			view.Items[i].Name = name
		}
		view.Items[i].Missing = false
	}
}

// GetRoomByID returns a single room by ID
func (handler *RoomsHandler) GetRoomByID(c *gin.Context) {
	id := c.Param("id")

	if room, err := handler.Service.FindByID(id); err == nil {
		c.JSON(http.StatusOK, room)
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
	}
}

// GetRoomOfTheDay returns the list of item templates
func (handler *RoomsHandler) GetRoomOfTheDay(c *gin.Context) {

	if rooms, err := handler.Service.FindAll(); err == nil {
		dayOfYear := time.Now().YearDay()
		rand.Seed(int64(dayOfYear))
		randomPick := rand.Int() % len(rooms)
		room := rooms[randomPick]

		c.JSON(http.StatusOK, room)
	} else {
		c.Error(err)
	}
}

// GetRoomValueHelp returns the list of item templates
func (handler *RoomsHandler) GetRoomValueHelp(c *gin.Context) {

	vh, _ := handler.Service.ValueHelp()

	c.JSON(http.StatusOK, vh)

}

// PostRoom ... creates a new charactersheet
func (handler *RoomsHandler) PostRoom(c *gin.Context) {

	var room rooms.Room
	if err := c.ShouldBindJSON(&room); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("room", room.Name).Info("Creating new room")

	if rejectInvalidRoom(c, handler.Facade, &room) {
		return
	}

	if room, err := handler.Service.Store(&room); err == nil {
		c.JSON(http.StatusOK, room)
	} else {
		c.Error(err)
	}
}

// PutRoom ... Updates a room
func (handler *RoomsHandler) PutRoom(c *gin.Context) {

	id := c.Param("id")
	var room rooms.Room
	if err := c.ShouldBindJSON(&room); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("room", room.Name).Info("Updating room")

	if rejectInvalidRoom(c, handler.Facade, &room) {
		return
	}

	if err := handler.Service.Update(id, &room); err == nil {
		c.JSON(http.StatusOK, gin.H{"status": "updated room"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

// DeleteRoom ... Updates a room
func (handler *RoomsHandler) DeleteRoom(c *gin.Context) {

	id := c.Param("id")

	log.WithField("room", id).Info("Deleting room")

	if err := handler.Service.Delete(id); err == nil {
		c.JSON(http.StatusOK, gin.H{"status": "deleted room"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

/*
//GetCharacterByID returns a single charactersheet
func (handler *RoomsHandler) GetCharacterByID(c *gin.Context) {

	id := c.Query("id")

	if character, err := handler.Service.GetCharacterSheetByID(id); err == nil {
		c.JSON(http.StatusOK, character)
	} else {
		c.Error(err)
	}
}

//DeleteCharacterByID returns a single charactersheet
func (handler *RoomsHandler) DeleteCharacterByID(c *gin.Context) {

	id := c.Query("id")

	if err := handler.Service.DeleteCharacterSheetByID(id); err == nil {
		c.JSON(http.StatusOK, "deleted")
	} else {
		c.Error(err)
	}
}

//UpdateCharacterByID creates a new charactersheet
func (handler *RoomsHandler) UpdateCharacterByID(c *gin.Context) {

	id := c.Query("id")
	var character e.CharacterSheet
	if err := c.ShouldBindJSON(&character); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("character", character.Name).Info("Updating character")

	if err := handler.Service.UpdateCharacterSheetByID(id, &character); err == nil {
		c.JSON(http.StatusOK, gin.H{"status": "updated character"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

//PostCharacter ... creates a new charactersheet
func (handler *RoomsHandler) PostCharacter(c *gin.Context) {

	var character e.CharacterSheet
	if err := c.ShouldBindJSON(&character); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("character", character.Name).Info("Creating new character")

	if newCharacter, err := handler.Service.CreateCharacterSheet(&character); err == nil {
		c.JSON(http.StatusOK, newCharacter)
	} else {
		c.Error(err)
	}
}
*/
