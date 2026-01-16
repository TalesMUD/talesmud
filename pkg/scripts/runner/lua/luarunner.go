package lua

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/talesmud/talesmud/pkg/mudserver/game/def"
	"github.com/talesmud/talesmud/pkg/scripts"
	"github.com/talesmud/talesmud/pkg/service"
)

// LuaRunner implements ScriptRunner for Lua scripts using gopher-lua
type LuaRunner struct {
	mu sync.RWMutex

	// Services for script API
	facade service.Facade
	game   def.GameCtrl

	// VM pool for performance
	pool *VMPool

	// Sandbox configuration
	sandbox *SandboxConfig

	// Module loaders (set by modules package)
	moduleLoaders map[string]func(*lua.LState, *LuaRunner) int
}

// NewLuaRunner creates a new Lua script runner
func NewLuaRunner() *LuaRunner {
	runner := &LuaRunner{
		sandbox:       DefaultSandboxConfig(),
		moduleLoaders: make(map[string]func(*lua.LState, *LuaRunner) int),
	}

	// Create VM pool with factory
	runner.pool = NewVMPool(10, func() *lua.LState {
		return runner.createState()
	})

	return runner
}

// SetServices injects the required services into the runner
func (r *LuaRunner) SetServices(facade service.Facade, game def.GameCtrl) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.facade = facade
	r.game = game
}

// GetFacade returns the service facade
func (r *LuaRunner) GetFacade() service.Facade {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.facade
}

// GetGame returns the game controller
func (r *LuaRunner) GetGame() def.GameCtrl {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.game
}

// RegisterModule registers a custom module loader
func (r *LuaRunner) RegisterModule(name string, loader func(*lua.LState, *LuaRunner) int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.moduleLoaders[name] = loader
}

// SupportsLanguage returns true if the runner supports the given language
func (r *LuaRunner) SupportsLanguage(lang scripts.ScriptLanguage) bool {
	return lang == scripts.ScriptLanguageLua
}

// Shutdown gracefully shuts down the script runner
func (r *LuaRunner) Shutdown() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.pool != nil {
		r.pool.Close()
	}
}

// Run executes a Lua script with the given context (backward compatibility)
func (r *LuaRunner) Run(script scripts.Script, ctx interface{}) interface{} {
	scriptCtx := scripts.NewScriptContext()
	scriptCtx.Set("ctx", ctx)

	result := r.RunWithResult(script, scriptCtx)
	if !result.Success {
		logrus.WithField("script", script.Name).WithField("error", result.Error).Error("Script execution failed")
		return result.Error
	}
	return result.Result
}

// RunWithResult executes a Lua script and returns detailed result information
func (r *LuaRunner) RunWithResult(script scripts.Script, ctx *scripts.ScriptContext) *scripts.ScriptResult {
	start := time.Now()

	logrus.WithField("Script", script.Name).WithField("Language", "lua").Info("Executing script...")

	// Get a Lua state from the pool
	L := r.pool.Get()
	if L == nil {
		return &scripts.ScriptResult{
			Success:  false,
			Error:    "failed to get Lua state from pool",
			Duration: time.Since(start),
		}
	}
	defer r.pool.Put(L)

	// Set up timeout context
	execCtx, cancel := r.sandbox.CreateContext()
	defer cancel()
	r.sandbox.SetupInterrupt(L, execCtx)

	// Set context variables
	r.setContext(L, ctx)

	// Execute the script
	result := r.executeWithTimeout(L, script.Code, execCtx)
	result.Duration = time.Since(start)

	return result
}

// createState creates a new Lua state with modules and sandbox applied
func (r *LuaRunner) createState() *lua.LState {
	L := lua.NewState(lua.Options{
		SkipOpenLibs: true, // Don't open all libs, we'll be selective
	})

	// Open safe standard libraries
	lua.OpenBase(L)
	lua.OpenString(L)
	lua.OpenTable(L)
	lua.OpenMath(L)

	// Apply sandbox restrictions
	r.sandbox.Apply(L)

	// Register custom modules
	r.registerTalesModule(L)

	return L
}

// registerTalesModule registers the main "tales" module
func (r *LuaRunner) registerTalesModule(L *lua.LState) {
	// Create the main tales table
	tales := L.NewTable()

	// Register sub-modules
	r.mu.RLock()
	for name, loader := range r.moduleLoaders {
		subMod := L.NewTable()
		L.Push(subMod)
		loader(L, r)
		tales.RawSetString(name, L.Get(-1))
		L.Pop(1)
	}
	r.mu.RUnlock()

	// If no modules registered yet, create empty sub-tables for basic structure
	if tales.RawGetString("items") == lua.LNil {
		tales.RawSetString("items", L.NewTable())
	}
	if tales.RawGetString("rooms") == lua.LNil {
		tales.RawSetString("rooms", L.NewTable())
	}
	if tales.RawGetString("characters") == lua.LNil {
		tales.RawSetString("characters", L.NewTable())
	}
	if tales.RawGetString("npcs") == lua.LNil {
		tales.RawSetString("npcs", L.NewTable())
	}
	if tales.RawGetString("dialogs") == lua.LNil {
		tales.RawSetString("dialogs", L.NewTable())
	}
	if tales.RawGetString("game") == lua.LNil {
		tales.RawSetString("game", L.NewTable())
	}
	if tales.RawGetString("utils") == lua.LNil {
		r.registerUtilsModule(L, tales)
	}

	L.SetGlobal("tales", tales)
}

// registerUtilsModule registers basic utility functions
func (r *LuaRunner) registerUtilsModule(L *lua.LState, tales *lua.LTable) {
	utils := L.NewTable()

	// tales.utils.random(min, max)
	utils.RawSetString("random", L.NewFunction(func(L *lua.LState) int {
		min := L.CheckInt(1)
		max := L.CheckInt(2)
		if max < min {
			min, max = max, min
		}
		// Use Lua's math.random which is already seeded
		L.Push(lua.LNumber(min + int(L.GetGlobal("math").(*lua.LTable).RawGetString("random").(*lua.LFunction).GFunction(L))))
		return 1
	}))

	// tales.utils.uuid()
	utils.RawSetString("uuid", L.NewFunction(func(L *lua.LState) int {
		// Simple UUID v4 generation
		L.Push(lua.LString(generateUUID()))
		return 1
	}))

	// tales.utils.now()
	utils.RawSetString("now", L.NewFunction(func(L *lua.LState) int {
		L.Push(lua.LNumber(time.Now().Unix()))
		return 1
	}))

	// tales.utils.log(level, message)
	utils.RawSetString("log", L.NewFunction(func(L *lua.LState) int {
		level := L.CheckString(1)
		message := L.CheckString(2)

		switch level {
		case "debug":
			logrus.Debug("[Script] " + message)
		case "info":
			logrus.Info("[Script] " + message)
		case "warn":
			logrus.Warn("[Script] " + message)
		case "error":
			logrus.Error("[Script] " + message)
		default:
			logrus.Info("[Script] " + message)
		}
		return 0
	}))

	tales.RawSetString("utils", utils)
}

// setContext sets the context variables in the Lua state
func (r *LuaRunner) setContext(L *lua.LState, ctx *scripts.ScriptContext) {
	if ctx == nil {
		return
	}

	// Create a ctx table
	ctxTable := L.NewTable()

	for key, value := range ctx.Data {
		// Use luar to convert Go values to Lua values
		ctxTable.RawSetString(key, luar.New(L, value))
	}

	L.SetGlobal("ctx", ctxTable)
}

// executeWithTimeout executes Lua code with timeout protection
func (r *LuaRunner) executeWithTimeout(L *lua.LState, code string, ctx context.Context) *scripts.ScriptResult {
	// Channel to receive execution result
	done := make(chan *scripts.ScriptResult, 1)

	go func() {
		err := L.DoString(code)
		if err != nil {
			done <- &scripts.ScriptResult{
				Success: false,
				Error:   err.Error(),
			}
			return
		}

		// Try to get return value
		result := r.getReturnValue(L)
		done <- &scripts.ScriptResult{
			Success: true,
			Result:  result,
		}
	}()

	select {
	case result := <-done:
		return result
	case <-ctx.Done():
		// Timeout - the context cancellation will interrupt the Lua state
		return &scripts.ScriptResult{
			Success: false,
			Error:   "script execution timeout exceeded",
		}
	}
}

// getReturnValue gets the return value from the Lua stack
func (r *LuaRunner) getReturnValue(L *lua.LState) interface{} {
	// Check if there's a return value on the stack
	if L.GetTop() > 0 {
		val := L.Get(-1)
		return luaValueToGo(val)
	}

	// Check if ctx was modified
	ctx := L.GetGlobal("ctx")
	if ctx != lua.LNil {
		return luaValueToGo(ctx)
	}

	return nil
}

// luaValueToGo converts a Lua value to a Go value
func luaValueToGo(val lua.LValue) interface{} {
	switch v := val.(type) {
	case lua.LBool:
		return bool(v)
	case lua.LNumber:
		return float64(v)
	case lua.LString:
		return string(v)
	case *lua.LTable:
		return luaTableToMap(v)
	case *lua.LNilType:
		return nil
	default:
		return v.String()
	}
}

// luaTableToMap converts a Lua table to a Go map or slice
func luaTableToMap(tbl *lua.LTable) interface{} {
	// Check if it's an array (consecutive integer keys starting at 1)
	isArray := true
	maxIndex := 0
	tbl.ForEach(func(key, _ lua.LValue) {
		if num, ok := key.(lua.LNumber); ok {
			idx := int(num)
			if idx > maxIndex {
				maxIndex = idx
			}
		} else {
			isArray = false
		}
	})

	if isArray && maxIndex > 0 {
		arr := make([]interface{}, maxIndex)
		tbl.ForEach(func(key, value lua.LValue) {
			if num, ok := key.(lua.LNumber); ok {
				arr[int(num)-1] = luaValueToGo(value)
			}
		})
		return arr
	}

	// Convert to map
	m := make(map[string]interface{})
	tbl.ForEach(func(key, value lua.LValue) {
		keyStr := ""
		switch k := key.(type) {
		case lua.LString:
			keyStr = string(k)
		case lua.LNumber:
			keyStr = k.String()
		default:
			keyStr = key.String()
		}
		m[keyStr] = luaValueToGo(value)
	})
	return m
}

// generateUUID generates a UUID v4
func generateUUID() string {
	return uuid.New().String()
}
