package characters

//CharacterTemplates TODO: move to DB or yml?
var CharacterTemplates = []Character{
	Character{
		Name:             "Mr. Warrior",
		Description:      "Select to create a new warrior",
		Race:             RaceHuman,
		Class:            ClassWarrior,
		CurrentHitPoints: 25,
		MaxHitPoints:     25,
		XP:               0,
		Level:            1,
		// basepoints for level 1 should sum to 50
		Attributes: createBaseAttributes(11, 8, 5, 6, 20),
	},
	Character{
		Name:             "Mrs. Longstride",
		Description:      "Select to create a new ranger",
		Race:             RaceElve,
		Class:            ClassRanger,
		CurrentHitPoints: 20,
		MaxHitPoints:     20,
		XP:               0,
		Level:            1,
		// basepoints for level 1 should sum to 50
		Attributes: createBaseAttributes(7, 15, 5, 6, 17),
	},
	Character{
		Name:             "Gumdalf the Bay",
		Description:      "Select to create the allmighty wizard",
		Race:             RaceDwarf,
		Class:            ClassWizard,
		CurrentHitPoints: 17,
		MaxHitPoints:     17,
		XP:               0,
		Level:            1,
		// basepoints for level 1 should sum to 50
		Attributes: createBaseAttributes(5, 5, 15, 12, 13),
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
