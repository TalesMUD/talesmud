package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/items"
	r "github.com/talesmud/talesmud/pkg/repository"
)

//--- Interface Definitions

// ItemsService delivers logical functions on top of the items repository
type ItemsService interface {
	r.ItemsRepository

	// Template-specific methods
	FindAllTemplates(query r.ItemsQuery) ([]*items.Item, error)
	CreateInstanceFromTemplate(templateID string) (*items.Item, error)

	// Metadata helpers
	ItemSlots() items.ItemSlots
	ItemQualities() items.ItemQualities
	ItemTypes() items.ItemTypes
	ItemSubTypes() items.ItemSubTypes
}

//--- Implementations

type itemsService struct {
	r.ItemsRepository
}

// NewItemsService creates a new item service
func NewItemsService(itemsRepo r.ItemsRepository) ItemsService {
	return &itemsService{
		ItemsRepository: itemsRepo,
	}
}

func (srv *itemsService) ItemSlots() items.ItemSlots {
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

func (srv *itemsService) ItemQualities() items.ItemQualities {
	return items.ItemQualities{
		items.ItemQualityNormal,
		items.ItemQualityMagic,
		items.ItemQualityRare,
		items.ItemQualityLegendary,
		items.ItemQualityMythic,
	}
}

func (srv *itemsService) ItemTypes() items.ItemTypes {
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

func (srv *itemsService) ItemSubTypes() items.ItemSubTypes {
	return items.ItemSubTypes{
		items.ItemSubTypeSword,
		items.ItemSubTypeTwoHandSword,
		items.ItemSubTypeAxe,
		items.ItemSubTypeSpear,
		items.ItemSubTypeShield,
	}
}

// CreateInstanceFromTemplate creates a new item instance from a template
func (srv *itemsService) CreateInstanceFromTemplate(templateID string) (*items.Item, error) {
	template, err := srv.FindByID(templateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}
	if template == nil {
		return nil, fmt.Errorf("template %s not found", templateID)
	}
	if !template.IsTemplate {
		return nil, fmt.Errorf("item %s is not a template", templateID)
	}

	// Generate unique suffix
	suffix := uuid.New().String()[:8]

	// Create instance (deep copy template fields)
	instance := &items.Item{
		Entity:      entities.NewEntity(),
		Name:        template.Name,
		Description: template.Description,
		Type:        template.Type,
		SubType:     template.SubType,
		Slot:        template.Slot,
		Quality:     template.Quality,
		Level:       template.Level,
		Properties:  copyMap(template.Properties),
		Attributes:  copyMap(template.Attributes),
		Tags:        append([]string{}, template.Tags...),
		NoPickup:    template.NoPickup,

		// Template reference
		IsTemplate:     false,
		TemplateID:     templateID,
		InstanceSuffix: suffix,

		Created: time.Now(),
	}

	// Copy container fields if applicable
	instance.Closed = template.Closed
	instance.Locked = template.Locked
	instance.LockedBy = template.LockedBy
	instance.MaxItems = template.MaxItems

	// Copy consumable/stacking fields
	instance.Consumable = template.Consumable
	instance.OnUseScriptID = template.OnUseScriptID
	instance.Stackable = template.Stackable
	instance.Quantity = template.Quantity
	instance.MaxStack = template.MaxStack
	instance.BasePrice = template.BasePrice

	// Copy LookAt trait
	instance.LookAt = template.LookAt

	// Copy meta if present
	if template.Meta != nil {
		instance.Meta = &struct {
			Img string `bson:"img,omitempty" json:"img,omitempty"`
		}{Img: template.Meta.Img}
	}

	// Save the instance to the database
	savedInstance, err := srv.Store(instance)
	if err != nil {
		return nil, fmt.Errorf("failed to save item instance: %w", err)
	}

	return savedInstance, nil
}

// copyMap creates a shallow copy of a map
func copyMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
