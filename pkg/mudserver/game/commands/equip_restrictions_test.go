package commands

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
)

func TestEquipRestriction(t *testing.T) {
	warrior := &characters.Character{Class: characters.Class{ID: "warrior", ArmorType: characters.ArmorTypePlate}, Level: 5}
	rogue := &characters.Character{Class: characters.Class{ID: "rogue", ArmorType: characters.ArmorTypeLeather}, Level: 5}
	cases := []struct {
		name      string
		item      items.Item
		character *characters.Character
		want      string
	}{
		{"class allowed", items.Item{Tags: []string{"class:warrior"}}, warrior, ""},
		{"class blocked", items.Item{Tags: []string{"class:warrior"}}, rogue, "Requires warrior class."},
		{"level blocked", items.Item{Level: 7}, warrior, "Requires level 7."},
		{"plate blocked", items.Item{Type: items.ItemTypeArmor, Properties: map[string]interface{}{"armorWeight": "plate"}}, rogue, "Requires plate armor training."},
		{"plate allowed", items.Item{Type: items.ItemTypeArmor, Tags: []string{"armor:plate"}}, warrior, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := equipRestriction(&tc.item, tc.character); got != tc.want {
				t.Fatalf("equipRestriction = %q, want %q", got, tc.want)
			}
		})
	}
}
