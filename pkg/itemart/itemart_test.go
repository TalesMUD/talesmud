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

func TestSkillGenericKeyFireball(t *testing.T) {
	if SkillGenericKey("mage_fireball", "Fireball") != "generic-spell-fire" {
		t.Fatal("fireball")
	}
}

func TestSkillGenericKeyHeal(t *testing.T) {
	if SkillGenericKey("cleric_heal", "") != "generic-spell-heal" {
		t.Fatal("heal")
	}
}

func TestSkillGenericURL(t *testing.T) {
	if SkillGenericURL("mage_fireball", "") != "/api/item-art/generic-spell-fire.png" {
		t.Fatalf("got %q", SkillGenericURL("mage_fireball", ""))
	}
}

func TestActionGenericKeyLook(t *testing.T) {
	if ActionGenericKey("look") != "generic-action-look" {
		t.Fatal("look")
	}
}
