package recipes

import "github.com/talesmud/talesmud/pkg/entities"

// SeedRecipes returns a tiny sample crafting list used when no recipe YAML
// is present. Item ids are placeholders, not a content-pack catalog.
// Packs should ship data/recipes (or set RECIPES_PATH).
func SeedRecipes() []*Recipe {
	return []*Recipe{
		{
			Entity:      &entities.Entity{ID: "RCP_SAMPLE_STEW"},
			Key:         "sample_stew",
			Name:        "Sample Stew",
			Description: "A sample recipe. Content packs replace this list.",
			Category:    "food",
			Station:     "",
			Ingredients: []Ingredient{{Item: "sample-herb", Qty: 2}},
			Output:      Output{Item: "sample-stew", Qty: 1},
		},
		{
			Entity:      &entities.Entity{ID: "RCP_SAMPLE_BLADE"},
			Key:         "sample_blade",
			Name:        "Sample Blade",
			Description: "A sample forge recipe with placeholder item ids.",
			Category:    "weapon",
			Station:     "forge",
			Ingredients: []Ingredient{{Item: "sample-ore", Qty: 2}},
			Output:      Output{Item: "sample-blade", Qty: 1},
		},
	}
}
