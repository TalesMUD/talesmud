package characters

// Race type
type Race struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Heritage    string `json:"heritage"`
}

// TODO: Move this to DB or YML
var (
	RaceDwarf Race = Race{
		ID:          "dwarf",
		Name:        "Dwarf",
		Description: "Blunt weapons deal 10% more damage.",
		Heritage:    "Deep below the mountains",
	}
	RaceHuman Race = Race{
		ID:          "human",
		Name:        "Human",
		Description: "No weapon bonus. A few extra coins when you start.",
		Heritage:    "Big cities",
	}
	// RaceElf is the canonical elf. Saved rows and portraits may still say elve.
	RaceElf Race = Race{
		ID:          "elf",
		Name:        "Elf",
		Description: "Bows deal 10% more damage.",
		Heritage:    "Near the forest",
	}
	// RaceElve is the old name for RaceElf. New creates store id elf.
	RaceElve           = RaceElf
	RaceConstruct Race = Race{
		ID:          "construct",
		Name:        "Construct",
		Description: "Poison never sticks. No weapon bonus.",
		Heritage:    "Workshops",
	}
)
