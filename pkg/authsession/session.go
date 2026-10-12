// Package authsession issues short-lived server sessions for confirmed device logins.
package authsession

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	e "github.com/talesmud/talesmud/pkg/entities"
)

const (
	Issuer   = "talesmud-server-session"
	Audience = "talesmud-api"
	TTL      = 30 * time.Minute
)

var (
	ErrCredentials = errors.New("invalid server session")
	ErrBanned      = errors.New("account banned")
)

type Users interface {
	FindByID(string) (*e.User, error)
	FindByRefID(string) (*e.User, error)
}

type Service struct {
	secret []byte
	users  Users
}

type claims struct {
	jwt.RegisteredClaims
	Kind string `json:"kind"`
	SV   int    `json:"sv"`
}

// New uses the existing SESSION_SECRET / game-mode session key resolver.
func New(secret string, users Users) *Service {
	return &Service{secret: []byte(secret), users: users}
}

// Issue only accepts an existing, eligible confirmed account; it never creates users.
func (s *Service) Issue(ref string) (string, int, error) {
	if s == nil || len(s.secret) == 0 || s.users == nil {
		return "", 0, ErrCredentials
	}
	user, err := s.users.FindByRefID(ref)
	if err != nil || user == nil || user.IsGuest || user.IsBanned || user.ID == "" {
		return "", 0, ErrCredentials
	}
	now := time.Now()
	c := claims{RegisteredClaims: jwt.RegisteredClaims{
		Subject: user.ID, Issuer: Issuer, Audience: jwt.ClaimStrings{Audience},
		IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(TTL)),
	}, Kind: "device-session", SV: user.SessionVersion}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(s.secret)
	return token, int(TTL / time.Second), err
}

// IsIssuer is only a routing hint. UserFromToken always verifies the signature.
func IsIssuer(raw string) bool {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	var peek struct {
		Iss  string `json:"iss"`
		Kind string `json:"kind"`
	}
	if json.Unmarshal(payload, &peek) != nil {
		return false
	}
	return peek.Iss == Issuer || peek.Kind == "device-session"
}

func (s *Service) UserFromToken(raw string) (*e.User, error) {
	if s == nil || len(s.secret) == 0 || s.users == nil {
		return nil, ErrCredentials
	}
	c := &claims{}
	token, err := jwt.ParseWithClaims(raw, c, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrCredentials
		}
		return s.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(Issuer), jwt.WithAudience(Audience), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil || token == nil || !token.Valid || c.Issuer != Issuer ||
		c.Kind != "device-session" || c.Subject == "" || c.ExpiresAt == nil || !c.ExpiresAt.After(time.Now()) {
		return nil, ErrCredentials
	}
	user, err := s.users.FindByID(c.Subject)
	if err != nil || user == nil || user.IsGuest || user.SessionVersion != c.SV {
		return nil, ErrCredentials
	}
	if user.IsBanned {
		return nil, ErrBanned
	}
	return user, nil
}
