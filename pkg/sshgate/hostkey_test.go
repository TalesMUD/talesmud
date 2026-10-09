package sshgate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/talesmud/talesmud/pkg/gamemode"
)

func TestHostKeyCreateModeAndRefuseLoose(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "keys", "host_ed25519")
	signer, fp, err := LoadOrCreateHostKey(path)
	if err != nil {
		t.Fatal(err)
	}
	if signer.PublicKey().Type() != "ssh-ed25519" || len(fp) < len("SHA256:") {
		t.Fatalf("key %s %s", signer.PublicKey().Type(), fp)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("mode %o", info.Mode().Perm())
	}
	_, fp2, err := LoadOrCreateHostKey(path)
	if err != nil || fp2 != fp {
		t.Fatalf("reload %v %s", err, fp2)
	}
	loose := filepath.Join(dir, "loose")
	if err := os.WriteFile(loose, []byte("not a key"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := LoadOrCreateHostKey(loose); err == nil {
		t.Fatal("world-readable host key was accepted")
	}
}

func TestListenRequiresAddressAndKey(t *testing.T) {
	cfg := gamemode.SSHConfig{Enabled: true, HostKeyPath: filepath.Join(t.TempDir(), "k")}
	if _, err := Listen(cfg, Deps{}); err == nil {
		t.Fatal("empty listen")
	}
	cfg.Listen = "127.0.0.1:0"
	cfg.HostKeyPath = ""
	if _, err := Listen(cfg, Deps{}); err == nil {
		t.Fatal("empty host key")
	}
	off, err := Listen(gamemode.SSHConfig{}, Deps{})
	if err != nil || off.PublicInfo().Enabled || off.Addr() != "" {
		t.Fatalf("disabled %+v %v", off.PublicInfo(), err)
	}
}
