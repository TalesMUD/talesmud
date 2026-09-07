package runner

import (
	"testing"

	"github.com/talesmud/talesmud/pkg/scripts"
)

func TestJavaScriptScriptsDisabledByDefault(t *testing.T) {
	t.Setenv("TALESMUD_ALLOW_JS_SCRIPTS", "")
	r := NewMultiRunner()
	script := scripts.Script{Name: "legacy", Language: scripts.ScriptLanguageJavaScript, Code: "1+1"}
	result := r.RunWithResult(script, scripts.NewScriptContext())
	if result.Success {
		t.Fatal("expected JS script to fail when disabled")
	}
	if result.Error != "javascript scripts are disabled" {
		t.Fatalf("got error %q", result.Error)
	}
}

func TestJavaScriptScriptsEnabledWithEnv(t *testing.T) {
	t.Setenv("TALESMUD_ALLOW_JS_SCRIPTS", "1")
	r := NewMultiRunner()
	script := scripts.Script{Name: "legacy", Language: scripts.ScriptLanguageJavaScript, Code: "1+1"}
	result := r.RunWithResult(script, scripts.NewScriptContext())
	if !result.Success {
		t.Fatalf("expected JS script to run, got %q", result.Error)
	}
}
