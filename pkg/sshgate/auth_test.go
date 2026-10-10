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
	"github.com/talesmud/talesmud/pkg/gamemode"
	"github.com/talesmud/talesmud/pkg/mudserver"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
	"github.com/talesmud/talesmud/pkg/sshkeys"
)

func TestDoorDeviceLobbySkipsGuestSplash(t *testing.T) {
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "door.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	user, err := facade.UsersService().Create(&entities.User{
		RefID:    "local:dooruser",
		Nickname: "dooruser",
		Username: "dooruser",
		Role:     entities.RolePlayer,
	})
	if err != nil {
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
	cfg := gamemode.SSHConfig{
		Enabled:     true,
		Listen:      "127.0.0.1:0",
		HostKeyPath: filepath.Join(t.TempDir(), "host_ed25519"),
		Guest:       gamemode.SSHGuestConfig{Enabled: true},
		Keys:        gamemode.SSHKeysConfig{Enabled: true},
		Device: gamemode.SSHDeviceConfig{
			Enabled:     true,
			ActivateURL: "http://127.0.0.1:8031/activate",
		},
		IdleTimeout: gamemode.Duration(30 * time.Minute),
		Door:        gamemode.SSHDoorConfig{SplashHold: gamemode.Duration(30 * time.Millisecond)},
	}
	gate, err := Listen(cfg, Deps{
		Mud: mud, Guests: facade.GuestService(), Users: facade.UsersService(),
		Door: true, Keys: keys, Pending: pending, Devices: devices,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer gate.Close()

	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := ssh.Dial("tcp", gate.Addr(), &ssh.ClientConfig{
		User: "player",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
			ssh.KeyboardInteractive(func(string, string, []string, []bool) ([]string, error) {
				return nil, nil
			}),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	session, stdin, out := doorShell(t, conn, "xterm-256color", 80, 25, "")
	defer session.Close()
	text := out.wait(t, []string{"[G] guest"}, 15*time.Second)
	if strings.Contains(text, "DOOR-FRAME-MARKER") || hook.lastUser() != nil {
		t.Fatal("device lobby reached the game before confirm")
	}
	match := regexp.MustCompile(`[BCDFGHJKLMNPQRSTVWXZ]{4}-[BCDFGHJKLMNPQRSTVWXZ]{4}`).FindString(text)
	if match == "" {
		t.Fatalf("no code in %q", text)
	}
	view, err := devices.Lookup(match, user.RefID, "127.0.0.1", func(string) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if view.KeyFP == "" || view.KeyFP == MaskFingerprint(view.KeyFP) {
		t.Fatalf("stored fingerprint was already masked: %s", view.KeyFP)
	}
	if err := devices.Confirm(match, user.RefID, "127.0.0.1", view.CSRF, func(string) error { return nil }); err != nil {
		t.Fatal(err)
	}
	out.wait(t, []string{"Remember this computer?"}, 15*time.Second)
	if _, err := io.WriteString(stdin, "y"); err != nil {
		t.Fatal(err)
	}
	linked := out.wait(t, []string{"Key linked.", "DOOR-FRAME-MARKER"}, 15*time.Second)
	if strings.Contains(linked, "Reconnect once") {
		t.Fatal("one-step link asked for a reconnect")
	}
	rows, err := keys.List(user.RefID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].CreatedVia != "device" {
		t.Fatalf("signed key was not stored: %+v", rows)
	}
}
