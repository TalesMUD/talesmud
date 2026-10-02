package mudserver

import "testing"

func TestOriginAllowed(t *testing.T) {
	allowed := []string{"https://talesmud.io", "http://localhost:8010"}
	if !OriginAllowed("", allowed) {
		t.Fatal("empty origin should be allowed for non-browser clients")
	}
	if !OriginAllowed("https://talesmud.io", allowed) {
		t.Fatal("listed origin should be allowed")
	}
	if OriginAllowed("https://evil.example", allowed) {
		t.Fatal("unknown origin should be rejected")
	}
}
