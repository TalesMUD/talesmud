package characters

import (
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/traits"
)

// Attribute data
type Attribute struct {
	Name  string `json:"name"`
	Short string `json:"short"`
	Value int32  `json:"value"`
}

//Attributes ...
type Attributes []Attribute

//NewAttribute ...
func NewAttribute(name, short string, value int32) Attribute {
	return Attribute{
		Name:  name,
		Short: short,
		Value: value,
	}
}

//Character data
type Character struct {
	*entities.Entity   `bson:",inline"`
	traits.BelongsUser `bson:",inline"`
	traits.CurrentRoom `bson:",inline"`

	Name        string `json:"name"`
	Description string `json:"description"`
	Race        Race   `json:"race"`
	Class       Class  `json:"class"`

	CurrentHitPoints int32 `json:"currentHitPoints"`
	MaxHitPoints     int32 `json:"maxHitPoints"`

	XP    int32 `json:"xp"`
	Level int32 `json:"level"`
	Gold  int64 `json:"gold"`

	Created    time.Time  `bson:"created" json:"created,omitempty"`
	Attributes Attributes `bson:"attributes" json:"attributes,omitempty"`

	// complex character fields
	Inventory     items.Inventory                `bson:"inventory" json:"inventory"`
	EquippedItems map[items.ItemSlot]*items.Item `bson:"equippedItems" json:"equippedItems"`

	// Combat state
	InCombat         bool   `bson:"inCombat" json:"inCombat"`
	CombatInstanceID string `bson:"combatInstanceId,omitempty" json:"combatInstanceId,omitempty"`

	// Respawn binding - room where player respawns on death
	BoundRoomID string `bson:"boundRoomId,omitempty" json:"boundRoomId,omitempty"`

	// track alltime stats in character object but dont expose as json by default
	AllTimeStats struct {
		PlayersKilled   int32 `bson:"playersKilled" json:"playersKilled"`
		GoldCollected   int32 `bson:"goldCollected" json:"goldCollected"`
		QuestsCompleted int32 `bson:"questsCompleted" json:"questsCompleted"`
	} `bson:"allTimeStats" json:"_"`
}

// GetAttribute returns the value of an attribute by its short name (STR, DEX, etc.)
func (c *Character) GetAttribute(short string) int32 {
	for _, attr := range c.Attributes {
		if attr.Short == short {
			return attr.Value
		}
	}
	return 10 // Default value if not found
}

// GetAttributeModifier returns the modifier for an attribute (value - 10) / 2
func (c *Character) GetAttributeModifier(short string) int {
	value := c.GetAttribute(short)
	return int(value-10) / 2
}

// GetSTRMod returns the strength modifier
func (c *Character) GetSTRMod() int {
	return c.GetAttributeModifier("STR")
}

// GetDEXMod returns the dexterity modifier
func (c *Character) GetDEXMod() int {
	return c.GetAttributeModifier("DEX")
}

// GetCONMod returns the constitution modifier
func (c *Character) GetCONMod() int {
	return c.GetAttributeModifier("CON")
}

// GetWeaponDamage returns the damage value of the equipped main hand weapon
// Returns 1 for unarmed combat
func (c *Character) GetWeaponDamage() int32 {
	if c.EquippedItems == nil {
		return 1
	}
	weapon, ok := c.EquippedItems[items.ItemSlotMainHand]
	if !ok || weapon == nil {
		return 1 // Unarmed
	}
	if damage, ok := weapon.Attributes["damage"]; ok {
		if dmg, ok := damage.(float64); ok {
			return int32(dmg)
		}
		if dmg, ok := damage.(int); ok {
			return int32(dmg)
		}
		if dmg, ok := damage.(int32); ok {
			return dmg
		}
	}
	return 1
}

// GetArmorDefense returns the total defense from all equipped armor
func (c *Character) GetArmorDefense() int32 {
	if c.EquippedItems == nil {
		return 0
	}
	var totalDefense int32 = 0
	for _, item := range c.EquippedItems {
		if item == nil {
			continue
		}
		if defense, ok := item.Attributes["defense"]; ok {
			if def, ok := defense.(float64); ok {
				totalDefense += int32(def)
			} else if def, ok := defense.(int); ok {
				totalDefense += int32(def)
			} else if def, ok := defense.(int32); ok {
				totalDefense += def
			}
		}
		if armor, ok := item.Attributes["armor"]; ok {
			if arm, ok := armor.(float64); ok {
				totalDefense += int32(arm)
			} else if arm, ok := armor.(int); ok {
				totalDefense += int32(arm)
			} else if arm, ok := armor.(int32); ok {
				totalDefense += arm
			}
		}
	}
	return totalDefense
}
