package sshkeys

import (
	"sync"
	"time"
)

// Pending is an in-memory "reconnect to finish" link. It is not a stored key.
type Pending struct {
	mu   sync.Mutex
	ttl  time.Duration
	now  func() time.Time
	byFP map[string]pendingRow
}

type pendingRow struct {
	ID      string
	FP      string
	UserRef string
	Expires time.Time
}

// NewPending keeps offers for ttl (10 minutes when ttl is zero).
func NewPending(ttl time.Duration) *Pending {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &Pending{ttl: ttl, now: time.Now, byFP: map[string]pendingRow{}}
}

// Offer records that this account asked to link this fingerprint.
func (p *Pending) Offer(fp, userRef string) (string, bool) {
	if p == nil || fp == "" || userRef == "" {
		return "", false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.purge()
	id := fp
	p.byFP[fp] = pendingRow{ID: id, FP: fp, UserRef: userRef, Expires: p.now().Add(p.ttl)}
	return id, true
}

// Get returns a live offer for the fingerprint.
func (p *Pending) Get(fp string) (userRef, id string, ok bool) {
	if p == nil || fp == "" {
		return "", "", false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.purge()
	row, found := p.byFP[fp]
	if !found {
		return "", "", false
	}
	return row.UserRef, row.ID, true
}

// Drop removes an offer.
func (p *Pending) Drop(fp string) {
	if p == nil {
		return
	}
	p.mu.Lock()
	delete(p.byFP, fp)
	p.mu.Unlock()
}

func (p *Pending) purge() {
	now := p.now()
	for fp, row := range p.byFP {
		if !now.Before(row.Expires) {
			delete(p.byFP, fp)
		}
	}
}
