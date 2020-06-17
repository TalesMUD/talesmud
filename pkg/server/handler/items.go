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

//ItemsQuery ...
type ItemsQuery struct {
	Name string         `form:"name"`
	Type items.ItemType `form:"type"`
	Slot items.ItemSlot `form:"slot"`
}

//GetItems returns the list of items
// TODO: add filters
func (csh *ItemsHandler) GetItems(c *gin.Context) {

	var query ItemsQuery

	if c.ShouldBindQuery(&query) == nil {
		log.Println("====== Only Bind By Query String ======")
		log.Println(query.Name)
		log.Println(query.Type)
		log.Println(query.Slot)
	}

	if items, err := csh.Service.Items().FindAll(); err == nil {
		c.JSON(http.StatusOK, items)
	} else {
		c.Error(err)
	}
}

//GetItemSlots ...
func (csh *ItemsHandler) GetItemSlots(c *gin.Context) {
	c.JSON(http.StatusOK, csh.Service.ItemSlots())
}

//GetItemQualities ...
func (csh *ItemsHandler) GetItemQualities(c *gin.Context) {
	c.JSON(http.StatusOK, csh.Service.ItemQualities())
}

//GetItemTypes ...
func (csh *ItemsHandler) GetItemTypes(c *gin.Context) {
	c.JSON(http.StatusOK, csh.Service.ItemTypes())
}

//GetItemSubTypes ...
func (csh *ItemsHandler) GetItemSubTypes(c *gin.Context) {
	c.JSON(http.StatusOK, csh.Service.ItemSubTypes())
}

//GetItemTemplates returns the list of item templates
func (csh *ItemsHandler) GetItemTemplates(c *gin.Context) {

	var query ItemsQuery

	if c.ShouldBindQuery(&query) == nil {
		log.Println("====== Only Bind By Query String ======")
		log.Println(query.Name)
		log.Println(query.Type)
		log.Println(query.Slot)
	}

	if items, err := csh.Service.ItemTemplates().FindAll(); err == nil {
		c.JSON(http.StatusOK, items)
	} else {
		c.Error(err)
	}
}

//GetItemByID returns a item
func (csh *ItemsHandler) GetItemByID(c *gin.Context) {

	id := c.Param("id")

	if item, err := csh.Service.Items().FindByID(id); err == nil {
		c.JSON(http.StatusOK, item)
	} else {
		c.Error(err)
	}
}

//GetItemTemplateByID returns a item
func (csh *ItemsHandler) GetItemTemplateByID(c *gin.Context) {

	id := c.Param("id")

	if item, err := csh.Service.ItemTemplates().FindByID(id); err == nil {
		c.JSON(http.StatusOK, item)
	} else {
		c.Error(err)
	}
}

//DeleteItemByID deletes an item
func (csh *ItemsHandler) DeleteItemByID(c *gin.Context) {

	id := c.Param("id")

	if err := csh.Service.Items().Delete(id); err == nil {
		c.JSON(http.StatusOK, "deleted")
	} else {
		c.Error(err)
	}
}

//DeleteItemTemplateByID deletes an item
func (csh *ItemsHandler) DeleteItemTemplateByID(c *gin.Context) {

	id := c.Param("id")

	if err := csh.Service.ItemTemplates().Delete(id); err == nil {
		c.JSON(http.StatusOK, "deleted")
	} else {
		c.Error(err)
	}
}

//CreateItemFromTemplateID creates a item
func (csh *ItemsHandler) CreateItemFromTemplateID(c *gin.Context) {

	id := c.Param("templateId")

	if item, err := csh.Service.CreateItemFromTemplate(id); err == nil {
		c.JSON(http.StatusOK, item)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	if err := csh.Service.Items().Update(id, &item); err == nil {
		c.JSON(http.StatusOK, gin.H{"status": "updated item"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

//UpdateItemTemplateByID creates a item
func (csh *ItemsHandler) UpdateItemTemplateByID(c *gin.Context) {

	id := c.Param("id")
	var itemTemplate items.ItemTemplate
	if err := c.ShouldBindJSON(&itemTemplate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("itemtemplate", itemTemplate.Name).Info("Updating itemtemplate")

	if err := csh.Service.ItemTemplates().Update(id, &itemTemplate); err == nil {
		c.JSON(http.StatusOK, gin.H{"status": "updated item template"})
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

	if newItem, err := csh.Service.Items().Store(&item); err == nil {
		c.JSON(http.StatusOK, newItem)
	} else {
		c.Error(err)
	}
}

//PostItemTemplate ... creates a new item template
func (csh *ItemsHandler) PostItemTemplate(c *gin.Context) {

	var itemTemplate items.ItemTemplate
	if err := c.ShouldBindJSON(&itemTemplate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("item", itemTemplate.Name).Info("Creating new item template")

	if newItem, err := csh.Service.ItemTemplates().Store(&itemTemplate); err == nil {
		c.JSON(http.StatusOK, newItem)
	} else {
		c.Error(err)
	}
}
