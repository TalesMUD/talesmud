package service

import (
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/items"
	r "github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/scripts"
)

//--- Interface Definitions

//ItemsService delives logical functions on top of the charactersheets Repo
type ItemsService interface {
	Items() r.ItemsRepository
	ItemTemplates() r.ItemTemplatesRepository

	CreateItemFromTemplate(templateId string) (*items.Item, error)
}

//--- Implementations

type itemsService struct {
	r.ItemsRepository
	r.ItemTemplatesRepository
	scripts.ScriptRunner
}

//NewItemsService creates a nwe item service
func NewItemsService(itemsRepo r.ItemsRepository, itemsTemplateRepo r.ItemTemplatesRepository) ItemsService {
	return &itemsService{
		itemsRepo,
		itemsTemplateRepo,
		scripts.ScriptRunner{},
	}
}

//Items...
func (itemsService *itemsService) Items() r.ItemsRepository {
	return itemsService.ItemsRepository
}

//itemTemplates...
func (itemsService *itemsService) ItemTemplates() r.ItemTemplatesRepository {
	return itemsService.ItemTemplatesRepository
}

func createItemFromTemplate(itemTemplate *items.ItemTemplate) *items.Item {

	item := itemTemplate.Item
	item.Entity = entities.NewEntity()

	return &item
}

//CreateItemFromTemplate ...
func (itemsService *itemsService) CreateItemFromTemplate(templateID string) (*items.Item, error) {

	// get item template
	if template, err := itemsService.ItemTemplates().FindByID(templateID); err != nil {
		return nil, err
	} else {

		item := createItemFromTemplate(template)

		// run script after item creation
		if template.OnAfterCreate != nil {
			itemsService.ScriptRunner.Run(*template.OnAfterCreate, item)
		}

		return item, nil
	}

	//	return nil, fmt.Errorf("Could not create item from templateID %v", templateID)
}
