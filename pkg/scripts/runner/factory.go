package runner

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/scripts"
	luarunner "github.com/talesmud/talesmud/pkg/scripts/runner/lua"
	"github.com/talesmud/talesmud/pkg/scripts/runner/lua/modules"
	"github.com/talesmud/talesmud/pkg/service"
)

func jsScriptsAllowed() bool {
	return os.Getenv("TALESMUD_ALLOW_JS_SCRIPTS") == "1"
}

func jsDisabledResult(script scripts.Script) *scripts.ScriptResult {
	logrus.WithField("script", script.Name).Warn("JavaScript scripts are disabled; set TALESMUD_ALLOW_JS_SCRIPTS=1 to enable")
	return &scripts.ScriptResult{
		Success: false,
		Error:   "javascript scripts are disabled",
	}
}

// MultiRunner is a ScriptRunner that delegates to the appropriate runner
// based on the script's language
type MultiRunner struct {
	jsRunner  *DefaultScriptRunner
	luaRunner *luarunner.LuaRunner
}

// NewMultiRunner creates a new multi-language script runner
func NewMultiRunner() *MultiRunner {
	luaRunner := luarunner.NewLuaRunner()

	// Register all Lua API modules
	modules.RegisterAllModules(luaRunner)

	return &MultiRunner{
		jsRunner:  NewDefaultScriptRunner(),
		luaRunner: luaRunner,
	}
}

// SetServices injects the required services into both runners
func (r *MultiRunner) SetServices(facade service.Facade, game def.GameCtrl) {
	r.jsRunner.SetServices(facade, game)
	r.luaRunner.SetServices(facade, game)
}

// GetLuaRunner returns the Lua runner for module registration
func (r *MultiRunner) GetLuaRunner() *luarunner.LuaRunner {
	return r.luaRunner
}

// Run executes a script with the given context, routing to the appropriate runner
func (r *MultiRunner) Run(script scripts.Script, ctx interface{}) interface{} {
	switch script.GetLanguage() {
	case scripts.ScriptLanguageLua:
		return r.luaRunner.Run(script, ctx)
	case scripts.ScriptLanguageJavaScript:
		if !jsScriptsAllowed() {
			return jsDisabledResult(script).Error
		}
		logrus.WithField("script", script.Name).Warn("JavaScript scripts are deprecated, please migrate to Lua")
		return r.jsRunner.Run(script, ctx)
	default:
		return r.luaRunner.Run(script, ctx)
	}
}

// RunWithResult executes a script and returns detailed result information
func (r *MultiRunner) RunWithResult(script scripts.Script, ctx *scripts.ScriptContext) *scripts.ScriptResult {
	switch script.GetLanguage() {
	case scripts.ScriptLanguageLua:
		return r.luaRunner.RunWithResult(script, ctx)
	case scripts.ScriptLanguageJavaScript:
		if !jsScriptsAllowed() {
			return jsDisabledResult(script)
		}
		logrus.WithField("script", script.Name).Warn("JavaScript scripts are deprecated, please migrate to Lua")
		return r.jsRunner.RunWithResult(script, ctx)
	default:
		return r.luaRunner.RunWithResult(script, ctx)
	}
}

// SupportsLanguage returns true if any runner supports the given language
func (r *MultiRunner) SupportsLanguage(lang scripts.ScriptLanguage) bool {
	return r.jsRunner.SupportsLanguage(lang) || r.luaRunner.SupportsLanguage(lang)
}

// Shutdown gracefully shuts down all runners
func (r *MultiRunner) Shutdown() {
	r.jsRunner.Shutdown()
	r.luaRunner.Shutdown()
}
