package messages

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestHelloJSON(t *testing.T) {
	for proto, status := range map[int]string{0: "too-old", 1: "ok", 2: "upgrade-recommended"} {
		msg := NewHelloMessage("user", proto)
		data, err := json.Marshal(msg)
		if err != nil {
			t.Fatal(err)
		}
		var wire map[string]interface{}
		if err := json.Unmarshal(data, &wire); err != nil {
			t.Fatal(err)
		}
		if wire["type"] != "hello" || wire["proto"] != float64(1) || wire["status"] != status || !strings.HasPrefix(wire["server"].(string), "talesmud/") || wire["message"] != "" {
			t.Fatalf("hello: %s", data)
		}
		if msg.GetAudienceID() != "user" {
			t.Fatal("wrong recipient")
		}
	}
	for _, msg := range []interface{}{HelloMessage{}, IncomingMessage{}} {
		data, err := json.Marshal(msg)
		if err != nil {
			t.Fatal(err)
		}
		var wire map[string]interface{}
		if err := json.Unmarshal(data, &wire); err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"proto", "server", "status", "client"} {
			if _, ok := wire[key]; ok {
				t.Fatalf("empty field %s: %s", key, data)
			}
		}
	}
	var incoming IncomingMessage
	if err := json.Unmarshal([]byte(`{"type":"hello","proto":1,"client":"terminal/1.0","message":""}`), &incoming); err != nil {
		t.Fatal(err)
	}
	if incoming.Proto != 1 || incoming.Client != "terminal/1.0" || incoming.Message != "" {
		t.Fatalf("incoming: %+v", incoming)
	}
}
