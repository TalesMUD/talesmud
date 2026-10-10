package devicecode

import (
	"errors"
	"testing"
	"time"
)

func allowOK(string) error { return nil }

func setClock(s *Store, now func() time.Time) {
	if s == nil || now == nil {
		return
	}
	s.mu.Lock()
	s.now = now
	s.mu.Unlock()
}

func TestDeviceCodeSingleUseAndExpiry(t *testing.T) {
	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	s := New(Config{TTL: time.Minute, PerUser: 10, PerIP: 20})
	setClock(s, func() time.Time { return now })
	id, display, _, err := s.Begin("203.0.113.5", "door", "SSH-2.0-OpenSSH", []string{"SHA256:abcdefghijklmnopqrstuvwxyz012345"})
	if err != nil {
		t.Fatal(err)
	}
	norm, ok := Normalize(display)
	if !ok || Format(norm) != display {
		t.Fatalf("code %s", display)
	}
	if _, bad := Normalize("ABCD-EFGH"); bad {
		t.Fatal("vowels accepted")
	}
	view, err := s.Lookup(display, "local:ada", "198.51.100.8", allowOK)
	if err != nil {
		t.Fatal(err)
	}
	if view.IP != "203.0.113.5" || view.KeyFP == "" || view.CSRF == "" || view.Mode != "door" {
		t.Fatalf("view %+v", view)
	}
	if err := s.Confirm(display, "local:ada", "198.51.100.8", "nope", allowOK); err == nil {
		t.Fatal("bad csrf")
	}
	if err := s.Confirm(display, "local:ada", "198.51.100.8", view.CSRF, allowOK); err != nil {
		t.Fatal(err)
	}
	if err := s.Confirm(display, "local:ada", "198.51.100.8", view.CSRF, allowOK); !errors.Is(err, ErrNotFound) {
		t.Fatalf("reuse %v", err)
	}
	ref, ok := s.Take(id)
	if !ok || ref != "local:ada" {
		t.Fatalf("take %q %v", ref, ok)
	}
	if _, ok := s.Take(id); ok {
		t.Fatal("taken twice")
	}
	id2, display2, _, err := s.Begin("203.0.113.5", "mud", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(2 * time.Minute)
	if s.State(id2) != "expired" && s.State(id2) != "missing" {
		t.Fatalf("state %s", s.State(id2))
	}
	if _, err := s.Lookup(display2, "local:ada", "198.51.100.8", allowOK); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expired %v", err)
	}
}

func TestDeviceCodeLimitsAndGuest(t *testing.T) {
	s := New(Config{MaxPendingPerIP: 1, PerUser: 2, PerIP: 2})
	if _, _, _, err := s.Begin("203.0.113.9", "door", "", nil); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := s.Begin("203.0.113.9", "door", "", nil); !errors.Is(err, ErrLimited) {
		t.Fatalf("pending %v", err)
	}
	_, display, _, err := s.Begin("203.0.113.10", "door", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	guest := func(string) error { return ErrForbidden }
	if _, err := s.Lookup(display, "guest", "198.51.100.1", guest); !errors.Is(err, ErrForbidden) {
		t.Fatal(err)
	}
	if _, err := s.Lookup(display, "local:ada", "198.51.100.1", allowOK); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Lookup(display, "local:ada", "198.51.100.1", allowOK); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Lookup(display, "local:ada", "198.51.100.1", allowOK); !errors.Is(err, ErrLimited) {
		t.Fatalf("rate %v", err)
	}
}
