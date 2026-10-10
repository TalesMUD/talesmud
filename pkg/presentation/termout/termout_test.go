package termout

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestCP437BlockBytes(t *testing.T) {
	got := Encode("\x1b[32m░▒▓█\x1b[0m", CP437)
	want := append([]byte("\x1b[32m"), 0xb0, 0xb1, 0xb2, 0xdb)
	want = append(want, []byte("\x1b[0m")...)
	if !bytes.Equal(got, want) {
		t.Fatalf("cp437 %x", got)
	}
	if !bytes.Equal(Encode("░▒▓█", ASCII), []byte(".:#@")) {
		t.Fatalf("ascii %q", Encode("░▒▓█", ASCII))
	}
	if !bytes.Equal(Encode("é", CP437), []byte{0x82}) {
		t.Fatalf("e-acute %x", Encode("é", CP437))
	}
	if !bytes.Equal(Encode("é", ASCII), []byte("?")) {
		t.Fatal("ascii unmappable")
	}
	if utf8.RuneCountInString(cp437Runes) != 128 || len(cp437Rev) != 128 {
		t.Fatalf("table runes %d map %d", utf8.RuneCountInString(cp437Runes), len(cp437Rev))
	}
	raw := make([]byte, 128)
	for i := range raw {
		raw[i] = byte(0x80 + i)
	}
	if !bytes.Equal(Encode(DecodeANS(raw), CP437), raw) {
		t.Fatal("cp437 round trip")
	}
}

func TestDetectCharset(t *testing.T) {
	if Detect("xterm-256color", "C", "", "", "cp437", "utf8") != CP437 {
		t.Fatal("tales charset wins")
	}
	if Detect("ansi-bbs", "en_US.UTF-8", "", "", "", "utf8") != CP437 {
		t.Fatal("ansi term")
	}
	if Detect("syncterm", "", "", "", "", "") != CP437 || Detect("pcansi", "", "", "", "", "") != CP437 {
		t.Fatal("bbs terms")
	}
	if Detect("dumb", "en_US.UTF-8", "", "", "", "") != ASCII {
		t.Fatal("dumb beats locale")
	}
	if Detect("vt52", "", "", "", "", "") != ASCII {
		t.Fatal("vt52")
	}
	if Detect("vt100", "en_US.UTF-8", "", "", "", "ascii") != UTF8 {
		t.Fatal("locale")
	}
	if Detect("xterm", "C", "", "", "", "ascii") != UTF8 {
		t.Fatal("modern term")
	}
	if Detect("vt100", "C", "", "", "", "cp437") != CP437 {
		t.Fatal("fallback")
	}
	if Detect("", "", "", "", "", "auto") != UTF8 {
		t.Fatal("default")
	}
}

func TestLetterbox100x30(t *testing.T) {
	blob := frameBlob()
	s := NewScreen()
	out := s.Paint(blob, 100, 30, UTF8, false)
	for _, seq := range []string{"\x1b[?1049h", "\x1b[?25l", "\x1b[?2026h", "\x1b[?2026l", "\x1b[3;11H"} {
		if !bytes.Contains(out, []byte(seq)) {
			t.Fatalf("missing %q in %q", seq, out)
		}
	}
	if !bytes.Contains(out, []byte("\x1b[3;11HROW-00")) {
		t.Fatalf("first row %q", out)
	}
	if !bytes.Contains(out, []byte("\x1b[4;11HROW-01")) {
		t.Fatal("second row")
	}
	if again := s.Paint(blob, 100, 30, UTF8, false); len(again) != 0 {
		t.Fatalf("unchanged paint %q", again)
	}
	forced := s.Paint(blob, 100, 30, UTF8, true)
	if !bytes.Contains(forced, []byte("\x1b[3;11HROW-00")) {
		t.Fatal("force")
	}
	top, left := Offsets(100, 30)
	if top != 2 || left != 10 {
		t.Fatalf("offsets %d %d", top, left)
	}
}

func TestDiffSkipsUnchangedRows(t *testing.T) {
	s := NewScreen()
	blob := frameBlob()
	if out := s.Paint(blob, 80, 25, UTF8, false); !bytes.Contains(out, []byte("ROW-00")) {
		t.Fatal(string(out))
	}
	lines := baseLines()
	lines[4] = "CHANGED-ROW"
	next := "\x1b[2J\x1b[H" + strings.Join(lines, "\r\n")
	out := s.Paint(next, 80, 25, UTF8, false)
	if bytes.Contains(out, []byte("ROW-00")) || !bytes.Contains(out, []byte("CHANGED-ROW")) {
		t.Fatalf("diff %q", out)
	}
}

func TestTooSmall70x20(t *testing.T) {
	s := NewScreen()
	out := s.Paint(frameBlob()+"DOOR-FRAME-MARKER", 70, 20, UTF8, false)
	if !bytes.Contains(out, []byte("70x20")) {
		t.Fatalf("notice %q", out)
	}
	if bytes.Contains(out, []byte("DOOR-FRAME-MARKER")) || bytes.Contains(out, []byte("ROW-00")) {
		t.Fatalf("painted the frame %q", out)
	}
	if again := s.Paint(frameBlob(), 70, 20, UTF8, false); len(again) != 0 {
		t.Fatal("repeat small")
	}
	big := s.Paint(frameBlob(), 100, 30, UTF8, false)
	if !bytes.Contains(big, []byte("\x1b[3;11HROW-00")) || bytes.Contains(big, []byte("70x20")) {
		t.Fatalf("resize %q", big)
	}
}

func TestEchoLineLetterbox(t *testing.T) {
	out := EchoLine("> hi", 100, 30, UTF8)
	if !bytes.Contains(out, []byte("\x1b[25;11H\x1b[2K> hi")) {
		t.Fatalf("%q", out)
	}
	if EchoLine("> hi", 70, 20, UTF8) != nil {
		t.Fatal("echo on a small window")
	}
}

func TestGenericSplashAndSubstitute(t *testing.T) {
	splash := GenericSplash()
	if !strings.Contains(splash, "TalesMUD") || !strings.Contains(splash, "do you see ░▒▓█ correctly?") {
		t.Fatal(splash)
	}
	if strings.Contains(strings.ToLower(splash), "aethermoor") {
		t.Fatal("world name in the engine splash")
	}
	got := Substitute("code {{CODE}} {{URL}} {{EXPIRES}}", "AB\x1b[2J12", "https://play.example/a", "10m")
	if strings.Contains(got, "\x1b") || !strings.Contains(got, "AB12") || !strings.Contains(got, "https://play.example/a") {
		t.Fatal(got)
	}
}

func TestSubstituteKeepsFrame(t *testing.T) {
	shade, bar := "░", "║"
	interior := strings.Repeat(" ", 27) + "\x1b[1;33;40m{{CODE}}\x1b[0;37;40m" + strings.Repeat(" ", 27)
	codeLine := strings.Repeat(shade, 8) + bar + interior + bar + strings.Repeat(shade, 8)
	if visibleLen(codeLine) != 80 {
		t.Fatalf("code fixture width %d", visibleLen(codeLine))
	}
	got := Substitute(codeLine, "ABCD-EFGH", "", "")
	if visibleLen(got) != 80 || !strings.Contains(got, "ABCD-EFGH") || !strings.HasSuffix(visibleText(got), shade) {
		t.Fatalf("code line %d %q", visibleLen(got), got)
	}
	expInterior := "expires {{EXPIRES}}"
	expPad := (62 - visibleLen(expInterior)) / 2
	expLine := strings.Repeat(shade, 8) + bar + strings.Repeat(" ", expPad) + expInterior + strings.Repeat(" ", 62-expPad-visibleLen(expInterior)) + bar + strings.Repeat(shade, 8)
	if visibleLen(expLine) != 80 {
		t.Fatalf("expires fixture %d", visibleLen(expLine))
	}
	got = Substitute(expLine, "", "", "9m30s")
	if visibleLen(got) != 80 || !strings.Contains(got, "9m30s") || strings.Contains(got, "{{") {
		t.Fatalf("expires %d %q", visibleLen(got), got)
	}
	url := "https://play.example/activate?code=ABCD-EFGH&signup=1"
	urlInterior := strings.Repeat(" ", 27) + "{{URL}}" + strings.Repeat(" ", 28)
	urlLine := strings.Repeat(shade, 8) + bar + urlInterior + bar + strings.Repeat(shade, 8)
	padInterior := strings.Repeat(" ", 62)
	padLine := strings.Repeat(shade, 8) + bar + padInterior + bar + strings.Repeat(shade, 8)
	screen := urlLine + "\n" + padLine
	got = Substitute(screen, "", url, "")
	parts := strings.Split(got, "\n")
	if len(parts) != 2 || visibleLen(parts[0]) != 80 || visibleLen(parts[1]) != 80 {
		t.Fatalf("url wrap widths %v", parts)
	}
	flat := visibleText(parts[0]) + visibleText(parts[1])
	if !strings.Contains(strings.ReplaceAll(flat, " ", ""), strings.ReplaceAll(url, " ", "")) {
		t.Fatalf("url missing in %q", flat)
	}
	if strings.Contains(got, "{{URL}}") {
		t.Fatal("placeholder remained")
	}
	signup := SubstituteSignup("{{SIGNUP_URL}}\n{{URL}}", "ABCD-EFGH", "http://login.example/a", "1m", "http://signup.example/new")
	if !strings.Contains(signup, "http://signup.example/new") || !strings.Contains(signup, "http://login.example/a") {
		t.Fatal(signup)
	}
}

func TestWrapTextWords(t *testing.T) {
	lines := WrapText("Remember this computer? Link key SHA256:abcd…wxyz to account [Y/N]", 24)
	for _, line := range lines {
		if visibleLen(line) > 24 {
			t.Fatalf("overwide %q", line)
		}
		if strings.Contains(line, "accoun") && !strings.Contains(line, "account") {
			t.Fatalf("split a word %q", line)
		}
	}
	if strings.Join(lines, " ") == "" || !strings.Contains(strings.Join(lines, " "), "Remember") {
		t.Fatal(lines)
	}
}

func visibleText(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == 0x1b {
			i = skipESC([]byte(s), i)
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if size < 1 {
			break
		}
		b.WriteRune(r)
		i += size
	}
	return b.String()
}

func TestLetterboxFill(t *testing.T) {
	s := NewScreen()
	s.SetFill("·")
	out := s.Paint(frameBlob(), 82, 26, UTF8, false)
	if !bytes.Contains(out, []byte("·")) {
		t.Fatal("fill missing")
	}
	s.SetFill("\x1b[31m")
	if s.fill != 0 {
		t.Fatal("control fill accepted")
	}
}

func frameBlob() string {
	return "\x1b[2J\x1b[H" + strings.Join(baseLines(), "\r\n")
}

func baseLines() []string {
	lines := make([]string, 25)
	for i := range lines {
		lines[i] = "ROW-" + string(rune('0'+i/10)) + string(rune('0'+i%10))
	}
	return lines
}
