package service

import (
	e "github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
)

// equipStartingItems copies each named starter onto the character.
// Each copy gets a new entity id and is not a template. The map key is the
// starting-item slot. Unknown names are skipped. An empty list leaves
// EquippedItems unchanged.
func equipStartingItems(ch *characters.Character, starting []characters.StartingItem) {
	if ch == nil || len(starting) == 0 {
		return
	}
	ch.EquippedItems = make(map[items.ItemSlot]*items.Item)
	for _, si := range starting {
		itemTemplate := items.StarterItemTemplateByName(si.ItemTemplateName)
		if itemTemplate == nil {
			continue
		}
		itemCopy := *itemTemplate
		itemCopy.Entity = e.NewEntity()
		itemCopy.IsTemplate = false
		if si.Slot != "" {
			ch.EquippedItems[si.Slot] = &itemCopy
		}
	}
}
