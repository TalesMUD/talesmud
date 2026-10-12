package sshgate

import (
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/gamemode"
)

func TestIPv6LimitsSharePrefix(t *testing.T) {
	cfg := gamemode.SSHConfig{
		MaxConnections:      10,
		MaxPerIP:            1,
		NewConnsPerIPPerMin: 10,
		AuthFailBan: gamemode.SSHBanConfig{
			Failures: 2,
			Window:   gamemode.Duration(time.Minute),
			Ban:      gamemode.Duration(time.Minute),
		},
		Guest: gamemode.SSHGuestConfig{MaxConcurrent: 10, PerIPPerHour: 1},
	}
	lim := newLimits(cfg)
	now := time.Now()
	if err := lim.acquire("2001:db8:abcd::1", now); err != nil {
		t.Fatal(err)
	}
	if err := lim.acquire("2001:db8:abcd::2", now); err != errCap {
		t.Fatalf("sibling /64 %v", err)
	}
	if err := lim.acquire("2001:db8:abce::1", now); err != nil {
		t.Fatalf("other /64 %v", err)
	}
	lim.release("2001:db8:abcd::2")
	if err := lim.acquire("2001:db8:abcd::3", now); err != nil {
		t.Fatalf("release sibling %v", err)
	}
	if err := lim.acquire("::ffff:203.0.113.8", now); err != nil {
		t.Fatal(err)
	}
	if err := lim.acquire("203.0.113.8", now); err != errCap {
		t.Fatalf("mapped ipv4 %v", err)
	}
	lim.authFail("2001:db8:9999::1", now)
	lim.authFail("2001:db8:9999::2", now)
	if !lim.bannedNow("2001:db8:9999::3", now.Add(time.Second)) {
		t.Fatal("ban did not cover the /64")
	}
	if err := lim.acquireGuest("2001:db8:7777::1", now); err != nil {
		t.Fatal(err)
	}
	if err := lim.acquireGuest("2001:db8:7777::2", now); err != errRate {
		t.Fatalf("guest /64 %v", err)
	}
}
