package authlocal

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"

	"github.com/talesmud/talesmud/pkg/gamemode"
)

// ResolveSecret returns the HMAC key for local session tokens.
// DOOR_SESSION_SECRET wins. Otherwise a key file is created on first use.
func ResolveSecret(cfg gamemode.Config) (string, error) {
	if s := strings.TrimSpace(os.Getenv("DOOR_SESSION_SECRET")); s != "" {
		return s, nil
	}
	if s := strings.TrimSpace(cfg.SessionSecret); s != "" {
		return s, nil
	}
	path := cfg.SecretPath
	if path == "" {
		path = "data/door.session.key"
	}
	if existing, err := os.ReadFile(path); err == nil {
		s := strings.TrimSpace(string(existing))
		if s != "" {
			return s, nil
		}
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	encoded := hex.EncodeToString(buf)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(encoded+"\n"), 0o600); err != nil {
		return "", err
	}
	return encoded, nil
}
