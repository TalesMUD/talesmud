package recipes

import "github.com/talesmud/talesmud/pkg/entities"

// SeedRecipes returns the default v1 crafting recipes (used when no YAML is present).
func SeedRecipes() []*Recipe {
	return []*Recipe{
		{
			Entity: &entities.Entity{ID: "RCP0001"}, Key: "meadow_stew", Name: "Meadow Stew",
			Description: "A warming stew of meadow herbs and wild apple.",
			Category:    "food", Station: "",
			Ingredients: []Ingredient{{Item: "ITM0004", Qty: 2}, {Item: "ITM0218", Qty: 1}},
			Output:      Output{Item: "ITM0220", Qty: 1},
		},
		{
			Entity: &entities.Entity{ID: "RCP0002"}, Key: "healing_salve", Name: "Healing Salve",
			Description: "Crushed herbs worked into a soothing paste.",
			Category:    "food", Station: "",
			Ingredients: []Ingredient{{Item: "ITM0004", Qty: 3}},
			Output:      Output{Item: "ITM0221", Qty: 1},
		},
		{
			Entity: &entities.Entity{ID: "RCP0003"}, Key: "patchwork_jerkin", Name: "Patchwork Jerkin",
			Description: "Wolf pelts stitched into a rough chest piece.",
			Category:    "armor", Station: "",
			Ingredients: []Ingredient{{Item: "ITM0013", Qty: 2}},
			Output:      Output{Item: "ITM0222", Qty: 1},
		},
		{
			Entity: &entities.Entity{ID: "RCP0004"}, Key: "fur_cap", Name: "Fur-Lined Cap",
			Description: "A soft cap lined with wolf fur.",
			Category:    "armor", Station: "",
			Ingredients: []Ingredient{{Item: "ITM0013", Qty: 1}},
			Output:      Output{Item: "ITM0223", Qty: 1},
		},
		{
			Entity: &entities.Entity{ID: "RCP0005"}, Key: "tusk_spike", Name: "Tusk Spike",
			Description: "Boar tusks bound to a sturdy branch — crude but sharp.",
			Category:    "weapon", Station: "forge",
			Ingredients: []Ingredient{{Item: "ITM0014", Qty: 2}, {Item: "ITM0021", Qty: 1}},
			Output:      Output{Item: "ITM0224", Qty: 1},
		},
		{
			Entity: &entities.Entity{ID: "RCP0006"}, Key: "copper_shiv", Name: "Copper Shiv",
			Description: "A short blade hammered from copper ore.",
			Category:    "weapon", Station: "forge",
			Ingredients: []Ingredient{{Item: "ITM0219", Qty: 2}},
			Output:      Output{Item: "ITM0225", Qty: 1},
		},
	}
}
