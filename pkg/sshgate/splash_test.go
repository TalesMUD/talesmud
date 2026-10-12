package sshgate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePackFile(t *testing.T) {
	base := t.TempDir()
	inside := filepath.Join(base, "screens")
	if err := os.MkdirAll(inside, 0o755); err != nil {
		t.Fatal(err)
	}
	art := filepath.Join(inside, "splash.ans")
	if err := os.WriteFile(art, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := resolvePackFile(base, "screens/splash.ans"); got != art {
		t.Fatalf("inside %s", got)
	}
	if got := resolvePackFile(base, "../etc/passwd"); got != "" {
		t.Fatalf("escape %s", got)
	}
	outside := filepath.Join(t.TempDir(), "secret.ans")
	if err := os.WriteFile(outside, []byte("nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(inside, "linked.ans")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	if got := resolvePackFile(base, "screens/linked.ans"); got != "" {
		t.Fatalf("symlink %s", got)
	}
	abs := filepath.Join(t.TempDir(), "explicit.ans")
	if got := resolvePackFile(base, abs); got != abs {
		t.Fatalf("absolute %s", got)
	}
	if resolvePackFile(base, "") != "" || resolvePackFile(base, "bad\x00path") != "" {
		t.Fatal("empty")
	}
}
