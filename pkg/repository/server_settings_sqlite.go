package repository

import (
	"encoding/json"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/entities/settings"
)

const serverSettingsID = "server-settings"

type sqliteServerSettingsRepository struct {
	*sqliteGenericRepo
}

// NewSQLiteServerSettingsRepository creates a new SQLite server settings repository.
func NewSQLiteServerSettingsRepository(client *dbsqlite.Client) ServerSettingsRepository {
	return &sqliteServerSettingsRepository{
		sqliteGenericRepo: newSQLiteGenericRepo(client.DB(), "server_settings", func() interface{} {
			return &settings.ServerSettings{}
		}),
	}
}

func (repo *sqliteServerSettingsRepository) Get() (*settings.ServerSettings, error) {
	result, err := repo.sqliteGenericRepo.FindByID(serverSettingsID)
	if err != nil {
		return settings.NewDefaultServerSettings(), nil
	}
	return result.(*settings.ServerSettings), nil
}

func (repo *sqliteServerSettingsRepository) Upsert(s *settings.ServerSettings) error {
	if s.Entity == nil {
		s.Entity = entities.NewEntity()
	}
	s.Entity.ID = serverSettingsID
	payload, err := json.Marshal(s)
	if err != nil {
		return err
	}
	_, err = repo.db.Exec(
		"INSERT OR REPLACE INTO server_settings (id, data) VALUES (?, ?)",
		serverSettingsID,
		string(payload),
	)
	return err
}
