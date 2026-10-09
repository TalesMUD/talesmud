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
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/gamemode"
	"github.com/talesmud/talesmud/pkg/mudserver"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

func TestGuestMudOverSSH(t *testing.T) {
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "guest.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	if _, err := facade.RoomsService().Store(&rooms.Room{
		Entity:      &entities.Entity{ID: "R0001"},
		Name:        "Harbor",
		Description: "SSH-SMOKE-HARBOR stands quiet.",
		LookAt:      traits.LookAt{Detail: "SSH-SMOKE-HARBOR stands quiet."},
	}); err != nil {
		t.Fatal(err)
	}
	mud := mudserver.New(facade)
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
	})
	if err != nil {
		t.Fatal(err)
	}
	defer gate.Close()
	if !gate.PublicInfo().Enabled || !gate.PublicInfo().GuestEnabled || len(gate.PublicInfo().HostKeyFingerprints) != 1 {
		t.Fatalf("info %+v", gate.PublicInfo())
	}

	t.Run("play and replace", func(t *testing.T) {
		conn := dialGuest(t, gate.Addr())
		defer conn.Close()
		session, err := conn.NewSession()
		if err != nil {
			t.Fatal(err)
		}
		defer session.Close()
		if err := session.RequestPty("xterm", 80, 24, ssh.TerminalModes{}); err != nil {
			t.Fatal(err)
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
		out := startCollector(stdout)
		out.wait(t, []string{"Connected to [TalesMUD]", "SSH-SMOKE-HARBOR"}, 20*time.Second)
		if _, err := io.WriteString(stdin, "look\r"); err != nil {
			t.Fatal(err)
		}
		got := out.wait(t, []string{"You look around", "SSH-SMOKE-HARBOR"}, 15*time.Second)
		if strings.Count(got, "SSH-SMOKE-HARBOR") < 2 {
			t.Fatalf("look output %q", got)
		}
		users, err := facade.UsersService().FindAll()
		if err != nil {
			t.Fatal(err)
		}
		var guest *entities.User
		for _, user := range users {
			if user != nil && user.IsGuest {
				guest = user
				break
			}
		}
		if guest == nil {
			t.Fatal("no guest user")
		}
		mud.AttachExternal(guest, discardTransport{ip: "127.0.0.1"})
		moved := out.wait(t, []string{"Session moved to another client."}, 10*time.Second)
		if strings.Contains(moved, guest.ID) && strings.Contains(moved, "BEGIN OPENSSH") {
			t.Fatal("session text leaked a key")
		}
	})

	t.Run("password refused", func(t *testing.T) {
		_, err := ssh.Dial("tcp", gate.Addr(), &ssh.ClientConfig{
			User:            "player",
			Auth:            []ssh.AuthMethod{ssh.Password("not-a-password")},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         5 * time.Second,
		})
		if err == nil {
			t.Fatal("password auth succeeded")
		}
	})

	t.Run("exec subsystem forwarding and no pty", func(t *testing.T) {
		reject := func(check func(*ssh.Client) error) {
			t.Helper()
			conn := dialGuest(t, gate.Addr())
			defer conn.Close()
			if err := check(conn); err == nil {
				t.Fatal("request was accepted")
			}
		}
		reject(func(conn *ssh.Client) error {
			sess, err := conn.NewSession()
			if err != nil {
				return err
			}
			return sess.Run("id")
		})
		reject(func(conn *ssh.Client) error {
			sess, err := conn.NewSession()
			if err != nil {
				return err
			}
			return sess.RequestSubsystem("sftp")
		})
		reject(func(conn *ssh.Client) error {
			_, err := conn.Dial("tcp", "127.0.0.1:1")
			return err
		})
		reject(func(conn *ssh.Client) error {
			sess, err := conn.NewSession()
			if err != nil {
				return err
			}
			return sess.Shell()
		})
	})
}

func dialGuest(t *testing.T, addr string) *ssh.Client {
	t.Helper()
	conn, err := ssh.Dial("tcp", addr, &ssh.ClientConfig{
		User:            "guest",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	return conn
}

type discardTransport struct{ ip string }

func (d discardTransport) Send(any) error    { return nil }
func (d discardTransport) Close(int, string) {}
func (d discardTransport) RemoteIP() string  { return d.ip }
func (d discardTransport) Kind() string      { return "ws" }

type collector struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func startCollector(r io.Reader) *collector {
	c := &collector{}
	go func() {
		tmp := make([]byte, 512)
		for {
			n, err := r.Read(tmp)
			if n > 0 {
				c.mu.Lock()
				c.buf.Write(tmp[:n])
				c.mu.Unlock()
			}
			if err != nil {
				return
			}
		}
	}()
	return c
}

func (c *collector) wait(t *testing.T, needles []string, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		c.mu.Lock()
		text := c.buf.String()
		c.mu.Unlock()
		ready := true
		for _, needle := range needles {
			if !strings.Contains(text, needle) {
				ready = false
				break
			}
		}
		if ready {
			return text
		}
		time.Sleep(20 * time.Millisecond)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	t.Fatalf("missing %v in %q", needles, c.buf.String())
	return ""
}
