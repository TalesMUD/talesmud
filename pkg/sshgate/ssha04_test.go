package sshgate

import (
	"strings"
	"testing"

	"github.com/talesmud/talesmud/pkg/entities"
)

func TestAccountLabelStripsLineBreakAndBidi(t *testing.T) {
	got := accountLabel(&entities.User{Nickname: "Vic\r\nFAKE-ACCOUNT \u202eBob"})
	if strings.ContainsAny(got, "\r\n") || strings.ContainsRune(got, '\u202e') {
		t.Fatalf("%q", got)
	}
	if got != "VicFAKE-ACCOUNT Bob" {
		t.Fatalf("%q", got)
	}
	long := accountLabel(&entities.User{Nickname: strings.Repeat("á", 40)})
	if len([]rune(long)) != 32 {
		t.Fatalf("clipped %d %q", len([]rune(long)), long)
	}
}
