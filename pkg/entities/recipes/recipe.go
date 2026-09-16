package recipes

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/talesmud/talesmud/pkg/entities"
)

// Ingredient is one required material for a recipe.
type Ingredient struct {
	Item string `bson:"item" json:"item" yaml:"item"`
	Qty  int32  `bson:"qty" json:"qty" yaml:"qty"`
}

// Output is what a recipe produces.
type Output struct {
	Item string `bson:"item" json:"item" yaml:"item"`
	Qty  int32  `bson:"qty" json:"qty" yaml:"qty"`
}

// Recipe is a crafting recipe. Everyone can see and use all recipes (no profession gate).
type Recipe struct {
	*entities.Entity `bson:",inline"`

	Key         string       `bson:"key" json:"key" yaml:"key"`
	Name        string       `bson:"name" json:"name" yaml:"name"`
	Description string       `bson:"description" json:"description" yaml:"description"`
	Category    string       `bson:"category" json:"category" yaml:"category"` // food, armor, weapon, other
	Station     string       `bson:"station" json:"station" yaml:"station"`    // "", campfire, forge
	Ingredients []Ingredient `bson:"ingredients" json:"ingredients" yaml:"ingredients"`
	Output      Output       `bson:"output" json:"output" yaml:"output"`
}

// yamlRecipe is the on-disk shape (avoids Entity embedding quirks).
type yamlRecipe struct {
	ID          string       `yaml:"id"`
	Key         string       `yaml:"key"`
	Name        string       `yaml:"name"`
	Description string       `yaml:"description"`
	Category    string       `yaml:"category"`
	Station     string       `yaml:"station"`
	Ingredients []Ingredient `yaml:"ingredients"`
	Output      Output       `yaml:"output"`
}

var (
	cache   []*Recipe
	cacheMu sync.RWMutex
	loaded  bool
)

// All returns every loaded recipe.
func All() []*Recipe {
	EnsureLoaded()
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	out := make([]*Recipe, len(cache))
	copy(out, cache)
	return out
}

// Find looks up a recipe by id, key, or partial name (case-insensitive).
func Find(query string) *Recipe {
	EnsureLoaded()
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}
	cacheMu.RLock()
	defer cacheMu.RUnlock()

	for _, r := range cache {
		if strings.EqualFold(r.ID, q) || strings.EqualFold(r.Key, q) || strings.EqualFold(r.Name, q) {
			return r
		}
	}
	for _, r := range cache {
		if strings.HasPrefix(strings.ToLower(r.Key), q) || strings.HasPrefix(strings.ToLower(r.Name), q) {
			return r
		}
	}
	for _, r := range cache {
		if strings.Contains(strings.ToLower(r.Key), q) || strings.Contains(strings.ToLower(r.Name), q) {
			return r
		}
	}
	return nil
}

// SetCache replaces the in-memory recipe list (tests / explicit load).
func SetCache(list []*Recipe) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cache = list
	loaded = true
}

// ResetForTest clears the loaded flag (unit tests only).
func ResetForTest() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cache = nil
	loaded = false
}

// EnsureLoaded loads recipes from disk once, falling back to seed data.
func EnsureLoaded() {
	cacheMu.RLock()
	ok := loaded
	cacheMu.RUnlock()
	if ok {
		return
	}
	cacheMu.Lock()
	defer cacheMu.Unlock()
	if loaded {
		return
	}
	paths := candidateDirs()
	var loadedList []*Recipe
	for _, dir := range paths {
		list, err := LoadDir(dir)
		if err != nil {
			log.WithError(err).WithField("dir", dir).Debug("recipes: skip load dir")
			continue
		}
		if len(list) > 0 {
			loadedList = list
			log.WithField("dir", dir).WithField("count", len(list)).Info("recipes: loaded from directory")
			break
		}
	}
	if len(loadedList) == 0 {
		loadedList = SeedRecipes()
		log.WithField("count", len(loadedList)).Info("recipes: using seed recipes")
	}
	cache = loadedList
	loaded = true
}

func candidateDirs() []string {
	var dirs []string
	if p := strings.TrimSpace(os.Getenv("RECIPES_PATH")); p != "" {
		dirs = append(dirs, p)
	}
	dirs = append(dirs,
		filepath.Join("import", "mvp-rpg-1", "data", "recipes"),
		filepath.Join("data", "recipes"),
	)
	return dirs
}

// LoadDir reads all *.yaml recipe files from a directory.
func LoadDir(dir string) ([]*Recipe, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []*Recipe
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		var y yamlRecipe
		if err := yaml.Unmarshal(raw, &y); err != nil {
			return nil, err
		}
		r := &Recipe{
			Entity:      &entities.Entity{ID: y.ID},
			Key:         y.Key,
			Name:        y.Name,
			Description: y.Description,
			Category:    y.Category,
			Station:     y.Station,
			Ingredients: y.Ingredients,
			Output:      y.Output,
		}
		if r.Key == "" {
			r.Key = strings.ToLower(strings.ReplaceAll(r.Name, " ", "_"))
		}
		if r.Output.Qty < 1 {
			r.Output.Qty = 1
		}
		for i := range r.Ingredients {
			if r.Ingredients[i].Qty < 1 {
				r.Ingredients[i].Qty = 1
			}
		}
		if r.ID == "" || r.Name == "" || r.Output.Item == "" || len(r.Ingredients) == 0 {
			log.WithField("file", name).Warn("recipes: skipping incomplete recipe")
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

// StationOK returns true if the room tags satisfy the recipe station requirement.
func StationOK(recipe *Recipe, roomTags []string) bool {
	if recipe == nil || recipe.Station == "" {
		return true
	}
	need := strings.ToLower(recipe.Station)
	for _, t := range roomTags {
		t = strings.ToLower(t)
		switch need {
		case "forge":
			if t == "forge" || t == "crafting" {
				return true
			}
		case "campfire":
			if t == "campfire" || t == "kitchen" || t == "hearth" {
				return true
			}
		default:
			if t == need {
				return true
			}
		}
	}
	return false
}

// StationHint explains where to craft.
func StationHint(station string) string {
	switch strings.ToLower(station) {
	case "forge":
		return "Requires a forge (Ironhand's Forge, Great Forge, etc.)."
	case "campfire":
		return "Requires a campfire or kitchen."
	default:
		return ""
	}
}
