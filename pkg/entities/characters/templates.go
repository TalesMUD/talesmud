package characters

// CharacterCreationTemplate is the legacy hardcoded template type used by the old
// character creation endpoint. This will be replaced by DB-backed templates.
type CharacterCreationTemplate struct {
	Character
	TemplateID int32 `json:"templateId"`
}

// CharacterCreationTemplates TODO: move to DB or yml?
var CharacterCreationTemplates = []CharacterCreationTemplate{
	{
		TemplateID: 1,
		Character: Character{
			Name:             "Warrior",
			Description:      "Frontline melee fighter with high durability",
			Race:             RaceHuman,
			Class:            ClassWarrior,
			CurrentHitPoints: 25,
			MaxHitPoints:     25,
			XP:               0,
			Level:            1,
			// basepoints for level 1 should sum to 50
			Attributes: createBaseAttributes(14, 7, 4, 5, 20),
		},
	},
	{

		TemplateID: 2,
		Character: Character{
			Name:             "Hunter",
			Description:      "Ranged specialist with great mobility and stamina",
			Race:             RaceElve,
			Class:            ClassHunter,
			CurrentHitPoints: 20,
			MaxHitPoints:     20,
			XP:               0,
			Level:            1,
			// basepoints for level 1 should sum to 50
			Attributes: createBaseAttributes(8, 18, 6, 6, 12),
		},
	},
	{
		TemplateID: 3,
		Character: Character{
			Name:             "Wizard",
			Description:      "Spellcaster with high intelligence and wisdom",
			Race:             RaceDwarf,
			Class:            ClassWizard,
			CurrentHitPoints: 17,
			MaxHitPoints:     17,
			XP:               0,
			Level:            1,
			// basepoints for level 1 should sum to 50
			Attributes: createBaseAttributes(4, 6, 18, 14, 8),
		},
	},
	{
		TemplateID: 4,
		Character: Character{
			Name:             "Rogue",
			Description:      "Swift melee skirmisher focused on dexterity and precision",
			Race:             RaceHuman,
			Class:            ClassRogue,
			CurrentHitPoints: 19,
			MaxHitPoints:     19,
			XP:               0,
			Level:            1,
			// basepoints for level 1 should sum to 50
			Attributes: createBaseAttributes(10, 18, 6, 5, 11),
		},
	},
}

func createBaseAttributes(str, dex, _int, wis, sta int32) Attributes {
	return []Attribute{
		NewAttribute("Strength", "str", str),
		NewAttribute("Dexterity", "dex", dex),
		NewAttribute("Intelligence", "int", _int),
		NewAttribute("Wisdom", "wis", wis),
		NewAttribute("Stamina", "sta", sta),
	}
}
