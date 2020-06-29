package service

import (
	"github.com/sirupsen/logrus"
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

	CreateItemFromTemplate(templateID string) (*items.Item, error)

	ItemSlots() items.ItemSlots
	ItemQualities() items.ItemQualities
	ItemTypes() items.ItemTypes
	ItemSubTypes() items.ItemSubTypes
}

//--- Implementations

type itemsService struct {
	r.ItemsRepository
	r.ItemTemplatesRepository
	ScriptsService
	scripts.ScriptRunner
}

//NewItemsService creates a nwe item service
func NewItemsService(itemsRepo r.ItemsRepository, itemsTemplateRepo r.ItemTemplatesRepository, scriptService ScriptsService, runner scripts.ScriptRunner) ItemsService {
	return &itemsService{
		itemsRepo,
		itemsTemplateRepo,
		scriptService,
		runner,
	}
}

func (itemsService *itemsService) ItemSlots() items.ItemSlots {
	return items.ItemSlots{
		items.ItemSlotInventory,
		items.ItemSlotContainer,
		items.ItemSlotPurse,
		items.ItemSlotHead,
		items.ItemSlotChest,
		items.ItemSlotLegs,
		items.ItemSlotBoots,
		items.ItemSlotNeck,
		items.ItemSlotRing1,
		items.ItemSlotRing2,
		items.ItemSlotHands,
		items.ItemSlotMainHand,
		items.ItemSlotOffHand,
	}
}

func (itemsService *itemsService) ItemQualities() items.ItemQualities {
	return items.ItemQualities{
		items.ItemQualityNormal,
		items.ItemQualityMagic,
		items.ItemQualityRare,
		items.ItemQualityLegendary,
		items.ItemQualityMythic,
	}
}

func (itemsService *itemsService) ItemTypes() items.ItemTypes {
	return items.ItemTypes{
		items.ItemTypeCurrency,
		items.ItemTypeConsumable,
		items.ItemTypeArmor,
		items.ItemTypeWeapon,
		items.ItemTypeCollectible,
		items.ItemTypeQuest,
		items.ItemTypeCraftingMaterial,
	}
}

func (itemsService *itemsService) ItemSubTypes() items.ItemSubTypes {
	return items.ItemSubTypes{
		items.ItemSubTypeSword,
		items.ItemSubTypeTwoHandSword,
		items.ItemSubTypeAxe,
		items.ItemSubTypeSpear,
		items.ItemSubTypeShield,
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
		if template.Script != nil {
			if script, err := itemsService.ScriptsService.FindByID(*template.Script); err == nil {

				logrus.WithField("item", item.Name).WithField("script", script.Name).Info("Executing script on created item")
				itemsService.ScriptRunner.Run(*script, item)
			}
		}
		return item, nil
	}

	//	return nil, fmt.Errorf("Could not create item from templateID %v", templateID)
}
