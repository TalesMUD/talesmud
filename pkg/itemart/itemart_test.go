package itemart

import "testing"

func TestItemArtURLUsesTemplate(t *testing.T) {
	got := URL("torch-instance", "ITM0007")
	if got != "/api/item-art/ITM0007.png" {
		t.Fatalf("got %q", got)
	}
}

func TestGenericKeyTorchSubtype(t *testing.T) {
	if GenericKey("consumable", "torch") != "torch" {
		t.Fatal("expected torch")
	}
}

func TestGenericKeyWeapon(t *testing.T) {
	if GenericKey("weapon", "sword") != "weapon" {
		t.Fatal("expected weapon")
	}
}
