package sshgate

import (
	"strings"
	"testing"
	"time"
)

func TestDecorateURL(t *testing.T) {
	login := decorateURL("http://127.0.0.1:8031/activate", "BCDF-GHJK", false)
	if login != "http://127.0.0.1:8031/activate?code=BCDF-GHJK" {
		t.Fatal(login)
	}
	signup := decorateURL("http://127.0.0.1:8031/activate?from=ssh", "BCDF-GHJK", true)
	if !strings.Contains(signup, "code=BCDF-GHJK") || !strings.Contains(signup, "signup=1") || !strings.Contains(signup, "from=ssh") {
		t.Fatal(signup)
	}
	if decorateURL("the activate page", "BCDF-GHJK", true) != "the activate page" {
		t.Fatal("unparsed url was rewritten")
	}
}

func TestClassicLobbySignupCopy(t *testing.T) {
	when := time.Now().Add(time.Minute)
	off := classicLobbyText("BCDF-GHJK", "http://mud.example/activate?code=BCDF-GHJK", "", when, true)
	if strings.Contains(off, "[N]") || strings.Contains(off, "signup=1") || !strings.Contains(off, "New players: open the same page.") {
		t.Fatal(off)
	}
	if !strings.Contains(off, "[G] continue as guest") || !strings.Contains(off, "[L] I have an account") {
		t.Fatal(off)
	}
	on := classicLobbyText("BCDF-GHJK", "http://door.example/activate?code=BCDF-GHJK", "http://door.example/activate?code=BCDF-GHJK&signup=1", when, false)
	if !strings.Contains(on, "[N] New player") || !strings.Contains(on, "signup=1") || strings.Contains(on, "[G]") {
		t.Fatal(on)
	}
	framed := genericActivateTemplate(false, true)
	if strings.Contains(framed, "[N]") || strings.Contains(framed, "{{SIGNUP_URL}}") || !strings.Contains(framed, "[G] guest") {
		t.Fatal(framed)
	}
	open := genericActivateTemplate(true, false)
	if !strings.Contains(open, "[N] New player") || !strings.Contains(open, "{{SIGNUP_URL}}") || strings.Contains(open, "[G]") {
		t.Fatal(open)
	}
}
