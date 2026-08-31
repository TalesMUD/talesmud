package items

import "testing"

func testArmor() *Item {
	return &Item{
		Name:       "Leather Jerkin",
		Type:       ItemTypeArmor,
		Slot:       ItemSlotChest,
		Attributes: map[string]interface{}{"defense": 8.0},
	}
}

func TestArmorTakesDamageWithoutDeletion(t *testing.T) {
	a := testArmor()
	if !a.DamageDurability(1) {
		t.Fatal("expected damage")
	}
	if a.Durability != 3 || a.ConditionLabel() != "worn" {
		t.Fatalf("durability=%d condition=%s", a.Durability, a.ConditionLabel())
	}
	if a.Name == "" {
		t.Fatal("item must not be deleted")
	}
}

func TestBrokenArmorGivesNoDefense(t *testing.T) {
	a := testArmor()
	a.DamageDurability(4)
	if a.ConditionLabel() != "broken" {
		t.Fatalf("condition %s", a.ConditionLabel())
	}
	if a.DefenseMultiplier() != 0 {
		t.Fatalf("multiplier %v", a.DefenseMultiplier())
	}
}

func TestRepairRestoresArmor(t *testing.T) {
	a := testArmor()
	a.DamageDurability(4)
	a.Repair()
	if a.Durability != a.MaxDurability || a.ConditionLabel() != "fine" {
		t.Fatalf("after repair dur=%d cond=%s", a.Durability, a.ConditionLabel())
	}
	if a.DefenseMultiplier() != 1 {
		t.Fatalf("multiplier %v", a.DefenseMultiplier())
	}
}
