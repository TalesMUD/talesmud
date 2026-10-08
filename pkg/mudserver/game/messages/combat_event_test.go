package messages

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMessageResponseOmitsCombatEventWhenEmpty(t *testing.T) {
	raw, err := json.Marshal(MessageResponse{
		Type:     MessageTypeDefault,
		Username: "SYSTEM",
		Message:  "the bricks drip",
	})
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	for _, key := range []string{`"style"`, `"hook"`, `"source"`} {
		if strings.Contains(text, key) {
			t.Fatalf("empty event marshaled %s in %s", key, text)
		}
	}
}

func TestMessageResponseCarriesCombatEventFields(t *testing.T) {
	raw, err := json.Marshal(MessageResponse{
		Type:     MessageTypeDefault,
		Username: "SYSTEM",
		Message:  "The chanter hits the drum once.",
		Style:    "combatEvent",
		Hook:     "onAggro",
		Source:   "The Barrow Warden",
	})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got["type"] != "message" || got["username"] != "SYSTEM" {
		t.Fatalf("wire type/username = %v %v", got["type"], got["username"])
	}
	if got["style"] != "combatEvent" || got["hook"] != "onAggro" || got["source"] != "The Barrow Warden" {
		t.Fatalf("event fields = %s", raw)
	}
}
