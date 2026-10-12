package worldindex

import "sync"

// Cache keeps one built index until Invalidate bumps the generation.
type Cache struct {
	mu         sync.Mutex
	generation uint64
	built      uint64
	index      *Index
	loads      int
}

// Invalidate marks the cached index stale. The next Index call reloads.
func (c *Cache) Invalidate() {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.generation++
	c.mu.Unlock()
}

// Loads reports how many times the loader has run.
func (c *Cache) Loads() int {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.loads
}

// Index returns the cached graph. load runs only after Invalidate, or the first call.
// A write that lands while load is running leaves the generation ahead, so the next call reloads.
func (c *Cache) Index(load func() (Snapshot, error)) (*Index, error) {
	if c == nil {
		snap, err := load()
		if err != nil {
			return nil, err
		}
		return Build(snap), nil
	}
	c.mu.Lock()
	gen := c.generation
	if c.index != nil && c.built == gen {
		ix := c.index
		c.mu.Unlock()
		return ix, nil
	}
	c.mu.Unlock()

	snap, err := load()
	if err != nil {
		return nil, err
	}
	ix := Build(snap)

	c.mu.Lock()
	defer c.mu.Unlock()
	c.loads++
	c.index = ix
	c.built = gen
	return ix, nil
}

// Live is the creator search index. Creator writes invalidate it.
var Live = &Cache{}
