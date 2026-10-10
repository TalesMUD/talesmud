package sshgate

import (
	"crypto/ed25519"
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

func TestRevokeClosesLiveSession(t *testing.T) {
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "revoke.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	if _, err := facade.RoomsService().Store(&rooms.Room{
		Entity:      &entities.Entity{ID: "R0001"},
		Name:        "Harbor",
		Description: "SSH-REVOKE-HARBOR stands quiet.",
		LookAt:      traits.LookAt{Detail: "SSH-REVOKE-HARBOR stands quiet."},
	}); err != nil {
		t.Fatal(err)
	}
	keys, err := sshkeys.Open(client.DB())
	if err != nil {
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
	}, Deps{Mud: mud, Users: facade.UsersService(), Keys: keys})
	if err != nil {
		t.Fatal(err)
	}
	defer gate.Close()

	ada, adaSigner := revokeUser(t, facade, keys, "local:ada", "ada")
	bea, beaSigner := revokeUser(t, facade, keys, "local:bea", "bea")
	adaOut, adaClose := revokeShell(t, gate.Addr(), adaSigner)
	beaOut, beaClose := revokeShell(t, gate.Addr(), beaSigner)
	defer adaClose()
	defer beaClose()
	adaOut.wait(t, []string{"Connected to [TalesMUD]"}, 8*time.Second)
	beaOut.wait(t, []string{"Connected to [TalesMUD]"}, 8*time.Second)

	fp := ssh.FingerprintSHA256(adaSigner.PublicKey())
	rows, err := keys.List(ada.RefID)
	if err != nil || len(rows) != 1 {
		t.Fatalf("ada keys %+v %v", rows, err)
	}
	gone, err := keys.Delete(ada.RefID, rows[0].ID)
	if err != nil || gone != fp {
		t.Fatalf("delete %q %v", gone, err)
	}
	gate.CloseFingerprint(fp)
	adaOut.wait(t, []string{"This session was closed."}, 8*time.Second)
	if strings.Contains(collectorText(beaOut), "This session was closed.") {
		t.Fatal("revoking one key closed the other session")
	}
	if _, err := ssh.Dial("tcp", gate.Addr(), &ssh.ClientConfig{
		User:            "player",
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(adaSigner)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}); err == nil {
		t.Fatal("deleted key authenticated")
	}

	gate.CloseUser(bea.ID)
	beaOut.wait(t, []string{"This session was closed."}, 8*time.Second)
}

func revokeUser(t *testing.T, facade service.Facade, keys *sshkeys.Store, ref, name string) (*entities.User, ssh.Signer) {
	t.Helper()
	user, err := facade.UsersService().Create(&entities.User{
		RefID:    ref,
		Nickname: name,
		Username: name,
		Role:     entities.RolePlayer,
	})
	if err != nil || user == nil || user.ID == "" {
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
	return user, signer
}

func revokeShell(t *testing.T, addr string, signer ssh.Signer) (*collector, func()) {
	t.Helper()
	conn, err := ssh.Dial("tcp", addr, &ssh.ClientConfig{
		User:            "player",
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	session, err := conn.NewSession()
	if err != nil {
		_ = conn.Close()
		t.Fatal(err)
	}
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
	return startCollector(stdout), func() {
		_ = stdin.Close()
		_ = session.Close()
		_ = conn.Close()
	}
}

func collectorText(c *collector) string {
	if c == nil {
		return ""
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buf.String()
}
