package repository

import (
	"database/sql"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
)

const contentHealthStateID = "state"

type sqliteContentHealthRepository struct {
	db *sql.DB
}

// NewSQLiteContentHealthRepository stores the content-health baseline in one row.
func NewSQLiteContentHealthRepository(client *dbsqlite.Client) ContentHealthRepository {
	return &sqliteContentHealthRepository{db: client.DB()}
}

func (repo *sqliteContentHealthRepository) Get() ([]byte, error) {
	var data string
	err := repo.db.QueryRow(
		`SELECT data FROM content_health WHERE id = ?`,
		contentHealthStateID,
	).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return []byte(data), nil
}

func (repo *sqliteContentHealthRepository) Save(payload []byte) error {
	_, err := repo.db.Exec(
		`INSERT OR REPLACE INTO content_health (id, data) VALUES (?, ?)`,
		contentHealthStateID,
		string(payload),
	)
	return err
}
