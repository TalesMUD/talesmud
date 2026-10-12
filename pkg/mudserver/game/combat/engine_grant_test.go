package combat

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
)

func TestCreateCombatantAddsScriptGrants(t *testing.T) {
	e := NewEngine(NewManager(), nil)
	char := &characters.Character{
		Entity:           &entities.Entity{ID: "grant"},
		Name:             "Grant",
		Class:            characters.ClassWarrior,
		Level:            1,
		MaxHitPoints:     25,
		CurrentHitPoints: 25,
		BonusAttack:      3,
		BonusDefense:     4,
		Attributes: characters.Attributes{
			characters.NewAttribute("Strength", "STR", 14),
		},
		EquippedItems: map[items.ItemSlot]*items.Item{
			items.ItemSlotMainHand: {
				Type:       items.ItemTypeWeapon,
				Slot:       items.ItemSlotMainHand,
				Attributes: map[string]interface{}{"damage": 2},
			},
			items.ItemSlotChest: {
				Type:       items.ItemTypeArmor,
				Slot:       items.ItemSlotChest,
				Attributes: map[string]interface{}{"defense": 1},
			},
		},
	}
	ref := e.CreateCombatantFromCharacter(char)
	// weapon 2 + STR mod 2 + grant 3
	if ref.AttackPower != 7 || ref.Defense != 5 {
		t.Fatalf("grant atk %d def %d", ref.AttackPower, ref.Defense)
	}

	char.BonusAttack = 0
	char.BonusDefense = 0
	ref = e.CreateCombatantFromCharacter(char)
	if ref.AttackPower != 4 || ref.Defense != 1 {
		t.Fatalf("zero grant atk %d def %d", ref.AttackPower, ref.Defense)
	}

	char.BonusAttack = -20
	char.BonusDefense = -20
	ref = e.CreateCombatantFromCharacter(char)
	if ref.AttackPower != 1 || ref.Defense != 0 {
		t.Fatalf("negative grant atk %d def %d", ref.AttackPower, ref.Defense)
	}
}
