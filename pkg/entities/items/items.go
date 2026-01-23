package items

import (
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/traits"
)

//ItemType type
type ItemType string

//ItemTypes ...
type ItemTypes []ItemType

const (
	ItemTypeCurrency    ItemType = "currency"
	ItemTypeConsumable           = "consumable"
	ItemTypeArmor                = "armor"
	ItemTypeWeapon               = "weapon"
	ItemTypeCollectible          = "collectible"
	ItemTypeQuest                = "quest"

	ItemTypeCraftingMaterial = "crafting_material"
)

//ItemSubType type
type ItemSubType string

//ItemSubTypes ...
type ItemSubTypes []ItemSubType

const (
	// weapons
	ItemSubTypeSword        ItemSubType = "sword"
	ItemSubTypeTwoHandSword ItemSubType = "twohandsword"
	ItemSubTypeAxe                      = "axe"
	ItemSubTypeSpear                    = "spear"

	// shields
	ItemSubTypeShield = "shield"
)

//ItemSlot type
type ItemSlot string

//ItemSlots type
type ItemSlots []ItemSlot

const (
	ItemSlotInventory ItemSlot = "inventory"
	ItemSlotContainer          = "container"
	ItemSlotPurse              = "purse"
	ItemSlotHead               = "head"
	ItemSlotChest              = "chest"
	ItemSlotLegs               = "legs"
	ItemSlotBoots              = "boots"
	ItemSlotNeck               = "neck"
	ItemSlotRing1              = "ring1"
	ItemSlotRing2              = "ring2"
	ItemSlotHands              = "hands"
	ItemSlotMainHand           = "main_hand"
	ItemSlotOffHand            = "off_hand"
)

//ItemQuality ...
type ItemQuality string

//ItemQualities type
type ItemQualities []ItemQuality

const (
	ItemQualityNormal    ItemQuality = "normal"
	ItemQualityMagic                 = "magic"
	ItemQualityRare                  = "rare"
	ItemQualityLegendary             = "legendary"
	ItemQualityMythic                = "mythic"
)

//Item data
type Item struct {
	*entities.Entity `bson:",inline"`
	traits.LookAt    `bson:",inline"` // "detail"

	// Template System (following NPC pattern)
	// IsTemplate indicates this Item is a blueprint for creating instances
	IsTemplate bool `bson:"isTemplate" json:"isTemplate"`
	// TemplateID references the source template for created instances
	TemplateID string `bson:"templateId,omitempty" json:"templateId,omitempty"`
	// InstanceSuffix is a unique suffix for created instances (e.g., "abc123")
	InstanceSuffix string `bson:"instanceSuffix,omitempty" json:"instanceSuffix,omitempty"`

	Name        string `bson:"name,omitempty" json:"name"`
	Description string `bson:"description,omitempty" json:"description"`

	Type    ItemType    `bson:"type,omitempty" json:"type"`
	SubType ItemSubType `bson:"subType,omitempty" json:"subType"`
	Slot    ItemSlot    `bson:"slot,omitempty" json:"slot"`
	Quality ItemQuality `bson:"quality,omitempty" json:"quality"`
	Level   int32       `bson:"level,omitempty" json:"level,omitempty"`

	// custom item properties
	Properties map[string]interface{} `bson:"properties,omitempty" json:"properties,omitempty"`
	// "stats"
	Attributes map[string]interface{} `bson:"attributes,omitempty" json:"attributes,omitempty"`

	// container specifics
	Closed   bool   `bson:"closed,omitempty" json:"closed,omitempty"`
	Locked   bool   `bson:"locked,omitempty" json:"locked,omitempty"`
	LockedBy string `bson:"lockedBy,omitempty" json:"lockedBy,omitempty"`
	Items    Items  `bson:"items,omitempty" json:"items,omitempty"`
	MaxItems int32  `bson:"maxItems,omitempty" json:"maxItems,omitempty"`

	// misc
	NoPickup bool `bson:"noPickup,omitempty" json:"noPickup,omitempty"`

	// stacking and economy
	Stackable bool  `bson:"stackable,omitempty" json:"stackable,omitempty"`
	Quantity  int32 `bson:"quantity,omitempty" json:"quantity,omitempty"`
	MaxStack  int32 `bson:"maxStack,omitempty" json:"maxStack,omitempty"`
	BasePrice int64 `bson:"basePrice,omitempty" json:"basePrice,omitempty"`

	// scripts

	// metainfo
	Tags      []string  `bson:"tags,omitempty" json:"tags"`
	Created   time.Time `bson:"created,omitempty" json:"created,omitempty"`
	CreatedBy string    `bson:"createdBy,omitempty" json:"createdBy,omitempty"`

	// additional non game critical meta information to enhance player experience on client
	Meta *struct {
		Img string `bson:"img,omitempty" json:"img,omitempty"`
	} `bson:"meta,omitempty" meta:"coords,omitempty"`
}

//Items type
type Items []*Item

// IsInstance returns true if this Item was created from a template
func (item *Item) IsInstance() bool {
	return item.TemplateID != "" && item.InstanceSuffix != ""
}

// GetDisplayName returns the name shown to players
func (item *Item) GetDisplayName() string {
	return item.Name
}

// GetTargetName returns the unique name for targeting commands
func (item *Item) GetTargetName() string {
	if item.InstanceSuffix != "" {
		return item.Name + "-" + item.InstanceSuffix
	}
	return item.Name
}

// CanEquipToSlot checks if this item can be equipped to the given slot
func (item *Item) CanEquipToSlot(slot ItemSlot) bool {
	// Item must have an equipment slot defined
	if item.Slot == "" || item.Slot == ItemSlotInventory || item.Slot == ItemSlotContainer || item.Slot == ItemSlotPurse {
		return false
	}

	// Two-handed weapons can go in main hand or off hand (they occupy both)
	if item.SubType == ItemSubTypeTwoHandSword {
		return slot == ItemSlotMainHand || slot == ItemSlotOffHand
	}

	// Direct slot match
	return item.Slot == slot
}

// IsTwoHanded returns true if this item requires both hand slots
func (item *Item) IsTwoHanded() bool {
	return item.SubType == ItemSubTypeTwoHandSword
}

// IsEquippable returns true if this item can be equipped
func (item *Item) IsEquippable() bool {
	return item.Slot != "" && item.Slot != ItemSlotInventory && item.Slot != ItemSlotContainer && item.Slot != ItemSlotPurse
}
