package server

import (
	"strings"
	"testing"
)

func TestBrandDoorPageUsesTheConfigTitle(t *testing.T) {
	page := `<title>TalesMUD Door</title><h1 id="title">TalesMUD Door</h1><p class="sub" id="sub">A text client on TalesMUD.</p>`
	got := brandDoorPage(page, "Sample Door", "A sample subtitle.")
	if strings.Count(got, "Sample Door") != 2 || strings.Contains(got, "TalesMUD Door") {
		t.Fatalf("title: %s", got)
	}
	if !strings.Contains(got, "A sample subtitle.") {
		t.Fatalf("subtitle: %s", got)
	}
	same := brandDoorPage(page, "TalesMUD Door", "A text client on TalesMUD.")
	if same != page {
		t.Fatalf("defaults changed: %s", same)
	}
	escaped := brandDoorPage(page, `A & B`, "ok")
	if !strings.Contains(escaped, "A &amp; B") {
		t.Fatalf("title was not escaped: %s", escaped)
	}
}
