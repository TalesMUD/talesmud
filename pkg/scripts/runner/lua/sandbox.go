package lua

import (
	"context"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// SandboxConfig defines security restrictions for Lua script execution
type SandboxConfig struct {
	// MaxExecutionTime is the maximum time a script can run before being terminated
	MaxExecutionTime time.Duration

	// MaxMemoryBytes is the maximum memory a script can allocate (not enforced by gopher-lua)
	MaxMemoryBytes int64

	// AllowedModules lists the modules that can be loaded
	AllowedModules []string

	// DisabledGlobals lists global functions/modules to remove
	DisabledGlobals []string
}

// DefaultSandboxConfig returns the default sandbox configuration
func DefaultSandboxConfig() *SandboxConfig {
	return &SandboxConfig{
		MaxExecutionTime: 5 * time.Second,
		MaxMemoryBytes:   10 * 1024 * 1024, // 10MB
		AllowedModules: []string{
			"string",
			"table",
			"math",
			"tales", // Our custom module
		},
		DisabledGlobals: []string{
			"os",
			"io",
			"debug",
			"loadfile",
			"dofile",
			"load",
			"loadstring",
			"package",
		},
	}
}

// Apply applies the sandbox configuration to a Lua state
func (s *SandboxConfig) Apply(L *lua.LState) {
	// Remove dangerous globals
	for _, name := range s.DisabledGlobals {
		L.SetGlobal(name, lua.LNil)
	}

	// Remove potentially dangerous functions from allowed modules
	s.sanitizeStringModule(L)
}

// sanitizeStringModule removes dangerous functions from the string module
func (s *SandboxConfig) sanitizeStringModule(L *lua.LState) {
	// string.dump can be used to dump bytecode
	if strMod := L.GetGlobal("string"); strMod != lua.LNil {
		if tbl, ok := strMod.(*lua.LTable); ok {
			tbl.RawSetString("dump", lua.LNil)
		}
	}
}

// CreateContext creates a context with timeout for script execution
func (s *SandboxConfig) CreateContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), s.MaxExecutionTime)
}

// SetupInterrupt sets up the interrupt mechanism for the Lua state
func (s *SandboxConfig) SetupInterrupt(L *lua.LState, ctx context.Context) {
	L.SetContext(ctx)
}
