package items

import (
	"errors"
	"strings"
)

//Inventory data
type Inventory struct {
	Size  int32   `json:"size"`
	Items []*Item `json:"items"`
}

// Count returns the number of items in the inventory
func (inv *Inventory) Count() int {
	return len(inv.Items)
}

// IsFull returns true if the inventory is at capacity
func (inv *Inventory) IsFull() bool {
	return inv.Size > 0 && int32(len(inv.Items)) >= inv.Size
}

// AddItem adds an item to the inventory with stacking logic
func (inv *Inventory) AddItem(item *Item) error {
	if item == nil {
		return errors.New("cannot add nil item")
	}

	// Check if the item is stackable and we can find an existing stack
	if item.Stackable {
		for _, existing := range inv.Items {
			// Stack by TemplateID if both have one, otherwise stack by Name
			sameTemplate := existing.TemplateID != "" && existing.TemplateID == item.TemplateID
			sameName := existing.TemplateID == "" && item.TemplateID == "" && existing.Name == item.Name
			if (sameTemplate || sameName) && existing.Stackable {
				// Check max stack limit
				if existing.MaxStack > 0 && existing.Quantity+item.Quantity > existing.MaxStack {
					// Can't fully stack, but add what we can
					spaceAvailable := existing.MaxStack - existing.Quantity
					if spaceAvailable > 0 {
						existing.Quantity += spaceAvailable
						item.Quantity -= spaceAvailable
						// Continue to add remainder as new stack
					}
				} else {
					// Can fully stack
					if item.Quantity == 0 {
						item.Quantity = 1
					}
					existing.Quantity += item.Quantity
					return nil
				}
			}
		}
	}

	// Check capacity for non-stackable or new stack
	if inv.IsFull() {
		return errors.New("inventory is full")
	}

	// Ensure quantity is at least 1
	if item.Quantity == 0 {
		item.Quantity = 1
	}

	inv.Items = append(inv.Items, item)
	return nil
}

// RemoveItem removes an item by ID and returns it
func (inv *Inventory) RemoveItem(itemID string) (*Item, error) {
	for i, item := range inv.Items {
		if item.ID == itemID {
			inv.Items = append(inv.Items[:i], inv.Items[i+1:]...)
			return item, nil
		}
	}
	return nil, errors.New("item not found in inventory")
}

// RemoveItemByName removes the first item matching the name and returns it
func (inv *Inventory) RemoveItemByName(name string) (*Item, error) {
	item := inv.FindItemByName(name)
	if item == nil {
		return nil, errors.New("item not found in inventory")
	}
	return inv.RemoveItem(item.ID)
}

// FindItemByName finds an item by partial name match (case-insensitive)
func (inv *Inventory) FindItemByName(name string) *Item {
	nameLower := strings.ToLower(name)

	// First try exact match
	for _, item := range inv.Items {
		if strings.ToLower(item.Name) == nameLower {
			return item
		}
	}

	// Then try prefix match
	for _, item := range inv.Items {
		if strings.HasPrefix(strings.ToLower(item.Name), nameLower) {
			return item
		}
	}

	// Then try contains match
	for _, item := range inv.Items {
		if strings.Contains(strings.ToLower(item.Name), nameLower) {
			return item
		}
	}

	return nil
}

// FindItemByTargetName finds an item by its target name (name-suffix format)
func (inv *Inventory) FindItemByTargetName(targetName string) *Item {
	targetLower := strings.ToLower(targetName)

	for _, item := range inv.Items {
		if strings.ToLower(item.GetTargetName()) == targetLower {
			return item
		}
	}

	// Also check by ID suffix if targetName looks like one
	for _, item := range inv.Items {
		if item.InstanceSuffix != "" && strings.HasSuffix(strings.ToLower(item.ID), targetLower) {
			return item
		}
	}

	return nil
}

// FindItemByID finds an item by its ID
func (inv *Inventory) FindItemByID(itemID string) *Item {
	for _, item := range inv.Items {
		if item.ID == itemID {
			return item
		}
	}
	return nil
}

// GetItemsByType returns all items of a specific type
func (inv *Inventory) GetItemsByType(itemType ItemType) []*Item {
	var result []*Item
	for _, item := range inv.Items {
		if item.Type == itemType {
			result = append(result, item)
		}
	}
	return result
}
