package textline

import (
	"strings"
	"testing"
)

func TestSanitizeStripsBidiAndFormat(t *testing.T) {
	payload := "hi \x1b]52;c;SGVsbG8=\x07 \u202eBIDI\u2066iso\u200bzw\r\nNEXT"
	got := Sanitize(payload)
	if strings.Contains(got, "\x1b") || strings.ContainsRune(got, '\u202e') || strings.ContainsRune(got, '\u2066') || strings.ContainsRune(got, '\u200b') {
		t.Fatalf("controls survived %q", got)
	}
	if !strings.Contains(got, "\r\nNEXT") {
		t.Fatalf("multiline sanitize %q", got)
	}
	line := SingleLine("Vic\r\nFAKE-ACCOUNT \u202eBob")
	if strings.ContainsAny(line, "\r\n") || strings.ContainsRune(line, '\u202e') {
		t.Fatalf("label %q", line)
	}
	if line != "VicFAKE-ACCOUNT Bob" {
		t.Fatalf("label %q", line)
	}
}
