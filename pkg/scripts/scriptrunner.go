package scripts

import "time"

// ScriptResult contains the result of a script execution
type ScriptResult struct {
	Success  bool          `json:"success"`
	Result   interface{}   `json:"result,omitempty"`
	Error    string        `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
}

// CombatHook is the enemy-hook stamp for room lines sent during one script run.
// It is not copied into the Lua ctx table.
type CombatHook struct {
	Hook   string
	Source string
}

// ScriptContext provides context data for script execution
type ScriptContext struct {
	Data       map[string]interface{}
	combatHook *CombatHook
}

// NewScriptContext creates a new script context
func NewScriptContext() *ScriptContext {
	return &ScriptContext{
		Data: make(map[string]interface{}),
	}
}

// Set adds a value to the context
func (c *ScriptContext) Set(key string, value interface{}) *ScriptContext {
	c.Data[key] = value
	return c
}

// Get retrieves a value from the context
func (c *ScriptContext) Get(key string) (interface{}, bool) {
	val, ok := c.Data[key]
	return val, ok
}

// SetCombatHook marks room lines from this run as an enemy-hook combat event.
func (c *ScriptContext) SetCombatHook(hook, source string) {
	if c == nil || hook == "" {
		return
	}
	c.combatHook = &CombatHook{Hook: hook, Source: source}
}

// CombatHook returns the enemy-hook stamp for this run.
func (c *ScriptContext) CombatHook() (CombatHook, bool) {
	if c == nil || c.combatHook == nil || c.combatHook.Hook == "" {
		return CombatHook{}, false
	}
	return *c.combatHook, true
}

// ScriptRunner defines the interface for script execution engines
type ScriptRunner interface {
	// Run executes a script with the given context
	// For backward compatibility, ctx can be any interface{}
	Run(script Script, ctx interface{}) interface{}

	// RunWithResult executes a script and returns detailed result information
	RunWithResult(script Script, ctx *ScriptContext) *ScriptResult

	// SupportsLanguage returns true if the runner supports the given language
	SupportsLanguage(lang ScriptLanguage) bool

	// Shutdown gracefully shuts down the script runner
	Shutdown()
}
