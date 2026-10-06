package service

import (
	"testing"
	"time"
)

func TestGuestRateLimitAllowsSixtyThenBlocks(t *testing.T) {
	gs := &guestService{rateLimits: map[string][]time.Time{}}
	ip := "203.0.113.50"
	for i := 0; i < GuestRateLimitPerIP; i++ {
		if !gs.checkRateLimit(ip) {
			t.Fatalf("create %d of %d was rejected", i+1, GuestRateLimitPerIP)
		}
	}
	if gs.checkRateLimit(ip) {
		t.Fatal("guest create past the hourly cap was allowed")
	}
	if GuestRateLimitPerIP != 60 {
		t.Fatalf("playtest cap = %d, want 60", GuestRateLimitPerIP)
	}
}

func TestGuestRateLimitExpiresAfterOneHour(t *testing.T) {
	gs := &guestService{rateLimits: map[string][]time.Time{}}
	ip := "203.0.113.51"
	old := time.Now().Add(-61 * time.Minute)
	gs.rateLimits[ip] = make([]time.Time, GuestRateLimitPerIP)
	for i := range gs.rateLimits[ip] {
		gs.rateLimits[ip][i] = old
	}
	if !gs.checkRateLimit(ip) {
		t.Fatal("stale hour-old creates still blocked a new guest")
	}
}
