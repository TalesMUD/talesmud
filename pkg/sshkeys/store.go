// Package sshkeys stores linked SSH public keys and short-lived pending links.
// A pending link is not a linked key. The key row is written only after the
// client proves it holds the private key and the account owner confirms.
package sshkeys

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

const maxLabel = 64

var (
	ErrExists   = errors.New("key already linked")
	ErrFull     = errors.New("too many keys")
	ErrRejected = errors.New("key rejected")
	ErrNotFound = errors.New("key not found")
)

// Key is one linked public key. PublicKey is an authorized_keys line.
type Key struct {
	ID          string    `json:"id"`
	UserRefID   string    `json:"user_ref_id"`
	Fingerprint string    `json:"fingerprint"`
	KeyType     string    `json:"key_type"`
	PublicKey   string    `json:"public_key"`
	Label       string    `json:"label"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedVia  string    `json:"created_via"`
	LastUsedAt  time.Time `json:"last_used_at,omitempty"`
	LastUsedIP  string    `json:"last_used_ip,omitempty"`
}

// Store is the ssh_keys table. The schema is created with the rest of the DB.
type Store struct {
	db *sql.DB
	mu sync.Mutex
}

// Open uses the process SQLite connection.
func Open(db *sql.DB) (*Store, error) {
	if db == nil {
		return nil, errors.New("nil db")
	}
	return &Store{db: db}, nil
}

// ParseAuthorizedKey accepts one public authorized_keys line.
func ParseAuthorizedKey(line string) (ssh.PublicKey, string, error) {
	line = strings.TrimSpace(line)
	if line == "" || strings.Contains(line, "PRIVATE KEY") || strings.Contains(line, "\n") {
		return nil, "", ErrRejected
	}
	pub, _, _, _, err := ssh.ParseAuthorizedKey([]byte(line))
	if err != nil {
		return nil, "", ErrRejected
	}
	switch pub.Type() {
	case ssh.KeyAlgoED25519, ssh.KeyAlgoSKED25519, ssh.KeyAlgoECDSA256, ssh.KeyAlgoECDSA384, ssh.KeyAlgoECDSA521, ssh.KeyAlgoRSA:
	default:
		return nil, "", ErrRejected
	}
	return pub, ssh.FingerprintSHA256(pub), nil
}

// Add links a public key the client proved during device sign-in.
// via must be "device". A web paste is refused so it cannot occupy the
// unique fingerprint index or authenticate as this account.
func (s *Store) Add(userRef, line, label, via string, max int) (*Key, error) {
	if s == nil || strings.TrimSpace(userRef) == "" || via != "device" {
		return nil, ErrRejected
	}
	pub, fp, err := ParseAuthorizedKey(line)
	if err != nil {
		return nil, err
	}
	if max <= 0 {
		max = 10
	}
	label = cleanLabel(label)
	row := &Key{
		ID:          uuid.NewString(),
		UserRefID:   userRef,
		Fingerprint: fp,
		KeyType:     pub.Type(),
		PublicKey:   strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pub))),
		Label:       label,
		CreatedAt:   time.Now().UTC(),
		CreatedVia:  "device",
	}
	raw, err := json.Marshal(row)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	n, err := s.count(userRef)
	if err != nil {
		return nil, err
	}
	if n >= max {
		return nil, ErrFull
	}
	if existing, err := s.byFP(fp); err != nil {
		return nil, err
	} else if existing != nil {
		return nil, ErrExists
	}
	if _, err := s.db.Exec(`INSERT INTO ssh_keys (id, data) VALUES (?, ?)`, row.ID, string(raw)); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return nil, ErrExists
		}
		return nil, err
	}
	return row, nil
}

// List returns the account's keys, oldest first.
func (s *Store) List(userRef string) ([]Key, error) {
	if s == nil {
		return nil, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	rows, err := s.db.Query(`SELECT data FROM ssh_keys WHERE json_extract(data, '$.user_ref_id') = ? ORDER BY json_extract(data, '$.created_at')`, userRef)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Key
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var key Key
		if err := json.Unmarshal([]byte(raw), &key); err != nil {
			return nil, err
		}
		out = append(out, key)
	}
	return out, rows.Err()
}

// ByFingerprint returns the linked key, or nil when it is unknown.
func (s *Store) ByFingerprint(fp string) (*Key, error) {
	if s == nil || strings.TrimSpace(fp) == "" {
		return nil, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.byFP(fp)
}

// Delete removes one of the account's keys and returns its fingerprint.
func (s *Store) Delete(userRef, id string) (string, error) {
	if s == nil {
		return "", ErrNotFound
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key, err := s.byID(id)
	if err != nil {
		return "", err
	}
	if key == nil || key.UserRefID != userRef {
		return "", ErrNotFound
	}
	fp := key.Fingerprint
	res, err := s.db.Exec(`DELETE FROM ssh_keys WHERE id = ?`, id)
	if err != nil {
		return "", err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return "", ErrNotFound
	}
	return fp, nil
}

// Touch records a successful key login. The address is stored, not logged here.
func (s *Store) Touch(fp, ip string, at time.Time) error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key, err := s.byFP(fp)
	if err != nil || key == nil {
		return err
	}
	key.LastUsedAt = at.UTC()
	key.LastUsedIP = ip
	raw, err := json.Marshal(key)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`UPDATE ssh_keys SET data = ? WHERE id = ?`, string(raw), key.ID)
	return err
}

func (s *Store) count(userRef string) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM ssh_keys WHERE json_extract(data, '$.user_ref_id') = ?`, userRef).Scan(&n)
	return n, err
}

func (s *Store) byFP(fp string) (*Key, error) {
	var raw string
	err := s.db.QueryRow(`SELECT data FROM ssh_keys WHERE json_extract(data, '$.fingerprint') = ?`, fp).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var key Key
	if err := json.Unmarshal([]byte(raw), &key); err != nil {
		return nil, err
	}
	return &key, nil
}

func (s *Store) byID(id string) (*Key, error) {
	var raw string
	err := s.db.QueryRow(`SELECT data FROM ssh_keys WHERE id = ?`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var key Key
	if err := json.Unmarshal([]byte(raw), &key); err != nil {
		return nil, err
	}
	return &key, nil
}

func cleanLabel(s string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(s) {
		if r == '\t' || (r >= 0x20 && r != 0x7f && !(r >= 0x80 && r <= 0x9f)) {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if len(out) > maxLabel {
		out = out[:maxLabel]
	}
	return out
}
