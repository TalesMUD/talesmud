package items

// StarterItemTemplatePresets returns a list of starter item templates for character creation.
// These are meant to be seeded alongside character templates.
func StarterItemTemplatePresets() []*Item {
	return []*Item{
		// Weapons
		{
			IsTemplate:  true,
			Name:        "Rusty Sword",
			Description: "A worn iron sword, chipped but serviceable.",
			Type:        ItemTypeWeapon,
			SubType:     ItemSubTypeSword,
			Slot:        ItemSlotMainHand,
			Quality:     ItemQualityNormal,
			Level:       1,
			Attributes: map[string]interface{}{
				"damage":      5,
				"attackSpeed": 1.2,
			},
		},
		{
			IsTemplate:  true,
			Name:        "Worn Dagger",
			Description: "A slim blade favored by those who strike from shadows.",
			Type:        ItemTypeWeapon,
			SubType:     "dagger",
			Slot:        ItemSlotMainHand,
			Quality:     ItemQualityNormal,
			Level:       1,
			Attributes: map[string]interface{}{
				"damage":      3,
				"attackSpeed": 1.8,
				"critChance":  5,
			},
		},
		{
			IsTemplate:  true,
			Name:        "Apprentice Staff",
			Description: "A simple wooden staff imbued with faint magical energy.",
			Type:        ItemTypeWeapon,
			SubType:     "staff",
			Slot:        ItemSlotMainHand,
			Quality:     ItemQualityNormal,
			Level:       1,
			Attributes: map[string]interface{}{
				"damage":      2,
				"spellPower":  5,
				"attackSpeed": 1.0,
			},
		},
		{
			IsTemplate:  true,
			Name:        "Simple Mace",
			Description: "A sturdy mace blessed by a minor shrine.",
			Type:        ItemTypeWeapon,
			SubType:     "mace",
			Slot:        ItemSlotMainHand,
			Quality:     ItemQualityNormal,
			Level:       1,
			Attributes: map[string]interface{}{
				"damage":      4,
				"holyDamage":  2,
				"attackSpeed": 1.1,
			},
		},
		{
			IsTemplate:  true,
			Name:        "Short Bow",
			Description: "A compact hunting bow, accurate at medium range.",
			Type:        ItemTypeWeapon,
			SubType:     "bow",
			Slot:        ItemSlotMainHand,
			Quality:     ItemQualityNormal,
			Level:       1,
			Attributes: map[string]interface{}{
				"damage":      4,
				"range":       30,
				"attackSpeed": 1.4,
			},
		},
		{
			IsTemplate:  true,
			Name:        "Wooden Staff",
			Description: "A gnarled branch shaped by nature's hand, thrumming with primal energy.",
			Type:        ItemTypeWeapon,
			SubType:     "staff",
			Slot:        ItemSlotMainHand,
			Quality:     ItemQualityNormal,
			Level:       1,
			Attributes: map[string]interface{}{
				"damage":      2,
				"naturePower": 5,
				"attackSpeed": 1.0,
			},
		},
		// Armor
		{
			IsTemplate:  true,
			Name:        "Leather Armor",
			Description: "Simple tanned leather offering basic protection.",
			Type:        ItemTypeArmor,
			SubType:     "chest",
			Slot:        ItemSlotChest,
			Quality:     ItemQualityNormal,
			Level:       1,
			Attributes: map[string]interface{}{
				"armor":   8,
				"agility": 1,
			},
		},
		{
			IsTemplate:  true,
			Name:        "Cloth Robe",
			Description: "A plain robe favored by scholars and spellcasters.",
			Type:        ItemTypeArmor,
			SubType:     "chest",
			Slot:        ItemSlotChest,
			Quality:     ItemQualityNormal,
			Level:       1,
			Attributes: map[string]interface{}{
				"armor":      3,
				"spellPower": 2,
				"mana":       10,
			},
		},
	}
}

// StarterItemTemplateByName returns the starter item template with the given name, or nil if not found.
func StarterItemTemplateByName(name string) *Item {
	for _, t := range StarterItemTemplatePresets() {
		if t.Name == name {
			return t
		}
	}
	return nil
}
