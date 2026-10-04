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
		Description: "The door. You stand in it until they don't. Brace once when a blow comes in.",
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
		Description: "Back street. Two cuts, then you Slip the first one that comes back.",
		ArmorType:   ArmorTypeLeather,
		CombatType:  CombatTypeMelee,
	}
	ClassWizard Class = Class{
		ID:          "wizard",
		Name:        "Rune Hand",
		Description: "Vault runes. One heavy swing, then you Inscribe. The mark burns for three rounds.",
		ArmorType:   ArmorTypeCloth,
		CombatType:  CombatTypeMagic,
	}
	ClassWard Class = Class{
		ID:          "ward",
		Name:        "Ward",
		Description: "Heavy plate. You start slow. Hits you take stack Grit, and Slam and the hit you throw back get heavier.",
		ArmorType:   ArmorTypePlate,
		CombatType:  CombatTypeMelee,
	}
	ClassRigger Class = Class{
		ID:          "rigger",
		Name:        "Rigger",
		Description: "Constructs. Bolt scrap onto someone in the room; the next hit still lands, and the attacker takes the same amount back. Rig drops a turret that does not chase.",
		ArmorType:   ArmorTypeLeather,
		CombatType:  CombatTypeMelee,
	}
)
