package items

import (
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/traits"
)

//ItemType type
type ItemType string

const (
	ItemTypeCurrency         ItemType = "currency"
	ItemTypeConsumable                = "consumable"
	ItemTypeArmor                     = "armor"
	ItemTypeWeapon                    = "weapon"
	ItemTypeCollectible               = "collectible"
	ItemTypeCraftingMaterial          = "crafting_material"
)

//ItemSlot type
type ItemSlot string

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

//Item data
type Item struct {
	*entities.Entity `bson:",inline"`
	traits.LookAt    `bson:",inline"` // "detail"

	Name        string `bson:"name,omitempty" json:"name"`
	Description string `bson:"description,omitempty" json:"description"`

	Type ItemType `bson:"type,omitempty" json:"type"`
	Slot ItemSlot `bson:"slot,omitempty" json:"slot"`

	// custom item properties
	Properties map[string]string `bson:"properties,omitempty" json:"properties,omitempty"`
	// "stats"
	Attributes map[string]string `bson:"attributes,omitempty" json:"attributes,omitempty"`

	// container specifics
	Closed   bool   `bson:"closed,omitempty" json:"closed,omitempty"`
	Locked   bool   `bson:"locked,omitempty" json:"locked,omitempty"`
	LockedBy string `bson:"lockedBy,omitempty" json:"lockedBy,omitempty"`
	Items    Items  `bson:"items,omitempty" json:"items,omitempty"`
	MaxItems int32  `bson:"maxItems,omitempty" json:"maxItems,omitempty"`

	// misc
	NoPickup bool `bson:"noPickup,omitempty" json:"noPickup,omitempty"`

	// scripts

	// metainfo
	Tags    []string  `bson:"tags,omitempty" json:"tags"`
	Created time.Time `bson:"created,omitempty" json:"created,omitempty"`
}

//Items type
type Items []*Item
