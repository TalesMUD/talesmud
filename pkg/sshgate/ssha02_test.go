package sshgate

import (
	"crypto/ed25519"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/devicecode"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/rooms"
	"github.com/talesmud/talesmud/pkg/entities/traits"
	"github.com/talesmud/talesmud/pkg/gamemode"
	"github.com/talesmud/talesmud/pkg/mudserver"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
	"github.com/talesmud/talesmud/pkg/sshkeys"
)

func TestExactYN(t *testing.T) {
	if exactYN([]byte("maybe")) != "" || exactYN([]byte("yes")) != "" || exactYN([]byte("yn")) != "" {
		t.Fatal("a chunk that merely contains y is not yes")
	}
	if exactYN([]byte("y")) != "y" || exactYN([]byte("Y\r")) != "y" || exactYN([]byte("Y\r\n")) != "y" {
		t.Fatal("exact y")
	}
	if exactYN([]byte("n")) != "n" || exactYN([]byte("N\n")) != "n" {
		t.Fatal("exact n")
	}
	if exactYN([]byte("q")) != "q" || exactYN([]byte{3}) != "q" || exactYN([]byte{4}) != "q" {
		t.Fatal("quit")
	}
}

// TestPendingNDisconnects is the SSHA-02 PoC: "maybe" does not link, and N
// on the reconnect drops the pending row and does not enter that account.
func TestPendingNDisconnects(t *testing.T) {
	facade, keys, pending, devices, hook, gate := pendingGate(t)
	user, err := facade.UsersService().Create(&entities.User{
		RefID: "local:pendingatk", Nickname: "Vic\r\nFAKE-ACCOUNT \u202eBob", Username: "pendingatk", Role: entities.RolePlayer,
	})
	if err != nil || user == nil {
		t.Fatal(err)
	}
	signer := newEdSigner(t)
	fp := ssh.FingerprintSHA256(signer.PublicKey())
	conn := dialAuth(t, gate.Addr(), signer, true)
	session, stdin, out := doorShell(t, conn, "xterm", 80, 24, "")
	text := out.wait(t, []string{"Code:"}, 15*time.Second)
	code := regexp.MustCompile(`[BCDFGHJKLMNPQRSTVWXZ]{4}-[BCDFGHJKLMNPQRSTVWXZ]{4}`).FindString(text)
	if code == "" {
		t.Fatalf("no code in %q", text)
	}
	view, err := devices.Lookup(code, user.RefID, "127.0.0.1", func(string) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if view.KeyFP != fp {
		t.Fatalf("offer %s signed %s", view.KeyFP, fp)
	}
	if err := devices.Confirm(code, user.RefID, "127.0.0.1", view.CSRF, func(string) error { return nil }); err != nil {
		t.Fatal(err)
	}
	out.wait(t, []string{"Remember this computer?"}, 15*time.Second)
	if _, err := io.WriteString(stdin, "maybe"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(400 * time.Millisecond)
	if strings.Contains(out.snapshot(), "Reconnect once") {
		t.Fatal("chunk containing y confirmed the link")
	}
	if _, _, ok := pending.Get(fp); ok {
		t.Fatal("maybe stored a pending link")
	}
	if _, err := io.WriteString(stdin, "y"); err != nil {
		t.Fatal(err)
	}
	out.wait(t, []string{"Reconnect once"}, 15*time.Second)
	if ref, _, ok := pending.Get(fp); !ok || ref != user.RefID {
		t.Fatalf("pending %s %v", ref, ok)
	}
	_ = session.Close()
	_ = conn.Close()

	hook.clear()
	conn = dialAuth(t, gate.Addr(), signer, false)
	session, stdin, out = doorShell(t, conn, "xterm", 80, 24, "")
	defer session.Close()
	defer conn.Close()
	ntext := out.wait(t, []string{"[Y/N]"}, 15*time.Second)
	if !strings.Contains(ntext, "Link key") {
		t.Fatalf("prompt %q", ntext)
	}
	if _, err := io.WriteString(stdin, "n"); err != nil {
		t.Fatal(err)
	}
	out.wait(t, []string{"Key not linked."}, 15*time.Second)
	time.Sleep(300 * time.Millisecond)
	if u := hook.lastUser(); u != nil {
		t.Fatalf("N admitted %s", u.RefID)
	}
	if _, _, ok := pending.Get(fp); ok {
		t.Fatal("N left the pending row")
	}
	rows, err := keys.List(user.RefID)
	if err != nil || len(rows) != 0 {
		t.Fatalf("N stored a key %+v %v", rows, err)
	}
}

// TestCrossAccountNDoesNotAdmit is the SSHA-02 cross-account PoC: the victim
// presses N and must not enter the account that created the pending link.
func TestCrossAccountNDoesNotAdmit(t *testing.T) {
	facade, _, pending, devices, hook, gate := pendingGate(t)
	atk, err := facade.UsersService().Create(&entities.User{
		RefID: "local:atk2", Nickname: "AtkTwo", Username: "atk2", Role: entities.RolePlayer,
	})
	if err != nil || atk == nil {
		t.Fatal(err)
	}
	signer := newEdSigner(t)
	fp := ssh.FingerprintSHA256(signer.PublicKey())
	conn := dialAuth(t, gate.Addr(), signer, true)
	session, stdin, out := doorShell(t, conn, "xterm", 80, 24, "")
	text := out.wait(t, []string{"Code:"}, 15*time.Second)
	code := regexp.MustCompile(`[BCDFGHJKLMNPQRSTVWXZ]{4}-[BCDFGHJKLMNPQRSTVWXZ]{4}`).FindString(text)
	if code == "" {
		t.Fatalf("no code in %q", text)
	}
	view, err := devices.Lookup(code, atk.RefID, "127.0.0.1", func(string) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if view.KeyFP != fp {
		t.Fatalf("offer is not the signed key: %s", view.KeyFP)
	}
	if err := devices.Confirm(code, atk.RefID, "127.0.0.1", view.CSRF, func(string) error { return nil }); err != nil {
		t.Fatal(err)
	}
	out.wait(t, []string{"Remember this computer?"}, 15*time.Second)
	if _, err := io.WriteString(stdin, "y"); err != nil {
		t.Fatal(err)
	}
	out.wait(t, []string{"Reconnect once"}, 15*time.Second)
	_ = session.Close()
	_ = conn.Close()

	hook.clear()
	conn = dialAuth(t, gate.Addr(), signer, false)
	session, stdin, out = doorShell(t, conn, "xterm", 80, 24, "")
	defer session.Close()
	defer conn.Close()
	ntext := out.wait(t, []string{"[Y/N]"}, 15*time.Second)
	if !strings.Contains(ntext, "AtkTwo") {
		t.Fatalf("prompt %q", ntext)
	}
	if _, err := io.WriteString(stdin, "n"); err != nil {
		t.Fatal(err)
	}
	out.wait(t, []string{"Key not linked."}, 15*time.Second)
	time.Sleep(300 * time.Millisecond)
	if u := hook.lastUser(); u != nil {
		t.Fatalf("cross N admitted %s", u.RefID)
	}
	if _, _, ok := pending.Get(fp); ok {
		t.Fatal("cross N left the pending row")
	}
}

func pendingGate(t *testing.T) (service.Facade, *sshkeys.Store, *sshkeys.Pending, *devicecode.Store, *frameHook, *Gate) {
	t.Helper()
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "pending.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close() })
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	if _, err := facade.RoomsService().Store(&rooms.Room{
		Entity:      &entities.Entity{ID: "R0001"},
		Name:        "Harbor",
		Description: "SSH-PENDING-HARBOR stands quiet.",
		LookAt:      traits.LookAt{Detail: "SSH-PENDING-HARBOR stands quiet."},
	}); err != nil {
		t.Fatal(err)
	}
	keys, err := sshkeys.Open(client.DB())
	if err != nil {
		t.Fatal(err)
	}
	pending := sshkeys.NewPending(10 * time.Minute)
	devices := devicecode.New(devicecode.Config{})
	mud := mudserver.New(facade)
	hook := &frameHook{}
	mud.SetSessionHook(hook)
	mud.Run()
	gate, err := Listen(gamemode.SSHConfig{
		Enabled:     true,
		Listen:      "127.0.0.1:0",
		HostKeyPath: filepath.Join(t.TempDir(), "host_ed25519"),
		Keys:        gamemode.SSHKeysConfig{Enabled: true, MaxPerAccount: 10},
		Device:      gamemode.SSHDeviceConfig{Enabled: true, ActivateURL: "http://127.0.0.1/activate"},
		IdleTimeout: gamemode.Duration(30 * time.Minute),
	}, Deps{
		Mud: mud, Users: facade.UsersService(), Keys: keys, Pending: pending, Devices: devices,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { gate.Close() })
	return facade, keys, pending, devices, hook, gate
}

func newEdSigner(t *testing.T) ssh.Signer {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	return signer
}

func dialAuth(t *testing.T, addr string, signer ssh.Signer, interactive bool) *ssh.Client {
	t.Helper()
	methods := []ssh.AuthMethod{ssh.PublicKeys(signer)}
	if interactive {
		methods = append(methods, ssh.KeyboardInteractive(func(string, string, []string, []bool) ([]string, error) {
			return nil, nil
		}))
	}
	conn, err := ssh.Dial("tcp", addr, &ssh.ClientConfig{
		User:            "player",
		Auth:            methods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	return conn
}
