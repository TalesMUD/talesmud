package lua

import (
	"sync"

	lua "github.com/yuin/gopher-lua"
)

// VMPool manages a pool of Lua states for efficient script execution
type VMPool struct {
	pool    chan *lua.LState
	factory func() *lua.LState
	size    int
	mu      sync.Mutex
	closed  bool
}

// NewVMPool creates a new VM pool with the given size and factory function
func NewVMPool(size int, factory func() *lua.LState) *VMPool {
	if size <= 0 {
		size = 10
	}

	p := &VMPool{
		pool:    make(chan *lua.LState, size),
		factory: factory,
		size:    size,
	}

	// Pre-populate the pool
	for i := 0; i < size; i++ {
		p.pool <- factory()
	}

	return p
}

// Get retrieves a Lua state from the pool, creating a new one if necessary
func (p *VMPool) Get() *lua.LState {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.mu.Unlock()

	select {
	case L := <-p.pool:
		return L
	default:
		// Pool exhausted, create a new one
		return p.factory()
	}
}

// Put returns a Lua state to the pool
// If the pool is full or closed, the state is closed instead
func (p *VMPool) Put(L *lua.LState) {
	if L == nil {
		return
	}

	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		L.Close()
		return
	}
	p.mu.Unlock()

	// Reset the state for reuse
	resetState(L)

	select {
	case p.pool <- L:
		// Successfully returned to pool
	default:
		// Pool is full, close the state
		L.Close()
	}
}

// Close closes all Lua states in the pool
func (p *VMPool) Close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	p.mu.Unlock()

	close(p.pool)
	for L := range p.pool {
		L.Close()
	}
}

// Size returns the current number of states in the pool
func (p *VMPool) Size() int {
	return len(p.pool)
}

// resetState resets a Lua state for reuse
func resetState(L *lua.LState) {
	// Clear the stack
	L.SetTop(0)

	// Clear any user-defined globals that might have been set
	// We keep the base modules intact
	G := L.Get(lua.GlobalsIndex).(*lua.LTable)

	// List of globals to preserve (built-in)
	preserve := map[string]bool{
		"_G":       true,
		"_VERSION": true,
		"assert":   true,
		"error":    true,
		"ipairs":   true,
		"next":     true,
		"pairs":    true,
		"pcall":    true,
		"print":    true,
		"rawequal": true,
		"rawget":   true,
		"rawset":   true,
		"select":   true,
		"setmetatable": true,
		"getmetatable": true,
		"tonumber": true,
		"tostring": true,
		"type":     true,
		"unpack":   true,
		"xpcall":   true,
		"string":   true,
		"table":    true,
		"math":     true,
		"coroutine": true,
		"tales":    true, // Our custom module
		"ctx":      true, // Context is reset per execution anyway
	}

	// Collect keys to remove
	var toRemove []lua.LValue
	G.ForEach(func(key, value lua.LValue) {
		if keyStr, ok := key.(lua.LString); ok {
			if !preserve[string(keyStr)] {
				toRemove = append(toRemove, key)
			}
		}
	})

	// Remove non-preserved globals
	for _, key := range toRemove {
		G.RawSet(key, lua.LNil)
	}
}
