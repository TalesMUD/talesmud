package characters

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/traits"
)

// toInt32 converts an interface{} value to int32, handling float64, int, int32,
// and string types. Returns (value, true) on success, (0, false) on failure.
// This is needed because the Creator UI stores attribute values as strings
// (HTML inputs always produce strings), while Go code and JSON deserialization
// produce numeric types.
func toInt32(val interface{}) (int32, bool) {
	switch v := val.(type) {
	case float64:
		return int32(v), true
	case int:
		return int32(v), true
	case int32:
		return v, true
	case int64:
		return int32(v), true
	case string:
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			return int32(n), true
		}
		if n, err := strconv.Atoi(v); err == nil {
			return int32(n), true
		}
	case fmt.Stringer:
		if n, err := strconv.Atoi(v.String()); err == nil {
			return int32(n), true
		}
	}
	return 0, false
}

// NormalizeAttributeShorts ensures all attribute Short fields are uppercase.
// Call this after loading a character from the database to handle legacy data.
func (c *Character) NormalizeAttributeShorts() {
	for i := range c.Attributes {
		c.Attributes[i].Short = strings.ToUpper(c.Attributes[i].Short)
	}
}

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

	// Mana (for caster classes)
	CurrentMana int32 `bson:"currentMana" json:"currentMana"`
	MaxMana     int32 `bson:"maxMana" json:"maxMana"`

	XP    int32 `json:"xp"`
	Level int32 `json:"level"`
	Gold  int64 `json:"gold"`

	Created    time.Time  `bson:"created" json:"created,omitempty"`
	Attributes Attributes `bson:"attributes" json:"attributes,omitempty"`

	// Distributable attribute points (earned on level-up, spent by player)
	UnspentAttributePoints int32            `bson:"unspentAttributePoints" json:"unspentAttributePoints"`
	SpentAttributePoints   map[string]int32 `bson:"spentAttributePoints,omitempty" json:"spentAttributePoints,omitempty"`

	// complex character fields
	Inventory     items.Inventory                `bson:"inventory" json:"inventory"`
	EquippedItems map[items.ItemSlot]*items.Item `bson:"equippedItems" json:"equippedItems"`

	// Equipped skill/spell IDs (max 4, locked during combat)
	EquippedSkills []string `bson:"equippedSkills,omitempty" json:"equippedSkills,omitempty"`

	// Combat state
	InCombat         bool   `bson:"inCombat" json:"inCombat"`
	CombatInstanceID string `bson:"combatInstanceId,omitempty" json:"combatInstanceId,omitempty"`

	// Respawn binding - room where player respawns on death
	BoundRoomID string `bson:"boundRoomId,omitempty" json:"boundRoomId,omitempty"`

	// Game flags for scripting (puzzles, quests, state tracking)
	Flags map[string]interface{} `bson:"flags,omitempty" json:"flags,omitempty"`

	// Discovered rooms and areas for exploration XP tracking
	DiscoveredRooms map[string]bool `bson:"discoveredRooms,omitempty" json:"-"`
	DiscoveredAreas map[string]bool `bson:"discoveredAreas,omitempty" json:"-"`

	// Per-character revealed exits (hidden exits unlocked via scripts)
	// Maps roomID → list of exit names the character has revealed
	RevealedExits map[string][]string `bson:"revealedExits,omitempty" json:"revealedExits,omitempty"`

	// track alltime stats in character object but dont expose as json by default
	AllTimeStats struct {
		PlayersKilled   int32 `bson:"playersKilled" json:"playersKilled"`
		GoldCollected   int32 `bson:"goldCollected" json:"goldCollected"`
		QuestsCompleted int32 `bson:"questsCompleted" json:"questsCompleted"`
		RoomsDiscovered int32 `bson:"roomsDiscovered" json:"roomsDiscovered"`
	} `bson:"allTimeStats" json:"_"`
}

// GetAttribute returns the value of an attribute by its short name (STR, DEX, etc.)
func (c *Character) GetAttribute(short string) int32 {
	short = strings.ToUpper(short)
	for _, attr := range c.Attributes {
		if strings.EqualFold(attr.Short, short) {
			return attr.Value
		}
	}
	return 10 // Default value if not found
}

// GetAttributeModifier returns the modifier for an attribute (value - 10) / 2
func (c *Character) GetAttributeModifier(short string) int {
	value := c.GetAttribute(strings.ToUpper(short))
	return int(value-10) / 2
}

// GetPrimaryAttackAttribute returns the attribute short name used for basic attack scaling.
// Each class uses a different primary attribute for auto-attacks.
func (c *Character) GetPrimaryAttackAttribute() string {
	switch strings.ToLower(c.Class.ID) {
	case "warrior":
		return "STR"
	case "rogue":
		return "DEX"
	case "hunter", "ranger":
		return "DEX"
	case "wizard", "mage":
		return "INT"
	case "cleric":
		return "WIS"
	case "druid":
		return "INT"
	default:
		return "STR"
	}
}

// GetPrimaryAttackMod returns the modifier from the class's primary attack attribute.
func (c *Character) GetPrimaryAttackMod() int {
	return c.GetAttributeModifier(c.GetPrimaryAttackAttribute())
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

// GetINTMod returns the intelligence modifier
func (c *Character) GetINTMod() int {
	return c.GetAttributeModifier("INT")
}

// GetWISMod returns the wisdom modifier
func (c *Character) GetWISMod() int {
	return c.GetAttributeModifier("WIS")
}

// CalculateMaxMana returns the max mana for this character based on class, level, and INT
func (c *Character) CalculateMaxMana() int32 {
	classID := strings.ToLower(c.Class.ID)
	switch classID {
	case "mage", "wizard", "cleric", "druid":
		intMod := c.GetINTMod()
		mana := int32(20) + (c.Level * 5) + int32(intMod*4)
		if mana < 10 {
			mana = 10
		}
		return mana
	default:
		return 0 // Physical classes don't use mana
	}
}

// CalculateManaRegen returns in-combat mana regen per round
func (c *Character) CalculateManaRegen() int32 {
	if c.MaxMana <= 0 {
		return 0
	}
	regen := int32(1) + int32(c.GetWISMod())
	if regen < 1 {
		regen = 1
	}
	return regen
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
		if dmg, ok := toInt32(damage); ok {
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
			if def, ok := toInt32(defense); ok {
				totalDefense += def
			}
		}
		if armor, ok := item.Attributes["armor"]; ok {
			if arm, ok := toInt32(armor); ok {
				totalDefense += arm
			}
		}
	}
	return totalDefense
}

// HasCollectedCopyItem checks if the character has already collected a CopyOnPickup item
func (c *Character) HasCollectedCopyItem(templateID string) bool {
	if c.Flags == nil {
		return false
	}
	key := "collected_item:" + templateID
	val, ok := c.Flags[key]
	if !ok {
		return false
	}
	if b, ok := val.(bool); ok {
		return b
	}
	return false
}

// MarkCollectedCopyItem records that the character has collected a CopyOnPickup item
func (c *Character) MarkCollectedCopyItem(templateID string) {
	if c.Flags == nil {
		c.Flags = make(map[string]interface{})
	}
	c.Flags["collected_item:"+templateID] = true
}

// HasRevealedExit checks if the character has revealed a specific hidden exit in a room
func (c *Character) HasRevealedExit(roomID, exitName string) bool {
	if c.RevealedExits == nil {
		return false
	}
	for _, name := range c.RevealedExits[roomID] {
		if strings.EqualFold(name, exitName) {
			return true
		}
	}
	return false
}

// RevealExit records that the character has revealed a hidden exit in a room
func (c *Character) RevealExit(roomID, exitName string) {
	if c.RevealedExits == nil {
		c.RevealedExits = make(map[string][]string)
	}
	// Don't add duplicates
	if c.HasRevealedExit(roomID, exitName) {
		return
	}
	c.RevealedExits[roomID] = append(c.RevealedExits[roomID], exitName)
}
