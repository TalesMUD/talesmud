package commands

import (
	"strings"
	"testing"

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
