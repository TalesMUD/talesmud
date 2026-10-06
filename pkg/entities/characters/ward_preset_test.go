package characters

import (
	"strings"
	"testing"
)

func TestWardPresetReplacesHitch(t *testing.T) {
	if PresetByID("tpl-hitch") != nil {
		t.Fatal("tpl-hitch is still a public preset")
	}
	ward := PresetByID("tpl-ward")
	if ward == nil {
		t.Fatal("missing tpl-ward")
	}
	if ward.Name != "Ward" || ward.Class.ID != "ward" || ward.Class.Name != "Ward" {
		t.Fatalf("identity %+v %+v", ward.Name, ward.Class)
	}
	const blurb = "Heavy plate. You start slow. Hits you take stack Grit, and Slam and the hit you throw back get heavier."
	if ward.Description != blurb || ward.Class.Description != blurb {
		t.Fatalf("blurb preset %q class %q", ward.Description, ward.Class.Description)
	}
	if strings.Contains(ward.Description, "Warrior") || strings.Contains(ward.Class.Name, "Warrior") {
		t.Fatal("warrior subtitle")
	}
	if ward.Backstory != "You let the first blows land. The plate holds, and the answer gets heavier." {
		t.Fatalf("backstory %q", ward.Backstory)
	}
	if ward.OriginArea != "Gatehouse" || ward.Archetype != "ward" {
		t.Fatalf("origin %s archetype %s", ward.OriginArea, ward.Archetype)
	}
	if ward.Class.ArmorType != ArmorTypePlate || ward.MaxHitPoints != 26 {
		t.Fatalf("plate/hp %s %d", ward.Class.ArmorType, ward.MaxHitPoints)
	}
	if len(ward.DefaultSkills) != 1 || ward.DefaultSkills[0] != "ward_guard" {
		t.Fatalf("skills %v", ward.DefaultSkills)
	}
	if len(ward.StartingItems) != 2 || ward.StartingItems[0].ItemTemplateName != "Rusty Sword" || ward.StartingItems[1].ItemTemplateName != "Leather Armor" {
		t.Fatalf("items %+v", ward.StartingItems)
	}
	for _, preset := range SystemCharacterTemplatePresets() {
		blob := strings.ToLower(preset.Name + " " + preset.Class.Name + " " + preset.Class.ID + " " + preset.Entity.ID)
		if strings.Contains(blob, "hitch") {
			t.Fatalf("public roster still says hitch: %s", blob)
		}
	}
}

func TestNormalizeClassRewritesHitch(t *testing.T) {
	ch := &Character{
		Level:          1,
		Class:          Class{ID: "hitch", Name: "Hitch", Description: "Fen rope."},
		EquippedSkills: []string{"hitch_pin"},
	}
	ch.NormalizeClass()
	if ch.Class.ID != "ward" || ch.Class.Name != "Ward" {
		t.Fatalf("%+v", ch.Class)
	}
	if strings.Contains(ch.Class.Description, "Hitch") || strings.Contains(ch.Class.Description, "Warrior") {
		t.Fatal(ch.Class.Description)
	}
	if len(ch.EquippedSkills) != 2 || ch.EquippedSkills[0] != "ward_guard" || ch.EquippedSkills[1] != "ward_slam" {
		t.Fatalf("level 1 bar %v", ch.EquippedSkills)
	}

	ch = &Character{
		Level:          8,
		Class:          Class{ID: "Hitch", Name: "Hitch"},
		EquippedSkills: []string{"hitch_pin", "hitch_reel"},
	}
	ch.NormalizeClass()
	if ch.Class.ID != "ward" || len(ch.EquippedSkills) != 2 || ch.EquippedSkills[0] != "ward_guard" || ch.EquippedSkills[1] != "ward_slam" {
		t.Fatalf("id %s bar %v", ch.Class.ID, ch.EquippedSkills)
	}
}
