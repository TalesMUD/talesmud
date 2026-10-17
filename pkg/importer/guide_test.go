package importer

import "testing"

func TestYAMLQuestGuideCopies(t *testing.T) {
	on := (&YAMLQuest{ID: "q1", Name: "Show me", Guide: true}).ToEntity()
	if on == nil || !on.Guide {
		t.Fatalf("guide flag dropped: %+v", on)
	}
	off := (&YAMLQuest{ID: "q2", Name: "Quiet"}).ToEntity()
	daily := (&YAMLQuest{ID: "q3", Name: "Board", Repeatable: true, RepeatReset: "daily"}).ToEntity()
	if !daily.Repeatable || daily.RepeatReset != "daily" || daily.RepeatKind() != "daily" {
		t.Fatalf("daily import %+v", daily)
	}
	cool := (&YAMLQuest{ID: "q4", Name: "Later", Repeatable: true, RepeatReset: "cooldown", RepeatCooldownSeconds: 90}).ToEntity()
	if cool.RepeatCooldownSeconds != 90 || cool.RepeatKind() != "cooldown" {
		t.Fatalf("cooldown import %+v", cool)
	}
	if off == nil || off.Guide {
		t.Fatalf("empty guide became true: %+v", off)
	}
}
