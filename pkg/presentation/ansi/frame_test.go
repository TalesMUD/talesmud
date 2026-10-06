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
