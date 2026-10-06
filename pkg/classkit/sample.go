package classkit

func kit(id, name, effect string, level int32, keeps, once bool, cd int, mult float64, tip string) SkillSpec {
	return SkillSpec{
		ID: id, Name: name, Effect: effect, Level: level,
		KeepsSwing: keeps, OncePerFight: once, Cooldown: cd, Multiplier: mult,
		Tooltip: tip, Target: "enemy",
	}
}

func sampleDefs() []*Def {
	warriorSkills := []SkillSpec{
		kit("warrior_brace", "Brace", "brace", 1, true, true, 0, 0,
			"Once a fight. The next hit on you is halved. You still swing."),
		kit("warrior_slam", "Slam", "slam", 4, false, false, 4, 1.40,
			"A heavy swing, 1.40×. Replaces your swing this round."),
		kit("warrior_stand", "Stand", "stand", 8, false, true, 0, 0,
			"Once a fight, for two rounds hits aimed at others hit you."),
	}
	rogueSkills := []SkillSpec{
		kit("rogue_slip", "Slip", "slip", 1, false, true, 0, 0,
			"Once a fight. Drop the fight and take one exit."),
		kit("rogue_nick", "Nick", "nick", 4, false, false, 3, 0.55,
			"Extra 0.55× swing on top of your cadence. Replaces the autoattack this round."),
		kit("rogue_smoke", "Smoke", "smoke", 8, false, true, 0, 0,
			"Once a fight. The target misses their next swing."),
	}
	mageSkills := []SkillSpec{
		kit("mage_inscribe", "Inscribe", "inscribe", 1, false, false, 4, 0,
			"Mark them. 4 a round for 3 rounds. Refresh, no stack. Costs no mana."),
		kit("mage_sear", "Sear", "sear", 4, false, false, 4, 1.80,
			"A cast at 1.80×. No burn. Replaces your swing this round."),
		kit("mage_glyph", "Glyph", "glyph", 8, false, true, 0, 0,
			"Once a fight. The next hit on you is reduced by 4."),
	}
	return []*Def{
		{
			ID: "warrior", Name: "Warrior", Order: 1,
			Description: "Sample warrior. Plate, one swing, and a brace that halves the next hit.",
			ArmorType:   "Plate", CombatType: "Melee", Primary: "STR", SkillClass: "warrior",
			Races: []string{"human", "dwarf"}, PortraitClass: "warrior",
			HPMultiplier: 1.20, Hotbar: 4, BraceCharges: 1,
			Balance: Row{DamageDealt: 1, DamageTaken: 0.90, BehindDealt: 1.15, Swings: 1},
			Skills:  warriorSkills,
			Template: &Template{
				ID: "tpl-warrior", Backstory: "You learned to hold a line.",
				Origin: "Town", Archetype: "warrior", Race: "human",
				Str: 14, Dex: 7, Int: 4, Wis: 5, Sta: 20,
				Items: []StartItem{
					{Slot: "main_hand", Name: "Rusty Sword"},
					{Slot: "chest", Name: "Leather Armor"},
				},
				DefaultSkills: []string{"warrior_brace"},
			},
		},
		{
			ID: "rogue", Name: "Rogue", Order: 2,
			Description: "Sample rogue. Leather, two lighter swings, and one slip out of a fight.",
			ArmorType:   "Leather", CombatType: "Melee", Primary: "DEX", SkillClass: "rogue",
			Also:  []string{"ranger", "hunter"},
			Races: []string{"human", "dwarf", "elf"}, PortraitClass: "rogue",
			HPMultiplier: 0.85, Hotbar: 4, SlipCharges: 1,
			Balance: Row{DamageDealt: 0.55, DamageTaken: 1.15, BehindDealt: 1.15, Swings: 2},
			Skills:  rogueSkills,
			Template: &Template{
				ID: "tpl-rogue", Backstory: "You learned the lanes with a knife.",
				Origin: "Town", Archetype: "rogue", Race: "human",
				Str: 10, Dex: 18, Int: 6, Wis: 5, Sta: 11,
				Items: []StartItem{
					{Slot: "main_hand", Name: "Worn Dagger"},
					{Slot: "chest", Name: "Leather Armor"},
				},
				DefaultSkills: []string{"rogue_slip"},
			},
		},
		{
			ID: "wizard", Name: "Mage", Order: 3,
			Description: "Sample mage. Cloth, one heavy swing, then a mark that burns.",
			ArmorType:   "Cloth", CombatType: "Magic", Primary: "INT", SkillClass: "mage",
			Aliases: []string{"mage"},
			Races:   []string{"human", "elf"}, PortraitClass: "mage",
			HPMultiplier: 0.75, Caster: true, Hotbar: 4, InscribeBasic: true,
			Balance: Row{DamageDealt: 1.40, DamageTaken: 1.25, BehindDealt: 1.15, Swings: 1},
			Skills:  mageSkills,
			Template: &Template{
				ID: "tpl-mage", Backstory: "You scratch a mark and it keeps burning.",
				Origin: "Town", Archetype: "mage", Race: "human",
				Str: 4, Dex: 6, Int: 18, Wis: 14, Sta: 8, Mana: 41,
				Items: []StartItem{
					{Slot: "main_hand", Name: "Apprentice Staff"},
					{Slot: "chest", Name: "Cloth Robe"},
				},
				DefaultSkills: []string{"mage_inscribe"},
			},
		},
	}
}
