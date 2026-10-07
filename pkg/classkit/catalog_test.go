package classkit

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/items"
)

func TestPackLoadedClasses(t *testing.T) {
	if !FromPack() {
		t.Fatal("testdata pack was not loaded")
	}
	want := []struct {
		id    string
		name  string
		blurb string
		races string
		skill string
		dealt float64
	}{
		{"warrior", "Sentinel", "The door.", "human,dwarf", "warrior_brace", 1},
		{"rogue", "Cutpurse", "Back street.", "human,dwarf,elf", "rogue_slip", 0.55},
		{"wizard", "Runecaster", "Vault runes.", "human,elf", "mage_inscribe", 1.40},
		{"ward", "Ward", "Heavy plate.", "human,dwarf", "ward_guard", 0.95},
		{"rigger", "Rigger", "Constructs.", "construct", "rigger_bolt", 0.85},
	}
	playable := Playable()
	if len(playable) != len(want) {
		t.Fatalf("playable %d", len(playable))
	}
	for i, w := range want {
		d := playable[i]
		if d.ID != w.id || d.Name != w.name || !strings.Contains(d.Description, w.blurb) {
			t.Fatalf("class %+v", d)
		}
		if strings.Join(d.Races, ",") != w.races {
			t.Fatalf("%s races %v", w.id, d.Races)
		}
		if len(d.Skills) == 0 || d.Skills[0].ID != w.skill {
			t.Fatalf("%s skill %+v", w.id, d.Skills)
		}
		if d.Balance.DamageDealt != w.dealt {
			t.Fatalf("%s dealt %v", w.id, d.Balance.DamageDealt)
		}
		requireStartingItems(t, d)
	}

	hitch := Lookup("hitch")
	if hitch == nil || hitch.ID != "ward" || hitch.Name != "Ward" {
		t.Fatalf("hitch %+v", hitch)
	}
	if SkillClass("wizard") != "mage" || SkillClass("mage") != "mage" {
		t.Fatalf("skill class wizard=%s mage=%s", SkillClass("wizard"), SkillClass("mage"))
	}
	if Lookup("ranger") != nil {
		t.Fatal("ranger must not be a kit class")
	}
	if row, ok := Balance("ranger"); !ok || row.Swings != 2 {
		t.Fatalf("ranger balance %+v %v", row, ok)
	}
	if HotbarCap("ranger") != 0 || HotbarCap("warrior") != 4 {
		t.Fatalf("caps ranger=%d warrior=%d", HotbarCap("ranger"), HotbarCap("warrior"))
	}
	view := ClientView(nil)
	if view.Source != "pack" || len(view.Classes) != 5 {
		t.Fatalf("view %+v", view.Source)
	}
	if !strings.Contains(RaceBlurbOverrides()["construct"], "Poison never sticks") {
		t.Fatal("construct blurb override missing")
	}
}

func TestNoPackFallback(t *testing.T) {
	UseDefaults()
	t.Cleanup(func() {
		if dir := findUp("pkg/classkit/testdata/classes"); dir != "" {
			if _, err := LoadDir(dir); err != nil {
				t.Errorf("restore pack: %v", err)
			}
		}
	})
	if FromPack() {
		t.Fatal("defaults are not a pack")
	}
	playable := Playable()
	if len(playable) != 3 {
		t.Fatalf("sample count %d", len(playable))
	}
	if playable[0].Name != "Warrior" || playable[1].Name != "Rogue" || playable[2].Name != "Mage" {
		t.Fatalf("names %s %s %s", playable[0].Name, playable[1].Name, playable[2].Name)
	}
	if playable[2].ID != "wizard" || playable[2].skillClass() != "mage" || !playable[2].InscribeBasic {
		t.Fatalf("sample mage %+v", playable[2])
	}
	if Lookup("hitch") != nil || Lookup("rigger") != nil {
		t.Fatal("sample catalog invented a pack class")
	}
	for _, d := range playable {
		blob := d.Name + " " + d.Description
		if strings.Contains(blob, "Sentinel") || strings.Contains(blob, "Cutpurse") || strings.Contains(blob, "Runecaster") || strings.Contains(blob, "Rigger") {
			t.Fatalf("sample leaked a pack name: %s", blob)
		}
		requireStartingItems(t, d)
	}
}

func requireStartingItems(t *testing.T, d *Def) {
	t.Helper()
	if d == nil || d.Template == nil || len(d.Template.Items) == 0 {
		id := ""
		if d != nil {
			id = d.ID
		}
		t.Fatalf("%s has no starting items", id)
	}
	for _, it := range d.Template.Items {
		if it.Slot == "" || items.StarterItemTemplateByName(it.Name) == nil {
			t.Fatalf("%s starter %+v", d.ID, it)
		}
	}
}
