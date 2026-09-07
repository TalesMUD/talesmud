package util

import "testing"

func TestRedactAccessToken(t *testing.T) {
	in := "/ws?access_token=eyJhbGciOi.secret.sig&x=1"
	got := RedactAccessToken(in)
	want := "/ws?access_token=[REDACTED]&x=1"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	ticketIn := "/ws?ticket=abc123&x=1"
	ticketGot := RedactAccessToken(ticketIn)
	if ticketGot != "/ws?ticket=[REDACTED]&x=1" {
		t.Fatalf("ticket redact got %q", ticketGot)
	}
	if RedactAccessToken("/api/rooms") != "/api/rooms" {
		t.Fatal("expected untouched path")
	}
}

func TestRequestPathWithoutQuery(t *testing.T) {
	if got := RequestPathWithoutQuery("/ws?ticket=secret"); got != "/ws" {
		t.Fatalf("got %q", got)
	}
	if got := RequestPathWithoutQuery("/api/rooms"); got != "/api/rooms" {
		t.Fatalf("got %q", got)
	}
}
