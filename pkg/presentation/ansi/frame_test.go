package ansi

import (
	"strings"
	"testing"
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


