package npc

import "testing"

func TestTwoMerchantsSameTemplateDifferentPrices(t *testing.T) {
	cheap := &MerchantTrait{BuyMultiplier: 1.0, SellMultiplier: 0.8}
	pricey := &MerchantTrait{BuyMultiplier: 1.5, SellMultiplier: 0.3}
	stock := MerchantItem{ItemTemplateID: "ITM0012", Quantity: 5, MaxQuantity: 5}
	base := int64(8)
	buyCheap := cheap.GetBuyPrice(&stock, base)
	buyPricey := pricey.GetBuyPrice(&stock, base)
	if buyCheap == buyPricey {
		t.Fatalf("expected different buy prices, both %d", buyCheap)
	}
	if buyCheap != 8 || buyPricey != 12 {
		t.Fatalf("buy cheap=%d pricey=%d", buyCheap, buyPricey)
	}
	if cheap.GetSellPrice(base) == pricey.GetSellPrice(base) {
		t.Fatal("expected different sell prices")
	}
}

func TestMerchantPriceOverrideBeatsMultiplier(t *testing.T) {
	mt := &MerchantTrait{BuyMultiplier: 2.0}
	item := MerchantItem{ItemTemplateID: "ITM0012", PriceOverride: 5}
	if mt.GetBuyPrice(&item, 8) != 5 {
		t.Fatal("override ignored")
	}
}
