package messages

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/mudserver/game/leveling"
)

func TestNewLevelUpMessageStructured(t *testing.T) {
	result := &leveling.LevelUpResult{
		OldLevel:               1,
		NewLevel:               2,
		LevelsGained:           1,
		HPGained:               15,
		ManaGained:             0,
		AttributeGains:         map[string]int32{"STR": 1},
		AttributePointsGained:  2,
		UnspentAttributePoints: 2,
		MaxHitPoints:           40,
		MaxMana:                0,
		Message:                "LEVEL UP!",
	}
	msg := NewLevelUpMessage("user-1", result)
	if msg == nil {
		t.Fatal("expected LevelUpMessage")
	}
	if msg.Type != MessageTypeLevelUp {
		t.Fatalf("type=%q", msg.Type)
	}
	if msg.AudienceID != "user-1" || msg.NewLevel != 2 || msg.HPGained != 15 {
		t.Fatalf("unexpected fields: %+v", msg)
	}
	if msg.AttributeGains["STR"] != 1 || msg.UnspentAttributePoints != 2 {
		t.Fatalf("gains/unspent wrong: %+v", msg)
	}
	if msg.Message != "LEVEL UP!" {
		t.Fatalf("message fallback missing")
	}
}

func TestNewLevelUpMessageNil(t *testing.T) {
	if NewLevelUpMessage("", &leveling.LevelUpResult{LevelsGained: 1}) != nil {
		t.Fatal("empty user should return nil")
	}
	if NewLevelUpMessage("u", nil) != nil {
		t.Fatal("nil result should return nil")
	}
	if NewLevelUpMessage("u", &leveling.LevelUpResult{LevelsGained: 0}) != nil {
		t.Fatal("zero levels should return nil")
	}
}
