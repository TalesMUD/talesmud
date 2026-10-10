package sshgate

import (
	"crypto/ed25519"
	"net"
	"testing"

	"golang.org/x/crypto/ssh"

	"github.com/talesmud/talesmud/pkg/devicecode"
)

// stubMeta is enough ConnMetadata for noteSigned. It is not a connection.
type stubMeta struct {
	id []byte
}

func (s stubMeta) User() string          { return "player" }
func (s stubMeta) SessionID() []byte     { return s.id }
func (s stubMeta) ClientVersion() []byte { return []byte("SSH-2.0-test") }
func (s stubMeta) ServerVersion() []byte { return []byte("SSH-2.0-TalesMUD") }
func (s stubMeta) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("203.0.113.50"), Port: 22}
}
func (s stubMeta) LocalAddr() net.Addr { return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 2222} }

func mintKey(t *testing.T) (line, fp string) {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	return string(ssh.MarshalAuthorizedKey(signer.PublicKey())), ssh.FingerprintSHA256(signer.PublicKey())
}

// TestStoredOfferIsTheKeyThatWillBeStored is SSHB-01. Two verified keys in
// one handshake used to show the first fingerprint and store the last.
func TestStoredOfferIsTheKeyThatWillBeStored(t *testing.T) {
	g := &Gate{}
	meta := stubMeta{id: []byte("handshake-1")}
	line1, fp1 := mintKey(t)
	line2, fp2 := mintKey(t)
	if fp1 == fp2 {
		t.Fatal("keys collided")
	}
	g.noteSigned(meta, fp1, line1)
	g.noteSigned(meta, fp2, line2)
	note := g.takeNote(meta)
	if len(note.fps) != 2 || note.fps[0] != fp1 || note.fps[1] != fp2 {
		t.Fatal("handshake did not keep both verified keys in order")
	}
	offers := storedOffers(note.line)
	if len(offers) != 1 || offers[0] != fp2 || offers[0] == fp1 {
		t.Fatal("confirm offer is not the last signed key")
	}
	if offers[0] != signedFingerprint(note.line, nil) {
		t.Fatal("confirm offer is not the key Y stores")
	}
	if signedFingerprint("", note.fps) != fp1 {
		t.Fatal("unsigned fallback is not the first offer, which the page must ignore")
	}
	if storedOffers("") != nil {
		t.Fatal("an empty line offered a key")
	}

	devices := devicecode.New(devicecode.Config{})
	_, code, _, err := devices.Begin("203.0.113.50", "door", "SSH-2.0-test", offers)
	if err != nil {
		t.Fatal(err)
	}
	view, err := devices.Lookup(code, "local:ada", "203.0.113.50", func(string) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if view.KeyFP != fp2 {
		t.Fatal("confirm page key is not the key Y stores")
	}
	_, firstCode, _, err := devices.Begin("198.51.100.50", "door", "SSH-2.0-test", note.fps)
	if err != nil {
		t.Fatal(err)
	}
	first, err := devices.Lookup(firstCode, "local:ada", "198.51.100.50", func(string) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if first.KeyFP != fp1 {
		t.Fatal("the old offer list no longer shows its first key")
	}
}
