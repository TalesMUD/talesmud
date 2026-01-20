package characters

import (
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/items"
)

// StartingItem defines an item template to be created and optionally equipped when a character is created from a template.
// If Slot is "inventory", the item is just added to the inventory.
type StartingItem struct {
	Slot             items.ItemSlot `bson:"slot,omitempty" json:"slot"`
	ItemTemplateID   string         `bson:"itemTemplateId,omitempty" json:"itemTemplateId"`
	ItemTemplateName string         `bson:"itemTemplateName,omitempty" json:"itemTemplateName,omitempty"` // convenience for UI
}

// CharacterTemplate is an editable class/archetype template used for character creation.
type CharacterTemplate struct {
	*entities.Entity `bson:",inline"`

	// Display / lore
	Name        string `bson:"name,omitempty" json:"name"`
	Description string `bson:"description,omitempty" json:"description"`
	Backstory   string `bson:"backstory,omitempty" json:"backstory,omitempty"`
	OriginArea  string `bson:"originArea,omitempty" json:"originArea,omitempty"`
	Archetype   string `bson:"archetype,omitempty" json:"archetype,omitempty"` // e.g. warrior, rogue, mage

	// Gameplay base
	Race Race  `bson:"race,omitempty" json:"race"`
	Class Class `bson:"class,omitempty" json:"class"`

	Level int32 `bson:"level,omitempty" json:"level"`

	CurrentHitPoints int32 `bson:"currentHitPoints,omitempty" json:"currentHitPoints"`
	MaxHitPoints     int32 `bson:"maxHitPoints,omitempty" json:"maxHitPoints"`

	Attributes Attributes `bson:"attributes,omitempty" json:"attributes,omitempty"`

	StartingItems []StartingItem `bson:"startingItems,omitempty" json:"startingItems,omitempty"`

	// Meta
	Source  string    `bson:"source,omitempty" json:"source,omitempty"` // "db" | "system"
	Created time.Time `bson:"created,omitempty" json:"created,omitempty"`
	Updated time.Time `bson:"updated,omitempty" json:"updated,omitempty"`
}

