package mudserver

import "sync"

type clientRegistry struct {
	mu      sync.RWMutex
	clients map[string]*Connection
}

func newClientRegistry() *clientRegistry {
	return &clientRegistry{clients: make(map[string]*Connection)}
}

func (r *clientRegistry) Set(id string, con *Connection) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[id] = con
}

func (r *clientRegistry) Get(id string) (*Connection, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	con, ok := r.clients[id]
	return con, ok
}

func (r *clientRegistry) Delete(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, id)
}

func (r *clientRegistry) DeleteIf(id string, con *Connection) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if current, ok := r.clients[id]; ok && current == con {
		delete(r.clients, id)
		return true
	}
	return false
}

func (r *clientRegistry) ForEach(fn func(id string, con *Connection)) {
	r.mu.RLock()
	snapshot := make(map[string]*Connection, len(r.clients))
	for id, con := range r.clients {
		snapshot[id] = con
	}
	r.mu.RUnlock()

	for id, con := range snapshot {
		fn(id, con)
	}
}
