package characters

//Race type
type Race struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Heritage    string `json:"heritage"`
}

//TODO: Move this to DB or YML
var (
	RaceDwarf Race = Race{
		ID:          "dwarf",
		Name:        "Dwarf",
		Description: "Small, but dont underestimate them",
		Heritage:    "Deep below the mountains",
	}
	RaceHuman Race = Race{
		ID:          "human",
		Name:        "Human",
		Description: "The common race",
		Heritage:    "Big cities",
	}
	RaceElve Race = Race{
		ID:          "elve",
		Name:        "Elve",
		Description: "Splendid forestwalkers, great sight during nights",
		Heritage:    "Near the forest",
	}
)
