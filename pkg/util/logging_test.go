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
	leaked := "/activate?code=BCDF-GHJK /play/?activate=BCDF-GHJK&access_token=SUPERSECRETTOKEN"
	got = RedactAccessToken(leaked)
	want = "/activate?code=[REDACTED] /play/?activate=[REDACTED]&access_token=[REDACTED]"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	shaped := RedactAccessToken("/play/?user_code=KEEP-ME&code=DROP-ME&activate=ALSO-DROP")
	if shaped != "/play/?user_code=KEEP-ME&code=[REDACTED]&activate=[REDACTED]" {
		t.Fatalf("query boundary: %s", shaped)
	}
}
