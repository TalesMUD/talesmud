package termout

import (
	"strings"
	"testing"
)

func TestEncodeDropsOSCAndKeepsPackCSI(t *testing.T) {
	payload := "hi \x1b]52;c;SGVsbG8=\x07 \x1b]0;pwned\x07 \x1bPdcspayload\x1b\\ \x9b31m \u202eB \x1b[32mgreen\x1b[0m \x1b[2J\x1b[H"
	for _, cs := range []Charset{UTF8, CP437} {
		got := string(Encode(payload, cs))
		for _, bad := range []string{"\x1b]52", "\x1b]0;", "\x1bP"} {
			if strings.Contains(got, bad) {
				t.Fatalf("%s kept %q in %q", cs, bad, got)
			}
		}
		if strings.ContainsRune(got, '\u202e') || strings.Contains(got, "\x9b") {
			t.Fatalf("%s kept c1 or bidi %q", cs, got)
		}
		if !strings.Contains(got, "\x1b[32m") || !strings.Contains(got, "\x1b[0m") {
			t.Fatalf("%s dropped SGR %q", cs, got)
		}
		if !strings.Contains(got, "\x1b[2J") || !strings.Contains(got, "\x1b[H") {
			t.Fatalf("%s dropped pack CSI %q", cs, got)
		}
	}
}
