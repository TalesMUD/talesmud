package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities/audit"
)

// AuditRepository stores the change log and raw entity snapshots for undo.
type AuditRepository interface {
	Insert(entry *audit.Entry) error
	Update(entry *audit.Entry) error
	FindByID(id string) (*audit.Entry, error)
	List(entityType, entityID string, limit int) ([]*audit.Entry, error)
	LoadRaw(entityType, id string) (json.RawMessage, error)
	SaveRaw(entityType, id string, raw json.RawMessage) error
	DeleteRaw(entityType, id string) error
}

// entityTables is the creator CRUD set. Keys are audit entity types.
var entityTables = map[string]string{
	"rooms":               "rooms",
	"npcs":                "npcs",
	"items":               "items",
	"loottables":          "loot_tables",
	"spawners":            "npc_spawners",
	"dialogs":             "dialogs",
	"quests":              "quests",
	"scripts":             "scripts",
	"skills":              "skills",
	"character-templates": "charactertemplates",
	"settings":            "server_settings",
}

type sqliteAuditRepository struct {
	db *sql.DB
}

// NewSQLiteAuditRepository creates the audit log repository.
func NewSQLiteAuditRepository(client *dbsqlite.Client) AuditRepository {
	return &sqliteAuditRepository{db: client.DB()}
}

func (repo *sqliteAuditRepository) Insert(entry *audit.Entry) error {
	if entry == nil || entry.ID == "" {
		return errors.New("audit entry id is required")
	}
	payload, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	_, err = repo.db.Exec(`INSERT INTO audit_log (id, data) VALUES (?, ?)`, entry.ID, string(payload))
	return err
}

func (repo *sqliteAuditRepository) Update(entry *audit.Entry) error {
	if entry == nil || entry.ID == "" {
		return errors.New("audit entry id is required")
	}
	payload, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	res, err := repo.db.Exec(`UPDATE audit_log SET data = ? WHERE id = ?`, string(payload), entry.ID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("audit entry not found")
	}
	return nil
}

func (repo *sqliteAuditRepository) FindByID(id string) (*audit.Entry, error) {
	if id == "" {
		return nil, errors.New("empty id")
	}
	var payload string
	err := repo.db.QueryRow(`SELECT data FROM audit_log WHERE id = ?`, id).Scan(&payload)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("audit entry not found")
		}
		return nil, err
	}
	entry := &audit.Entry{}
	if err := json.Unmarshal([]byte(payload), entry); err != nil {
		return nil, err
	}
	return entry, nil
}

func (repo *sqliteAuditRepository) List(entityType, entityID string, limit int) ([]*audit.Entry, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	query := `SELECT data FROM audit_log`
	args := make([]interface{}, 0, 3)
	where := make([]string, 0, 2)
	if entityType != "" {
		where = append(where, `json_extract(data, '$.entityType') = ?`)
		args = append(args, entityType)
	}
	if entityID != "" {
		where = append(where, `json_extract(data, '$.entityId') = ?`)
		args = append(args, entityID)
	}
	if len(where) > 0 {
		query += " WHERE " + where[0]
		for _, clause := range where[1:] {
			query += " AND " + clause
		}
	}
	query += ` ORDER BY json_extract(data, '$.time') DESC LIMIT ?`
	args = append(args, limit)

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*audit.Entry, 0)
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		entry := &audit.Entry{}
		if err := json.Unmarshal([]byte(payload), entry); err != nil {
			continue
		}
		out = append(out, entry)
	}
	return out, rows.Err()
}

func (repo *sqliteAuditRepository) table(entityType string) (string, error) {
	table, ok := entityTables[entityType]
	if !ok {
		return "", fmt.Errorf("unknown entity type %q", entityType)
	}
	return table, nil
}

func (repo *sqliteAuditRepository) LoadRaw(entityType, id string) (json.RawMessage, error) {
	table, err := repo.table(entityType)
	if err != nil {
		return nil, err
	}
	if id == "" {
		return nil, nil
	}
	var payload string
	err = repo.db.QueryRow(fmt.Sprintf(`SELECT data FROM %s WHERE id = ?`, table), id).Scan(&payload)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return json.RawMessage(payload), nil
}

func (repo *sqliteAuditRepository) SaveRaw(entityType, id string, raw json.RawMessage) error {
	table, err := repo.table(entityType)
	if err != nil {
		return err
	}
	if id == "" {
		return errors.New("entity id is required")
	}
	if len(raw) == 0 {
		return errors.New("entity payload is empty")
	}
	_, err = repo.db.Exec(
		fmt.Sprintf(`INSERT INTO %s (id, data) VALUES (?, ?) ON CONFLICT(id) DO UPDATE SET data = excluded.data`, table),
		id,
		string(raw),
	)
	return err
}

func (repo *sqliteAuditRepository) DeleteRaw(entityType, id string) error {
	table, err := repo.table(entityType)
	if err != nil {
		return err
	}
	if id == "" {
		return errors.New("entity id is required")
	}
	_, err = repo.db.Exec(fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, table), id)
	return err
}
