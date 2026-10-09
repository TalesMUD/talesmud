package sshgate

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

// LoadOrCreateHostKey loads an ed25519 host key or creates one.
// The directory is 0700 and the file is 0600. A group- or world-readable
// file is refused. The key bytes are never logged.
func LoadOrCreateHostKey(path string) (ssh.Signer, string, error) {
	path = filepath.Clean(path)
	if path == "" || path == "." {
		return nil, "", errors.New("ssh host key path is empty")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, "", fmt.Errorf("ssh host key dir: %w", err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return nil, "", fmt.Errorf("ssh host key dir mode: %w", err)
	}
	info, err := os.Stat(path)
	if err == nil {
		if info.Mode().Perm()&0o077 != 0 {
			return nil, "", fmt.Errorf("ssh host key %s is group or world readable", path)
		}
		return parseHostKey(path)
	}
	if !os.IsNotExist(err) {
		return nil, "", err
	}
	if err := writeNewHostKey(path); err != nil {
		return nil, "", err
	}
	return parseHostKey(path)
}

func writeNewHostKey(path string) error {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	block, err := ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := pem.Encode(f, block); err != nil {
		return err
	}
	return f.Chmod(0o600)
}

func parseHostKey(path string) (ssh.Signer, string, error) {
	// #nosec G304 -- path is the operator host_key_path, never request input.
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	key, err := ssh.ParseRawPrivateKey(raw)
	if err != nil {
		return nil, "", errors.New("ssh host key could not be parsed")
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		return nil, "", err
	}
	if signer.PublicKey().Type() != ssh.KeyAlgoED25519 {
		return nil, "", errors.New("ssh host key must be ed25519")
	}
	return signer, ssh.FingerprintSHA256(signer.PublicKey()), nil
}
