package characters

import "github.com/talesmud/talesmud/pkg/classkit"

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

// Class vars mirror the catalog. A world pack overwrites the sample names.
// Ranger and hunter stay their own classes; they only share a damage row.
var (
	ClassWarrior Class = Class{
		ID:          "warrior",
		Name:        "Warrior",
		Description: "Sample warrior. Plate, one swing, and a brace that halves the next hit.",
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
		Name:        "Rogue",
		Description: "Sample rogue. Leather, two lighter swings, and one slip out of a fight.",
		ArmorType:   ArmorTypeLeather,
		CombatType:  CombatTypeMelee,
	}
	ClassWizard Class = Class{
		ID:          "wizard",
		Name:        "Mage",
		Description: "Sample mage. Cloth, one heavy swing, then a mark that burns.",
		ArmorType:   ArmorTypeCloth,
		CombatType:  CombatTypeMagic,
	}
	ClassWard Class = Class{
		ID:          "ward",
		Name:        "Ward",
		Description: "Plate. Hits you take stack, and the answer gets heavier.",
		ArmorType:   ArmorTypePlate,
		CombatType:  CombatTypeMelee,
	}
)

func init() {
	classkit.OnChange(syncClassVars)
	syncClassVars()
}

func syncClassVars() {
	if d := classkit.Lookup("warrior"); d != nil {
		ClassWarrior = classFromDef(d)
	}
	if d := classkit.Lookup("rogue"); d != nil {
		ClassRogue = classFromDef(d)
	}
	if d := classkit.Lookup("wizard"); d != nil {
		ClassWizard = classFromDef(d)
	}
	if d := classkit.Lookup("ward"); d != nil {
		ClassWard = classFromDef(d)
	}
}

func classFromDef(d *classkit.Def) Class {
	if d == nil {
		return Class{}
	}
	return Class{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		ArmorType:   ArmorType(d.ArmorType),
		CombatType:  CombatType(d.CombatType),
	}
}

// ClassByID resolves a catalog id, alias, template id, or display name.
// Ranger and hunter are not catalog classes.
func ClassByID(id string) (Class, bool) {
	d := classkit.Lookup(id)
	if d == nil {
		return Class{}, false
	}
	return classFromDef(d), true
}
