package sshgate

import (
	"crypto/ed25519"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/gamemode"
	"github.com/talesmud/talesmud/pkg/mudserver"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
	"github.com/talesmud/talesmud/pkg/sshkeys"
)

// TestWebPasteDoesNotAuthenticate is the SSHA-01 squat: the attacker's paste
// of the victim's public key must not create a row, and a signature from
// that key must not enter the attacker's account.
func TestWebPasteDoesNotAuthenticate(t *testing.T) {
	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "squat.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	attacker, err := facade.UsersService().Create(&entities.User{
		RefID: "local:attacker", Nickname: "AttackerNick", Username: "attacker", Role: entities.RolePlayer,
	})
	if err != nil || attacker == nil {
		t.Fatal(err)
	}
	if _, err := facade.UsersService().Create(&entities.User{
		RefID: "local:victim", Nickname: "VictimNick", Username: "victim", Role: entities.RolePlayer,
	}); err != nil {
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
	line := string(ssh.MarshalAuthorizedKey(signer.PublicKey()))
	keys, err := sshkeys.Open(client.DB())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := keys.Add(attacker.RefID, line, "squat", "web", 10); !errors.Is(err, sshkeys.ErrRejected) {
		t.Fatalf("web paste stored a key: %v", err)
	}
	fp := ssh.FingerprintSHA256(signer.PublicKey())
	got, err := keys.ByFingerprint(fp)
	if err != nil || got != nil {
		t.Fatalf("fingerprint occupied: %+v %v", got, err)
	}
	mud := mudserver.New(facade)
	hook := &frameHook{}
	mud.SetSessionHook(hook)
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
	_, err = ssh.Dial("tcp", gate.Addr(), &ssh.ClientConfig{
		User:            "victim",
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	})
	if err == nil {
		t.Fatal("victim key authenticated after a web paste")
	}
	if u := hook.lastUser(); u != nil {
		t.Fatalf("paste landed in %s", u.RefID)
	}
}
