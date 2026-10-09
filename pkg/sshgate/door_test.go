package sshgate

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/gamemode"
	"github.com/talesmud/talesmud/pkg/mudserver"
	"github.com/talesmud/talesmud/pkg/presentation/ansi"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

func TestGuestDoorOverSSH(t *testing.T) {
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "door.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	mud := mudserver.New(facade)
	hook := &frameHook{}
	mud.SetSessionHook(hook)
	mud.Run()

	cfg := gamemode.SSHConfig{
		Enabled:     true,
		Listen:      "127.0.0.1:0",
		HostKeyPath: filepath.Join(t.TempDir(), "host_ed25519"),
		Guest:       gamemode.SSHGuestConfig{Enabled: true},
		IdleTimeout: gamemode.Duration(30 * time.Minute),
	}
	gate, err := Listen(cfg, Deps{
		Mud:    mud,
		Guests: facade.GuestService(),
		Users:  facade.UsersService(),
		Door:   true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer gate.Close()

	t.Run("splash letterbox then frame", func(t *testing.T) {
		conn := dialGuest(t, gate.Addr())
		defer conn.Close()
		session, stdin, out := doorShell(t, conn, "xterm-256color", 100, 30, "")
		defer session.Close()
		got := out.wait(t, []string{"do you see", "\x1b[3;11H", "\x1b[?1049h"}, 15*time.Second)
		if !bytes.Contains([]byte(got), []byte("░▒▓█")) {
			t.Fatal("utf-8 blocks missing")
		}
		if _, err := io.WriteString(stdin, "\r"); err != nil {
			t.Fatal(err)
		}
		played := out.wait(t, []string{"DOOR-FRAME-MARKER"}, 15*time.Second)
		if !bytes.Contains([]byte(played), []byte("░▒▓█")) {
			t.Fatal("frame blocks missing")
		}
		if _, err := io.WriteString(stdin, "n"); err != nil {
			t.Fatal(err)
		}
		waitKey(t, hook, "n")
		if _, err := io.WriteString(stdin, "\r"); err != nil {
			t.Fatal(err)
		}
		waitKey(t, hook, "")
	})

	t.Run("too small", func(t *testing.T) {
		conn := dialGuest(t, gate.Addr())
		defer conn.Close()
		session, _, out := doorShell(t, conn, "xterm", 70, 20, "")
		defer session.Close()
		got := out.wait(t, []string{"70x20"}, 15*time.Second)
		if strings.Contains(got, "DOOR-FRAME-MARKER") || strings.Contains(got, "do you see") {
			t.Fatalf("small window painted the door %q", got)
		}
	})

	t.Run("cp437 splash", func(t *testing.T) {
		conn := dialGuest(t, gate.Addr())
		defer conn.Close()
		session, _, out := doorShell(t, conn, "xterm", 100, 30, "cp437")
		defer session.Close()
		deadline := time.Now().Add(15 * time.Second)
		for time.Now().Before(deadline) {
			text := out.snapshot()
			if bytes.Contains([]byte(text), []byte{0xb0, 0xb1, 0xb2, 0xdb}) && strings.Contains(text, "do you see") {
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
		t.Fatalf("cp437 splash %x", out.snapshot())
	})

	t.Run("replace restores the screen", func(t *testing.T) {
		conn := dialGuest(t, gate.Addr())
		defer conn.Close()
		session, stdin, out := doorShell(t, conn, "xterm", 80, 25, "")
		defer session.Close()
		out.wait(t, []string{"do you see"}, 15*time.Second)
		if _, err := io.WriteString(stdin, "\r"); err != nil {
			t.Fatal(err)
		}
		out.wait(t, []string{"DOOR-FRAME-MARKER"}, 15*time.Second)
		guest := hook.lastUser()
		if guest == nil {
			t.Fatal("no guest")
		}
		mud.AttachExternal(guest, discardTransport{ip: "127.0.0.1"})
		moved := out.wait(t, []string{"Session moved to another client."}, 10*time.Second)
		leave := strings.Index(moved, "\x1b[?1049l")
		notice := strings.Index(moved, "Session moved to another client.")
		if leave < 0 || notice < 0 || leave > notice {
			t.Fatalf("alt screen restore order leave %d notice %d", leave, notice)
		}
	})
}

func doorShell(t *testing.T, conn *ssh.Client, term string, cols, rows int, charset string) (*ssh.Session, io.Writer, *collector) {
	t.Helper()
	session, err := conn.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	// RequestPty takes height, then width.
	if err := session.RequestPty(term, rows, cols, ssh.TerminalModes{}); err != nil {
		t.Fatal(err)
	}
	if charset != "" {
		if err := session.Setenv("TALES_CHARSET", charset); err != nil {
			t.Fatal(err)
		}
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdin, err := session.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Shell(); err != nil {
		t.Fatal(err)
	}
	return session, stdin, startCollector(stdout)
}

type frameHook struct {
	mu   sync.Mutex
	keys []string
	last *entities.User
}

func (h *frameHook) Active() bool { return true }

func (h *frameHook) OnConnect(user *entities.User, send func(any)) {
	h.mu.Lock()
	h.last = user
	h.mu.Unlock()
	lines := make([]string, 25)
	lines[3] = "DOOR-FRAME-MARKER ░▒▓█"
	send(ansi.Frame{
		Type:      "door_frame",
		ANSI:      "\x1b[2J\x1b[H" + strings.Join(lines, "\r\n"),
		InputMode: "hotkey",
		Accepts:   []string{"hotkey"},
		Prompt:    ">",
		Cols:      80,
		Rows:      25,
	})
}

func (h *frameHook) OnInput(user *entities.User, text string, send func(any)) bool {
	h.mu.Lock()
	h.keys = append(h.keys, text)
	h.mu.Unlock()
	return true
}

func (h *frameHook) OnDisconnect(user *entities.User) {}

func (h *frameHook) OnNotice(user *entities.User, text, kind string, gen uint64, send func(any)) {
}

func (h *frameHook) saw(key string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, k := range h.keys {
		if k == key {
			return true
		}
	}
	return false
}

func (h *frameHook) lastUser() *entities.User {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.last
}

func waitKey(t *testing.T, h *frameHook, key string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if h.saw(key) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("missing key %q in %v", key, h.snapshot())
}

func (h *frameHook) snapshot() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]string, len(h.keys))
	copy(out, h.keys)
	return out
}

func (c *collector) snapshot() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buf.String()
}
