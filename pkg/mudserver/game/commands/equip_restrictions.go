package commands

import (
	"fmt"
	"strings"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
)

// equipRestriction explains a requirement before the item is moved out of the bag.
func equipRestriction(item *items.Item, character *characters.Character) string {
	if item == nil || character == nil {
		return ""
	}
	classID := strings.ToLower(character.Class.ID)
	allowed := []string{}
	for _, tag := range item.Tags {
		if strings.HasPrefix(strings.ToLower(tag), "class:") {
			allowed = append(allowed, strings.TrimPrefix(strings.ToLower(tag), "class:"))
		}
	}
	if len(allowed) > 0 {
		matched := false
		for _, id := range allowed {
			if id == classID {
				matched = true
				break
			}
		}
		if !matched {
			return "Requires " + strings.Join(allowed, " or ") + " class."
		}
	}
	if item.Level > character.Level {
		return fmt.Sprintf("Requires level %d.", item.Level)
	}
	if item.Type != items.ItemTypeArmor {
		return ""
	}
	weight := ""
	if value, ok := item.Properties["armorWeight"].(string); ok {
		weight = strings.ToLower(value)
	}
	if weight == "" {
		for _, tag := range item.Tags {
			if strings.HasPrefix(strings.ToLower(tag), "armor:") {
				weight = strings.TrimPrefix(strings.ToLower(tag), "armor:")
				break
			}
		}
	}
	ranks := map[string]int{"cloth": 1, "leather": 2, "plate": 3}
	trained := strings.ToLower(string(character.Class.ArmorType))
	if ranks[weight] > 0 && ranks[trained] > 0 && ranks[weight] > ranks[trained] {
		return "Requires " + weight + " armor training."
	}
	return ""
}
