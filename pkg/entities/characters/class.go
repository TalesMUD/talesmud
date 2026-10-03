package characters

type Class struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ArmorType   ArmorType  `json:"armorType"`
	CombatType  CombatType `json:"combatType"`
	//Spells Spells[]
}

// ArmorType type
type ArmorType string

// armor types
const (
	ArmorTypeCloth   ArmorType = "Cloth"
	ArmorTypeLeather           = "Leather"
	ArmorTypePlate             = "Plate"
)

// CombatType ...
type CombatType string

// combat types
const (
	CombatTypeMelee CombatType = "Melee"
	CombaTypeRanged            = "Ranged"
	CombatTypeMagic            = "Magic"
)

// TODO: Move this to Database or YML files
var (
	ClassWarrior Class = Class{
		ID:          "warrior",
		Name:        "Fenwatch",
		Description: "One swing, and you mean it. Extra health, hits land lighter. Brace once when a blow comes in.",
		ArmorType:   ArmorTypePlate,
		CombatType:  CombatTypeMelee,
	}
	ClassRanger Class = Class{
		ID:          "ranger",
		Name:        "Ranger",
		Description: "Quick Bow wielding ranged combatant",
		ArmorType:   ArmorTypeLeather,
		CombatType:  CombaTypeRanged,
	}
	ClassHunter Class = Class{
		ID:          "hunter",
		Name:        "Hunter",
		Description: "Woodsman and tracker, deadly at range and in the wilds",
		ArmorType:   ArmorTypeLeather,
		CombatType:  CombaTypeRanged,
	}
	ClassRogue Class = Class{
		ID:          "rogue",
		Name:        "Alley",
		Description: "Two short swings, dagger or bow. Less health, you feel hits more. Slip the first one.",
		ArmorType:   ArmorTypeLeather,
		CombatType:  CombatTypeMelee,
	}
	ClassWizard Class = Class{
		ID:          "wizard",
		Name:        "Rune Hand",
		Description: "One heavy swing, then the rune burns for three rounds. Thin on health. The basic costs no mana.",
		ArmorType:   ArmorTypeCloth,
		CombatType:  CombatTypeMagic,
	}
	ClassHitch Class = Class{
		ID:          "hitch",
		Name:        "Hitch",
		Description: "One careful swing. A bit more health. Pin once: the next time they try to leave, they stay.",
		ArmorType:   ArmorTypeLeather,
		CombatType:  CombatTypeMelee,
	}
)
