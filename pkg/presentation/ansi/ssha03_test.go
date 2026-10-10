package ansi

import (
	"strings"
	"testing"
)

func TestRenderDropsOSCAndC1(t *testing.T) {
	payload := "hi \x1b]52;c;SGVsbG8=\x07 \x1b]0;pwned\x07 \x1b]8;;http://evil.example\x1b\\ \x1bPdcspayload\x1b\\ \x9b31mC1 \u009bUTF8C1 \u202eBIDI\r\nNEXT \x1b[32mgreen\x1b[0m"
	frame := Render(Page{
		Title:    "Room \x1b]0;title\x07",
		Location: "Hall \x1b]0;title\x07",
		Body:     []string{payload, "exit \x1b]52;c;QQ==\x07"},
	})
	raw := frame.ANSI
	for _, bad := range []string{"\x1b]52", "\x1b]0;", "\x1b]8;;", "\x1bP", "\x9b31m"} {
		if strings.Contains(raw, bad) {
			t.Fatalf("frame kept %q", bad)
		}
	}
	if strings.ContainsRune(raw, '\u202e') || strings.ContainsRune(raw, '\u009b') {
		t.Fatal("frame kept bidi or C1")
	}
	if strings.Contains(raw, "\r\nNEXT") {
		t.Fatal("frame kept a label line break")
	}
	if !strings.Contains(raw, "\x1b[32mgreen\x1b[0m") {
		t.Fatal("SGR was dropped")
	}
	if !strings.Contains(raw, "\x1b[2J\x1b[H") {
		t.Fatal("frame clear was dropped")
	}
	if strings.Contains(frame.ANSI, "\x1b]0;title") {
		t.Fatal("title kept OSC")
	}
}
