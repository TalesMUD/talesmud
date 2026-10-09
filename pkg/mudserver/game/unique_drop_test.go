package game

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func TestUniqueDropMessageIsRoomChip(t *testing.T) {
	msg := uniqueDropMessage("R1", "Warden", "Relic")
	if msg.Type != messages.MessageTypeDefault || msg.Username != "SYSTEM" {
		t.Fatalf("type %s user %s", msg.Type, msg.Username)
	}
	if msg.Audience != messages.MessageAudienceRoom || msg.AudienceID != "R1" {
		t.Fatalf("audience %v %s", msg.Audience, msg.AudienceID)
	}
	if msg.Style != "combatEvent" || msg.Hook != "unique" || msg.Source != "Warden" {
		t.Fatalf("chip %+v", msg)
	}
	if msg.Message != "UNIQUE: Relic" {
		t.Fatalf("message %q", msg.Message)
	}
}

func TestLootRevealMarksUnique(t *testing.T) {
	got := lootReveals([]*items.Item{{
		Name:    "Relic",
		Quality: items.ItemQualityRare,
		Unique:  true,
	}})
	if len(got) != 1 || !got[0].Unique || got[0].Quality != "rare" || got[0].Name != "Relic" {
		t.Fatalf("%+v", got)
	}
}
