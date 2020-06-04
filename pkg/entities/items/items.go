package items

import (
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
)

//ItemType type
type ItemType int

const (
	itemTypeCurrency ItemType = iota + 1
	itemTypeConsumable
	itemTypeArmor
	itemTypeWeapon
	itemTypeCollectible
	itemTypeCraftingMaterial
)

func (it ItemType) String() string {
	return [...]string{"currency", "consumable", "armor", "weapon", "collectible", "crafting_material"}[it]
}

//ItemSlot type
type ItemSlot int

const (
	itemSlotInventory ItemSlot = iota + 1
	itemSlotPurse
	itemSlotHead
	itemSlotChest
	itemSlotLegs
	itemSlotBoots
	itemSlotNeck
	itemSlotRing1
	itemSlotRing2
	itemSlotHands
	itemSlotMainHand
	itemSlotOffHand
)

func (is ItemSlot) String() string {
	return [...]string{"inventory", "purse", "head", "chest", "legs", "boots", "neck", "ring1", "ring2", "hands", "main_hand", "off_hand"}[is]
}

//Item data
type Item struct {
	*entities.Entity `bson:",inline"`

	Name        string `bson:"name,omitempty" json:"name"`
	Description string `bson:"description,omitempty" json:"description"`

	ItemType ItemType `bson:"itemType,omitempty" json:"itemType"`
	ItemSlot ItemSlot `bson:"itemSlot,omitempty" json:"itemSlot"`

	Created    time.Time         `bson:"created,omitempty" json:"created,omitempty"`
	Attributes map[string]string `bson:"attributes,omitempty" json:"attributes,omitempty"`
	Properties map[string]string `bson:"properties,omitempty" json:"properties,omitempty"`
}

//Items type
type Items []*Item
