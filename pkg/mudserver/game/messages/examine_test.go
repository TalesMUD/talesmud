package messages

import (
	"encoding/json"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities/items"
	"github.com/talesmud/talesmud/pkg/entities/traits"
)

func TestExamineResultJSON(t *testing.T) {
	item := &items.Item{Name: "Blade", Slot: items.ItemSlotMainHand, Quality: items.ItemQualityMagic,
		Attributes: map[string]interface{}{"damage": 3}, BasePrice: 12, Description: "A blue blade.",
		LookAt: traits.LookAt{Detail: "Runes glow."}, OnUseScriptID: "SCR1",
		Effects: []items.ItemEffect{{Trigger: items.EffectTriggerOnHit, Name: "Burn", Text: "2 damage."}}}
	msg := NewExamineResult("user", item, func(string) string { return "Glow" })
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"type":"examineResult","username":"","message":"","name":"Blade","slot":"main_hand","quality":"magic","stats":{"damage":3},"value":12,"effects":["On hit — Burn: 2 damage.","On use — Glow: Special effect"],"description":"A blue blade.","detail":"Runes glow."}`
	if string(data) != want {
		t.Fatalf("got %s\nwant %s", data, want)
	}
	if msg.GetAudienceID() != "user" {
		t.Fatal("wrong recipient")
	}
}

func TestExamineResultOmitsEmptyFields(t *testing.T) {
	data, err := json.Marshal(NewExamineResult("user", &items.Item{}, nil))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"type":"examineResult","username":"","message":""}` {
		t.Fatalf("empty fields: %s", data)
	}
	if NewExamineResult("user", nil, nil) != nil {
		t.Fatal("nil item should not produce a result")
	}
}
