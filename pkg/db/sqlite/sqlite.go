package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	_ "modernc.org/sqlite"
)

// Client wraps a SQLite connection and handles initialization.
type Client struct {
	db *sql.DB
}

// Open opens a SQLite database file and applies basic pragmas.
func Open(path string) (*Client, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	client := &Client{db: db}
	if err := client.applyPragmas(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := client.InitSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := client.runMigrations(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return client, nil
}

// DB returns the underlying sql.DB.
func (c *Client) DB() *sql.DB {
	return c.db
}

// Close closes the database connection.
func (c *Client) Close() error {
	return c.db.Close()
}

func (c *Client) applyPragmas() error {
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA synchronous=NORMAL;",
		"PRAGMA busy_timeout=5000;",
	}
	for _, stmt := range pragmas {
		if _, err := c.db.Exec(stmt); err != nil {
			return fmt.Errorf("sqlite pragma failed: %w", err)
		}
	}
	return nil
}

// InitSchema creates the tables used by the repositories.
func (c *Client) InitSchema() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (id TEXT PRIMARY KEY, data TEXT NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS characters (id TEXT PRIMARY KEY, data TEXT NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS charactertemplates (id TEXT PRIMARY KEY, data TEXT NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS rooms (id TEXT PRIMARY KEY, data TEXT NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS items (id TEXT PRIMARY KEY, data TEXT NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS scripts (id TEXT PRIMARY KEY, data TEXT NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS npcs (id TEXT PRIMARY KEY, data TEXT NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS npc_spawners (id TEXT PRIMARY KEY, data TEXT NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS dialogs (id TEXT PRIMARY KEY, data TEXT NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS conversations (id TEXT PRIMARY KEY, data TEXT NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS parties (id TEXT PRIMARY KEY, data TEXT NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS loot_tables (id TEXT PRIMARY KEY, data TEXT NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS server_settings (id TEXT PRIMARY KEY, data TEXT NOT NULL);`,
	}
	for _, stmt := range stmts {
		if _, err := c.db.Exec(stmt); err != nil {
			return fmt.Errorf("sqlite schema init failed: %w", err)
		}
	}
	return nil
}

// runMigrations performs data migrations after schema init.
func (c *Client) runMigrations() error {
	if err := c.migrateItemTemplates(); err != nil {
		return fmt.Errorf("item templates migration failed: %w", err)
	}
	return nil
}

// migrateItemTemplates migrates data from the old itemtemplates table to items table.
// It adds isTemplate: true to each item and removes the script field.
func (c *Client) migrateItemTemplates() error {
	// Check if itemtemplates table exists
	var count int
	err := c.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='itemtemplates'").Scan(&count)
	if err != nil || count == 0 {
		return nil // No migration needed
	}

	log.Info("Migrating itemtemplates to items table...")

	// Read all itemtemplates first (collect before writing to avoid deadlock with single connection)
	type templateRow struct {
		id   string
		data string
	}
	var templates []templateRow

	rows, err := c.db.Query("SELECT id, data FROM itemtemplates")
	if err != nil {
		return err
	}
	for rows.Next() {
		var t templateRow
		if err := rows.Scan(&t.id, &t.data); err != nil {
			log.WithError(err).Warn("Failed to scan itemtemplate row")
			continue
		}
		templates = append(templates, t)
	}
	rows.Close() // Close before writing

	if err := rows.Err(); err != nil {
		return err
	}

	// Now insert into items table
	migratedCount := 0
	for _, t := range templates {
		// Parse JSON, add isTemplate flag, remove script field
		var itemData map[string]interface{}
		if err := json.Unmarshal([]byte(t.data), &itemData); err != nil {
			log.WithError(err).WithField("id", t.id).Warn("Failed to parse itemtemplate JSON")
			continue
		}

		itemData["isTemplate"] = true
		delete(itemData, "script") // Remove script field

		newData, err := json.Marshal(itemData)
		if err != nil {
			log.WithError(err).WithField("id", t.id).Warn("Failed to marshal migrated item")
			continue
		}

		// Insert into items table (upsert to handle conflicts)
		_, err = c.db.Exec(
			"INSERT OR REPLACE INTO items (id, data) VALUES (?, ?)",
			t.id, string(newData),
		)
		if err != nil {
			log.WithError(err).WithField("id", t.id).Warn("Failed to migrate itemtemplate")
			continue
		}
		migratedCount++
	}

	// Drop the old itemtemplates table
	_, err = c.db.Exec("DROP TABLE IF EXISTS itemtemplates")
	if err != nil {
		return fmt.Errorf("failed to drop itemtemplates table: %w", err)
	}

	log.WithField("count", migratedCount).Info("Item templates migration complete")
	return nil
}
