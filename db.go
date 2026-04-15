package main

import (
	"crypto/sha1"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

// DB wraps a sql.DB connection to the shadow SQLite database.
type DB struct {
	db *sql.DB
}

// dbPath returns the path to the global knockout database.
// Uses $XDG_STATE_HOME/knockout/knockout.db, defaulting to ~/.local/state/knockout/knockout.db.
func dbPath() string {
	stateHome := os.Getenv("XDG_STATE_HOME")
	if stateHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		stateHome = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(stateHome, "knockout", "knockout.db")
}

// OpenDB opens (or creates) the shadow database and runs pending migrations.
func OpenDB() (*DB, error) {
	path := dbPath()
	if path == "" {
		return nil, fmt.Errorf("cannot determine database path")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	sqlDB, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	// Enable WAL and foreign keys
	if _, err := sqlDB.Exec("PRAGMA journal_mode = WAL"); err != nil {
		sqlDB.Close()
		return nil, err
	}
	if _, err := sqlDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		sqlDB.Close()
		return nil, err
	}

	d := &DB{db: sqlDB}
	if err := d.migrate(); err != nil {
		sqlDB.Close()
		return nil, err
	}
	return d, nil
}

// Close closes the database connection.
func (d *DB) Close() {
	if d.db != nil {
		d.db.Close()
	}
}

// migrate runs the schema if not already applied.
func (d *DB) migrate() error {
	// Check if schema_migrations exists
	var name string
	err := d.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&name)
	if err == sql.ErrNoRows {
		// Fresh database: run full schema
		if _, err := d.db.Exec(schemaSQL); err != nil {
			return fmt.Errorf("schema init: %w", err)
		}
		_, err := d.db.Exec("INSERT INTO schema_migrations (version, applied_at) VALUES (2, ?)",
			time.Now().UTC().Format(time.RFC3339))
		return err
	}
	if err != nil {
		return err
	}

	// Check current version, run incremental migrations if needed.
	var version int
	if err := d.db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version); err != nil {
		return err
	}

	if version < 2 {
		if err := d.migrateV2(); err != nil {
			return fmt.Errorf("migrate v2: %w", err)
		}
	}

	return nil
}

// migrateV2 removes the UNIQUE constraint from projects.prefix.
// The prefix column is metadata describing what prefix tickets use in that directory,
// but uniqueness is enforced at the Registry layer, not the DB layer.
func (d *DB) migrateV2() error {
	// SQLite doesn't support ALTER TABLE DROP CONSTRAINT, so recreate the table.
	// DROP IF EXISTS handles partial previous runs.
	// Disable foreign keys during migration since tickets references projects.
	migrations := []string{
		`PRAGMA foreign_keys = OFF`,
		`DROP TABLE IF EXISTS projects_new`,
		`CREATE TABLE projects_new (
			id          INTEGER PRIMARY KEY,
			tag         TEXT    NOT NULL UNIQUE,
			prefix      TEXT    NOT NULL,
			tickets_dir TEXT    NOT NULL UNIQUE,
			is_default  INTEGER NOT NULL DEFAULT 0 CHECK (is_default IN (0, 1)),
			created_at  TEXT    NOT NULL
		)`,
		`INSERT INTO projects_new SELECT * FROM projects`,
		`DROP TABLE projects`,
		`ALTER TABLE projects_new RENAME TO projects`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_projects_single_default
			ON projects(is_default) WHERE is_default = 1`,
		`PRAGMA foreign_keys = ON`,
	}

	for _, m := range migrations {
		if _, err := d.db.Exec(m); err != nil {
			preview := m
			if len(preview) > 40 {
				preview = preview[:40]
			}
			return fmt.Errorf("exec %q: %w", preview, err)
		}
	}

	_, err := d.db.Exec("INSERT INTO schema_migrations (version, applied_at) VALUES (2, ?)",
		time.Now().UTC().Format(time.RFC3339))
	return err
}

// Lazy global DB handle. Initialized on first shadow write.
var (
	shadowOnce sync.Once
	shadowDB   *DB
)

// getShadowDB returns the lazily-initialized shadow database.
// Returns nil if the DB cannot be opened (best-effort).
func getShadowDB() *DB {
	shadowOnce.Do(func() {
		db, err := OpenDB()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ko: shadow db: %v\n", err)
			return
		}
		shadowDB = db
	})
	return shadowDB
}

// knockoutNamespace is a fixed UUID used as the namespace for deterministic
// UUID v5 generation. Chosen arbitrarily; must never change.
var knockoutNamespace = [16]byte{
	0x6b, 0x6f, 0x2d, 0x74, 0x69, 0x63, 0x6b, 0x65,
	0x74, 0x2d, 0x6e, 0x73, 0x00, 0x00, 0x00, 0x01,
}

// ticketUUID generates a deterministic UUID v5 from a project prefix and ticket ID.
// The same prefix+ticketID always produces the same UUID.
func ticketUUID(prefix, ticketID string) string {
	h := sha1.New()
	h.Write(knockoutNamespace[:])
	h.Write([]byte(prefix + ":" + ticketID))
	sum := h.Sum(nil)
	// Set version 5 (bits 4-7 of byte 6)
	sum[6] = (sum[6] & 0x0f) | 0x50
	// Set variant (bits 6-7 of byte 8)
	sum[8] = (sum[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		sum[0:4], sum[4:6], sum[6:8], sum[8:10], sum[10:16])
}
