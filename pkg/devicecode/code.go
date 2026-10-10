// Package devicecode is the in-memory device-code login.
// The SSH listener and the activate API share one process, so a restart
// drops pending codes. Codes are stored as hashes.
package devicecode

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// alphabet has no vowels, so a code is not a word, and no look-alike digits.
const alphabet = "BCDFGHJKLMNPQRSTVWXZ"

// Normalize accepts "XXXX-XXXX" or the same letters without the dash.
func Normalize(code string) (string, bool) {
	var b strings.Builder
	for _, r := range strings.ToUpper(strings.TrimSpace(code)) {
		if r == '-' || r == ' ' || r == '\t' {
			continue
		}
		if !strings.ContainsRune(alphabet, r) {
			return "", false
		}
		b.WriteRune(r)
	}
	if b.Len() != 8 {
		return "", false
	}
	return b.String(), true
}

// Format renders an 8-letter code as XXXX-XXXX.
func Format(norm string) string {
	if len(norm) != 8 {
		return ""
	}
	return norm[:4] + "-" + norm[4:]
}

// Hash is the stored form of a normalised code.
func Hash(norm string) string {
	sum := sha256.Sum256([]byte(norm))
	return hex.EncodeToString(sum[:])
}

func randomCode() (string, error) {
	const span = 240 // 20 * 12, so the modulo has no bias
	out := make([]byte, 8)
	buf := make([]byte, 1)
	for i := 0; i < len(out); {
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		if buf[0] >= span {
			continue
		}
		out[i] = alphabet[int(buf[0])%len(alphabet)]
		i++
	}
	return string(out), nil
}

func randomID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
