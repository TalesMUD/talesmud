package service

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	log "github.com/sirupsen/logrus"
	e "github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/repository"
)

// GuestStatsSummary holds aggregate statistics returned to the admin dashboard.
type GuestStatsSummary struct {
	TotalAllTime int                       `json:"totalAllTime"`
	Today        int                       `json:"today"`
	ThisWeek     int                       `json:"thisWeek"`
	ActiveNow    int                       `json:"activeNow"`
	Daily        []*repository.DailyGuestStat `json:"daily"`
}

// GuestStatsService tracks guest session analytics without storing personal data.
type GuestStatsService interface {
	RecordSessionStart(guestRefID string) error
	RecordSessionEnd(guestRefID string, endedAt time.Time) error
	GetSummary() (*GuestStatsSummary, error)
	GetDailyStats(days int) ([]*repository.DailyGuestStat, error)
}

type guestStatsService struct {
	repo repository.GuestStatsRepository
}

// NewGuestStatsService creates a new guest statistics service.
func NewGuestStatsService(repo repository.GuestStatsRepository) GuestStatsService {
	return &guestStatsService{repo: repo}
}

func generateStatID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "gstat_" + hex.EncodeToString(b)
}

// RecordSessionStart records the start of a new guest session.
func (s *guestStatsService) RecordSessionStart(guestRefID string) error {
	now := time.Now().UTC()
	session := &e.GuestStatSession{
		ID:         generateStatID(),
		GuestRefID: guestRefID,
		Date:       now.Format("2006-01-02"),
		StartedAt:  now,
	}
	if err := s.repo.Store(session); err != nil {
		log.WithError(err).Warn("GuestStats: failed to record session start")
		return err
	}
	return nil
}

// RecordSessionEnd marks an existing session as ended and records its duration.
// If the session is not found (e.g., server restart), this is a no-op.
func (s *guestStatsService) RecordSessionEnd(guestRefID string, endedAt time.Time) error {
	session, err := s.repo.FindByGuestRefID(guestRefID)
	if err != nil {
		// Session not found — likely due to a server restart or stats recording failure. Not an error.
		return nil
	}
	duration := int64(endedAt.Sub(session.StartedAt).Seconds())
	if duration < 0 {
		duration = 0
	}
	if err := s.repo.MarkEnded(session.ID, endedAt, duration); err != nil {
		log.WithError(err).Warn("GuestStats: failed to record session end")
		return err
	}
	return nil
}

// GetSummary returns aggregate guest statistics for the admin dashboard.
func (s *guestStatsService) GetSummary() (*GuestStatsSummary, error) {
	total, err := s.repo.CountAll()
	if err != nil {
		return nil, err
	}

	active, err := s.repo.CountActive()
	if err != nil {
		return nil, err
	}

	daily, err := s.repo.GetDailyStats(30)
	if err != nil {
		return nil, err
	}

	// Derive today and this-week counts from the daily slice (last entry = today).
	today := 0
	if len(daily) > 0 {
		today = daily[len(daily)-1].Count
	}

	thisWeek := 0
	start := len(daily) - 7
	if start < 0 {
		start = 0
	}
	for _, d := range daily[start:] {
		thisWeek += d.Count
	}

	return &GuestStatsSummary{
		TotalAllTime: total,
		Today:        today,
		ThisWeek:     thisWeek,
		ActiveNow:    active,
		Daily:        daily,
	}, nil
}

// GetDailyStats returns per-day session counts for the last N days.
func (s *guestStatsService) GetDailyStats(days int) ([]*repository.DailyGuestStat, error) {
	return s.repo.GetDailyStats(days)
}
