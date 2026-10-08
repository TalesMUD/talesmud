package lua

import (
	"strings"
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/scripts"
)

// TestTimedOutScriptDoesNotCrash locks the gopher-lua timeout race: a cancelled
// context panics inside the VM goroutine, and the runner must recover and wait
// for that goroutine before the state is returned to the pool.
func TestTimedOutScriptDoesNotCrash(t *testing.T) {
	runner := NewLuaRunner()
	defer runner.Shutdown()
	runner.sandbox.MaxExecutionTime = 40 * time.Millisecond

	hung := scripts.Script{
		Name:     "infinite",
		Language: scripts.ScriptLanguageLua,
		Code:     "while true do end",
	}
	first := runner.RunWithResult(hung, scripts.NewScriptContext())
	if first == nil {
		t.Fatal("nil result")
	}
	if first.Success {
		t.Fatal("infinite script succeeded")
	}
	if !timeoutError(first.Error) {
		t.Fatalf("expected a timeout error, got %q", first.Error)
	}
	if first.Duration > 3*time.Second {
		t.Fatalf("timeout took %s; goroutine did not unwind", first.Duration)
	}

	follow := scripts.Script{
		Name:     "follow",
		Language: scripts.ScriptLanguageLua,
		Code:     "return 40 + 2",
	}
	second := runner.RunWithResult(follow, scripts.NewScriptContext())
	if second == nil || !second.Success {
		err := ""
		if second != nil {
			err = second.Error
		}
		t.Fatalf("follow-up script failed: %s", err)
	}
	got, ok := second.Result.(float64)
	if !ok || got != 42 {
		t.Fatalf("follow-up result = %#v", second.Result)
	}
}

func timeoutError(err string) bool {
	e := strings.ToLower(err)
	return strings.Contains(e, "timeout") || strings.Contains(e, "deadline exceeded") || strings.Contains(e, "context canceled")
}
