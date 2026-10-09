package sshgate

import (
	"errors"
	"sync"
	"time"

	"github.com/talesmud/talesmud/pkg/gamemode"
)

var (
	errBanned = errors.New("banned")
	errRate   = errors.New("rate")
	errCap    = errors.New("cap")
)

// limits tracks connection caps, new-connection rate, auth-fail bans, and
// the extra SSH guest caps. It does not replace GuestService's own limits.
type limits struct {
	mu      sync.Mutex
	cfg     gamemode.SSHConfig
	total   int
	perIP   map[string]int
	opened  map[string][]time.Time
	fails   map[string][]time.Time
	banned  map[string]time.Time
	guests  int
	guestIP map[string][]time.Time
}

func newLimits(cfg gamemode.SSHConfig) *limits {
	return &limits{
		cfg:     cfg,
		perIP:   map[string]int{},
		opened:  map[string][]time.Time{},
		fails:   map[string][]time.Time{},
		banned:  map[string]time.Time{},
		guestIP: map[string][]time.Time{},
	}
}

func (l *limits) acquire(ip string, now time.Time) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if until, ok := l.banned[ip]; ok {
		if now.Before(until) {
			return errBanned
		}
		delete(l.banned, ip)
	}
	window := now.Add(-time.Minute)
	recent := prune(l.opened[ip], window)
	if len(recent) >= l.cfg.NewConnsPerIPPerMin {
		l.opened[ip] = recent
		return errRate
	}
	if l.total >= l.cfg.MaxConnections {
		return errCap
	}
	if l.perIP[ip] >= l.cfg.MaxPerIP {
		return errCap
	}
	l.total++
	l.perIP[ip]++
	l.opened[ip] = append(recent, now)
	return nil
}

func (l *limits) release(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.total > 0 {
		l.total--
	}
	if n := l.perIP[ip]; n > 1 {
		l.perIP[ip] = n - 1
	} else {
		delete(l.perIP, ip)
	}
}

func (l *limits) authFail(ip string, now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	window := now.Add(-l.cfg.AuthFailBan.Window.Duration())
	fails := append(prune(l.fails[ip], window), now)
	l.fails[ip] = fails
	if len(fails) >= l.cfg.AuthFailBan.Failures {
		l.banned[ip] = now.Add(l.cfg.AuthFailBan.Ban.Duration())
		delete(l.fails, ip)
	}
}

func (l *limits) acquireGuest(ip string, now time.Time) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	window := now.Add(-time.Hour)
	hits := prune(l.guestIP[ip], window)
	if len(hits) >= l.cfg.Guest.PerIPPerHour {
		l.guestIP[ip] = hits
		return errRate
	}
	if l.guests >= l.cfg.Guest.MaxConcurrent {
		return errCap
	}
	l.guests++
	l.guestIP[ip] = append(hits, now)
	return nil
}

func (l *limits) releaseGuest(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.guests > 0 {
		l.guests--
	}
	_ = ip
}

func (l *limits) bannedNow(ip string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	until, ok := l.banned[ip]
	return ok && now.Before(until)
}

func prune(times []time.Time, after time.Time) []time.Time {
	out := times[:0]
	for _, ts := range times {
		if ts.After(after) {
			out = append(out, ts)
		}
	}
	return out
}
