package commands_test

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/mudserver/game/commands"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
)

func TestExamineKeepsProseAndAddsResult(t *testing.T) {
	for _, verb := range []string{"examine", "inspect"} {
		t.Run(verb, func(t *testing.T) {
			g, facade := newSocialTestGame(t)
			user, ch := storeSocialPlayer(t, facade, "u", "r", "c", "Tester", "")
			ch.Inventory.Items = []*items.Item{{Name: "Stone", Type: items.ItemTypeCollectible, Description: "A smooth stone."}}
			(&commands.ExamineCommand{}).Execute(g, &messages.Message{FromUser: user, Character: ch, Data: verb + " Stone"})
			out := drainSocialMessages(g.SendMessage())
			if len(out) != 2 {
				t.Fatalf("got %d messages", len(out))
			}
			prose, ok := out[0].(messages.MessageResponse)
			want := "=== Stone ===\n\nA smooth stone.\n\n--- Item Details ---\nType: Collectible\n"
			if !ok || prose.Message != want {
				t.Fatalf("prose changed: %#v", out[0])
			}
			result, ok := out[1].(*messages.ExamineResult)
			if !ok || result.Name != "Stone" || result.GetAudienceID() != user.ID {
				t.Fatalf("missing result: %#v", out[1])
			}
			(&commands.ExamineCommand{}).Execute(g, &messages.Message{FromUser: user, Character: ch, Data: verb + " Missing"})
			out = drainSocialMessages(g.SendMessage())
			if len(out) != 1 || socialReplyText(out[0]) != "You don't see a 'Missing' here." {
				t.Fatalf("missing-item prose changed: %#v", out)
			}
		})
	}
}
