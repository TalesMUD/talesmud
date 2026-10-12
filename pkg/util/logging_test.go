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
	if got := RedactAccessToken("/api/device/poll?device_code=POLLSECRET&token=SESSION"); got != "/api/device/poll?device_code=[REDACTED]&token=[REDACTED]" {
		t.Fatal("device secrets leaked in access log")
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

func TestRedactSecretQueryKeysWithDuplicates(t *testing.T) {
	in := "/ws?TICKET=one&ticket=two&access_token=jwt&code=callback&activate=device&device_code=poll&token=session&user_code=VISIBLE&my_ticket=VISIBLE"
	want := "/ws?TICKET=[REDACTED]&ticket=[REDACTED]&access_token=[REDACTED]&code=[REDACTED]&activate=[REDACTED]&device_code=[REDACTED]&token=[REDACTED]&user_code=VISIBLE&my_ticket=VISIBLE"
	if got := RedactAccessToken(in); got != want {
		t.Fatalf("secret query values leaked or non-secret keys changed: %q", got)
	}
}
