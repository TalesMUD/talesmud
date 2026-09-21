// Package daily is the first-class daily budget store.
// Quest tags are not used for fight allowances.
package daily

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

const forestFightsKey = "forest_fights"

// ForestFightsKey is the budget consumed by a walk into the forest.
const ForestFightsKey = forestFightsKey

// ErrExhausted is returned when a consume would drop the balance below zero.
var ErrExhausted = errors.New("daily resource exhausted")

// Resource is one character's balance for a calendar day in the server timezone.
type Resource struct {
	ID          string    `json:"id"`
	CharacterID string    `json:"characterId"`
	Key         string    `json:"key"`
	Day         string    `json:"day"`
	Allowance   int       `json:"allowance"`
	Remaining   int       `json:"remaining"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Store persists daily resources in SQLite.
type Store struct {
	db  *sql.DB
	loc *time.Location
	now func() time.Time
}

// New opens a store. loc controls the calendar day used for rollover.
func New(db *sql.DB, loc *time.Location) *Store {
	if loc == nil {
		loc = time.UTC
	}
	s := &Store{db: db, loc: loc, now: time.Now}
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS daily_resources (id TEXT PRIMARY KEY, data TEXT NOT NULL);`)
	return s
}

// SetNow overrides the clock. Tests use this to cross midnight.
func (s *Store) SetNow(now func() time.Time) {
	if now != nil {
		s.now = now
	}
}

// Get returns today's balance, refilling when the calendar day changes.
func (s *Store) Get(characterID, key string, allowance int) (Resource, error) {
	if allowance < 0 {
		allowance = 0
	}
	today := s.today()
	res, err := s.load(idFor(characterID, key))
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return Resource{}, err
		}
		res = Resource{
			ID:          idFor(characterID, key),
			CharacterID: characterID,
			Key:         key,
			Day:         today,
			Allowance:   allowance,
			Remaining:   allowance,
		}
	}
	if res.Day != today {
		res.Day = today
		res.Allowance = allowance
		res.Remaining = allowance
	}
	res.UpdatedAt = s.now()
	if err := s.save(res); err != nil {
		return Resource{}, err
	}
	return res, nil
}

// Consume subtracts n from today's balance.
func (s *Store) Consume(characterID, key string, allowance, n int) (Resource, error) {
	res, err := s.Get(characterID, key, allowance)
	if err != nil {
		return res, err
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
		return Resource{}, err
	}
	return res, nil
}

func (s *Store) today() string {
	return s.now().In(s.loc).Format("2006-01-02")
}

func idFor(characterID, key string) string {
	return characterID + "|" + key
}

func (s *Store) load(id string) (Resource, error) {
	var payload string
	if err := s.db.QueryRow(`SELECT data FROM daily_resources WHERE id = ?`, id).Scan(&payload); err != nil {
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
		`INSERT INTO daily_resources (id, data) VALUES (?, ?)
		 ON CONFLICT(id) DO UPDATE SET data = excluded.data`,
		res.ID, string(payload),
	)
	return err
}
