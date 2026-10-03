package characters

import (
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/mudserver/game/balance"
)

// Playable roster. HP uses balance.ScaleClassHP once off a shared base of 25.
// Ranger and hunter are not classes here; Alley uses those weapons.
const classHPBase int32 = 25

func presetEntity(id string) *entities.Entity {
	return &entities.Entity{ID: id}
}

// SystemCharacterTemplatePresets is the signed create roster: Fenwatch, Alley, Rune Hand, Hitch, Rigger.
func SystemCharacterTemplatePresets() []*CharacterTemplate {
	fenHP := balance.ScaleClassHP("warrior", classHPBase)
	alleyHP := balance.ScaleClassHP("rogue", classHPBase)
	runeHP := balance.ScaleClassHP("wizard", classHPBase)
	hitchHP := balance.ScaleClassHP("hitch", classHPBase)
	riggerHP := balance.ScaleClassHP("rigger", classHPBase)
	return []*CharacterTemplate{
		{
			Entity:           presetEntity("tpl-fenwatch"),
			Name:             "Fenwatch",
			Description:      "The door. You stand in it until they don't. Brace once when a blow comes in.",
			Backstory:        "You held a wet road until the wagons were through.",
			OriginArea:       "Oldtown",
			Archetype:        "fenwatch",
			Race:             RaceHuman,
			Class:            ClassWarrior,
			Level:            1,
			CurrentHitPoints: fenHP,
			MaxHitPoints:     fenHP,
			Attributes:       createBaseAttributes(14, 7, 4, 5, 20),
			StartingItems: []StartingItem{
				{Slot: items.ItemSlotMainHand, ItemTemplateName: "Rusty Sword"},
				{Slot: items.ItemSlotChest, ItemTemplateName: "Leather Armor"},
			},
			DefaultSkills: []string{"warrior_power_strike"},
			Source:        "system",
		},
		{
			Entity:           presetEntity("tpl-alley"),
			Name:             "Alley",
			Description:      "Back street. Two cuts, then you Slip the first one that comes back.",
			Backstory:        "You learned the lanes with a knife, and the tree line with a bow.",
			OriginArea:       "Dockside",
			Archetype:        "alley",
			Race:             RaceHuman,
			Class:            ClassRogue,
			Level:            1,
			CurrentHitPoints: alleyHP,
			MaxHitPoints:     alleyHP,
			Attributes:       createBaseAttributes(10, 18, 6, 5, 11),
			StartingItems: []StartingItem{
				{Slot: items.ItemSlotMainHand, ItemTemplateName: "Worn Dagger"},
				{Slot: items.ItemSlotChest, ItemTemplateName: "Leather Armor"},
			},
			DefaultSkills: []string{"rogue_backstab"},
			Source:        "system",
		},
		{
			Entity:           presetEntity("tpl-runehand"),
			Name:             "Rune Hand",
			Description:      "Vault runes. One heavy swing, then you Inscribe. The mark burns for three rounds.",
			Backstory:        "You scratch a mark and it keeps burning after the staff goes still.",
			OriginArea:       "Arcane Tower",
			Archetype:        "runehand",
			Race:             RaceHuman,
			Class:            ClassWizard,
			Level:            1,
			CurrentHitPoints: runeHP,
			MaxHitPoints:     runeHP,
			CurrentMana:      41,
			MaxMana:          41,
			Attributes:       createBaseAttributes(4, 6, 18, 14, 8),
			StartingItems: []StartingItem{
				{Slot: items.ItemSlotMainHand, ItemTemplateName: "Apprentice Staff"},
				{Slot: items.ItemSlotChest, ItemTemplateName: "Cloth Robe"},
			},
			DefaultSkills: []string{"mage_fireball", "mage_frost_shield"},
			Source:        "system",
		},
		{
			Entity:           presetEntity("tpl-hitch"),
			Name:             "Hitch",
			Description:      "Fen rope. One careful swing. Pin once: the next time they try to leave, they stay.",
			Backstory:        "You set the snare and wait. When they turn to run, the line is already tight.",
			OriginArea:       "Forest Edge",
			Archetype:        "hitch",
			Race:             RaceHuman,
			Class:            ClassHitch,
			Level:            1,
			CurrentHitPoints: hitchHP,
			MaxHitPoints:     hitchHP,
			Attributes:       createBaseAttributes(12, 12, 6, 8, 12),
			StartingItems: []StartingItem{
				{Slot: items.ItemSlotMainHand, ItemTemplateName: "Worn Dagger"},
				{Slot: items.ItemSlotChest, ItemTemplateName: "Leather Armor"},
			},
			DefaultSkills: []string{},
			Source:        "system",
		},
		{
			Entity:           presetEntity("tpl-rigger"),
			Name:             "Rigger",
			Description:      "Constructs. Bolt scrap onto someone in the room; the next hit still lands, and the attacker takes the same amount back. Rig drops a turret that does not chase.",
			Backstory:        "You bolt scrap onto a target and leave a turret where you stood.",
			OriginArea:       "Gear Yard",
			Archetype:        "rigger",
			Race:             RaceConstruct,
			Class:            ClassRigger,
			Level:            1,
			CurrentHitPoints: riggerHP,
			MaxHitPoints:     riggerHP,
			Attributes:       createBaseAttributes(12, 10, 8, 6, 14),
			StartingItems: []StartingItem{
				{Slot: items.ItemSlotMainHand, ItemTemplateName: "Rusty Sword"},
				{Slot: items.ItemSlotChest, ItemTemplateName: "Leather Armor"},
			},
			DefaultSkills: []string{},
			Source:        "system",
		},
	}
}

// PresetByID finds a signed create template by its stable id.
func PresetByID(id string) *CharacterTemplate {
	if id == "" {
		return nil
	}
	for _, preset := range SystemCharacterTemplatePresets() {
		if preset != nil && preset.Entity != nil && preset.Entity.ID == id {
			return preset
		}
	}
	return nil
}
