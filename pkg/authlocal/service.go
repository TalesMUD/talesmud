package authlocal

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	e "github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/service"
)

const (
	issuer         = "talesmud-local"
	sessionTTL     = 7 * 24 * time.Hour
	resetTTL       = 30 * time.Minute
	minPasswordLen = 8
	maxPasswordLen = 128
)

var (
	ErrValidation  = errors.New("invalid registration")
	ErrExists      = errors.New("account already exists")
	ErrCredentials = errors.New("invalid username or password")
	ErrBanned      = errors.New("account banned")
	ErrToken       = errors.New("invalid or expired token")

	usernamePattern = regexp.MustCompile(`^[a-z0-9_]{3,20}$`)
)

// PublicUser is the account view safe to return from HTTP.
type PublicUser struct {
	ID            string `json:"id"`
	Username      string `json:"username,omitempty"`
	Email         string `json:"email,omitempty"`
	Nickname      string `json:"nickname,omitempty"`
	Role          string `json:"role"`
	LastCharacter string `json:"lastCharacter,omitempty"`
}

// Public strips secrets from a user record.
func Public(u *e.User) PublicUser {
	if u == nil {
		return PublicUser{}
	}
	return PublicUser{
		ID:            u.ID,
		Username:      u.Username,
		Email:         u.Email,
		Nickname:      u.Nickname,
		Role:          u.GetRole(),
		LastCharacter: u.LastCharacter,
	}
}

type resetRecord struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
	Used      bool      `json:"used"`
}

type sessionClaims struct {
	jwt.StandardClaims
	SV   int    `json:"sv"`
	Kind string `json:"kind"`
}

// Service registers and authenticates local username/password accounts.
type Service struct {
	db      *sql.DB
	users   service.UsersService
	secret  []byte
	mailer  Mailer
	now     func() time.Time
	dummy   string
	dummyMu sync.Once
}

// New builds a local auth service. secret is the HMAC session key.
func New(db *sql.DB, users service.UsersService, secret string, mailer Mailer) *Service {
	if mailer == nil {
		mailer = OutboxMailer{}
	}
	s := &Service{
		db:     db,
		users:  users,
		secret: []byte(secret),
		mailer: mailer,
		now:    time.Now,
	}
	_ = s.ensureTable()
	return s
}

func (s *Service) ensureTable() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS auth_tokens (id TEXT PRIMARY KEY, data TEXT NOT NULL);`)
	return err
}

// Register creates a user and returns a session token. Passwords are stored as Argon2id.
func (s *Service) Register(username, email, password string) (string, PublicUser, error) {
	username, email, err := normalizeIdentity(username, email, password)
	if err != nil {
		return "", PublicUser{}, err
	}
	if _, err := s.users.FindByUsername(username); err == nil {
		return "", PublicUser{}, ErrExists
	}
	if _, err := s.users.FindByEmail(email); err == nil {
		return "", PublicUser{}, ErrExists
	}
	hash, err := Hash(password)
	if err != nil {
		return "", PublicUser{}, err
	}
	user := &e.User{
		Name:         username,
		Nickname:     username,
		Email:        email,
		Username:     username,
		PasswordHash: hash,
		Role:         e.RolePlayer,
		RefID:        "local:" + username,
		Created:      s.now().UTC(),
		IsNewUser:    false,
	}
	created, err := s.users.Create(user)
	if err != nil {
		return "", PublicUser{}, err
	}
	token, err := s.sign(created)
	if err != nil {
		return "", PublicUser{}, err
	}
	return token, Public(created), nil
}

// Login verifies a password and returns a session token.
func (s *Service) Login(username, password string) (string, PublicUser, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	if username == "" || password == "" {
		s.burn(password)
		return "", PublicUser{}, ErrCredentials
	}
	user, err := s.users.FindByUsername(username)
	if err != nil || user.PasswordHash == "" {
		s.burn(password)
		return "", PublicUser{}, ErrCredentials
	}
	if !Verify(password, user.PasswordHash) {
		return "", PublicUser{}, ErrCredentials
	}
	if user.IsBanned {
		return "", PublicUser{}, ErrBanned
	}
	token, err := s.sign(user)
	if err != nil {
		return "", PublicUser{}, err
	}
	return token, Public(user), nil
}

// Forgot stores a hashed one-time token and asks the mailer to deliver the raw value.
// Unknown emails return nil so callers can answer every request the same way.
func (s *Service) Forgot(email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil
	}
	user, err := s.users.FindByEmail(email)
	if err != nil || user.PasswordHash == "" {
		return nil
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return err
	}
	token := hex.EncodeToString(raw)
	rec := resetRecord{
		ID:        tokenHash(token),
		UserID:    user.ID,
		ExpiresAt: s.now().Add(resetTTL),
		Used:      false,
	}
	if err := s.storeToken(rec); err != nil {
		return err
	}
	if err := s.mailer.SendPasswordReset(user.Email, token); err != nil {
		_ = s.deleteToken(rec.ID)
		return err
	}
	return nil
}

// Reset sets a new password from a one-time token and revokes existing sessions.
func (s *Service) Reset(rawToken, newPassword string) error {
	rawToken = strings.TrimSpace(rawToken)
	if len(newPassword) < minPasswordLen || len(newPassword) > maxPasswordLen {
		return ErrValidation
	}
	rec, err := s.loadToken(tokenHash(rawToken))
	if err != nil || rec.Used || s.now().After(rec.ExpiresAt) {
		return ErrToken
	}
	user, err := s.users.FindByID(rec.UserID)
	if err != nil {
		return ErrToken
	}
	hash, err := Hash(newPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	user.SessionVersion++
	if err := s.users.Update(user.RefID, user); err != nil {
		return err
	}
	rec.Used = true
	return s.storeToken(*rec)
}

// UserFromToken loads the user for a local session JWT.
func (s *Service) UserFromToken(tokenStr string) (*e.User, error) {
	claims, err := s.parse(tokenStr)
	if err != nil {
		return nil, err
	}
	user, err := s.users.FindByID(claims.Subject)
	if err != nil {
		return nil, ErrCredentials
	}
	if user.SessionVersion != claims.SV {
		return nil, ErrCredentials
	}
	if user.IsBanned {
		return nil, ErrBanned
	}
	return user, nil
}

// IsLocalIssuer reports whether the unverified JWT payload claims to be ours.
// Invalid signatures still fail closed in UserFromToken.
func IsLocalIssuer(tokenStr string) bool {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return false
	}
	raw, err := jwt.DecodeSegment(parts[1])
	if err != nil {
		return false
	}
	var peek struct {
		Iss  string `json:"iss"`
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(raw, &peek); err != nil {
		return false
	}
	return peek.Iss == issuer || peek.Kind == "local"
}

func (s *Service) sign(user *e.User) (string, error) {
	now := s.now()
	claims := sessionClaims{
		StandardClaims: jwt.StandardClaims{
			Subject:   user.ID,
			Issuer:    issuer,
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(sessionTTL).Unix(),
		},
		SV:   user.SessionVersion,
		Kind: "local",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *Service) parse(tokenStr string) (*sessionClaims, error) {
	claims := &sessionClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid || claims.Issuer != issuer || claims.Kind != "local" || claims.Subject == "" {
		return nil, ErrCredentials
	}
	return claims, nil
}

func (s *Service) burn(password string) {
	s.dummyMu.Do(func() {
		encoded, err := Hash("not-a-real-account-password")
		if err != nil {
			s.dummy = ""
			return
		}
		s.dummy = encoded
	})
	if s.dummy != "" && password != "" {
		Verify(password, s.dummy)
	}
}

func normalizeIdentity(username, email, password string) (string, string, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	email = strings.ToLower(strings.TrimSpace(email))
	if !usernamePattern.MatchString(username) {
		return "", "", ErrValidation
	}
	if _, err := mail.ParseAddress(email); err != nil || !strings.Contains(email, ".") {
		return "", "", ErrValidation
	}
	if len(password) < minPasswordLen || len(password) > maxPasswordLen {
		return "", "", ErrValidation
	}
	return username, email, nil
}

func tokenHash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (s *Service) storeToken(rec resetRecord) error {
	payload, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		`INSERT INTO auth_tokens (id, data) VALUES (?, ?)
		 ON CONFLICT(id) DO UPDATE SET data = excluded.data`,
		rec.ID, string(payload),
	)
	return err
}

func (s *Service) loadToken(id string) (*resetRecord, error) {
	var payload string
	if err := s.db.QueryRow(`SELECT data FROM auth_tokens WHERE id = ?`, id).Scan(&payload); err != nil {
		return nil, err
	}
	var rec resetRecord
	if err := json.Unmarshal([]byte(payload), &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (s *Service) deleteToken(id string) error {
	_, err := s.db.Exec(`DELETE FROM auth_tokens WHERE id = ?`, id)
	return err
}
