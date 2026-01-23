package npc

import "time"

// MerchantTrait contains merchant-specific configuration for NPCs
type MerchantTrait struct {
	// MerchantType categorizes the merchant (e.g., "general", "blacksmith", "alchemist")
	MerchantType string `json:"merchantType"`

	// Inventory holds the items this merchant sells
	Inventory []MerchantItem `json:"inventory,omitempty"`

	// RestockMinutes is how often the merchant restocks inventory (0 = never)
	RestockMinutes int32 `json:"restockMinutes,omitempty"`

	// LastRestock tracks when inventory was last restocked
	LastRestock time.Time `json:"lastRestock,omitempty"`

	// BuyMultiplier adjusts buy prices (1.0 = normal, 1.2 = 20% more expensive)
	BuyMultiplier float64 `json:"buyMultiplier"`

	// SellMultiplier adjusts sell prices (1.0 = normal, 0.5 = sells at half price)
	SellMultiplier float64 `json:"sellMultiplier"`

	// AcceptedTypes limits what item types the merchant will buy (empty = accepts all)
	AcceptedTypes []string `json:"acceptedTypes,omitempty"`

	// RejectedTags prevents buying items with these tags (e.g., "soulbound", "quest")
	RejectedTags []string `json:"rejectedTags,omitempty"`
}

// MerchantItem represents an item in a merchant's inventory
type MerchantItem struct {
	// ItemTemplateID is the ID of the item template to sell
	ItemTemplateID string `json:"itemTemplateId"`

	// BasePrice is the default price (0 = use item's BasePrice)
	BasePrice int64 `json:"basePrice,omitempty"`

	// PriceOverride forces a specific price (ignores multipliers if > 0)
	PriceOverride int64 `json:"priceOverride,omitempty"`

	// Quantity is the current stock (-1 = unlimited)
	Quantity int32 `json:"quantity"`

	// MaxQuantity is the max stock after restock
	MaxQuantity int32 `json:"maxQuantity"`

	// RequiredLevel is the minimum player level to purchase
	RequiredLevel int32 `json:"requiredLevel,omitempty"`
}

// NewMerchantTrait creates a MerchantTrait with default values
func NewMerchantTrait() *MerchantTrait {
	return &MerchantTrait{
		MerchantType:   "general",
		Inventory:      make([]MerchantItem, 0),
		BuyMultiplier:  1.0,
		SellMultiplier: 0.5,
	}
}

// GetBuyPrice calculates the price a player pays to buy an item
func (mt *MerchantTrait) GetBuyPrice(item *MerchantItem, baseItemPrice int64) int64 {
	if item.PriceOverride > 0 {
		return item.PriceOverride
	}

	price := item.BasePrice
	if price == 0 {
		price = baseItemPrice
	}

	return int64(float64(price) * mt.BuyMultiplier)
}

// GetSellPrice calculates the price a player receives for selling an item
func (mt *MerchantTrait) GetSellPrice(baseItemPrice int64) int64 {
	return int64(float64(baseItemPrice) * mt.SellMultiplier)
}

// CanBuyItem checks if the merchant will accept an item for purchase
func (mt *MerchantTrait) CanBuyItem(itemType string, itemTags []string) bool {
	// Check accepted types
	if len(mt.AcceptedTypes) > 0 {
		accepted := false
		for _, t := range mt.AcceptedTypes {
			if t == itemType {
				accepted = true
				break
			}
		}
		if !accepted {
			return false
		}
	}

	// Check rejected tags
	for _, tag := range itemTags {
		for _, rejected := range mt.RejectedTags {
			if tag == rejected {
				return false
			}
		}
	}

	return true
}

// NeedsRestock checks if merchant inventory needs restocking
func (mt *MerchantTrait) NeedsRestock() bool {
	if mt.RestockMinutes <= 0 {
		return false
	}
	return time.Since(mt.LastRestock) >= time.Duration(mt.RestockMinutes)*time.Minute
}

// Restock refills merchant inventory to max quantities
func (mt *MerchantTrait) Restock() {
	for i := range mt.Inventory {
		if mt.Inventory[i].MaxQuantity > 0 {
			mt.Inventory[i].Quantity = mt.Inventory[i].MaxQuantity
		}
	}
	mt.LastRestock = time.Now()
}

// FindInventoryItem finds a merchant item by template ID
func (mt *MerchantTrait) FindInventoryItem(templateID string) *MerchantItem {
	for i := range mt.Inventory {
		if mt.Inventory[i].ItemTemplateID == templateID {
			return &mt.Inventory[i]
		}
	}
	return nil
}
