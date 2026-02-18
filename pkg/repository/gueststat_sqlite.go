package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	e "github.com/talesmud/talesmud/pkg/entities"
)

type sqliteGuestStatsRepo struct {
	db *sql.DB
}

// NewSQLiteGuestStatsRepository returns a SQLite-backed guest statistics repository.
func NewSQLiteGuestStatsRepository(client *dbsqlite.Client) GuestStatsRepository {
	return &sqliteGuestStatsRepo{db: client.DB()}
}

func (r *sqliteGuestStatsRepo) Store(session *e.GuestStatSession) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(
		"INSERT INTO guest_statistics (id, data) VALUES (?, ?)",
		session.ID, string(data),
	)
	return err
}

func (r *sqliteGuestStatsRepo) FindByGuestRefID(refID string) (*e.GuestStatSession, error) {
	row := r.db.QueryRow(
		"SELECT data FROM guest_statistics WHERE json_extract(data, '$.guestRefID') = ? LIMIT 1",
		refID,
	)
	var payload string
	if err := row.Scan(&payload); err != nil {
		return nil, err
	}
	var session e.GuestStatSession
	if err := json.Unmarshal([]byte(payload), &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sqliteGuestStatsRepo) MarkEnded(id string, endedAt time.Time, durationSeconds int64) error {
	row := r.db.QueryRow("SELECT data FROM guest_statistics WHERE id = ?", id)
	var payload string
	if err := row.Scan(&payload); err != nil {
		return err
	}
	var session e.GuestStatSession
	if err := json.Unmarshal([]byte(payload), &session); err != nil {
		return err
	}
	session.EndedAt = &endedAt
	session.DurationSeconds = durationSeconds
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(
		"UPDATE guest_statistics SET data = ? WHERE id = ?",
		string(data), id,
	)
	return err
}

func (r *sqliteGuestStatsRepo) CountAll() (int, error) {
	row := r.db.QueryRow("SELECT COUNT(*) FROM guest_statistics")
	var count int
	err := row.Scan(&count)
	return count, err
}

// CountActive counts sessions that started within the last 35 minutes and have not yet ended.
// The 35-minute window covers the 30-minute guest session duration plus the 5-minute cleanup interval.
// This prevents orphaned entries (from server restarts) from inflating the count.
func (r *sqliteGuestStatsRepo) CountActive() (int, error) {
	cutoff := time.Now().UTC().Add(-35 * time.Minute).Format(time.RFC3339)
	row := r.db.QueryRow(
		`SELECT COUNT(*) FROM guest_statistics
		 WHERE json_extract(data, '$.endedAt') IS NULL
		 AND json_extract(data, '$.startedAt') > ?`,
		cutoff,
	)
	var count int
	err := row.Scan(&count)
	return count, err
}

// GetDailyStats returns session counts and average durations for the last N days (UTC).
func (r *sqliteGuestStatsRepo) GetDailyStats(days int) ([]*DailyGuestStat, error) {
	now := time.Now().UTC()
	stats := make([]*DailyGuestStat, 0, days)

	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")

		var count int
		row := r.db.QueryRow(
			"SELECT COUNT(*) FROM guest_statistics WHERE json_extract(data, '$.date') = ?",
			date,
		)
		if err := row.Scan(&count); err != nil {
			count = 0
		}

		var avgDuration float64
		row2 := r.db.QueryRow(
			`SELECT COALESCE(AVG(CAST(json_extract(data, '$.durationSeconds') AS REAL)), 0)
			 FROM guest_statistics
			 WHERE json_extract(data, '$.date') = ?
			 AND json_extract(data, '$.durationSeconds') > 0`,
			date,
		)
		_ = row2.Scan(&avgDuration)

		stats = append(stats, &DailyGuestStat{
			Date:               date,
			Count:              count,
			AvgDurationSeconds: avgDuration,
		})
	}
	return stats, nil
}
