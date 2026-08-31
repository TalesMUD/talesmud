package items

const defaultArmorMaxDurability int32 = 4

// IsArmorPiece is true for worn armor (not weapons).
func (item *Item) IsArmorPiece() bool {
	if item == nil {
		return false
	}
	return item.Type == ItemTypeArmor || item.Slot == ItemSlotHead || item.Slot == ItemSlotChest ||
		item.Slot == ItemSlotLegs || item.Slot == ItemSlotBoots || item.Slot == ItemSlotHands ||
		item.SubType == ItemSubTypeShield
}

// EnsureArmorDurability sets defaults on first wear/damage.
func (item *Item) EnsureArmorDurability() {
	if item == nil || !item.IsArmorPiece() {
		return
	}
	if item.MaxDurability <= 0 {
		item.MaxDurability = defaultArmorMaxDurability
	}
	if item.Durability == 0 && !item.brokenOnce() {
		item.Durability = item.MaxDurability
	}
}

func (item *Item) brokenOnce() bool {
	if item.Properties == nil {
		return false
	}
	v, ok := item.Properties["broken"]
	b, _ := v.(bool)
	return ok && b
}

// DamageDurability reduces remaining hits until repair. Does not delete the item.
func (item *Item) DamageDurability(amount int32) bool {
	if item == nil || !item.IsArmorPiece() || amount <= 0 {
		return false
	}
	item.EnsureArmorDurability()
	if item.Durability <= 0 {
		return false
	}
	item.Durability -= amount
	if item.Durability < 0 {
		item.Durability = 0
	}
	if item.Durability == 0 {
		if item.Properties == nil {
			item.Properties = map[string]interface{}{}
		}
		item.Properties["broken"] = true
	}
	return true
}

// Repair restores durability to max.
func (item *Item) Repair() {
	if item == nil || !item.IsArmorPiece() {
		return
	}
	item.EnsureArmorDurability()
	item.Durability = item.MaxDurability
	if item.Properties != nil {
		delete(item.Properties, "broken")
	}
}

// NeedsRepair is true when the piece is worn down.
func (item *Item) NeedsRepair() bool {
	if item == nil || !item.IsArmorPiece() {
		return false
	}
	item.EnsureArmorDurability()
	return item.Durability < item.MaxDurability
}

// DefenseMultiplier scales protection by remaining durability.
func (item *Item) DefenseMultiplier() float64 {
	if item == nil || !item.IsArmorPiece() {
		return 1
	}
	item.EnsureArmorDurability()
	if item.MaxDurability <= 0 || item.Durability <= 0 {
		return 0
	}
	return float64(item.Durability) / float64(item.MaxDurability)
}

// ConditionLabel is a short player-facing durability state.
func (item *Item) ConditionLabel() string {
	if item == nil || !item.IsArmorPiece() {
		return ""
	}
	item.EnsureArmorDurability()
	if item.Durability <= 0 {
		return "broken"
	}
	ratio := float64(item.Durability) / float64(item.MaxDurability)
	switch {
	case ratio >= 1:
		return "fine"
	case ratio >= 0.5:
		return "worn"
	default:
		return "damaged"
	}
}
