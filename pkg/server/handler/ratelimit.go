package handler

import (
	"sync"
	"time"
)

type windowLimiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	max    int
	window time.Duration
}

func newWindowLimiter(max int, window time.Duration) *windowLimiter {
	return &windowLimiter{
		hits:   make(map[string][]time.Time),
		max:    max,
		window: window,
	}
}

func (l *windowLimiter) Allow(key string) bool {
	if l == nil {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-l.window)
	prev := l.hits[key]
	recent := make([]time.Time, 0, len(prev)+1)
	for _, t := range prev {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	if len(recent) >= l.max {
		l.hits[key] = recent
		return false
	}
	l.hits[key] = append(recent, now)
	return true
}
