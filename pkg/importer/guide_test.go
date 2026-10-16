package importer

import "testing"

func TestYAMLQuestGuideCopies(t *testing.T) {
	on := (&YAMLQuest{ID: "q1", Name: "Show me", Guide: true}).ToEntity()
	if on == nil || !on.Guide {
		t.Fatalf("guide flag dropped: %+v", on)
	}
	off := (&YAMLQuest{ID: "q2", Name: "Quiet"}).ToEntity()
	if off == nil || off.Guide {
		t.Fatalf("empty guide became true: %+v", off)
	}
}
