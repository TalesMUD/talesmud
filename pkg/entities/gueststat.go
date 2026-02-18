package entities

import "time"

// GuestStatSession records a single guest session for analytics purposes.
// No personal data (IP, location, user-agent, etc.) is stored.
type GuestStatSession struct {
	ID              string     `json:"id"`
	GuestRefID      string     `json:"guestRefID"`      // links to the guest user RefID for cleanup tracking
	Date            string     `json:"date"`            // YYYY-MM-DD (UTC), for daily grouping
	StartedAt       time.Time  `json:"startedAt"`
	EndedAt         *time.Time `json:"endedAt,omitempty"`
	DurationSeconds int64      `json:"durationSeconds"` // 0 while session is still active
}
