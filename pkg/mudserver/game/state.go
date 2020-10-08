package game

import (
	"sync"
	"time"

	cache "github.com/patrickmn/go-cache"
)

var once sync.Once

// State ...
type State interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
}

// State ...
type state struct {
	c *cache.Cache
}

func (s *state) Get(key string) (interface{}, bool) {
	return s.c.Get(key)
}

func (s *state) Set(key string, value interface{}) {
	s.c.Set(key, value, cache.NoExpiration)
}

var (
	instance State
)

// GetState ... return global MUD state cache
// can only contained cached states up to 60 minutes
// Should not be used for permament persistence
func GetState() State {

	once.Do(func() {
		instance = &state{
			c: cache.New(60*time.Minute, 90*time.Minute),
		}
	})

	return instance
}
