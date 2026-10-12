package sshgate

import (
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/talesmud/talesmud/pkg/gamemode"
)

func TestSlowReaderDoesNotBlock(t *testing.T) {
	link := &sshLink{queue: make(chan outbound, 1), fail: func(int) {}}
	if err := link.Send("one"); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	if err := link.Send("two"); err == nil {
		t.Fatal("full queue was accepted")
	}
	if time.Since(start) > 50*time.Millisecond {
		t.Fatal("send blocked the caller")
	}
}

func TestAuthTimeoutClosesSilentTCP(t *testing.T) {
	gate, err := Listen(gamemode.SSHConfig{
		Enabled:     true,
		Listen:      "127.0.0.1:0",
		HostKeyPath: filepath.Join(t.TempDir(), "host_ed25519"),
		AuthTimeout: gamemode.Duration(400 * time.Millisecond),
		IdleTimeout: gamemode.Duration(time.Minute),
	}, Deps{})
	if err != nil {
		t.Fatal(err)
	}
	defer gate.Close()
	conn, err := net.Dial("tcp", gate.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	start := time.Now()
	buf := make([]byte, 64)
	for {
		_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		_, err := conn.Read(buf)
		if err != nil {
			if time.Since(start) > 2*time.Second {
				t.Fatalf("silent handshake stayed open: %v", err)
			}
			return
		}
	}
}

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
