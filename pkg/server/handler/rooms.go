package handler

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

//RoomsHandler ...
type RoomsHandler struct {
	Service service.RoomsService
}

//GetRooms returns the list of item templates
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

//GetRoomByID returns a single room by ID
func (handler *RoomsHandler) GetRoomByID(c *gin.Context) {
	id := c.Param("id")

	if room, err := handler.Service.FindByID(id); err == nil {
		c.JSON(http.StatusOK, room)
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
	}
}

//GetRoomOfTheDay returns the list of item templates
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

//GetRoomValueHelp returns the list of item templates
func (handler *RoomsHandler) GetRoomValueHelp(c *gin.Context) {

	vh, _ := handler.Service.ValueHelp()

	c.JSON(http.StatusOK, vh)

}

//PostRoom ... creates a new charactersheet
func (handler *RoomsHandler) PostRoom(c *gin.Context) {

	var room rooms.Room
	if err := c.ShouldBindJSON(&room); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("room", room.Name).Info("Creating new room")

	if room, err := handler.Service.Store(&room); err == nil {
		c.JSON(http.StatusOK, room)
	} else {
		c.Error(err)
	}
}

//PutRoom ... Updates a room
func (handler *RoomsHandler) PutRoom(c *gin.Context) {

	id := c.Param("id")
	var room rooms.Room
	if err := c.ShouldBindJSON(&room); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("room", room.Name).Info("Updating room")

	if err := handler.Service.Update(id, &room); err == nil {
		c.JSON(http.StatusOK, gin.H{"status": "updated room"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

//DeleteRoom ... Updates a room
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
