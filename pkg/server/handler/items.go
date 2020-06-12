package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/service"
)

//ItemsHandler ...
type ItemsHandler struct {
	Service service.ItemsService
}

//GetItems returns the list of item templates
func (csh *ItemsHandler) GetItems(c *gin.Context) {

	if items, err := csh.Service.FindAll(); err == nil {
		c.JSON(http.StatusOK, items)
	} else {
		c.Error(err)
	}
}

//GetItemByID returns a item
func (csh *ItemsHandler) GetItemByID(c *gin.Context) {

	id := c.Param("id")

	if item, err := csh.Service.FindByID(id); err == nil {
		c.JSON(http.StatusOK, item)
	} else {
		c.Error(err)
	}
}

//DeleteItemByID returns a item
func (csh *ItemsHandler) DeleteItemByID(c *gin.Context) {

	id := c.Param("id")

	if err := csh.Service.Delete(id); err == nil {
		c.JSON(http.StatusOK, "deleted")
	} else {
		c.Error(err)
	}
}

//UpdateItemByID creates a item
func (csh *ItemsHandler) UpdateItemByID(c *gin.Context) {

	id := c.Param("id")
	var item items.Item
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("item", item.Name).Info("Updating item")

	if err := csh.Service.Update(id, &item); err == nil {
		c.JSON(http.StatusOK, gin.H{"status": "updated item"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

//PostItem ... creates a new item
func (csh *ItemsHandler) PostItem(c *gin.Context) {

	var item items.Item
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("item", item.Name).Info("Creating new item")

	if newItem, err := csh.Service.Store(&item); err == nil {
		c.JSON(http.StatusOK, newItem)
	} else {
		c.Error(err)
	}
}
