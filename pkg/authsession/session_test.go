package authsession

import (
	"errors"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	e "github.com/talesmud/talesmud/pkg/entities"
)

type testUsers struct{ user *e.User }

func (u *testUsers) FindByID(id string) (*e.User, error) {
	if u.user != nil && u.user.ID == id {
		return u.user, nil
	}
	return nil, errors.New("missing")
}
func (u *testUsers) FindByRefID(ref string) (*e.User, error) {
	if u.user != nil && u.user.RefID == ref {
		return u.user, nil
	}
	return nil, errors.New("missing")
}

func TestServerSessionsVerifyAndRevoke(t *testing.T) {
	users := &testUsers{&e.User{Entity: &e.Entity{ID: "account-id"}, RefID: "auth0|account"}}
	s := New("test-server-secret", users)
	raw, ttl, err := s.Issue(users.user.RefID)
	if err != nil || ttl != 1800 || !IsIssuer(raw) {
		t.Fatal("session issuance failed")
	}
	if user, err := s.UserFromToken(raw); err != nil || user != users.user {
		t.Fatal("account session rejected")
	}
	if _, err := New("other-secret", users).UserFromToken(raw); err != ErrCredentials {
		t.Fatal("bad signature accepted")
	}
	if _, err := (*Service)(nil).UserFromToken(raw); err != ErrCredentials {
		t.Fatal("unconfigured verifier accepted")
	}
	users.user.SessionVersion++
	if _, err := s.UserFromToken(raw); err != ErrCredentials {
		t.Fatal("revoked session accepted")
	}
	users.user.SessionVersion--
	users.user.IsBanned = true
	if _, err := s.UserFromToken(raw); err != ErrBanned {
		t.Fatal("banned session accepted")
	}
	if _, _, err := s.Issue(users.user.RefID); err != ErrCredentials {
		t.Fatal("banned account issued token")
	}
	users.user.IsBanned = false
	users.user.IsGuest = true
	if _, err := s.UserFromToken(raw); err != ErrCredentials {
		t.Fatal("guest session accepted")
	}
	if _, _, err := s.Issue(users.user.RefID); err != ErrCredentials {
		t.Fatal("guest issued token")
	}
	if _, _, err := s.Issue("missing"); err != ErrCredentials {
		t.Fatal("unknown account issued token")
	}
}

func TestServerSessionRequiresSignedClaims(t *testing.T) {
	users := &testUsers{&e.User{Entity: &e.Entity{ID: "account-id"}, RefID: "auth0|account"}}
	s := New("test-server-secret", users)
	for _, name := range []string{"missing exp", "expired", "wrong audience", "wrong issuer", "wrong kind", "missing subject", "future issued", "wrong algorithm"} {
		t.Run(name, func(t *testing.T) {
			c := jwt.MapClaims{"sub": "account-id", "iss": Issuer, "aud": Audience, "kind": "device-session", "sv": 0, "exp": time.Now().Add(TTL).Unix()}
			method := jwt.SigningMethodHS256
			switch name {
			case "missing exp":
				delete(c, "exp")
			case "expired":
				c["exp"] = time.Now().Add(-time.Second).Unix()
			case "wrong audience":
				c["aud"] = "elsewhere"
			case "wrong issuer":
				c["iss"] = "elsewhere"
			case "wrong kind":
				c["kind"] = "guest"
			case "missing subject":
				delete(c, "sub")
			case "future issued":
				c["iat"] = time.Now().Add(time.Hour).Unix()
			case "wrong algorithm":
				method = jwt.SigningMethodHS384
			}
			raw, err := jwt.NewWithClaims(method, c).SignedString(s.secret)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.UserFromToken(raw); err != ErrCredentials {
				t.Fatal("invalid claims accepted")
			}
		})
	}
}
