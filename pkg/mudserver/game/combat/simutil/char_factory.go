package simutil

import (
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/skills"
	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
)

// ClassCleric is the Cleric class (not in class.go package vars)
var ClassCleric = characters.Class{
	ID:          "cleric",
	Name:        "Cleric",
	Description: "Holy spellcaster and protector",
	ArmorType:   characters.ArmorTypeLeather,
	CombatType:  characters.CombatTypeMagic,
}

// ClassDruid is the Druid class (not in class.go package vars)
var ClassDruid = characters.Class{
	ID:          "druid",
	Name:        "Druid",
	Description: "Nature spellcaster and shapeshifter",
	ArmorType:   characters.ArmorTypeLeather,
	CombatType:  characters.CombatTypeMagic,
}

// ClassConfig holds the base stats for a character class
type ClassConfig struct {
	Name       string
	Class      characters.Class
	Race       characters.Race
	STR        int32
	DEX        int32
	INT        int32
	WIS        int32
	STA        int32
	HP         int32
	WeaponName string
	ArmorName  string
}

// AllClassConfigs returns configurations for all 6 starter classes
func AllClassConfigs() []ClassConfig {
	return []ClassConfig{
		{"Warrior", characters.ClassWarrior, characters.RaceHuman, 14, 7, 4, 5, 20, 25, "Rusty Sword", "Leather Armor"},
		{"Rogue", characters.ClassRogue, characters.RaceHuman, 10, 18, 6, 5, 11, 19, "Worn Dagger", "Leather Armor"},
		{"Mage", characters.ClassWizard, characters.RaceDwarf, 4, 6, 18, 14, 8, 17, "Apprentice Staff", "Cloth Robe"},
		{"Cleric", ClassCleric, characters.RaceHuman, 6, 7, 10, 18, 9, 22, "Simple Mace", "Cloth Robe"},
		{"Ranger", characters.ClassRanger, characters.RaceElve, 8, 18, 6, 6, 12, 20, "Short Bow", "Leather Armor"},
		{"Druid", ClassDruid, characters.RaceElve, 5, 8, 14, 18, 5, 20, "Wooden Staff", "Cloth Robe"},
	}
}

// ClassConfigByName returns the class config for the given class name, or nil
func ClassConfigByName(name string) *ClassConfig {
	for _, c := range AllClassConfigs() {
		if c.Name == name {
			return &c
		}
	}
	return nil
}

// CreateCharacter creates a fully equipped character at the specified level.
// It uses the real leveling system to apply HP and attribute gains.
func CreateCharacter(config ClassConfig, level int32) *characters.Character {
	char := &characters.Character{
		Entity: entities.NewEntity(),
		Name:   config.Name,
		Class:  config.Class,
		Race:   config.Race,
		Level:  1,
		XP:     0,
		MaxHitPoints:     config.HP,
		CurrentHitPoints: config.HP,
		Attributes: characters.Attributes{
			characters.NewAttribute("Strength", "STR", config.STR),
			characters.NewAttribute("Dexterity", "DEX", config.DEX),
			characters.NewAttribute("Intelligence", "INT", config.INT),
			characters.NewAttribute("Wisdom", "WIS", config.WIS),
			characters.NewAttribute("Stamina", "STA", config.STA),
		},
		EquippedItems: make(map[items.ItemSlot]*items.Item),
	}

	// Equip weapon
	if weapon := items.StarterItemTemplateByName(config.WeaponName); weapon != nil {
		weaponCopy := *weapon
		weaponCopy.Entity = entities.NewEntity()
		char.EquippedItems[items.ItemSlotMainHand] = &weaponCopy
	}

	// Equip armor
	if armor := items.StarterItemTemplateByName(config.ArmorName); armor != nil {
		armorCopy := *armor
		armorCopy.Entity = entities.NewEntity()
		char.EquippedItems[items.ItemSlotChest] = &armorCopy
	}

	// Level up from 1 to target level using the real leveling system
	if level > 1 {
		leveling.ApplyLevelUp(char, int(level-1))
	}

	// Ensure full HP after leveling
	char.CurrentHitPoints = char.MaxHitPoints

	// Set up mana for caster classes
	char.MaxMana = char.CalculateMaxMana()
	char.CurrentMana = char.MaxMana

	// Equip all available skills for the character's class and level
	classID := char.Class.ID
	available := skills.AvailableSkills(classID, char.Level)
	maxSlots := skills.MaxSkillSlots(classID, char.Level)
	equipped := make([]string, 0, maxSlots)
	for i, s := range available {
		if i >= maxSlots {
			break
		}
		equipped = append(equipped, s.ID)
	}
	char.EquippedSkills = equipped

	return char
}

// CreateCharacterWithGear creates a character with custom weapon and armor stats.
// Useful for testing different equipment tiers.
func CreateCharacterWithGear(config ClassConfig, level int32, weaponDamage int, armorDefense int) *characters.Character {
	char := CreateCharacter(config, level)

	// Override weapon
	if weaponDamage > 0 {
		char.EquippedItems[items.ItemSlotMainHand] = &items.Item{
			Entity: entities.NewEntity(),
			Name:   "Test Weapon",
			Type:   items.ItemTypeWeapon,
			Slot:   items.ItemSlotMainHand,
			Attributes: map[string]interface{}{
				"damage": weaponDamage,
			},
		}
	}

	// Override armor
	if armorDefense > 0 {
		char.EquippedItems[items.ItemSlotChest] = &items.Item{
			Entity: entities.NewEntity(),
			Name:   "Test Armor",
			Type:   items.ItemTypeArmor,
			Slot:   items.ItemSlotChest,
			Attributes: map[string]interface{}{
				"armor": armorDefense,
			},
		}
	}

	return char
}
