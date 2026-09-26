package ruleset

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestShippedRulesetEqualsBuiltin(t *testing.T) {
	path := filepath.Join("..", "..", "config", "ruleset.yaml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	saved := current
	mu.Unlock()
	t.Cleanup(func() {
		mu.Lock()
		current = saved
		mu.Unlock()
	})

	if err := LoadBytes(raw); err != nil {
		t.Fatal(err)
	}
	mu.RLock()
	got := current
	mu.RUnlock()
	want := builtin()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("shipped ruleset drifted from builtin\n got %#v\nwant %#v", got, want)
	}
}
