// Package resources is a per-character balance store.
// Allowances come from configuration. A balance refills when its calendar day
// or interval bucket changes. Nothing is configured until a caller says so,
// so an empty store never grants a use.
package resources

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	// ResetCalendar refills on the calendar date in the allowance timezone.
	ResetCalendar = "calendar"
	// ResetInterval refills when the wall clock crosses a fixed interval.
	ResetInterval = "interval"
)

var (
	// ErrExhausted is returned when a consume would drop the balance below zero.
	ErrExhausted = errors.New("resource exhausted")
	// ErrUnknown is returned when the key is not configured.
	ErrUnknown = errors.New("resource not configured")
)

// Allowance is the configured budget for one key.
type Allowance struct {
	Key      string
	Amount   int
	Reset    string
	Timezone string
	Interval time.Duration
}

// Resource is one character's balance for the current period.
type Resource struct {
	ID          string    `json:"id"`
	CharacterID string    `json:"characterId"`
	Key         string    `json:"key"`
	Period      string    `json:"period"`
	Allowance   int       `json:"allowance"`
	Remaining   int       `json:"remaining"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Modifier adjusts the allowance used on the next refill.
// It does not change a balance already open in the current period.
type Modifier func(characterID, key string, allowance int) int

// Store persists balances in SQLite.
type Store struct {
	db   *sql.DB
	now  func() time.Time
	mu   sync.Mutex
	spec map[string]Allowance
	mods []Modifier
}

// New opens a store on db and ensures the table exists.
func New(db *sql.DB) (*Store, error) {
	if db == nil {
		return nil, errors.New("resource store requires a database")
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS character_resources (id TEXT PRIMARY KEY, data TEXT NOT NULL);`); err != nil {
		return nil, fmt.Errorf("create character_resources: %w", err)
	}
	return &Store{
		db:   db,
		now:  time.Now,
		spec: map[string]Allowance{},
	}, nil
}

// SetNow overrides the clock. Tests use this to cross a period boundary.
func (s *Store) SetNow(now func() time.Time) {
	if s == nil || now == nil {
		return
	}
	s.mu.Lock()
	s.now = now
	s.mu.Unlock()
}

// Configure replaces the allowance catalog. Keys that are no longer listed
// stop refilling; rows already stored are left in place and are not readable
// until the key is configured again.
func (s *Store) Configure(list []Allowance) {
	if s == nil {
		return
	}
	next := map[string]Allowance{}
	for _, a := range list {
		a.Key = strings.TrimSpace(a.Key)
		if a.Key == "" {
			continue
		}
		if a.Amount < 0 {
			a.Amount = 0
		}
		next[a.Key] = a
	}
	s.mu.Lock()
	s.spec = next
	s.mu.Unlock()
}

// AddModifier registers a refill adjustment. Modifiers run in registration order.
func (s *Store) AddModifier(m Modifier) {
	if s == nil || m == nil {
		return
	}
	s.mu.Lock()
	s.mods = append(s.mods, m)
	s.mu.Unlock()
}

// Get returns the current balance for a configured key, refilling when the
// period has changed. The boolean is false when the key is not configured.
func (s *Store) Get(characterID, key string) (Resource, bool, error) {
	if s == nil {
		return Resource{}, false, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.getLocked(characterID, key)
}

// Consume subtracts n from the current period. n <= 0 returns the balance unchanged.
func (s *Store) Consume(characterID, key string, n int) (Resource, error) {
	if s == nil {
		return Resource{}, ErrUnknown
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	res, ok, err := s.getLocked(characterID, key)
	if err != nil {
		return res, err
	}
	if !ok {
		return res, ErrUnknown
	}
	if n <= 0 {
		return res, nil
	}
	if res.Remaining < n {
		return res, ErrExhausted
	}
	res.Remaining -= n
	res.UpdatedAt = s.now()
	if err := s.save(res); err != nil {
		return res, err
	}
	return res, nil
}

func (s *Store) getLocked(characterID, key string) (Resource, bool, error) {
	characterID = strings.TrimSpace(characterID)
	key = strings.TrimSpace(key)
	if characterID == "" || key == "" {
		return Resource{}, false, nil
	}
	allowance, ok := s.spec[key]
	if !ok {
		return Resource{}, false, nil
	}
	now := s.now()
	period, err := periodOf(allowance, now)
	if err != nil {
		return Resource{}, false, err
	}
	effective := s.effective(characterID, allowance)
	id := characterID + "|" + key
	res, err := s.load(id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return Resource{}, false, err
		}
		res = Resource{
			ID:          id,
			CharacterID: characterID,
			Key:         key,
			Period:      period,
			Allowance:   effective,
			Remaining:   effective,
		}
	} else if res.Period != period {
		res.Period = period
		res.Allowance = effective
		res.Remaining = effective
	} else {
		res.Allowance = effective
	}
	res.UpdatedAt = now
	if err := s.save(res); err != nil {
		return Resource{}, false, err
	}
	return res, true, nil
}

func (s *Store) effective(characterID string, allowance Allowance) int {
	n := allowance.Amount
	if n < 0 {
		n = 0
	}
	for _, mod := range s.mods {
		if mod == nil {
			continue
		}
		n = mod(characterID, allowance.Key, n)
		if n < 0 {
			n = 0
		}
	}
	return n
}

func periodOf(allowance Allowance, now time.Time) (string, error) {
	switch strings.TrimSpace(allowance.Reset) {
	case "", ResetCalendar:
		return now.In(location(allowance.Timezone)).Format("2006-01-02"), nil
	case ResetInterval:
		if allowance.Interval < time.Second {
			return "", fmt.Errorf("interval reset for %s requires an interval of at least 1s", allowance.Key)
		}
		step := int64(allowance.Interval / time.Second)
		return fmt.Sprintf("i:%d", now.Unix()/step), nil
	default:
		return "", fmt.Errorf("unknown reset %q", allowance.Reset)
	}
}

func location(name string) *time.Location {
	name = strings.TrimSpace(name)
	if name == "" || strings.EqualFold(name, "UTC") {
		return time.UTC
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}

func (s *Store) load(id string) (Resource, error) {
	var payload string
	err := s.db.QueryRow(`SELECT data FROM character_resources WHERE id = ?`, id).Scan(&payload)
	if err != nil {
		return Resource{}, err
	}
	var res Resource
	if err := json.Unmarshal([]byte(payload), &res); err != nil {
		return Resource{}, err
	}
	return res, nil
}

func (s *Store) save(res Resource) error {
	payload, err := json.Marshal(res)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		`INSERT INTO character_resources (id, data) VALUES (?, ?)
		 ON CONFLICT(id) DO UPDATE SET data = excluded.data`,
		res.ID, string(payload),
	)
	return err
}
