package contenthealth

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/importer"
	"gopkg.in/yaml.v3"
)

func TestRoomNoAggroYAMLRoundTrip(t *testing.T) {
	var authored importer.YAMLRoom
	if err := yaml.Unmarshal([]byte("id: R1\nname: Chapel\nnoAggro: true\n"), &authored); err != nil {
		t.Fatal(err)
	}
	room := authored.ToEntity()
	if room == nil || !room.NoAggro {
		t.Fatalf("import noAggro = %+v", room)
	}

	out, err := MarshalYAML("room", room)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "noAggro: true") {
		t.Fatalf("export dropped noAggro:\n%s", out)
	}
	var back importer.YAMLRoom
	if err := yaml.Unmarshal(out, &back); err != nil {
		t.Fatal(err)
	}
	if got := back.ToEntity(); got == nil || !got.NoAggro {
		t.Fatalf("reimport noAggro = %+v", got)
	}

	room.NoAggro = false
	quiet, err := MarshalYAML("room", room)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(quiet), "noAggro") {
		t.Fatalf("false noAggro should stay omitted:\n%s", quiet)
	}
}
