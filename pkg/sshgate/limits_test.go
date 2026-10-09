package sshgate

import (
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/gamemode"
)

func TestLimitsBanRateAndGuestCap(t *testing.T) {
	cfg := gamemode.SSHConfig{
		MaxConnections:      2,
		MaxPerIP:            1,
		NewConnsPerIPPerMin: 2,
		AuthFailBan: gamemode.SSHBanConfig{
			Failures: 2,
			Window:   gamemode.Duration(time.Minute),
			Ban:      gamemode.Duration(time.Minute),
		},
		Guest: gamemode.SSHGuestConfig{MaxConcurrent: 1, PerIPPerHour: 1},
	}
	lim := newLimits(cfg)
	now := time.Now()
	if err := lim.acquire("10.0.0.1", now); err != nil {
		t.Fatal(err)
	}
	if err := lim.acquire("10.0.0.1", now); err != errCap {
		t.Fatalf("per ip %v", err)
	}
	if err := lim.acquire("10.0.0.2", now); err != nil {
		t.Fatal(err)
	}
	if err := lim.acquire("10.0.0.3", now); err != errCap {
		t.Fatalf("global %v", err)
	}
	lim.release("10.0.0.1")
	if err := lim.acquire("10.0.0.9", now); err != nil {
		t.Fatal(err)
	}
	lim.authFail("10.1.1.1", now)
	lim.authFail("10.1.1.1", now)
	if !lim.bannedNow("10.1.1.1", now.Add(time.Second)) {
		t.Fatal("expected ban")
	}
	if err := lim.acquire("10.1.1.1", now.Add(time.Second)); err != errBanned {
		t.Fatalf("banned acquire %v", err)
	}
	if err := lim.acquireGuest("10.2.2.2", now); err != nil {
		t.Fatal(err)
	}
	if err := lim.acquireGuest("10.2.2.2", now); err != errRate {
		t.Fatalf("guest rate %v", err)
	}
	if err := lim.acquireGuest("10.2.2.3", now); err != errCap {
		t.Fatalf("guest cap %v", err)
	}
}
