package commands

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/traits"
)

func TestFormatItemSubTypeHidesInternalIds(t *testing.T) {
	if got := formatItemSubType(items.ItemSubType("artifact_fragment")); got != "" {
		t.Fatalf("artifact_fragment should be hidden, got %q", got)
	}
	if got := formatItemSubType(items.ItemSubTypeSword); got != "Sword" {
		t.Fatalf("sword: got %q", got)
	}
}

func TestExamineItemOmitsInternalSubtype(t *testing.T) {
	item := &items.Item{
		Name:        "Tarnished Medal",
		Description: "A corroded medal bearing an unfamiliar crest.",
		LookAt:      traits.LookAt{Detail: "This medal was once precious."},
		Type:        items.ItemTypeQuest,
		SubType:     items.ItemSubType("artifact_fragment"),
		Quality:     items.ItemQualityNormal,
	}
	out := examineItem(item, nil)
	if strings.Contains(out, "artifact_fragment") {
		t.Fatalf("examine dump still contains internal id:\n%s", out)
	}
	if !strings.Contains(out, "Type: Quest Item\n") {
		t.Fatalf("expected clean Quest Item type line, got:\n%s", out)
	}
	if !strings.Contains(out, "=== Tarnished Medal ===") {
		t.Fatalf("expected header for client parser, got:\n%s", out)
	}
}

func TestExamineItemMarksUnique(t *testing.T) {
	item := &items.Item{
		Name:    "Relic",
		Type:    items.ItemTypeWeapon,
		Quality: items.ItemQualityRare,
		Unique:  true,
	}
	out := examineItem(item, nil)
	if !strings.Contains(out, "Mark: Unique\n") {
		t.Fatalf("missing unique mark:\n%s", out)
	}
}

func TestExamineItemListsEffects(t *testing.T) {
	item := &items.Item{Name: "Blade", Type: items.ItemTypeWeapon, OnHitScriptID: "SCR9",
		Effects: []items.ItemEffect{{Trigger: items.EffectTriggerOnHit, Name: "Vigil Burn", Text: "2 damage each turn for 3 turns."}}}
	out := examineItem(item, nil)
	if !strings.Contains(out, "--- Effects ---\nOn hit — Vigil Burn: 2 damage each turn for 3 turns.") {
		t.Fatalf("missing effect line:\n%s", out)
	}
	bare := &items.Item{Name: "Odd Knife", Type: items.ItemTypeWeapon, OnHitScriptID: "SCR9"}
	out = examineItem(bare, nil, func(string) string { return "Odd Proc" })
	if !strings.Contains(out, "On hit — Odd Proc: Special effect") {
		t.Fatalf("missing script fallback:\n%s", out)
	}
}

func TestBackfillItemEffects(t *testing.T) {
	tmpl := &items.Item{Effects: []items.ItemEffect{{Trigger: items.EffectTriggerOnHit, Text: "x"}}}
	carried := &items.Item{Name: "Old copy", TemplateID: "ITM1"}
	worn := &items.Item{Name: "Worn copy", TemplateID: "ITM1"}
	ch := &characters.Character{EquippedItems: map[items.ItemSlot]*items.Item{items.ItemSlotMainHand: worn}}
	ch.Inventory.Items = []*items.Item{carried, {Name: "Rock", TemplateID: "ITM2"}}
	lookup := func(id string) *items.Item {
		if id == "ITM1" {
			return tmpl
		}
		return nil
	}
	changed := backfillItemEffects(ch, lookup)
	if !changed || len(carried.Effects) != 1 || len(worn.Effects) != 1 {
		t.Fatalf("backfill failed: changed=%v carried=%+v worn=%+v", changed, carried.Effects, worn.Effects)
	}
	if backfillItemEffects(ch, lookup) {
		t.Fatal("second backfill should change nothing")
	}
}
