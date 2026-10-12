package devicecode

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAppPollConfirmationAndConcurrentConsumption(t *testing.T) {
	now := time.Now()
	s := New(Config{})
	setClock(s, func() time.Time { return now })
	code, err := s.BeginApp("203.0.113.1", "some-client", "0.1")
	if err != nil {
		t.Fatal(err)
	}
	if code.DeviceCode == code.UserCode || len(code.DeviceCode) != 64 || code.Interval != 5 || code.ExpiresIn != 600 {
		t.Fatal("invalid start contract")
	}
	if _, _, err := s.Poll(code.UserCode, "203.0.113.1", nil); err != ErrExpired {
		t.Fatalf("human code as poll secret: %v", err)
	}
	if _, _, err := s.Poll(strings.Repeat("0", 64), "203.0.113.1", nil); err != ErrExpired {
		t.Fatalf("bad secret: %v", err)
	}
	if _, _, err := s.Poll(code.DeviceCode, "203.0.113.1", nil); err != ErrPending {
		t.Fatalf("pending: %v", err)
	}
	view, err := s.Lookup(code.UserCode, "auth0|account", "198.51.100.1", allowOK)
	if err != nil {
		t.Fatal(err)
	}
	if view.ClientKind != "app" || view.ClientName != "some-client" || view.ClientVersion != "0.1" {
		t.Fatal("missing client metadata")
	}
	if err := s.Confirm(code.UserCode, "auth0|account", "198.51.100.1", view.CSRF, allowOK); err != nil {
		t.Fatal(err)
	}
	now = now.Add(PollInterval)
	var issued, won atomic.Int32
	issue := func(ref string) (string, int, error) {
		if ref != "auth0|account" {
			t.Error("wrong confirmed account")
		}
		issued.Add(1)
		return "signed-session", 1800, nil
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			token, ttl, err := s.Poll(code.DeviceCode, "203.0.113.1", issue)
			if err == nil {
				if token != "signed-session" || ttl != 1800 {
					t.Error("invalid token response")
				}
				won.Add(1)
			} else if err != ErrExpired {
				t.Errorf("replay: %v", err)
			}
		}()
	}
	wg.Wait()
	if issued.Load() != 1 || won.Load() != 1 {
		t.Fatal("token not consumed exactly once")
	}
	now = now.Add(10 * time.Minute)
	s.State("missing")
	if len(s.byPollHash) != 0 || len(s.byID) != 0 {
		t.Fatal("consumed app record not purged")
	}
}

func TestAppPollSlowDownAndIssuerRetry(t *testing.T) {
	now := time.Now()
	s := New(Config{})
	setClock(s, func() time.Time { return now })
	code, _ := s.BeginApp("203.0.113.1", "client", "")
	if _, _, err := s.Poll(code.DeviceCode, "203.0.113.1", nil); err != ErrPending {
		t.Fatal(err)
	}
	if _, _, err := s.Poll(code.DeviceCode, "203.0.113.1", nil); err != ErrSlowDown {
		t.Fatal(err)
	}
	now = now.Add(10 * time.Second)
	if _, _, err := s.Poll(code.DeviceCode, "203.0.113.1", nil); err != ErrPending {
		t.Fatal(err)
	}
	now = now.Add(5 * time.Second)
	if _, _, err := s.Poll(code.DeviceCode, "203.0.113.2", nil); err != ErrSlowDown {
		t.Fatal("interval did not grow, or IP change bypassed it")
	}
	now = now.Add(15 * time.Second)
	view, _ := s.Lookup(code.UserCode, "account", "198.51.100.1", allowOK)
	if err := s.Confirm(code.UserCode, "account", "198.51.100.1", view.CSRF, allowOK); err != nil {
		t.Fatal(err)
	}
	issueErr := errors.New("issuer temporarily unavailable")
	if _, _, err := s.Poll(code.DeviceCode, "203.0.113.1", func(string) (string, int, error) { return "", 0, issueErr }); err != issueErr {
		t.Fatal(err)
	}
	now = now.Add(15 * time.Second)
	if _, _, err := s.Poll(code.DeviceCode, "203.0.113.1", func(string) (string, int, error) { return "session", 1800, nil }); err != nil {
		t.Fatal(err)
	}
}

func TestAppDenyExpiryAndGuest(t *testing.T) {
	for _, state := range []string{"pending", "confirmed", "denied"} {
		t.Run(state, func(t *testing.T) {
			now := time.Now()
			s := New(Config{TTL: 20 * time.Minute})
			setClock(s, func() time.Time { return now })
			code, _ := s.BeginApp("203.0.113.1", "client", "")
			if code.ExpiresIn != 600 {
				t.Fatal("app TTL exceeded ten minutes")
			}
			if err := s.Confirm(code.UserCode, "guest", "198.51.100.1", "", func(string) error { return ErrForbidden }); err != ErrForbidden {
				t.Fatal("guest confirmed")
			}
			view, err := s.Lookup(code.UserCode, "account", "198.51.100.1", allowOK)
			if err != nil {
				t.Fatal(err)
			}
			if state == "confirmed" {
				err = s.Confirm(code.UserCode, "account", "198.51.100.1", view.CSRF, allowOK)
			}
			if state == "denied" {
				err = s.Deny(code.UserCode, "account", "198.51.100.1", view.CSRF, allowOK)
				if _, _, pollErr := s.Poll(code.DeviceCode, "203.0.113.1", nil); pollErr != ErrDenied {
					t.Fatal(pollErr)
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			now = now.Add(10 * time.Minute)
			if _, _, err := s.Poll(code.DeviceCode, "203.0.113.1", nil); err != ErrExpired {
				t.Fatal(err)
			}
			if len(s.byID) != 0 || len(s.byHash) != 0 || len(s.byPollHash) != 0 {
				t.Fatal("expired app retained")
			}
		})
	}
}

func TestAppStartAndPollLimits(t *testing.T) {
	now := time.Now()
	s := New(Config{StartsPerIP: 1, StartsGlobal: 2, MaxPending: 10})
	setClock(s, func() time.Time { return now })
	if _, err := s.BeginApp("2001:db8::1", "client", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := s.BeginApp("2001:db8::2", "client", ""); err != ErrLimited {
		t.Fatal("start /64 not limited")
	}
	if _, err := s.BeginApp("203.0.113.1", "client", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := s.BeginApp("203.0.113.2", "client", ""); err != ErrLimited {
		t.Fatal("global start limit missing")
	}
	now = now.Add(11 * time.Minute)
	if _, err := s.BeginApp("2001:db8::2", "client", ""); err != nil {
		t.Fatal("start window did not reset")
	}
	p := New(Config{PollsPerIP: 1})
	setClock(p, func() time.Time { return now })
	if _, _, err := p.Poll("bad", "2001:db8::1", nil); err != ErrExpired {
		t.Fatal(err)
	}
	if _, _, err := p.Poll("other", "2001:db8::2", nil); err != ErrSlowDown {
		t.Fatal("bad secrets did not share poll IP limit")
	}
	now = now.Add(time.Minute)
	if _, _, err := p.Poll("bad", "2001:db8::2", nil); err != ErrExpired {
		t.Fatal("poll window did not reset")
	}
	shared := New(Config{MaxPending: 1})
	if _, _, _, err := shared.Begin("203.0.113.1", "mud", "ssh", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := shared.BeginApp("203.0.113.2", "client", ""); err != ErrLimited {
		t.Fatal("app bypassed shared global cap")
	}
}
