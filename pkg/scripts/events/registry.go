package events

import (
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/talesmud/talesmud/pkg/scripts"
)

// ScriptHandler defines a script that handles an event
type ScriptHandler struct {
	ScriptID string    `json:"scriptId"`
	Priority int       `json:"priority"` // Lower = runs first
	Filter   string    `json:"filter"`   // Optional Lua expression for filtering
	Async    bool      `json:"async"`    // Run in goroutine
	Enabled  bool      `json:"enabled"`
}

// ScriptResult contains the result of a script execution for an event
type ScriptResult struct {
	ScriptID string        `json:"scriptId"`
	Success  bool          `json:"success"`
	Error    string        `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
	Canceled bool          `json:"canceled"` // If script requested event cancellation
}

// ScriptsRepository interface for loading scripts
type ScriptsRepository interface {
	FindByID(id string) (*scripts.Script, error)
}

// EventRegistry manages script handlers for events
type EventRegistry struct {
	mu       sync.RWMutex
	handlers map[EventType][]ScriptHandler
	runner   scripts.ScriptRunner
	scripts  ScriptsRepository
}

// NewEventRegistry creates a new event registry
func NewEventRegistry(runner scripts.ScriptRunner) *EventRegistry {
	return &EventRegistry{
		handlers: make(map[EventType][]ScriptHandler),
		runner:   runner,
	}
}

// SetScriptsRepository sets the scripts repository for loading scripts
func (r *EventRegistry) SetScriptsRepository(repo ScriptsRepository) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.scripts = repo
}

// Register registers a script handler for an event
func (r *EventRegistry) Register(eventType EventType, handler ScriptHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if handler.Priority == 0 {
		handler.Priority = 100 // Default priority
	}
	if !handler.Enabled {
		handler.Enabled = true
	}

	r.handlers[eventType] = append(r.handlers[eventType], handler)

	// Sort by priority
	sort.Slice(r.handlers[eventType], func(i, j int) bool {
		return r.handlers[eventType][i].Priority < r.handlers[eventType][j].Priority
	})

	logrus.WithField("event", eventType).WithField("script", handler.ScriptID).Info("Registered event handler")
}

// Unregister removes a script handler for an event
func (r *EventRegistry) Unregister(eventType EventType, scriptID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	handlers := r.handlers[eventType]
	for i, h := range handlers {
		if h.ScriptID == scriptID {
			r.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			logrus.WithField("event", eventType).WithField("script", scriptID).Info("Unregistered event handler")
			return
		}
	}
}

// UnregisterAll removes all handlers for a script
func (r *EventRegistry) UnregisterAll(scriptID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for eventType, handlers := range r.handlers {
		filtered := make([]ScriptHandler, 0, len(handlers))
		for _, h := range handlers {
			if h.ScriptID != scriptID {
				filtered = append(filtered, h)
			}
		}
		r.handlers[eventType] = filtered
	}
}

// GetHandlers returns all handlers for an event type
func (r *EventRegistry) GetHandlers(eventType EventType) []ScriptHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handlers := r.handlers[eventType]
	result := make([]ScriptHandler, len(handlers))
	copy(result, handlers)
	return result
}

// Dispatch dispatches an event to all registered handlers
func (r *EventRegistry) Dispatch(eventType EventType, ctx *EventContext) []ScriptResult {
	handlers := r.GetHandlers(eventType)
	if len(handlers) == 0 {
		return nil
	}

	if r.scripts == nil {
		logrus.Warn("Event registry has no scripts repository, cannot dispatch events")
		return nil
	}

	results := make([]ScriptResult, 0, len(handlers))
	var mu sync.Mutex

	for _, handler := range handlers {
		if !handler.Enabled {
			continue
		}

		// Load the script
		script, err := r.scripts.FindByID(handler.ScriptID)
		if err != nil || script == nil {
			results = append(results, ScriptResult{
				ScriptID: handler.ScriptID,
				Success:  false,
				Error:    "script not found",
			})
			continue
		}

		// Execute the script
		if handler.Async {
			// Run asynchronously
			go func(h ScriptHandler, s *scripts.Script) {
				result := r.executeScript(s, ctx)
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
			}(handler, script)
		} else {
			// Run synchronously
			result := r.executeScript(script, ctx)
			results = append(results, result)

			// Check if script requested event cancellation
			if result.Canceled {
				break
			}
		}
	}

	return results
}

// executeScript executes a single script with the event context
func (r *EventRegistry) executeScript(script *scripts.Script, ctx *EventContext) ScriptResult {
	start := time.Now()

	// Create script context from event context
	scriptCtx := scripts.NewScriptContext()
	for k, v := range ctx.ToMap() {
		scriptCtx.Set(k, v)
	}

	// Execute
	result := r.runner.RunWithResult(*script, scriptCtx)

	scriptResult := ScriptResult{
		ScriptID: script.ID,
		Success:  result.Success,
		Duration: time.Since(start),
	}

	if !result.Success {
		scriptResult.Error = result.Error
	}

	// Check for cancellation request in result
	if resultMap, ok := result.Result.(map[string]interface{}); ok {
		if canceled, ok := resultMap["cancel"].(bool); ok && canceled {
			scriptResult.Canceled = true
		}
	}

	return scriptResult
}

// HasHandlers returns true if there are handlers for the event type
func (r *EventRegistry) HasHandlers(eventType EventType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.handlers[eventType]) > 0
}

// Clear removes all handlers
func (r *EventRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers = make(map[EventType][]ScriptHandler)
}

// Stats returns statistics about registered handlers
func (r *EventRegistry) Stats() map[string]int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := make(map[string]int)
	for eventType, handlers := range r.handlers {
		stats[string(eventType)] = len(handlers)
	}
	return stats
}
