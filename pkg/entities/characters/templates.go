package characters

//CharacterTemplate ...
type CharacterTemplate struct {
	Character
	TemplateID int32 `json:"templateId"`
}

//CharacterTemplates TODO: move to DB or yml?
var CharacterTemplates = []CharacterTemplate{
	{
		TemplateID: 1,
		Character: Character{
			Name:             "Mr. Warrior",
			Description:      "Create a new Warrior",
			Race:             RaceHuman,
			Class:            ClassWarrior,
			CurrentHitPoints: 25,
			MaxHitPoints:     25,
			XP:               0,
			Level:            1,
			// basepoints for level 1 should sum to 50
			Attributes: createBaseAttributes(11, 8, 5, 6, 20),
		},
	},
	{

		TemplateID: 2,
		Character: Character{
			Name:             "Mrs. Longstride",
			Description:      "Create a new Ranger",
			Race:             RaceElve,
			Class:            ClassRanger,
			CurrentHitPoints: 20,
			MaxHitPoints:     20,
			XP:               0,
			Level:            1,
			// basepoints for level 1 should sum to 50
			Attributes: createBaseAttributes(7, 15, 5, 6, 17),
		},
	},
	{
		TemplateID: 3,
		Character: Character{
			Name:             "Gumdalf the Bay",
			Description:      "Create a new Wizard",
			Race:             RaceDwarf,
			Class:            ClassWizard,
			CurrentHitPoints: 17,
			MaxHitPoints:     17,
			XP:               0,
			Level:            1,
			// basepoints for level 1 should sum to 50
			Attributes: createBaseAttributes(5, 5, 15, 12, 13),
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
