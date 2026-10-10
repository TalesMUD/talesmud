package sshgate

import (
	"crypto/ed25519"
	"io"
	"path/filepath"
	"strings"
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
	"github.com/talesmud/talesmud/pkg/sshkeys"
)

func TestAccountMaxSessionExpires(t *testing.T) {
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "max.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	if _, err := facade.RoomsService().Store(&rooms.Room{
		Entity:      &entities.Entity{ID: "R0001"},
		Name:        "Harbor",
		Description: "SSH-MAX-HARBOR stands quiet.",
		LookAt:      traits.LookAt{Detail: "SSH-MAX-HARBOR stands quiet."},
	}); err != nil {
		t.Fatal(err)
	}
	user, err := facade.UsersService().Create(&entities.User{
		RefID:    "local:maxuser",
		Nickname: "maxuser",
		Username: "maxuser",
		Role:     entities.RolePlayer,
	})
	if err != nil || user == nil {
		t.Fatal(err)
	}
	keys, err := sshkeys.Open(client.DB())
	if err != nil {
		t.Fatal(err)
	}
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := keys.Add(user.RefID, string(ssh.MarshalAuthorizedKey(signer.PublicKey())), "laptop", "device", 10); err != nil {
		t.Fatal(err)
	}
	mud := mudserver.New(facade)
	mud.Run()
	gate, err := Listen(gamemode.SSHConfig{
		Enabled:     true,
		Listen:      "127.0.0.1:0",
		HostKeyPath: filepath.Join(t.TempDir(), "host_ed25519"),
		Keys:        gamemode.SSHKeysConfig{Enabled: true, MaxPerAccount: 10},
		IdleTimeout: gamemode.Duration(30 * time.Minute),
		MaxSession:  gamemode.Duration(1200 * time.Millisecond),
	}, Deps{Mud: mud, Users: facade.UsersService(), Keys: keys})
	if err != nil {
		t.Fatal(err)
	}
	defer gate.Close()

	conn, err := ssh.Dial("tcp", gate.Addr(), &ssh.ClientConfig{
		User:            "player",
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	session, err := conn.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := session.RequestPty("xterm", 24, 80, ssh.TerminalModes{}); err != nil {
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
	defer stdin.Close()
	if err := session.Shell(); err != nil {
		t.Fatal(err)
	}
	out := startCollector(stdout)
	got := out.wait(t, []string{"Your session has expired. Goodbye."}, 8*time.Second)
	if strings.Contains(got, "guest session") {
		t.Fatalf("account cap used the guest notice: %q", got)
	}
}

func TestClassicIdleTimeout(t *testing.T) {
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "idle.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	if _, err := facade.RoomsService().Store(&rooms.Room{
		Entity:      &entities.Entity{ID: "R0001"},
		Name:        "Harbor",
		Description: "SSH-IDLE-HARBOR stands quiet.",
		LookAt:      traits.LookAt{Detail: "SSH-IDLE-HARBOR stands quiet."},
	}); err != nil {
		t.Fatal(err)
	}
	mud := mudserver.New(facade)
	mud.Run()
	gate, err := Listen(gamemode.SSHConfig{
		Enabled:     true,
		Listen:      "127.0.0.1:0",
		HostKeyPath: filepath.Join(t.TempDir(), "host_ed25519"),
		Guest:       gamemode.SSHGuestConfig{Enabled: true, MaxSession: gamemode.Duration(30 * time.Minute)},
		IdleTimeout: gamemode.Duration(1200 * time.Millisecond),
	}, Deps{Mud: mud, Guests: facade.GuestService(), Users: facade.UsersService()})
	if err != nil {
		t.Fatal(err)
	}
	defer gate.Close()
	conn := dialGuest(t, gate.Addr())
	defer conn.Close()
	session, err := conn.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := session.RequestPty("xterm", 24, 80, ssh.TerminalModes{}); err != nil {
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
	out.wait(t, []string{"Connected to [TalesMUD]"}, 8*time.Second)
	_ = stdin
	out.wait(t, []string{"Idle timeout. Goodbye."}, 8*time.Second)
	_ = io.EOF
}
