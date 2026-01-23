package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

// ItemsHandler ...
type ItemsHandler struct {
	Service service.ItemsService
}

// GetItems returns the list of items
// Use ?isTemplate=true to get only templates, ?isTemplate=false for instances only
func (h *ItemsHandler) GetItems(c *gin.Context) {
	var query repository.ItemsQuery

	if c.ShouldBindQuery(&query) == nil {
		log.Println("====== Only Bind By Query String ======")
		log.Println(query.Name)
		log.Println(query.Type)
		log.Println(query.Slot)
	}

	// Check for isTemplate filter
	isTemplateStr := c.Query("isTemplate")

	var result []*items.Item
	var err error

	switch isTemplateStr {
	case "true":
		result, err = h.Service.FindAllTemplates(query)
	case "false":
		result, err = h.Service.FindAllInstances(query)
	default:
		result, err = h.Service.FindAll(query)
	}

	if err == nil {
		c.JSON(http.StatusOK, result)
	} else {
		c.Error(err)
	}
}

// GetItemSlots ...
func (h *ItemsHandler) GetItemSlots(c *gin.Context) {
	c.JSON(http.StatusOK, h.Service.ItemSlots())
}

// GetItemQualities ...
func (h *ItemsHandler) GetItemQualities(c *gin.Context) {
	c.JSON(http.StatusOK, h.Service.ItemQualities())
}

// GetItemTypes ...
func (h *ItemsHandler) GetItemTypes(c *gin.Context) {
	c.JSON(http.StatusOK, h.Service.ItemTypes())
}

// GetItemSubTypes ...
func (h *ItemsHandler) GetItemSubTypes(c *gin.Context) {
	c.JSON(http.StatusOK, h.Service.ItemSubTypes())
}

// GetItemByID returns a item (template or instance)
func (h *ItemsHandler) GetItemByID(c *gin.Context) {
	id := c.Param("id")

	if item, err := h.Service.FindByID(id); err == nil {
		c.JSON(http.StatusOK, item)
	} else {
		c.Error(err)
	}
}

// DeleteItemByID deletes an item (template or instance)
func (h *ItemsHandler) DeleteItemByID(c *gin.Context) {
	id := c.Param("id")

	if err := h.Service.Delete(id); err == nil {
		c.JSON(http.StatusOK, "deleted")
	} else {
		c.Error(err)
	}
}

// CreateInstanceFromTemplate creates an item instance from a template
func (h *ItemsHandler) CreateInstanceFromTemplate(c *gin.Context) {
	templateID := c.Param("templateId")

	instance, err := h.Service.CreateInstanceFromTemplate(templateID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, instance)
}

// UpdateItemByID updates an item (template or instance)
func (h *ItemsHandler) UpdateItemByID(c *gin.Context) {
	id := c.Param("id")
	var item items.Item
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("item", item.Name).Info("Updating item")

	if err := h.Service.Update(id, &item); err == nil {
		c.JSON(http.StatusOK, gin.H{"status": "updated item"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

// PostItem creates a new item (template or instance based on isTemplate field)
func (h *ItemsHandler) PostItem(c *gin.Context) {
	var item items.Item
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.WithField("item", item.Name).WithField("isTemplate", item.IsTemplate).Info("Creating new item")

	if newItem, err := h.Service.Store(&item); err == nil {
		c.JSON(http.StatusOK, newItem)
	} else {
		c.Error(err)
	}
}
