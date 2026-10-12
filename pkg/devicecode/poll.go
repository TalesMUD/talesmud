package devicecode

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

const PollInterval = 5 * time.Second

var (
	ErrPending  = errors.New("authorization_pending")
	ErrSlowDown = errors.New("slow_down")
	ErrDenied   = errors.New("access_denied")
	ErrExpired  = errors.New("expired_token")
)

// AppCode contains the only plaintext copies of the human code and poll secret.
type AppCode struct {
	UserCode   string
	DeviceCode string
	ExpiresIn  int
	Interval   int
}

// BeginApp uses the SSH store and confirmation flow, with a separate 256-bit
// poll secret. Public app codes live at most ten minutes regardless of SSH TTL.
func (s *Store) BeginApp(ip, name, version string) (AppCode, error) {
	if s == nil {
		return AppCode{}, ErrLimited
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeLocked()
	if !s.hit(s.startHits, "ip:"+limitKey(ip), s.cfg.StartsPerIP) ||
		!s.hit(s.startHits, "global", s.cfg.StartsGlobal) {
		return AppCode{}, ErrLimited
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return AppCode{}, err
	}
	id, display, _, err := s.beginLocked(ip, "app", version, nil)
	if err != nil {
		return AppCode{}, err
	}
	req := s.byID[id]
	req.kind = "app"
	req.name = clip(name, maxField)
	req.interval = PollInterval
	if max := req.created.Add(10 * time.Minute); req.expires.After(max) {
		req.expires = max
	}
	poll := hex.EncodeToString(secret)
	req.pollHash = Hash(poll)
	s.byPollHash[req.pollHash] = id
	return AppCode{display, poll, int((req.expires.Sub(req.created) + time.Second - 1) / time.Second), int(PollInterval / time.Second)}, nil
}

// Poll issues a confirmed account token under the store lock, then consumes the
// record exactly once. Issuer failures leave it confirmed for a later retry.
// Unknown, expired and consumed secrets share the same response.
func (s *Store) Poll(secret, ip string, issue func(string) (string, int, error)) (string, int, error) {
	if s == nil {
		return "", 0, ErrExpired
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeLocked()
	if !s.hitWindow(s.pollHits, limitKey(ip), s.cfg.PollsPerIP, time.Minute) {
		return "", 0, ErrSlowDown
	}
	if len(secret) != 64 {
		return "", 0, ErrExpired
	}
	req := s.byID[s.byPollHash[Hash(secret)]]
	if req == nil || req.kind != "app" || req.status == "consumed" {
		return "", 0, ErrExpired
	}
	if req.status == "denied" {
		return "", 0, ErrDenied
	}
	now := s.now()
	if now.Before(req.nextPoll) {
		req.interval += 5 * time.Second
		req.nextPoll = now.Add(req.interval)
		return "", 0, ErrSlowDown
	}
	req.nextPoll = now.Add(req.interval)
	if req.status != "confirmed" {
		return "", 0, ErrPending
	}
	if issue == nil {
		return "", 0, errors.New("session unavailable")
	}
	token, expires, err := issue(req.userRef)
	if err != nil {
		return "", 0, err
	}
	if token == "" || expires <= 0 {
		return "", 0, errors.New("session unavailable")
	}
	req.status = "consumed"
	req.userRef = ""
	return token, expires, nil
}
