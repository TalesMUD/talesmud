package characters

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/items"
)

func TestDeathDamagesArmorAndRepairRestoresIt(t *testing.T) {
	jerkin := &items.Item{
		Entity:     &entities.Entity{ID: "armor-1"},
		Name:       "Leather Jerkin",
		Type:       items.ItemTypeArmor,
		Slot:       items.ItemSlotChest,
		Attributes: map[string]interface{}{"defense": float64(8)},
	}
	ch := &Character{
		Entity:        &entities.Entity{ID: "c1"},
		EquippedItems: map[items.ItemSlot]*items.Item{items.ItemSlotChest: jerkin},
	}
	full := ch.GetArmorDefense()
	if full != 8 {
		t.Fatalf("full defense %d", full)
	}
	names := ch.DamageEquippedArmor()
	if len(names) != 1 {
		t.Fatalf("damaged %v", names)
	}
	if ch.GetArmorDefense() >= full {
		t.Fatalf("expected reduced defense, got %d", ch.GetArmorDefense())
	}
	for i := 0; i < 5; i++ {
		ch.DamageEquippedArmor()
	}
	if ch.GetArmorDefense() != 0 {
		t.Fatalf("broken armor still defending %d", ch.GetArmorDefense())
	}
	if jerkin.Name == "" {
		t.Fatal("armor must not be deleted")
	}
	if ch.RepairEquippedArmor() < 1 {
		t.Fatal("expected repair")
	}
	if ch.GetArmorDefense() != 8 {
		t.Fatalf("repaired defense %d", ch.GetArmorDefense())
	}
}
