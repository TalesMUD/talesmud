package characters

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/items"
)

func TestPrepareUniquePickupTrimsThenBlocks(t *testing.T) {
	c := &Character{Inventory: items.Inventory{Items: []*items.Item{
		{TemplateID: "ITM0020", Name: "Sample Relic Shard"},
		{TemplateID: "ITM0020", Name: "Sample Relic Shard"},
		{TemplateID: "ITM0020", Name: "Sample Relic Shard"},
	}}}
	blocked, trimmed := c.PrepareUniquePickup("ITM0020", true)
	if !blocked || trimmed != 2 {
		t.Fatalf("blocked=%v trimmed=%d", blocked, trimmed)
	}
	if c.Inventory.CountMatchingTemplate("ITM0020") != 1 {
		t.Fatalf("count %d", c.Inventory.CountMatchingTemplate("ITM0020"))
	}
	blocked, trimmed = c.PrepareUniquePickup("ITM0020", true)
	if !blocked || trimmed != 0 {
		t.Fatalf("second blocked=%v trimmed=%d", blocked, trimmed)
	}
	blocked, trimmed = c.PrepareUniquePickup("ITM0012", false)
	if blocked || trimmed != 0 {
		t.Fatalf("non-unique blocked=%v trimmed=%d", blocked, trimmed)
	}
}
