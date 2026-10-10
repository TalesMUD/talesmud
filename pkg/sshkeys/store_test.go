package sshkeys

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
)

func TestKeyStoreUniqueAndCap(t *testing.T) {
	client, err := dbsqlite.Open(t.TempDir() + "/keys.db")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	store, err := Open(client.DB())
	if err != nil {
		t.Fatal(err)
	}
	one := testSigner(t)
	two := testSigner(t)
	line := string(ssh.MarshalAuthorizedKey(one.PublicKey()))
	if _, _, err := ParseAuthorizedKey("-----BEGIN OPENSSH PRIVATE KEY-----\nsecret"); !errors.Is(err, ErrRejected) {
		t.Fatal(err)
	}
	row, err := store.Add("local:ada", line, "laptop\x1b[2J", "device", 1)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(row.Label, "\x1b") || row.Fingerprint == "" || row.UserRefID != "local:ada" || row.CreatedVia != "device" {
		t.Fatalf("%+v", row)
	}
	if _, err := store.Add("local:ada", line, "again", "device", 1); !errors.Is(err, ErrExists) && !errors.Is(err, ErrFull) {
		t.Fatalf("dup %v", err)
	}
	otherLine := string(ssh.MarshalAuthorizedKey(two.PublicKey()))
	if _, err := store.Add("local:ada", otherLine, "", "device", 1); !errors.Is(err, ErrFull) {
		t.Fatalf("cap %v", err)
	}
	if _, err := store.Add("local:bea", line, "", "device", 10); !errors.Is(err, ErrExists) {
		t.Fatalf("cross account %v", err)
	}
	got, err := store.ByFingerprint(row.Fingerprint)
	if err != nil || got == nil || got.ID != row.ID {
		t.Fatal(got, err)
	}
	list, err := store.List("local:ada")
	if err != nil || len(list) != 1 {
		t.Fatalf("list %d %v", len(list), err)
	}
	if err := store.Delete("local:bea", row.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if err := store.Delete("local:ada", row.ID); err != nil {
		t.Fatal(err)
	}
	if got, err := store.ByFingerprint(row.Fingerprint); err != nil || got != nil {
		t.Fatal("still linked")
	}
}

func TestWebPasteDoesNotSquat(t *testing.T) {
	client, err := dbsqlite.Open(t.TempDir() + "/squat.db")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	store, err := Open(client.DB())
	if err != nil {
		t.Fatal(err)
	}
	line := string(ssh.MarshalAuthorizedKey(testSigner(t).PublicKey()))
	if _, err := store.Add("local:attacker", line, "squat", "web", 10); !errors.Is(err, ErrRejected) {
		t.Fatalf("web paste stored a key: %v", err)
	}
	pub, fp, err := ParseAuthorizedKey(line)
	if err != nil || pub == nil || fp == "" {
		t.Fatal(err)
	}
	got, err := store.ByFingerprint(fp)
	if err != nil || got != nil {
		t.Fatalf("paste occupied the fingerprint: %+v %v", got, err)
	}
	list, err := store.List("local:attacker")
	if err != nil || len(list) != 0 {
		t.Fatalf("attacker list %v %v", list, err)
	}
}

func TestPendingOffer(t *testing.T) {
	p := NewPending(time.Minute)
	if _, ok := p.Offer("SHA256:abc", "local:ada"); !ok {
		t.Fatal("offer")
	}
	if ref, _, ok := p.Get("SHA256:abc"); !ok || ref != "local:ada" {
		t.Fatal(ref, ok)
	}
	if _, ok := p.Offer("SHA256:abc", "local:bea"); ok {
		t.Fatal("overwrote another account")
	}
	if ref, _, ok := p.Get("SHA256:abc"); !ok || ref != "local:ada" {
		t.Fatalf("overwrite stuck %s %v", ref, ok)
	}
	if _, ok := p.Offer("SHA256:abc", "local:ada"); !ok {
		t.Fatal("same account refresh")
	}
	p.Drop("SHA256:abc")
	if _, _, ok := p.Get("SHA256:abc"); ok {
		t.Fatal("dropped")
	}
}

func testSigner(t *testing.T) ssh.Signer {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	return signer
}
