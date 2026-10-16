package commands

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/ruleset"
)

type hintSink struct {
	def.GameCtrl
	sent chan interface{}
}

func (h *hintSink) SendMessage() chan interface{} { return h.sent }

func TestGuideDisabledSendsNothing(t *testing.T) {
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)
	spy := &hintSink{sent: make(chan interface{}, 1)}

	PushGuideHintAt(spy, "user-1", nil, "")
	select {
	case <-spy.sent:
	default:
		t.Fatal("default guide sent no message")
	}

	if err := ruleset.LoadBytes([]byte("guide:\n  enabled: false\n")); err != nil {
		t.Fatal(err)
	}
	PushGuideHintAt(spy, "user-1", nil, "")
	select {
	case msg := <-spy.sent:
		t.Fatalf("disabled guide sent %#v", msg)
	default:
	}
}
