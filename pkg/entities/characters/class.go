package characters

type Class struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ArmorType   ArmorType  `json:"armorType"`
	CombatType  CombatType `json:"combatType"`
	//Spells Spells[]
}

//ArmorType type
type ArmorType string

const (
	ArmorTypeCloth   ArmorType = "Cloth"
	ArmorTypeLeather           = "Leather"
	ArmorTypePlate             = "Plate"
)

type CombatType string

const (
	CombatTypeMelee CombatType = "Melee"
	CombaTypeRanged            = "Ranged"
	CombatTypeMagic            = "Magic"
)

//TODO: Move this to Database or YML files
var (
	ClassWarrior Class = Class{
		ID:          "warrior",
		Name:        "Warrior",
		Description: "Strong plate wearing melee warrior",
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
	ClassWizard Class = Class{
		ID:          "wizard",
		Name:        "Wizard",
		Description: "Master of the elements wielder of magic spells",
		ArmorType:   ArmorTypeCloth,
		CombatType:  CombatTypeMagic,
	}
)
