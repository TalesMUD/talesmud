package devicecode

import (
	"errors"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	maxCandidates = 5
	maxField      = 64
)

var (
	// ErrNotFound is the only code error returned to clients.
	ErrNotFound = errors.New("code not found or expired")
	// ErrForbidden is a banned or guest account.
	ErrForbidden = errors.New("account cannot confirm")
	// ErrLimited is too many attempts or too many open codes.
	ErrLimited = errors.New("too many attempts")
)

// Config is the device-code policy. Zero values get the documented defaults.
type Config struct {
	TTL             time.Duration
	MaxPendingPerIP int
	MaxPending      int // whole process, not one address
	PerUser         int
	PerIP           int
	Window          time.Duration
}

// View is what a logged-in account may see. KeyFP is the offered fingerprint.
type View struct {
	IP            string
	Created       time.Time
	Mode          string
	ClientVersion string
	KeyFP         string
	CSRF          string
}

type request struct {
	id      string
	hash    string
	ip      string
	mode    string
	client  string
	created time.Time
	expires time.Time
	status  string // pending, confirmed, denied, consumed
	userRef string
	fps     []string
	csrf    map[string]string
}

// Store holds pending device codes for one process.
type Store struct {
	mu       sync.Mutex
	now      func() time.Time
	cfg      Config
	byID     map[string]*request
	byHash   map[string]string
	userHits map[string][]time.Time
	ipHits   map[string][]time.Time
}

// New builds a store. now may be nil.
func New(cfg Config) *Store {
	if cfg.TTL <= 0 {
		cfg.TTL = 10 * time.Minute
	}
	if cfg.MaxPendingPerIP <= 0 {
		cfg.MaxPendingPerIP = 3
	}
	if cfg.MaxPending <= 0 {
		cfg.MaxPending = 100
	}
	if cfg.PerUser <= 0 {
		cfg.PerUser = 10
	}
	if cfg.PerIP <= 0 {
		cfg.PerIP = 20
	}
	if cfg.Window <= 0 {
		cfg.Window = 10 * time.Minute
	}
	return &Store{
		now:      time.Now,
		cfg:      cfg,
		byID:     map[string]*request{},
		byHash:   map[string]string{},
		userHits: map[string][]time.Time{},
		ipHits:   map[string][]time.Time{},
	}
}

// Begin opens one code for an SSH lobby. display is the only copy of the code.
func (s *Store) Begin(ip, mode, client string, fps []string) (id, display string, expires time.Time, err error) {
	if s == nil {
		return "", "", time.Time{}, ErrLimited
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeLocked()
	if s.pendingTotal() >= s.cfg.MaxPending || s.pendingCount(ip) >= s.cfg.MaxPendingPerIP {
		return "", "", time.Time{}, ErrLimited
	}
	norm, err := randomCode()
	if err != nil {
		return "", "", time.Time{}, err
	}
	rid, err := randomID()
	if err != nil {
		return "", "", time.Time{}, err
	}
	now := s.now()
	req := &request{
		id:      rid,
		hash:    Hash(norm),
		ip:      clip(ip, maxField),
		mode:    clip(mode, 16),
		client:  clip(client, maxField),
		created: now,
		expires: now.Add(s.cfg.TTL),
		status:  "pending",
		fps:     clipFPs(fps),
		csrf:    map[string]string{},
	}
	s.byID[req.id] = req
	s.byHash[req.hash] = req.id
	return req.id, Format(norm), req.expires, nil
}

// State reports pending, confirmed, denied, consumed, expired, or missing.
func (s *Store) State(id string) string {
	if s == nil {
		return "missing"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeLocked()
	req := s.byID[id]
	if req == nil {
		return "missing"
	}
	return req.status
}

// Cancel drops a lobby that quit or continued as a guest.
func (s *Store) Cancel(id string) {
	if s == nil || id == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if req := s.byID[id]; req != nil && req.status == "pending" {
		req.status = "denied"
	}
}

// Take returns the confirmed account once.
func (s *Store) Take(id string) (string, bool) {
	if s == nil {
		return "", false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeLocked()
	req := s.byID[id]
	if req == nil || req.status != "confirmed" || req.userRef == "" {
		return "", false
	}
	req.status = "consumed"
	return req.userRef, true
}

// Lookup checks a code for a signed-in account and returns a csrf nonce.
func (s *Store) Lookup(code, userRef, ip string, allow func(string) error) (View, error) {
	if err := allow(userRef); err != nil {
		return View{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	req, err := s.findLocked(code, userRef, ip)
	if err != nil {
		return View{}, err
	}
	token := req.csrf[userRef]
	if token == "" {
		token, err = randomID()
		if err != nil {
			return View{}, err
		}
		req.csrf[userRef] = token
	}
	return s.view(req, token), nil
}

// Confirm binds the account. The waiting lobby takes the result once.
func (s *Store) Confirm(code, userRef, ip, csrf string, allow func(string) error) error {
	if err := allow(userRef); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	req, err := s.findLocked(code, userRef, ip)
	if err != nil {
		return err
	}
	if req.status != "pending" || req.csrf[userRef] == "" || req.csrf[userRef] != csrf {
		return ErrNotFound
	}
	req.status = "confirmed"
	req.userRef = userRef
	return nil
}

// Deny consumes a pending code.
func (s *Store) Deny(code, userRef, ip, csrf string, allow func(string) error) error {
	if err := allow(userRef); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	req, err := s.findLocked(code, userRef, ip)
	if err != nil {
		return err
	}
	if req.status != "pending" || req.csrf[userRef] == "" || req.csrf[userRef] != csrf {
		return ErrNotFound
	}
	req.status = "denied"
	return nil
}

func (s *Store) findLocked(code, userRef, ip string) (*request, error) {
	norm, ok := Normalize(code)
	if !ok || userRef == "" {
		return nil, ErrNotFound
	}
	s.purgeLocked()
	if !s.hit(s.userHits, userRef, s.cfg.PerUser) || !s.hit(s.ipHits, LimitKey(ip), s.cfg.PerIP) {
		return nil, ErrLimited
	}
	req := s.byID[s.byHash[Hash(norm)]]
	if req == nil || req.status != "pending" || !s.now().Before(req.expires) {
		return nil, ErrNotFound
	}
	return req, nil
}

func (s *Store) view(req *request, csrf string) View {
	fp := ""
	if len(req.fps) > 0 {
		fp = req.fps[0]
	}
	return View{
		IP:            req.ip,
		Created:       req.created,
		Mode:          req.mode,
		ClientVersion: req.client,
		KeyFP:         fp,
		CSRF:          csrf,
	}
}

func (s *Store) pendingCount(ip string) int {
	key := LimitKey(ip)
	n := 0
	now := s.now()
	for _, req := range s.byID {
		if LimitKey(req.ip) == key && req.status == "pending" && now.Before(req.expires) {
			n++
		}
	}
	return n
}

func (s *Store) pendingTotal() int {
	n := 0
	now := s.now()
	for _, req := range s.byID {
		if req.status == "pending" && now.Before(req.expires) {
			n++
		}
	}
	return n
}

// LimitKey buckets IPv6 by /64. IPv4, including IPv4-mapped IPv6, stays one address.
// The stored address on a code stays the raw peer for the confirm page.
func LimitKey(ip string) string {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		if strings.TrimSpace(ip) == "" {
			return "-"
		}
		return ip
	}
	if v4 := parsed.To4(); v4 != nil {
		return v4.String()
	}
	return parsed.Mask(net.CIDRMask(64, 128)).String() + "/64"
}

func (s *Store) purgeLocked() {
	now := s.now()
	for id, req := range s.byID {
		if !now.Before(req.expires) && req.status == "pending" {
			req.status = "expired"
			delete(s.byHash, req.hash)
			delete(s.byID, id)
		}
	}
}

func (s *Store) hit(book map[string][]time.Time, key string, limit int) bool {
	if key == "" {
		key = "-"
	}
	now := s.now()
	cut := now.Add(-s.cfg.Window)
	kept := book[key][:0]
	for _, ts := range book[key] {
		if ts.After(cut) {
			kept = append(kept, ts)
		}
	}
	if len(kept) >= limit {
		book[key] = kept
		return false
	}
	book[key] = append(kept, now)
	return true
}

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) > n {
		return s[:n]
	}
	return s
}

func clipFPs(fps []string) []string {
	if len(fps) == 0 {
		return nil
	}
	out := make([]string, 0, len(fps))
	for _, fp := range fps {
		fp = strings.TrimSpace(fp)
		if fp == "" {
			continue
		}
		if len(fp) > 128 {
			fp = fp[:128]
		}
		out = append(out, fp)
		if len(out) >= maxCandidates {
			break
		}
	}
	return out
}
