package ansi

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestRenderIncludesLocationAndFixedSize(t *testing.T) {
	frame := Render(Page{
		Title:    "Sample",
		Location: "Market",
		Body:     []string{"A stall.", "Exits: north"},
		Prompt:   ">",
		Footer:   "n north",
	})
	if frame.Type != "door_frame" || frame.Cols != 80 || frame.Rows != 25 {
		t.Fatalf("%+v", frame)
	}
	if !strings.Contains(frame.ANSI, "Market") || !strings.Contains(frame.ANSI, "A stall.") {
		t.Fatalf("frame missing body:\n%s", frame.ANSI)
	}
	lines := 1
	for _, r := range frame.ANSI {
		if r == '\n' {
			lines++
		}
	}
	if lines != 25 {
		t.Fatalf("lines = %d", lines)
	}
}

func TestClipKeepsMultibyteBarsIntact(t *testing.T) {
	line := strings.Repeat("═", 90)
	frame := Render(Page{
		Title:    "Sample",
		Location: "Market",
		Body:     []string{line},
		Prompt:   ">",
		Footer:   "n north",
	})
	if !utf8.ValidString(frame.ANSI) {
		t.Fatal("frame split a multibyte character")
	}
	if strings.Contains(frame.ANSI, "\uFFFD") {
		t.Fatal("frame inserted a replacement character")
	}
	if strings.Count(frame.ANSI, "═") != 80 {
		t.Fatalf("bar runes = %d", strings.Count(frame.ANSI, "═"))
	}
}

func TestRailsReplacesTheDashedRule(t *testing.T) {
	classic := Render(Page{Title: "Sample", Location: "Market", Body: []string{"A stall."}})
	if !strings.Contains(classic.ANSI, strings.Repeat("-", 80)) {
		t.Fatal("classic frame lost the dashed rule")
	}
	frame := Render(Page{
		Style:    StyleRails,
		Title:    "Sample",
		Location: "Market",
		Status:   "\x1b[0;37mLevel \x1b[1;37m3\x1b[0;37m   HP \x1b[1;32m11/20\x1b[0m",
		Body:     []string{"A stall."},
		Prompt:   ">",
		Footer:   "n north",
	})
	if frame.Cols != 80 || frame.Rows != 25 {
		t.Fatalf("%+v", frame)
	}
	if strings.Contains(frame.ANSI, strings.Repeat("-", 40)) {
		t.Fatal("rails kept a dashed rule")
	}
	if !strings.Contains(frame.ANSI, "Market") || !strings.Contains(frame.ANSI, "11/20") || !strings.Contains(frame.ANSI, "▀") || !strings.Contains(frame.ANSI, "░") {
		t.Fatalf("rails frame:\n%s", frame.ANSI)
	}
	if !utf8.ValidString(frame.ANSI) {
		t.Fatal("rails frame is not utf-8")
	}
	lines := strings.Split(strings.TrimPrefix(frame.ANSI, "\x1b[2J\x1b[H"), "\r\n")
	if len(lines) != 25 {
		t.Fatalf("lines = %d", len(lines))
	}
}
