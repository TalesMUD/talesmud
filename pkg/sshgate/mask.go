package sshgate

import "strings"

// MaskFingerprint shortens an OpenSSH SHA256 fingerprint for logs.
// SHA256:ab12…wxyz keeps the first and last four characters of the hash.
func MaskFingerprint(fp string) string {
	const prefix = "SHA256:"
	fp = strings.TrimSpace(fp)
	if !strings.HasPrefix(fp, prefix) {
		return prefix + "…"
	}
	rest := fp[len(prefix):]
	if len(rest) <= 8 {
		return prefix + "…"
	}
	return prefix + rest[:4] + "…" + rest[len(rest)-4:]
}

// MaskUserCode hides a device code. "BCDF-GHJK" becomes "BC**-****".
func MaskUserCode(code string) string {
	var compact strings.Builder
	for _, r := range strings.ToUpper(code) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			compact.WriteRune(r)
		}
	}
	raw := compact.String()
	if len(raw) < 2 {
		return "**-****"
	}
	return raw[:2] + "**-****"
}
