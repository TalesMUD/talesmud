package commands_test

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/characters"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func TestCombatStatusClearsStaleCombatState(t *testing.T) {
	g, facade := newTradeTestGame(t)

	character, err := facade.CharactersService().Store(&characters.Character{
		Name:             "Stale Fighter",
		BelongsUser:      *traits.BelongsToUser("user-combat"),
		InCombat:         true,
		CombatInstanceID: "missing-combat-instance",
	})
	if err != nil {
		t.Fatalf("store character: %v", err)
	}

	msg := &messages.Message{
		FromUser:  &entities.User{Entity: &entities.Entity{ID: "user-combat"}},
		Character: character,
		Data:      "status",
	}
	if !(&commands.CombatStatusCommand{}).Execute(g, msg) {
		t.Fatal("status command did not handle message")
	}

	stored, err := facade.CharactersService().FindByID(character.ID)
	if err != nil {
		t.Fatalf("find character: %v", err)
	}
	if stored.InCombat || stored.CombatInstanceID != "" {
		t.Fatalf("expected stale combat state to be cleared, got InCombat=%v CombatInstanceID=%q", stored.InCombat, stored.CombatInstanceID)
	}

	var sawNotInCombat bool
	for _, out := range drainTradeMessages(g.SendMessage()) {
		if rsp, ok := out.(messages.MessageResponse); ok && strings.Contains(rsp.Message, "not in combat") {
			sawNotInCombat = true
		}
	}
	if !sawNotInCombat {
		t.Fatal("expected not-in-combat response")
	}
}
