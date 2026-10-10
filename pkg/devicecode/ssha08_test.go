package devicecode

import (
	"errors"
	"testing"
)

func TestIPv6PendingSharesPrefixAndGlobalCap(t *testing.T) {
	s := New(Config{MaxPendingPerIP: 1, MaxPending: 3})
	if _, _, _, err := s.Begin("2001:db8:abcd::1", "mud", "", nil); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := s.Begin("2001:db8:abcd::2", "mud", "", nil); !errors.Is(err, ErrLimited) {
		t.Fatalf("sibling /64 %v", err)
	}
	id, _, _, err := s.Begin("2001:db8:abce::1", "mud", "", nil)
	if err != nil {
		t.Fatalf("other /64 %v", err)
	}
	if viewIP := s.byID[id].ip; viewIP != "2001:db8:abce::1" {
		t.Fatalf("display ip %q", viewIP)
	}
	if _, _, _, err := s.Begin("203.0.113.1", "mud", "", nil); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := s.Begin("203.0.113.2", "mud", "", nil); !errors.Is(err, ErrLimited) {
		t.Fatalf("global cap %v", err)
	}
	if _, _, _, err := s.Begin("198.51.100.1", "mud", "", nil); !errors.Is(err, ErrLimited) {
		t.Fatalf("second global %v", err)
	}

	lookups := New(Config{PerIP: 1, PerUser: 10, MaxPending: 10})
	_, display, _, err := lookups.Begin("2001:db8:abcd::1", "mud", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lookups.Lookup(display, "local:ada", "2001:db8:abcd::9", allowOK); err != nil {
		t.Fatal(err)
	}
	if _, err := lookups.Lookup(display, "local:bea", "2001:db8:abcd::a", allowOK); !errors.Is(err, ErrLimited) {
		t.Fatalf("lookup /64 %v", err)
	}
}
