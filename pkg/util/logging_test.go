package util

import "testing"

func TestRedactAccessToken(t *testing.T) {
	in := "/ws?access_token=eyJhbGciOi.secret.sig&x=1"
	got := RedactAccessToken(in)
	want := "/ws?access_token=[REDACTED]&x=1"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if RedactAccessToken("/api/rooms") != "/api/rooms" {
		t.Fatal("expected untouched path")
	}
}
