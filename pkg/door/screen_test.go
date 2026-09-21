package door

import (
	"strings"
	"testing"
)

func TestFrameIs80x25(t *testing.T) {
	body := []string{
		hotkey("F", "The Bramblewood", "walks left"),
		"Soot-warm cobbles and a gate that counts your days.",
	}
	frame := view{
		Location:  "Ashmarket Square",
		Body:      body,
		Prompt:    "Your command, Mara Quinn? ",
		Footer:    "Mara Quinn  Lv 1  HP 30/30  coin 50  vault 0  walks 15/15  12:00",
		InputMode: "hotkey",
	}.frame()
	if frame.Type != "door_frame" || frame.Cols != 80 || frame.Rows != 25 {
		t.Fatalf("frame meta = %+v", frame)
	}
	if !strings.HasPrefix(frame.ANSI, "\x1b[2J\x1b[H") {
		t.Fatal("frame does not clear the screen")
	}
	payload := strings.TrimPrefix(frame.ANSI, "\x1b[2J\x1b[H")
	lines := strings.Split(payload, "\r\n")
	if len(lines) != Rows {
		t.Fatalf("rows = %d", len(lines))
	}
	for i, ln := range lines {
		if n := visibleLen(ln); n != Cols {
			t.Fatalf("line %d visible width %d: %q", i, n, stripANSI(ln))
		}
	}
	plain := stripANSI(frame.ANSI)
	for _, banned := range []string{"Legend of the Red Dragon", "Seth Able", "Turgon", "Abdul"} {
		if strings.Contains(plain, banned) {
			t.Fatalf("frame contains locked IP text %q", banned)
		}
	}
}
